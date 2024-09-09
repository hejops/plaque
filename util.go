package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func checkRelPaths(items []string) {
	// stop checking as soon as one entry is valid. errors are only logged;
	// no action is taken
	for _, item := range items {
		info, err := os.Stat(filepath.Join(config.Library.Root, item))
		if err != nil {
			log.Println("not exist:", item)
			continue
		}
		if !info.IsDir() {
			log.Println("not dir:", item)
			continue
		}
		break
	}
}

func timer(name string) func() {
	// https://stackoverflow.com/a/45766707
	start := time.Now() // at time of defer
	return func() {
		dur := time.Since(start) // at end of func scope
		if dur >= time.Second {
			log.Printf(
				"%s took %v\n",
				name,
				dur,
			)
		}
	}
}

func ensure(c bool) {
	if !c {
		log.Fatalln("assertion failed")
	}
}

func intRange(n int) []int {
	ints := make([]int, n)
	for i := range n {
		ints[i] = i
	}
	return ints
}

// base should always be a valid absolute path
//
// returns basenames of immediate children
func descend(base string) ([]string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		// TODO: remove from queue
		// log.Println("not a valid dir:", base)
		return []string{}, err
	}
	ch := []string{}
	for _, e := range entries {
		ch = append(ch, e.Name())
	}
	return ch, nil
}

// Pretty-print arbitrary http (json) response without needing to know its
// schema
//
// Warning: resp will be closed
func debugResponse(resp *http.Response) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		panic(err)
	}
	x, _ := json.MarshalIndent(data, "", "    ")
	fmt.Println(string(x))
}

var alnumChars = func() map[rune]any {
	chars := make(map[rune]any)
	for _, c := range "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890 " {
		chars[c] = nil
	}
	return chars
}()

func alnum(s string) string {
	// https://en.wikipedia.org/wiki/Unicode_equivalence#Normal_forms
	// https://www.unicode.org/glossary/#normalization_form_d
	// https://www.unicode.org/reports/tr15/#Norm_Forms
	// https://www.unicode.org/versions/Unicode15.1.0/ch03.pdf#G49537
	norm, _, err := transform.String(norm.NFD, s)
	if err != nil {
		panic(err.Error())
	}

	// // for benchmark only
	// alnumChars2 := []*unicode.RangeTable{
	// 	unicode.Letter,
	// 	unicode.Number,
	// 	unicode.Space,
	// }

	var out []rune
	for _, c := range norm {
		switch {
		// https://www.unicode.org/reports/tr44/#GC_Values_Table
		case unicode.IsPunct(c): //,unicode.IsSymbol(c):
			out = append(out, ' ')

		// 415 ns/op
		case c <= unicode.MaxASCII:
			if _, ok := alnumChars[c]; ok {
				out = append(out, c)
			}

			// // 425 ns/op
			// case c <= unicode.MaxASCII && unicode.In(c, alnumChars2...):
			// 	out = append(out, c)

		}
	}
	return string(out)
}

// Given a slice of items, return a slice of indices of each item that contains
// the target word
func fuzzySearch(items []string, target string) (matchIdxs []int) {
	for i, rel := range items {
		// TODO: if b.input has ' ', strings.Fields and match each word
		// (b.items -> []map[string]nil?)
		if strings.Contains(strings.ToLower(rel), strings.ToLower(target)) {
			matchIdxs = append(matchIdxs, i)
		}
	}
	return matchIdxs
}

// "a", "b [c]"
// "a c", "b"
func movePerfsToArtist(artist string, album string) (string, string) {
	// TODO:
	// strings.LastIndex(album, " [")
	x := strings.SplitN(album, " [", 2) // "b", "c]"
	perfs := strings.TrimSuffix(x[1], "]")
	return artist + " " + perfs, x[0]
}

// Sort a slice of albums by year suffix (" (YYYY)"). Sorting is performed
// inplace.
func sortByYear(albums []string) {
	slices.SortFunc(albums, func(a string, b string) int {
		if a[len(a)-1] != ')' {
			return -1
		}
		if b[len(b)-1] != ')' {
			return 1
		}

		ay := a[len(a)-5:]
		by := b[len(b)-5:]

		switch {
		case ay < by:
			return -1
		case ay > by:
			return 1
		default:
			return 0
		}
	})
}

// Given a middle number n, construct a slice whose first item is n, odd values
// are increments of n, and even values are decrements of n. The slice has len
// 2 * width + 1, and all values are constrained in the range [0, limit].
func surround(middle int, limit int, width int) (ints []int) {
	if width == 0 {
		return []int{middle}
	}

	ints = make([]int, 2*width+1)
	for idx := range ints {
		switch {
		case idx == 0:
			ints[0] = middle
		case (idx % 2) == 1:
			next := middle + (idx+1)/2
			if next > limit {
				next -= limit + 1
			}
			ints[idx] = next
		default:
			prev := middle - (idx+1)/2
			if prev < 0 {
				prev += limit + 1
			}
			ints[idx] = prev
		}
	}
	// TODO: constrain ends to start,end of eb.artists (i wish go had Option)
	return ints
}

// generics {{{

// hacky function that uses generics (v1.18) to deserialize a http.Response
// into an arbitrary target type T, without any error handling whatsoever
func deserialize[T any](resp *http.Response, t T) (data T) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	_ = json.Unmarshal(body, &data) // errors are ignored!
	return data
}

// If target is not found in the dereferenced slice, the slice is unchanged.
// This is not an in-place operation (due to language constraints?).
//
// Warning: for performance, order is not preserved!
func remove[T comparable](ptr *[]T, target T) *[]T {
	s := *ptr
	for i, item := range s {
		if item == target {
			// swap last item with target, to prevent costly
			// shifting of items
			// https://stackoverflow.com/a/37335777
			s[i] = s[len(s)-1]
			// re-assignment always reallocates, so a new ptr must
			// be returned
			// https://stackoverflow.com/a/56394697
			s = s[:len(s)-1]
			return &s
		}
	}
	return ptr
}

func Map[T any](items []T, f func(T) T) []T {
	var out []T
	for _, i := range items {
		out = append(out, f(i))
	}
	return out
}

func anyValue[T comparable](m map[T]bool) bool {
	if len(m) == 0 {
		return false
	}
	for _, v := range m {
		if v {
			return true
		}
	}
	return false
}

// }}}
