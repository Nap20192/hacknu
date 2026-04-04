package domain

import "time"

type Metric struct {
	Name  string  `json:"n"`
	Value float64 `json:"v"`
}

type TelemetryBatch struct {
	LocoID  string    `json:"loco_id"`
	TS      time.Time `json:"ts"`
	Payload []Metric  `json:"payload"`
}

// IssueLevel represents the severity of a detected problem.
type IssueLevel uint8

const (
	LevelInfo     IssueLevel = iota // informational, no state impact
	LevelWarning                    // performance degradation risk
	LevelCritical                   // immediate action required
)

func (l IssueLevel) String() string {
	switch l {
	case LevelInfo:
		return "Info"
	case LevelWarning:
		return "Warning"
	case LevelCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// Issue describes a single detected problem for a metric.
type Issue struct {
	Code         string // unique rule code, e.g. "CRIT_ABOVE_ENGINE_TEMP"
	Level        IssueLevel
	Target       string  // metric name that triggered the issue
	Message      string  // human-readable description for logs / UI
	HealthWeight float32 // copied from MetricRule; used for cumulative Degraded calc
}

// MetricRule mirrors the thresholds stored in the metric_definitions table.
type MetricRule struct {
	Name         string
	WarnAbove    *float32
	WarnBelow    *float32
	CritAbove    *float32
	CritBelow    *float32
	HealthWeight float32
}
