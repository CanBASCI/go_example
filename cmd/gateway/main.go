// Gateway: reverse proxy with round-robin for user-service, single upstream for order-service.
package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/proxy"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

const (
	defaultPort = "8080"
	userSvc1   = "http://user-service-1:8081"
	userSvc2   = "http://user-service-2:8082"
	orderSvc   = "http://order-service:8091"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	app := fiber.New()
	app.Use(recover.New())
	app.Use(logger.New())

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "UP"})
	})

	// Round-robin across user-service instances
	app.All("/users", proxy.BalancerForward([]string{userSvc1, userSvc2}))
	app.All("/users/*", proxy.BalancerForward([]string{userSvc1, userSvc2}))

	// Order-service: use Do with base+path because Forward does not preserve path
	app.All("/orders", func(c fiber.Ctx) error {
		return proxy.Do(c, orderSvc+c.Path())
	})
	app.All("/orders/*", func(c fiber.Ctx) error {
		return proxy.Do(c, orderSvc+c.Path())
	})

	log.Printf("gateway listening on :%s", port)
	if err := app.Listen(":"+port, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
		log.Fatalf("gateway: %v", err)
	}
}
