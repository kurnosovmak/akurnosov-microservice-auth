package mailer

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendVerificationEmail(to, token string) error {
	from := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	addr := os.Getenv("SMTP_HOST") + ":" + os.Getenv("SMTP_PORT")

	subject := "Subject: Verify your email\n"
	body := fmt.Sprintf("Click here to verify your email: %s/verify?token=%s", os.Getenv("APP_URL"), token)
	msg := []byte(subject + "\n" + body)

	auth := smtp.PlainAuth("", from, pass, os.Getenv("SMTP_HOST"))
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}
