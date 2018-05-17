package mail

import (
	"net/smtp"
	"net"
	"fmt"

	"github.com/ljjjustin/themis/config"
	"github.com/ljjjustin/themis/database"
	"github.com/coreos/pkg/capnslog"
	"encoding/base64"
	"crypto/tls"
	"errors"
)

var plog = capnslog.NewPackageLogger("github.com/ljjjustin/themis", "mail")

func SendAlert(config *config.ThemisConfig, host *database.Host) {

	host.Notified = true
	host.FencedTimes += 1
	saveHost(host)
	plog.Debugf("SendAlert set host notified true.")

	subject := "[warning]节点" + host.Name + "down, 请及时处理"

	body := "host HA 检测到 节点 " + host.Name + " down,请及时处理。"

	to := config.Mail.SendTo

	plog.Info("begin send notification mail to ", to)

	err := sendMail(config.Mail.SmtpUser, config.Mail.SmtpPassword, config.Mail.SmtpHost, subject, body, to)
	if err != nil {
		plog.Warningf("send notification mail to %s failed : %s", to, err)
	}

	plog.Info("send notification mail to ", to, " success.")

	return
}

func saveHost(host *database.Host) {
	database.HostUpdateFields(host, "notified", "fenced_times")
}

func sendMail(user, password, host, subject, body string, to []string) error {

	auth := NewLoginAuth(user, password)

	msg := []byte(body)

	err := SendToMail(host, auth, user, to, msg, subject)

	return err
}

func SendToMail(addr string, a smtp.Auth, from string, to []string, msg []byte, subject string) error {
	c, err := smtp.Dial(addr)
	host, _, _ := net.SplitHostPort(addr)
	if err != nil {
		plog.Debug("call dial")
		return err
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: host, InsecureSkipVerify: true}
		if err = c.StartTLS(config); err != nil {
			plog.Debug("call start tls")
			return err
		}
	}

	if a != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(a); err != nil {
				plog.Debugf("check auth with err:", err)
				return err
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}

	header := make(map[string]string)
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString(msg)

	_, err = w.Write([]byte(message))

	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

type LoginAuth struct {
	username, password string
}

func NewLoginAuth(username, password string) smtp.Auth {
	return &LoginAuth{username, password}
}

func (a *LoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *LoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unknown fromServer")
		}
	}
	return nil, nil
}