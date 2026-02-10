package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"

	"go_example/internal/events"
)

// Producer publishes order events to Kafka.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Producer.
func NewProducer(brokers []string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers[0]),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

// Close closes the producer.
func (p *Producer) Close() error {
	return p.writer.Close()
}

// PublishOrderCreated publishes OrderCreatedEvent to order.created.
func (p *Producer) PublishOrderCreated(ctx context.Context, evt events.OrderCreatedEvent) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{Topic: "order.created", Value: body})
}

// PublishOrderCanceled publishes OrderCanceledEvent to order.canceled.
func (p *Producer) PublishOrderCanceled(ctx context.Context, evt events.OrderCanceledEvent) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{Topic: "order.canceled", Value: body})
}
