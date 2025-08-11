package scraper

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ScrapedData holds the metadata extracted from a URL
type ScrapedData struct {
	Title       string
	Description string
	ImageURL    string
}

// ScrapeMetadata fetches a URL and extracts OpenGraph or standard metadata.
func ScrapeMetadata(url string) (*ScrapedData, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	data := &ScrapedData{}

	data.Title = doc.Find("meta[property='og:title']").AttrOr("content", "")
	if data.Title == "" {
		data.Title = doc.Find("title").First().Text()
	}

	data.Description = doc.Find("meta[property='og:description']").AttrOr("content", "")
	if data.Description == "" {
		data.Description = doc.Find("meta[name='description']").AttrOr("content", "")
	}

	data.ImageURL = doc.Find("meta[property='og:image']").AttrOr("content", "")

	data.Title = strings.TrimSpace(data.Title)
	data.Description = strings.TrimSpace(data.Description)
	data.ImageURL = strings.TrimSpace(data.ImageURL)

	log.Printf("Scraped from %s: Title='%s'", url, data.Title)
	return data, nil
}
