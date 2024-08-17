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

func discogsReq(urlpath string, method string, data map[string]any) *http.Response {
	_url, err := url.JoinPath(ApiPrefix, urlpath)
	if err != nil {
		panic(err)
	}

	// map -> []byte -> bytes.Buffer

	var req *http.Request
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		// https://stackoverflow.com/a/24455606
		req, err = http.NewRequest(method, _url, bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, _url, nil)
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
}

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

	fmt.Print("rating: ")
	var input string
	fmt.Scanln(&input)
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
