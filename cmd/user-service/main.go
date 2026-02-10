// User-service: manages users and their balances; listens to order events for saga.
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
	"go_example/cmd/user-service/config"
	"go_example/cmd/user-service/handler"
	"go_example/cmd/user-service/kafka"
	"go_example/cmd/user-service/repository"
	"go_example/cmd/user-service/service"
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

	userRepo := repository.NewUserRepository(pool)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc, cfg.OrderServiceURL)

	consumer := kafka.NewConsumer(userSvc, cfg.Kafka.Brokers)
	defer consumer.Close()

	metrics.RegisterHTTPMetrics("user-service")

	app := fiber.New()
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(metrics.HTTPMiddleware())
	app.Get("/metrics", metrics.MetricsHandler())
	app.Get("/health", func(c fiber.Ctx) error { return c.JSON(fiber.Map{"status": "UP"}) })
	app.Post("/users", userHandler.CreateUser)
	app.Get("/users/:id/orders", userHandler.GetUserWithOrders)
	app.Get("/users/:id", userHandler.GetByID)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := app.Listen(":"+cfg.ServerPort, fiber.ListenConfig{DisableStartupMessage: true}); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	go consumer.Run(ctx)

	<-ctx.Done()
	log.Println("user-service shutting down")
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

// advisoryLockID ensures only one instance runs migrations when multiple share the same DB.
const advisoryLockID int64 = 0x75736572 // "user"

func runMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	// Single migration runner when multiple instances share the same DB.
	_, err = conn.Exec(ctx, "SELECT pg_advisory_lock($1)", advisoryLockID)
	if err != nil {
		return err
	}
	defer conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", advisoryLockID)
	data, err := migrationsFS.ReadFile("migrations/000001_create_users_table.up.sql")
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, string(data))
	return err
}
