package watchers

import (
	"bytes"
	"fmt"
	"github.com/streamrail/sendgrid-go"
	"github.com/streamrail/watchdog/models"
	"net"
	"net/http"
	"time"
)

const (
	sendgridUsername = "sr_ops"
	sendgridKey      = "FoUx^l1c"
	mailFrom         = "status@streamrail.com"
	subjectPrefix    = "[SR Status]"
	debugMode        = false
)

var (
	sendgridClient = sendgrid.NewSendGridClient(sendgridUsername, sendgridKey)
)

func SendNotification(c *models.Check, err error) {
	subject := ""
	text := ""
	alert := true
	if err == nil {
		alert = false
		subject = fmt.Sprintf("%s Resolved: Back to normal for %s (%s)", subjectPrefix, c.Name, c.Type)
		text = fmt.Sprintf("Yo Dawg,\n\nThe test %s (type %s) now suffices it's specified condition again after a previous failure.\n\n Full test spec: \n%s\n\nBark bark, \n\nWatchdog",
			c.Name, c.Type, c.ToJsonString())
	} else {
		subject = fmt.Sprintf("%s Alert: Issue on %s (%s): %s", subjectPrefix, c.Name, c.Type, err.Error())
		text = fmt.Sprintf("Yo Dawg,\n\nThe test %s (type %s) failed to suffice it's specified condition for at least %d times. \n\nError: %s.\n\n Full test spec: \n%s\n\nBark bark, \n\nWatchdog :O)",
			c.Name, c.Type, c.AlertAfter+1, err.Error(), c.ToJsonString())
	}
	if len(c.SlackWebHookUrl) > 0 {
		SendSlackMessage(c.SlackWebHookUrl, subject, text, alert)
	}
	if len(c.Mailto) > 0 {
		SendEmail(subject, text, mailFrom, c.Mailto)
	}
}

func SendSlackMessage(webHookUrl string, subject string, text string, alert bool) {
	color := "danger"
	if !alert {
		color = "good"
	}

	s := []byte(fmt.Sprintf(`{"attachments": [{ "fallback": "%s", "color": "%s", "author_name": "Status Bot", "author_link": "http://status.streamrail.com", "title": "%s", "title_link": "http://status.streamrail.com/", "fields": [ { "title": "Priority", "value": "High", "short": false } ] } ] }`, subject, color, subject))
	if req, err := http.NewRequest("POST", webHookUrl, bytes.NewBuffer(s)); err != nil {
	} else {
		tr := http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, 30*time.Second)
			},
		}
		req.Header.Set("Content-Type", "application/json")
		if debugMode {
			fmt.Println(subject)
			fmt.Println(text)
			return
		}
		if res, err := tr.RoundTrip(req); err != nil {
			return
		} else {
			defer res.Body.Close()
		}
	}
}

func SendEmail(subject string, text string, mailFrom string, mailTo string) {
	message := sendgrid.NewMail()
	message.AddTo(mailTo)
	message.SetSubject(subject)
	message.SetText(text)
	message.SetFrom(mailFrom)
	if debugMode {
		fmt.Println(subject)
		fmt.Println(text)
		return
	}
	sendgridClient.Send(message)
}
