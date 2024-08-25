package xiter

import "testing"

func TestCoroutine(t *testing.T) {
	yield, stop := Coroutine(func(first int, yield func(int) (int, bool)) {
		if first != 0 {
			t.Fatal(first)
		}

		prev := first
		for {
			v, ok := yield(prev + 1)
			if !ok {
				if prev != 10 {
					t.Fatal(v)
				}
				return
			}
			if v != prev+2 {
				t.Fatal(v)
			}
			prev = v
		}
	})
	defer stop()

	prev, ok := yield(0)
	if !ok {
		t.Fatal(prev)
	}
	for prev < 10 {
		v, ok := yield(prev + 1)
		if !ok {
			t.Fatal(v)
		}
		if v != prev+2 {
			t.Fatal(v)
		}
		prev = v
	}
}
