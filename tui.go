package main

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// no error-handling
func LibraryRoot() string {
	// TODO: lazy_static equivalent?
	return os.Getenv("MU")
}

func GetQueue(n int) []string {
	// TODO: symlink file to here
	queueFile := os.ExpandEnv("$HOME/dita/dita/play/queue.txt")
	b, err := os.ReadFile(queueFile)
	if err != nil {
		panic(err)
	}
	items := strings.Split(string(b), "\n")
	// https://stackoverflow.com/a/12267471
	for i := range items {
		j := rand.Intn(i + 1)
		items[i], items[j] = items[j], items[i]
	}
	paths := []string{}
	root := LibraryRoot()
	for _, x := range items[:n] {
		p := filepath.Join(root, x)
		paths = append(paths, p)
	}
	return paths
}

type Mode int

const (
	Queue Mode = iota
	Artists
	Albums
)

// mostly copied from https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics

type model struct {
	// base   string
	mode Mode
	// guaranteed to be valid fullpaths
	items  []string
	cursor int
	play   bool
}

// func initialModel(base string, maxItems int) model {

// all items must be valid fullpaths
func newModel(items []string, mode Mode) model {
	info, err := os.Stat(items[0])
	if err != nil {
		panic(err)
	}
	if !info.IsDir() {
		panic("not dir")
	}

	return model{
		// base:  base,
		items: items,
		mode:  mode,
		play:  true,
	}
}

func (m model) Init() tea.Cmd {
	// not sure if this is a good idea
	return tea.ClearScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { // {{{
	tea.ClearScreen()

	selected := m.items[m.cursor] // nolint:all // this var is only used in a few branches

	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q", "esc":
			// os.Exit(0) // ungraceful exit
			// return nil, tea.Quit // bad pointer
			return m, tea.Quit // graceful exit

		// TODO: pgup/down

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		// https://github.com/antonmedv/walk/blob/ba821ed78f31e0ebd46eeef19cfe642fc1ec4330/main.go#L259 (?)
		case "enter":
			switch m.mode {
			case Artists:
				return newModel(descend(selected, true), Albums),
					tea.ClearScreen

			case Queue:
				// clear screen (?)
				// play album (mpv)
				// rate
				// remove from queue
				// then prepare new model from queue
				// play(selected)

				return newModel(GetQueue(10), Queue), play(selected)

			case Albums:

				if m.play { // uncommon in real use
					return newModel(GetQueue(10), Queue), play(selected)
				} else {
					// tea.Println("queued", selected)
					// TODO: append to queue file
					return m, tea.Sequence(tea.ClearScreen, tea.Quit)
				}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
} // }}}

// base should always be a valid absolute path
//
// returns fullpaths of immediate children if join is true (otherwise basenames)
func descend(base string, join bool) []string {
	entries, err := os.ReadDir(base)
	if err != nil {
		panic(err)
	}
	ch := []string{}
	for _, e := range entries {
		if join {
			ch = append(ch, filepath.Join(base, e.Name()))
		} else {
			ch = append(ch, e.Name())
		}
	}

	return ch
}

func (m model) View() string {
	// split screen into 2 vertical panes, place preview window on right
	// https://github.com/charmbracelet/bubbletea/blob/master/examples/split-editors/main.go
	// this does not look good at all, but it works for now
	// TODO: pane separator, fixed pane widths

	root := LibraryRoot()
	lines := []string{}
	for i, item := range m.items {

		cursor := " "
		if m.cursor == i {
			cursor = "→"
		}

		// TODO: condition should be "if relpath found in queue"
		checked := " "
		// if _, ok := m.selected[i]; ok {
		// 	checked = "x"
		// }

		var base string
		if m.mode == Queue {
			base, _ = filepath.Rel(root, item)
		} else {
			base = path.Base(item)
		}

		lines = append(lines, fmt.Sprintf("%s [%s] %s", cursor, checked, base))
	}
	left := strings.Join(lines, "\n")

	right := strings.Join(descend(m.items[m.cursor], false), "\n")

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right) //+ "\n\n"

	// s := "Select:\n\n"
	// // s += "\nPress q to quit.\n"
	// return s
}
