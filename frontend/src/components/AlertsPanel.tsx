import { useState } from 'react';
import { useAlertsPolling } from '../hooks/useAlertsPolling';
import type { Alert } from '../types/alerts';
import './AlertsPanel.css';

interface AlertsPanelProps {
  locomotiveId: string | null;
  pollingInterval?: number;
}

type AlertFilter = 'all' | 'warning' | 'critical';

export const AlertsPanel = ({ locomotiveId, pollingInterval = 5000 }: AlertsPanelProps) => {
  const [acknowledging, setAcknowledging] = useState<Set<number>>(new Set());
  const [filter, setFilter] = useState<AlertFilter>('all');
  const { active, loading, error, lastUpdated, acknowledgeAlert } = useAlertsPolling(
    locomotiveId,
    { interval: pollingInterval, timeout: 5000, maxRetries: 5 }
  );

  const handleAcknowledge = async (alertId: number) => {
    setAcknowledging((prev) => new Set(prev).add(alertId));
    try {
      await acknowledgeAlert(alertId);
    } finally {
      setAcknowledging((prev) => {
        const next = new Set(prev);
        next.delete(alertId);
        return next;
      });
    }
  };

  const filteredAlerts = active.filter((a) => {
    if (filter === 'all') return true;
    return a.severity === filter;
  });

  const unacked = active.filter((a) => !a.acknowledged).length;
  const critical = active.filter((a) => a.severity === 'critical' && !a.acknowledged).length;
  const warnings = active.filter((a) => a.severity === 'warning' && !a.acknowledged).length;

  return (
    <div className="alerts-panel">
      <h2 className="panel-title">Системные алерты</h2>
      <div className="alerts-stats">
        <div className="stat">
          <span className="stat-label">Всего:</span>
          <span className="stat-value">{active.length}</span>
        </div>
        {unacked > 0 && (
          <div className="stat">
            <span className="stat-label">Не подтв.:</span>
            <span className="stat-value warning">{unacked}</span>
          </div>
        )}
        {critical > 0 && (
          <div className="stat">
            <span className="stat-label">Критичных:</span>
            <span className="stat-critical-value">{critical}</span>
          </div>
        )}
        {lastUpdated && (
          <div className="stat stat-updated">
            <span className="stat-time">{lastUpdated.toLocaleTimeString('ru-RU')}</span>
          </div>
        )}
      </div>

      <div className="alerts-filters">
        <button
          className={`filter-btn ${filter === 'all' ? 'active' : ''}`}
          onClick={() => setFilter('all')}
        >
          Все
          <span className="filter-badge">{active.length}</span>
        </button>
        <button
          className={`filter-btn filter-btn--warning ${filter === 'warning' ? 'active' : ''}`}
          onClick={() => setFilter('warning')}
        >
          <span className="filter-dot filter-dot--warning" />
          Warn
          <span className="filter-badge">{warnings}</span>
        </button>
        <button
          className={`filter-btn filter-btn--critical ${filter === 'critical' ? 'active' : ''}`}
          onClick={() => setFilter('critical')}
        >
          <span className="filter-dot filter-dot--critical" />
          Crit
          <span className="filter-badge">{critical}</span>
        </button>
      </div>

      {loading && active.length === 0 && <div className="alerts-loading">Загрузка...</div>}
      {error && <div className="alerts-error">Ошибка: {error}</div>}
      {!loading && !error && filteredAlerts.length === 0 && (
        <div className="alerts-empty">
          {active.length === 0 ? 'Нет активных алертов' : 'Нет алертов в этой категории'}
        </div>
      )}
      {filteredAlerts.length > 0 && (
        <div className="alerts-list">
          {filteredAlerts.map((alert) => (
            <AlertRow
              key={alert.id}
              alert={alert}
              isAcknowledging={acknowledging.has(alert.id)}
              onAcknowledge={handleAcknowledge}
            />
          ))}
        </div>
      )}
    </div>
  );
};

function AlertRow({
  alert,
  isAcknowledging,
  onAcknowledge,
}: {
  alert: Alert;
  isAcknowledging: boolean;
  onAcknowledge: (id: number) => void;
}) {
  return (
    <div className={`alert-item severity-${alert.severity} ${alert.acknowledged ? 'alert-acknowledged' : ''}`}>
      <div className="alert-message">
        <div className="alert-code">{alert.code}</div>
        <div className="alert-text">{alert.message}</div>
        {alert.recommendation && (
          <div className="alert-recommend">💡 {alert.recommendation}</div>
        )}
      </div>
      {!alert.acknowledged && (
        <button
          className="alert-ack-btn"
          onClick={() => onAcknowledge(alert.id)}
          disabled={isAcknowledging}
        >
          {isAcknowledging ? '...' : 'OK'}
        </button>
      )}
    </div>
  );
}
