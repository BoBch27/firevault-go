package firevault

import (
	"slices"
	"strconv"
	"time"
)

// asInt returns the parameter as a int64, or panics if it can't convert
func asInt(param string) int64 {
	i, err := strconv.ParseInt(param, 0, 64)
	if err != nil {
		panic(err)
	}

	return i
}

// asUint returns the parameter as a uint64, or panics if it can't convert
func asUint(param string) uint64 {
	u, err := strconv.ParseUint(param, 0, 64)
	if err != nil {
		panic(err)
	}

	return u
}

// asFloat returns the parameter as a float64, or panics if it can't convert
func asFloat(param string) float64 {
	f, err := strconv.ParseFloat(param, 32)
	if err != nil {
		panic(err)
	}

	return f
}

// asTime returns the parameter as a time.Time, or panics if it can't convert
func asTime(param string) time.Time {
	t, err := time.Parse(param, param)
	if err != nil {
		panic(err)
	}

	return t
}

// delSliceItem deletes an item from a slice
func delSliceItem[T comparable](slice []T, item T) []T {
	index := slices.Index(slice, item)
	if index == -1 {
		return slice
	}

	return slices.Delete(slice, index, index+1)
}
