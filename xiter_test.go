package xiter

import "testing"

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

func TestMerge(t *testing.T) {
	s1 := Slice([]int{2, 3, 5})
	s2 := Slice([]int{1, 2, 3, 4, 5})
	r := Collect(Merge(s1, s2))
	if [8]int(r) != [...]int{1, 2, 2, 3, 3, 4, 5, 5} {
		t.Fatalf("unexpected result: %v", r)
	}
}
