package article

import (
	"log"
	"net/http"
	"strconv"

	"github.com/cheildo/deeli-api/pkg/scraper"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler holds the repository dependency.
type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// CreateArticleRequest defines the expected JSON for creating an article.
type CreateArticleRequest struct {
	URL string `json:"url" binding:"required,url"`
}

// CreateArticle handles POST /articles
func (h *Handler) CreateArticle(c *gin.Context) {
	var req CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(uint)

	article := &Article{
		URL:    req.URL,
		UserID: userID,
		Status: StatusPending,
	}

	if err := h.repo.CreateArticle(article); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Article from this URL already exists for this user"})
		return
	}

	// Start scraping in a background goroutine so the API returns immediately.
	go func() {
		log.Printf("Starting initial scrape for article ID %d", article.ID)
		scrapedData, err := scraper.ScrapeMetadata(article.URL)
		if err != nil {
			log.Printf("Scrape failed for article ID %d: %v", article.ID, err)
			article.Status = StatusFailed
		} else {
			article.Title = scrapedData.Title
			article.Description = scrapedData.Description
			article.ImageURL = scrapedData.ImageURL
			article.Status = StatusCompleted
		}

		if err := h.repo.UpdateArticle(article); err != nil {
			log.Printf("Failed to update article ID %d after scrape: %v", article.ID, err)
		}
		log.Printf("Finished initial scrape for article ID %d with status %s", article.ID, article.Status)
	}()

	c.JSON(http.StatusAccepted, article)
}

// GetArticles handles GET /articles
func (h *Handler) GetArticles(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	articles, err := h.repo.GetArticlesByUserID(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve articles"})
		return
	}

	c.JSON(http.StatusOK, articles)
}

// DeleteArticle handles DELETE /articles/:id
func (h *Handler) DeleteArticle(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	if err := h.repo.DeleteArticle(uint(articleID), userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found or you don't have permission to delete it"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete article"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// RateArticleRequest defines the JSON for rating an article.
type RateArticleRequest struct {
	Score int `json:"score" binding:"required,min=1,max=5"`
}

// RateArticle handles POST /articles/:id/rate
func (h *Handler) RateArticle(c *gin.Context) {
	var req RateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(uint)
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	if _, err := h.repo.GetArticleByIDAndUserID(uint(articleID), userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found or you don't own it"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	rating := &Rating{
		UserID:    userID,
		ArticleID: uint(articleID),
		Score:     req.Score,
	}

	if err := h.repo.CreateOrUpdateRating(rating); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save rating"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rating saved successfully"})
}

// GetRating handles GET /articles/:id/rate
func (h *Handler) GetRating(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	rating, err := h.repo.GetRating(uint(articleID), userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "No rating found for this article"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rating"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"score": rating.Score})
}

// DeleteRating handles DELETE /articles/:id/rate
func (h *Handler) DeleteRating(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	if err := h.repo.DeleteRating(uint(articleID), userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rating not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rating"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
