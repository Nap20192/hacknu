package pipeline

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/Nap20192/hacknu/gen/sqlc"
	"github.com/Nap20192/hacknu/internal/domain"
	"github.com/Nap20192/hacknu/internal/hub"
	"github.com/Nap20192/hacknu/internal/spec"
)

// LocoUpdate — JSON-фрейм, который сервер отправляет дашборду по WebSocket.
// Содержит и снапшот здоровья, и нормализованные метрики для обновления графиков.
type LocoUpdate struct {
	LocoID   string             `json:"loco_id"`
	Ts       time.Time          `json:"ts"`
	State    string             `json:"state"`
	Score    int16              `json:"score"`
	Category string             `json:"category"`
	Issues   []IssueWire        `json:"issues"`
	Metrics  map[string]float64 `json:"metrics"` // name → EMA-сглаженное значение
}

// IssueWire — сериализуемое представление одной проблемы.
type IssueWire struct {
	Code         string  `json:"code"`
	Level        string  `json:"level"`
	Target       string  `json:"target"`
	Message      string  `json:"message"`
	HealthWeight float32 `json:"health_weight"`
}

// AgregatorService читает сырые фреймы из hub, маршрутизирует по loco_id
// к воркерам и управляет их жизненным циклом.
type AgregatorService struct {
	queries       sqlc.Querier
	registry      *spec.RuleRegistry
	engine        *spec.Engine
	hub           *hub.Manager
	bufCap        int
	flushInterval time.Duration
	workers       map[string]*agregatorWorker
	mu            sync.Mutex
	wg            sync.WaitGroup
}

func NewAgregatorService(
	queries sqlc.Querier,
	registry *spec.RuleRegistry,
	engine *spec.Engine,
	h *hub.Manager,
	bufCap int,
	flushInterval time.Duration,
) *AgregatorService {
	return &AgregatorService{
		queries:       queries,
		registry:      registry,
		engine:        engine,
		hub:           h,
		bufCap:        bufCap,
		flushInterval: flushInterval,
		workers:       make(map[string]*agregatorWorker),
	}
}

// Run читает фреймы из hub.ReadChannel() до закрытия ctx.
// Вызывать в отдельной горутине.
func (s *AgregatorService) Run(ctx context.Context, readCh <-chan hub.ReadFromWs) {
	slog.Info("aggregator service started")
	for {
		select {
		case <-ctx.Done():
			s.shutdown()
			return
		case msg, ok := <-readCh:
			if !ok {
				s.shutdown()
				return
			}
			s.ingest(msg.Payload)
		}
	}
}

// ingest разбирает сырой JSON и отправляет батч в нужный воркер.
func (s *AgregatorService) ingest(raw []byte) {
	var batch domain.TelemetryBatch
	if err := json.Unmarshal(raw, &batch); err != nil {
		slog.Warn("aggregator: malformed batch", "err", err)
		return
	}
	if batch.LocoID == "" || len(batch.Payload) == 0 {
		slog.Debug("aggregator: empty batch, skipping")
		return
	}
	if batch.TS.IsZero() {
		batch.TS = time.Now()
	}

	s.dispatch(batch, raw)
}

// dispatch маршрутизирует батч к воркеру loco_id.
// Если воркер не существует — создаёт и запускает новую горутину.
func (s *AgregatorService) dispatch(batch domain.TelemetryBatch, raw []byte) {
	s.mu.Lock()
	w, ok := s.workers[batch.LocoID]
	if !ok {
		w = newWorker(batch.LocoID, s.registry, s.engine, s.queries, s.hub, s.bufCap, s.flushInterval)
		s.workers[batch.LocoID] = w
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			w.run()
		}()
		slog.Info("aggregator: new worker spawned", "loco_id", batch.LocoID)
	}
	s.mu.Unlock()

	// Сначала отправляем сырые байты (не блокируем если канал полный)
	select {
	case w.rawIn <- raw:
	default:
	}

	// Затем — распарсенный батч
	select {
	case w.in <- batch:
	default:
		slog.Warn("aggregator: worker channel full, batch dropped", "loco_id", batch.LocoID)
	}
}

// shutdown останавливает все воркеры и ждёт завершения.
func (s *AgregatorService) shutdown() {
	s.mu.Lock()
	for _, w := range s.workers {
		w.stop()
	}
	s.mu.Unlock()
	s.wg.Wait()
	slog.Info("aggregator service stopped")
}

// buildLocoUpdate собирает broadcast-пакет из снапшота и нормализованных метрик.
func buildLocoUpdate(snap spec.HealthSnapshot, metrics []domain.Metric) LocoUpdate {
	metricsMap := make(map[string]float64, len(metrics))
	for _, m := range metrics {
		metricsMap[m.Name] = m.Value
	}

	issues := make([]IssueWire, len(snap.Issues))
	for i, iss := range snap.Issues {
		issues[i] = IssueWire{
			Code:         iss.Code,
			Level:        iss.Level.String(),
			Target:       iss.Target,
			Message:      iss.Message,
			HealthWeight: iss.HealthWeight,
		}
	}

	return LocoUpdate{
		LocoID:   snap.LocoID,
		Ts:       snap.Ts,
		State:    snap.State.String(),
		Score:    snap.Score,
		Category: locoCategory(snap.State),
		Issues:   issues,
		Metrics:  metricsMap,
	}
}

func locoCategory(s spec.LocoState) string {
	switch s {
	case spec.StateEmergency:
		return "Critical"
	case spec.StateDegraded:
		return "Warning"
	case spec.StateMaintenance:
		return "Maintenance"
	default:
		return "Normal"
	}
}
