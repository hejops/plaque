//

package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtil(t *testing.T) {
	strData := map[string]any{"1": "1"}
	intData := map[string]any{"1": 1}
	strBytes, _ := json.Marshal(strData)
	intBytes, _ := json.Marshal(intData)
	assert.NotEqual(t, strBytes, intBytes)

	a, b := movePerfsToArtist("a", "b [c]")
	assert.Equal(t, a, "a c")
	assert.Equal(t, b, "b")

	assert.Equal(
		t,
		alnum("foo'bar a(b)c d@e"),
		"foo bar a b c d e",
	)

	assert.Equal(
		t,
		alnum("András"),
		"Andras",
	)

	assert.Equal(
		t,
		alnum("Neworldisorder ⁽⁽⁽ᵗ·ʷ·ᵒ·ˢ· ᵛ-₆₆₆₎₎₎"),
		"Neworldisorder             ",
	)

	ints := []int{1, 2, 3, 4, 5}
	assert.Equal(
		t,
		*remove(&ints, 3),
		[]int{1, 2, 5, 4},
	)

	albums := []string{"a (1990)", "b (1989)", "c (1988)"}
	sortByYear(albums)
	assert.Equal(t, albums, []string{"c (1988)", "b (1989)", "a (1990)"})

	t.Run("mock", func(t *testing.T) {
		// setup mock dirs + cfg
		// invoke entry point (queue)
		// TODO: simulate keypresses?
	})

	// assert.Equal(
	// 	t,
	// 	&ints,
	// 	&removed,
	// )

	assert.Equal(t, surround(0, 5, 2), []int{0, 1, 5, 2, 4})
	assert.Equal(t, surround(0, 6, 2), []int{0, 1, 6, 2, 5})
	assert.Equal(t, surround(1, 6, 2), []int{1, 2, 0, 3, 6})
	assert.Equal(t, surround(6, 6, 2), []int{6, 0, 5, 1, 4})
}

func BenchmarkAlnum(b *testing.B) {
	// 564271
	// c := 'x'
	for i := 0; i < b.N; i++ {
		alnum("foo'bar a(b)c d@e")

		// if c <= unicode.MaxASCII {
		// 	if _, ok := alnumChars[c]; ok {
		// 	}
		// }
	}
}

// func BenchmarkAlnum2(b *testing.B) {
// 	c := 'x'
// 	for i := 0; i < b.N; i++ {
// 		if c <= unicode.MaxASCII && unicode.In(c, alnumChars2...) {
// 		}
// 	}
// }

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
