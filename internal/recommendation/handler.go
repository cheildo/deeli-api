package recommendation

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetRecommendations handles the GET /recommendations request.
func (h *Handler) GetRecommendations(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	recommendations, err := h.service.GetRecommendationsForUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate recommendations"})
		return
	}

	c.JSON(http.StatusOK, recommendations)
}
