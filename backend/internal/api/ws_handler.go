package api

import (
	"net/http"

	"github.com/Nap20192/hacknu/internal/hub"
	"github.com/Nap20192/hacknu/internal/spec"
	"github.com/google/uuid"
)

// ServeWS upgrades an HTTP connection to WebSocket and registers it with the hub.
// gorilla/websocket requires net/http — mounted via gofiber/adaptor in handler.go.
func ServeWS(h *hub.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeWS(w, r, uuid.New())
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
