package discogs

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var alnumChars = make(map[rune]any)

func init() {
	for _, c := range "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890 " {
		alnumChars[c] = nil
	}
}

func ensure(c bool) {
	if !c {
		log.Fatalln("assertion failed")
	}
}

// hacky function that uses generics (v1.18) to deserialize a http.Response
// into an arbitrary target type T, without any error handling whatsoever
func deserialize[T any](resp *http.Response, _ T) (data T) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	_ = json.Unmarshal(body, &data) // errors are ignored!
	return data
}

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
