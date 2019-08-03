package teak

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/smtp"
)

//EmailConfig - configuration for sending email
type EmailConfig struct {
	AppEMail         string `json:"appEMail"`
	AppEMailPassword string `json:"appEMailPassword"`
	SMTPHost         string `json:"smtpHost"`
	SMTPPort         int    `json:"smtpPort"`
}

//SendEmail - sends an email with given information. Uses the package level
//variable emainConfig for SMTP configuration - smtp.gmail.com:587
func SendEmail(to, subject, meesage string) (err error) {
	var emailConfig EmailConfig
	found := GetConfig("emailConfig", &emailConfig)
	if !found {
		err = errors.New("Could not find EMail config")
		return LogError("t.net.email", err)
	}
	// DumpJSON(emailConfig)
	msg := "From: " + emailConfig.AppEMail + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		meesage
	smtpURL := fmt.Sprintf("%s:%d", emailConfig.SMTPHost, emailConfig.SMTPPort)
	auth := smtp.PlainAuth("",
		emailConfig.AppEMail,
		emailConfig.AppEMailPassword,
		emailConfig.SMTPHost)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpURL,
	}
	var conn *tls.Conn
	conn, err = tls.Dial("tcp", smtpURL, tlsConfig)
	if err != nil {
		return LogError("t.net.email", err)
	}
	var client *smtp.Client
	client, err = smtp.NewClient(conn, emailConfig.SMTPHost)
	if err != nil {
		return LogError("t.net.email", err)
	}
	err = client.Auth(auth)
	if err != nil {
		return LogError("t.net.email", err)
	}

	err = client.Mail(emailConfig.AppEMail)
	if err != nil {
		return LogError("t.net.email", err)
	}

	client.Rcpt(to)
	var writer io.WriteCloser
	writer, err = client.Data()
	if err != nil {
		return LogError("t.net.email", err)
	}

	_, err = writer.Write([]byte(msg))
	if err != nil {
		return LogError("t.net.email", err)
	}

	err = writer.Close()
	if err != nil {
		return LogError("t.net.email", err)
	}

	client.Quit()
	return LogError("t.net.email", err)
	// err = smtp.SendMail(
	// 	smtpURL,
	// 	auth,
	// 	emailConfig.From,
	// 	[]string{to},
	// 	[]byte(msg))
}
