// Order-service: manages orders and saga coordination via Kafka.
package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"

	"go_example/internal/metrics"
	"go_example/cmd/order-service/config"
	"go_example/cmd/order-service/handler"
	"go_example/cmd/order-service/kafka"
	"go_example/cmd/order-service/repository"
	"go_example/cmd/order-service/service"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DB.DSN())
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	if err := runMigrations(pool); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	producer := kafka.NewProducer(cfg.Kafka.Brokers)
	defer producer.Close()

	orderRepo := repository.NewOrderRepository(pool)
	orderSvc := service.NewOrderService(orderRepo, producer)
	orderHandler := handler.NewOrderHandler(orderSvc)

	consumer := kafka.NewConsumer(orderSvc, cfg.Kafka.Brokers)

	metrics.RegisterHTTPMetrics("order-service")

	app := fiber.New()
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(metrics.HTTPMiddleware())
	app.Get("/metrics", metrics.MetricsHandler())
	app.Get("/health", func(c fiber.Ctx) error { return c.JSON(fiber.Map{"status": "UP"}) })
	app.Post("/orders", orderHandler.CreateOrder)
	app.Get("/orders", orderHandler.ListByUserID)
	app.Get("/orders/:id", orderHandler.GetByID)
	app.Delete("/orders/:id", orderHandler.CancelOrder)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := app.Listen(":"+cfg.ServerPort, fiber.ListenConfig{DisableStartupMessage: true}); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	go consumer.Run(ctx)

	<-ctx.Done()
	log.Println("order-service shutting down")
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

// advisoryLockID ensures only one instance runs migrations when multiple share the same DB.
const advisoryLockID int64 = 0x6f72646572 // "order"

func runMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	_, err = conn.Exec(ctx, "SELECT pg_advisory_lock($1)", advisoryLockID)
	if err != nil {
		return err
	}
	defer conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", advisoryLockID)
	data, err := migrationsFS.ReadFile("migrations/000001_create_orders_table.up.sql")
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, string(data))
	return err
}
