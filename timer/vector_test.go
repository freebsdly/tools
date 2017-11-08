//
package timer

import (
	"testing"
)

func TestNewTimeVector(t *testing.T) {
	var num = 10
	tv, err := newTimeVector(num)
	if err != nil {
		t.Fatalf("test newTimeVector failed with num: %d\n", num)
	}
	if len(tv.vector) != num {
		t.Fatalf("test newTimeVector failed, timeVector.vector's length is't eque the test number %d\n", num)
	}

}

func TestNewTimeVectorFailed(t *testing.T) {
	var num = 300
	tv, err := newTimeVector(num)
	if err == nil {
		t.Fatalf("test newTimeVector failed with num: %d\n", num)
	}

	if len(tv.vector) > maxListNumber {
		t.Fatalf("test newTimeVector failed, timeVector.vector's length large than the max list number %d\n", maxListNumber)
	}
}
