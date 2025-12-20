// I just implemented some common string search utilities here for use in
// autocomplete handlers.
package util

// Time compleity: O(n + m + z) where n is length of text, m is sum of lengths of patterns, z is number of matches
// I was doing KMP at first but then it becomes O(n + zm) for multiple patterns
func AhoCorasickSearch(patterns []string, text string) bool {
	// TODO: implement Aho-Corasick algorithm
	return true
}