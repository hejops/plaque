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

	// b := Browser{
	// 	items: []string{"A", "B", "C"},
	// 	previews: map[string][]string{
	// 		"A": {"1"},
	// 		"B": {"2"},
	// 		"C": {"3"},
	// 	},
	// }
	// if _, err := tea.NewProgram(
	// 	&b,
	// 	tea.WithAltScreen(),
	// ).Run(); err != nil {
	// 	panic(err)
	// }
	// return

	if _, err := tea.NewProgram(
		queueBrowser(true),
		// &Browser{},
		tea.WithAltScreen(),
	).Run(); err != nil {
		panic(err)
	}

	// if _, err := tea.NewProgram(artistBrowser()).Run(); err != nil {
	// 	panic(err)
	// }
}
