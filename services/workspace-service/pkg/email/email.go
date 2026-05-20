package email

import (
	"context"
	"fmt"
	"net/smtp"
)

type EmailClient interface {
	SendInvite(ctx context.Context, to, token string) error
	SendReminder(ctx context.Context, to string) error
}

type gmailClient struct {
	smtpHost string
	smtpPort string
	from     string
	password string
}

func NewGmailClient(host, port, from, password string) EmailClient {
	return &gmailClient{
		smtpHost: host,
		smtpPort: port,
		from:     from,
		password: password,
	}
}

func (g *gmailClient) SendInvite(ctx context.Context, to, token string) error {
	auth := smtp.PlainAuth("", g.from, g.password, g.smtpHost)

	inviteLink := fmt.Sprintf("https://your-app.com/invite/accept?token=%s", token)
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: Workspace Invitation\r\n"+
		"\r\n"+
		"You have been invited to join the workspace. Click here to accept: %s\r\n"+
		"This link will expire in 14 days.\r\n", to, inviteLink))

	err := smtp.SendMail(g.smtpHost+":"+g.smtpPort, auth, g.from, []string{to}, msg)
	return err
}

func (g *gmailClient) SendReminder(ctx context.Context, to string) error {
	auth := smtp.PlainAuth("", g.from, g.password, g.smtpHost)

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: Reminder: Workspace Invitation\r\n"+
		"\r\n"+
		"Friendly reminder! Your workspace invitation is pending. It will expire in 4 days.\r\n", to))

	err := smtp.SendMail(g.smtpHost+":"+g.smtpPort, auth, g.from, []string{to}, msg)
	return err
}
