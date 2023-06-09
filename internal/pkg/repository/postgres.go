package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

const (
	usersTable        = "users"
	courseTable       = "course"
	usersCoursesTable = "users_courses"
	itemTable         = "item"
	itemsCoursesTable = "items_courses"
	sessionTable      = "session"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDb(cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Open(
		"postgres",
		fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.SSLMode))

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
