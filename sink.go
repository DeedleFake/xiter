package xiter

import "cmp"

// AppendTo appends the values of seq to s, returning the new slice.
func AppendTo[T any, S ~[]T](seq Seq[T], s S) S {
	seq(func(v T) bool {
		s = append(s, v)
		return true
	})
	return s
}

// Collect returns a slice of the elements of seq.
func Collect[T any](seq Seq[T]) []T {
	return AppendTo(seq, []T(nil))
}

// Find returns the first value of seq for which f(value) returns
// true.
func Find[T any](seq Seq[T], f func(T) bool) (r T, ok bool) {
	seq(func(v T) bool {
		if !f(v) {
			return true
		}
		r = v
		ok = true
		return false
	})
	return r, ok
}

// Contains returns true if v is an element of seq.
func Contains[T comparable](seq Seq[T], v T) bool {
	_, ok := Find(seq, func(e T) bool { return v == e })
	return ok
}

// Reduce calls reducer on each value of seq, passing it initial as
// its first argument on the first call and then the result of the
// previous call for each call after that. It returns the final value
// returned by reducer.
//
// Reduce can be somewhat complicated to get the hang of, but very
// powerful. For example, a simple summation of values can be done as
//
//	sum := Reduce(seq, 0, func(total, v int) int { return total + v })
func Reduce[T, R any](seq Seq[T], initial R, reducer func(R, T) R) R {
	seq(func(v T) bool {
		initial = reducer(initial, v)
		return true
	})
	return initial
}

// Sum returns the values of seq added together in the order that they
// are yielded.
func Sum[T Addable](seq Seq[T]) T {
	var zero T
	return Reduce(seq, zero, func(total, v T) T { return total + v })
}

// IsSorted returns true if each element of seq is greater than or
// equal to the previous one.
func IsSorted[T cmp.Ordered](seq Seq[T]) bool {
	return IsSortedFunc(seq, cmp.Compare)
}

// IsSortedFunc is like [IsSorted] but uses a custom comparison
// function.
func IsSortedFunc[T any](seq Seq[T], compare func(T, T) int) bool {
	var prev T
	c := func(T, T) int { return -1 }

	sorted := true
	seq(func(v T) bool {
		sorted = c(prev, v) <= 0
		c, prev = compare, v
		return sorted
	})
	return sorted
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
	p1, stop := Pull(seq1)
	defer stop()
	p2, stop := Pull(seq2)
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

// Drain empties seq, discarding every single value and returning once
// it's finished.
func Drain[T any](seq Seq[T]) {
	seq(func(T) bool { return true })
}

// CollectSplit is like [Collect], but for a SplitSeq.
func CollectSplit[T1, T2 any](seq SplitSeq[T1, T2]) (y1 []T1, y2 []T2) {
	seq(
		func(v T1) bool {
			y1 = append(y1, v)
			return true
		},
		func(v T2) bool {
			y2 = append(y2, v)
			return true
		},
	)
	return y1, y2
}

// Partition returns two slices, one containing all of the elements of
// seq for which f(element) is true and one containing all of those
// for which it is false.
func Partition[T any](seq Seq[T], f func(T) bool) (true, false []T) {
	return CollectSplit(Split(seq, f))
}
