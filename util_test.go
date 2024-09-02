package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtil(t *testing.T) {
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
	// _ = relpaths
}
