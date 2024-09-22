package discogs

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)

const ApiPrefix = "https://api.discogs.com"

// urlpath -cannot- contain query params; these should be passed as data
// instead.
//
// data should either be GET query params (in which case all values must be
// strings), or PUT json data (in which case values must be correctly typed by
// the caller).
func makeReq(urlpath string, method string, data map[string]any) *http.Response {
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

	req.Header.Set("Authorization", "Discogs token="+Config.Key)
	req.Header.Set("Cache-Control", "no-cache")

	log.Println(method, u.RequestURI())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	return resp
}
