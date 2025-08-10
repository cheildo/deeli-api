package user

import "github.com/cheildo/deeli-api/pkg/database"

type Repository interface {
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id uint) (*User, error)
}

type repository struct{}

func NewRepository() Repository {
	return &repository{}
}

func (r *repository) CreateUser(user *User) error {
	return database.DB.Create(user).Error
}

func (r *repository) GetUserByEmail(email string) (*User, error) {
	var user User
	err := database.DB.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *repository) GetUserByID(id uint) (*User, error) {
	var user User
	err := database.DB.First(&user, id).Error
	return &user, err
}
