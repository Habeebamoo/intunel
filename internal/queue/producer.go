package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Habeebamoo/intunel-backend/internal/configs"
	"github.com/Habeebamoo/intunel-backend/internal/models"
	"github.com/redis/go-redis/v9"
)

const NotificationStream = "notifications:stream"

type Producer struct {
	cfg    *configs.Config
	client *redis.Client
}

func NewProducer(cfg *configs.Config, client *redis.Client) *Producer {
	return &Producer{cfg, client}
}

func (p *Producer) Publish(ctx context.Context, n models.Notification) error {
	payload, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	return p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: NotificationStream,
		MaxLen: 10000,
		Approx: true,
		Values: map[string]interface{}{
			"data": payload,
		},
	}).Err()
}

func (p *Producer) RedriveDLQ(ctx context.Context) error {
	dlqMessages, err := p.client.XRangeN(ctx, p.cfg.DlqStream, "-", "+", 50).Result()
	if err != nil { 
		return err
	}

	replayed := 0

	for _, msg := range dlqMessages {
		data, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}

		err := p.client.XAdd(ctx, &redis.XAddArgs{
			Stream: NotificationStream,
			Values: map[string]interface{}{
				"data": data,
			},
		}).Err()

		if err == nil {
			p.client.XDel(ctx, p.cfg.DlqStream, msg.ID)
			replayed++
		}
	}

	return nil
}