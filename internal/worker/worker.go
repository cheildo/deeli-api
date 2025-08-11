package worker

import (
	"log"
	"time"

	"github.com/cheildo/deeli-api/internal/article"
	"github.com/cheildo/deeli-api/pkg/scraper"
)

const maxRetries = 3

// Worker holds dependencies for the background job processor.
type Worker struct {
	articleRepo article.Repository
}

func NewWorker(repo article.Repository) *Worker {
	return &Worker{articleRepo: repo}
}

// Start runs the background worker loop. It should be called in a goroutine.
func (w *Worker) Start() {
	log.Println("Starting background worker...")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	w.processFailedArticles()

	for range ticker.C {
		log.Println("Worker tick: checking for failed metadata scrapes...")
		w.processFailedArticles()
	}
}

// processFailedArticles is the core logic that is run periodically.
func (w *Worker) processFailedArticles() {
	articlesToRetry, err := w.articleRepo.GetFailedArticlesToRetry(maxRetries)
	if err != nil {
		log.Printf("Worker error fetching articles to retry: %v", err)
		return
	}

	if len(articlesToRetry) == 0 {
		log.Println("Worker: No articles to retry.")
		return
	}

	log.Printf("Worker: Found %d articles to retry.", len(articlesToRetry))
	for _, art := range articlesToRetry {
		articleToProcess := art

		log.Printf("Retrying article ID %d (URL: %s), attempt #%d", articleToProcess.ID, articleToProcess.URL, articleToProcess.RetryCount+1)
		scrapedData, err := scraper.ScrapeMetadata(articleToProcess.URL)

		articleToProcess.RetryCount++
		if err != nil {
			log.Printf("Retry failed for article ID %d: %v", articleToProcess.ID, err)
		} else {
			log.Printf("Retry successful for article ID %d", articleToProcess.ID)
			articleToProcess.Title = scrapedData.Title
			articleToProcess.Description = scrapedData.Description
			articleToProcess.ImageURL = scrapedData.ImageURL
			articleToProcess.Status = article.StatusCompleted
		}

		if err := w.articleRepo.UpdateArticle(&articleToProcess); err != nil {
			log.Printf("Worker failed to update article ID %d after retry: %v", articleToProcess.ID, err)
		}
	}
}
