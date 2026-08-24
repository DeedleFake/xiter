package xiter

import (
	"cmp"
	"context"
	"iter"
	"slices"
	"strings"
)

// Collect collects all items yielded by seq into a slice. It is
// equivalent to [slices.Collect] and is provided purely for
// convenience.
func (seq Seq[T]) Collect() []T {
	return slices.Collect(seq.Seq())
}

// CollectSize pre-allocates the slice being collected into to the
// given size. It is provided purely for convenience.
func (seq Seq[T]) CollectSize(len int) []T {
	return slices.AppendSeq(make([]T, 0, len), seq.Seq())
}

// Find returns the first value of seq for which f(value) returns
// true.
func (seq Seq[T]) Find(f func(T) bool) (r T, ok bool) {
	for v := range seq {
		if f(v) {
			return v, true
		}
	}
	return r, ok
}

// Contains returns true if v is an element of seq.
func Contains[T comparable](seq Seq[T], v T) bool {
	_, ok := seq.Find(func(e T) bool { return v == e })
	return ok
}

// Any returns true if f(element) is true for any elements of seq.
func (seq Seq[T]) Any(f func(T) bool) bool {
	_, ok := seq.Find(f)
	return ok
}

// All returns true if f(element) is true for every element of seq.
func (seq Seq[T]) All(f func(T) bool) bool {
	return !seq.Any(func(v T) bool { return !f(v) })
}

// Reduce calls reducer on each value of seq, passing it initial as
// its first argument on the first call and then the result of the
// previous call for each call after that. It returns the final value
// returned by reducer.
//
// Reduce can be somewhat complicated to get the hang of, but very
// powerful. For example, a simple summation of values can be done as
//
//	sum := seq.Reduce(0, func(total, v int) int { return total + v })
func (seq Seq[T]) Reduce[R any](initial R, reducer func(R, T) R) R {
	for v := range seq {
		initial = reducer(initial, v)
	}
	return initial
}

// Fold performs a [Seq.Reduce] but uses the first value yielded by
// seq instead of a provided initial value. If seq doesn't yield any
// values, the zero value of T is returned.
func (seq Seq[T]) Fold(reducer func(T, T) T) T {
	var prev T
	r := func(v1, v2 T) T { return v2 }
	for v := range seq {
		prev = r(prev, v)
		r = reducer
	}
	return prev
}

// Sum returns the values of seq added together in the order that they
// are yielded.
func Sum[T Addable](seq Seq[T]) T {
	return seq.Fold(func(total, v T) T { return total + v })
}

// Product returns the values of seq multiplied together. It returns
// 1 if no values are yielded.
func Product[T Multiplyable](seq Seq[T]) T {
	return seq.Reduce(1, func(product, v T) T { return product * v })
}

// IsSorted returns true if each element of seq is greater than or
// equal to the previous one.
func IsSorted[T cmp.Ordered](seq Seq[T]) bool {
	return seq.IsSortedFunc(cmp.Compare)
}

// IsSortedFunc is like [IsSorted] but uses a custom comparison
// function.
func (seq Seq[T]) IsSortedFunc(compare func(T, T) int) bool {
	var prev T
	c := func(T, T) int { return -1 }

	for v := range seq {
		if c(prev, v) > 0 {
			return false
		}
		c, prev = compare, v
	}
	return true
}

// Equal returns true if seq1 and seq2 are the same length and each
// element of each is equal to the element at the same point in the
// sequence of the other.
func Equal[T cmp.Ordered](seq1, seq2 Seq[T]) bool {
	return EqualFunc(seq1, seq2, func(v1, v2 T) bool { return v1 == v2 })
}

// EqualFunc is like [Equal] but uses a custom comparison function to
// determine the equivalence of the elements of each sequence.
func EqualFunc[T1, T2 any](seq1 Seq[T1], seq2 Seq[T2], equal func(T1, T2) bool) bool {
	p1, stop := iter.Pull(seq1.Seq())
	defer stop()
	p2, stop := iter.Pull(seq2.Seq())
	defer stop()

	for {
		v1, ok1 := p1()
		v2, ok2 := p2()
		if !ok1 && !ok2 {
			return true
		}
		if (ok1 != ok2) || !equal(v1, v2) {
			return false
		}
	}
}

// Drain empties seq, returning the last value yielded. If no values
// are yielded, ok will be false.
func (seq Seq[T]) Drain() (v T, ok bool) {
	for val := range seq {
		v = val
		ok = true
	}
	return v, ok
}

// CollectSplit is like [Collect], but for a SplitSeq.
func (seq SplitSeq[T1, T2]) CollectSplit() (y1 []T1, y2 []T2) {
	return seq.AppendSplitTo(y1, y2)
}

// AppendSplitTo collects the elements of seq by appending them to
// existing slices.
func (seq SplitSeq[T1, T2]) AppendSplitTo(s1 []T1, s2 []T2) ([]T1, []T2) {
	seq(
		func(v T1) bool {
			s1 = append(s1, v)
			return true
		},
		func(v T2) bool {
			s2 = append(s2, v)
			return true
		},
	)
	return s1, s2
}

// Partition returns two slices, one containing all of the elements of
// seq for which f(element) is true and one containing all of those
// for which it is false.
func (seq Seq[T]) Partition(f func(T) bool) (true, false []T) {
	return seq.PartitionInto(f, true, false)
}

// PartitionInto performs a [Seq.Partition] by appending to two
// existing slices.
func (seq Seq[T]) PartitionInto(f func(T) bool, true, false []T) ([]T, []T) {
	return seq.Split(f).AppendSplitTo(true, false)
}

// Min returns the minimum element yielded by seq or the zero value if
// seq doesn't yield anything.
func Min[T cmp.Ordered](seq Seq[T]) T {
	return seq.Fold(func(v1, v2 T) T { return min(v1, v2) })
}

// Max returns maximum element yielded by seq or the zero value if seq
// doesn't yield anything.
func Max[T cmp.Ordered](seq Seq[T]) T {
	return seq.Fold(func(v1, v2 T) T { return max(v1, v2) })
}

// SendContext sends values from seq to c repeatedly until either the
// sequence ends or ctx is canceled. It blocks until one of those two
// things happens.
func (seq Seq[T]) SendContext(ctx context.Context, c chan<- T) {
	for v := range seq {
		select {
		case <-ctx.Done():
			return
		case c <- v:
		}
	}
}

// StringJoin works exactly like [strings.Join] but it operates on a
// [Seq] instead of a []string.
func StringJoin(seq Seq[string], sep string) string {
	var buf strings.Builder
	var s string
	for v := range seq {
		buf.WriteString(s)
		buf.WriteString(v)
		s = sep
	}
	return buf.String()
}
