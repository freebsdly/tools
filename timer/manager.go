//
package timer

import (
	"errors"
	"sync/atomic"
)

// define errors
var (
	jidDoesNotExistError = errors.New("jid does not exist")
)

// manager manage jobs
type manager struct {
	tw      *timeWheel
	jids    map[uint32]*job
	counter uint32
}

// Add new job to manager
func (p *manager) Add(j Jober, period uint32, onetime bool) (jid uint32, err error) {
	wj := &job{
		period:  period,
		oneTime: onetime,
		jober:   j,
	}
	jid = atomic.AddUint32(&p.counter, 1)
	p.jids[jid] = wj
	go p.tw.add(wj)
	return jid, nil

}

// Remove job with it's job id
func (p *manager) Remove(jid uint32) error {
	jb, exist := p.jids[jid]
	if !exist {
		return jidDoesNotExistError
	}
	p.tw.remove(jb)
	return nil
}

// Start Timer
func (p *manager) Start() error {
	go p.tw.tickerRun()
	return nil
}

// return a new manager
func newManager() *manager {
	var mgr *manager = &manager{
		tw:   newTimeWheel(),
		jids: make(map[uint32]*job),
	}
	return mgr

}
