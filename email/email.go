package email

import (
	"gopkg.in/gomail.v2"
	"v2ray-admin/backend/conf"
)

func Send(subject string, body string, emailType string, to string, cc string) error {
	c := conf.App.Smtp

	m := gomail.NewMessage()
	m.SetAddressHeader("From", c.Username, c.From)
	m.SetHeader("To", to)
	if cc != "" {
		m.SetHeader("Cc", cc)
	}
	m.SetHeader("Subject", subject)
	m.SetBody(emailType, body)

	d := gomail.NewDialer(c.Host, c.Port, c.Username, c.Password)

	return d.DialAndSend(m)
}
