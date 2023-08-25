package xiter

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
