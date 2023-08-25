package xiter

import "testing"

func TestFind(t *testing.T) {
	s, _ := Find(Windows(Generate(
		0, 1),
		3),
		func(win []int) bool { return Sum(Slice(win)) >= 100 })
	if [3]int(s) != [...]int{33, 34, 35} {
		t.Fatal(s)
	}
}

func TestSum(t *testing.T) {
	s := Sum(Slice([]string{"a", " ", "test"}))
	if s != "a test" {
		t.Fatal(s)
	}
}
