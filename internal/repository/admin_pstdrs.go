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

type AdminPostgres struct {
	db      *sqlx.DB
	storage *storage.DataStorage
	logger  *logging.Logger
}

func NewAdminPostgres(db *sqlx.DB, storage *storage.DataStorage, logger *logging.Logger) *AdminPostgres {
	return &AdminPostgres{
		db:      db,
		storage: storage,
		logger:  logger,
	}
}

func (adm *AdminPostgres) AddCode(key string, val int) {
	adm.storage.Add(key, val)
}

func (adm *AdminPostgres) GetCode(key string) (int, bool) {
	return adm.storage.Get(key)
}

func (adm *AdminPostgres) DeleteCode(key string) error {
	if err := adm.storage.Delete(key); err != nil {
		adm.logger.Errorf("error while deleting admin auth code from storage: %s", err.Error())
		return err
	}
	return nil
}

func (adm *AdminPostgres) CreateSession(userId int, SID string, expiredDate time.Time) (int, error) {
	query := fmt.Sprintf("INSERT INTO %s (user_id, session, expired_date) VALUES ($1, $2, $3) RETURNING id;", sessionTable)
	row := adm.db.QueryRow(query, userId, SID, expiredDate)

	var id int
	if err := row.Scan(&id); err != nil {
		adm.logger.Errorf("error while scanning &id in auth_pstgrs/CreateSession: %s", err.Error())
		return 0, err
	}

	return id, nil

}

func (adm *AdminPostgres) CheckSession(sessionId string) (int, error) {
	query := fmt.Sprintf("SELECT user_id FROM %s WHERE ((user_id IN (SELECT user_id FROM %s WHERE is_admin=TRUE)) AND session.session=$1);", sessionTable, usersTable)
	row := adm.db.QueryRow(query, sessionId)

	var id int
	if err := row.Scan(&id); err != nil {
		adm.logger.Errorf("error sending request to db while checking admin session")
		return 0, err
	}

	return id, nil
}

func (adm *AdminPostgres) SelectAllUsers() ([]courses.User, error) {
	query := fmt.Sprintf("SELECT id, last_name, first_name, email, phone_number, vk, vk_id FROM %s WHERE is_admin=false;", usersTable)

	rows, err := adm.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			adm.logger.Errorf("memory leak in admin_pstgrs.SelectAllUsers: %s", err.Error())
		}
	}(rows)

	var users []courses.User
	for rows.Next() {
		var user courses.User
		if err := rows.Scan(&user.Id, &user.LastName, &user.FirstName, &user.Email, &user.PhoneNumber, &user.Vk, &user.VkId); err != nil {
			adm.logger.Errorf("error while scanning user in admin_pstgrs.SelectAllUsers: %s", err.Error())
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		adm.logger.Errorf("error while rows.Err(): %s", err.Error())
	}
	return users, err
}
