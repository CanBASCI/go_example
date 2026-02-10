// Gateway: reverse proxy with round-robin for user-service, single upstream for order-service.
package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/proxy"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"go_example/cmd/gateway/config"
	"go_example/internal/metrics"
)

func main() {
	cfg := config.Load()

	metrics.RegisterHTTPMetrics("gateway")

	app := fiber.New()
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(metrics.HTTPMiddleware())
	app.Get("/metrics", metrics.MetricsHandler())
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "UP"})
	})

	if len(cfg.UserServiceURLs) == 0 {
		log.Fatal("gateway: at least one USER_SERVICE_URLS entry required")
	}
	app.All("/users", proxy.BalancerForward(cfg.UserServiceURLs))
	app.All("/users/*", proxy.BalancerForward(cfg.UserServiceURLs))

	orderSvc := cfg.OrderServiceURL
	app.All("/orders", func(c fiber.Ctx) error {
		return proxy.Do(c, orderSvc+c.Path())
	})
	app.All("/orders/*", func(c fiber.Ctx) error {
		return proxy.Do(c, orderSvc+c.Path())
	})

	log.Printf("gateway listening on :%s", cfg.Port)
	if err := app.Listen(":"+cfg.Port, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
		log.Fatalf("gateway: %v", err)
	}
}
