// The Browser can exist in one of three states (Modes). Each state leverages
// the same list-based TUI to present a (different) set of items to the user:
//
//	1. Queue: paths of depth 2, typically loaded from a (local) file
//	2. Artists: immediate children directories of root, generated via traversal
//	3. Albums: directories under an artist (i.e. depth 2)
//
// The lists are implemented as a simple fzf-like menu with basic non-fuzzy
// substring matching.
//
// For simplicity of rendering, all items must be valid directories, relative
// to the library root. On selecting an item, the Browser transitions to the
// next state, crudely represented by the following finite state machine:
//
//                  /----> [Playback]  /---> End
//                  |           |     /
//                  |           v    /
//                Queue <--- Albums /--- Artists
//                  ^                       ^
//                   \------ Start --------/
//
// - playback (and the associated post-playback actions) is always blocking
// - on startup, Queue and Artists modes are available
//   - only Queue mode can (and must) transition to playback
//   - Artists mode transitions to Albums mode, then always exits
// - the program can be gracefully exited in any Mode

package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/charmbracelet/x/term"
)

// https://leg100.github.io/en/posts/building-bubbletea-programs/

var IsSelected = map[bool]string{
	true:  "→",
	false: " ",
}

var IsQueued = map[bool]string{
	true:  "Q",
	false: " ",
}

type Mode int

const (
	Queue Mode = iota
	Artists
	Albums
)

// mostly copied from https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics

type Browser struct {
	mode     Mode
	items    []string            // valid relpaths
	queued   map[string]bool     // keys correspond to items
	previews map[string][]string // keys correspond to items

	c      chan string
	noquit bool
	// may only be true in Albums mode

	width  int
	height int

	offset  int
	cursor  int
	input   string
	matches []int
}

// All items must be valid relpaths (relative to root)
func newBrowser(items []string, mode Mode) *Browser {
	// TODO: on cold start, slow os.Stat prevents View from being called
	defer timer("cold stat")()

	// putting this in a goroutine does not prevent blocking (unless the
	// newBrowser call itself is also async). in any case, this is just a
	// guard rail which i intend to remove sooner or later
	go checkRelPaths(items)

	// init window correctly; a "recursively" spawned Browser is
	// initialised with zeroed dimensions!
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		panic("failed to get terminal size")
	}

	return &Browser{
		mode:    mode,
		items:   items,
		matches: intRange(len(items)),
		c:       make(chan string),

		width:  width,
		height: height,
	}
}

// TODO: group the 3 funcs into one: notice that Albums needs a string arg, and
// Queue needs an int arg. these args could be passed as a struct

// type BrowserOpts struct {
// 	mode   Mode
// 	queue  int // number of items to sample
// 	artist string
// }

var firstRun = true

func queueBrowser( /*resume bool*/ ) (b *Browser) {
	// resume should only be true on the first invocation (i.e. on startup)

	if firstRun {
		timer := time.NewTimer(time.Second * 2)
		defer timer.Stop()
		go func() {
			fmt.Println("please wait...", <-timer.C)
			// log.Println("cold disk...", <-timer.C)
		}()
	}

	resumes := getResumes()
	switch {
	case firstRun && resumes != nil:
		b = newBrowser(*resumes, Queue)
		b.noquit = true
		firstRun = false
	default:
		b = newBrowser(getQueue(QueueCount), Queue)
	}

	return b
}

func artistBrowser() *Browser {
	items, _ := descend(config.Library.Root)
	return newBrowser(items, Artists)
}

// Browser.items will be sorted by year.
func albumsBrowser(artist string) *Browser {
	// more complex since we need to check queue and populate the `queued`
	// field

	allQueued := make(map[string]any)
	for _, x := range getQueue(0) {
		allQueued[x] = nil
	}

	// TODO: the rest is i/o; could be goroutine'd?
	albums, err := descend(filepath.Join(config.Library.Root, artist))
	if err != nil {
		panic(err)
	}

	sortByYear(albums)

	items := []string{}
	// queued := make(map[string]any)
	// int keys are much easier to index (for View), but require correct sort
	// queued := make(map[int]bool)
	queued := make(map[string]bool)
	previews := make(map[string][]string)
	for _, alb := range albums {
		// newBrowser requires valid relpaths
		relpath := filepath.Join(artist, alb)
		fullpath := filepath.Join(config.Library.Root, relpath)
		items = append(items, relpath)

		_, q := allQueued[relpath]
		queued[relpath] = q

		p, err := descend(fullpath)
		if err != nil {
			panic(err)
		}
		previews[relpath] = p

	}

	b := newBrowser(items, Albums)
	b.queued = queued
	b.previews = previews

	return b
}

func (b *Browser) updateSearch(msg tea.KeyMsg) {
	// https://github.com/antonmedv/walk/blob/ba821ed78f31e0ebd46eeef19cfe642fc1ec4330/main.go#L427
	// note the pointer; we are mutating Browser

	switch {

	case b.input == "":
		// return all indices
		b.matches = intRange(len(b.items))
		return

	case b.mode == Albums:
		// b.items is relpath, but we want basenames
		b.matches = fuzzySearch(Map(b.items, filepath.Base), b.input)

	default:
		b.matches = fuzzySearch(b.items, b.input)
	}

	if len(b.matches) > 0 {
		b.cursor = 0
	}
}

func (b *Browser) Init() tea.Cmd {
	if b.mode == Queue {
		go func() {
			previews := make(map[string][]string)
			for _, item := range b.items {
				p, err := descend(item)
				if err != nil {
					continue
				}
				previews[item] = p

			}
			b.previews = previews
		}()
	}

	return nil
}

func (b *Browser) Update(msg tea.Msg) (tea.Model, tea.Cmd) { // {{{
	// log.Println("msg:", msg) // not terribly informative

	// // https://leg100.github.io/en/posts/building-bubbletea-programs/
	// spew.Fdump(b.dump, msg)

	// notice this (subtle) reassignment
	switch msg := msg.(type) {

	// https://github.com/charmbracelet/bubbletea/discussions/818#discussioncomment-6914769
	case tea.WindowSizeMsg: // only triggered when window resized?
		b.width = msg.Width
		b.height = msg.Height
		return b, tea.ClearScreen

	case tea.KeyMsg:

		if msg.Type == tea.KeyRunes {
			b.input += string(msg.Runes)
			b.updateSearch(msg)
			return b, nil
		}

		// TODO: consider using `bubbles/key` for key.Matches()
		// https://github.com/antonmedv/walk/blob/ba821ed78f31e0ebd46eeef19cfe642fc1ec4330/main.go#L252

		switch msg.String() {

		case "ctrl+w":
			b.input = "" // TODO: delete last word
			b.cursor = 0 // fzf resets cursor pos
			b.updateSearch(msg)
			return b, nil

		case "ctrl+c", "esc", "ctrl+\\":
			// os.Exit(0) // ungraceful exit
			// return nil, tea.Quit // bad pointer!

			// allow just going back to Queue
			if b.noquit {
				return queueBrowser(), nil
			}

			log.Println("quitting")
			return b, tea.Quit // graceful exit

		case "backspace":
			if len(b.input) > 0 {
				b.input = b.input[:len(b.input)-1]
				b.updateSearch(msg)
			}

		case "up", "ctrl+k":
			b.cursor--
			if b.cursor < 0 {
				b.cursor = len(b.matches) - 1
			}

		case "pgup":
			b.cursor = 0
			b.offset -= b.height

		case "pgdown":
			b.cursor = 0
			b.offset += b.height

		case "down", "ctrl+j":
			b.cursor++
			if b.cursor > len(b.matches)-1 {
				b.cursor = 0
			}

			// if b.cursor > b.height {
			// 	b.offset = b.cursor - b.height
			// }

		// https://github.com/antonmedv/walk/blob/ba821ed78f31e0ebd46eeef19cfe642fc1ec4330/main.go#L259 (?)
		case "enter":
			if len(b.matches) == 0 {
				return b, nil // do nothing
			}

			return b.getNewState()

		}
		// default:
		// 	return b, nil
	}

	// panic("unreachable")
	return b, nil
} // }}}

func (b *Browser) getNewState() (*Browser, tea.Cmd) {
	pos := b.matches[b.cursor]
	sel := b.items[pos] // relpath
	switch b.mode {
	case Artists:
		// return albumsBrowser(sel), nil
		return albumsBrowser(sel), tea.ClearScreen

	case Queue:

		// note: we need to split artist here (even though we do it
		// again in `play`)
		ensure(strings.Contains(sel, "/"))
		artist := strings.Split(sel, "/")[0]

		nb := albumsBrowser(artist)

		// allow user to back out of the selection without quitting the
		// program
		nb.noquit = true

		// sel will be removed from queue after play, but we need to
		// preempt that removal here
		nb.queued[sel] = false

		// after `play`, start View in Albums mode
		// TODO: detect if sel was deleted (in which case, go back to Queue)
		return nb, play(sel)

	case Albums:
		// add selected item to queue (if it exists), then go back to
		// Queue
		if _, err := os.Stat(filepath.Join(config.Library.Root, sel)); err == nil {
			q := getQueue(0)
			nq := append(q, sel)
			ensure(len(nq)-len(q) == 1)
			writeQueue(nq)
			log.Println("queued:", sel)
		}

		return queueBrowser(), nil

	default:
		panic("Invalid state")
	}
}

// Split screen into 2 vertical panes, with preview window on right
func (b *Browser) View() string {
	// The TUI is not very appealing, but this is ~by design~, as 1) I
	// really don't care about styling, 2) most of the time is spent in
	// mpv, and 3) the program is meant to just get out of the way and not
	// be distracting.
	//
	// [input]
	// [item1]|[preview1]
	// [item2]|[preview2]
	// ...    |...
	// (where | represents the border)

	// log.Println("view:", b)

	// https://github.com/charmbracelet/bubbletea/blob/master/examples/split-editors/main.go

	if len(b.matches) == 0 {
		return "no matches; please clear input"
	}

	sel := b.items[b.matches[b.cursor]]

	// TODO: another struct field?
	enu := func(_ list.Items, index int) string {
		return IsSelected[index == b.cursor]
	}
	leftItems := list.New().Enumerator(enu)

	anyQueued := b.mode == Albums && anyValue(b.queued)

	for i, idx := range b.matches {
		if i < b.offset {
			continue
		}
		item := b.items[idx] // idx is the actual index that points to the item

		switch {
		case anyQueued:
			base := path.Base(item)
			item = IsQueued[b.queued[item]] + " " + base
			leftItems.Item(item) // inplace

		case b.mode == Albums:
			base := path.Base(item)
			leftItems.Item(base)

		default:
			leftItems.Item(item)
		}
	}

	rightItems := list.New().Enumerator(func(_ list.Items, _ int) string { return "" })
	preview, ok := b.previews[sel]

	p := filepath.Join(config.Library.Root, sel)
	previews, err := descend(p)

	switch {
	case ok: // usually only in Albums mode
		rightItems.Items(preview)
	case err != nil:
		rightItems.Item("error")
	default:
		rightItems.Items(previews)
	}

	panes := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().
			// Inline(true). // forces multiline string back to single line
			Width(b.width*3/5). // takes priority over right
			// important: Height prioritises displaying last items,
			// MaxHeight prioritises displaying first items
			MaxHeight(b.height-3).
			Render(leftItems.String()),
		lipgloss.NewStyle().
			MaxHeight(b.height-3). // should always be MaxHeight, never Height
			BorderLeft(true).      // if omitted, assumes full border
			BorderStyle(lipgloss.NormalBorder()).
			Render(rightItems.String()),
	)

	return lipgloss.JoinVertical(lipgloss.Left, b.input, panes)
}
