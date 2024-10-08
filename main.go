package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// TODO: cli flags OR tui main menu [queue/artists/albums]; the latter only
// makes sense once i have a good grasp of panes etc

func main() {
	// TODO: zerolog
	// https://github.com/kubefirst/kubefirst/blob/49665b715b1899887e82454a920506f493ba5448/main.go#L118
	lf, _ := tea.LogToFile("/tmp/tea.log", "plaque")
	defer lf.Close()

	// browseArtists(discogsSearchArtist("rira")).rate()
	// return

	// WithAltScreen should always be used, to avoid janky rendering
	var p tea.Model
	switch mpvRunning() {
	case true:
		p = artistBrowser()
	case false:
		p = queueBrowser()
	}

	if _, err := tea.NewProgram(p, tea.WithAltScreen()).Run(); err != nil {
		panic(err)
	}
}
