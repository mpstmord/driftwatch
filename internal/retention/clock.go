package retention

// SetClock replaces the internal clock used by the Store. Intended for
// testing only; production code should rely on the default time.Now clock
// set by New.
func (s *Store) SetClock(fn func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.now = fn
}
