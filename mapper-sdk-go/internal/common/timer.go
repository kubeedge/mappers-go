package common

import (
	"time"
)

// Timer is to call a function periodically.
type Timer struct {
	Function func()
	Duration time.Duration
	Times    int
	stopFlag bool
}

// Start start a timer.
func (t *Timer) Start() {
	ticker := time.NewTicker(t.Duration)
	if t.Times > 0 {
		for i := 0; i < t.Times; i++ {
			if t.stopFlag == true {
				break
			}
			<-ticker.C
			t.Function()
		}
	} else {
		for range ticker.C {
			if t.stopFlag == true {
				break
			}
			t.Function()
		}
	}
}

// Stop is a method to stop a timer
func (t *Timer) Stop() {
	t.stopFlag = true
}
