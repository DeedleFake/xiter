// Package xiter provides iterator-related functionality compatible
// with, but not requiring, CL 510541.
package xiter

import "sync"

// Seq represents an iterator over a sequence of values. When called,
// the passed yield function is called for each successive value.
// Returning false from yield causes the iterator to stop, equivalent
// to a break statement. The return value of the Seq function itself
// is completely ignored, but present to be compatible with the CL
// 510541 prototype.
type Seq[T any] func(yield func(T) bool) bool

// Pull simulates a pull-iterator using Go's built-in concurrency
// primitives in lieu of coroutines. It handles all synchronization
// internally, so despite running the iterator in a new thread, there
// shouldn't be any data races, but there is some performance
// overhead.
//
// The returned stop function must be called when the iterator is no
// longer in use.
func Pull[T any](seq Seq[T]) (iter func() (T, bool), stop func()) {
	next := make(chan struct{})
	yield := make(chan T)

	go func() {
		defer close(yield)

		_, ok := <-next
		if !ok {
			return
		}

		seq(func(v T) bool {
			yield <- v
			_, ok := <-next
			return ok
		})
	}()

	return func() (v T, ok bool) {
			select {
			case <-yield:
				return v, false
			case next <- struct{}{}:
				v, ok := <-yield
				return v, ok
			}
		}, sync.OnceFunc(func() {
			close(next)
			<-yield
		})
}
