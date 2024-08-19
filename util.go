package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"unicode"
)

const QueuedSymbol = "[Q]"

func intRange(n int) []int {
	ints := []int{}
	for i := range n {
		ints = append(ints, i)
	}
	return ints
}

// base should always be a valid absolute path
//
// returns basenames of immediate children
func descend(base string) ([]string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return []string{}, err
		// panic(err)
	}
	ch := []string{}
	for _, e := range entries {
		ch = append(ch, e.Name())
	}
	return ch, nil
}

// pretty-print arbitrary http (json) response without needing to know its
// schema
//
// warning: resp will be closed
func debugResponse(resp *http.Response) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var data map[string]any
	json.Unmarshal(body, &data)
	x, _ := json.MarshalIndent(data, "", "    ")
	fmt.Println(string(x))
}

// hacky function that uses generics (v1.18) to deserialize a http.Response
// into an arbitrary target type T. simply pass a zeroed t
func deserialize[T any](resp *http.Response, t T) T {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var data T
	json.Unmarshal(body, &data)
	return data
}

func alnum(s string) string {
	var out []rune
	for _, c := range s {
		if unicode.IsLetter(c) || unicode.IsNumber(c) || c == ' ' {
			out = append(out, c)
		}
	}
	return string(out)
}

// If target is not found in the dereferenced slice, the slice is unchanged.
// This is not an in-place operation (though I want it to be).
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
