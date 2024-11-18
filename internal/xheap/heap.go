package xheap

import "cmp"

func Init[T cmp.Ordered](h []T) {
	n := len(h)
	for i := n/2 - 1; i >= 0; i-- {
		down(h, i, n)
	}
}

func InitFunc[T any](h []T, compare func(T, T) int) {
	n := len(h)
	for i := n/2 - 1; i >= 0; i-- {
		downfunc(h, i, n, compare)
	}
}

//func Push[T cmp.Ordered](h []T, x T) []T {
//	h = append(h, x)
//	up(h, len(h)-1)
//	return h
//}

func Pop[T cmp.Ordered](h []T) (T, []T) {
	n := len(h) - 1
	swap(h, 0, n)
	down(h, 0, n)
	return h[n], h[:n]
}

func PopFunc[T any](h []T, compare func(T, T) int) (T, []T) {
	n := len(h) - 1
	swap(h, 0, n)
	downfunc(h, 0, n, compare)
	return h[n], h[:n]
}

//func Remove(h []T, i int) T {
//	n := len(h) - 1
//	if n != i {
//		h.Swap(i, n)
//		if !down(h, i, n) {
//			up(h, i)
//		}
//	}
//	return h.Pop()
//}

//func Fix(h []T, i int) {
//	if !down(h, i, len(h)) {
//		up(h, i)
//	}
//}

//func up[T cmp.Ordered](h []T, j int) {
//	for {
//		i := (j - 1) / 2 // parent
//		if i == j || h[j] >= h[i] {
//			break
//		}
//		swap(h, i, j)
//		j = i
//	}
//}

func down[T cmp.Ordered](h []T, i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && h[j2] < h[j1] {
			j = j2 // = 2*i + 2  // right child
		}
		if h[j] >= h[i] {
			break
		}
		swap(h, i, j)
		i = j
	}
	return i > i0
}

func downfunc[T any](h []T, i0, n int, compare func(T, T) int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && compare(h[j2], h[j1]) < 0 {
			j = j2 // = 2*i + 2  // right child
		}
		if compare(h[j], h[i]) >= 0 {
			break
		}
		swap(h, i, j)
		i = j
	}
	return i > i0
}

func swap[T any](s []T, i1, i2 int) {
	s[i1], s[i2] = s[i2], s[i1]
}
