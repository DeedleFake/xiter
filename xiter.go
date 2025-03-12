// Package xiter provides iterator-related functionality compatible
// with, but not requiring, Go 1.23.
package xiter

import "iter"

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

// Addable is a type that should probably exist in the standard
// library somewhere because it's quite common and kind of a pain to
// write every time I need it.
type Addable interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64 | complex64 | complex128 | string
}

type Multiplyable interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64 | complex64 | complex128
}

// CoroutineFunc is the signature of a coroutine function as passed to
// [Coroutine].
type CoroutineFunc[In, Out any] = func(first In, yield CoroutineYieldFunc[Out, In]) Out

// CoroutineYieldFunc is the signature of a coroutine yield function
// as returned by [Coroutine].
type CoroutineYieldFunc[In, Out any] = func(In) (Out, bool)

// Coroutine starts the provided function as a coroutine. This is
// similar to a pull iterator as returned by [iter.Pull], but allows
// full, two-way communication with the suspended function. The
// returned yield function can be used to pass data into the
// coroutine, while the function given to the coroutine function
// itself can be used to pass data back out, suspending the coroutine
// where it was. All of the caveats and warnings that apply to
// iter.Pull also apply to this.
//
// Coroutine is a somewhat complicated function with a lot of nested
// function types that can be kind of confusing to work with. For a
// simpler function that can handle some cases that Coroutine might be
// necessary for, see [Push].
//
// The coroutine provided is not started until the first call to the
// returned yield function. On the first call, the coroutine is called
// with the data passed to yield as its first argument. All subsequent
// calls to yield will cause the yield function inside of the
// coroutine to return the data provided instead.
//
// The returned stop function returns the final return value of the
// coroutine function. If the coroutine was never started, this will
// return the zero value.
//
// After stop is called, the value returned by yield inside of the
// coroutine will be the last value that was yielded to the coroutine
// before the call to stop.
func Coroutine[In, Out any](coroutine CoroutineFunc[In, Out]) (yield CoroutineYieldFunc[In, Out], stop func() Out) {
	var in In
	var r Out
	next, pstop := iter.Pull(func(yield func(Out) bool) {
		r = coroutine(in, func(v Out) (In, bool) {
			ok := yield(v)
			return in, ok
		})
	})

	yield = func(v In) (Out, bool) {
		in = v
		return next()
	}

	stop = func() Out {
		pstop()
		return r
	}

	return yield, stop
}

// Push is the opposite, in some ways, of [iter.Pull]. Where iter.Pull
// creates a coroutine from which values can be yielded, Push creates
// a coroutine into which values can be yielded. This is useful for
// wrapping certain types of APIs to make them interact more cleanly
// with [iter.Seq], like that shown in the example.
//
// If full, two-way communication with coroutine is necessary, see
// [Coroutine].
func Push[In, Out any](coroutine func(iter.Seq[In]) Out) (yield func(In) bool, stop func() Out) {
	next, stop := Coroutine(func(v In, next func(Out) (In, bool)) Out {
		return coroutine(func(yield func(In) bool) {
			if !yield(v) {
				return
			}

			var zero Out
			for {
				v, ok := next(zero)
				if !ok || !yield(v) {
					return
				}
			}
		})
	})

	yield = func(v In) bool {
		_, ok := next(v)
		return ok
	}

	return yield, stop
}
