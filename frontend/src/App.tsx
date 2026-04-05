import { useState } from 'react'
import { useLocomotives } from './hooks/useLocomotives'
import { useTelemetry } from './hooks/useTelemetry'
import { useTheme } from './hooks/useTheme'
import { HealthIndex } from './components/HealthIndex'
import { MetricsGrid } from './components/MetricsGrid'
import { IssuesPanel } from './components/IssuesPanel'
import { AlertsPanel } from './components/AlertsPanel'
import { ConnectionBadge } from './components/ConnectionBadge'
import { ThemeToggle } from './components/ThemeToggle'
import './App.css'

export default function App() {
  const { locomotives, loading } = useLocomotives()
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const { theme, toggleTheme } = useTheme()

  const locoId = selectedId ?? (locomotives.length > 0 ? locomotives[0].id : null)
  const { latest, history, wsStatus } = useTelemetry(locoId)

  const selectedLoco = locomotives.find((l) => l.id === locoId)

  return (
    <div className={`app ${theme}`}>
      <header className="app-header">
        <div className="header-left">
          <span className="app-logo">🚂</span>
          <h1 className="app-title">Цифровой двойник локомотива</h1>
        </div>
        <div className="header-center">
          {loading ? (
            <span className="loco-loading">Загрузка...</span>
          ) : (
            <select
              className="loco-select"
              value={locoId ?? ''}
              onChange={(e) => setSelectedId(e.target.value || null)}
            >
              {locomotives.map((l) => (
                <option key={l.id} value={l.id}>{l.display_name}</option>
              ))}
            </select>
          )}
          {selectedLoco && <span className="loco-type">{selectedLoco.loco_type}</span>}
        </div>
        <div className="header-right">
          <ConnectionBadge status={wsStatus} />
          {latest && (
            <span className="last-update">
              {new Date(latest.ts).toLocaleTimeString('ru-RU')}
            </span>
          )}
          <ThemeToggle theme={theme} onToggle={toggleTheme} />
        </div>
      </header>

      <main className="app-main">
        <div className="left-column">
          <HealthIndex update={latest} />
          <IssuesPanel issues={latest?.issues ?? []} />
          {locoId && <AlertsPanel locomotiveId={locoId} pollingInterval={5000} />}
        </div>
        <div className="right-column">
          <MetricsGrid update={latest} history={history} />
        </div>
      </main>
    </div>
  )
}
