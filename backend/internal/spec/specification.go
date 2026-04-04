package spec

import (
	"fmt"
	"strings"

	"github.com/Nap20192/hacknu/internal/domain"
)

// Specification evaluates a single metric against its rule.
// Returns *domain.Issue if a problem is detected, nil if the metric is within bounds.
//
// Returning a rich *Issue (instead of bool) keeps the caller free from
// knowing threshold details — all context lives in the Issue itself.
type Specification interface {
	Evaluate(metric domain.Metric, rule domain.MetricRule) *domain.Issue
}

// ThresholdSpecification checks the four threshold boundaries in priority order:
// critical-above → critical-below → warn-above → warn-below.
// First match wins; nil thresholds are skipped.
type ThresholdSpecification struct{}

func (ThresholdSpecification) Evaluate(metric domain.Metric, rule domain.MetricRule) *domain.Issue {
	v := float32(metric.Value)

	if rule.CritAbove != nil && v > *rule.CritAbove {
		return &domain.Issue{
			Code:         issueCode("CRIT_ABOVE", rule.Name),
			Level:        domain.LevelCritical,
			Target:       metric.Name,
			Message:      fmt.Sprintf("%s=%.4g exceeds critical upper bound %.4g", metric.Name, v, *rule.CritAbove),
			HealthWeight: rule.HealthWeight,
		}
	}
	if rule.CritBelow != nil && v < *rule.CritBelow {
		return &domain.Issue{
			Code:         issueCode("CRIT_BELOW", rule.Name),
			Level:        domain.LevelCritical,
			Target:       metric.Name,
			Message:      fmt.Sprintf("%s=%.4g is below critical lower bound %.4g", metric.Name, v, *rule.CritBelow),
			HealthWeight: rule.HealthWeight,
		}
	}
	if rule.WarnAbove != nil && v > *rule.WarnAbove {
		return &domain.Issue{
			Code:         issueCode("WARN_ABOVE", rule.Name),
			Level:        domain.LevelWarning,
			Target:       metric.Name,
			Message:      fmt.Sprintf("%s=%.4g exceeds warning upper bound %.4g", metric.Name, v, *rule.WarnAbove),
			HealthWeight: rule.HealthWeight,
		}
	}
	if rule.WarnBelow != nil && v < *rule.WarnBelow {
		return &domain.Issue{
			Code:         issueCode("WARN_BELOW", rule.Name),
			Level:        domain.LevelWarning,
			Target:       metric.Name,
			Message:      fmt.Sprintf("%s=%.4g is below warning lower bound %.4g", metric.Name, v, *rule.WarnBelow),
			HealthWeight: rule.HealthWeight,
		}
	}
	return nil
}

// issueCode produces a stable, uppercase code like "CRIT_ABOVE_ENGINE_TEMP".
func issueCode(prefix, metricName string) string {
	return prefix + "_" + strings.ToUpper(strings.ReplaceAll(metricName, ".", "_"))
}
