package providers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Habeebamoo/intunel-backend/internal/configs"
	"github.com/Habeebamoo/intunel-backend/internal/models"
	//"github.com/resend/resend-go/v3"
)

type EmailProvider struct {
	ResendApiKey string
}

func NewEmailProvider(cfg *configs.Config) *EmailProvider {
	return &EmailProvider{ResendApiKey: cfg.ResendApiKey}
}

func (e *EmailProvider) Send(ctx context.Context, n models.Notification) error {
	senderDomain := os.Getenv("SENDER_DOMAIN")

	if senderDomain == "" {
		return fmt.Errorf("Invalid Sender Domain")
	}

	// client := resend.NewClient(e.ResendApiKey)
	// title := fmt.Sprintf("%s <%s>", n.Title, senderDomain)

	// params := &resend.SendEmailRequest{
	// 	From: title,
	// 	To: []string{n.To},
	// 	Subject: n.Title,
	// 	Html: n.Body,
	// }

	// _, err := client.Emails.Send(params)
	// if err != nil {
	// 	return err
	// }

	time.Sleep(5*time.Second)

	log.Printf("Email sent: %v", n)
	return nil
}