package xiter

import (
	"context"
	"slices"
	"testing"
)

func TestFind(t *testing.T) {
	s, _ := _Find(_Windows(_Generate(
		0, 1),
		3),
		func(win []int) bool { return _Sum(_OfSlice(win)) >= 100 })
	if [3]int(s) != [...]int{33, 34, 35} {
		t.Fatal(s)
	}
}

func TestContains(t *testing.T) {
	c := _Contains(_Of(1, 2, 3), 2)
	if !c {
		t.Fatal(c)
	}
}

func TestSum(t *testing.T) {
	s := _Sum(_OfSlice([]string{"a", " ", "test"}))
	if s != "a test" {
		t.Fatal(s)
	}
}

func TestProduct(t *testing.T) {
	p := _Product(_Of(3, 2, -5))
	if p != -30 {
		t.Fatal(p)
	}
}

func TestPartition(t *testing.T) {
	s1, s2 := _Partition(_Of(1, 2, 3, 4, 5), func(v int) bool { return v%2 == 0 })
	if !_Equal(_OfSlice(s1), _Of(2, 4)) {
		t.Fatal(s1)
	}
	if !_Equal(_OfSlice(s2), _Of(1, 3, 5)) {
		t.Fatal(s2)
	}
}

func TestExtent(t *testing.T) {
	s := _Of(3, 2, 5, 1, 6, -2, 10)
	min := _Min(s)
	max := _Max(s)
	if min != -2 {
		t.Fatal(min)
	}
	if max != 10 {
		t.Fatal(max)
	}
}

func TestAny(t *testing.T) {
	r := _Any(_Of(2, 4, 6, 7), func(v int) bool { return v%2 != 0 })
	if !r {
		t.Fatal(r)
	}
}

func TestAll(t *testing.T) {
	r := _All(_Of(2, 4, 6, 7), func(v int) bool { return v%2 == 0 })
	if r {
		t.Fatal(r)
	}
}

func TestSendContext(t *testing.T) {
	c := make(chan int, 3)
	_SendContext(_Of(3, 2, 5), context.Background(), c)
	s := []int{<-c, <-c, <-c}
	select {
	case v := <-c:
		t.Fatal(v)
	default:
	}
	if !slices.Equal(s, []int{3, 2, 5}) {
		t.Fatal(s)
	}
}

func FuzzSendRecvContext(f *testing.F) {
	f.Add([]byte("The quick brown fox jumps over the lazy dog."))

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		c := make(chan byte, len(data))
		_SendContext(_OfSlice(data), ctx, c)
		close(c)
		s := _Collect(_RecvContext(ctx, c))
		if !slices.Equal(data, s) {
			t.Fatal(s)
		}
	})
}
