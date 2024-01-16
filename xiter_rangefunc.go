//go:build goexperiment.rangefunc

package xiter

import "iter"

// _Seq represents an iterator over a sequence of values. When called,
// the passed yield function is called for each successive value.
// Returning false from yield causes the iterator to stop, equivalent
// to a break statement.
type _Seq[T any] iter.Seq[T] // Type alias would be nice, but not supported for generic types.

// _Seq2 represents a two-value iterator.
type _Seq2[T1, T2 any] iter.Seq2[T1, T2]

func Pull[T any](seq iter.Seq[T]) (iterator func() (T, bool), stop func()) {
	return iter.Pull[T](seq)
}
