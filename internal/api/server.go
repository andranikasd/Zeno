package api

import (
	"database/sql"

	"github.com/andranikasd/Zeno/internal/config"
	"github.com/gofiber/fiber/v2"
)

func StartServer(cfg *config.Config, db *sql.DB) error {
	app := fiber.New()

	app.Get("/services", func(c *fiber.Ctx) error {
		return c.JSON([]string{"AmazonEC2", "AmazonS3"}) // Stub
	})

	app.Get("/costs", func(c *fiber.Ctx) error {
		from := c.Query("from")
		to := c.Query("to")
		service := c.Query("service")
		return c.JSON(fiber.Map{
			"from": from,
			"to": to,
			"service": service,
			"cost": 123.45, // Stub
		})
	})

	return app.Listen(cfg.ServerAddr)
}
