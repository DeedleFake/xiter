package xiter

import "testing"

func BenchmarkZip2Pull(b *testing.B) {
	b.ReportAllocs()
	slice1 := []int{1, 2, 3, 4, 5}
	slice2 := []int{2, 3, 4, 5, 6}

	for i := 0; i < b.N; i++ {
		s1 := OfSlice(slice1)
		s2 := OfSlice(slice2)
		seq := zip2Pull(s1, s2)
		seq(func(v Zipped[int, int]) bool {
			return true
		})
	}
}

// zip2Pull returns a new Seq that yields the values of seq1 and seq2
// simultaneously.  This is the straightforward 2-Pull version.
func zip2Pull[T1, T2 any](seq1 Seq[T1], seq2 Seq[T2]) Seq[Zipped[T1, T2]] {
	return func(yield func(Zipped[T1, T2]) bool) {
		p1, stop := Pull(seq1)
		defer stop()
		p2, stop := Pull(seq2)
		defer stop()

		for {
			var val Zipped[T1, T2]
			val.V1, val.OK1 = p1()
			val.V2, val.OK2 = p2()
			if (!val.OK1 && !val.OK2) || !yield(val) {
				return
			}
		}
	}
}
