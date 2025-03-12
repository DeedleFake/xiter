package xiter_test

import (
	"fmt"

	"deedles.dev/xiter"
)

type Aggregate[S, R any] interface {
	Step(S)
	Result() R
}

type aggregate struct {
	next func(int) bool
	stop func() int
}

func (s *aggregate) Step(v int) {
	if s.next != nil {
		if !s.next(v) {
			s.next = nil
		}
	}
}

func (s *aggregate) Result() int { return s.stop() }

func ExamplePush() {
	// This example demonstrates creating an implementation of an
	// interface that wraps an iter.Seq in a roundabout way. This is
	// useful for handling cases where some API provides an interface
	// that the user must implement but that does not map cleanly to the
	// iter.Seq API.

	yield, stop := xiter.Push(xiter.Sum[int])
	s := &aggregate{next: yield, stop: stop}

	// This simulates the API using sum via the interface.
	for n := range 10 {
		s.Step(n)
	}
	fmt.Println(s.Result())
	// Output: 45
}
