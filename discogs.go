package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const ApiPrefix = "https://api.discogs.com"

// urlpath -cannot- contain query params; these should be passed as data
// instead.
//
// data should either be GET query params (in which case all values must be
// strings), or PUT json data.
func discogsReq(
	urlpath string,
	method string,
	data map[string]any,
) *http.Response { // {{{

	// urlpath must be a string (not Url) to make it easy for callers.
	// however, because urlpaths that contain query will be joined with
	// undesirable escape ("?" -> "%3f"), queries have to be added
	// separately, and -after- joining (not before). to avoid needing a
	// fourth arg, `data` is forced to serve double duty (i.e. it is either
	// GET query params, or PUT json data).

	if urlpath == "" {
		panic("empty urlpath!")
	}

	u, _ := url.Parse(ApiPrefix)
	u = u.JoinPath(urlpath)

	// map -> []byte -> bytes.Buffer

	var req *http.Request
	var err error

	switch method {

	case "GET", "POST":
		if data != nil {
			query := url.Values{}
			for k, v := range data {
				query.Set(k, v.(string))
			}
			u.RawQuery = query.Encode()
		}
		req, err = http.NewRequest(method, u.String(), nil)

	case "PUT":
		b, mErr := json.Marshal(data)
		if mErr != nil {
			panic(mErr)
		}
		// https://stackoverflow.com/a/24455606
		req, err = http.NewRequest(method, u.String(), bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")

	default:
		panic("invalid method")

	}

	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", "Discogs token="+config.Discogs.Key)
	req.Header.Set("Cache-Control", "no-cache")

	log.Println(method, u.RequestURI())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	return resp
} // }}}

// remember: all fields must be uppercase, and any fields in camelcase must be
// marked in order to be parsed

// A general-purpose struct that is shared across several contexts: search
// results (which may or may not be master), artist releases, master releases,
// and 'actual' releases.
//
// This is because Discogs stores releases in a variety of representations with
// subtle differences in schema.
type Release struct { // {{{
	// The meaning of Id depends on ReleaseType; i.e. Id will be a master
	// id if ReleaseType = "master", or a 'regular' release id if
	// ReleaseType = "release"
	Id int

	// For search results, artist will be included in this field (i.e.
	// "artist - title"). For this reason, treating search results as
	// Release is discouraged.
	Title string

	Artists     []Artist
	ArtistsSort string `json:"artists_sort"`
	ResourceUrl string `json:"resource_url"`
	Year        int    // may be string (in search?)

	MasterId  int    `json:"master_id"`    // may be 0 (if no master)
	MasterUrl string `json:"master_url"`   // may be empty (if no master)
	Primary   int    `json:"main_release"` // master-only

	// Formats []map[string]any // full release

	// search-only (?)

	Community   map[string]int
	Genre       []string
	ReleaseType string `json:"type"` // search-only ('release' or 'master')

	// artist-only

	Artist  string // artist-only
	Label   string
	Role    string                    // typically "Main"
	Stats   map[string]map[string]int // 4 keys: "community"/"stats" -> "in_collection"/"in_wantlist"
	Formats string                    // ", "-delimited, may be empty
} // }}}

func (r *Release) inCollection() bool {
	return r.Stats["user"]["in_collection"] > 0
}

// requires r.Id (callers should override r.Id with r.Primary for now)
//
// does nothing if release already rated
func (r *Release) rate() int { // {{{
	noopInt := -1
	if r.Id == 0 {
		return noopInt
	}

	// TODO: leaky abstraction that should be refactored out
	switch {
	case r.Primary > 0: // master release
		r = deserialize(
			discogsReq("/releases/"+strconv.Itoa(r.Primary), "GET", nil),
			&Release{},
		)
	case r.Artist != "": // artist release
		r = deserialize(
			discogsReq("/releases/"+strconv.Itoa(r.Id), "GET", nil),
			&Release{},
		)
	}

	// releases/{r.Id}/rating/{username}
	urlpath, _ := url.JoinPath(
		"releases",
		strconv.Itoa(r.Id),
		"rating",
		config.Discogs.Username,
	)

	resp := discogsReq(urlpath, "GET", nil)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var currentRating map[string]any
	// an error here usually means incorrect was Id supplied (i.e. master
	// id instead of release id)
	if err := json.Unmarshal(body, &currentRating); err != nil {
		return noopInt
	}
	if int(currentRating["rating"].(float64)) != 0 {
		log.Println("already rated:", r.Id, r.Title, currentRating)
		return noopInt
	}

	fmt.Println(r.Year, "::", r.Artists[0].Name, "::", r.Title)
	fmt.Printf("https://www.discogs.com/release/%d\n", r.Id)
	fmt.Print("rating: ")
	var input string

	// can only exit with ctrl+\, not ctrl+c
	// empty input is an error; ignore this
	_, _ = fmt.Scanln(&input)

	var newRating int
	switch input {

	case "1", "2", "3", "4", "5":
		newRating, _ = strconv.Atoi(input)
		discogsReq(
			urlpath,
			"PUT",
			map[string]any{
				"username":   config.Discogs.Username,
				"release_id": r.Id,
				"rating":     newRating,
			},
		)

	case "x":
		// TODO: return some enum variant, to signal to caller to do
		// something
		panic("not impl")

	case "":
		return noopInt

	default:
		// TODO: should loop until input in [12345] or empty
		log.Println("invalid rating:", input)
		return noopInt

	}

	postUrlPath, err := url.JoinPath(
		"users",
		config.Discogs.Username,
		"collection/folders/1/releases",
		strconv.Itoa(r.Id),
	)
	if err != nil {
		panic(err)
	}

	discogsReq(postUrlPath, "POST", nil)
	return newRating
} // }}}

type SearchResult struct {
	Pagination map[string]any
	Results    []Release
}

// releases
func discogsSearch(artist string, album string) SearchResult {
	// returning SearchResult (instead of []Release) might look weird
	// (compared to discogsSearchArtist), but i want to be able to get
	// primary via a method for clearer intent (i.e. `result.Primary()`
	// instead of `getPrimary(releases)`)
	log.Println("searching", artist, album)
	resp := discogsReq(
		"/database/search",
		"GET",
		// compiler does -not- allow map[string]string, which is silly
		map[string]any{"artist": alnum(artist), "release_title": alnum(album)},
	)
	return deserialize(resp, SearchResult{})
}

// If r.Results contains a master release (correctness is not checked), returns
// the first master. Otherwise returns the first result (as it is probably
// still meaningful to callers).
//
// Note: a GET call is always performed.
//
// If no results are found, returns empty Release (Id = 0); callers should
// check Release.Id.
func (r *SearchResult) Primary() Release {
	// TODO: return *Release (can check nil = clearer intent)
	if len(r.Results) == 0 {
		return Release{}
	}
	for i, res := range r.Results {
		if i > config.Discogs.MaxResults {
			break
		}

		if res.MasterId == 0 {
			time.Sleep(time.Second)
			continue
		}

		m := deserialize(
			// TODO: should use joinpath, but i'm lazy to handle errors
			discogsReq("/masters/"+strconv.Itoa(res.MasterId), "GET", nil),
			Release{},
		)
		// log.Println("foo", m)
		ensure(len(m.Artists) > 0)
		return m

	}
	return deserialize(
		discogsReq("/releases/"+strconv.Itoa(r.Results[0].Id), "GET", nil),
		Release{},
	)
}

type Artist struct {
	Id          int
	ResourceUrl string          `json:"resource_url"`
	UserData    map[string]bool `json:"user_data"` // in_collection

	// TODO: in search, json key is 'title', otherwise 'name' in all other
	// contexts. this is very footgun-y, so i need to do something about it

	Name  string
	Title string // search-only
}

// additional heuristics/tui will usually be required to select the correct
// artist; this is left to callers
func discogsSearchArtist(artist string) []Artist {
	resp := discogsReq(
		"/database/search",
		"GET",
		map[string]any{"q": alnum(artist), "type": "artist"},
	)
	return deserialize(resp, struct {
		Results []Artist
	}{}).Results
}

// Returns artist releases (which are not full releases)
func (a *Artist) Releases() []Release {
	// /artists/{a.id}/releases
	urlpath, _ := url.JoinPath(
		"artists",
		strconv.Itoa(a.Id),
		"releases",
	)

	resp := discogsReq(
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

	"DVD-V":   nil,
	"Shellac": nil,
	"Single":  nil,
}

// Currently only supports artist releases
func (r *Release) ignored() bool {
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

// rate all releases
func (a *Artist) rate() {
	i := 0
	for _, rel := range a.Releases() {
		if i > 100 || rel.inCollection() || rel.Role != "Main" || rel.ignored() {
			i++
			continue
		}
		log.Println("chk:", filepath.Join(config.Library.Root, rel.Artist, rel.Title))
		rel.rate()
		return
	}
}
