package pages

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	uniswap "github.com/hirokimoto/uniswap-api"
	unitrade "github.com/hirokimoto/uniswap-api/swap"
	"github.com/hirokimoto/uniswap-auto-gui/data"
	"github.com/skratchdot/open-golang/open"
)

func trackFavorites(_ fyne.Window) fyne.CanvasObject {
	var selected uniswap.Swaps

	pairs := data.ReadFavorites()
	records, _ := data.ReadTrackSettings()

	oldNames = make([]string, 0)
	oldPrices = make([]float64, 0)
	oldChanges = make([]float64, 0)
	oldDurations = make([]float64, 0)
	oldTransactions = make([]string, 0)

	for _, _ = range pairs {
		oldNames = append(oldNames, "")
		oldPrices = append(oldPrices, 0.0)
		oldChanges = append(oldChanges, 0.0)
		oldDurations = append(oldDurations, 0.0)
		oldTransactions = append(oldTransactions, "")
	}

	name := widget.NewEntry()
	name.SetPlaceHolder("0x385769E84B650C070964398929DB67250B7ff72C")
	append := widget.NewButton("Append", func() {
		if name.Text != "" {
			isExisted := false
			for _, item := range pairs {
				if item == name.Text {
					isExisted = true
				}
			}
			if !isExisted {
				pairs = append(pairs, name.Text)
				oldNames = append(oldNames, "")
				oldPrices = append(oldPrices, 0.0)
				oldChanges = append(oldChanges, 0.0)
				oldDurations = append(oldDurations, 0.0)
				oldTransactions = append(oldTransactions, "")
				data.SaveTrackPairs(pairs)
			}
		}
	})

	control := container.NewVBox(name, append)

	rightList := widget.NewList(
		func() int {
			return len(selected.Data.Swaps)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.DocumentIcon()), widget.NewLabel("target"), widget.NewLabel("price"), widget.NewLabel("amount"), widget.NewLabel("amount1"), widget.NewLabel("amount2"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if len(selected.Data.Swaps) > 1 {
				price, target, amount, amount1, amount2 := unitrade.Trade(selected.Data.Swaps[id])

				item.(*fyne.Container).Objects[1].(*widget.Label).SetText(target)
				item.(*fyne.Container).Objects[2].(*widget.Label).SetText(fmt.Sprintf("$%f", price))
				item.(*fyne.Container).Objects[3].(*widget.Label).SetText(amount)
				item.(*fyne.Container).Objects[4].(*widget.Label).SetText(amount1)
				item.(*fyne.Container).Objects[5].(*widget.Label).SetText(amount2)
			}
		},
	)

	table := widget.NewTable(
		func() (int, int) { return len(pairs), 8 },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			switch id.Col {
			case 0:
				label.SetText(fmt.Sprintf("%d", id.Row+1))
			case 1:
				label.SetText("<")
			case 2:
				label.SetText(">")
			case 3:
				if len(oldNames[id.Row]) > 30 {
					label.SetText(oldNames[id.Row][0:30] + "...")
				} else {
					label.SetText(oldNames[id.Row])
				}
			case 4:
				label.SetText(fmt.Sprintf("%f", oldPrices[id.Row]))
			case 5:
				label.SetText(fmt.Sprintf("%.2f%%", 100*oldChanges[id.Row]/oldPrices[id.Row]))
			case 6:
				label.SetText(fmt.Sprintf("%f", oldDurations[id.Row]))
			case 7:
				label.SetText("-")
			default:
			}
		})
	table.SetColumnWidth(0, 40)
	table.SetColumnWidth(1, 25)
	table.SetColumnWidth(2, 25)
	table.SetColumnWidth(3, 250)
	table.SetColumnWidth(4, 100)
	table.SetColumnWidth(5, 100)
	table.SetColumnWidth(6, 100)
	table.SetColumnWidth(7, 30)
	table.OnSelected = func(id widget.TableCellID) {
		pair := pairs[id.Row]
		if id.Col == 0 {
			open.Run(fmt.Sprintf("https://www.dextools.io/app/ether/pair-explorer/%s", pair))
		}
		if id.Col == 1 {
			if id.Row > 0 {
				temp := pairs[id.Row-1]
				pairs[id.Row-1] = pairs[id.Row]
				pairs[id.Row] = temp

				trackPair(pairs[id.Row-1], id.Row-1, records, table)
				trackPair(pairs[id.Row], id.Row, records, table)
			}
		}
		if id.Col == 2 {
			if id.Row < len(pairs)-1 {
				temp := pairs[id.Row+1]
				pairs[id.Row+1] = pairs[id.Row]
				pairs[id.Row] = temp

				trackPair(pairs[id.Row], id.Row, records, table)
				trackPair(pairs[id.Row+1], id.Row+1, records, table)
			}
		}
		if id.Col == 1 || id.Col == 2 {
			data.SaveTrackPairs(pairs)
			table.Refresh()
		}
		if id.Col == 3 {
			go func() {
				for {
					var swaps uniswap.Swaps
					cc := make(chan string, 1)
					go uniswap.SwapsByCounts(cc, 20, pair)

					msg := <-cc
					json.Unmarshal([]byte(msg), &swaps)

					selected = swaps
					rightList.Refresh()

					time.Sleep(time.Second * 1)
				}
			}()
		}
		if id.Col == 4 {
			showSettings(pair)
		}
		if id.Col == 7 {
			pairs[id.Row] = pairs[len(pairs)-1]
			pairs[len(pairs)-1] = ""
			pairs = pairs[:len(pairs)-1]
			data.SaveTrackPairs(pairs)
			table.Refresh()
		}
	}

	listPanel := container.NewBorder(nil, control, nil, nil, table)

	go func() {
		for {
			for index, pair := range pairs {
				var wg sync.WaitGroup
				wg.Add(1)
				fmt.Print(".")
				trackPair(pair, index, records, table)
				wg.Done()
			}
			time.Sleep(time.Second * 1)
		}
	}()

	return container.NewHSplit(listPanel, rightList)
}
