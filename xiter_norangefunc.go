//go:build !goexperiment.rangefunc

package xiter

type _Seq[T any] func(yield func(T) bool)
type _Seq2[T1, T2 any] func(yield func(T1, T2) bool)

// Seq represents an iterator over a sequence of values. When called,
// the passed yield function is called for each successive value.
// Returning false from yield causes the iterator to stop, equivalent
// to a break statement.
type Seq[T any] _Seq[T]

// Seq2 represents a two-value iterator.
type Seq2[T1, T2 any] _Seq2[T1, T2]

func _Pull[T any](seq _Seq[T]) (iter func() (T, bool), stop func()) {
	return _GoPull[T](seq)
}
