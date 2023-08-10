package goevent

type Event interface {
	Topic() string
}

type Listener func(event Event)

type stoppable interface {
	IsPropagationStopped() bool
}

type Stoppable struct {
	propagation bool
}

func (s *Stoppable) IsPropagationStopped() bool {
	return s.propagation
}
func (s *Stoppable) StopPropagation() {
	s.propagation = true
}
