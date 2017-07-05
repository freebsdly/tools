// timewheel 时间轮定时器

package timewheel

import (
	"container/list"
	"errors"
	"log"
	"sync"
	"time"
	"tools/pool"
)

//
const (
	TV_COUNT = 5
	TVR_BITS = 8
	TVN_BITS = 6
	TVR_SIZE = 1 << TVR_BITS
	TVN_SIZE = 1 << TVN_BITS
	TVR_MASK = TVR_SIZE - 1
	TVN_MASK = TVN_SIZE - 1
)

// time vector contains a list array, it stories
// timeEvent in the list.
type timeVector struct {
	vector []*list.List
}

// init time vector.
func (p *timeVector) init(num int) error {
	if num <= 0 {
		return errors.New("number must large than 0")
	}
	for i := 0; i < num; i++ {
		p.vector = append(p.vector, list.New())
	}

	return nil
}

// return a time vector with given length list array
func newTimeVector(num int) *timeVector {
	var t *timeVector = &timeVector{}
	t.init(num)
	return t

}

// tvecbase have TV_COUNT timewheel,first have TVR_SIZE length
// and othse have TVN_SIZE length
type TimeWheel struct {
	// jiffies store the time counter start from program run.
	jiffies uint32

	// time vector array. all the time event will be storied in
	// this array.
	tvs [TV_COUNT]*timeVector

	// lock
	sync.RWMutex
}

// init and return a time wheel
func NewTimeWheel() *TimeWheel {
	var t *TimeWheel = &TimeWheel{}
	t.Init()
	return t
}

// Init TimeWheel
func (p *TimeWheel) Init() {
	p.tvs[0] = newTimeVector(TVR_SIZE)
	for i := 1; i < TV_COUNT; i++ {
		p.tvs[i] = newTimeVector(TVN_SIZE)
	}
}

// return the current jiffies
func (p *TimeWheel) Jiffies() uint32 {
	p.RLock()
	i := p.jiffies
	p.RUnlock()
	return i
}

// return current index of the time vector n [0,1,2,3,4]
func (p *TimeWheel) timeVectorIndex(n int) uint32 {
	if n == 0 {
		return p.Jiffies() & TVR_MASK
	}
	return (p.Jiffies() >> (TVR_BITS + uint32(n-1)*TVN_BITS)) & TVN_MASK
}

// add time event to timewheel's time vector
func (p *TimeWheel) AddTimeEvent(t *TimeEvent) (*TimeEvent, error) {
	var (
		offset uint32 = t.Expires - p.Jiffies()
		index  uint32
		iwheel int
	)
	if offset <= 0 {
		index = p.timeVectorIndex(0)
		iwheel = 0
	} else if offset < TVR_SIZE {
		index = t.Expires & TVR_MASK
		iwheel = 0
	} else if offset < 1<<(TVR_BITS+TVN_BITS) {
		index = (t.Expires >> TVR_BITS) & TVN_MASK
		iwheel = 1
	} else if offset < 1<<(TVR_BITS+2*TVN_BITS) {
		index = (t.Expires >> (TVR_BITS + TVN_BITS)) & TVN_MASK
		iwheel = 2
	} else if offset < 1<<(TVR_BITS+3*TVN_BITS) {
		index = (t.Expires >> (TVR_BITS + 2*TVN_BITS)) & TVN_MASK
		iwheel = 3
	} else {
		index = (t.Expires >> (TVR_BITS + 3*TVN_BITS)) & TVN_MASK
		iwheel = 4
	}
	p.Lock()
	e := p.tvs[iwheel].vector[index].PushBack(t)
	t.element = e
	t.lst = p.tvs[iwheel].vector[index]
	p.Unlock()
	return t, nil
}

// Delete Time Event from timewheel
func (p *TimeWheel) DeleteTimeEvent(t *TimeEvent) error {
	if t.element == nil {
		return errors.New("TimeEvent.element is nil.")
	}
	if t.lst == nil {
		return errors.New("TimeEvent.lst is nil.")
	}
	t.lst.Remove(t.element)
	return nil
}

// execute timer on current index of tv
func (p *TimeWheel) runTimers() {
	var (
		index    = p.Jiffies() & TVR_MASK
		workList *list.List
	)
	p.cascade()
	workList = p.tvs[0].vector[index]
	p.deleteTimeList(uint32(0), index)
	go p.runList(workList)

}

// recompute and reput the timer to base slot
func (p *TimeWheel) cascade() {
	tv1_index := p.timeVectorIndex(0)
	if tv1_index != 0 {
		return
	}
	for i := 1; i < TV_COUNT; i++ {
		idx := p.timeVectorIndex(i)
		if idx == 0 {
			return
		}
		cascadelist := p.tvs[i].vector[idx]
		for e := cascadelist.Front(); e != nil; e = e.Next() {
			p.AddTimeEvent(e.Value.(*TimeEvent))
		}
		p.deleteTimeList(uint32(i), idx)
	}
}

// run event from list
func (p *TimeWheel) runList(workList *list.List) {
	pool.TaskPool.Init(pool.TaskNum, workList.Len())
	for e := workList.Front(); e != nil; e = e.Next() {
		tl := e.Value.(*TimeEvent)
		if tl.Func == nil {
			log.Printf("timeEvent name %s, func is nil.\n", tl.Name)
			continue
		}
		go p.runfunc(tl)

		if !tl.OneTime {
			tl.Init(p.Jiffies())
			p.AddTimeEvent(tl)
		}
	}
	pool.TaskPool.Wg.Wait()
}

// run callback function use goroutine
func (p *TimeWheel) runfunc(t *TimeEvent) {
	pool.TaskPool.AddTask()
	log.Printf("run TimeEvent, name :%s, period: %d\n", t.Name, t.Period)
	f := t.Func

	f(t.Args)

	pool.TaskPool.DeleteTask()
}

func (p *TimeWheel) deleteTimeList(iw, index uint32) {
	switch iw {
	case 0:
		if index > 255 || index < 0 {
			return
		}
	case 1, 2, 3, 4:
		if index > 63 || index < 0 {
			return
		}
	default:
		return
	}
	p.Lock()
	p.tvs[iw].vector[index] = list.New()
	p.Unlock()
}

// tickerRun make base.jiffies auto increate self.
func (p *TimeWheel) tickerRun() {
	ticker := time.NewTicker(time.Duration(1000) * time.Millisecond)
	for {
		<-ticker.C
		go p.runTimers()
		p.jiffies += 1
	}
	defer ticker.Stop()
}

// run timewheel
func (p *TimeWheel) Run() {
	go p.tickerRun()
}

type TimeEvent struct {
	Name    string            //the name of this time event
	Period  uint32            // time event will be repeated exection at the specific time interval
	Expires uint32            // unix timestramp + period
	Func    func(interface{}) // the callback function
	Args    interface{}       // the argvments for the callback function
	OneTime bool
	lst     *list.List
	element *list.Element
}

// Init TimeEvent
func (p *TimeEvent) Init(now uint32) {
	p.lst = list.New()
	p.element = nil
	p.Expires = p.Period + now
}
