package api

import "time"

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
