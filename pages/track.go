package pages

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	gosxnotifier "github.com/deckarep/gosx-notifier"
	uniswap "github.com/hirokimoto/uniswap-api"
	unitrade "github.com/hirokimoto/uniswap-api/swap"
	unitrades "github.com/hirokimoto/uniswap-api/swaps"
	"github.com/hirokimoto/uniswap-auto-gui/data"
	"github.com/hirokimoto/uniswap-auto-gui/services"
	"github.com/skratchdot/open-golang/open"
)

var oldNames []string
var oldPrices []float64
var oldChanges []float64
var oldDurations []float64
var oldTransactions []string

func trackScreen(_ fyne.Window) fyne.CanvasObject {
	var selected uniswap.Swaps

	pairs := data.ReadTrackPairs()
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
		func() (int, int) { return len(pairs), 9 },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			switch id.Col {
			case 0:
				label.SetText(fmt.Sprintf("%d", id.Row+1))
			case 1:
				label.SetText("#")
			case 2:
				label.SetText("<")
			case 3:
				label.SetText(">")
			case 4:
				if len(oldNames[id.Row]) > 30 {
					label.SetText(oldNames[id.Row][0:30] + "...")
				} else {
					label.SetText(oldNames[id.Row])
				}
			case 5:
				label.SetText(fmt.Sprintf("%f", oldPrices[id.Row]))
			case 6:
				label.SetText(fmt.Sprintf("%.2f%%", 100*oldChanges[id.Row]/oldPrices[id.Row]))
			case 7:
				label.SetText(fmt.Sprintf("%f", oldDurations[id.Row]))
			case 8:
				label.SetText("-")
			default:
			}
		})
	table.SetColumnWidth(0, 40)
	table.SetColumnWidth(1, 30)
	table.SetColumnWidth(2, 25)
	table.SetColumnWidth(3, 25)
	table.SetColumnWidth(4, 250)
	table.SetColumnWidth(5, 100)
	table.SetColumnWidth(6, 100)
	table.SetColumnWidth(7, 100)
	table.SetColumnWidth(8, 30)
	table.OnSelected = func(id widget.TableCellID) {
		pair := pairs[id.Row]
		if id.Col == 0 {
			open.Run(fmt.Sprintf("https://www.dextools.io/app/ether/pair-explorer/%s", pair))
		}
		if id.Col == 1 {

		}
		if id.Col == 2 {
			if id.Row > 0 {
				temp := pairs[id.Row-1]
				pairs[id.Row-1] = pairs[id.Row]
				pairs[id.Row] = temp

				trackPair(pairs[id.Row-1], id.Row-1, records, table)
				trackPair(pairs[id.Row], id.Row, records, table)
			}
		}
		if id.Col == 3 {
			if id.Row < len(pairs)-1 {
				temp := pairs[id.Row+1]
				pairs[id.Row+1] = pairs[id.Row]
				pairs[id.Row] = temp

				trackPair(pairs[id.Row], id.Row, records, table)
				trackPair(pairs[id.Row+1], id.Row+1, records, table)
			}
		}
		if id.Col == 2 || id.Col == 3 {
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
		if id.Col == 5 {
			showSettings(pair)
		}
		if id.Col == 8 {
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

func trackPair(pair string, index int, records [][]string, table *widget.Table) {
	var swaps uniswap.Swaps
	cc := make(chan string, 1)
	go uniswap.SwapsByCounts(cc, 2, pair)

	msg := <-cc
	json.Unmarshal([]byte(msg), &swaps)

	if len(swaps.Data.Swaps) == 0 || swaps.Data.Swaps == nil {
		time.Sleep(time.Second * 1)
		return
	}

	if swaps.Data.Swaps[0].Id != oldTransactions[index] {
		n := unitrade.Name(swaps.Data.Swaps[0])
		p, _ := unitrade.Price(swaps.Data.Swaps[0])
		_, c := unitrades.WholePriceChanges(swaps)
		_, _, d := unitrades.Duration(swaps)

		oldNames[index] = n
		oldChanges[index] = c
		oldDurations[index] = d
		oldTransactions[index] = swaps.Data.Swaps[0].Id

		_, _, coin, usd := data.ReadSetting(records, pair)
		gas := 100.0
		if coin != 0.0 && usd != 0.0 {
			earn := p*coin - usd - gas
			oldNames[index] += fmt.Sprintf("(%.2f$)", earn)
		}

		if oldPrices[index] != 0.0 || c/p > 0.1 {
			go alert(records, pair, n, p, c, d)
		}
		oldPrices[index] = p

		table.Refresh()
	}
}

func alert(records [][]string, pair string, n string, p float64, c float64, d float64) {
	message := fmt.Sprintf("%s: %f %f%% %f", n, p, 100*c/p, d)
	title := "Priced Up!"
	if c < 0 {
		title = "Priced Down!"
	}
	link := fmt.Sprintf("https://www.dextools.io/app/ether/pair-explorer/%s", pair)

	min, max, coin, usd := data.ReadSetting(records, pair)
	gas := 100.0
	sound := gosxnotifier.Morse

	earn := p*coin - usd - gas
	if p < min {
		title = fmt.Sprintf("Warning Low! Watch %s", n)
		sound = gosxnotifier.Default
	}
	if p > max {
		title = fmt.Sprintf("Warning High! Watch %s", n)
		sound = gosxnotifier.Default
	}

	if coin != 0.0 && usd != 0.0 {
		if earn > 0.0 {
			percent := 100 * earn / usd
			if percent > 0.0 && percent < 5.0 {
				title = fmt.Sprintf("You can sell(+%f$)! Watch %s", earn, n)
			} else if percent > 5.0 && percent < 10.0 {
				title = fmt.Sprintf("Sell(+%f$)! Watch %s", earn, n)
			} else if percent > 10.0 {
				title = fmt.Sprintf("Sell sell sell(+%f$) %s!", earn, n)
			}
		} else {
			title += fmt.Sprintf("   %f$", earn)
		}
	}

	services.Alert(title, message, link, sound)
}

func showSettings(pair string) {
	records, _ := data.ReadTrackSettings()
	w := fyne.CurrentApp().NewWindow("Settings")

	min, max, coin, usd := data.ReadSetting(records, pair)
	mindata := binding.BindFloat(&min)
	minLabel := widget.NewLabel("Minimum")
	minEntry := widget.NewEntryWithData(binding.FloatToString(mindata))
	minPanel := container.NewGridWithColumns(2, minLabel, minEntry)

	maxdata := binding.BindFloat(&max)
	maxLabel := widget.NewLabel("Maximum")
	maxEntry := widget.NewEntryWithData(binding.FloatToString(maxdata))
	maxPanel := container.NewGridWithColumns(2, maxLabel, maxEntry)

	coindata := binding.BindFloat(&coin)
	coinLabel := widget.NewLabel("Coin")
	coinEntry := widget.NewEntryWithData(binding.FloatToString(coindata))
	coinPanel := container.NewGridWithColumns(2, coinLabel, coinEntry)

	usddata := binding.BindFloat(&usd)
	usdLabel := widget.NewLabel("USD")
	usdEntry := widget.NewEntryWithData(binding.FloatToString(usddata))
	usdPanel := container.NewGridWithColumns(2, usdLabel, usdEntry)

	btnSave := widget.NewButton("Save", func() {
		data.SaveTrackSettings(pair, min, max, coin, usd)
	})

	settingsPanel := container.NewVBox(minPanel, maxPanel, coinPanel, usdPanel, btnSave)
	w.SetContent(settingsPanel)

	w.Resize(fyne.NewSize(340, 200))
	w.SetFixedSize(true)
	w.Show()
}
