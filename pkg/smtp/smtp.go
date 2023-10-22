package smtp

import (
	"bytes"
	"html/template"
	"net"
	"net/mail"
	"os"
	"strings"

	"github.com/danzelVash/courses-marketplace"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

type EmailParams struct {
	TemplateName string
	TemplateVars interface{}
	Destination  string
	Subject      string
}

func SendEmail(params EmailParams) error {
	address, err := mail.ParseAddress(params.Destination)
	if err != nil {
		return courses.BadEmail
	}

	domain := strings.Split(address.Address, "@")[1]

	mx, err := net.LookupMX(domain)
	if err != nil || len(mx) == 0 {
		return courses.BadEmail
	}

	var body bytes.Buffer
	t, err := template.ParseFiles(params.TemplateName)
	if err != nil {
		return errors.Errorf("error while parsing email template %s: %s", params.TemplateName, err.Error())
	}

	err = t.Execute(&body, params.TemplateVars)
	if err != nil {
		return errors.Errorf("error while executing email template %s: %s", params.TemplateName, err.Error())
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", viper.GetString("gmail.supportAddress"))
	msg.SetHeader("To", params.Destination)
	msg.SetHeader("Subject", params.Subject)
	msg.SetBody("text/html", body.String())

	d := gomail.NewDialer("smtp.gmail.com", 587, viper.GetString("gmail.supportAddress"), os.Getenv("GMAIL_SUPPORT_SECRET"))

	if err = d.DialAndSend(msg); err != nil {
		return courses.ErrorSendingMail
	}

	return nil
}
