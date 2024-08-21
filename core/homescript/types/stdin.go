package types

import "sync"

//
// STDIN
//

type StdinBuffer struct {
	inputs []string
	lock   sync.RWMutex
}

func NewStdinBuffer() *StdinBuffer {
	return &StdinBuffer{
		inputs: make([]string, 0),
		lock:   sync.RWMutex{},
	}
}

func (s *StdinBuffer) Send(msg string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.inputs = append(s.inputs, msg)
}

func (s *StdinBuffer) Poll() *string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if len(s.inputs) > 0 {
		first := s.inputs[0]
		s.inputs = s.inputs[1:]
		return &first
	}

	return nil
}

//
// END STDIN.
//
