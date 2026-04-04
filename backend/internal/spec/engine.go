package spec

import (
	"log/slog"
	"time"

	"github.com/Nap20192/hacknu/internal/domain"
	"github.com/google/uuid"
)

// HealthSnapshot is the result of one full diagnostic cycle.
// It is ready to be persisted to the health_snapshots TimescaleDB table.
type HealthSnapshot struct {
	LocoID uuid.UUID
	Ts     time.Time
	State  LocoState
	// Score is a 0–100 integer health index (100 = all green, 0 = catastrophic).
	Score  int16
	Issues []domain.Issue
}

// Engine executes the diagnostic pipeline for each incoming TelemetryBatch.
//
// Pipeline stages:
//  1. For every metric in the batch, look up its rule in the RuleRegistry.
//  2. Run the Specification against (metric, rule) → collect Issues.
//  3. Feed the Issue list into the Reactive State Machine → derive LocoState.
//  4. Compute a composite health score.
//  5. Return a HealthSnapshot ready for DB insertion.
type Engine struct {
	registry *RuleRegistry
	spec     Specification
}

// NewEngine creates an Engine with the given registry and specification.
// Inject ThresholdSpecification{} as the default spec implementation.
func NewEngine(registry *RuleRegistry, spec Specification) *Engine {
	return &Engine{registry: registry, spec: spec}
}

// Process is the hot path, called once per WebSocket batch.
//
// Allocation strategy:
//   - issues slice is pre-allocated to the batch payload length to avoid
//     incremental growth in the append loop (typical batch size: 100–500 metrics).
//   - Range-by-index (for i := range) avoids copying Metric structs.
func (e *Engine) Process(batch domain.TelemetryBatch) HealthSnapshot {
	issues := make([]domain.Issue, 0, len(batch.Payload))

	for i := range batch.Payload {
		rule, ok := e.registry.Lookup(batch.Payload[i].Name)
		if !ok {
			// Metric has no rule in DB → treat as informational, skip evaluation.
			continue
		}
		if issue := e.spec.Evaluate(batch.Payload[i], rule); issue != nil {
			issues = append(issues, *issue)
		}
	}

	// Reactive transition: state is derived purely from the current issue set.
	// No memory of the previous state is needed or used.
	state := CalculateState(issues)

	if len(issues) > 0 {
		issueCodes := make([]string, len(issues))
		for i := range issues {
			issueCodes[i] = issues[i].Code
		}
		slog.Info("diagnostic cycle complete",
			"loco_id", batch.LocoID,
			"state", state.String(),
			"issue_count", len(issues),
			"issues", issueCodes,
		)
	}

	return HealthSnapshot{
		LocoID: batch.LocoID,
		Ts:     batch.TS,
		State:  state,
		Score:  calcHealthScore(issues),
		Issues: issues,
	}
}

// calcHealthScore converts the issue list into a 0–100 health index.
//
// Penalty formula:
//   - Critical issue: weight × 200  (double-weight because it forces Emergency)
//   - Warning issue:  weight × 100
//
// The resulting penalty is clamped to [0, 100] and subtracted from 100.
func calcHealthScore(issues []domain.Issue) int16 {
	if len(issues) == 0 {
		return 100
	}
	var penalty float32
	for i := range issues {
		switch issues[i].Level {
		case domain.LevelCritical:
			penalty += issues[i].HealthWeight * 2
		case domain.LevelWarning:
			penalty += issues[i].HealthWeight
		}
	}
	score := 100 - int16(penalty*100)
	if score < 0 {
		return 0
	}
	return score
}
