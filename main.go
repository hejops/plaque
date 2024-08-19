package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// to be able to import functions from same package (need not be public), need:
// go run *.go
// https://stackoverflow.com/a/43953582
//
// to also avoid test.go files (cringe):
// ls *.go | grep -v _test | xargs go run
// note that this breaks stdin

// TODO: check resume
// TODO: cli flags
// TODO: tui main menu

func main() {
	lf, _ := tea.LogToFile("/tmp/tea.log", "plaque")
	defer lf.Close()

	if _, err := tea.NewProgram(
		newBrowser(getQueue(10), Queue),
		tea.WithAltScreen(),
	).Run(); err != nil {
		panic(err)
	}

	// if _, err := tea.NewProgram(artistBrowser()).Run(); err != nil {
	// 	panic(err)
	// }
}
