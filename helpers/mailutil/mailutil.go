// Package mailutil provides mail sending utilities
package mailutil

import (
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
)

type Config struct {
	SmtpHost     string
	SmtpPort     int
	SmtpUsername string
	SmtpPassword string
	From         string
}

func Send(config Config, to []string, subject, message string) (err error) {
	var headers []string
	headers = append(headers, fmt.Sprintf("From: %s", config.From))
	headers = append(headers, fmt.Sprintf("To: %s", strings.Join(to, ", ")))
	headers = append(headers, fmt.Sprintf("Subject: %s", subject))
	headers = append(headers, "MIME-version: 1.0", "Content-Type: text/html; charset=\"UTF-8\"")
	allheaders := strings.Join(headers, "\r\n") + "\r\n\r\n"
	auth := smtp.PlainAuth("", config.SmtpUsername, config.SmtpPassword, config.SmtpHost)
	err = smtp.SendMail(config.SmtpHost+":"+strconv.Itoa(config.SmtpPort), auth, config.From, to, []byte(allheaders+message))
	if err != nil {
		return fmt.Errorf("Error when sending mail: %w", err)
	}
	return nil
}
