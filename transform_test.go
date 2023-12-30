package xiter

import (
	"bytes"
	"cmp"
	"slices"
	"testing"
)

func TestMap(t *testing.T) {
	s := OfSlice([]int{1, 2, 3})
	n := Collect(Map(s, func(v int) float64 { return float64(v * 2) }))
	if [3]float64(n) != [...]float64{2, 4, 6} {
		t.Fatal(n)
	}
}

func TestFilter(t *testing.T) {
	s := OfSlice([]int{1, 2, 3})
	n := Collect(Filter(s, func(v int) bool { return v%2 != 0 }))
	if [2]int(n) != [...]int{1, 3} {
		t.Fatal(n)
	}
}

func TestSkip(t *testing.T) {
	s := Collect(Skip(Limit(Generate(
		0, 1),
		3),
		2),
	)
	if !Equal(OfSlice(s), Of(2)) {
		t.Fatal(s)
	}
}

func TestLimit(t *testing.T) {
	s := Collect(Limit(Generate(
		0, 2),
		3),
	)
	if [3]int(s) != [...]int{0, 2, 4} {
		t.Fatal(s)
	}
}

func TestConcat(t *testing.T) {
	s := Collect(Concat(OfSlice([]int{1, 2, 3}), OfSlice([]int{3, 2, 5})))
	if [6]int(s) != [...]int{1, 2, 3, 3, 2, 5} {
		t.Fatal(s)
	}
}

func TestZip(t *testing.T) {
	s1 := OfSlice([]int{1, 2, 3, 4, 5})
	s2 := OfSlice([]int{2, 3, 4, 5, 6})
	seq := Zip(s1, s2)
	seq(func(v Zipped[int, int]) bool {
		if v.V2-v.V1 != 1 {
			t.Fatalf("unexpected values: %+v", v)
		}
		return true
	})
}

func TestZipShort1(t *testing.T) {
	s1 := OfSlice([]int{1, 2, 3, 4})
	s2 := OfSlice([]int{2, 3, 4, 5, 1})
	seq := Zip(s1, s2)
	seq(func(v Zipped[int, int]) bool {
		if v.V2-v.V1 != 1 {
			t.Fatalf("unexpected values: %+v", v)
		}
		return true
	})
	t.Logf("Greetings from TestZipShort1")
}

func TestZipShort2(t *testing.T) {
	s1 := OfSlice([]int{1, 2, 3, 4, -1, -2})
	s2 := OfSlice([]int{2, 3, 4, 5})
	seq := Zip(s1, s2)
	seq(func(v Zipped[int, int]) bool {
		if v.V1 == -2 {
			return false
		}
		if v.V2-v.V1 != 1 {
			t.Fatalf("unexpected values: %+v", v)
		}
		return true
	})
	t.Logf("Greetings from TestZipShort2")
}

func BenchmarkZip(b *testing.B) {
	slice1 := []int{1, 2, 3, 4, 5}
	slice2 := []int{2, 3, 4, 5, 6}

	for i := 0; i < b.N; i++ {
		s1 := OfSlice(slice1)
		s2 := OfSlice(slice2)
		seq := Zip(s1, s2)
		seq(func(v Zipped[int, int]) bool {
			return true
		})
	}
}

func TestIsSorted(t *testing.T) {
	if IsSorted(OfSlice([]int{1, 2, 3, 2})) {
		t.Fatal("is not sorted")
	}
	if !IsSorted(OfSlice([]int{1, 2, 3, 4, 5})) {
		t.Fatal("is sorted")
	}
	if !IsSorted(OfSlice([]int{48, 48})) {
		t.Fatal("is sorted")
	}
}

func TestMerge(t *testing.T) {
	s1 := OfSlice([]int{2, 3, 5})
	s2 := OfSlice([]int{1, 2, 3, 4, 5})
	r := Collect(Merge(s1, s2))
	if [8]int(r) != [...]int{1, 2, 2, 3, 3, 4, 5, 5} {
		t.Fatal(r)
	}
}

func splitmerge[T cmp.Ordered](s []T) Seq[T] {
	if len(s) <= 1 {
		return OfSlice(s)
	}

	left := splitmerge(s[:len(s)/2])
	right := splitmerge(s[len(s)/2:])
	return Merge(left, right)
}

func mergesort[T cmp.Ordered](s []T) {
	AppendTo(splitmerge(s), s[:0])
}

func TestMergeSort(t *testing.T) {
	s := []int{3, 2, 5, 1, 6, 2}
	mergesort(s)
	if [6]int(s) != [...]int{1, 2, 2, 3, 5, 6} {
		t.Fatal(s)
	}
}

func FuzzMergeSort(f *testing.F) {
	f.Add([]byte("The quick brown fox jumped over the lazy dog."))
	f.Fuzz(func(t *testing.T, s []byte) {
		check := bytes.Clone(s)
		slices.Sort(check)

		mergesort(s)
		if !Equal(OfSlice(s), OfSlice(check)) {
			t.Fatal(s)
		}
	})
}

func TestChunks(t *testing.T) {
	s := Collect(Map(Chunks(OfSlice([]int{1, 2, 3, 4, 5}),
		2),
		slices.Clone),
	)
	if !slices.EqualFunc(s, [][]int{{1, 2}, {3, 4}, {5}}, slices.Equal) {
		t.Fatal(s)
	}
}

func TestSplit2(t *testing.T) {
	s1, s2 := CollectSplit(Split2(FromPair(OfSlice([]Pair[int32, int64]{{1, 2}, {3, 4}, {5, 6}}))))
	if !slices.Equal(s1, []int32{1, 3, 5}) {
		t.Fatal(s1)
	}
	if !slices.Equal(s2, []int64{2, 4, 6}) {
		t.Fatal(s2)
	}
}

func TestCache(t *testing.T) {
	var i int
	f := func(yield func(int) bool) {
		yield(i)
		i++
		return
	}
	seq := Cache(f)
	if s := Collect(seq); !slices.Equal(s, []int{0}) {
		t.Fatal(s)
	}
	if s := Collect(seq); !slices.Equal(s, []int{0}) {
		t.Fatal(s)
	}
}

func TestEnumerate(t *testing.T) {
	s := Collect(ToPair(Enumerate(Limit(Generate(0, 2), 3))))
	if !slices.Equal(s, []Pair[int, int]{{0, 0}, {1, 2}, {2, 4}}) {
		t.Fatal(s)
	}
}

func TestOr(t *testing.T) {
	s := Collect(Or(Of[int](), nil, Of(1, 2, 3), Of(4, 5, 6)))
	if !slices.Equal(s, []int{1, 2, 3}) {
		t.Fatal(s)
	}
}
