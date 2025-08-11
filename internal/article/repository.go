package article

import (
	"github.com/cheildo/deeli-api/pkg/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository defines the interface for article and rating database operations.
type Repository interface {
	CreateArticle(article *Article) error
	GetArticlesByUserID(userID uint, page, limit int) ([]Article, error)
	GetArticleByIDAndUserID(articleID, userID uint) (*Article, error)
	GetArticleByID(articleID uint) (*Article, error)
	UpdateArticle(article *Article) error
	DeleteArticle(articleID, userID uint) error
	GetFailedArticlesToRetry(maxRetries int) ([]Article, error)
	CreateOrUpdateRating(rating *Rating) error
	GetRating(articleID, userID uint) (*Rating, error)
	DeleteRating(articleID, userID uint) error
	GetHighlyRatedArticleIDsForUser(userID uint, minScore int) ([]uint, error)
	FindPeerUsers(userID uint, articleIDs []uint, minScore int) ([]uint, error)
	GetHighlyRatedArticlesByUsers(userIDs []uint, minScore int) ([]Rating, error)
	GetArticleIDsSavedByUser(userID uint) ([]uint, error)
	GetArticlesByIDs(articleIDs []uint) ([]Article, error)
}

type repository struct{}

func NewRepository() Repository {
	return &repository{}
}

func (r *repository) CreateArticle(article *Article) error {
	return database.DB.Create(article).Error
}

func (r *repository) GetArticlesByUserID(userID uint, page, limit int) ([]Article, error) {
	var articles []Article
	offset := (page - 1) * limit
	err := database.DB.Where("user_id = ?", userID).Order("created_at desc").Offset(offset).Limit(limit).Find(&articles).Error
	return articles, err
}

func (r *repository) GetArticleByIDAndUserID(articleID, userID uint) (*Article, error) {
	var article Article
	err := database.DB.Where("id = ? AND user_id = ?", articleID, userID).First(&article).Error
	return &article, err
}

func (r *repository) GetArticleByID(articleID uint) (*Article, error) {
	var article Article
	err := database.DB.First(&article, articleID).Error
	return &article, err
}

func (r *repository) UpdateArticle(article *Article) error {
	return database.DB.Save(article).Error
}

func (r *repository) DeleteArticle(articleID, userID uint) error {
	result := database.DB.Where("id = ? AND user_id = ?", articleID, userID).Delete(&Article{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // No article found with that ID for that user
	}
	return nil
}

func (r *repository) GetFailedArticlesToRetry(maxRetries int) ([]Article, error) {
	var articles []Article
	err := database.DB.Where("status = ? AND retry_count < ?", StatusFailed, maxRetries).Find(&articles).Error
	return articles, err
}

// This will INSERT a new rating, or if a rating with the same
// user_id and article_id already exists, it will UPDATE the score.
func (r *repository) CreateOrUpdateRating(rating *Rating) error {
	return database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "article_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"score"}),
	}).Create(rating).Error
}

func (r *repository) GetRating(articleID, userID uint) (*Rating, error) {
	var rating Rating
	err := database.DB.Where("article_id = ? AND user_id = ?", articleID, userID).First(&rating).Error
	return &rating, err
}

func (r *repository) DeleteRating(articleID, userID uint) error {
	result := database.DB.Where("article_id = ? AND user_id = ?", articleID, userID).Delete(&Rating{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *repository) GetHighlyRatedArticleIDsForUser(userID uint, minScore int) ([]uint, error) {
	var articleIDs []uint
	err := database.DB.Model(&Rating{}).
		Where("user_id = ? AND score >= ?", userID, minScore).
		Pluck("article_id", &articleIDs).Error
	return articleIDs, err
}

func (r *repository) FindPeerUsers(userID uint, articleIDs []uint, minScore int) ([]uint, error) {
	var peerIDs []uint
	if len(articleIDs) == 0 {
		return peerIDs, nil
	}
	err := database.DB.Model(&Rating{}).
		Distinct("user_id").
		Where("user_id != ? AND article_id IN ? AND score >= ?", userID, articleIDs, minScore).
		Pluck("user_id", &peerIDs).Error
	return peerIDs, err
}

func (r *repository) GetHighlyRatedArticlesByUsers(userIDs []uint, minScore int) ([]Rating, error) {
	var ratings []Rating
	if len(userIDs) == 0 {
		return ratings, nil
	}
	err := database.DB.Model(&Rating{}).
		Where("user_id IN ? AND score >= ?", userIDs, minScore).
		Find(&ratings).Error
	return ratings, err
}

func (r *repository) GetArticleIDsSavedByUser(userID uint) ([]uint, error) {
	var articleIDs []uint
	err := database.DB.Model(&Article{}).
		Where("user_id = ?", userID).
		Pluck("id", &articleIDs).Error
	return articleIDs, err
}

func (r *repository) GetArticlesByIDs(articleIDs []uint) ([]Article, error) {
	var articles []Article
	if len(articleIDs) == 0 {
		return articles, nil
	}
	err := database.DB.Where("id IN ?", articleIDs).Find(&articles).Error
	return articles, err
}
