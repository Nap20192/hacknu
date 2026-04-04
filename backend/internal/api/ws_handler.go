package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Nap20192/hacknu/internal/domain"
	"github.com/Nap20192/hacknu/internal/hub"
	"github.com/Nap20192/hacknu/internal/spec"
	"github.com/google/uuid"
)

// TelemetryProcessor reads raw WebSocket frames from the hub, decodes them
// as TelemetryBatch, runs them through the diagnostic Engine, then broadcasts
// the resulting HealthSnapshot JSON back to all connected dashboard clients.
//
// This function blocks and should be run in its own goroutine.
// Cancel ctx to stop processing.
type TelemetryProcessor struct {
	hub      *hub.Manager
	engine   *spec.Engine
}

// NewTelemetryProcessor wires the hub and engine together.
func NewTelemetryProcessor(h *hub.Manager, e *spec.Engine) *TelemetryProcessor {
	return &TelemetryProcessor{hub: h, engine: e}
}

// Run reads from hub.ReadChannel() until it is closed.
func (p *TelemetryProcessor) Run() {
	slog.Info("TelemetryProcessor started")
	for msg := range p.hub.ReadChannel() {
		p.handle(msg)
	}
	slog.Info("TelemetryProcessor stopped")
}

func (p *TelemetryProcessor) handle(msg hub.ReadFromWs) {
	var req TelemetryBatchRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		slog.Warn("ws: malformed telemetry batch", "producer", msg.ProducerID, "err", err)
		return
	}

	if req.LocoID == "" || len(req.Payload) == 0 {
		slog.Warn("ws: empty batch, skipping", "producer", msg.ProducerID)
		return
	}

	// Convert DTO → domain
	metrics := make([]domain.Metric, len(req.Payload))
	for i, m := range req.Payload {
		metrics[i] = domain.Metric{Name: m.Name, Value: m.Value}
	}
	ts := req.Ts
	if ts.IsZero() {
		ts = time.Now()
	}
	batch := domain.TelemetryBatch{
		LocoID:  req.LocoID,
		TS:      ts,
		Payload: metrics,
	}

	snap := p.engine.Process(batch)

	out, err := json.Marshal(snapshotToDTO(snap))
	if err != nil {
		slog.Error("ws: failed to marshal snapshot", "err", err)
		return
	}

	// Broadcast to all connected dashboard clients.
	p.hub.Broadcast(out)
}

// ServeWS upgrades an HTTP connection to WebSocket and registers it with the hub.
// gorilla/websocket requires net/http — mount via gofiber/adaptor.
func ServeWS(h *hub.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeWS(w, r, uuid.New())
	}
}

// snapshotToDTO converts the internal spec.HealthSnapshot to an API DTO.
func snapshotToDTO(s spec.HealthSnapshot) HealthSnapshotDTO {
	issues := make([]IssueDTO, len(s.Issues))
	for i, iss := range s.Issues {
		issues[i] = IssueDTO{
			Code:         iss.Code,
			Level:        iss.Level.String(),
			Target:       iss.Target,
			Message:      iss.Message,
			HealthWeight: iss.HealthWeight,
		}
	}
	return HealthSnapshotDTO{
		LocomotiveID: s.LocoID,
		Ts:           s.Ts,
		State:        s.State.String(),
		Score:        s.Score,
		Category:     stateToCategory(s.State),
		Issues:       issues,
	}
}

func stateToCategory(s spec.LocoState) string {
	switch s {
	case spec.StateEmergency:
		return "Critical"
	case spec.StateDegraded:
		return "Warning"
	case spec.StateMaintenance:
		return "Maintenance"
	default:
		return "Normal"
	}
}
