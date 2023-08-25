package xiter

import (
	"cmp"
	"slices"
	"testing"
)

func TestRunes(t *testing.T) {
	s := Collect(Runes("これはテストです。"))
	if [9]rune(s) != [9]rune([]rune("これはテストです。")) {
		t.Fatal(s)
	}
}

func TestMapEntries(t *testing.T) {
	s := Collect(MapEntries(map[string]string{"this": "is", "a": "test"}))
	slices.SortFunc(s, func(e1, e2 MapEntry[string, string]) int { return cmp.Compare(e1.Key, e2.Key) })
	if [2]MapEntry[string, string](s) != [...]MapEntry[string, string]{{"a", "test"}, {"this", "is"}} {
		t.Fatal(s)
	}
}
