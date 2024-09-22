// TUI for discogs data

package discogs

import (
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
)

// Similar to Browser
type discogsBrowser struct {
	artists       []Artist
	releases      map[int]*[]Release
	releasesTable map[int]*table.Model
	cursor        int

	// releases map[*Artist]*[]Release
	// // pointers are non-unique! (and structs can't even be used as keys)

	width  int
	height int
}

// additional heuristics/tui will usually be required to select the correct
// artist; this is left to callers
func SearchArtist(artist string) []Artist {
	resp := makeReq(
		"/database/search",
		"GET",
		map[string]any{"q": alnum(artist), "type": "artist"},
	)
	return deserialize(resp, struct {
		Results []Artist
	}{}).Results
}

// Start a bubbletea program to browse artist discographies, and return
// selected artist. Returns nil if no selection was made.
//
// Panics if an empty slice is passed
func BrowseArtists(artists []Artist) *Artist {
	// // sort by artist name len (usually silly)
	// slices.SortFunc(artists, func(a Artist, b Artist) int {
	// 	an := len(a.Title)
	// 	bn := len(b.Title)
	//
	// 	switch {
	// 	case an < bn:
	// 		return -1
	// 	case an > bn:
	// 		return 1
	// 	default:
	// 		return 0
	// 	}
	// })

	// TODO: levenshtein

	eb := discogsBrowser{
		artists: artists,
		// make map necessary?
		releases:      make(map[int]*[]Release, len(artists)),
		releasesTable: make(map[int]*table.Model, len(artists)),
	}

	m, err := tea.NewProgram(&eb, tea.WithAltScreen()).Run()
	if err != nil {
		panic(err)
	}
	x := m.(*discogsBrowser)
	if x.cursor < 0 {
		return nil
	}
	art := x.artists[x.cursor]
	return &art
}

// Fetch n artists on each side of the cursor
func (db *discogsBrowser) getReleases(n int) {
	idxs := surround(db.cursor, len(db.artists)-1, n)
	for _, idx := range idxs {
		// check oob because surround does not constrain result slice
		// to the bounds of eb.artists
		if idx < 0 || idx >= len(db.artists) {
			continue
		}
		artist := db.artists[idx]
		releases := db.releases[artist.Id]
		if releases == nil {
			r := artist.Releases()
			db.releases[artist.Id] = &r

			time.Sleep(time.Second)

			var rows []table.Row

			for _, rel := range r {
				row := table.Row{strconv.Itoa(rel.Year), rel.Title}
				rows = append(rows, row)
			}

			// log.Println("rows:", rows)

			t := table.New(
				table.WithRows(rows),
				table.WithColumns(
					[]table.Column{
						{Title: "Year", Width: 4},
						{Title: "Title", Width: 50},
					},
				),
			)
			// log.Println("table:", t)
			db.releasesTable[artist.Id] = &t

		}

	}
}

// Init is the first function that will be called. It returns an optional
// initial command. To not perform an initial command return nil.
func (db *discogsBrowser) Init() tea.Cmd {
	// always populate map value for 1st artist
	db.getReleases(0)

	go db.getReleases(2)
	return nil
}

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.
func (db *discogsBrowser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		db.width = msg.Width
		db.height = msg.Height
		return db, tea.ClearScreen

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return db, tea.Quit

		case "esc", "q":
			db.cursor = -1
			return db, tea.Quit

		case "j":
			db.cursor++
			if db.cursor > len(db.artists)-1 {
				db.cursor = 0
			}
			go db.getReleases(2)

		case "k":
			db.cursor--
			if db.cursor < 0 {
				db.cursor = len(db.artists) - 1
			}
			go db.getReleases(2)

		}
	}
	return db, nil
}

var isSelected = map[bool]string{true: "→", false: " "}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (db *discogsBrowser) View() string {
	artists := []string{}
	for _, a := range db.artists {
		artists = append(artists, a.Title)
	}

	left := list.New(artists).Enumerator(func(items list.Items, index int) (arrow string) {
		arrow = isSelected[index == db.cursor]

		switch db.artists[index].UserData["in_collection"] {
		case true:
			arrow += " /"
		case false:
			arrow += "  "
		}

		return arrow
	})

	var right string
	if len(db.artists) == 0 {
		return "no artists"
	}
	t := db.releasesTable[db.artists[db.cursor].Id]
	if t == nil {
		right = "wait..."
	} else {
		right = t.View()
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().MaxHeight(db.height).Render(left.String()),
		lipgloss.NewStyle().MaxHeight(db.height).Render(right),
	)
}
