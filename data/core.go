package data

import (
	"bufio"
	"fmt"
	"os"
)

func WriteOnePair(pair string) error {
	err := writeOnePair(pair)
	return err
}

func RemoveOnePair(pair string) error {
	err := removeOnePair(pair)
	return err
}

func addOnePair(pair string, filename string) error {
	path := absolutePath() + filename
	pairs, _ := readLines(path)
	_pairs := []string{}
	for _, v := range pairs {
		if v != pair {
			_pairs = append(_pairs, v)
		}
	}
	err := writeLines(_pairs, path)
	return err
}

func writeOnePair(pair string) error {
	path := absolutePath() + "/pairs.txt"
	pairs, _ := readLines(path)
	pairs = append(pairs, pair)
	err := writeLines(pairs, path)
	return err
}

func removeOnePair(pair string) error {
	path := absolutePath() + "/pairs.txt"
	pairs, _ := readLines(path)
	_pairs := []string{}
	for _, v := range pairs {
		if v != pair {
			_pairs = append(_pairs, v)
		}
	}
	fmt.Println(pairs)
	fmt.Println(_pairs)
	err := writeLines(_pairs, path)
	return err
}

func absolutePath() string {
	path, _ := os.UserCacheDir()
	return path
}

func readLines(path string) ([]string, error) {
	var lines []string
	file, err := os.Open(path)
	if err != nil {
		writeLines(lines, path)
		return lines, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}
