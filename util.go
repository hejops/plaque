package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const QueuedSymbol = "[Q]"

func intRange(n int) []int {
	ints := []int{}
	for i := range n {
		ints = append(ints, i)
	}
	return ints
}

// Select n random items from the queue file (containing relpaths), and return
// them as fullpaths
func getQueue(n int) []string {
	b, err := os.ReadFile(config.Library.Queue)
	if err != nil {
		panic(err)
	}
	relpaths := strings.Split(string(b), "\n")
	// TODO: split off sampling
	// TODO: use rand.Shuffle instead?
	// https://stackoverflow.com/a/12267471
	for i := range relpaths {
		j := rand.Intn(i + 1)
		relpaths[i], relpaths[j] = relpaths[j], relpaths[i]
	}
	paths := []string{}
	root := config.Library.Root
	if n < 0 {
		panic("not impl yet")
	}
	for _, rel := range relpaths[:n] {
		p := filepath.Join(root, rel)
		paths = append(paths, p)
	}
	return paths
}

// base should always be a valid absolute path
//
// returns fullpaths of immediate children if join is true (otherwise basenames)
func descend(base string, join bool) ([]string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return []string{}, err
		// panic(err)
	}
	ch := []string{}
	for _, e := range entries {
		if join {
			ch = append(ch, filepath.Join(base, e.Name()))
		} else {
			ch = append(ch, e.Name())
		}
	}
	return ch, nil
}

// pretty-print arbitrary http (json) response without needing to know its
// schema
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
