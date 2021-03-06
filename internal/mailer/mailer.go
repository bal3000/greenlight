package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	// Initialize a new mail.Dialer instance with the given SMTP server settings. We
	// also configure this to use a 5-second timeout whenever we send an email.
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Send takes the recipient email address
// as the first parameter, the name of the file containing the templates, and any
// dynamic data for the templates as an interface{} parameter. TODO: Generics
func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	tmpl, err := template.New("email").ParseFS(templateFS, fmt.Sprintf("templates/%s", templateFile))
	if err != nil {
		return err
	}

	// Execute the named template "subject", passing in the dynamic data and storing the
	// result in a bytes.Buffer variable.
	var subject bytes.Buffer
	err = tmpl.ExecuteTemplate(&subject, "subject", data)
	if err != nil {
		return err
	}

	var plainBody bytes.Buffer
	err = tmpl.ExecuteTemplate(&plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	var htmlBody bytes.Buffer
	err = tmpl.ExecuteTemplate(&htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())

	// It's important to note that AddAlternative() should
	// always be called *after* SetBody().
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// Try sending the email up to three times before aborting and returning the final
	// error. We sleep for 500 milliseconds between each attempt.
	for i := 1; i <= 3; i++ {
		// Call the DialAndSend() method on the dialer, passing in the message to send. This
		// opens a connection to the SMTP server, sends the message, then closes the
		// connection. If there is a timeout, it will return a "dial tcp: i/o timeout"
		// error.
		err = m.dialer.DialAndSend(msg)
		if err == nil {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}
