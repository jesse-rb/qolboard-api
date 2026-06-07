package email

import (
	"context"
	"fmt"
	service "qolboard-api/services"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

// SESClient implements the EmailClient interface, and sends emails through AWS SES
type SESClient struct {
	client    *sesv2.Client
	fromEmail string
	fromName  string
}

func NewSESClient(ctx context.Context, fromEmail string, fromName string) (*SESClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("ap-southeast-2"))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &SESClient{
		client:    sesv2.NewFromConfig(cfg),
		fromEmail: fromEmail,
		fromName:  fromName,
	}, nil
}

func (s *SESClient) sendEmail(ctx context.Context, to, subject, htmlBody, textBody string) error {
	input := &sesv2.SendEmailInput{
		FromEmailAddress: service.ToPointer(fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data:    service.ToPointer(subject),
					Charset: service.ToPointer("UTF-8"),
				},
				Body: &types.Body{
					Html: &types.Content{
						Data:    service.ToPointer(htmlBody),
						Charset: service.ToPointer("UTF-8"),
					},
					Text: &types.Content{
						Data:    service.ToPointer(textBody),
						Charset: service.ToPointer("UTF-8"),
					},
				},
			},
		},
	}

	_, err := s.client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
