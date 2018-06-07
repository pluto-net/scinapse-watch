package twitter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type TwitItem struct {
	Content   string   `json:"content"`
	FullName  string   `json:"full_name"`
	Username  string   `json:"username"`
	Link      string   `json:"link"`
	Timestamp string   `json:"timestamp"`
	DesLinks  []string `json:"destination_link"`
}

func NewTwitItem() TwitItem {
	return TwitItem{DesLinks: make([]string, 0)}
}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf.String()
}

func getTwitItem(node *html.Node, twitItem *TwitItem) {
	if node.Type == html.ElementNode {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "fullname" {
				fullName := node.FirstChild.Data
				if len(fullName) > 0 {
					twitItem.FullName = fullName
				}

			}

			if attr.Key == "class" && attr.Val == "username" {
				userName := strings.TrimSpace(node.LastChild.Data)
				if len(userName) > 0 {
					twitItem.Username = userName
				}

			}

			if attr.Key == "data-url" && node.Data == "a" {
				if len(attr.Val) > 0 && !strings.Contains(attr.Val, "twitter.com") { // To avoid twitter image links
					twitItem.DesLinks = append(twitItem.DesLinks, attr.Val)
				}
			}

			if attr.Key == "class" && attr.Val == "tweet-text" {
				text := renderNode(node)

				if len(text) > 0 {
					twitItem.Content = text
				}
			}

			if attr.Key == "class" && attr.Val == "timestamp" {
				timeStamp := node.FirstChild.NextSibling.Attr[0].Val

				if len(timeStamp) > 0 {
					twitItem.Timestamp = timeStamp
				}
			}

			if attr.Key == "href" && node.Data == "table" {
				link := attr.Val

				if len(link) > 0 {
					twitItem.Link = link
				}
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		getTwitItem(child, twitItem)
	}
}

func findTwitTableNode(node *html.Node) (*html.Node, error) {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "timeline" {
				return node, nil
			}
		}

	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		nextNode, err := findTwitTableNode(child)
		if err == nil {
			return nextNode, err
		}
	}

	return nil, errors.New("No table to crawl")
}

func (t TwitItem) isValid() bool {
	return len(t.Content) > 0 && len(t.Username) > 0 && len(t.FullName) > 0 && len(t.Link) > 0
}

func parseTwitTable(node *html.Node) []*TwitItem {
	result := make([]*TwitItem, 0)

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		twitItem := NewTwitItem()
		getTwitItem(child, &twitItem)

		if twitItem.isValid() {
			result = append(result, &twitItem)
		}
	}

	return result
}

func Crawl() []*TwitItem {
	t := string(time.Now().Unix())
	resp, err := http.Get(fmt.Sprintf("https://mobile.twitter.com/search?q=scinapse.io&s=typd&x=0&y=0&t=%s", t))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	htmlResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	htmlString := string(htmlResponse)

	parsedHTML, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		log.Fatal(err)
	}

	node, err := findTwitTableNode(parsedHTML)
	if err != nil {
		log.Fatal(err)
	}

	return parseTwitTable(node)
}
