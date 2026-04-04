import type { Alert } from '../types/alerts';
import './AlertsList.css';

interface AlertsListProps {
  alerts: Alert[];
  loading: boolean;
  error: string | null;
  onAcknowledge: (alertId: number) => void;
  onAcknowledging?: Set<number>;
}

export const AlertsList = ({
  alerts,
  loading,
  error,
  onAcknowledge,
  onAcknowledging = new Set(),
}: AlertsListProps) => {
  if (loading && alerts.length === 0) {
    return (
      <div className="alerts-loading">
        <span>Загрузка алертов...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="alerts-error">
        <span>Ошибка: {error}</span>
      </div>
    );
  }

  if (alerts.length === 0) {
    return (
      <div className="alerts-empty">
        <span>Нет активных алертов</span>
      </div>
    );
  }

  return (
    <div className="alerts-container">
      <div className="alerts-header">
        <h2>Активные алерты ({alerts.length})</h2>
      </div>

      <div className="alerts-list">
        {alerts.map((alert) => (
          <AlertCard
            key={alert.id}
            alert={alert}
            isAcknowledging={onAcknowledging.has(alert.id)}
            onAcknowledge={onAcknowledge}
          />
        ))}
      </div>
    </div>
  );
};

interface AlertCardProps {
  alert: Alert;
  isAcknowledging: boolean;
  onAcknowledge: (alertId: number) => void;
}

const AlertCard = ({ alert, isAcknowledging, onAcknowledge }: AlertCardProps) => {
  const severityClass = `severity-${alert.severity}`;
  const triggeredDate = new Date(alert.triggered_at);
  const timeAgo = getTimeAgo(triggeredDate);

  return (
    <div className={`alert-card ${severityClass}`}>
      <div className="alert-header">
        <div className="alert-title">
          <span className={`severity-badge ${alert.severity}`}>{alert.severity}</span>
          <h3>{alert.code}</h3>
        </div>
        <span className="alert-time">{timeAgo}</span>
      </div>

      <div className="alert-body">
        <p className="alert-message">{alert.message}</p>

        {alert.metric_name && (
          <div className="alert-metric">
            <span className="metric-label">{alert.metric_name}</span>
            <span className="metric-value">
              {alert.metric_value?.toFixed(2)}
              {alert.threshold && ` / ${alert.threshold.toFixed(2)}`}
            </span>
          </div>
        )}

        {alert.recommendation && (
          <div className="alert-recommendation">
            <strong>Рекомендация:</strong> {alert.recommendation}
          </div>
        )}
      </div>

      <div className="alert-footer">
        <span className="alert-loco">Локомотив: {alert.locomotive_id}</span>
        <button
          className="btn-acknowledge"
          onClick={() => onAcknowledge(alert.id)}
          disabled={alert.acknowledged || isAcknowledging}
          title={alert.acknowledged ? 'Алерт уже подтвержден' : 'Подтвердить алерт'}
        >
          {isAcknowledging ? 'Подтверждение...' : alert.acknowledged ? '✓ Подтвержден' : 'Подтвердить'}
        </button>
      </div>
    </div>
  );
};

function getTimeAgo(date: Date): string {
  const now = new Date();
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (seconds < 60) return `${seconds}с назад`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}м назад`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}ч назад`;
  return `${Math.floor(seconds / 86400)}д назад`;
}
