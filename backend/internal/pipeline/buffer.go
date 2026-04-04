package pipeline

import "github.com/Nap20192/hacknu/internal/domain"

const defaultEmaAlpha = float32(0.3)

// metricSamples накапливает сырые значения одной метрики за окно
// и хранит EMA-состояние между флашами.
type metricSamples struct {
	values []float64
	ema    float64
	hasEMA bool
}

// MetricBuffer накапливает сырые Metric-семплы для одного локомотива.
//
// При флаше выполняет два шага нормализации:
//  1. mean(window)   — среднее по окну, убирает внутриоконный шум
//  2. EMA(α, mean)   — сглаживает между окнами, убирает межоконные скачки
//
// EMA-стейт сохраняется между флашами для непрерывного сглаживания.
// Не потокобезопасен — вызывается только из одной горутины воркера.
type MetricBuffer struct {
	data  map[string]*metricSamples
	count int // суммарное число семплов по всем метрикам
	cap   int // порог для capacity-based flush
}

func newMetricBuffer(cap int) *MetricBuffer {
	return &MetricBuffer{
		data: make(map[string]*metricSamples),
		cap:  cap,
	}
}

// add добавляет семпл. Возвращает true когда буфер достиг cap.
func (b *MetricBuffer) add(m domain.Metric) (full bool) {
	s, ok := b.data[m.Name]
	if !ok {
		s = &metricSamples{}
		b.data[m.Name] = s
	}
	s.values = append(s.values, m.Value)
	b.count++
	return b.count >= b.cap
}

// flush вычисляет нормализованные значения, сбрасывает окно, возвращает результат.
// alphaFn возвращает EMA-alpha для конкретной метрики.
// Возвращает nil если нет накопленных данных.
func (b *MetricBuffer) flush(alphaFn func(string) float32) []domain.Metric {
	if b.count == 0 {
		return nil
	}
	out := make([]domain.Metric, 0, len(b.data))
	for name, s := range b.data {
		if len(s.values) == 0 {
			continue
		}
		// Шаг 1: среднее по окну
		mean := windowMean(s.values)

		// Шаг 2: EMA через окна
		alpha := float64(alphaFn(name))
		if alpha <= 0 || alpha > 1 {
			alpha = float64(defaultEmaAlpha)
		}
		if !s.hasEMA {
			s.ema = mean // cold start: seed EMA первым средним
			s.hasEMA = true
		} else {
			s.ema = alpha*mean + (1-alpha)*s.ema
		}

		out = append(out, domain.Metric{Name: name, Value: s.ema})
		s.values = s.values[:0] // сбрасываем окно, EMA сохраняем
	}
	b.count = 0
	return out
}

func windowMean(vals []float64) float64 {
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}
