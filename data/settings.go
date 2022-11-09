package data

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
)

func SaveTrackSettings(address string, min float64, max float64, coin float64, usd float64) {
	filePath := absolutePath() + "/settings.csv"
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
	}
	defer f.Close()

	csvWriter := csv.NewWriter(f)

	oldRecords, _ := ReadTrackSettings()
	var newRecords [][]string

	isAdded := false
	for _, record := range oldRecords {
		caddress := string(record[0])
		cmin := string(record[1])
		cmax := string(record[2])
		ccoin := string(record[3])
		cusd := string(record[4])

		if address == caddress {
			isAdded = true
			newRecords = append(newRecords, []string{caddress, fmt.Sprintf("%f", min), fmt.Sprintf("%f", max), fmt.Sprintf("%f", coin), fmt.Sprintf("%f", usd)})
		} else {
			newRecords = append(newRecords, []string{caddress, cmin, cmax, ccoin, cusd})
		}
	}
	if !isAdded {
		newRecords = append(newRecords, []string{address, fmt.Sprintf("%f", min), fmt.Sprintf("%f", max), fmt.Sprintf("%f", coin), fmt.Sprintf("%f", usd)})
	}

	err = csvWriter.WriteAll(newRecords)

	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	if err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Failed saving settings!",
		})
		return
	}

	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   "Success",
		Content: "Saved settings successfully!",
	})
}

func ReadTrackSettings() ([][]string, error) {
	filePath := absolutePath() + "/settings.csv"
	f, err := os.Open(filePath)
	if err != nil {
		return [][]string{}, nil
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()

	if err != nil {
		return [][]string{}, nil
	}

	return records, nil
}

func ReadSetting(records [][]string, address string) (float64, float64, float64, float64) {
	min := 0.0
	max := 0.0
	coin := 0.0
	usd := 0.0
	for _, record := range records {
		caddress := string(record[0])
		if address == caddress {
			min, _ = strconv.ParseFloat(record[1], 64)
			max, _ = strconv.ParseFloat(record[2], 64)
			coin, _ = strconv.ParseFloat(record[3], 64)
			usd, _ = strconv.ParseFloat(record[4], 64)
		}
	}
	return min, max, coin, usd
}
