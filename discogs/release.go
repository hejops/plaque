package discogs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"strconv"
)

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

func (r *Release) inCollection() bool { return r.Stats["user"]["in_collection"] > 0 }

func (r *Release) IsRateable() bool { return !r.inCollection() && r.Role == "Main" && !r.ignored() }

// TODO: private/namespace? these errors are only relevant for 1 function.
// enums address this somewhat, but cannot be nil'd
var (
	ErrNotFound     = errors.New("Release not found")
	ErrNotRated     = errors.New("Release not rated")
	ErrNotRateable  = errors.New("Release not eligible for rating")
	ErrAlreadyRated = errors.New("Release already rated") // in Rust, would contain an inner value
	ErrUnhandled    = errors.New("Failed to parse JSON (probably)")

	// i would have preferred an enum, but an int cannot be nil'd, and
	// leads to unclear intent
	// https://old.reddit.com/r/golang/comments/fg6527/simple_error_enum/fk53no5/
)

func (r *Release) Rate() (int, error) { // {{{
	if r.Id == 0 {
		return 0, ErrNotFound
	}

	if !r.IsRateable() {
		return 0, ErrNotRateable
	}

	// TODO: leaky abstraction that should be handled at lower level
	switch {
	case r.Primary > 0: // master release
		r = deserialize(
			makeReq("/releases/"+strconv.Itoa(r.Primary), "GET", nil),
			&Release{},
		)
	case r.Artist != "": // artist release
		r = deserialize(
			makeReq("/releases/"+strconv.Itoa(r.Id), "GET", nil),
			&Release{},
		)
	}

	// releases/{r.Id}/rating/{username}
	urlpath, _ := url.JoinPath(
		"releases",
		strconv.Itoa(r.Id),
		"rating",
		Config.Username,
	)

	resp := makeReq(urlpath, "GET", nil)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var currentRating map[string]any
	// an error here usually means incorrect was Id supplied (i.e. master
	// id instead of release id)
	if err := json.Unmarshal(body, &currentRating); err != nil {
		return 0, ErrUnhandled
	}
	if int(currentRating["rating"].(float64)) != 0 {
		log.Println("already rated:", r.Id, r.Title, currentRating)
		return 0, ErrAlreadyRated
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
		makeReq(
			urlpath,
			"PUT",
			map[string]any{
				"username":   Config.Username,
				"release_id": r.Id,
				"rating":     newRating,
			},
		)

	case "x":
		// TODO: return some enum variant, to signal to caller to do
		// something
		panic("not impl")

	case "":
		return 0, ErrNotRated

	default:
		// TODO: should loop until input in [12345] or empty
		log.Println("invalid rating:", input)
		return 0, ErrNotRated

	}

	postUrlPath, err := url.JoinPath(
		"users",
		Config.Username,
		"collection/folders/1/releases",
		strconv.Itoa(r.Id),
	)
	if err != nil {
		panic(err)
	}

	makeReq(postUrlPath, "POST", nil)
	return newRating, nil
} // }}}
