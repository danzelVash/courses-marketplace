package service

import (
	courses "mom"
	"mom/internal/libraries/logging"
	"mom/internal/pkg/repository"
	"time"
)

var cleanersNum int

type Cleaner interface {
	CleanSessionStorage()
}

type SessionCleaner struct {
	repos  repository.Authorization
	logger *logging.Logger
}

func NewSessionCleaner(repos repository.Authorization, logger *logging.Logger) *SessionCleaner {
	return &SessionCleaner{
		repos:  repos,
		logger: logger,
	}
}

func (sc *SessionCleaner) deleteExpiredSessions() error {
	return sc.repos.DeleteExpiredSessions()
}

func (sc *SessionCleaner) CleanSessionStorage() {
	if cleanersNum == 0 {
		cleanersNum++
		ticker := time.Tick(time.Hour * 24)
		for now := range ticker {
			err := sc.deleteExpiredSessions()
			if err != nil {
				sc.logger.Errorf("error while deleting expired sessions in session_cleaner: %s", err.Error())
			} else {
				sc.logger.Infof("session storage was cleaned at %v", now)
			}
		}
	}
	sc.logger.Error(courses.CleanerAlreadyRunning)
}
