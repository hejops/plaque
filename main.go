package main

import (
	"fmt"
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

func play(dir string) tea.Cmd {
	mpv_args := strings.Split("--mute=no --no-audio-display --pause=no --start=0%", " ")
	mpv_args = append(mpv_args, dir)
	cmd := exec.Command("mpv", mpv_args...)

	// // handover std streams + keyboard control to mpv
	// // https://github.com/search?type=code&q=exec.Command(%22mpv%22
	// // https://github.com/aynakeya/blivechat/blob/9c4a8ddddc9c5295a9a8d368ac5dab62557397c5/app/heiting/heiting.go#L136
	// cmd.Stdout = os.Stdout
	// cmd.Stdin = os.Stdin
	// cmd.Stderr = os.Stderr
	//
	// if err := cmd.Run(); err != nil {
	// 	panic(err)
	// }

	return tea.Sequence(
		// if the altscreen is not used, new
		// model is rendered before (above) mpv
		tea.EnterAltScreen,
		tea.ExecProcess(cmd, nil),
		tea.ExitAltScreen,
		tea.Println("rating..."),
		tea.ClearScreen,
	)
}

func main() {
	if _, err := tea.NewProgram(newModel(GetQueue(10), Queue), tea.WithAltScreen()).Run(); err != nil {
		panic(err)
	}
	return

	root := LibraryRoot()
	items, err := os.ReadDir(root)
	if err != nil {
		panic(err)
	}

	maxItems := 10
	// note: if terminal currently has n rows, and len(m.items) > n, only
	// the last n rows will be displayed
	if len(items) > maxItems {
		items = items[:maxItems]
	}
	items2 := []string{}
	for _, x := range items {
		items2 = append(items2, filepath.Join(root, x.Name()))
	}

	if _, err := tea.NewProgram(newModel(items2, Artists)).Run(); err != nil {
		panic(err)
	}

	fmt.Println("end")

	// dir := filepath.Join(LibraryRoot(), "Metallica", "Ride the Lightning (1984)")

	// TODO: relpath -> search -> primary release id -> rate

	// discogsGet(4319735)
	// rateRelease(4319735)

	// _ = p
}
