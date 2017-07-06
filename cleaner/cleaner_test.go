package cleaner

import (
	"testing"
)

var (
	testValue int
)

func callBack1() {
	testValue = 1
}

func callBack2() {
	testValue = 2
}

func TestCleanerF(t *testing.T) {
	cc := new(Cleaner)
	cc.HandleFunc(callBack1)
	cc.F()
	if testValue != 1 {
		t.Errorf("Cleaner's F run failed\n")
	}
}

func TestCleanerExitClean(t *testing.T) {
	cc := new(Cleaner)
	cc.HandleFunc(callBack2)
	cc.ExitClean()
	if testValue != 2 {
		t.Errorf("Cleaner's ExitClean run failed\n")
	}
}
