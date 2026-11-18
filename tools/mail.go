package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
)

type MailTool struct {
	Addr     string
	MailFrom string
	MailTo   string
}

type MailContent struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func (c *MailTool) Name() string {
	return "mail"
}

func (c *MailTool) Description() string {
	return `Mail tool to send emails, to use just provide the email content you want to send 
	the body of the email can be in HTML format and should be easily readable even in bed half asleep, 
	so green and red and important stuff big.
	which should include the subject and body in the following format:
	"{
		"subject": "Your subject here", "body": "Your email body here"
	}"`
}

func (c *MailTool) Execute(ctx context.Context, input string) (string, error) {
	var mailContent MailContent
	if err := json.Unmarshal([]byte(input), &mailContent); err != nil {
		return "", fmt.Errorf("failed to unmarshal mail content: %w", err)
	}

	// use postfix to send the email
	// Build RFC-2822 email
	toHeader := c.MailTo
	from := mail.Address{Name: "", Address: c.MailFrom}
	fromHeader := from.String()
	subjectHeader := mailContent.Subject
	header := make(map[string]string)
	header["To"] = toHeader
	header["From"] = fromHeader
	header["Subject"] = subjectHeader
	header["Content-Type"] = `text/html; charset="UTF-8"`
	msg := ""

	for k, v := range header {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	msg += "\r\n" + mailContent.Body
	bMsg := []byte(msg)

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", c.Addr)
	if err != nil {
		return "", fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	host, _, _ := net.SplitHostPort(c.Addr)

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return "", fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	if err = client.Mail(c.MailFrom); err != nil {
		return "", fmt.Errorf("failed to set sender: %w", err)
	}

	if err = client.Rcpt(c.MailTo); err != nil {
		return "", fmt.Errorf("failed to set recipient: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return "", fmt.Errorf("failed to create data writer: %w", err)
	}
	_, err = w.Write(bMsg)
	if err != nil {
		return "", fmt.Errorf("failed to write email data: %w", err)
	}

	err = w.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close data writer: %w", err)
	}

	err = client.Quit()
	// Or alternatively, send with remote service like Amazon SES
	// err = smtp.SendMail(addr, auth, fromEmail, toEmails, bMsg)
	// Handle response from local postfix or remote service
	if err != nil {
		return "", fmt.Errorf("failed to quit SMTP connection: %w", err)
	}

	msg = fmt.Sprintf("Email sent successfully to %s with subject: %s", c.MailTo, mailContent.Subject)
	log.Printf("[INFO] [tool: mail] %s", msg)
	return msg, err
}
