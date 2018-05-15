package mail

import (
	"net/smtp"
	"strings"
	"time"

	"github.com/ljjjustin/themis/config"
	"github.com/ljjjustin/themis/database"
	"github.com/coreos/pkg/capnslog"
)

var plog = capnslog.NewPackageLogger("github.com/ljjjustin/themis", "mail")

func SendAlert(config *config.ThemisConfig, host *database.Host) {

	host.Notified = true
	host.FencedTimes += 1
	saveHost(host)
	plog.Debugf("SendAlert set host notified true.")

	subject := "[warning]节点" + host.Name + "down, 请及时处理"

	mailtype := "html"

	body := `
		<html>
		<body>
		<h3>
			host HA 检测到 节点 ` + host.Name + ` down,请及时处理。
		</h3>
		</body>
		</html>
		`
	to := config.Mail.SendTo[0]
	for i := 1 ; i < len(config.Mail.SendTo) ; i ++ {

		to = to + ";" + config.Mail.SendTo[i]
	}

	for {
		plog.Info("send notification mail to ", to)

		err := sendMail(config.Mail.SmtpUser, config.Mail.SmtpPassword, config.Mail.SmtpHost, to, subject, body, mailtype)
		if err != nil {
			plog.Warningf("send notification mail to %s failed : %s", to, err)
		}

		//check db if send false -> break
		hostFromDB, err := database.HostGetById(host.Id)
		if err != nil {
			plog.Warningf("Can't find Host %s.", host.Id)
		}

		if !hostFromDB.Notified {
			plog.Debug("bd notified from true to false....stop notify")
			break
		}

		time.Sleep(60 * time.Second)
	}

	return
}

func saveHost(host *database.Host) {
	host.UpdatedAt = time.Now()
	database.HostUpdateFields(host, "notified", "updated_at", "fenced_times")
}

func sendMail(user, passwword, host, to, subject, body, mailtype string) error {

	hp := strings.Split(host, ":")

	auth := smtp.PlainAuth("", user, passwword, hp[0])

	var content_type string

	if mailtype == "html" {
		content_type = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain; charset=UTF-8"
	}

	msg := []byte("To: " + to + "\r\nFrom: " + user + "\r\nSubject:" + subject + "\r\n" + content_type + "\r\n\r\n" + body)

	send_to := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, user, send_to, msg)

	return err
}