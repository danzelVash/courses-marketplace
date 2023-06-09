package courses

import (
	"database/sql"
	"github.com/pkg/errors"
)

var (
	BadPassword           = errors.New("incorrect password")
	BadAdminAuth          = errors.New("incorrect input params for admin")
	ErrorSendingMail      = errors.New("error while try send mail")
	NoActiveCodesForEmail = errors.New("code like your isn`t in server memory")
	IncorrectAuthCode     = errors.New("on mail was sent other code")
	ErrNoRows             = sql.ErrNoRows
	AlreadyRegistered     = errors.New("this email is already registered")
	BadCode               = errors.New("your code is incorrect or expired")
	BadEmail              = errors.New("email is invalid")
	BadVkInteraction      = errors.New("error of interaction with the VK service")
	CleanerAlreadyRunning = errors.New("cleaner is already running")
)
