package clock

import "time"

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (sc *SystemClock) Now() time.Time {
	return time.Now()
}
