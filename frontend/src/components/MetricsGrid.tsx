import type { LocoUpdate, MetricHistory } from '../types/telemetry'
import { MetricChart } from './MetricChart'

interface MetricMeta {
  name: string
  display: string
  unit: string
  warnAbove?: number
  warnBelow?: number
  critAbove?: number
  critBelow?: number
}

const METRICS: MetricMeta[] = [
  { name: 'speed_kmh', display: 'Скорость', unit: 'км/ч', warnAbove: 120, critAbove: 160 },
  { name: 'engine_temp_c', display: 'Температура двигателя', unit: '°C', warnAbove: 95, critAbove: 110 },
  { name: 'brake_pressure_bar', display: 'Давление тормозов', unit: 'бар', warnBelow: 4.5, critBelow: 3.0 },
  { name: 'fuel_level_pct', display: 'Уровень топлива', unit: '%', warnBelow: 20, critBelow: 10 },
  { name: 'oil_pressure_bar', display: 'Давление масла', unit: 'бар', warnBelow: 3.0, critBelow: 2.0 },
  { name: 'traction_amps', display: 'Ток тяги', unit: 'А', warnAbove: 600, critAbove: 750 },
  { name: 'voltage_v', display: 'Напряжение', unit: 'В', warnAbove: 135, critAbove: 150 },
  { name: 'axle_temp_c', display: 'Температура букс', unit: '°C', warnAbove: 60, critAbove: 80 },
]

function getMetricStatus(_name: string, value: number | undefined, meta: MetricMeta): 'normal' | 'warning' | 'critical' {
  if (value === undefined) return 'normal'
  if (meta.critAbove !== undefined && value >= meta.critAbove) return 'critical'
  if (meta.critBelow !== undefined && value <= meta.critBelow) return 'critical'
  if (meta.warnAbove !== undefined && value >= meta.warnAbove) return 'warning'
  if (meta.warnBelow !== undefined && value <= meta.warnBelow) return 'warning'
  return 'normal'
}

interface Props {
  update: LocoUpdate | null
  history: MetricHistory
}

export function MetricsGrid({ update, history }: Props) {
  return (
    <div className="metrics-grid">
      {METRICS.map((meta) => {
        const value = update?.metrics[meta.name]
        const hist = history[meta.name] ?? []
        const status = getMetricStatus(meta.name, value, meta)
        return (
          <MetricChart
            key={meta.name}
            name={meta.name}
            display={meta.display}
            unit={meta.unit}
            value={value}
            history={hist}
            warnAbove={meta.warnAbove}
            warnBelow={meta.warnBelow}
            critAbove={meta.critAbove}
            critBelow={meta.critBelow}
            status={status}
          />
        )
      })}
    </div>
  )
}
