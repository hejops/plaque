// Playback and queue management. Playback is delegated to mpv, which is
// blocking.
//
// Scrobbling is out of scope of this program; consider
// https://github.com/Feqzz/mpv-lastfm-scrobbler

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Select n random items from the queue file (containing relpaths), and return
// them as fullpaths
//
// If n = 0, the entire queue is returned without shuffling
func getQueue(n int) []string {
	if n < 0 {
		panic("invalid")
	}

	// my queue file is about 8000, so it is worth doing some optimisation

	// b, err := os.ReadFile(config.Library.Queue)
	// if err != nil {
	// 	panic(err)
	// }
	// relpaths := strings.Split(string(b), "\n")

	// https://stackoverflow.com/a/16615559
	var relpaths []string
	file, err := os.Open(config.Library.Queue)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	sc := bufio.NewScanner(file)
	// chance := float32(n) / float32(foo)
	for sc.Scan() {
		// could potentially move rng stuff in here (e.g. chance =
		// n/len, or generate idxs in advance); but need to get number
		// of newlines in advance (so need to read whole file
		// anyway...)
		// if chance < rand.Float32() {
		// }
		relpaths = append(relpaths, sc.Text())
	}
	if sc.Err() != nil {
		log.Fatal(err)
	}

	// TODO: split off sampling
	switch n {
	case 0:
		return relpaths
	default:
		var sel []string
		idxs := rand.Perm(len(relpaths) - 1)
		for _, idx := range idxs[:n] {
			sel = append(sel, relpaths[idx])
		}
		return sel
	}
}

func writeQueue(items []string) {
	os.WriteFile(config.Library.Queue+"_", []byte(strings.Join(items, "\n")), 0666)
}

// https://github.com/picosh/pico/blob/4632c9cd3d7bc37c9c0c92bdc3dc8a64928237d8/tui/senpai.go#L10

// wrapper to call functions in a blocking manner
type rateCmd struct {
	relpath string
}

// required methods for tea.ExecCommand

func (c *rateCmd) Run() error {
	// if resume, return early

	artist, album := filepath.Split(c.relpath)
	// TODO: artist endswith ) -> remove translation
	if strings.HasSuffix(album, ")") { // " (YYYY)"
		album = album[:len(album)-7]
	}
	// if strings.HasSuffix(album, "]") { // " [performer, ...]"
	// 	album = album[:len(album)-7]
	// }
	log.Println(artist, album)

	res := discogsSearch(artist, album)
	pri := res.Primary()
	if pri > 0 {
		fmt.Println(pri)
		r := Release{Id: pri} // awkward, but whatever
		r.rate()
		// rateRelease(pri)
		// rateArtist(artist)
	}

	q := getQueue(0)
	nq := remove(q, c.relpath)
	writeQueue(nq)

	return nil
}
func (c *rateCmd) SetStderr(io.Writer) {}
func (c *rateCmd) SetStdin(io.Reader)  {}
func (c *rateCmd) SetStdout(io.Writer) {}

func play(relpath string) tea.Cmd {
	mpv_args := strings.Split("--mute=no --no-audio-display --pause=no --start=0%", " ")
	mpv_args = append(mpv_args, filepath.Join(config.Library.Root, relpath))
	cmd := exec.Command("mpv", mpv_args...)

	return tea.Sequence(
		// if the altscreen is not used, new model is (inexplicably)
		// rendered before (above) mpv
		// tea.EnterAltScreen,
		tea.ExecProcess(cmd, nil),
		// tea.ExitAltScreen,
		// tea.ClearScreen,

		// // this func/Msg actually works (program will wait for 1 line
		// // of stdin, then proceed with put/post), but since it is not
		// // blocking, very bizarre behaviour will be observed (e.g. new
		// // selector is rendered on top of prompt)
		// func() tea.Msg {
		// 	rateRelease(4319735)
		// 	return nil
		// },

		tea.Exec(&rateCmd{relpath: relpath}, nil),

		// tea.ExitAltScreen,
		// tea.ClearScreen,
	)
}
