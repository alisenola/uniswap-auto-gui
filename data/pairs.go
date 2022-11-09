package data

import (
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/hirokimoto/uniswap-auto-gui/services"
)

func ReadFavorites() []string {
	pairs := readPairs("/favorites.txt")
	return pairs
}

func SaveFavorites(pair string) {
	pairs := readPairs("/favorites.txt")
	pairs = append(pairs, pair)
	addOnePair(pair, "/pairs.txt")
	writePairs("/favorites.txt", pairs)
}

func ReadBabies() []string {
	pairs := readPairs("/babies.txt")
	return pairs
}

func SaveBabies(pair string) {
	pairs := readPairs("/babies.txt")
	pairs = append(pairs, pair)
	addOnePair(pair, "/pairs.txt")
	writePairs("/babies.txt", pairs)
}

func SaveTrackPairs(pairs []string) {
	writePairs("/pairs.txt", pairs)
}

func ReadTrackPairs() []string {
	pairs := readPairs("/pairs.txt")
	return pairs
}

func SaveTradePairs(pairs []string) {
	writePairs("/trades.txt", pairs)
}

func ReadTradePairs() []string {
	pairs := readPairs("/trades.txt")
	return pairs
}

func readPairs(file string) []string {
	path := absolutePath() + file
	pairs, err := readLines(path)

	if err != nil {
		fmt.Println(err)
		return []string{
			"0x7a99822968410431edd1ee75dab78866e31caf39",
			"0x3dd49f67e9d5bc4c5e6634b3f70bfd9dc1b6bd74"}
	}

	return pairs
}

func writePairs(file string, pairs []string) {
	path := absolutePath() + file
	err := writeLines(pairs, path)

	if err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Failed tracking pairs!",
		})
		return
	}

	services.Notify("Success", "Saved tracking pairs successfully!")
}
