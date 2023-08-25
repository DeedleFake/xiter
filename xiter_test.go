package xiter

import (
	"bytes"
	"cmp"
	"slices"
	"testing"
)

func TestZip(t *testing.T) {
	s1 := Slice([]int{1, 2, 3, 4, 5})
	s2 := Slice([]int{2, 3, 4, 5, 6})
	seq := Zip(s1, s2)
	seq(func(v Zipped[int, int]) bool {
		if v.V2-v.V1 != 1 {
			t.Fatalf("unexpected values: %+v", v)
		}
		return true
	})
}

func BenchmarkZip(b *testing.B) {
	slice1 := []int{1, 2, 3, 4, 5}
	slice2 := []int{2, 3, 4, 5, 6}

	for i := 0; i < b.N; i++ {
		s1 := Slice(slice1)
		s2 := Slice(slice2)
		seq := Zip(s1, s2)
		seq(func(v Zipped[int, int]) bool {
			return true
		})
	}
}

func TestIsSorted(t *testing.T) {
	if IsSorted(Slice([]int{1, 2, 3, 2})) {
		t.Fatal("is not sorted")
	}
	if !IsSorted(Slice([]int{1, 2, 3, 4, 5})) {
		t.Fatal("is sorted")
	}
	if !IsSorted(Slice([]int{48, 48})) {
		t.Fatal("is sorted")
	}
}

func TestMerge(t *testing.T) {
	s1 := Slice([]int{2, 3, 5})
	s2 := Slice([]int{1, 2, 3, 4, 5})
	r := Collect(Merge(s1, s2))
	if [8]int(r) != [...]int{1, 2, 2, 3, 3, 4, 5, 5} {
		t.Fatal(r)
	}
}

func splitmerge[T cmp.Ordered](s []T) Seq[T] {
	if len(s) <= 1 {
		return Slice(s)
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
		if !Equal(Slice(s), Slice(check)) {
			t.Fatal(s)
		}
	})
}
