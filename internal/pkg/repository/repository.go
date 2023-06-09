package repository

import (
	"github.com/jmoiron/sqlx"
	courses "mom"
	"mom/internal/libraries/logging"
	"mom/internal/libraries/storage"
	"time"
)

type Authorization interface {
	CreateUser(user courses.User) (int, error)
	GetUser(login string) (courses.User, error)
	CreateSession(userId int, sessionId string, expiredDate time.Time) (int, error)
	DeleteSession(sessionId string) error
	CheckSession(sessionId string) (int, error)
	CheckEmail(email string) (int, error)
	CheckVkId(VkId int) (int, error)
	UpdatePassword(email, password string, salt int) (int, error)
	DeleteExpiredSessions() error

	AddCode(key string, val int)
	GetCode(key string) (int, bool)
	DeleteCode(key string) error
}

type Administration interface {
	CreateSession(userId int, sessionId string, expiredDate time.Time) (int, error)
	CheckSession(sessionId string) (int, error)
	SelectAllUsers() ([]courses.User, error)

	AddCode(key string, val int)
	GetCode(key string) (int, bool)
	DeleteCode(key string) error
}

type CourseList interface {
}

type CourseItem interface {
}

type Repository struct {
	Authorization
	Administration
	CourseList
	CourseItem
}

func NewRepository(db *sqlx.DB, storage *storage.DataStorage, logger *logging.Logger) *Repository {
	return &Repository{
		Authorization:  NewAuthPostgres(db, storage, logger),
		Administration: NewAdminPostgres(db, storage, logger),
	}
}
