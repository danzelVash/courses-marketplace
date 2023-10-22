package repository

import (
	"database/sql"
	"fmt"

	"github.com/danzelVash/courses-marketplace"
	"github.com/danzelVash/courses-marketplace/pkg/logging"
	"github.com/danzelVash/courses-marketplace/pkg/storage"
	"github.com/jmoiron/sqlx"

	"time"
)

type AuthPostgres struct {
	db      *sqlx.DB
	storage *storage.DataStorage
	logger  *logging.Logger
}

func NewAuthPostgres(db *sqlx.DB, storage *storage.DataStorage, logger *logging.Logger) *AuthPostgres {
	return &AuthPostgres{
		db:      db,
		storage: storage,
		logger:  logger,
	}
}

func (a *AuthPostgres) AddCode(key string, val int) {
	a.storage.Add(key, val)
}

func (a *AuthPostgres) GetCode(key string) (int, bool) {
	return a.storage.Get(key)
}

func (a *AuthPostgres) DeleteCode(key string) error {
	if err := a.storage.Delete(key); err != nil {
		a.logger.Errorf("error while deleting admin auth code from storage: %s", err.Error())
		return err
	}
	return nil
}

func (a *AuthPostgres) CreateUser(user courses.User) (int, error) {
	query := fmt.Sprintf(
		"INSERT INTO %s (last_name, first_name, email, phone_number, password_hash, salt, vk, vk_id) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id;",
		usersTable)

	row := a.db.QueryRow(query, user.LastName, user.FirstName, user.Email, user.PhoneNumber, user.Password, user.Salt, user.Vk, user.VkId)

	var id int
	if err := row.Scan(&id); err != nil {
		a.logger.Errorf("error while scanning id in auth_pstgrs/CreateUser: %s", err.Error())
		return 0, err
	}

	return id, nil
}

func (a *AuthPostgres) GetUser(login string) (courses.User, error) {
	var user courses.User

	query := fmt.Sprintf("SELECT id, salt, password_hash FROM %s WHERE email=$1;", usersTable)

	if err := a.db.Get(&user, query, login); err == sql.ErrNoRows {
		return user, courses.ErrNoRows
	} else {
		return user, err
	}
}

func (a *AuthPostgres) CreateSession(userId int, SID string, expiredDate time.Time) (int, error) {
	query := fmt.Sprintf("INSERT INTO %s (user_id, session, expired_date) VALUES ($1, $2, $3) RETURNING id;", sessionTable)
	row := a.db.QueryRow(query, userId, SID, expiredDate)

	var id int
	if err := row.Scan(&id); err != nil {
		a.logger.Errorf("error while scanning id in auth_pstgrs/CreateSession: %s", err.Error())
		return 0, err
	}

	return id, nil
}

func (a *AuthPostgres) DeleteSession(SID string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE session = $1;", sessionTable)
	row := a.db.QueryRow(query, SID)
	return row.Err()
}

func (a *AuthPostgres) CheckSession(sessionId string) (int, error) {
	query := fmt.Sprintf("SELECT user_id FROM %s WHERE session=$1;", sessionTable)

	row := a.db.QueryRow(query, sessionId)

	var id int
	if err := row.Scan(&id); err != nil {
		a.logger.Error("error while scanning id in auth_pstgrs/CheckSession: %s", err.Error())
		return 0, err
	}

	return id, nil
}

func (a *AuthPostgres) CheckEmail(email string) (int, error) {
	query := fmt.Sprintf("SELECT id FROM %s WHERE email=$1;", usersTable)

	row := a.db.QueryRow(query, email)

	var id int
	err := row.Scan(&id)
	switch err {
	case sql.ErrNoRows:
		return 0, courses.ErrNoRows
	case nil:
		return id, nil
	default:
		a.logger.Errorf("error while scanning id in auth_pstgrs/CheckEmail: %s", err.Error())
		return 0, err
	}
}

func (a *AuthPostgres) UpdatePassword(email, password string, salt int) (int, error) {
	query := fmt.Sprintf("UPDATE %s SET password_hash = $1, salt = $2 WHERE email = $3 RETURNING id;", usersTable)

	row := a.db.QueryRow(query, password, salt, email)

	var id int
	err := row.Scan(&id)
	switch err {
	case sql.ErrNoRows:
		a.logger.Error("error in repository/auth_pstgrs.UpdatePassword: error = sql.ErrNoRows, but before was checked")
		return 0, courses.ErrNoRows
	case nil:
		return id, nil
	default:
		a.logger.Errorf("error while scanning id in auth_pstgrs/CheckEmail: %s", err.Error())
		return 0, err
	}
}

func (a *AuthPostgres) CheckVkId(VkId int) (int, error) {
	query := fmt.Sprintf("SELECT id FROM %s WHERE vk=TRUE AND vk_id=$1;", usersTable)

	row := a.db.QueryRow(query, VkId)

	var id int
	err := row.Scan(&id)
	switch err {
	case sql.ErrNoRows:
		return 0, courses.ErrNoRows
	case nil:
		return id, nil
	default:
		a.logger.Errorf("error while scanning id in auth_pstgrs/CheckEmail: %s", err.Error())
		return 0, err
	}
}

func (a *AuthPostgres) DeleteExpiredSessions() error {
	now := time.Now().Round(time.Hour)
	query := fmt.Sprintf("DELETE FROM %s WHERE expired_date < $1;", sessionTable)
	row := a.db.QueryRow(query, now)
	return row.Err()
}
