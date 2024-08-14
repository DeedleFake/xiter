//go:build !goexperiment.rangefunc && !go1.23

package xiter

// Seq represents an iterator over a sequence of values. When called,
// the passed yield function is called for each successive value.
// Returning false from yield causes the iterator to stop, equivalent
// to a break statement.
type Seq[T any] func(yield func(T) bool)

// Seq2 represents a two-value iterator.
type Seq2[T1, T2 any] func(yield func(T1, T2) bool)

func Pull[T any](seq Seq[T]) (iter func() (T, bool), stop func()) {
	return GoPull[T](seq)
}
