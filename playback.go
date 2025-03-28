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
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"plaque/discogs"
)

const QueueCount = 5

func mpvRunning() bool {
	// https://github.com/mitchellh/go-ps/blob/master/process_linux.go
	running := false
	_ = filepath.WalkDir("/proc", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() || filepath.Base(path) != "stat" {
			return nil
		}
		b, e := os.ReadFile(path)
		if e != nil {
			panic(e)
		}

		s := string(b)
		start := strings.IndexRune(s, '(')
		end := strings.IndexRune(s, ')')
		if end < 0 {
			return nil
		}
		if s[start+1:end] == "mpv" {
			running = true
			return fs.SkipAll
		}
		return nil
	})
	return running
}

func getResumes() *[]string {
	// When mpv is quit with the `quit_watch_later` command, a file is
	// written to this dir, containing the full path to the file.

	if config == nil {
		panic("init was not done")
	}

	var resumes []string
	err := filepath.WalkDir(
		config.Mpv.WatchLaterDir,
		func(path string, d fs.DirEntry, err error) error {
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
				strings.HasPrefix(file, config.Library.Root) {
				rel, _ := filepath.Rel(config.Library.Root, filepath.Dir(file))
				resumes = append(resumes, rel)
			}
			return nil
		},
	)
	if err != nil {
		panic(err)
	}
	if len(resumes) == 0 {
		return nil
	}
	return &resumes
}

func willResume(relpath string) (resume bool) {
	path := filepath.Join(config.Library.Root, relpath)
	_ = filepath.WalkDir(config.Mpv.WatchLaterDir, func(p string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			panic(err)
		}
		if strings.Contains(string(b), path) {
			resume = true
			return fs.SkipAll
		}
		return nil
	})
	return resume
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

	// according to a simple benchmark, os.ReadFile() is almost 2-3x as
	// fast as bufio.NewScanner(). NewScanner can probably only be faster
	// if we know how to stop scanning early (which we don't)

	b, err := os.ReadFile(config.Library.Queue)
	if err != nil {
		panic(err)
	}
	relpaths := strings.Split(string(b), "\n")

	// TODO: split off sampling
	switch n {
	case 0:
		// 'valid' text file should always end with a trailing newline.
		// in this case, last element will be empty string
		return relpaths[:len(relpaths)-1]
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
		// []byte(strings.Join(items, "\n"))
		[]byte(strings.Join(items, "\n")+"\n"), // file must have trailing newline
		0666,
	)
	if err != nil {
		panic(err)
	}
}

// https://github.com/picosh/pico/blob/4632c9cd3d7bc37c9c0c92bdc3dc8a64928237d8/tui/senpai.go#L10

// wrapper to call functions in a blocking manner (via Run)
type postPlaybackCmd struct{ relpath string }

// required methods for tea.ExecCommand

func (c *postPlaybackCmd) Run() error {
	if willResume(c.relpath) {
		// we -could- propagate some error to tea.Exec, which can be
		// handled there. for practical purposes, all we need to do is
		// just return to Queue
		log.Println("will resume:", c.relpath)
		os.Exit(0)
		return nil
		// // TODO: figure out how to return a 'real' error
		// return fmt.Errorf("resume")
	}

	log.Println("playback done")

	q := getQueue(0)
	nq := *remove(&q, c.relpath)
	ensure(len(q)-len(nq) == 1)
	writeQueue(nq)
	log.Println("removed:", c.relpath)

	if !discogsEnabled {
		log.Println("no discogs key, skipping rate")
		return nil
	}

	artist, album := filepath.Split(c.relpath)

	// remove possible translation
	if artist[len(artist)-1] == ')' {
		i := strings.LastIndex(artist, "(")
		artist = artist[:i-1]
	}

	// remove album suffix " (YYYY)"
	if album[len(album)-1] == ')' {
		album = album[:len(album)-7]
	}

	// aside from edge cases, only classical albums have " [performer, ...]" suffix
	var res discogs.SearchResult
	if album[len(album)-1] == ']' {
		res = discogs.Search(movePerfsToArtist(artist, album))
	} else {
		res = discogs.Search(artist, album)
	}

	rel := res.Primary()
	if rating, _ := rel.Rate(); rating == 1 &&
		// guard rail to prevent deleting classical artists
		album[len(album)-1] != ']' {
		p := filepath.Join(config.Library.Root, artist)
		if _, err := os.Stat(p); err != nil {
			return nil
		}

		fmt.Printf("Delete %s? [y/N] ", artist)
		var del string
		_, _ = fmt.Scanln(&del)
		if del == "y" {
			_ = os.RemoveAll(p)
			fmt.Println("Deleted", p)
		}
		return nil
	}

	// this is not terribly ergonomic; but wrapping the returned []Artist
	// in a struct seems even more annoying
	artists := discogs.SearchArtist(artist)
	if len(artists) == 0 {
		return nil
	}

	art := discogs.BrowseArtists(artists)
	if art != nil {
		return nil
	}

	// art.Rate(checkDir) // nonsensical api
	// art.Rate() // sane api, but no checkDir

	for _, rel := range art.Releases() {
		// if !rel.IsRateable() || checkDir(artist, rel.Title) {
		// 	continue
		// }
		if checkDir(artist, rel.Title) {
			continue
		}

		// if errors.Is(err, discogs.ErrAlreadyRated) {
		// 	continue
		// }

		_, err := rel.Rate()
		switch err {
		case discogs.ErrAlreadyRated, discogs.ErrNotRateable:
			continue
		}

		break
	}

	return nil
}

func (c *postPlaybackCmd) SetStderr(io.Writer) {}

func (c *postPlaybackCmd) SetStdin(io.Reader) {}

func (c *postPlaybackCmd) SetStdout(io.Writer) {}

func play(relpath string) tea.Cmd {
	timer := time.NewTimer(time.Second * 2)
	defer timer.Stop()
	go func() {
		// fmt.Println("please wait...", <-timer.C)
		// fmt won't work outside View
		log.Println("please wait...", <-timer.C)
	}()

	// TODO: online mode (search ytm)
	path := filepath.Join(config.Library.Root, relpath)
	mpvCmd := exec.Command("mpv", append(strings.Fields(config.Mpv.Args), path)...)
	log.Println("playing:", path)

	return tea.Sequence(
		func() tea.Msg {
			if config.Playback.Before == "" {
				return nil
			}
			args := strings.Fields(config.Playback.Before)
			var expandedArgs []string
			for _, arg := range args[1:] {
				if arg[0] == '$' {
					arg = os.ExpandEnv(arg)
				}
				expandedArgs = append(expandedArgs, arg)
			}
			cmd := exec.Command(args[0], expandedArgs...)
			err := cmd.Run()
			if err != nil {
				log.Println("could not run cmd", cmd, err)
			}
			return nil
		},
		tea.ExecProcess(mpvCmd, nil),
		tea.Exec(
			&postPlaybackCmd{relpath: relpath},
			nil,
			// // if you need to check/handle the error returned by
			// // Run and turn that into a Cmd, you could; otherwise,
			// // we just return to Queue
			// func(err error) tea.Msg {
			// 	// if err.Error() == "resume" {
			// 	// 	log.Println("quitting and will resume")
			// 	// 	return tea.Quit()
			// 	// 	// return nil
			// 	// }
			// 	return nil
			// },
		),
		tea.ClearScreen,
	)
}
