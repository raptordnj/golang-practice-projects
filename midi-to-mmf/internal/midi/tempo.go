package midi

// TimeMicros converts an absolute tick position to microseconds from the
// start of the song, integrating the tempo map. For SMPTE-timed songs it is
// a simple linear scale.
func (s *Song) TimeMicros(tick int64) int64 {
	if s.IsSMPTE {
		return tick * 1_000_000 / int64(s.TicksPerSecond)
	}
	// Accumulate tick*usec products exactly and divide once, so segment
	// boundaries don't compound truncation error.
	var acc int64
	var prevTick int64
	usPerQN := int64(500000)
	for _, tc := range s.TempoChanges {
		if tc.Tick >= tick {
			break
		}
		acc += (tc.Tick - prevTick) * usPerQN
		prevTick = tc.Tick
		usPerQN = int64(tc.A)
	}
	acc += (tick - prevTick) * usPerQN
	return acc / int64(s.TicksPerQN)
}

// TotalTicks is the tick position just past the last event.
func (s *Song) TotalTicks() int64 {
	var max int64
	for _, e := range s.Events {
		if e.Tick > max {
			max = e.Tick
		}
	}
	return max
}
