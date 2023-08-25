package xiter

// Map returns a Seq that yields the values of seq transformed via f.
func Map[T1, T2 any](seq Seq[T1], f func(T1) T2) Seq[T2] {
	return func(yield func(T2) bool) bool {
		return seq(func(v T1) bool {
			return yield(f(v))
		})
	}
}

// Filter returns a Seq that yields only the values of seq for which
// f(value) returns true.
func Filter[T any](seq Seq[T], f func(T) bool) Seq[T] {
	return func(yield func(T) bool) bool {
		return seq(func(v T) bool {
			if !f(v) {
				return true
			}
			return yield(v)
		})
	}
}

// Limit returns a Seq that yields at most n values from seq.
func Limit[T any](seq Seq[T], n int) Seq[T] {
	return func(yield func(T) bool) bool {
		return seq(func(v T) bool {
			if !yield(v) {
				return false
			}
			n--
			return n > 0
		})
	}
}

// Concat creates a new Seq that yields the values of each of the
// provided Seqs in turn.
func Concat[T any](seqs ...Seq[T]) Seq[T] {
	return func(yield func(T) bool) bool {
		for _, seq := range seqs {
			seq(yield)
		}
		return false
	}
}
