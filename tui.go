package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// mostly copied from https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics

type model struct {
	// navigable
	items []fs.DirEntry // TODO: generalise to any

	// not navigable
	previews []string
	cursor   int
	selected map[int]struct{}
}

func initialModel(items []fs.DirEntry) model {
	return model{
		items:    items,
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		// TODO: return item (instead of changing m state)
		case "enter": //, " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// split screen into 2 vertical panes, place preview window on right
	// https://github.com/charmbracelet/bubbletea/blob/master/examples/split-editors/main.go
	// this does not look good at all, but it works for now
	// TODO: pane separator, fixed pane widths?, left pane cursor

	lines := []string{}
	for i, item := range m.items {

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		// TODO: condition should be "if relpath found in queue"
		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		line := fmt.Sprintf(
			// "%s [%s] %s\n", cursor, checked, choice.Name(),
			"%s [%s] %s", cursor, checked, item.Name(),
		)

		lines = append(lines, line)
	}
	left := strings.Join(lines, "\n")

	subdirs, err := os.ReadDir(filepath.Join(library_root(), m.items[m.cursor].Name()))
	if err != nil {
		panic(err)
	}
	previews := []string{}
	for _, item := range subdirs {
		previews = append(previews, item.Name())
	}
	right := strings.Join(previews, "\n")

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right) //+ "\n\n"

	// s := "Select:\n\n"
	// // s += "\nPress q to quit.\n"
	// return s
}
