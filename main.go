package main

import (
	"scinapse-watch/twitter"
	"os"
	"log"
	"path/filepath"
	"encoding/json"
	"strconv"
	"scinapse-watch/slack"
)

func checkAndMakingLogDirectory() {
	_, err := os.Stat("./logs")
	if err != nil && os.IsNotExist(err) {
		os.Mkdir("./logs", 0777)
	}
}

func openOrCreateLogFile(twitterFilePath string) *os.File {
	twFile, err := os.OpenFile(twitterFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	return twFile
}

func main() {
	checkAndMakingLogDirectory()
	logFilePath, err := filepath.Abs("./logs/twitter.json")
	if err != nil {
		log.Fatal(err)
	}

	newTwitts := twitter.Crawl()
	oldTwitts := getOldTwitts(logFilePath)

	os.Remove(logFilePath)

	if len(oldTwitts) != len(newTwitts) {
		if len(oldTwitts) > 0 {
			sendNewComingTwitts(oldTwitts, newTwitts)
		}

		f := openOrCreateLogFile(logFilePath)


		enc := json.NewEncoder(f)
		err = enc.Encode(newTwitts)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getOldTwitts(path string) []*twitter.TwitItem {
	twFile := openOrCreateLogFile(path)
	defer twFile.Close()

	var oldTwitts = make([]*twitter.TwitItem, 0)
	info, err := twFile.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if info.Size() > 0 {
		dec := json.NewDecoder(twFile)
		err = dec.Decode(&oldTwitts)
		if err != nil {
			log.Fatal(err)
		}
	}

	return oldTwitts
}

func sendNewComingTwitts(oldTwitts []*twitter.TwitItem, newTwitts []*twitter.TwitItem) {
	var newcomerTwitts []*twitter.TwitItem
	var oldTimeStamp int64

	i, err := strconv.ParseInt(oldTwitts[0].Timestamp[6:], 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	oldTimeStamp = i
	for _, twit := range newTwitts {
		timestamp, err := strconv.ParseInt(twit.Timestamp[6:], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		if oldTimeStamp < timestamp {
			newcomerTwitts = append(newcomerTwitts, twit)
		}
	}

	if len(newcomerTwitts) > 0 {
		for _, twit := range newcomerTwitts {
			slack.SendTwitterInformation(twit)
		}
	}
}
