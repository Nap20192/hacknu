package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/Nap20192/hacknu/gen/sqlc"
	"github.com/Nap20192/hacknu/internal/domain"
	"github.com/Nap20192/hacknu/internal/spec"
)

// SimulationRepository is the subset of sqlc.Queries used by the simulator.
type SimulationRepository interface {
	InsertTelemetryEvent(ctx context.Context, arg sqlc.InsertTelemetryEventParams) (sqlc.TelemetryEvent, error)
	InsertHealthSnapshot(ctx context.Context, arg sqlc.InsertHealthSnapshotParams) (sqlc.HealthSnapshot, error)
	InsertAlert(ctx context.Context, arg sqlc.InsertAlertParams) (sqlc.Alert, error)
	ListMetricDefinitions(ctx context.Context) ([]sqlc.MetricDefinition, error)
	UpsertLocomotive(ctx context.Context, arg sqlc.UpsertLocomotiveParams) (sqlc.Locomotive, error)
	UpdateLocomotiveLastSeen(ctx context.Context, arg sqlc.UpdateLocomotiveLastSeenParams) (sqlc.Locomotive, error)
}

// metricState tracks the simulated current value for one metric.
type metricState struct {
	def     sqlc.MetricDefinition
	current float64
}

func (ms *metricState) normalMid() float64 {
	lo := f32OrDefault(ms.def.NormalMin, ms.def.PhysicalMin, 0)
	hi := f32OrDefault(ms.def.NormalMax, ms.def.PhysicalMax, 100)
	return (float64(lo) + float64(hi)) / 2.0
}

func (ms *metricState) normalRange() (float64, float64) {
	lo := f32OrDefault(ms.def.NormalMin, ms.def.PhysicalMin, 0)
	hi := f32OrDefault(ms.def.NormalMax, ms.def.PhysicalMax, 100)
	return float64(lo), float64(hi)
}

// issueScenario forces one metric toward a threshold violation, then returns it to normal.
type issueScenario struct {
	metricName  string
	targetValue float64 // push toward this (above/below a threshold)
	returnTo    float64 // mid-point to recover to
	pushing     bool    // true = ramping toward target; false = returning to normal
	ticksLeft   int
}

// SimulationService generates realistic telemetry and writes it to the database.
type SimulationService struct {
	repo     SimulationRepository
	registry *spec.RuleRegistry
	engine   *spec.Engine
	locoID   string
	interval time.Duration
	states   map[string]*metricState
	scenario *issueScenario
	rng      *rand.Rand
}

// NewSimulationService creates a service wired to the given repository.
// locoID is the locomotive identifier written into every DB row.
// interval controls how often a tick fires (e.g. time.Second for 1 Hz).
func NewSimulationService(repo SimulationRepository, locoID string, interval time.Duration) *SimulationService {
	registry := spec.NewRuleRegistry()
	engine := spec.NewEngine(registry, spec.ThresholdSpecification{})
	return &SimulationService{
		repo:     repo,
		registry: registry,
		engine:   engine,
		locoID:   locoID,
		interval: interval,
		states:   make(map[string]*metricState),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Run loads metric definitions from the DB, seeds locomotive, then starts
// the simulation loop. Blocks until ctx is cancelled.
func (s *SimulationService) Run(ctx context.Context) error {
	defs, err := s.repo.ListMetricDefinitions(ctx)
	if err != nil {
		return fmt.Errorf("load metric definitions: %w", err)
	}
	if len(defs) == 0 {
		return fmt.Errorf("no metric definitions found in DB; seed them first")
	}

	for i := range defs {
		ms := &metricState{def: defs[i]}
		ms.current = ms.normalMid()
		s.states[defs[i].Name] = ms
	}
	s.registry.Refresh(defs)

	now := time.Now()
	if _, err = s.repo.UpsertLocomotive(ctx, sqlc.UpsertLocomotiveParams{
		ID:          s.locoID,
		DisplayName: "Simulator-" + s.locoID,
		LocoType:    "sim",
		LastSeenAt:  &now,
		Active:      true,
	}); err != nil {
		return fmt.Errorf("upsert locomotive: %w", err)
	}

	slog.Info("simulation started",
		"loco_id", s.locoID,
		"metrics", len(defs),
		"interval", s.interval,
	)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	nextScenario := 20 + s.rng.Intn(20) // ticks until the first scenario
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
			if err := s.tick(ctx, t); err != nil {
				slog.Error("tick error", "err", err)
			}
		}
	}
}

// tick executes one simulation step: evolve values → run engine → write DB.
func (s *SimulationService) tick(ctx context.Context, t time.Time) error {
	if s.scenario != nil {
		s.stepScenario()
	}

	for _, ms := range s.states {
		s.evolveMetric(ms)
	}

	metrics := make([]domain.Metric, 0, len(s.states))
	metricsMap := make(map[string]float64, len(s.states))
	for name, ms := range s.states {
		metrics = append(metrics, domain.Metric{Name: name, Value: ms.current})
		metricsMap[name] = ms.current
	}

	batch := domain.TelemetryBatch{LocoID: s.locoID, TS: t, Payload: metrics}
	snapshot := s.engine.Process(batch)

	metricsJSON, _ := json.Marshal(metricsMap)

	if _, err := s.repo.InsertTelemetryEvent(ctx, sqlc.InsertTelemetryEventParams{
		LocomotiveID: s.locoID,
		Ts:           t,
		Metrics:      metricsJSON,
		Raw:          metricsJSON,
	}); err != nil {
		return fmt.Errorf("insert telemetry: %w", err)
	}

	factorsJSON, _ := marshalFactors(snapshot.Issues)
	if _, err := s.repo.InsertHealthSnapshot(ctx, sqlc.InsertHealthSnapshotParams{
		LocomotiveID: s.locoID,
		Ts:           t,
		Score:        snapshot.Score,
		Category:     stateCategory(snapshot.State),
		Factors:      factorsJSON,
		MetricsSnap:  metricsJSON,
	}); err != nil {
		return fmt.Errorf("insert health snapshot: %w", err)
	}

	for _, issue := range snapshot.Issues {
		if issue.Level == domain.LevelInfo {
			continue
		}
		severity := "warning"
		if issue.Level == domain.LevelCritical {
			severity = "critical"
		}
		mn := issue.Target
		var mv *float32
		if ms, ok := s.states[issue.Target]; ok {
			v := float32(ms.current)
			mv = &v
		}
		if _, err := s.repo.InsertAlert(ctx, sqlc.InsertAlertParams{
			LocomotiveID:   s.locoID,
			Severity:       severity,
			Code:           issue.Code,
			MetricName:     &mn,
			MetricValue:    mv,
			Threshold:      nil,
			Message:        issue.Message,
			Recommendation: recommendationFor(issue.Code),
		}); err != nil {
			slog.Warn("insert alert failed", "err", err)
		}
	}

	_, _ = s.repo.UpdateLocomotiveLastSeen(ctx, sqlc.UpdateLocomotiveLastSeenParams{
		ID:         s.locoID,
		LastSeenAt: &t,
	})

	slog.Debug("tick", "score", snapshot.Score, "state", snapshot.State.String(), "issues", len(snapshot.Issues))
	return nil
}

// evolveMetric applies a random walk with mean-reversion toward the normal mid-point.
// When an active scenario targets this metric, the value is driven toward the target.
func (s *SimulationService) evolveMetric(ms *metricState) {
	lo, hi := ms.normalRange()
	span := hi - lo
	if span <= 0 {
		span = 1
	}
	delta := (s.rng.Float64()*2 - 1) * span * 0.01 // ±1 % of normal span
	revert := (ms.normalMid() - ms.current) * 0.02 // 2 % pull toward centre

	if s.scenario != nil && s.scenario.metricName == ms.def.Name && s.scenario.pushing {
		drive := (s.scenario.targetValue - ms.current) * 0.15
		ms.current += drive + delta
	} else {
		ms.current += delta + revert
	}

	physLo := float64(f32OrVal(ms.def.PhysicalMin, -1e9))
	physHi := float64(f32OrVal(ms.def.PhysicalMax, 1e9))
	if ms.current < physLo {
		ms.current = physLo
	}
	if ms.current > physHi {
		ms.current = physHi
	}
}

// activateScenario picks one metric with a threshold and starts driving it there.
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

	physLo := float64(f32OrVal(ms.def.PhysicalMin, -1e9))
	physHi := float64(f32OrVal(ms.def.PhysicalMax, 1e9))
	if c.target > physHi {
		c.target = physHi
	}
	if c.target < physLo {
		c.target = physLo
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

// stepScenario advances the active scenario by one tick.
func (s *SimulationService) stepScenario() {
	scen := s.scenario
	scen.ticksLeft--

	if scen.pushing && scen.ticksLeft <= 0 {
		// Begin recovery phase.
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

// marshalFactors converts issues to the JSONB format stored in health_snapshots.factors.
func marshalFactors(issues []domain.Issue) ([]byte, error) {
	type factor struct {
		Code   string  `json:"code"`
		Level  string  `json:"level"`
		Metric string  `json:"metric"`
		Msg    string  `json:"msg"`
		Weight float32 `json:"weight"`
	}
	slice := make([]factor, len(issues))
	for i, iss := range issues {
		slice[i] = factor{
			Code:   iss.Code,
			Level:  iss.Level.String(),
			Metric: iss.Target,
			Msg:    iss.Message,
			Weight: iss.HealthWeight,
		}
	}
	return json.Marshal(slice)
}

func stateCategory(state spec.LocoState) string {
	switch state {
	case spec.StateEmergency:
		return "critical"
	case spec.StateDegraded, spec.StateMaintenance:
		return "warning"
	default:
		return "normal"
	}
}

// f32OrDefault returns the value of the first non-nil pointer, or the second,
// or the hard fallback.
func f32OrDefault(primary, secondary *float32, fallback float32) float32 {
	if primary != nil {
		return *primary
	}
	if secondary != nil {
		return *secondary
	}
	return fallback
}

func f32OrVal(p *float32, def float32) float32 {
	if p != nil {
		return *p
	}
	return def
}

func recommendationFor(code string) string {
	recs := map[string]string{
		"CRIT_ABOVE_ENGINE_TEMP_C":      "Снизить скорость, проверить охлаждение",
		"WARN_ABOVE_ENGINE_TEMP_C":      "Снизить нагрузку, следить за температурой",
		"CRIT_BELOW_BRAKE_PRESSURE_BAR": "Немедленная остановка",
		"WARN_BELOW_BRAKE_PRESSURE_BAR": "Проверить тормозную систему",
		"CRIT_BELOW_FUEL_LEVEL_PCT":     "Срочная заправка",
		"WARN_BELOW_FUEL_LEVEL_PCT":     "Запланировать заправку",
		"CRIT_BELOW_OIL_PRESSURE_BAR":   "Остановить двигатель, проверить масло",
		"WARN_BELOW_OIL_PRESSURE_BAR":   "Проверить уровень масла",
		"CRIT_ABOVE_TRACTION_AMPS":      "Снизить тяговое усилие немедленно",
		"WARN_ABOVE_TRACTION_AMPS":      "Снизить тяговое усилие",
		"CRIT_ABOVE_VOLTAGE_V":          "Проверить генераторы",
		"CRIT_BELOW_VOLTAGE_V":          "Проверить источники питания",
		"WARN_ABOVE_VOLTAGE_V":          "Проверить генераторы",
		"WARN_BELOW_VOLTAGE_V":          "Проверить источники питания",
		"CRIT_ABOVE_SPEED_KMH":          "Снизить скорость немедленно",
		"WARN_ABOVE_SPEED_KMH":          "Снизить скорость",
		"CRIT_ABOVE_AXLE_TEMP_C":        "Остановить поезд, проверить буксы",
		"WARN_ABOVE_AXLE_TEMP_C":        "Снизить скорость, следить за буксами",
	}
	if r, ok := recs[code]; ok {
		return r
	}
	return "Обратитесь к технику"
}
