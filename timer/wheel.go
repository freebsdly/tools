//
package timer

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

// define time vector args
const (
	tvCount      = 5                 // total number of the time vectors
	tvRootBits   = 8                 // number of the root time vector bits
	tvNormalBits = 6                 // number of hte normal time vector bits
	tvRootSize   = 1 << tvRootBits   // size of the root time vector
	tvNormalSize = 1 << tvNormalBits // size of the normal time vector
	tvRootMask   = tvRootSize - 1    // mask number of the root time vector
	tvNormalMask = tvNormalSize - 1  // mask number of the normal time vector
)

// tvecbase have tvCount timewheel, first have tvRootSize length
// and othse have tvNormalSize length
type timeWheel struct {
	// jiffies store the time counter start from program run.
	jiffies uint32

	// time vector array. all the time event will be storied in
	// this array.
	tvs [tvCount]*timeVector

	// lock
	sync.RWMutex
}

// init and return a time wheel
func newTimeWheel() *timeWheel {
	var tw *timeWheel = &timeWheel{}
	tw.init()
	return tw
}

// init TimeWheel
func (p *timeWheel) init() error {
	var err error
	p.tvs[0], err = newTimeVector(tvRootSize)
	if err != nil {
		return err
	}
	for i := 1; i < tvCount; i++ {
		p.tvs[i], err = newTimeVector(tvNormalSize)
		if err != nil {
			return err
		}
	}
	return nil
}

// return current index of the time vector n [0,1,2,3,4]
func (p *timeWheel) timeVectorIndex(n int) uint32 {
	if n == 0 {
		return atomic.LoadUint32(&p.jiffies) & tvRootMask
	}
	return (atomic.LoadUint32(&p.jiffies) >> (tvRootBits + uint32(n-1)*tvNormalBits)) & tvNormalMask
}

// add job to timewheel's time vector
func (p *timeWheel) add(j *job) {
	j.init(atomic.LoadUint32(&p.jiffies))
	var (
		expires        = j.expires
		offset  uint32 = expires - atomic.LoadUint32(&p.jiffies)
		index   uint32
		iwheel  int
	)
	if offset < 0 {
		index = p.timeVectorIndex(0)
		iwheel = 0
	} else if offset < tvRootSize {
		index = expires & tvRootMask
		iwheel = 0
	} else if offset < 1<<(tvRootBits+tvNormalBits) {
		index = (expires >> tvRootBits) & tvNormalMask
		iwheel = 1
	} else if offset < 1<<(tvRootBits+2*tvNormalBits) {
		index = (expires >> (tvRootBits + tvNormalBits)) & tvNormalMask
		iwheel = 2
	} else if offset < 1<<(tvRootBits+3*tvNormalBits) {
		index = (expires >> (tvRootBits + 2*tvNormalBits)) & tvNormalMask
		iwheel = 3
	} else {
		index = (expires >> (tvRootBits + 3*tvNormalBits)) & tvNormalMask
		iwheel = 4
	}
	p.Lock()
	j.lst = p.tvs[iwheel].vector[index]
	j.element = j.lst.PushBack(j)
	p.Unlock()
}

// remove
func (p *timeWheel) remove(j *job) {
	j.lst.Remove(j.element)
}

// execute timer on current index of tv
func (p *timeWheel) runTimers() {
	var (
		index    = p.timeVectorIndex(0)
		workList *list.List
	)
	p.cascade()
	workList = p.tvs[0].vector[index]
	p.deleteJobList(uint32(0), index)
	p.runJobList(workList)

}

// recompute and reput jobs into base slot
func (p *timeWheel) cascade() {
	tv1_index := p.timeVectorIndex(0)
	if tv1_index != 0 {
		return
	}
	for i := 1; i < tvCount; i++ {
		idx := p.timeVectorIndex(i)
		if idx == 0 {
			return
		}
		cascadelist := p.tvs[i].vector[idx]
		for e := cascadelist.Front(); e != nil; e = e.Next() {
			p.add(e.Value.(*job))
		}
		p.deleteJobList(uint32(i), idx)
	}
}

// run jobs in time vector's list
func (p *timeWheel) runJobList(workList *list.List) {
	var jb *job
	for e := workList.Front(); e != nil; e = e.Next() {
		jb = e.Value.(*job)
		go jb.jober.Run()
		if !jb.oneTime {
			p.add(jb)
		}
	}
}

// remove time vector's list
func (p *timeWheel) deleteJobList(iw, index uint32) {
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
func (p *timeWheel) tickerRun() {
	ticker := time.NewTicker(time.Duration(1000) * time.Millisecond)
	for {
		p.runTimers()
		<-ticker.C
		atomic.AddUint32(&p.jiffies, 1)
	}
	defer ticker.Stop()
}
