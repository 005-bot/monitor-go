package scheduler

import "time"

type Config struct {
	Interval     int
	CycleTimeout time.Duration
}

func (c Config) ResolveInterval() time.Duration {
	return time.Duration(c.Interval) * time.Second
}

func (c Config) ResolveCycleTimeout() time.Duration {
	cycleTimeout := min(c.CycleTimeout, time.Duration(c.Interval)*time.Second)
	if cycleTimeout == 0 {
		cycleTimeout = time.Duration(c.Interval) * time.Second
	}

	return cycleTimeout
}
