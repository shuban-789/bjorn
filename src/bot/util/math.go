package util

import "cmp"

// Clamp restricts a value to be within the specified minimum and maximum bounds.
// cmp.Ordered I saw was used in the definition of min and max so I used it here
func Clamp[T cmp.Ordered](value, minVal, maxVal T) T {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}