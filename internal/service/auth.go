package service

import (
	"math/rand"

	"github.com/danzelVash/courses-marketplace"
	"github.com/danzelVash/courses-marketplace/internal/repository"
	"github.com/danzelVash/courses-marketplace/pkg/logging"
	"github.com/danzelVash/courses-marketplace/pkg/smtp"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"

	"os"
	"strconv"
	"time"
)

const (
	weightOfHashing = 10
	passwordLen     = 8

	sessLen  = 40
	alphabet = "QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzxcvbnm1234567890!#*()_+=-<>/?"
)

type AuthService struct {
	repos   repository.Authorization
	cleaner Cleaner
	logger  *logging.Logger
}

func NewAuthService(repos repository.Authorization, logger *logging.Logger) *AuthService {
	return &AuthService{
		repos:   repos,
		cleaner: NewSessionCleaner(repos, logger),
		logger:  logger,
	}
}

func (a *AuthService) CreateUser(user courses.User) (int, error) {
	_, err := a.repos.GetUser(user.Email)
	if err != courses.ErrNoRows && err != nil {
		return 0, err
	} else if err == nil {
		return 0, courses.AlreadyRegistered
	}
	password := generateRandomString(passwordLen)

	passHash, salt, err := generatePasswordHash(password)
	if err != nil {
		a.logger.Errorf("error while generating hashed password: %s", err.Error())
		return 0, err
	}

	user.Password = passHash
	user.Salt = salt

	id, err := a.repos.CreateUser(user)
	if err != nil {
		a.logger.Errorf("error while creating user with email = %s", user.Email)
		return 0, err
	}

	emailParams := smtp.EmailParams{
		TemplateName: "templates/email_templates/sending_password.html",
		TemplateVars: struct {
			Password string
			Url      string
		}{
			Password: password,
			Url:      viper.GetString("domain"),
		},
		Destination: user.Email,
		Subject:     "Войдите в личный кабинет",
	}

	if err := smtp.SendEmail(emailParams); err != nil {
		a.logger.Errorf("error while sending email with password: %s", err.Error())
		return 0, err
	}

	return id, nil
}

func (a *AuthService) CreateSession(login, password string) (string, error) {
	user, err := a.repos.GetUser(login)

	if err != nil {
		return "", err
	}

	if err := checkPassEqualPassHash(password, user.Salt, user.Password); err != nil {
		return "", courses.BadPassword
	}

	SID := generateRandomString(sessLen)
	expiredDate := time.Now().AddDate(0, 0, 15).Round(time.Hour)
	if _, err := a.repos.CreateSession(user.Id, SID, expiredDate); err != nil {
		a.logger.Errorf("error while creating session in auth_pstgrs %s", err.Error())
		return "", err
	}

	return SID, nil
}

func (a *AuthService) DeleteSession(SID string) error {
	return a.repos.DeleteSession(SID)
}

func (a *AuthService) CheckSession(sessionId string) (int, error) {
	userId, err := a.repos.CheckSession(sessionId)

	if err != nil {
		return 0, err
	}

	return userId, err
}

func (a *AuthService) CheckEmailAndSendCode(email string) error {
	if _, err := a.repos.CheckEmail(email); err != nil {
		return err
	}

	code := generateAuthCode()
	emailParams := smtp.EmailParams{
		TemplateName: "templates/email_templates/password_update_code.html",
		TemplateVars: struct {
			Url  string
			Code int
		}{
			Url:  viper.GetString("domain"),
			Code: code,
		},
		Destination: email,
		Subject:     "Обновление пароля",
	}

	if err := smtp.SendEmail(emailParams); err != nil {
		a.logger.Errorf("error while sending email with auth code %s", err.Error())
		return err
	}

	a.repos.AddCode(email, code)

	return nil
}

func (a *AuthService) CheckCode(email string, code int) error {
	correctCode, exist := a.repos.GetCode(email)

	if !exist || correctCode != code {
		return courses.BadCode
	}

	return nil
}

func (a *AuthService) UpdatePassword(email string) error {
	password := generateRandomString(passwordLen)
	passHash, salt, err := generatePasswordHash(password)
	if err != nil {
		a.logger.Errorf("error while generating hashed password: %s", err.Error())
		return err
	}

	if _, err = a.repos.UpdatePassword(email, passHash, salt); err != nil {
		a.logger.Errorf("error while updating password: %s", err.Error())
		return err
	}

	emailParams := smtp.EmailParams{
		TemplateName: "templates/email_templates/updated_password.html",
		TemplateVars: struct {
			Password string
			Url      string
		}{
			Password: password,
			Url:      viper.GetString("domain"),
		},
		Destination: email,
		Subject:     "Пароль обновлен",
	}

	if err := smtp.SendEmail(emailParams); err != nil {
		a.logger.Errorf("error while sending mail with password: %s", err.Error())
		return err
	}
	return nil
}

func (a *AuthService) CleanExpiredSessions() {
	go a.cleaner.CleanSessionStorage()
}

func generatePasswordHash(pass string) (string, int, error) {
	rand.Seed(time.Now().UnixNano())
	salt := rand.Intn(10000)
	hash, err := bcrypt.GenerateFromPassword([]byte(os.Getenv("PASS_SALT")+pass+strconv.Itoa(salt)), weightOfHashing)

	return string(hash), salt, err
}

func checkPassEqualPassHash(password string, salt int, passwordHash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(os.Getenv("PASS_SALT")+password+strconv.Itoa(salt)))
	return err
}

func generateRandomString(strLen int) string {
	rand.Seed(time.Now().UnixNano())
	sid := ""
	for i := 0; i < strLen; i++ {
		sid += string(alphabet[rand.Intn(len(alphabet))])
	}
	return sid
}
