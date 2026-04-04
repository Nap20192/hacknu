import type { LocoUpdate } from '../types/telemetry'

interface Props {
  update: LocoUpdate | null
}

function getColor(category: string) {
  switch (category) {
    case 'Critical': return '#ef4444'
    case 'Warning': return '#f59e0b'
    case 'Maintenance': return '#8b5cf6'
    default: return '#22c55e'
  }
}

function getLabel(category: string) {
  switch (category) {
    case 'Critical': return 'Критично'
    case 'Warning': return 'Внимание'
    case 'Maintenance': return 'Обслуживание'
    default: return 'Норма'
  }
}

function describeState(state: string) {
  switch (state) {
    case 'Emergency': return 'Требуется немедленная остановка'
    case 'Degraded': return 'Снижена производительность'
    case 'Maintenance': return 'Плановое обслуживание'
    default: return 'Все системы в норме'
  }
}

export function HealthIndex({ update }: Props) {
  const score = update?.score ?? 100
  const category = update?.category ?? 'Normal'
  const state = update?.state ?? 'Operational'
  const color = getColor(category)

  const radius = 70
  const circumference = 2 * Math.PI * radius
  const offset = circumference - (score / 100) * circumference

  return (
    <div className="health-index-card">
      <h2 className="panel-title">Индекс здоровья</h2>
      <div className="health-gauge-wrap">
        <svg width="180" height="180" viewBox="0 0 180 180">
          <circle cx="90" cy="90" r={radius} fill="none" stroke="#334155" strokeWidth="14" />
          <circle
            cx="90" cy="90" r={radius}
            fill="none"
            stroke={color}
            strokeWidth="14"
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            strokeLinecap="round"
            transform="rotate(-90 90 90)"
            style={{ transition: 'stroke-dashoffset 0.5s ease, stroke 0.5s ease' }}
          />
          <text x="90" y="85" textAnchor="middle" fontSize="36" fontWeight="bold" fill={color}>{score}</text>
          <text x="90" y="110" textAnchor="middle" fontSize="13" fill="#94a3b8">{getLabel(category)}</text>
        </svg>
      </div>
      <p className="health-state-desc">{describeState(state)}</p>
      {update && update.issues.length > 0 && (
        <div className="health-top-issues">
          <p className="issues-header">Ключевые проблемы ({update.issues.length}):</p>
          {update.issues.slice(0, 5).map((iss) => (
            <div key={iss.code} className={`issue-row issue-${iss.level.toLowerCase()}`}>
              <span className="issue-bullet" />
              <span className="issue-msg">{iss.message}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
