package arrayutils

// Map applies the provided function fn to each element of the input slice in,
// returning a new slice containing the results.
//
// Example:
//
//	squares := Map([]int{1, 2, 3}, func(x int) int { return x * x })
//	// squares == []int{1, 4, 9}
func Map[T any, U any](in []T, fn func(T) U) []U {
	out := make([]U, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}
