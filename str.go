// String searching algorithms

package main

import (
	"slices"
	"strings"
	"sync"
)

var (
	Bigrams    map[string][]int
	bigramOnce sync.Once
)

func makeBigrams(items []string) map[string][]int {
	chars := "abcdefghijklmnoprstuvwxyz "
	big := make(map[string][]int)
	for _, a := range chars {
		for _, b := range chars {
			bi := string(a) + string(b)
			if _, done := big[bi]; !done {
				big[bi] = searchSubstring(items, bi)
			}
		}
	}
	// log.Println(len(big))
	return big
}

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

// In real world usage, this is an 8x speedup over searchSubstring.
//
// The caveat: this function relies on the global Bigrams map, which is
// relatively expensive (676 bigrams * 37 k items = 1.5 s), and calculated only
// once for the lifetime of the program.

// Note that because searches will be inherently fuzzy, false positives may be
// returned.
func searchSubstringBigram(items []string, target string) []int {
	// if we are generating the bigrams here, it is already too late; user
	// input will usually be faster than 1.5s
	//
	// bigramOnce.Do(func() {
	// 	go func() {
	// 		t := time.Now()
	// 		Bigrams = makeBigrams(items)
	// 		log.Println("bigram construction took", time.Since(t).Seconds())
	// 	}()
	// })

	if len(target) < 2 {
		return searchSubstring(items, target)
	} else if strings.Contains(target, ".") {
		// this might look like a really crappy impl, but it doesn't
		// feel that slow
		r := regexp.MustCompile("(?i)" + target)
		matches := []int{}
		for i, x := range items {
			if r.Match([]byte(x)) {
				matches = append(matches, i)
			}
		}
		return matches
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
