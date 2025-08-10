package worker

import (
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

// ScrapedData holds the metadata extracted from a URL
type ScrapedData struct {
	Title       string
	Description string
	ImageURL    string
}

// ScrapeMetadata fetches a URL and extracts OpenGraph metadata
func ScrapeMetadata(url string) (*ScrapedData, error) {
	res, err := http.Get(url)
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

	log.Printf("Scraped: Title='%s', Desc='%s', Img='%s'", data.Title, data.Description, data.ImageURL)
	return data, nil
}
