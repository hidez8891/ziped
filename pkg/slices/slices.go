package slices

func FindIndex[T comparable](xs []T, value T, offset int) int {
	for i, x := range xs[offset:] {
		if x == value {
			return i + offset
		}
	}
	return -1
}

func Split[T comparable](xs []T, value T) [][]T {
	out := make([][]T, 0)

	first := 0
	end := 0
	for {
		if end >= len(xs) {
			break
		}

		end = FindIndex(xs, value, first)
		if end < 0 {
			end = len(xs)
		}

		out = append(out, xs[first:end])
		first = end + 1 // skip delimiter value
	}

	return out
}
