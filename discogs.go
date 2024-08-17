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

const ApiPrefix = "https://api.discogs.com"

// HEADERS = {"Cache-Control": "no-cache"}

func addKey(req *http.Request) {
	req.Header.Add("Authorization", "Discogs token="+config.Discogs.Key)
}

func discogsGet(urlpath string) []byte {
	_url, err := url.JoinPath(ApiPrefix, urlpath)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("GET", _url, nil)
	if err != nil {
		panic(err)
	}

	// addKey + http.DefaultClient.Do can be refactored if needed
	addKey(req)
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return body
}

func discogsPut(urlpath string, data map[string]any) {
	_url, err := url.JoinPath(ApiPrefix, urlpath)
	if err != nil {
		panic(err)
	}

	// map -> []byte -> bytes.Buffer
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	// https://stackoverflow.com/a/24455606
	req, err := http.NewRequest("PUT", _url, bytes.NewBuffer(b))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	addKey(req)
	if resp, err := http.DefaultClient.Do(req); err != nil {
		panic(err)
	} else if resp.StatusCode != 201 {
		panic(resp)
	}
}

func rateRelease(id int) {
	//

	// "releases/{id}/rating/{username}"
	urlpath, _ := url.JoinPath(
		"releases",
		strconv.Itoa(id),
		"rating",
		config.Discogs.Username,
	)
	resp := discogsGet(urlpath)
	var currentRating map[string]any
	json.Unmarshal(resp, &currentRating)
	if int(currentRating["rating"].(float64)) != 0 {
		fmt.Println("already rated:", id)
		return
	}

	fmt.Print("rating: ")
	var input string
	fmt.Scanln(&input)
	newRating, _ := strconv.Atoi(input)
	discogsPut(
		urlpath,
		map[string]any{
			"username":   config.Discogs.Username,
			"release_id": id,
			"rating":     newRating,
		},
	)

	postUrl, err := url.JoinPath(
		ApiPrefix,
		"users",
		config.Discogs.Username,
		"collection/folders/1/releases",
		strconv.Itoa(id),
	)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", postUrl, nil)
	if err != nil {
		panic(err)
	}
	addKey(req)
	if resp, err := http.DefaultClient.Do(req); err != nil {
		panic(err)
	} else if resp.StatusCode != 201 {
		panic(resp)
	}
	// fmt.Println("OK")
}

// TODO: relpath -> search -> primary release id -> rate
