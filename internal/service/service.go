package service

import (
	"context"
	"io"

	"github.com/danzelVash/courses-marketplace"
	"github.com/danzelVash/courses-marketplace/internal/repository"
	"github.com/danzelVash/courses-marketplace/pkg/logging"
	"github.com/danzelVash/courses-marketplace/pkg/m3u8_generator"
)

type Authorization interface {
	CreateUser(user courses.User) (int, error)
	CreateSession(login, password string) (string, error)
	DeleteSession(sessionId string) error
	CheckSession(sessionId string) (int, error)
	CheckEmailAndSendCode(email string) error
	CheckCode(email string, code int) error
	UpdatePassword(email string) error
	CleanExpiredSessions()

	GenerateLinkForOauth() string
	ValidateParamsFromRedirect(stateTemp, code string) bool
	GenerateUrlForVKHandshake(code string) string
	GenerateUrlForVkApi(body io.ReadCloser) (string, error)
	AuthorizeVkUser(body io.ReadCloser) (string, error)
}

type Administration interface {
	SendAuthCode(admin courses.Administrator) error
	CreateAdminSession(inputCode int) (string, error)
	CheckAdminSession(sessionId string) (int, error)
	SelectAllUsers() ([]courses.User, error)

	CreateVideoLessonOnBucket(ctx context.Context, bytesChan <-chan *m3u8_generator.AnswerToCaller)
}

type CourseList interface {
}

type CourseItem interface {
}

type Service struct {
	Authorization
	Administration
	CourseList
	CourseItem
}

func NewService(repos *repository.Repository, logger *logging.Logger) *Service {
	return &Service{
		Authorization:  NewAuthService(repos.Authorization, logger),
		Administration: NewAdminService(repos.Administration, logger),
	}
}
