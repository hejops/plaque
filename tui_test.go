package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

// t.Run("mock", func(t *testing.T) {
// 	// setup mock dirs + cfg
// 	// invoke entry point (queue)
// 	// TODO: simulate keypresses?
// })

func checkModelOutput(
	t *testing.T,
	tm *teatest.TestModel,
	s string,
	waitOpts ...teatest.WaitForOption,
) {
	teatest.WaitFor(
		t,
		tm.Output(),
		func(bts []byte) bool {
			// if strings.Contains(string(bts), "a") {
			// 	fmt.Println(string(bts))
			// }
			if s == "" {
				return len(string(bts)) == 0
			}
			return strings.Contains(string(bts), s)
		},
		waitOpts...,
	)
}

func TestUIBasic(t *testing.T) {
	b := Browser{
		items: []string{"A", "B", "C"},
		previews: map[string][]string{
			"A": {"1"},
			"B": {"2"},
			"C": {"3"},
		},
	}

	tm := teatest.NewTestModel(t, &b)
	checkModelOutput(t, tm, "no matches; please clear input")

	// tm.Output 'clears' the TestModel, so for 'unit' tests that (naively)
	// modify the inner model directly, a new TestModel must be spawned
	// each time.

	b.matches = []int{0}
	tm = teatest.NewTestModel(t, &b)
	checkModelOutput(t, tm, "→ A│ 1")

	b.matches = []int{0, 1, 2}
	b.cursor = 1
	// typical size of my terminal
	tm = teatest.NewTestModel(t, &b, teatest.WithInitialTermSize(115, 15))
	checkModelOutput(
		t,
		tm,
		"  A                                                                  │ 2",
	)

	// on the other hand, for 'e2e' tests with tm.Send calls (which are
	// more realistic), the TestModel simply can be reused

	b.matches = []int{0, 1, 2}
	b.cursor = 0
	tm = teatest.NewTestModel(t, &b, teatest.WithInitialTermSize(115, 15))

	// cursor movement

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlJ})
	checkModelOutput(t, tm, "→ B")
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlJ})
	checkModelOutput(t, tm, "→ C")
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlJ})
	checkModelOutput(t, tm, "→ A")
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlK})
	checkModelOutput(t, tm, "→ C")

	// // pressing backspace seems to lead to erroneous failures
	// tm.Send(tea.KeyMsg{Type: tea.KeyBackspace})
	// checkModelOutput(t, tm, "→ C") // empty

	// filtering

	tm = teatest.NewTestModel(t, &b)
	tm.Type("b")
	checkModelOutput(t, tm, "→ B")
	tm.Send(tea.KeyMsg{Type: tea.KeyBackspace})
	checkModelOutput(t, tm, "→ A")
	tm.Type("x")
	checkModelOutput(t, tm, "no matches")
	tm.Send(tea.KeyMsg{Type: tea.KeyBackspace})
	checkModelOutput(t, tm, "→ A")

	// // doing a search that would return the already selected item also
	// // produces an erroneous failure
	// tm.Type("a")
	// checkModelOutput(t, tm, "→ A") // " {7}a"

	// tm.Send(tea.KeyMsg{Type: tea.KeyCtrlK})
	// checkModelOutput(t, tm, "→ C")
}
