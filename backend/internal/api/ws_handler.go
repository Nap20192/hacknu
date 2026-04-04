package api

import (
	"github.com/Nap20192/hacknu/internal/hub"
	"github.com/Nap20192/hacknu/internal/spec"
	fiberws "github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

// WSHandler returns a Fiber WebSocket handler that registers the upgraded conn with the hub.
func WSHandler(h *hub.Manager) func(*fiberws.Conn) {
	return func(c *fiberws.Conn) {
		h.ServeWS(c, uuid.New())
	}
}

// snapshotToDTO converts spec.HealthSnapshot → REST response DTO.
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
