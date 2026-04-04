package api

import (
	"github.com/Nap20192/hacknu/internal/hub"
	fiberws "github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

// WSHandler registers the connection as a PRODUCER (simulator -> server).
// Reads incoming telemetry frames and routes them to the aggregator via hub.
func WSHandler(h *hub.Manager) func(*fiberws.Conn) {
	return func(c *fiberws.Conn) {
		h.ServeWS(c, uuid.New())
	}
}

// LiveHandler registers the connection as a CONSUMER (server -> dashboard).
// The client only receives broadcast LocoUpdate frames; any writes from the
// client are silently ignored so the hub read-channel stays clean.
func LiveHandler(h *hub.Manager) func(*fiberws.Conn) {
	return func(c *fiberws.Conn) {
		h.ServeLive(c, uuid.New())
	}
}
