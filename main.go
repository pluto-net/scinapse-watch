package main

import (
	"scinapse-watch/twitter"
	"os"
	"log"
	"path/filepath"
	"encoding/json"
)

func checkAndMakingLogDirectory() {
	_, err := os.Stat("./logs")
	if err != nil && os.IsNotExist(err) {
		os.Mkdir("./logs", 0777)
	}
}

func checkAndMakingTwitterLogFile(twitterFilePath string) *os.File {
	_, err := os.Stat(twitterFilePath)
	if err != nil && os.IsNotExist(err) {
		twFile, createErr := os.Create(twitterFilePath)
		if createErr != nil {
			log.Fatal("Failed to create log file")
		}

		return twFile
	} else {
		twFile, err := os.Open(twitterFilePath)
		if err != nil {
			log.Fatal(err)
		}
		return twFile
	}

}

func main() {
	checkAndMakingLogDirectory()

	newTwitts := twitter.Crawl()

	twitterFilePath, fileError := filepath.Abs("./logs/twitter.json")
	if fileError != nil {
		log.Fatal("can not assign file path")
	}

	twFile := checkAndMakingTwitterLogFile(twitterFilePath)
	defer twFile.Close()

	var oldTwitts = make([]*twitter.TwitItem, 0)

	dec := json.NewDecoder(twFile)
	decodeError := dec.Decode(&oldTwitts)

	if decodeError != nil {
		if decodeError.Error() != "EOF" {
			log.Fatal(decodeError)
		}
	}

	if len(oldTwitts) != len(newTwitts) {
		enc := json.NewEncoder(twFile)
		err := enc.Encode(newTwitts)

		if err != nil {
			log.Fatal(err)
		}

		oldTwitts = newTwitts
	}
}