// timer privide a interface to use timer
package timer

import (
	"errors"
)

type TimerType string

// define errors
var (
	TimerTypeError = errors.New("unsupported Timer type")
)

// define timer type
const (
	TIMEWHEEL TimerType = "timewheel"
)

func NewTimer(tt TimerType) (t Timer, err error) {
	switch tt {
	case TIMEWHEEL:
		t = newManager()
	default:
		err = TimerTypeError
	}
	return t, err
}
