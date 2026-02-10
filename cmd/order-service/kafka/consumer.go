package kafka

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/segmentio/kafka-go"

	"go_example/internal/events"
	"go_example/cmd/order-service/service"
)

// Consumer runs Kafka consumers for order-service (user.credit-reserved, user.credit-reservation-failed).
type Consumer struct {
	orderSvc *service.OrderService
	brokers  []string
}

// NewConsumer creates a new Consumer.
func NewConsumer(orderSvc *service.OrderService, brokers []string) *Consumer {
	return &Consumer{orderSvc: orderSvc, brokers: brokers}
}

// Run starts consuming credit events.
func (c *Consumer) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		c.consumeCreditReserved(ctx)
	}()
	go func() {
		defer wg.Done()
		c.consumeCreditReservationFailed(ctx)
	}()
	wg.Wait()
}

func (c *Consumer) consumeCreditReserved(ctx context.Context) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		Topic:    "user.credit-reserved",
		GroupID:  "order-service-group",
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
				log.Printf("[order-service] user.credit-reserved read error: %v", err)
				continue
			}
			var evt events.UserCreditReservedEvent
			if err := json.Unmarshal(msg.Value, &evt); err != nil {
				log.Printf("[order-service] user.credit-reserved unmarshal error: %v", err)
				continue
			}
			log.Printf("[order-service] Received UserCreditReservedEvent: orderId=%s", evt.OrderID)
			if err := c.orderSvc.ConfirmOrder(ctx, evt.OrderID); err != nil {
				log.Printf("[order-service] confirm order error: %v", err)
			}
		}
	}
}

func (c *Consumer) consumeCreditReservationFailed(ctx context.Context) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		Topic:    "user.credit-reservation-failed",
		GroupID:  "order-service-group",
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
				log.Printf("[order-service] user.credit-reservation-failed read error: %v", err)
				continue
			}
			var evt events.UserCreditReservationFailedEvent
			if err := json.Unmarshal(msg.Value, &evt); err != nil {
				log.Printf("[order-service] user.credit-reservation-failed unmarshal error: %v", err)
				continue
			}
			log.Printf("[order-service] Received UserCreditReservationFailedEvent: orderId=%s reason=%s", evt.OrderID, evt.Reason)
			if err := c.orderSvc.CancelOrder(ctx, evt.OrderID); err != nil {
				log.Printf("[order-service] cancel order error: %v", err)
			}
		}
	}
}
