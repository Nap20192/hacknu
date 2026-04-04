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
