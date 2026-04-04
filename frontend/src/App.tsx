import { useState, useEffect } from 'react'
import { useTelemetrySimulator } from './hooks/useTelemetrySimulator'
import { Dashboard } from './components/Dashboard'
import './App.css'

function App() {
  const [isDarkMode, setIsDarkMode] = useState(true)
  const { data, history, isConnected } = useTelemetrySimulator(1000)

  useEffect(() => {
    // Check for saved theme preference
    const savedTheme = localStorage.getItem('theme')
    if (savedTheme === 'light') {
      setIsDarkMode(false)
    }
  }, [])

  const handleThemeToggle = () => {
    const newMode = !isDarkMode
    setIsDarkMode(newMode)
    localStorage.setItem('theme', newMode ? 'dark' : 'light')
  }

  return (
    <div className="app" data-theme={isDarkMode ? 'dark' : 'light'}>
      <Dashboard
        data={data}
        history={history}
        isConnected={isConnected}
        isDark={isDarkMode}
        onThemeToggle={handleThemeToggle}
      />
    </div>
  )
}

export default App
