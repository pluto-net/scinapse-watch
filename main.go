package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pluto-net/scinapse-watch/slack"
	"github.com/pluto-net/scinapse-watch/twitter"
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

	newTwits := twitter.Crawl()
	oldTwits := getOldTwits(logFilePath)

	os.Remove(logFilePath)

	if len(oldTwits) != len(newTwits) {
		if len(oldTwits) > 0 {
			sendNewComingTwits(oldTwits, newTwits)
		}

		f := openOrCreateLogFile(logFilePath)

		enc := json.NewEncoder(f)
		err = enc.Encode(newTwits)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getOldTwits(path string) []*twitter.TwitItem {
	twFile := openOrCreateLogFile(path)
	defer twFile.Close()

	var oldTwits = make([]*twitter.TwitItem, 0)
	info, err := twFile.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if info.Size() > 0 {
		dec := json.NewDecoder(twFile)
		err = dec.Decode(&oldTwits)
		if err != nil {
			log.Fatal(err)
		}
	}

	return oldTwits
}

func sendNewComingTwits(oldTwits []*twitter.TwitItem, newTwitts []*twitter.TwitItem) {
	var newcomerTwits []*twitter.TwitItem
	var oldTimeStamp int64

	i, err := strconv.ParseInt(oldTwits[0].Timestamp[6:], 10, 64)
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
			newcomerTwits = append(newcomerTwits, twit)
		}
	}

	if len(newcomerTwits) > 0 {
		for _, twit := range newcomerTwits {
			slack.SendTwitterInformation(twit)
		}
	}
}
