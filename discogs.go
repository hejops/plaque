package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// discogs rate (get, put, post)
// discogs artist

const ApiPrefix = "https://api.discogs.com"

// func addKey(req *http.Request) {
// 	req.Header.Add("Authorization", "Discogs token="+config.Discogs.Key)
// }

// urlpath -cannot- contain query params; these should be passed as data
// instead.
//
// data should either be GET query params, or PUT json data.
func discogsReq(urlpath string, method string, data map[string]any) *http.Response { // {{{

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

// TODO: turn into Release method

func rateRelease(id int) { // {{{
	//

	// "releases/{id}/rating/{username}"
	urlpath, _ := url.JoinPath(
		"releases",
		strconv.Itoa(id),
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
		fmt.Println("already rated:", id)
		return
	}

	// tea.Printf("rating: ")
	fmt.Printf("rating: ")
	var input string
	fmt.Scanln(&input)
	if input == "" {
		return
	}
	newRating, _ := strconv.Atoi(input)
	discogsReq(
		urlpath,
		"PUT",
		map[string]any{
			"username":   config.Discogs.Username,
			"release_id": id,
			"rating":     newRating,
		},
	)

	postUrlPath, err := url.JoinPath(
		"users",
		config.Discogs.Username,
		"collection/folders/1/releases",
		strconv.Itoa(id),
	)
	if err != nil {
		panic(err)
	}

	discogsReq(postUrlPath, "POST", nil)
} // }}}

// TODO: relpath -> search -> primary release id -> rate

// remember: all fields must be uppercase, and any fields in camelcase must be
// marked in order to be parsed

type SearchRelease struct {
	Id          int
	ReleaseType string `json:"type"` // 'release' or 'master'

	MasterId    int    `json:"master_id"`  // may be 0
	MasterUrl   string `json:"master_url"` // may be empty
	ResourceUrl string `json:"resource_url"`
	Title       string
	Year        string // may be empty

	Genre []string

	Community map[string]int
}

type SearchResult struct {
	// pagination map[string]any
	Results []SearchRelease
}

func (r *SearchResult) Master() Master {
	// iterate through all results
	// cases:
	// 1. master (Type) -- return
	// 2. release with master (MasterId!=0)
	// 3. release w/o master -- return 1st r.Id
	var m Master
	for _, r := range r.Results {
		if r.ReleaseType == "master" {
			resp := discogsReq("/masters/"+strconv.Itoa(r.Id), "GET", nil)
			return deserialize(resp, Master{})
		}
	}
	return m
}

type Master struct {
	Id      int
	Primary int `json:"main_release"`

	// LowestPrice float32
	Title       string
	Uri         string
	VersionsUrl string `json:"versions_url"`
	Year        string

	Artists   []map[string]any
	Genre     []string
	Tracklist []map[string]any
}

func discogsSearch(artist string, album string) SearchResult {
	resp := discogsReq(
		"/database/search",
		"GET",
		// compiler does -not- allow map[string]string, which is silly
		map[string]any{"artist": artist, "release_title": album},
	)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var data SearchResult
	json.Unmarshal(body, &data)
	return data
}
