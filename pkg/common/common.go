package common

import (
	"sort"
	"strings"
)

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Unique takes a slice of strings and returns a new slice with unique values, sorted in ascending order.
func Unique(input []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, val := range input {
		if _, ok := seen[val]; !ok {
			seen[val] = struct{}{}
			result = append(result, val)
		}
	}
	sort.Strings(result)
	return result
}
func UniqueLowercase(input []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, val := range input {
		lowercase := strings.ToLower(val)
		if _, ok := seen[lowercase]; !ok {
			seen[lowercase] = struct{}{}
			result = append(result, lowercase)
		}
	}
	sort.Strings(result)
	return result
}
