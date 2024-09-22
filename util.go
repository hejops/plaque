package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

func ensure(c bool) {
	if !c {
		log.Fatalln("assertion failed")
	}
}

func checkDir(artist string, album string) bool {
	dirs, err := os.ReadDir(filepath.Join(config.Library.Root, artist))
	if err != nil {
		return false
	}
	for _, dir := range dirs {
		if strings.HasPrefix(dir.Name(), album) {
			fmt.Println(filepath.Join(artist, dir.Name()))
			return true
		}
	}
	return false
}

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
	// ch := []string{}
	ch := make([]string, len(entries))
	for i, e := range entries {
		ch[i] = e.Name()
	}
	return ch, nil
}

// Pretty-print arbitrary http (json) response without needing to know its
// schema
//
// Warning: resp will be closed
// func debugResponse(resp *http.Response) {
// 	defer resp.Body.Close()
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		panic(err)
// 	}
// 	var data map[string]any
// 	if err := json.Unmarshal(body, &data); err != nil {
// 		panic(err)
// 	}
// 	x, _ := json.MarshalIndent(data, "", "    ")
// 	fmt.Println(string(x))
// }

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

// generics {{{

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

func intersect[T comparable](a []T, b []T) []T {
	var inter []T
	for _, x := range a {
		if slices.Contains(b, x) {
			inter = append(inter, x)
		}
	}
	return inter
}

// }}}
