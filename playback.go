// Playback and queue management. Playback is delegated to mpv, which is
// allowed to run in a blocking manner for full keyboard control. As such,
// multiple instances of the program are to be expected, but only one instance
// can be running mpv; other instances can only add to queue, and terminate
// immediately.
//
// Scrobbling is out of scope of this program; consider
// https://github.com/Feqzz/mpv-lastfm-scrobbler

package main

import (
	"bufio"
	"io"
	"io/fs"
	"log"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func getResumes() []string {
	// When mpv is quit with the `quit_watch_later` command, a file is
	// written to this dir, containing the full path to the file.

	var resumes []string
	err := filepath.WalkDir(WatchLaterDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		fo, _ := os.Open(path)
		defer fo.Close()
		sc := bufio.NewScanner(fo)
		sc.Scan() // only need to read 1st line
		line := sc.Text()

		file := line[2:] // # "# "
		if fi, e := os.Stat(file); e == nil &&
			!fi.IsDir() &&
			// TODO: startswith?
			strings.Contains(file, config.Library.Root) {
			rel, _ := filepath.Rel(config.Library.Root, filepath.Dir(file))
			resumes = append(resumes, rel)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return resumes
}

// Select n random items from the queue file (containing relpaths), and return
// them as fullpaths
//
// If n = 0, the entire queue is returned without shuffling
func getQueue(n int) []string {
	if n < 0 {
		panic("invalid")
	}

	// my queue file is about 8000, so it is worth doing some optimisation
	// https://scribe.rip/golicious/comparing-ioutil-readfile-and-bufio-scanner-ddd8d6f18463

	// https://stackoverflow.com/a/16615559
	var relpaths []string
	file, err := os.Open(config.Library.Queue)
	if err != nil {
		log.Fatalln(err)
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
		log.Fatalln(err)
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
	err := os.WriteFile(
		config.Library.Queue,
		[]byte(strings.Join(items, "\n")),
		0666,
	)
	if err != nil {
		panic(err)
	}
}

// https://github.com/picosh/pico/blob/4632c9cd3d7bc37c9c0c92bdc3dc8a64928237d8/tui/senpai.go#L10

// wrapper to call functions in a blocking manner
type postPlaybackCmd struct {
	relpath string
}

// required methods for tea.ExecCommand

func (c *postPlaybackCmd) Run() error {
	// if resume, return early
	_ = filepath.WalkDir(WatchLaterDir, func(path string, d fs.DirEntry, err error) error {
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		s := string(b)
		if strings.Contains(s, config.Library.Root) {
			log.Println("will resume:", s)
			return fs.SkipAll
		}
		return err
	})

	q := getQueue(0)
	nq := *remove(&q, c.relpath)
	ensure(len(q)-len(nq) == 1)
	writeQueue(nq)
	log.Println("removed:", c.relpath)

	artist, album := filepath.Split(c.relpath)

	// remove translation
	if artist[len(artist)-1] == ')' {
		i := strings.LastIndex(artist, "(")
		artist = artist[:i]
	}

	if album[len(album)-1] == ')' { // " (YYYY)"
		album = album[:len(album)-7]
	}

	var res SearchResult
	if album[len(album)-1] == ']' { // " [performer, ...]"
		res = discogsSearch(movePerfsToArtist(artist, album))
	} else {
		res = discogsSearch(artist, album)
	}
	rel := res.Primary()
	rel.rate()

	artists := discogsSearchArtist(artist)
	if len(artists) > 0 {
		// artists[0].rate()
		for _, a := range artists {
			// log.Println("artist:", a.Title, a.UserData)
			if !a.UserData["in_collection"] {
				continue
			}
			log.Println("rating releases of", a.Title)
			a.rate()
			break // do we need to try more than 1?
		}
	}

	return nil
}
func (c *postPlaybackCmd) SetStderr(io.Writer) {}
func (c *postPlaybackCmd) SetStdin(io.Reader)  {}
func (c *postPlaybackCmd) SetStdout(io.Writer) {}

func play(relpath string) tea.Cmd {
	path := filepath.Join(config.Library.Root, relpath)
	mpvCmd := exec.Command("mpv", append(mpvArgs, path)...)
	log.Println("playing:", path)

	return tea.Sequence(
		// tea.EnterAltScreen,
		tea.ExecProcess(mpvCmd, nil),
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

		tea.Exec(&postPlaybackCmd{relpath: relpath}, nil),

		// tea.ExitAltScreen,
		// tea.ClearScreen,
	)
}
