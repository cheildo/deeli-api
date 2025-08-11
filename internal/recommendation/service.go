package recommendation

import (
	"log"
	"sort"

	"github.com/cheildo/deeli-api/internal/article"
)

const (
	minRatingForRecommendation = 4
	recommendationLimit        = 10
)

// Service provides the recommendation logic.
type Service interface {
	GetRecommendationsForUser(userID uint) ([]article.Article, error)
}

type service struct {
	articleRepo article.Repository
}

// NewService creates a new recommendation service.
func NewService(articleRepo article.Repository) Service {
	return &service{articleRepo: articleRepo}
}

// GetRecommendationsForUser implements the collaborative filtering logic.
func (s *service) GetRecommendationsForUser(userID uint) ([]article.Article, error) {
	// 1. Get all articles the current user has rated highly.
	userFavorites, err := s.articleRepo.GetHighlyRatedArticleIDsForUser(userID, minRatingForRecommendation)
	if err != nil {
		log.Printf("Error getting user favorites for user %d: %v", userID, err)
		return nil, err
	}

	// Handle cold start: If the user has no favorite articles, we can't find peers.
	// A good fallback would be to return globally popular articles, but for now, we'll return empty.
	if len(userFavorites) == 0 {
		log.Printf("User %d has no highly-rated articles. Cannot generate personalized recommendations.", userID)
		return []article.Article{}, nil
	}

	// 2. Find "peer" users who also liked the same articles.
	peers, err := s.articleRepo.FindPeerUsers(userID, userFavorites, minRatingForRecommendation)
	if err != nil {
		log.Printf("Error finding peers for user %d: %v", userID, err)
		return nil, err
	}

	if len(peers) == 0 {
		log.Printf("No peers found for user %d. Cannot generate recommendations.", userID)
		return []article.Article{}, nil
	}

	// 3. Get all the favorite articles of those peers.
	peerRatings, err := s.articleRepo.GetHighlyRatedArticlesByUsers(peers, minRatingForRecommendation)
	if err != nil {
		log.Printf("Error getting peer ratings for user %d: %v", userID, err)
		return nil, err
	}

	// 4. Get all articles the current user has already saved, to filter them out.
	userSavedArticles, err := s.articleRepo.GetArticleIDsSavedByUser(userID)
	if err != nil {
		log.Printf("Error getting user's saved articles for user %d: %v", userID, err)
		return nil, err
	}

	// Create a map for quick lookups of articles the user has already saved.
	userSavedMap := make(map[uint]bool)
	for _, id := range userSavedArticles {
		userSavedMap[id] = true
	}

	// 5. Count the frequency of each article recommended by peers.
	recommendationScores := make(map[uint]int)
	for _, rating := range peerRatings {
		// If the article is not one the user has already saved...
		if !userSavedMap[rating.ArticleID] {
			// ...increment its recommendation score.
			recommendationScores[rating.ArticleID]++
		}
	}

	// 6. Sort the recommendations by score (frequency).
	type recommendation struct {
		ArticleID uint
		Score     int
	}

	var sortedRecs []recommendation
	for id, score := range recommendationScores {
		sortedRecs = append(sortedRecs, recommendation{ArticleID: id, Score: score})
	}

	sort.Slice(sortedRecs, func(i, j int) bool {
		return sortedRecs[i].Score > sortedRecs[j].Score // Sort descending
	})

	// 7. Get the top N article IDs.
	var finalArticleIDs []uint
	for i := 0; i < len(sortedRecs) && i < recommendationLimit; i++ {
		finalArticleIDs = append(finalArticleIDs, sortedRecs[i].ArticleID)
	}

	if len(finalArticleIDs) == 0 {
		return []article.Article{}, nil
	}

	// 8. Fetch the full article details for the top IDs.
	return s.articleRepo.GetArticlesByIDs(finalArticleIDs)
}
