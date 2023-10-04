package xiter

import "testing"

func BenchmarkOldZip(b *testing.B) {
	slice1 := []int{1, 2, 3, 4, 5}
	slice2 := []int{2, 3, 4, 5, 6}

	for i := 0; i < b.N; i++ {
		s1 := OfSlice(slice1)
		s2 := OfSlice(slice2)
		seq := oldZip(s1, s2)
		seq(func(v Zipped[int, int]) bool {
			return true
		})
	}
}

func oldZipSend[T any](done <-chan struct{}, c chan<- T) func(v T) bool {
	return func(v T) bool {
		select {
		case <-done:
			return false
		case c <- v:
			return true
		}
	}
}

func oldZip[T1, T2 any](seq1 Seq[T1], seq2 Seq[T2]) Seq[Zipped[T1, T2]] {
	done := make(chan struct{})

	c1 := make(chan T1)
	go func() {
		defer close(c1)
		seq1(oldZipSend(done, c1))
	}()

	c2 := make(chan T2)
	go func() {
		defer close(c2)
		seq2(oldZipSend(done, c2))
	}()

	return func(yield func(Zipped[T1, T2]) bool) {
		defer close(done)

		var c1hold chan T1
		var c2hold chan T2
		var val Zipped[T1, T2]

		for {
			select {
			case v, ok := <-c1:
				if !ok {
					c1 = nil
					goto send
				}
				c1, c1hold = nil, c1
				val.V1 = v
				val.OK1 = true
			case v, ok := <-c2:
				if !ok {
					c2 = nil
					goto send
				}
				c2, c2hold = nil, c2
				val.V2 = v
				val.OK2 = true
			}

		send: // Dear lord, what is even going on here?
			if (val.OK1 || (c1 == c1hold)) && (val.OK2 || (c2 == c2hold)) && !(c1 == c1hold && c2 == c2hold) {
				if !yield(val) {
					return
				}
				c1, c1hold, c2, c2hold = c1hold, nil, c2hold, nil
				val = Zipped[T1, T2]{}
			}

			if (c1 == c1hold) && (c2 == c2hold) {
				return
			}
		}
	}
}
