package aws

import (
	"time"
)

// AttemptStrategy represents a strategy for waiting for an action
// to complete successfully. This is an internal type used by the
// implementation of other goamz packages.
type AttemptStrategy struct {
	Total time.Duration // total duration of attempt.
	Delay time.Duration // interval between each try in the burst.
}

type Attempt struct {
	strategy AttemptStrategy
	last     time.Time
	end      time.Time
	force    bool
}

// Start begins a new sequence of attempts for the given strategy.
func (s AttemptStrategy) Start() *Attempt {
	return &Attempt{strategy: s}
}

// Next waits until it is time to perform the next attempt or returns
// false if it is time to stop trying.
func (a *Attempt) Next() bool {
	now := time.Now()
	if a.end.IsZero() {
		// First attempt.
		a.last = now
		a.end = now.Add(a.strategy.Total)
		a.force = false
		return true
	}
	if !a.force && !now.Before(a.end) {
		return false
	}
	a.force = false
	if sleep := a.strategy.Delay - now.Sub(a.last); sleep > 0 {
		time.Sleep(sleep)
		now = time.Now()
	}
	a.last = now
	return true
}

// HasNext returns whether another attempt will be made if the current
// one fails. If it returns true, the following call to Next is
// guaranteed to return true.
func (a *Attempt) HasNext() bool {
	if a.end.IsZero() || time.Now().Before(a.end) {
		a.force = true
		return true
	}
	return false
}
