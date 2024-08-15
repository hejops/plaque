package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

const API_PREFIX = "https://api.discogs.com"

func discogsGet(id int) {
	client := &http.Client{
		//
	}

	url := fmt.Sprintf("%s/releases/%s", API_PREFIX, strconv.Itoa(id))
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
	fmt.Println(string(body))
}

func rateRelease(id int) {
}
