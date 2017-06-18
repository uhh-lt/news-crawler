package sources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

func Run() error {
	urls := crawlFeedDirectories()
	feeds, err := scrapeFeedURLs(urls)
	if err != nil {
		return err
	}

	feeds = validateFeeds(feeds)
	fmt.Printf("Found %d feeds\n", len(feeds))
	err = store(feeds)
	return err
}

func store(feeds []string) error {
	feedsJSON, err := json.Marshal(feeds)
	if err != nil {
		return err
	}
	ioutil.WriteFile("feeds.json", feedsJSON, 0644)
	return nil
}

func crawlFeedDirectories() []string {
	urls := make([]string, 0)

	// http://www.rss-verzeichnis.net/
	for i := 1; i < 54; i++ {
		url := fmt.Sprintf("http://www.rss-verzeichnis.net/nachrichten-page%d.htm", i)
		urls = append(urls, url)
	}

	return urls
}

func scrapeFeedURLs(urls []string) ([]string, error) {
	feedURLs := make([]string, 0)
	for _, url := range urls {
		pageHTML, err := fetchURL(url)
		if err != nil {
			return nil, err
		}
		urls := extractFeedURLs(pageHTML)
		feedURLs = append(feedURLs, urls...)
	}

	feedURLs = unique(feedURLs)
	return feedURLs, nil
}

func fetchURL(url string) (string, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer rsp.Body.Close()
	html, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	return string(html), nil
}

func extractFeedURLs(html string) []string {
	feedReg := regexp.MustCompile(`(https?:\/\/([-\w\.]+)+(:\d+)?(\/([\w\/_\.]*(\?\S+)?)?)?(feed|rss)+([\w\/_\.\-]*(\?\S+)?)?)`)
	feeds := feedReg.FindAllString(html, -1)
	return feeds
}

func validateFeeds(feeds []string) []string {
	validFeeds := make([]string, 0)

	for _, feed := range feeds {
		timeout := time.Duration(30 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		req, err := http.NewRequest("GET", feed, nil)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		defer resp.Body.Close()

		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			continue
		}

		validFeeds = append(validFeeds, feed)
		fmt.Printf("Validated feed %s", feed)
	}

	return validFeeds
}

func unique(s []string) []string {
	uniq := make([]string, 0, len(s))
	seen := make(map[string]bool)

	for _, val := range s {
		if _, ok := seen[val]; ok {
			continue
		}
		seen[val] = true
		uniq = append(uniq, val)
	}

	return uniq
}