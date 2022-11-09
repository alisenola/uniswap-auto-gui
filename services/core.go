package services

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"

	"fyne.io/fyne/v2"
	gosxnotifier "github.com/deckarep/gosx-notifier"
	uniswap "github.com/hirokimoto/uniswap-api"
)

func UniswapMarkketPairs(target chan<- []uniswap.Pair) {
	var pairs []uniswap.Pair
	skip := 0

	for {
		var wg sync.WaitGroup
		wg.Add(1)

		cc := make(chan string, 1)
		go uniswap.RequestPairs(cc, 1000, 1000*skip)
		msg := <-cc
		var data uniswap.Pairs
		json.Unmarshal([]byte(msg), &data)
		counts := len(data.Data.Pairs)
		fmt.Println("Got ", counts, "...")
		if counts == 0 {
			fmt.Println("Completed to get pairs ", len(pairs), ".....")
			target <- pairs
			defer wg.Done()
			return
		}
		pairs = append(pairs, data.Data.Pairs...)
		fmt.Println(skip, "* 1000 pairs is coming...")
		skip += 1

		defer wg.Done()
	}
}

func Notify(title string, message string) {
	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   title,
		Content: message,
	})
}

func Alert(title string, message string, link string, sound gosxnotifier.Sound) {
	if runtime.GOOS == "windows" {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   title,
			Content: message,
		})
	} else {
		note := gosxnotifier.NewNotification(message)
		note.Title = title
		note.Sound = sound
		note.Link = link
		note.Push()
	}
}
