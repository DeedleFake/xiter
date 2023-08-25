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
