package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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
	u = u.JoinPath(urlpath) // no error returned, wtf?

	// map -> []byte -> bytes.Buffer

	var req *http.Request
	var err error

	switch method {

	case "GET", "POST":
		if data != nil {
			q := url.Values{}
			for k, v := range data {
				q.Set(k, v.(string))
			}
			u.RawQuery = q.Encode()
			req, err = http.NewRequest(method, u.String(), nil)
		} else {
			req, err = http.NewRequest(method, u.String(), nil)
		}

	case "PUT":
		b, err := json.Marshal(data)
		if err != nil {
			panic(err)
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

	req.Header.Add("Authorization", "Discogs token="+config.Discogs.Key)
	req.Header.Add("Cache-Control", "no-cache")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	return resp
} // }}}

// remember: all fields must be uppercase, and any fields in camelcase must be
// marked in order to be parsed

// this struct is shared across several contexts: search results, artist
// releases, and actual releases
type Release struct {
	Id          int
	ReleaseType string `json:"type"` // 'release' or 'master'

	MasterId    int    `json:"master_id"`  // may be 0
	MasterUrl   string `json:"master_url"` // may be empty
	ResourceUrl string `json:"resource_url"`
	Title       string
	Year        string // may be empty

	// search-only (?)

	Genre     []string
	Community map[string]int

	// artist-only

	Artist string
	Format string // ", "-delimited
	Label  string
	Role   string // typically "Main"
	Stats  map[string]map[string]int
}

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
	resp := discogsReq(
		"/database/search",
		"GET",
		// compiler does -not- allow map[string]string, which is silly
		map[string]any{"artist": alnum(artist), "release_title": alnum(album)},
	)
	return deserialize(resp, SearchResult{})
}

// if r.Results contains a master release (correctness is not checked), returns
// the id of the primary version of the first master. otherwise returns id of
// first result (as it is probably still meaningful to callers).
//
// if no results are found, returns empty Release (Id = 0); callers should
// check Release.Id
//
// in my use case, I have never actually needed to use the master id.
func (r *SearchResult) Primary() Release {
	if len(r.Results) == 0 {
		// return 0
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

		// merge into Release, if necessary
		// type Master struct {
		// 	Id      int
		// 	Primary int `json:"main_release"`
		//
		// 	// LowestPrice float32
		// 	Title       string
		// 	Uri         string
		// 	VersionsUrl string `json:"versions_url"`
		// 	Year        string
		//
		// 	Artists   []map[string]any
		// 	Genre     []string
		// 	Tracklist []map[string]any
		// }

		return deserialize(
			// TODO: should use joinpath, but i'm lazy to handle errors
			discogsReq("/masters/"+strconv.Itoa(res.MasterId), "GET", nil),
			Release{},
			// struct {
			// 	Id      int
			// 	Primary int `json:"main_release"`
			// }{},
		) //.Primary
	}
	return r.Results[0] //.Id
}

// does nothing if release already rated
func (r *Release) rate() { // {{{
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
	json.Unmarshal(body, &currentRating)
	if int(currentRating["rating"].(float64)) != 0 {
		fmt.Println("already rated:", r.Id)
		return
	}

	fmt.Print("rating: ")
	var input string
	fmt.Scanln(&input) // can only exit with ctrl+\, not ctrl+c

	switch input {

	case "1", "2", "3", "4", "5":
		newRating, _ := strconv.Atoi(input)
		discogsReq(
			urlpath,
			"PUT",
			map[string]any{
				"username":     config.Discogs.Username,
				"release_r.Id": r.Id,
				"rating":       newRating,
			},
		)

	case "":
		return

	default:
		fmt.Println("invalr.Id:", input)
		return

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
} // }}}

type Artist struct {
	Id          int
	Name        string          `json:"title"`
	ResourceUrl string          `json:"resource_url"`
	UserData    map[string]bool `json:"user_data"` // in_collection
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
