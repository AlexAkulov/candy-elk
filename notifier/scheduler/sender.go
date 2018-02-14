package scheduler

import (
	"bytes"
	"crypto/tls"
	"html/template"
	"io"

	gomail "gopkg.in/gomail.v2"

	"github.com/AlexAkulov/candy-elk"
)

var tpl = template.Must(template.New("mail").Parse(`
<html>
	<head>
		<style type="text/css">
			th, td { border: 1px solid black; padding: 3px}
		</style>
	</head>
	<body>
		<table style="border-collapse: collapse">
			<thead>
				<tr>
					<th>Alert</th>
					<th>Message</th>
					<th>Count</th>
				</tr>
			</thead>
			<tbody>
				{{range .Items}}
				<tr>
					<td>{{ .Alert }}</td>
					<td>
						<table>
						{{ range $field, $value := .Message.Fields }}
							<tr><td style="border:none">{{$field}}</td><td style="border:none"><pre>{{$value}}</pre></td></tr>
						{{end}}
						</table>
					</td>
					<td>{{ .Count }}</td>
				</tr>
				{{end}}
			</tbody>
		</table>
		<p>Please, do something!</p>
	</body>
</html>
`))

type templateRow struct {
	Alert   string
	Message *elkstreams.DecodedLogMessage
	Count   int
}

// MakeMessage is making smtp message from template
func makeMessage(config *MailConfig, notification Notification) *gomail.Message {
	var subjectBuffer bytes.Buffer
	var subject string
	for _, data := range notification.AlertsData {
		subjectBuffer.WriteString(data.Name)
		subjectBuffer.WriteRune(',')
	}
	if subjectBuffer.Len() > 0 {
		subjectBuffer.Truncate(subjectBuffer.Len() - 1)
		subject = subjectBuffer.String()
	} else {
		subject = "Error alert"
	}

	templateData := struct {
		Items []*templateRow
	}{
		Items: make([]*templateRow, 0, len(notification.AlertsData)),
	}

	for _, data := range notification.AlertsData {
		templateData.Items = append(templateData.Items, &templateRow{
			Alert:   data.Name,
			Message: data.Message,
			Count:   data.Count,
		})
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.From)
	m.SetHeader("To", notification.Recipient)
	m.SetHeader("Subject", subject)
	m.AddAlternativeWriter("text/html", func(w io.Writer) error {
		return tpl.Execute(w, templateData)
	})

	return m
}

// SendNotification is making mail message and send it via smtp server
func sendNotification(config *MailConfig, notification Notification) error {
	m := MakeMessage(config, notification)
	if len(notification.Recipient) == 0 {
		return nil
	}
	d := gomail.Dialer{
		Host: config.SMTPHost,
		Port: config.SMTPPort,
		TLSConfig:
		&tls.Config{
			InsecureSkipVerify:
			config.InsecureTLS,
		},
	}
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	log.Debugf("Successfully send notification for [%s] subject [%s]", m.GetHeader("To"), m.GetHeader("Subject"))
	return nil
}

// StartNotifier is receiving notification from channel and send it
func StartNotifier(config *MailConfig, ch <-chan Notification) {
	for notification := range ch {
		go SendNotification(config, notification)
	}
}
