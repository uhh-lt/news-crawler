package scraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/thesoenke/news-crawler/feedreader"
)

const (
	userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:53.0) Gecko/20100101 Firefox/53.0"
)

type Scraper struct {
	Feeds []feedreader.Feed
}

func New(feedsFile string) (Scraper, error) {
	scraper := Scraper{}

	feeds, err := loadFeeds(feedsFile)
	if err != nil {
		return scraper, nil
	}

	scraper.Feeds = feeds
	return scraper, nil
}

// ScrapeURLs downloads the content of the provide list of urls
func (scraper *Scraper) Scrape() error {
	for _, feed := range scraper.Feeds {
		for _, feedItem := range feed.Items {
			page, err := fetchPage(feedItem.URL)
			if err != nil {
				return err
			}

			_, err = extractContent(feedItem.URL, page)
			if err != nil {
				return err
			}

			// err = Store(feedItem.URL, content)
			// if err != nil {
			// 	return err
			// }
		}
	}

	return nil
}

func fetchPage(url string) (string, error) {
	timeout := time.Duration(30 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return "", fmt.Errorf("Site retuned status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func loadFeeds(path string) ([]feedreader.Feed, error) {
	articlesFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var feeds []feedreader.Feed
	err = json.Unmarshal(articlesFile, &feeds)
	if err != nil {
		return nil, err
	}

	var urls []string
	for _, feed := range feeds {
		for _, feedItem := range feed.Items {
			urls = append(urls, feedItem.URL)
		}
	}

	return feeds, nil
}