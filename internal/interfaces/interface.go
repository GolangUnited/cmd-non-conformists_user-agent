package interfaces

import (
	"github.com/golang-unitied-school/useragent/internal/models"
)

type EmptyUser struct{}

type UserDataManager interface {
	Init(connectionString string)
	Create(user *models.User) (string, error)
	Update(uuid, fname, sname, email, role string) error
	Delete(userId string) error
	GetById(userId string) (models.User, error)
	GetByEmail(email string) (models.User, error)
	GetPassword(userId string) (string, error)
	SetPassword(userId, newPass string) error
	Close() error
}
