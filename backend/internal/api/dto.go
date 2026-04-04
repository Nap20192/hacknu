package api

import (
	"time"

	"github.com/google/uuid"
)

// ---- generic envelope ----

// Response is the standard API envelope.
type Response[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// PagedResponse wraps paginated results.
type PagedResponse[T any] struct {
	Success bool `json:"success"`
	Data    []T  `json:"data"`
	Total   int  `json:"total"`
}

// ---- locomotives ----

// LocomotiveDTO is the API representation of a locomotive.
type LocomotiveDTO struct {
	ID           uuid.UUID  `json:"id"`
	DisplayName  string     `json:"display_name"`
	LocoType     string     `json:"loco_type"`
	RegisteredAt time.Time  `json:"registered_at"`
	LastSeenAt   *time.Time `json:"last_seen_at,omitempty"`
	Active       bool       `json:"active"`
}

// ---- metric definitions ----

// MetricDefinitionDTO is the API representation of a metric definition.
type MetricDefinitionDTO struct {
	Name         string   `json:"name"`
	Display      string   `json:"display"`
	Description  string   `json:"description"`
	Unit         string   `json:"unit"`
	WarnAbove    *float32 `json:"warn_above,omitempty"`
	WarnBelow    *float32 `json:"warn_below,omitempty"`
	CritAbove    *float32 `json:"crit_above,omitempty"`
	CritBelow    *float32 `json:"crit_below,omitempty"`
	HealthWeight float32  `json:"health_weight"`
}

// ---- alerts ----

// AlertDTO is the API representation of an alert.
type AlertDTO struct {
	ID             int64      `json:"id"`
	LocomotiveID   string     `json:"locomotive_id"`
	TriggeredAt    time.Time  `json:"triggered_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	Severity       string     `json:"severity"`
	Code           string     `json:"code"`
	MetricName     *string    `json:"metric_name,omitempty"`
	MetricValue    *float32   `json:"metric_value,omitempty"`
	Threshold      *float32   `json:"threshold,omitempty"`
	Message        string     `json:"message"`
	Recommendation string     `json:"recommendation"`
	Acknowledged   bool       `json:"acknowledged"`
}

