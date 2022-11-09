package pages

import (
	"fyne.io/fyne/v2"
)

// Page defines the data structure for a page
type Page struct {
	Title, Intro string
	View         func(w fyne.Window) fyne.CanvasObject
}

var (
	// Pages defines the metadata for each page
	Pages = map[string]Page{
		"welcome":   {"Welcome", "", welcomeScreen},
		"favorites": {"Favorites", "", trackFavorites},
		"babies":    {"Babies", "", trackBabies},
		"track":     {"Track", "", trackScreen},
		"trades":    {"Trades", "", tradesScreen},
	}

	// PageIndex  defines how our pages should be laid out in the index tree
	PageIndex = map[string][]string{
		"": {"welcome", "favorites", "babies", "track", "trades"},
	}
)
