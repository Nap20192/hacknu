import type { IssueWire } from '../types/telemetry'

interface Props {
  issues: IssueWire[]
}

const levelIcon: Record<string, string> = {
  Critical: '🔴',
  Warning: '🟡',
  Info: '🔵',
}

const recommendations: Record<string, string> = {
  CRIT_ABOVE_ENGINE_TEMP: 'Снизить скорость, проверить систему охлаждения',
  CRIT_BELOW_BRAKE_PRESSURE: 'Немедленная остановка!',
  WARN_BELOW_FUEL_LEVEL: 'Запланировать заправку на ближайшей станции',
  CRIT_BELOW_OIL_PRESSURE: 'Остановить двигатель, проверить уровень масла',
  CRIT_ABOVE_VOLTAGE: 'Проверить генераторы и регуляторы напряжения',
  WARN_ABOVE_TRACTION: 'Снизить тяговое усилие',
  WARN_ABOVE_AXLE_TEMP: 'Снизить скорость, проверить буксы',
}

export function IssuesPanel({ issues }: Props) {
  if (issues.length === 0) {
    return (
      <div className="issues-panel">
        <h2 className="panel-title">Активные проблемы</h2>
        <p className="issues-empty">Нет активных проблем</p>
      </div>
    )
  }

  const sorted = [...issues].sort((a, b) => {
    const order = { Critical: 0, Warning: 1, Info: 2 }
    return (order[a.level] ?? 3) - (order[b.level] ?? 3)
  })

  return (
    <div className="issues-panel">
      <h2 className="panel-title">Активные проблемы ({issues.length})</h2>
      <div className="issues-list">
        {sorted.map((iss) => (
          <div key={iss.code} className={`issue-item issue-item-${iss.level.toLowerCase()}`}>
            <div className="issue-item-header">
              <span>{levelIcon[iss.level]} {iss.message}</span>
              <span className="issue-weight">Вес: {(iss.health_weight * 100).toFixed(0)}%</span>
            </div>
            {recommendations[iss.code] && (
              <div className="issue-recommendation">💡 {recommendations[iss.code]}</div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
