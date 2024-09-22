package discogs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtil(t *testing.T) {
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
