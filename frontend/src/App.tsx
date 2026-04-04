import { useState } from 'react'
import { AlertsPanel } from './components/AlertsPanel'
import './App.css'

function App() {
  const [selectedLoco, setSelectedLoco] = useState<string>('loco-001')

  return (
    <div className="app">
      <header className="app-header">
        <h1>Locomotive Digital Twin - Alerts Monitor</h1>
        <div className="header-controls">
          <label htmlFor="loco-select">Выберите локомотив:</label>
          <select
            id="loco-select"
            value={selectedLoco}
            onChange={(e) => setSelectedLoco(e.target.value)}
          >
            <option value="loco-001">Локо-001</option>
            <option value="loco-002">Локо-002</option>
            <option value="loco-003">Локо-003</option>
          </select>
        </div>
      </header>

      <main className="app-main">
        <AlertsPanel locomotiveId={selectedLoco} pollingInterval={3000} />
      </main>
    </div>
  )
}

export default App

