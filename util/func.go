// Package util provides small utilities used through the rest
// of the project
package util

// Min returns the minimum of two ints
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two ints
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
