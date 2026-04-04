import { useState } from 'react';
import { useAlertsPolling } from '../hooks/useAlertsPolling';
import type { Alert } from '../types/alerts';

interface AlertsPanelProps {
  locomotiveId: string | null;
  pollingInterval?: number;
}

export const AlertsPanel = ({ locomotiveId, pollingInterval = 5000 }: AlertsPanelProps) => {
  const [acknowledging, setAcknowledging] = useState<Set<number>>(new Set());
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

  const unacked = active.filter((a) => !a.acknowledged).length;
  const critical = active.filter((a) => a.severity === 'critical' && !a.acknowledged).length;

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

      {loading && active.length === 0 && <div className="alerts-loading">Загрузка...</div>}
      {error && <div className="alerts-error">Ошибка: {error}</div>}
      {!loading && !error && active.length === 0 && (
        <div className="alerts-empty">Нет активных алертов</div>
      )}
      {active.length > 0 && (
        <div className="alerts-list">
          {active.map((alert) => (
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
