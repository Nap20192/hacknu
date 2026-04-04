package spec

import "github.com/Nap20192/hacknu/internal/domain"

// LocoState represents the operational state of a locomotive.
type LocoState uint8

const (
	StateOperational LocoState = iota // all systems nominal
	StateDegraded                     // power limited; maintenance recommended
	StateEmergency                    // immediate stop required
	StateMaintenance                  // taken out of service for scheduled work
)

func (s LocoState) String() string {
	switch s {
	case StateOperational:
		return "Operational"
	case StateDegraded:
		return "Degraded"
	case StateEmergency:
		return "Emergency"
	case StateMaintenance:
		return "Maintenance"
	default:
		return "Unknown"
	}
}

const (
	// degradedWarnCount: more than this many Warning issues → Degraded.
	degradedWarnCount = 3
	// degradedWarnWeight: cumulative health_weight of all Warning issues
	// exceeding this threshold also triggers Degraded state.
	degradedWarnWeight float32 = 0.5
)

// CalculateState is a pure, stateless function.
//
// REACTIVE semantics: the new state is derived exclusively from the *current*
// snapshot of issues — there is no "previous state" input. This means:
//   - If a critical sensor recovers, the state immediately drops back to
//     Operational (or Degraded if other warnings persist).
//   - Historical telemetry can be replayed through CalculateState to
//     reconstruct the exact sequence of state changes, enabling a full
//     "digital twin" audit trail.
//
// Transition priority (highest wins):
//  1. Any Critical issue          → Emergency
//  2. Warnings > degradedWarnCount
//     OR total warning weight > degradedWarnWeight → Degraded
//  3. No issues                   → Operational
func CalculateState(issues []domain.Issue) LocoState {
	var warnCount int
	var warnWeight float32

	for i := range issues {
		switch issues[i].Level {
		case domain.LevelCritical:
			// Short-circuit: one Critical is enough for Emergency.
			return StateEmergency
		case domain.LevelWarning:
			warnCount++
			warnWeight += issues[i].HealthWeight
		}
	}

	if warnCount > degradedWarnCount || warnWeight > degradedWarnWeight {
		return StateDegraded
	}

	return StateOperational
}
