package p

import (
	"crypto/tls"
	"os"

	"gopkg.in/gomail.v2"
)

var (
	NotifEmail = os.Getenv("NOTIFICATION_EMAIL")
	SMTPServer = "smtp.gmail.com"
	SMTPPort   = 465
	SMTPUser   = os.Getenv("SMTP_USER")
	SMTPPass   = os.Getenv("SMTP_PASS")
)

func SendEmail(from string, to []string, subject, content string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)

	m.SetBody("text/html", content)

	d := gomail.NewDialer(SMTPServer, SMTPPort, SMTPUser, SMTPPass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d.DialAndSend(m)
}
