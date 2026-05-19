package email

import (
	"context"
	"fmt"
)

type EmailClient interface {
	sendEmail(ctx context.Context, to string, subject string, htmlBody string, textBody string) error
}

// LogClient implements the EmailClient interface, but only logs emails to stdout, and never returns an error,
// useful for local development
type LogClient struct {
	fromEmail string
	fromName  string
}

func NewLogClient(fromEmail string, fromName string) *LogClient {
	return &LogClient{
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

func (c *LogClient) sendEmail(ctx context.Context, to string, subject string, htmlBody string, textBody string) error {
	fmt.Printf("from email: %s\nfrom name:%s\nto:%s\nsubject:%s\n\nhtml body:\n%s\n\ntext body:\n%s\n\n", c.fromEmail, c.fromName, to, subject, htmlBody, textBody)
	return nil
}
