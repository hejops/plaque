// String searching algorithms

package main

import (
	"log"
	"slices"
	"strings"
)

func makeBigrams(items []string) map[string][]int {
	chars := "abcdefghijklmnoprstuvwxyz "
	big := make(map[string][]int)
	for _, a := range chars {
		for _, b := range chars {
			bi := string(a) + string(b)
			if _, done := Bigrams[bi]; !done {
				big[bi] = searchSubstring(items, bi)
			}
		}
	}
	log.Println(len(big))
	return big
}

// TODO: if b.input has ' ', strings.Fields and match each word (b.items ->
// []map[string]nil?)
// might as well impl Aho-Corasick or trigram at that point

// Given a slice of items, return a slice of indices of each item that contains
// the target word. Normalisation is applied.
//
// Uses default Rabin-Karp algorithm for each string search
func searchSubstring(items []string, target string) []int {
	// fmt.Println(items[0])
	if target == "" {
		return intRange(len(items))
	}
	targetLower := strings.ToLower(target)
	matchIdxs := make([]int, len(items))
	var i int
	for j, rel := range items {
		if strings.Contains(strings.ToLower(rel), targetLower) {
			matchIdxs[i] = j
			i++
		}
	}
	// slices.Clip(matchIdxs)
	return matchIdxs[:i]
}

func searchSubstringCache(items []string, target string, inputCache map[string][]int) []int { // {{{
	if matches, ok := inputCache[target]; ok {
		return matches
	} else {
		matches = searchSubstring(items, target)
		inputCache[target] = matches
		return matches
	}
} // }}}

// 8x speedup, after constructing all the 676*len(items) bigrams (which takes
// about 1.5 s for 37 k items). Note that searches will be inherently fuzzy.
func searchSubstringBigram(items []string, target string) []int {
	if len(target) < 2 {
		return searchSubstring(items, target)
	}

	first := target[:2]
	idxs := Bigrams[first]

	// fmt.Println(first)
	for i := 1; i < len(target)-1; i++ {
		bi := target[i : i+2]
		// fmt.Println(bi)
		found := Bigrams[bi]
		if len(found) == 0 {
			return []int{}
		}
		idxs = intersect(idxs, found)
	}
	slices.Sort(idxs)
	return idxs
}
