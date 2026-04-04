import React, { useState } from 'react';
import { HealthIndex } from './HealthIndex';
import { SpeedPanel, FuelPanel, PressurePanel, ElectricPanel } from './MetricPanels';
import { TrendsChart } from './TrendsChart';
import { AlertsPanel } from './AlertsPanel';
import { RouteMap } from './RouteMap';
import type { TelemetryData } from '../types/telemetry';
import { calculateHealthIndex } from '../utils/healthIndex';

interface DashboardProps {
  data: TelemetryData | null;
  history: TelemetryData[];
  isConnected: boolean;
  isDark: boolean;
  onThemeToggle: () => void;
}

export const Dashboard: React.FC<DashboardProps> = ({
  data,
  history,
  isConnected,
  isDark,
  onThemeToggle,
}) => {
  const [timeWindow, setTimeWindow] = useState(10);
  const [showHistory, setShowHistory] = useState(false);

  if (!data) {
    return (
      <div className="dashboard loading">
        <div className="loading-content">
          <div className="spinner"></div>
          <p>Initializing telemetry...</p>
        </div>
      </div>
    );
  }

  const healthIndexData = calculateHealthIndex(data);

  const handleExport = () => {
    const csv =
      'timestamp,speed,fuel_level,engine_temp,air_pressure,voltage\n' +
      history
        .map(
          h =>
            `${new Date(h.timestamp).toISOString()},${h.speed},${h.fuel.level},${h.temperature.engine},${h.pressure.air},${h.electrical.voltage}`
        )
        .join('\n');

    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `telemetry_export_${Date.now()}.csv`;
    a.click();
  };

  return (
    <div className={`dashboard ${isDark ? 'dark' : 'light'}`}>
      {/* Header */}
      <header className="dashboard-header">
        <div className="header-left">
          <h1>🚂 Locomotive Digital Twin</h1>
          <div className="status-badge">
            <span className={`connection ${isConnected ? 'connected' : 'disconnected'}`}></span>
            <span>{isConnected ? 'LIVE' : 'DISCONNECTED'}</span>
          </div>
        </div>
        <div className="header-right">
          <button
            className="control-btn"
            onClick={() => setShowHistory(!showHistory)}
            title="Toggle history replay"
          >
            ⏱️ {showHistory ? 'Live' : 'History'}
          </button>
          <select
            className="control-select"
            value={timeWindow}
            onChange={e => setTimeWindow(Number(e.target.value))}
          >
            <option value={5}>5 min</option>
            <option value={10}>10 min</option>
            <option value={15}>15 min</option>
          </select>
          <button className="control-btn" onClick={handleExport} title="Export data">
            📥 Export
          </button>
          <button
            className="control-btn"
            onClick={onThemeToggle}
            title="Toggle theme"
          >
            {isDark ? '☀️' : '🌙'}
          </button>
        </div>
      </header>

      {/* Main Content */}
      <main className="dashboard-main">
        {/* Left Column: Health Index */}
        <aside className="sidebar-left">
          <HealthIndex data={healthIndexData} isConnected={isConnected} />
        </aside>

        {/* Center Column: Metrics and Charts */}
        <section className="main-content">
          {/* Metrics Row */}
          <div className="metrics-section">
            <SpeedPanel data={data} />
            <FuelPanel data={data} />
            <PressurePanel data={data} />
            <ElectricPanel data={data} />
          </div>

          {/* Charts Row */}
          <div className="charts-section">
            <TrendsChart history={showHistory ? history : [data]} timeWindow={timeWindow} />
          </div>

          {/* Alerts and Map Row */}
          <div className="lower-section">
            <AlertsPanel alerts={data.alerts} />
            <RouteMap data={data} history={history} />
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="dashboard-footer">
        <div className="footer-info">
          <span>Last update: {new Date(data.timestamp).toLocaleTimeString()}</span>
          <span>•</span>
          <span>Data points: {history.length}</span>
          <span>•</span>
          <span>FPS: ~{Math.round(1000 / (data.timestamp - (history[Math.max(0, history.length - 2)]?.timestamp || data.timestamp)))}</span>
        </div>
      </footer>

      <style>{`
        .dashboard {
          display: flex;
          flex-direction: column;
          height: 100vh;
          overflow: hidden;
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
          background: #0f0f23;
          color: rgba(255, 255, 255, 0.9);
        }

        .dashboard.light {
          background: #f5f5f5;
          color: #1a1a1a;
        }

        /* Header */
        .dashboard-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 16px 24px;
          border-bottom: 2px solid rgba(148, 225, 213, 0.2);
          background: rgba(0, 0, 0, 0.3);
          gap: 16px;
          flex-wrap: wrap;
        }

        .header-left {
          display: flex;
          align-items: center;
          gap: 16px;
        }

        .dashboard-header h1 {
          margin: 0;
          font-size: 24px;
          font-weight: 600;
          letter-spacing: 0.5px;
        }

        .status-badge {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 6px 12px;
          background: rgba(93, 214, 168, 0.1);
          border: 1px solid rgba(93, 214, 168, 0.3);
          border-radius: 4px;
          font-size: 12px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
          color: #5dd6a8;
        }

        .connection {
          width: 8px;
          height: 8px;
          border-radius: 50%;
          background: #5dd6a8;
        }

        .connection.disconnected {
          background: #ff5252;
          animation: pulse 1s infinite;
        }

        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }

        .header-right {
          display: flex;
          align-items: center;
          gap: 8px;
          flex-wrap: wrap;
        }

        .control-btn,
        .control-select {
          padding: 8px 12px;
          background: rgba(255, 255, 255, 0.05);
          border: 1px solid rgba(255, 255, 255, 0.1);
          border-radius: 4px;
          color: rgba(255, 255, 255, 0.8);
          cursor: pointer;
          font-size: 12px;
          text-transform: uppercase;
          letter-spacing: 0.3px;
          transition: all 0.2s;
        }

        .control-btn:hover,
        .control-select:hover {
          background: rgba(255, 255, 255, 0.1);
          border-color: rgba(255, 255, 255, 0.2);
        }

        /* Main Layout */
        .dashboard-main {
          display: flex;
          flex: 1;
          overflow: hidden;
          gap: 16px;
          padding: 16px 24px;
        }

        .sidebar-left {
          flex: 0 0 380px;
          overflow-y: auto;
        }

        .main-content {
          flex: 1;
          display: flex;
          flex-direction: column;
          gap: 16px;
          overflow-y: auto;
        }

        .metrics-section {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
          gap: 12px;
        }

        .charts-section {
          flex: 1;
          min-height: 400px;
          overflow-y: auto;
        }

        .lower-section {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 16px;
        }

        /* Footer */
        .dashboard-footer {
          padding: 12px 24px;
          border-top: 1px solid rgba(255, 255, 255, 0.1);
          background: rgba(0, 0, 0, 0.2);
          font-size: 11px;
          color: rgba(255, 255, 255, 0.4);
        }

        .footer-info {
          display: flex;
          gap: 12px;
          justify-content: center;
        }

        /* Loading State */
        .dashboard.loading {
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .loading-content {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 16px;
        }

        .spinner {
          width: 40px;
          height: 40px;
          border: 3px solid rgba(255, 255, 255, 0.1);
          border-top-color: #5dd6a8;
          border-radius: 50%;
          animation: spin 1s linear infinite;
        }

        @keyframes spin {
          to { transform: rotate(360deg); }
        }

        /* Responsive */
        @media (max-width: 1400px) {
          .sidebar-left {
            flex: 0 0 320px;
          }
          
          .lower-section {
            grid-template-columns: 1fr;
          }
        }

        @media (max-width: 1024px) {
          .dashboard-main {
            flex-direction: column;
          }

          .sidebar-left {
            flex: 0 0 auto;
            max-height: 400px;
          }

          .metrics-section {
            grid-template-columns: repeat(2, 1fr);
          }
        }

        @media (max-width: 768px) {
          .dashboard-header {
            flex-direction: column;
            align-items: stretch;
          }

          .header-right {
            justify-content: space-between;
          }

          .metrics-section {
            grid-template-columns: 1fr;
          }

          .dashboard-main {
            padding: 8px 12px;
          }
        }

        /* Scrollbar styling */
        ::-webkit-scrollbar {
          width: 8px;
          height: 8px;
        }

        ::-webkit-scrollbar-track {
          background: transparent;
        }

        ::-webkit-scrollbar-thumb {
          background: rgba(255, 255, 255, 0.1);
          border-radius: 4px;
        }

        ::-webkit-scrollbar-thumb:hover {
          background: rgba(255, 255, 255, 0.2);
        }
      `}</style>
    </div>
  );
};
