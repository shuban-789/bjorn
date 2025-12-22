// I just implemented some common string search utilities here for use in
// autocomplete handlers.
package util

import (
	"sort"
	"strings"
)

type TokenizedName interface {
	GetSearchTokens() []string
}

// Time compleity: O(n + m + z) where n is length of text, m is sum of lengths of patterns, z is number of matches
// I was doing KMP at first but then it becomes O(n + zm) for multiple patterns
func AhoCorasickSearch(patterns []string, text string) bool {
	// TODO: implement Aho-Corasick algorithm
	return true
}

// note: basically what I do here is just give a score to each region, then return the top n matches
// set maxResults to -1 to get all matches
//
// this only really works fast enough for autocomplete if the db is small (like regions) and each
// item only has a few tokens
func TokenizedSearch[T TokenizedName](db []T, query string, maxResults int) []T {
	// return all if empty query
	if query == "" {
		results := make([]T, 0)
		for i, item := range db {
			if maxResults != -1 && i >= maxResults {
				break
			}
			results = append(results, item)
		}
		return results
	}

	queryTokens := strings.Fields(normalizeName(query))

	type ScoredItem struct {
		item T
		score  int
	}
	scores := make([]ScoredItem, 0)
	
	// 3 for loops is icky but it's like 90 regions so it's chill
	for _, item := range db {
		score := 0
		for _, queryToken := range queryTokens {
			for _, regionToken := range item.GetSearchTokens() {
				// prioritize things starting w/ the query
				if strings.HasPrefix(regionToken, queryToken) {
					score += 2
				} else if strings.Contains(regionToken, queryToken) {
					score += 1
				}
			}
		}
		if score > 0 {
			scores = append(scores, ScoredItem{
				item: item,
				score:  score,
			})
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	results := make([]T, 0)
	for i, scoredItem := range scores {
		if maxResults != -1 && i >= maxResults {
			break
		}
		results = append(results, scoredItem.item)
	}
	return results
}

func GenerateNormalizedTokens(name string) []string {
	normalized := normalizeName(name)
	tokens := strings.Fields(normalized)
	return tokens
}

// remove dashes + convert to lowercase for easier searching
func normalizeName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, "-", " "))
}
