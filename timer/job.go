package timer

import (
	"container/list"
)

// define job
type job struct {
	period  uint32 // time event will be repeated exection at the specific time interval
	expires uint32 // unix timestramp + period
	jober   Jober  // the callback function
	oneTime bool
	lst     *list.List
	element *list.Element
}

// init job
func (p *job) init(now uint32) {
	p.lst = list.New()
	p.element = nil
	p.expires = p.period + now
}
