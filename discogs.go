package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

const API_PREFIX = "https://api.discogs.com"

func discogsGet(urlpath string) string {
	client := &http.Client{
		//
	}

	url := fmt.Sprintf("%s/%s", API_PREFIX, urlpath)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	token := os.Getenv("DISCOGS_TOKEN") // TODO: read from config
	req.Header.Add("Authorization", token)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	// TODO: return plain struct (don't deserialise)
	return string(body)
}

func rateRelease(id int) {
	username := os.Getenv("DISCOGS_USERNAME") // TODO: read from config
	url := fmt.Sprintf("releases/%d/rating/%s", id, username)
	resp := discogsGet(url)
	fmt.Println(resp)

	// put
	// data = json.dumps(  # dict -> json str
	//     {
	//         "username": dc.USERNAME,
	//         "release_id": release_id,
	//         "rating": int(rating),
	//     }
	// )

	// # add to collection -- must be done last to prevent duplicate additions
	// # (post is not idempotent)
	// # https://www.discogs.com/developers#page:user-collection,header:user-collection-add-to-collection-folder
	// response = json.loads(
	//     requests.post(
	//         url=dc.API_PREFIX
	//         + f"/users/{dc.USERNAME}/collection/folders/1/releases/{release_id}",
	//         headers=dc.HEADERS,
	//         timeout=3,
	//     ).content
	// )
}
