package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const QueuedSymbol = "Q"

func intRange(n int) []int {
	ints := []int{}
	for i := range n {
		ints = append(ints, i)
	}
	return ints
}

// no error-handling
func LibraryRoot() string {
	// TODO: read from config file
	return os.Getenv("MU")
}

// Select n random items from the queue file (containing relpaths), and return
// them as fullpaths
func GetQueue(n int) []string {
	// TODO: symlink file to here
	queueFile := os.ExpandEnv("$HOME/dita/dita/play/queue.txt")
	b, err := os.ReadFile(queueFile)
	if err != nil {
		panic(err)
	}
	relpaths := strings.Split(string(b), "\n")
	// TODO: split off sampling
	// https://stackoverflow.com/a/12267471
	for i := range relpaths {
		j := rand.Intn(i + 1)
		relpaths[i], relpaths[j] = relpaths[j], relpaths[i]
	}
	paths := []string{}
	root := LibraryRoot()
	if n < 0 {
		panic("not impl yet")
	}
	for _, rel := range relpaths[:n] {
		p := filepath.Join(root, rel)
		paths = append(paths, p)
	}
	return paths
}

// base should always be a valid absolute path
//
// returns fullpaths of immediate children if join is true (otherwise basenames)
func descend(base string, join bool) ([]string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return []string{}, err
		// panic(err)
	}
	ch := []string{}
	for _, e := range entries {
		if join {
			ch = append(ch, filepath.Join(base, e.Name()))
		} else {
			ch = append(ch, e.Name())
		}
	}
	return ch, nil
}

type Mode int

const (
	Queue Mode = iota
	Artists
	Albums
)

// mostly copied from https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics

type Browser struct {
	mode   Mode
	width  int
	height int
	play   bool

	// guaranteed to be valid fullpaths
	items   []string
	cursor  int
	input   string
	matches []int

	// only used in Albums mode
	queue map[string]any
}

// all items must be valid fullpaths
func newBrowser(items []string, mode Mode) Browser {
	// TODO: only pass mode and optional artist arg (items can be
	// auto-generated for Queue and Artists)

	info, err := os.Stat(items[0])
	if err != nil {
		panic(err)
	}
	if !info.IsDir() {
		panic("not dir")
	}

	return Browser{
		items:   items,
		mode:    mode,
		play:    true,
		matches: intRange(len(items)),
	}
}

func (_ Browser) Init() tea.Cmd {
	// not sure if this is a good idea
	return tea.ClearScreen
}

// https://github.com/antonmedv/walk/blob/ba821ed78f31e0ebd46eeef19cfe642fc1ec4330/main.go#L427
// important: we are mutating Browser

func (b *Browser) updateSearch(msg tea.KeyMsg) {
	if b.input == "" {
		b.matches = intRange(len(b.items))
	} else {
		root := LibraryRoot()
		matchIdxs := []int{}
		for i, item := range b.items {
			rel, _ := filepath.Rel(root, item)
			if strings.Contains(strings.ToLower(rel), b.input) {
				matchIdxs = append(matchIdxs, i)
			}
		}
		b.matches = matchIdxs
		if len(b.matches) > 0 {
			b.cursor = 0
		}
	}
}

func (b Browser) Update(msg tea.Msg) (tea.Model, tea.Cmd) { // {{{
	// tea.ClearScreen()

	switch msg := msg.(type) {

	// https://github.com/charmbracelet/bubbletea/discussions/818#discussioncomment-6914769
	case tea.WindowSizeMsg:
		b.width = msg.Width
		b.height = msg.Height

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
			b.input = ""
			b.cursor = 0 // does fzf reset cursor pos?
			b.updateSearch(msg)
			return b, nil

		case "ctrl+c", "esc":
			// os.Exit(0) // ungraceful exit
			// return nil, tea.Quit // bad pointer
			// fmt.Println(b.input)
			return b, tea.Quit // graceful exit

		case "backspace":
			if len(b.input) > 0 {
				b.input = b.input[:len(b.input)-1]
				b.updateSearch(msg)
			}

		case "up", "ctrl+k":
			if b.cursor > 0 {
				b.cursor--
			} else {
				b.cursor = len(b.matches) - 1
			}

		case "down", "ctrl+j":
			if b.cursor < len(b.matches)-1 {
				b.cursor++
			} else {
				b.cursor = 0
			}

		// https://github.com/antonmedv/walk/blob/ba821ed78f31e0ebd46eeef19cfe642fc1ec4330/main.go#L259 (?)
		case "enter":

			if b.cursor > len(b.matches) {
				return b, nil
			}
			idx := b.matches[b.cursor]
			selected := b.items[idx]

			switch b.mode {
			case Artists:
				items, err := descend(selected, true)
				exec.Command("notify-send", items[0]).Run()
				if err != nil {
					panic(err)
				}
				nb := newBrowser(items, Albums)

				// q := make(map[string]any)
				// GetQueue(-1)
				// nb.queue = q

				return nb,
					tea.ClearScreen

			case Queue:
				return newBrowser(GetQueue(10), Queue),
					play(selected)

			case Albums:
				if b.play { // uncommon in real use
					return newBrowser(GetQueue(10), Queue),
						play(selected)
				} else {
					// tea.Println("queued", selected)
					// TODO: append to queue file
					return b,
						tea.Sequence(tea.ClearScreen, tea.Quit)
				}
			}

		}

	}

	return b, nil
} // }}}

func (b Browser) View() string {
	// split screen into 2 vertical panes, with preview window on right
	// https://github.com/charmbracelet/bubbletea/blob/master/examples/split-editors/main.go

	root := LibraryRoot()
	lines := []string{}
	for i, idx := range b.matches {

		cursor := " "
		if b.cursor == i { // 'raw' 0-based index
			// cursor = "→" // messes with truncation, presumably because len > 1
			cursor = ">"
		}

		// exec.Command("notify-send", strconv.Itoa(idx)).Run()
		item := b.items[idx] // idx is the actual index that points to the item

		var base string
		if b.mode == Queue {
			base, _ = filepath.Rel(root, item)
		} else {
			base = path.Base(item)
		}

		// if b.mode == Albums {
		// 	_, queued := b.queue[item]
		// 	if queued {
		// 		base = QueuedSymbol + " " + base
		// 	}
		// }

		line := fmt.Sprintf("%s %s", cursor, base)

		// TODO: ellipsis properly
		// upper bound (half of term width)
		// note: lower bound is not enforced (i.e. not fixed width)
		if b.width > 0 && len(line) > b.width/2 {
			line = line[:b.width/2] + "..."
		}
		lines = append(lines, line)
	}
	left := strings.Join(lines, "\n")
	// return left
	// sep := strings.Repeat(" │ \n", max(0, b.height-1))

	var s string
	if len(b.matches) > 0 {
		sel := b.items[b.matches[b.cursor]]
		previewItems, err := descend(sel, false)
		// exec.Command("notify-send", fmt.Sprintf("%v", previewItems)).Run()
		if err != nil {
			previewItems = []string{"error"}
		} else if b.height > 0 && len(previewItems) > b.height {
			previewItems = previewItems[:max(0, b.height-1)]
			// previewItems = previewItems[:b.height]
		}
		right := strings.Join(previewItems, "\n")
		s = lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	}

	return lipgloss.JoinVertical(lipgloss.Left, b.input, s)
	// return lipgloss.PlaceHorizontal(25, lipgloss.Top, s)
}
