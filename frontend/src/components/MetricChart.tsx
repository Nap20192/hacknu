import { LineChart, Line, ResponsiveContainer, Tooltip, ReferenceLine } from 'recharts'
import type { MetricPoint } from '../types/telemetry'

interface Props {
  name: string
  display: string
  unit: string
  value: number | undefined
  history: MetricPoint[]
  warnAbove?: number
  warnBelow?: number
  critAbove?: number
  critBelow?: number
  status: 'normal' | 'warning' | 'critical'
}

function statusColor(s: string) {
  if (s === 'critical') return '#ef4444'
  if (s === 'warning') return '#f59e0b'
  return '#22c55e'
}

export function MetricChart({ display, unit, value, history, warnAbove, critAbove, status, name: _name }: Props) {
  const color = statusColor(status)
  const formatted = value !== undefined ? value.toFixed(1) : '—'

  return (
    <div className={`metric-card metric-${status}`}>
      <div className="metric-header">
        <span className="metric-name">{display}</span>
        <span className="metric-value" style={{ color }}>{formatted} <span className="metric-unit">{unit}</span></span>
      </div>
      <div className="metric-chart-wrap">
        <ResponsiveContainer width="100%" height={60}>
          <LineChart data={history.slice(-60)}>
            {warnAbove !== undefined && (
              <ReferenceLine y={warnAbove} stroke="#f59e0b" strokeDasharray="3 3" />
            )}
            {critAbove !== undefined && (
              <ReferenceLine y={critAbove} stroke="#ef4444" strokeDasharray="3 3" />
            )}
            <Line
              type="monotone"
              dataKey="value"
              stroke={color}
              strokeWidth={2}
              dot={false}
              isAnimationActive={false}
            />
            <Tooltip
              formatter={(v) => [`${(v as number).toFixed(2)} ${unit}`, display]}
              labelFormatter={() => ''}
              contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '6px', fontSize: '12px' }}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}
