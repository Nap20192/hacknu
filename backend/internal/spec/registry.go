package spec

import (
	"sync"

	"github.com/Nap20192/hacknu/gen/sqlc"
	"github.com/Nap20192/hacknu/internal/domain"
)

// RuleRegistry is a thread-safe, in-memory cache of domain.MetricRule entries
// loaded from the metric_definitions table.
//
// The map is replaced atomically on each Refresh call, so concurrent
// readers (Engine.Process goroutines) always see a consistent snapshot
// without holding a lock during the hot diagnostic loop.
type RuleRegistry struct {
	mu    sync.RWMutex
	rules map[string]domain.MetricRule
}

// NewRuleRegistry returns an empty registry. Call Refresh before first use.
func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{rules: make(map[string]domain.MetricRule)}
}

// Refresh atomically replaces the rule set with a fresh snapshot from DB.
// Intended to be called once at startup and then periodically (e.g. every 60 s)
// to pick up threshold changes without restarting the service.
func (r *RuleRegistry) Refresh(defs []sqlc.MetricDefinition) {
	next := make(map[string]domain.MetricRule, len(defs))
	for _, d := range defs {
		next[d.Name] = domain.MetricRule{
			Name:         d.Name,
			WarnAbove:    d.WarnAbove,
			WarnBelow:    d.WarnBelow,
			CritAbove:    d.CritAbove,
			CritBelow:    d.CritBelow,
			HealthWeight: d.HealthWeight,
		}
	}
	r.mu.Lock()
	r.rules = next
	r.mu.Unlock()
}

// Lookup returns the rule for a metric name.
// Returns (rule, true) if found, (zero, false) if the metric has no rule in DB.
// Callers that receive false should simply skip the metric.
func (r *RuleRegistry) Lookup(name string) (domain.MetricRule, bool) {
	r.mu.RLock()
	rule, ok := r.rules[name]
	r.mu.RUnlock()
	return rule, ok
}
