package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// to be able to import functions from same package (need not be public), need:
// go run *.go
// https://stackoverflow.com/a/43953582

// discogs rate (get, put, post)
// discogs artist
// check resume

func main() {
	if _, err := tea.NewProgram(newBrowser(getQueue(10), Queue), tea.WithAltScreen()).Run(); err != nil {
		panic(err)
	}
	return

	if _, err := tea.NewProgram(artistBrowser()).Run(); err != nil {
		panic(err)
	}

	fmt.Println("end")

	// TODO: relpath -> search -> primary release id -> rate

	// discogsGet(4319735)
	// rateRelease(4319735)

	// _ = p
}
