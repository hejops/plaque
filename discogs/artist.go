package discogs

import (
	"log"
	"net/url"
	"strconv"
	"strings"
)

type Artist struct {
	Id          int
	ResourceUrl string          `json:"resource_url"`
	UserData    map[string]bool `json:"user_data"` // in_collection

	// TODO: in search, json key is 'title', otherwise 'name' in all other
	// contexts. this is very footgun-y, so i need to do something about it

	Name  string // all other contexts
	Title string // search-only
}

// Returns artist releases (which are not full releases)
//
// Requires GET
func (a Artist) Releases() []Release {
	// /artists/{a.id}/releases
	urlpath, _ := url.JoinPath(
		"artists",
		strconv.Itoa(a.Id),
		"releases",
	)

	resp := makeReq(
		urlpath,
		"GET",
		// yes, the numbers need to be strings...
		map[string]any{
			"sort":     "year",
			"per_page": "100",
			"page":     "1",
		},
	)
	return deserialize(resp, struct {
		Releases []Release
	}{}).Releases
}

var IgnoredFormats = map[string]any{
	// maps are var-only

	"Compilation": nil,
	"DVD-V":       nil,
	"Shellac":     nil,
	"Single":      nil,
}

// Currently only supports artist releases. Note that this tends to produce
// false negatives, because the Formats field of an artist release tends to be
// empty, and we avoid a GET of the primary release for now.
func (r *Release) ignored() bool {
	log.Println(r.Formats)
	for _, format := range strings.Split(r.Formats, ", ") {
		// fmts := r.Formats[0]["descriptions"].([]string)
		// for _, format := range fmts {
		if _, ig := IgnoredFormats[format]; ig {
			return true
		}
	}
	// TODO: fetch actual release
	return false
}
