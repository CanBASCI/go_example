package kafka

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"go_example/internal/events"
	"go_example/cmd/user-service/service"
)

// Consumer runs Kafka consumers for user-service (order.created, order.canceled).
type Consumer struct {
	userSvc *service.UserService
	brokers []string
	writer  *kafka.Writer
}

// NewConsumer creates a new Consumer.
func NewConsumer(userSvc *service.UserService, brokers []string) *Consumer {
	return &Consumer{
		userSvc: userSvc,
		brokers: brokers,
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers[0]),
			Topic:    "", // set per message
			Balancer: &kafka.LeastBytes{},
		},
	}
}

// Close closes the Kafka writer.
func (c *Consumer) Close() error {
	return c.writer.Close()
}

// Run starts consuming order.created and order.canceled.
func (c *Consumer) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		c.consumeOrderCreated(ctx)
	}()
	go func() {
		defer wg.Done()
		c.consumeOrderCanceled(ctx)
	}()
	wg.Wait()
}

func (c *Consumer) consumeOrderCreated(ctx context.Context) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		Topic:    "order.created",
		GroupID:  "user-service-group",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer r.Close()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[user-service] order.created read error: %v", err)
				continue
			}
			var evt events.OrderCreatedEvent
			if err := json.Unmarshal(msg.Value, &evt); err != nil {
				log.Printf("[user-service] order.created unmarshal error: %v", err)
				continue
			}
			log.Printf("[user-service] Received OrderCreatedEvent: orderId=%s userId=%s amount=%d", evt.OrderID, evt.UserID, evt.Amount)
			reserved, err := c.userSvc.ReserveCredit(ctx, evt.UserID, evt.Amount)
			if err != nil {
				log.Printf("[user-service] reserve credit error: %v", err)
				c.publishCreditReservationFailed(ctx, evt.OrderID, evt.UserID, evt.Amount, err.Error())
				continue
			}
			if reserved {
				log.Printf("[user-service] Credit reserved for orderId=%s", evt.OrderID)
				c.publishCreditReserved(ctx, evt.OrderID, evt.UserID, evt.Amount)
			} else {
				log.Printf("[user-service] Insufficient balance for orderId=%s", evt.OrderID)
				c.publishCreditReservationFailed(ctx, evt.OrderID, evt.UserID, evt.Amount, "Insufficient balance")
			}
		}
	}
}

func (c *Consumer) consumeOrderCanceled(ctx context.Context) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		Topic:    "order.canceled",
		GroupID:  "user-service-group",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer r.Close()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[user-service] order.canceled read error: %v", err)
				continue
			}
			var evt events.OrderCanceledEvent
			if err := json.Unmarshal(msg.Value, &evt); err != nil {
				log.Printf("[user-service] order.canceled unmarshal error: %v", err)
				continue
			}
			log.Printf("[user-service] Received OrderCanceledEvent: orderId=%s userId=%s amount=%d", evt.OrderID, evt.UserID, evt.Amount)
			if err := c.userSvc.ReleaseCredit(ctx, evt.UserID, evt.Amount); err != nil {
				log.Printf("[user-service] release credit error: %v", err)
				continue
			}
			log.Printf("[user-service] Credit released for orderId=%s", evt.OrderID)
		}
	}
}

func (c *Consumer) publishCreditReserved(ctx context.Context, orderID, userID uuid.UUID, amount int64) {
	evt := events.UserCreditReservedEvent{OrderID: orderID, UserID: userID, Amount: amount}
	body, _ := json.Marshal(evt)
	_ = c.writeMessage(ctx, "user.credit-reserved", body)
}

func (c *Consumer) publishCreditReservationFailed(ctx context.Context, orderID, userID uuid.UUID, amount int64, reason string) {
	evt := events.UserCreditReservationFailedEvent{OrderID: orderID, UserID: userID, Amount: amount, Reason: reason}
	body, _ := json.Marshal(evt)
	_ = c.writeMessage(ctx, "user.credit-reservation-failed", body)
}

func (c *Consumer) writeMessage(ctx context.Context, topic string, value []byte) error {
	return c.writer.WriteMessages(ctx, kafka.Message{Topic: topic, Value: value})
}