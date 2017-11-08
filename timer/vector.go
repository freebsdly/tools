package timer

import (
	"container/list"
	"errors"
)

// define errors
var (
	timeVectorNumberError = errors.New("number must large than 0")
	maxListNumberError    = errors.New("time Vector initial over than max list number")
)

//
const (
	maxListNumber = 256
)

// time vector contains a list array, it stories
// timeEvent in the list.
type timeVector struct {
	vector []*list.List
}

// init time vectors.
func (p *timeVector) init(num int) error {
	if num <= 0 {
		return timeVectorNumberError
	}

	if num > maxListNumber {
		return maxListNumberError
	}
	p.vector = make([]*list.List, num, num)
	for i := 0; i < num; i++ {
		p.vector[i] = list.New()
	}

	return nil
}

// return a time vector with given length list array
func newTimeVector(num int) (tv *timeVector, err error) {
	tv = &timeVector{}
	err = tv.init(num)
	return tv, err

}
