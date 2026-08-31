package httpx

import (
	"log/slog"
	"time"

	"ipw/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// NewServer builds a configured Fiber app with the base middleware stack.
// Feature routes are attached by the caller onto the returned *fiber.App.
func NewServer(cfg config.Config, log *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler:          ErrorHandler,
		DisableStartupMessage: true,
		ReadTimeout:           15 * time.Second,
		WriteTimeout:          30 * time.Second,
		BodyLimit:             8 * 1024 * 1024, // 8 MiB; large uploads go straight to object storage
	})

	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(requestLogger(log))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     joinOrigins(cfg.HTTP.AllowedOrigins),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return OK(c, fiber.Map{"status": "ok"})
	})

	return app
}

func requestLogger(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.Info("http_request",
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("took", time.Since(start)),
			slog.String("request_id", c.Locals("requestid").(string)),
			slog.String("ip", c.IP()),
		)
		return err
	}
}

func joinOrigins(origins []string) string {
	out := ""
	for i, o := range origins {
		if i > 0 {
			out += ","
		}
		out += o
	}
	if out == "" {
		return "http://localhost:3000"
	}
	return out
}
