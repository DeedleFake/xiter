// Package xiter provides iterator-related functionality compatible
// with, but not requiring, Go 1.23.
package xiter

import "sync"

// A SplitSeq is like a Seq but can yield via either of two functions.
// It might not be useful, but is included anyways because it might
// be.
type SplitSeq[T1, T2 any] func(y1 func(T1) bool, y2 func(T2) bool)

// Pair contains two values of arbitrary types.
type Pair[T1, T2 any] struct {
	V1 T1
	V2 T2
}

// P returns a Pair containing v1 and v2.
func P[T1, T2 any](v1 T1, v2 T2) Pair[T1, T2] {
	return Pair[T1, T2]{V1: v1, V2: v2}
}

// Split is a convenience function that just returns the two values
// contained in the pair.
func (p Pair[T1, T2]) Split() (T1, T2) {
	return p.V1, p.V2
}

// GoPull simulates a pull-iterator using Go's built-in concurrency
// primitives in lieu of coroutines. It handles all synchronization
// internally, so, despite running the iterator in a new thread, there
// shouldn't be any data races, but there is some performance
// overhead.
//
// The returned stop function must be called when the iterator is no
// longer in use.
func GoPull[T any](seq Seq[T]) (iter func() (T, bool), stop func()) {
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

// Addable is a type that should probably exist in the standard
// library somewhere because it's quite command and kind of a pain to
// write every time I need it.
type Addable interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64 | complex64 | complex128 | string
}

type Multiplyable interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64 | complex64 | complex128
}
