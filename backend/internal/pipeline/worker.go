package pipeline

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"time"

	"github.com/Nap20192/hacknu/gen/sqlc"
	"github.com/Nap20192/hacknu/internal/domain"
	"github.com/Nap20192/hacknu/internal/hub"
	"github.com/Nap20192/hacknu/internal/spec"
	"github.com/google/uuid"
)

const chanSize = 256 // буфер входящего канала воркера

// agregatorWorker — одна горутина на один loco_id.
// Обрабатывает сырые батчи по пайплайну:
// validate → deduplicate → buffer → flush(normalize) → engine → persist+broadcast.
type agregatorWorker struct {
	locoID   uuid.UUID
	in       chan domain.TelemetryBatch
	rawIn    chan []byte // оригинальные JSON-байты для хранения в telemetry_events.raw
	buffer   *MetricBuffer
	lastTS   time.Time // для дедупликации по ts
	registry *spec.RuleRegistry
	engine   *spec.Engine
	queries  sqlc.Querier
	hub      *hub.Manager
	ticker   *time.Ticker
	done     chan struct{}
}

func newWorker(
	locoID uuid.UUID,
	registry *spec.RuleRegistry,
	engine *spec.Engine,
	queries sqlc.Querier,
	h *hub.Manager,
	bufCap int,
	flushEvery time.Duration,
) *agregatorWorker {
	return &agregatorWorker{
		locoID:   locoID,
		in:       make(chan domain.TelemetryBatch, chanSize),
		rawIn:    make(chan []byte, chanSize),
		buffer:   newMetricBuffer(bufCap),
		registry: registry,
		engine:   engine,
		queries:  queries,
		hub:      h,
		ticker:   time.NewTicker(flushEvery),
		done:     make(chan struct{}),
	}
}

func (w *agregatorWorker) run() {
	slog.Info("aggregator worker started", "loco_id", w.locoID)
	defer w.ticker.Stop()

	var pendingRaw []byte // последний сырой JSON для persist

	for {
		select {
		case <-w.done:
			slog.Info("aggregator worker stopped", "loco_id", w.locoID)
			return

		case raw := <-w.rawIn:
			pendingRaw = raw // сохраняем байты до следующего flush

		case batch, ok := <-w.in:
			if !ok {
				return
			}
			w.ingest(batch)

		case <-w.ticker.C:
			// Time-based flush: сбрасываем буфер даже если cap не достигнут
			w.flush(time.Now(), pendingRaw)
			pendingRaw = nil
		}
	}
}

// ingest выполняет validate + deduplicate + buffer.
// Если буфер заполнился — немедленно вызывает flush.
func (w *agregatorWorker) ingest(batch domain.TelemetryBatch) {
	// Deduplicate: одинаковый ts от одного локомотива — пропустить
	if !batch.TS.IsZero() && batch.TS.Equal(w.lastTS) {
		slog.Debug("aggregator: duplicate ts, skipping", "loco_id", w.locoID)
		return
	}
	w.lastTS = batch.TS

	full := false
	for i := range batch.Payload {
		if !w.validate(batch.Payload[i]) {
			continue
		}
		if w.buffer.add(batch.Payload[i]) {
			full = true
		}
	}

	// Capacity-based flush
	if full {
		// Забираем последний raw из канала без блокировки
		var raw []byte
		select {
		case raw = <-w.rawIn:
		default:
		}
		w.flush(batch.TS, raw)
	}
}

// validate отсекает физически невозможные и некорректные значения.
func (w *agregatorWorker) validate(m domain.Metric) bool {
	if math.IsNaN(m.Value) || math.IsInf(m.Value, 0) {
		slog.Debug("validate: NaN/Inf dropped", "metric", m.Name, "loco_id", w.locoID)
		return false
	}
	rule, ok := w.registry.Lookup(m.Name)
	if !ok {
		return true // нет правила — пропускаем без проверки границ
	}
	if rule.PhysicalMin != nil && float32(m.Value) < *rule.PhysicalMin {
		slog.Debug("validate: below physical_min", "metric", m.Name, "value", m.Value)
		return false
	}
	if rule.PhysicalMax != nil && float32(m.Value) > *rule.PhysicalMax {
		slog.Debug("validate: above physical_max", "metric", m.Name, "value", m.Value)
		return false
	}
	return true
}

// flush нормализует буфер, прогоняет через Engine, сохраняет в БД и отправляет в hub.
func (w *agregatorWorker) flush(ts time.Time, rawBytes []byte) {
	normalized := w.buffer.flush(func(name string) float32 {
		rule, ok := w.registry.Lookup(name)
		if !ok {
			return defaultEmaAlpha
		}
		return rule.EmaAlpha
	})
	if len(normalized) == 0 {
		return
	}

	batch := domain.TelemetryBatch{LocoID: w.locoID, TS: ts, Payload: normalized}

	// Engine: правила → issues → HealthSnapshot
	snap := w.engine.Process(batch)

	slog.Debug("aggregator flush",
		"loco_id", w.locoID,
		"metrics", len(normalized),
		"state", snap.State,
		"score", snap.Score,
		"issues", len(snap.Issues),
	)

	// Persist в фоне — не блокируем flush
	go w.persist(batch, snap, rawBytes)

	// Broadcast: LocoUpdate = health + нормализованные метрики для графиков
	update := buildLocoUpdate(snap, normalized)
	out, err := json.Marshal(update)
	if err != nil {
		slog.Error("aggregator: marshal LocoUpdate failed", "err", err)
		return
	}
	w.hub.Broadcast(out)
}

// persist сохраняет сырую телеметрию и снапшот здоровья в БД.
func (w *agregatorWorker) persist(batch domain.TelemetryBatch, snap spec.HealthSnapshot, rawBytes []byte) {
	ctx := context.Background()

	// Нормализованные метрики
	metricsJSON, _ := json.Marshal(batch.Payload)

	// Оригинальный сырой фрейм (если был передан)
	if rawBytes == nil {
		rawBytes = metricsJSON
	}

	if _, err := w.queries.InsertTelemetryEvent(ctx, sqlc.InsertTelemetryEventParams{
		LocomotiveID: batch.LocoID,
		Ts:           batch.TS,
		Metrics:      metricsJSON,
		Raw:          rawBytes,
	}); err != nil {
		slog.Warn("aggregator: insert telemetry_event failed", "loco_id", w.locoID, "err", err)
	}

	factorsJSON, _ := json.Marshal(snap.Issues)
	if _, err := w.queries.InsertHealthSnapshot(ctx, sqlc.InsertHealthSnapshotParams{
		LocomotiveID: snap.LocoID,
		Ts:           snap.Ts,
		Score:        snap.Score,
		Category:     snap.State.String(),
		Factors:      factorsJSON,
		MetricsSnap:  metricsJSON,
	}); err != nil {
		slog.Warn("aggregator: insert health_snapshot failed", "loco_id", w.locoID, "err", err)
	}
}

func (w *agregatorWorker) stop() {
	close(w.done)
}
