import { useState } from 'react';
import { useAlertsPolling } from '../hooks/useAlertsPolling';
import { AlertsList } from './AlertsList';
import './AlertsPanel.css';

interface AlertsPanelProps {
  locomotiveId: string | null;
  pollingInterval?: number;
}

export const AlertsPanel = ({ locomotiveId, pollingInterval = 3000 }: AlertsPanelProps) => {
  const [acknowledging, setAcknowledging] = useState<Set<number>>(new Set());

  const { active, loading, error, lastUpdated, acknowledgeAlert } = useAlertsPolling(
    locomotiveId,
    {
      interval: pollingInterval,
      timeout: 5000,
      maxRetries: 5,
    }
  );

  const handleAcknowledge = async (alertId: number) => {
    setAcknowledging((prev) => new Set(prev).add(alertId));

    try {
      const success = await acknowledgeAlert(alertId);
      if (!success) {
        console.error('Failed to acknowledge alert');
      }
    } finally {
      setAcknowledging((prev) => {
        const next = new Set(prev);
        next.delete(alertId);
        return next;
      });
    }
  };

  const unacknowledgedCount = active.filter((a) => !a.acknowledged).length;
  const criticalCount = active.filter((a) => a.severity === 'critical' && !a.acknowledged).length;

  return (
    <div className="alerts-panel">
      <div className="alerts-stats">
        <div className="stat">
          <span className="stat-label">Всего алертов:</span>
          <span className="stat-value">{active.length}</span>
        </div>
        <div className="stat">
          <span className="stat-label">Не подтвержено:</span>
          <span className={`stat-value ${unacknowledgedCount > 0 ? 'warning' : ''}`}>
            {unacknowledgedCount}
          </span>
        </div>
        {criticalCount > 0 && (
          <div className="stat">
            <span className="stat-label stat-critical">Критичных:</span>
            <span className="stat-value stat-critical-value">{criticalCount}</span>
          </div>
        )}
        {lastUpdated && (
          <div className="stat stat-updated">
            <span className="stat-label">Обновлено:</span>
            <span className="stat-time">{formatTime(lastUpdated)}</span>
          </div>
        )}
      </div>

      <AlertsList
        alerts={active}
        loading={loading}
        error={error}
        onAcknowledge={handleAcknowledge}
        onAcknowledging={acknowledging}
      />
    </div>
  );
};

function formatTime(date: Date): string {
  return date.toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}
