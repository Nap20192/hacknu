package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/Nap20192/hacknu/internal/domain"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// MetricDef holds the simulation parameters for one metric.
type MetricDef struct {
	Name        string
	PhysicalMin float32
	PhysicalMax float32
	NormalMin   float32
	NormalMax   float32
	WarnAbove   *float32
	WarnBelow   *float32
	CritAbove   *float32
	CritBelow   *float32
}

type metricState struct {
	def     MetricDef
	current float64
}

func (ms *metricState) normalMid() float64 {
	return (float64(ms.def.NormalMin) + float64(ms.def.NormalMax)) / 2.0
}

func (ms *metricState) normalRange() (float64, float64) {
	return float64(ms.def.NormalMin), float64(ms.def.NormalMax)
}

type issueScenario struct {
	metricName  string
	targetValue float64
	returnTo    float64
	pushing     bool
	ticksLeft   int
}

// SimulationService generates realistic telemetry and sends it via WebSocket.
// It does not connect to the database.
type SimulationService struct {
	wsURL    string
	locoID   uuid.UUID
	interval time.Duration
	defs     []MetricDef
	states   map[string]*metricState
	scenario *issueScenario
	rng      *rand.Rand
}

func NewSimulationService(wsURL string, locoID uuid.UUID, interval time.Duration, defs []MetricDef) *SimulationService {
	return &SimulationService{
		wsURL:    wsURL,
		locoID:   locoID,
		interval: interval,
		defs:     defs,
		states:   make(map[string]*metricState),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Run initialises metric states, connects to the WebSocket server and starts
// sending TelemetryBatch frames until ctx is cancelled.
func (s *SimulationService) Run(ctx context.Context) error {
	if len(s.defs) == 0 {
		return fmt.Errorf("no metric definitions provided")
	}
	for i := range s.defs {
		ms := &metricState{def: s.defs[i]}
		ms.current = ms.normalMid()
		s.states[s.defs[i].Name] = ms
	}

	slog.Info("simulation started", "loco_id", s.locoID, "metrics", len(s.defs), "interval", s.interval)

	conn, err := s.dialWithRetry(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	nextScenario := 20 + s.rng.Intn(20)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case t := <-ticker.C:
			nextScenario--
			if nextScenario <= 0 {
				s.activateScenario()
				nextScenario = 25 + s.rng.Intn(20)
			}
			if err := s.tick(conn, t); err != nil {
				slog.Warn("tick failed, reconnecting", "err", err)
				conn.Close()
				conn, err = s.dialWithRetry(ctx)
				if err != nil {
					return err
				}
			}
		}
	}
}

func (s *SimulationService) dialWithRetry(ctx context.Context) (*websocket.Conn, error) {
	for {
		conn, err := s.dial(ctx)
		if err == nil {
			return conn, nil
		}
		slog.Warn("ws connect failed, retrying in 3s", "err", err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}

func (s *SimulationService) dial(ctx context.Context) (*websocket.Conn, error) {
	dialer := websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	conn, _, err := dialer.DialContext(ctx, s.wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", s.wsURL, err)
	}
	slog.Info("ws: connected", "url", s.wsURL)
	return conn, nil
}

func (s *SimulationService) tick(conn *websocket.Conn, t time.Time) error {
	if s.scenario != nil {
		s.stepScenario()
	}
	for _, ms := range s.states {
		s.evolveMetric(ms)
	}

	metrics := make([]domain.Metric, 0, len(s.states))
	for name, ms := range s.states {
		metrics = append(metrics, domain.Metric{Name: name, Value: ms.current})
	}

	batch := domain.TelemetryBatch{LocoID: s.locoID, TS: t, Payload: metrics}
	payload, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("marshal batch: %w", err)
	}

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
		return fmt.Errorf("send batch: %w", err)
	}

	slog.Debug("tick sent", "loco_id", s.locoID, "metrics", len(metrics))
	return nil
}

func (s *SimulationService) evolveMetric(ms *metricState) {
	lo, hi := ms.normalRange()
	span := hi - lo
	if span <= 0 {
		span = 1
	}
	delta := (s.rng.Float64()*2 - 1) * span * 0.01
	revert := (ms.normalMid() - ms.current) * 0.02

	if s.scenario != nil && s.scenario.metricName == ms.def.Name && s.scenario.pushing {
		drive := (s.scenario.targetValue - ms.current) * 0.15
		ms.current += drive + delta
	} else {
		ms.current += delta + revert
	}

	physLo := float64(ms.def.PhysicalMin)
	physHi := float64(ms.def.PhysicalMax)
	if ms.current < physLo {
		ms.current = physLo
	}
	if ms.current > physHi {
		ms.current = physHi
	}
}

func (s *SimulationService) activateScenario() {
	type candidate struct {
		name   string
		target float64
	}
	var pool []candidate
	for name, ms := range s.states {
		d := ms.def
		if d.CritAbove != nil {
			pool = append(pool, candidate{name, float64(*d.CritAbove) * 1.05})
		}
		if d.CritBelow != nil {
			pool = append(pool, candidate{name, float64(*d.CritBelow) * 0.85})
		}
		if d.WarnAbove != nil {
			pool = append(pool, candidate{name, float64(*d.WarnAbove) * 1.03})
		}
		if d.WarnBelow != nil {
			pool = append(pool, candidate{name, float64(*d.WarnBelow) * 0.92})
		}
	}
	if len(pool) == 0 {
		return
	}

	c := pool[s.rng.Intn(len(pool))]
	ms := s.states[c.name]

	if c.target > float64(ms.def.PhysicalMax) {
		c.target = float64(ms.def.PhysicalMax)
	}
	if c.target < float64(ms.def.PhysicalMin) {
		c.target = float64(ms.def.PhysicalMin)
	}

	s.scenario = &issueScenario{
		metricName:  c.name,
		targetValue: c.target,
		returnTo:    ms.normalMid(),
		pushing:     true,
		ticksLeft:   10 + s.rng.Intn(10),
	}
	slog.Info("scenario activated", "metric", c.name, "target", c.target)
}

func (s *SimulationService) stepScenario() {
	scen := s.scenario
	scen.ticksLeft--
	if scen.pushing && scen.ticksLeft <= 0 {
		scen.pushing = false
		scen.targetValue = scen.returnTo
		scen.ticksLeft = 8 + s.rng.Intn(7)
		slog.Info("scenario returning to normal", "metric", scen.metricName)
		return
	}
	if !scen.pushing && scen.ticksLeft <= 0 {
		slog.Info("scenario finished", "metric", scen.metricName)
		s.scenario = nil
	}
}
