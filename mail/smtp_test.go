package mail

import "testing"

func TestSendMailList(t *testing.T) {

	user := ""

	password := ""

	host := "smtp.hnair.net:587"

	var to []string

	to1 := ""
	to2 := ""

	to = append(to, to2)
	to = append(to, to1)

	subject := "[test]ha发送多人"

	str := "test-baicheng"

	body := "host HA 检测到 节点" + str + "down,请及时处理。"

	t.Log("send email to ", to)

	err := sendMail(user, password, host, subject, body, to)
	if err != nil {
		t.Log("发送邮件失败")
		t.Log(err)
	} else {
		t.Log("send notification mail to ", to, " success.")
	}
}