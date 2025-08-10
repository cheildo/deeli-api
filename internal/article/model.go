package article

import "gorm.io/gorm"

type ArticleStatus string

const (
	StatusPending   ArticleStatus = "pending"
	StatusCompleted ArticleStatus = "completed"
	StatusFailed    ArticleStatus = "failed"
)

type Article struct {
	gorm.Model
	URL         string `gorm:"uniqueIndex:idx_user_url;not null"`
	Title       string
	Description string
	ImageURL    string
	Status      ArticleStatus `gorm:"default:'pending'"`
	RetryCount  int           `gorm:"default:0"`
	UserID      uint          `gorm:"uniqueIndex:idx_user_url;not null"`
}

type Rating struct {
	gorm.Model
	Score     int  `gorm:"not null"`
	ArticleID uint `gorm:"uniqueIndex:idx_user_article_rating;not null"`
	UserID    uint `gorm:"uniqueIndex:idx_user_article_rating;not null"`
}
