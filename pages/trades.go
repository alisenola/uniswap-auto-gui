package pages

import (
	"encoding/json"
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	regression "github.com/gaillard/go-online-linear-regression/v1"
	uniswap "github.com/hirokimoto/uniswap-api"
	unitrade "github.com/hirokimoto/uniswap-api/swap"
	unitrades "github.com/hirokimoto/uniswap-api/swaps"
	"github.com/hirokimoto/uniswap-auto-gui/data"
	"github.com/hirokimoto/uniswap-auto-gui/services"
	"github.com/skratchdot/open-golang/open"
)

type Trade struct {
	Pair     uniswap.Pair
	Swaps    uniswap.Swaps
	Name     string
	Price    float64
	Duration float64
	Status   string
}

func tradesScreen(_ fyne.Window) fyne.CanvasObject {
	var pairs []uniswap.Pair
	trades := map[string]Trade{}
	var actives []uniswap.Pair

	table := widget.NewTable(
		func() (int, int) { return len(actives), 5 },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			pair := actives[id.Row]
			switch id.Col {
			case 0:
				label.SetText(fmt.Sprintf("%d", id.Row+1))
			case 1:
				label.SetText(trades[pair.Id].Name)
			case 2:
				label.SetText(fmt.Sprintf("%f", trades[pair.Id].Price))
			case 3:
				label.SetText(fmt.Sprintf("%.2f", trades[pair.Id].Duration))
			case 4:
				label.SetText(trades[pair.Id].Status)
			}
		})
	table.SetColumnWidth(0, 60)
	table.SetColumnWidth(1, 250)
	table.SetColumnWidth(2, 150)
	table.SetColumnWidth(3, 100)
	table.SetColumnWidth(4, 100)
	table.OnSelected = func(id widget.TableCellID) {
		pair := actives[id.Row]
		if id.Col == 0 {
			open.Run(fmt.Sprintf("https://www.dextools.io/app/ether/pair-explorer/%s", pair.Id))
		}
	}

	infProgress := widget.NewProgressBarInfinite()
	infProgress.Start()

	find := widget.NewButton("Fetching Pairs", func() {
		infProgress.Start()

		go func() {
			for index, v := range pairs {
				var wg sync.WaitGroup
				wg.Add(1)

				var s uniswap.Swaps

				sc := make(chan string, 1)
				go uniswap.SwapsByCounts(sc, 1000, v.Id)
				msg := <-sc
				json.Unmarshal([]byte(msg), &s)

				if len(s.Data.Swaps) > 0 {
					n := unitrade.Name(s.Data.Swaps[0])
					p, _ := unitrade.Price(s.Data.Swaps[0])
					_, c := unitrades.WholePriceChanges(s)
					_, _, d := unitrades.Duration(s)
					old, _ := unitrade.Old(s.Data.Swaps[0])
					average := unitrades.AveragePrice(s.Data.Swaps)

					// Filter our some tokens which is in the active trading in recent3 days.
					if old < 3*24 && p > 0.0001 {
						slope, _, _ := testRegression(s)
						var isGoingUp = slope > 0
						var isGoingDown = slope < 0
						updown := "up"
						if isGoingUp {
							updown = "up"
							fmt.Println("Trending up token ", n, p, average, c, d)
						}
						if isGoingDown {
							updown = "down"
							fmt.Println("Trending down token ", n, p, average, c, d)
						}

						var t Trade
						t.Pair = v
						t.Swaps = s
						t.Name = n
						t.Price = p
						t.Duration = d
						t.Status = updown
						trades[v.Id] = t
						actives = append(actives, v)
						table.Refresh()
					}
				}
				fmt.Print(index, "|")

				defer wg.Done()
			}

			infProgress.Stop()
		}()
	})
	find.Disable()

	go func() {
		pc := make(chan []uniswap.Pair, 1)
		go services.UniswapMarkketPairs(pc)
		pairs = <-pc

		var list []string
		for _, item := range pairs {
			list = append(list, item.Id)
		}

		data.SaveTradePairs(list)
		find.Enable()
		find.SetText("Find Tradable Pairs")
		infProgress.Stop()
	}()

	controls := container.NewVBox(find, infProgress)
	return container.NewBorder(controls, nil, nil, nil, table)
}

func testRegression(swaps uniswap.Swaps) (float64, float64, float64) {
	r := regression.New(7)

	for i := 0; i < len(swaps.Data.Swaps); i++ {
		swap := swaps.Data.Swaps[i]
		price, _ := unitrade.Price(swap)
		r.Add(float64(i), price)
	}

	slope, intercept, stdError := r.CalculateWithStdError()
	return slope, intercept, stdError
}
