package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
)

type EmailData struct {
	To       string
	Token    string
	AppURL   string
	Username string
}

const emailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify your email</title>
</head>
<body>
    <h2>Welcome to our service!</h2>
    <p>Please click the link below to verify your email address:</p>
    <p><a href="{{.AppURL}}/verify?token={{.Token}}">Verify Email</a></p>
    <p>If you didn't request this, please ignore this email.</p>
</body>
</html>
`

func SendVerificationEmail(to, token string) error {
	from := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	appURL := os.Getenv("APP_URL")

	// Подробное логирование отсутствующих переменных
	missing := make([]string, 0)
	if from == "" {
		missing = append(missing, "SMTP_USER")
	}
	if pass == "" {
		missing = append(missing, "SMTP_PASS")
	}
	if host == "" {
		missing = append(missing, "SMTP_HOST")
	}
	if port == "" {
		missing = append(missing, "SMTP_PORT")
	}
	if appURL == "" {
		missing = append(missing, "APP_URL")
	}

	if len(missing) > 0 {
		log.Panicf("missing required email configuration: %v", missing)
	}

	data := EmailData{
		To:     to,
		Token:  token,
		AppURL: appURL,
	}

	// Создание HTML письма
	tmpl, err := template.New("email").Parse(emailTemplate)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}

	// Формирование заголовков письма
	headers := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: Verify your email\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n",
		from, to)

	// Отправка письма
	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", from, pass, host)
	msg := append([]byte(headers), body.Bytes()...)

	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}
