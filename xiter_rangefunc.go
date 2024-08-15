//go:build (goexperiment.rangefunc || go1.23) && !goexperiment.aliastypeparams

package xiter

import "iter"

// Seq represents an iterator over a sequence of values. When called,
// the passed yield function is called for each successive value.
// Returning false from yield causes the iterator to stop, equivalent
// to a break statement.
type Seq[T any] iter.Seq[T] // Type alias would be nice, but not supported for generic types.

// Seq2 represents a two-value iterator.
type Seq2[T1, T2 any] iter.Seq2[T1, T2]

// Pull converts a push iterator in the form of a [Seq] into a pull
// iterator. On versions of Go that have support for this capability
// natively, that is used, while on versions of Go prior to that this
// is simply a wrapper around [GoPull].
func Pull[T any](seq Seq[T]) (iterator func() (T, bool), stop func()) {
	return iter.Pull[T](iter.Seq[T](seq))
}
