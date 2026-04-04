package domain

type Metric struct {
	Name  string  `json:"n"`
	Value float64 `json:"v"`
}

type TelemetryBatch struct {
	LocoID  string   `json:"loco_id"`
	TS      int64    `json:"ts"`
	Payload []Metric `json:"payload"`
}
