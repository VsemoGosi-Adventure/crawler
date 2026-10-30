package producer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/VsemoGosi-Adventure/crawler/internal/config"
	"github.com/VsemoGosi-Adventure/crawler/internal/dto"
	"github.com/segmentio/kafka-go"
)

type EventProducer interface {
	PublishEvent(ctx context.Context, event dto.CrawledEventMessage) error
	Close() error
}

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(cfg config.ProducerConfig) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(cfg.Brokers...),
			Topic:    cfg.Topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) Close() error {
	if err := p.writer.Close(); err != nil {
		return err
	}
	return nil
}

func (p *Producer) PublishEvent(ctx context.Context, event dto.CrawledEventMessage) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal crawled event: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.ExternalSourceID),
		Value: payload,
	})
	if err != nil {
		return fmt.Errorf("failed to write message to redpanda: %w", err)
	}

	return nil
}
