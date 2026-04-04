package api

import (
	"net/http"

	"github.com/Nap20192/hacknu/gen/sqlc"
	"github.com/Nap20192/hacknu/internal/hub"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	fiberSwagger "github.com/gofiber/swagger"
)

// NewApp builds and returns a configured Fiber application.
//
//	@title			Locomotive Digital Twin API
//	@version		1.0
//	@description	Real-time telemetry ingestion, health index, and diagnostic API for locomotive digital twin.
//	@contact.name	HackNU Team
//	@host			localhost:8080
//	@BasePath		/
func NewApp(q *sqlc.Queries, h *hub.Manager) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(Response[any]{Error: err.Error()})
		},
	})

	// ── Middleware ────────────────────────────────────────────────────────────
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "${time} ${method} ${path} ${status} ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// ── Health check ──────────────────────────────────────────────────────────
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// ── Swagger UI ────────────────────────────────────────────────────────────
	// Access at http://localhost:8080/swagger/index.html
	// Generate docs: go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go
	app.Get("/swagger/*", fiberSwagger.HandlerDefault)

	// ── WebSocket telemetry ingest ────────────────────────────────────────────
	// gorilla/websocket needs net/http — bridge via adaptor.
	// Clients (simulator/device) connect here and send TelemetryBatchRequest JSON.
	// HealthSnapshot JSON is broadcast back to all connected clients.
	app.Get("/ws/telemetry", adaptor.HTTPHandlerFunc(wsHandler(h)))

	// ── REST routes ───────────────────────────────────────────────────────────
	registerRoutes(app, q)

	return app
}

// wsHandler wraps hub.ServeWS as a standard http.HandlerFunc.
func wsHandler(h *hub.Manager) http.HandlerFunc {
	return ServeWS(h)
}
