package slack

import (
	"bytes"
	"net/http"
	"io/ioutil"
	"scinapse-watch/twitter"
	"encoding/json"
	"log"
	"fmt"
	"os"
)

type TwitPayload struct {
	Text string `json:"text"`
}


func SendTwitterInformation(newTwitt *twitter.TwitItem) {
	twitSlackUrl := os.Getenv("TWIT_SLACK_URL")
	dstUrl := twitSlackUrl
	link := fmt.Sprintf("<https://twitter.com/%s>", newTwitt.Link)

	referUrls := ""
	if len(newTwitt.DesLinks) > 0 {
		for _, link := range newTwitt.DesLinks {
			if len(referUrls) > 0 {
				referUrls = referUrls + ", " + link
			} else {
				referUrls = link
			}

		}

	}

	textContent := fmt.Sprintf("`user`%s `link`(%s): `Referenced` %s", newTwitt.Username, link, referUrls )

	payload := TwitPayload{ Text: textContent}

	jsonStr, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}
	buf := bytes.NewBuffer(jsonStr)

	resp, err := http.Post(dstUrl, "application/json", buf)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if err == nil {
		str := string(respBody)
		println(str)
	}
}