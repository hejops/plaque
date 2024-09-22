package discogs

import (
	"log"
	"strconv"
	"time"
)

type SearchResult struct {
	Pagination map[string]any
	Results    []Release
}

// Search for releases
func Search(artist string, album string) SearchResult {
	// returning SearchResult (instead of []Release) might look weird
	// (compared to SearchArtist), but i want to be able to get primary via
	// a method for clearer intent (i.e. `result.Primary()` instead of
	// `getPrimary(releases)`)
	log.Println("searching", artist, album)
	resp := makeReq(
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
	// TODO: return *Release? (can check nil = clearer intent)
	if len(r.Results) == 0 {
		return Release{}
	}
	for i, res := range r.Results {
		if i > Config.MaxResults {
			break
		}

		if res.MasterId == 0 {
			time.Sleep(time.Second)
			continue
		}

		m := deserialize(
			// TODO: should use url.joinpath, but i'm lazy to handle errors
			makeReq("/masters/"+strconv.Itoa(res.MasterId), "GET", nil),
			Release{},
		)
		// log.Println("foo", m)
		ensure(len(m.Artists) > 0)
		return m

	}
	return deserialize(
		makeReq("/releases/"+strconv.Itoa(r.Results[0].Id), "GET", nil),
		Release{},
	)
}
