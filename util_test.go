//

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
)

func TestUtil(t *testing.T) {
	a, b := movePerfsToArtist("a", "b [c]")
	assert.Equal(t, a, "a c")
	assert.Equal(t, b, "b")

	ints := []int{1, 2, 3, 4, 5}
	assert.Equal(
		t,
		*remove(&ints, 3),
		[]int{1, 2, 5, 4},
	)

	albums := []string{"a (1990)", "b (1989)", "c (1988)"}
	sortByYear(albums)
	assert.Equal(t, albums, []string{"c (1988)", "b (1989)", "a (1990)"})

	// assert.Equal(
	// 	t,
	// 	&ints,
	// 	&removed,
	// )

	strData := map[string]any{"1": "1"}
	intData := map[string]any{"1": 1}
	strBytes, _ := json.Marshal(strData)
	intBytes, _ := json.Marshal(intData)
	assert.NotEqual(t, strBytes, intBytes)

	assert.False(t, anyValue(map[int]bool{1: false, 2: false}))
	assert.False(t, anyValue(map[int]bool{1: false}))
	assert.False(t, anyValue(map[int]bool{}))
	assert.True(t, anyValue(map[int]bool{1: true, 2: false}))
	assert.True(t, anyValue(map[int]bool{1: true}))
}

func BenchmarkReadfile(b *testing.B) {
	// https://stackoverflow.com/a/16615559
	by, err := os.ReadFile(config.Library.Queue)
	if err != nil {
		log.Fatalln(err)
	}
	strings.Split(string(by), "\n")
}

func BenchmarkNewScanner(b *testing.B) {
	// https://stackoverflow.com/a/16615559
	var relpaths []string
	file, err := os.Open(config.Library.Queue)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		relpaths = append(relpaths, sc.Text())
	}
	if sc.Err() != nil {
		log.Fatalln(err)
	}
	_ = relpaths
}

var (
	randStrings = make([]string, 10000)
	sameStrings = make([]string, 10000)
	artists     []string
)

func init() {
	x := rand.New(rand.NewSource(1))
	_ = x

	// tl;dr: bigrams are exceptional at random inputs, but still great at
	// real world inputs

	// BenchmarkSubstringSearchRKLower-12      1000000000      0.0008277 ns/op
	// BenchmarkSubstringSearchBigram-12       1000000000      0.0000006 ns/op

	// for i := range len(randStrings) {
	// 	var s []rune
	// 	for range 10 {
	// 		s = append(s, rune(65+x.Intn(57)))
	// 	}
	// 	randStrings[i] = string(s)
	// }
	// Bigrams = makeBigrams(randStrings)

	// BenchmarkSubstringSearchRKLower-12      1000000000      0.002818 ns/op
	// BenchmarkSubstringSearchBigram-12       1000000000      0.0000051 ns/op

	var s []rune
	for range 100 {
		s = append(s, rune(65+x.Intn(57)))
	}
	for i := range len(sameStrings) {
		sameStrings[i] = string(s)
	}
	Bigrams = makeBigrams(sameStrings)

	// BenchmarkSubstringSearchRKLower-12      1000000000      0.002496 ns/op
	// BenchmarkSubstringSearchBigram-12       1000000000      0.0001683 ns/op

	// artists, _ = descend(config.Library.Root)
	// Bigrams = makeBigrams(artists)
}

func benchmarkWrapper(f func([]string, string) []int) {
	// f(randStrings, "ab")
	f(sameStrings, "jp")
	// f(artists, "johann")
}

func BenchmarkSubstringSearchRKLower(b *testing.B) { benchmarkWrapper(searchSubstring) }

func BenchmarkSubstringSearchBigram(b *testing.B) { benchmarkWrapper(searchSubstringBigram) }

func TestSubstring(t *testing.T) {
	fmt.Println(sameStrings[0])
	for _, x := range []struct {
		needle string
		count  int
		countB int
	}{
		{needle: "jp", count: 10000, countB: 10000},
		{needle: "xoba", count: 0, countB: 10000}, // fuzzy!
	} {
		assert.Len(t, searchSubstring(sameStrings, x.needle), x.count, x.needle)
		assert.Len(t, searchSubstringBigram(sameStrings, x.needle), x.countB, x.needle)
	}
}

// func BenchmarkWalk(b *testing.B) { generateAlloc(100) }

// 5.6 s cold, 2.5 s warm
func TestWalkAlloc(t *testing.T) { generateQueue(100) }
