package clock

import "time"

// Clock abstracts time access for deterministic testing.
type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now().UTC() }

// FixedClock returns a constant time. Handy for tests.
type FixedClock struct{ T time.Time }

func (c FixedClock) Now() time.Time { return c.T }
