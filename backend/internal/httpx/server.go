package httpx

import (
	"errors"
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

		// The central ErrorHandler runs after this middleware returns, so derive
		// the eventual status from the error rather than the not-yet-written
		// response.
		status := c.Response().StatusCode()
		if err != nil {
			var de *DomainError
			var fe *fiber.Error
			switch {
			case errors.As(err, &de):
				status = de.Status
			case errors.As(err, &fe):
				status = fe.Code
			default:
				status = fiber.StatusInternalServerError
			}
		}

		attrs := []any{
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("took", time.Since(start)),
			slog.String("request_id", requestID(c)),
			slog.String("ip", c.IP()),
		}
		if err != nil && status >= 500 {
			log.Error("http_request", append(attrs, slog.Any("err", err))...)
		} else {
			log.Info("http_request", attrs...)
		}
		return err
	}
}

func requestID(c *fiber.Ctx) string {
	if id, ok := c.Locals("requestid").(string); ok {
		return id
	}
	return ""
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
