package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Mode int

const (
	Queue Mode = iota
	Artists
	Albums
)

// mostly copied from https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics

type Browser struct {
	mode   Mode
	width  int // calculated dynamically (?)
	height int // calculated dynamically (?)

	// if false, album will be queued
	play bool

	// should be valid relpaths
	items   []string
	offset  int
	cursor  int
	input   string
	matches []int

	// only used in Albums mode
	queue map[string]any
}

// all items must be valid relpaths (relative to root)
func newBrowser(items []string, mode Mode) Browser {
	// TODO: only pass mode and optional artist arg (items can be
	// auto-generated for Queue and Artists)

	info, err := os.Stat(filepath.Join(config.Library.Root, items[0]))
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

func artistBrowser() Browser {
	items, _ := descend(config.Library.Root)
	return newBrowser(items, Artists)
}

// https://github.com/antonmedv/walk/blob/ba821ed78f31e0ebd46eeef19cfe642fc1ec4330/main.go#L427
// note the pointer; we are mutating Browser

func (b *Browser) updateSearch(msg tea.KeyMsg) {
	switch b.input {
	case "":
		b.matches = intRange(len(b.items))
	default:
		var matchIdxs []int
		for i, rel := range b.items {
			// // this -should- always work...
			// rel, _ := filepath.Rel(config.Library.Root, item)
			// TODO: if b.input has ' ', split and match each word
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

// tea.Model interface; the required methods cannot use pointer receivers

func (_ Browser) Init() tea.Cmd {
	// not sure if this is needed
	return tea.ClearScreen
}

func (b Browser) Update(msg tea.Msg) (tea.Model, tea.Cmd) { // {{{

	switch msg := msg.(type) {

	// https://github.com/charmbracelet/bubbletea/discussions/818#discussioncomment-6914769
	case tea.WindowSizeMsg: // only triggered when window resized?
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
			b.cursor = 0 // fzf resets cursor pos
			b.updateSearch(msg)
			return b, nil

		case "ctrl+c", "esc", "ctrl+\\":
			// os.Exit(0) // ungraceful exit
			// return nil, tea.Quit // bad pointer
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

			// if b.cursor > b.height-1 {
			// 	// b.cursor = len(b.matches) - 1
			// 	// b.offset = b.cursor //- b.height
			// 	b.offset = b.height
			// 	// b.cursor = 0
			// }
			// // return b, nil

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

			// if b.cursor > len(b.matches) {
			// 	return b, nil
			// }
			// idx := b.matches[b.cursor]
			sel := b.items[b.matches[b.cursor]]

			switch b.mode {
			case Artists:
				albums, err := descend(filepath.Join(config.Library.Root, sel))
				var relpaths []string
				for _, alb := range albums {
					relpaths = append(relpaths, filepath.Join(sel, alb))
				}
				if err != nil {
					panic(err)
				}
				nb := newBrowser(relpaths, Albums)

				// init window correctly; a "recursively"
				// spawned Browser leaves b.height zeroed!
				nb.height = b.height
				nb.width = b.width

				// TODO: set .queue
				// q := make(map[string]any)
				// GetQueue(-1)
				// nb.queue = q

				return nb, tea.ClearScreen

			case Queue:
				return newBrowser(getQueue(10), Queue), play(sel)

			case Albums:
				if b.play { // uncommon in real use
					return newBrowser(getQueue(10), Queue), play(sel)
				} else {
					// tea.Println("queued", selected)
					// TODO: append to queue file
					return b, tea.Sequence(tea.ClearScreen, tea.Quit)
				}
			}
		}
	}

	return b, nil
} // }}}

func (b Browser) View() string {
	// split screen into 2 vertical panes, with preview window on right
	// https://github.com/charmbracelet/bubbletea/blob/master/examples/split-editors/main.go

	lines := []string{}

	// log.Println(b.cursor)

	for i, idx := range b.matches {

		if i < b.offset {
			continue
		}

		cursor := " "
		if b.cursor == i { // 'raw' 0-based index
			// cursor = "→" // messes with truncation, presumably because len > 1
			cursor = ">"
		}

		item := b.items[idx] // idx is the actual index that points to the item
		var base string
		if b.mode == Queue {
			base = item
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

		// log.Println(i, idx, item, line)
		lines = append(lines, line)

		// why -3? not sure...
		if i > b.height-3 {
			break
		}
	}
	// log.Println(len(lines), "lines, height", b.height)
	left := strings.Join(lines, "\n")
	// return left
	// sep := strings.Repeat(" │ \n", max(0, b.height-1))

	var s string
	if len(b.matches) > 0 {
		sel := b.items[b.matches[b.cursor]]
		p := filepath.Join(config.Library.Root, sel)
		previewItems, err := descend(p)
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
