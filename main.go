package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// to be able to import functions from same package (need not be public), need:
// go run *.go
// https://stackoverflow.com/a/43953582

// sample n items from queue
// directory tui (fzf-like -- bubbletea)
// discogs rate (get, put, post)
// discogs artist
// mpv album
// check resume

// no error-handling
func library_root() string {
	// TODO: lazy_static equivalent?
	return os.Getenv("MU")
}

func main() {
	dirs, err := os.ReadDir(library_root()) // sorted
	// dirs, err := os.ReadDir(library_root() + "/Metallica")
	if err != nil {
		log.Fatal(err)
	}

	p := tea.NewProgram(initialModel(dirs[:10])) //, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}

	dir := filepath.Join(library_root(), "Metallica", "Ride the Lightning (1984)")
	mpv_args := strings.Split("--mute=no --no-audio-display --pause=no --start=0%", " ")
	mpv_args = append(mpv_args, dir)

	// TODO: handover stdout + keyboard control to mpv
	exec.Command("mpv", mpv_args...).Run()

	// TODO: relpath -> search -> primary release id -> rate

	// discogsGet(4319735)

	// _ = p
}
