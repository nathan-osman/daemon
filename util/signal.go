package util

// Send a signal to any number of listeners.
type Signal struct {
	listeners []chan interface{}
}

// Add a new listener.
func (s *Signal) Listen() chan interface{} {
	c := make(chan interface{})
	s.listeners = append(s.listeners, c)
	return c
}

// Send a value to all listeners.
func (s *Signal) Emit(val interface{}) {
	for _, c := range s.listeners {
		c <- val
	}
}
