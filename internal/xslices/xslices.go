package xslices

func Map[From, To any](from []From, f func(From) To) []To {
	to := make([]To, 0, len(from))
	for _, v := range from {
		to = append(to, f(v))
	}
	return to
}
