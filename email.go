package teak

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"

	"github.com/varunamachi/vaali/vcmn"

	"github.com/varunamachi/vaali/vlog"
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
	err = vcmn.GetConfig("emailConfig", &emailConfig)
	if err != nil {
		return vlog.LogError("Net:EMail", err)
	}
	// vcmn.DumpJSON(emailConfig)
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
		return vlog.LogError("Net:EMail", err)
	}
	var client *smtp.Client
	client, err = smtp.NewClient(conn, emailConfig.SMTPHost)
	if err != nil {
		return vlog.LogError("Net:EMail", err)
	}
	err = client.Auth(auth)
	if err != nil {
		return vlog.LogError("Net:EMail", err)
	}

	err = client.Mail(emailConfig.AppEMail)
	if err != nil {
		return vlog.LogError("Net:EMail", err)
	}

	client.Rcpt(to)
	var writer io.WriteCloser
	writer, err = client.Data()
	if err != nil {
		return vlog.LogError("Net:EMail", err)
	}

	_, err = writer.Write([]byte(msg))
	if err != nil {
		return vlog.LogError("Net:EMail", err)
	}

	err = writer.Close()
	if err != nil {
		return vlog.LogError("Net:EMail", err)
	}

	client.Quit()
	return vlog.LogError("Net:EMail", err)
	// err = smtp.SendMail(
	// 	smtpURL,
	// 	auth,
	// 	emailConfig.From,
	// 	[]string{to},
	// 	[]byte(msg))
}
