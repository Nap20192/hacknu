import React from 'react';
import type { Alert } from '../types/telemetry';
import { format } from 'date-fns';

interface AlertsPanelProps {
  alerts: Alert[];
}

export const AlertsPanel: React.FC<AlertsPanelProps> = ({ alerts }) => {
  const criticalAlerts = alerts.filter(a => a.severity === 'critical');
  const warningAlerts = alerts.filter(a => a.severity === 'warning');
  const infoAlerts = alerts.filter(a => a.severity === 'info');

  const severityIcons = {
    critical: '🚨',
    warning: '⚠️',
    info: 'ℹ️',
  };

  const severityColors = {
    critical: '#ff5252',
    warning: '#ffd54f',
    info: '#64b5f6',
  };

  const AlertGroup: React.FC<{
    title: string;
    items: Alert[];
    severity: 'critical' | 'warning' | 'info';
  }> = ({ title, items, severity }) => (
    <div className="alert-group">
      <div className="group-header">
        <span className="group-icon">{severityIcons[severity]}</span>
        <h4>{title}</h4>
        <span className="group-count">{items.length}</span>
      </div>
      {items.length > 0 ? (
        <div className="alert-list">
          {items.map(alert => (
            <div key={alert.id} className="alert-item" style={{ borderLeftColor: severityColors[severity] }}>
              <div className="alert-time">{format(new Date(alert.timestamp), 'HH:mm:ss')}</div>
              <div className="alert-message">{alert.message}</div>
            </div>
          ))}
        </div>
      ) : (
        <div className="no-alerts">No {title.toLowerCase()}</div>
      )}
    </div>
  );

  return (
    <div className="alerts-panel">
      <h3>System Alerts</h3>
      <div className="alerts-content">
        <AlertGroup title="Critical" items={criticalAlerts} severity="critical" />
        <AlertGroup title="Warnings" items={warningAlerts} severity="warning" />
        <AlertGroup title="Info" items={infoAlerts} severity="info" />
      </div>

      <style>{`
        .alerts-panel {
          background: #1a1a2e;
          border: 1px solid rgba(255, 255, 255, 0.1);
          border-radius: 8px;
          padding: 16px;
        }

        .alerts-panel h3 {
          margin: 0 0 12px 0;
          font-size: 12px;
          text-transform: uppercase;
          color: rgba(255, 255, 255, 0.5);
          letter-spacing: 0.5px;
        }

        .alerts-content {
          display: flex;
          flex-direction: column;
          gap: 16px;
        }

        .alert-group {
          display: flex;
          flex-direction: column;
          gap: 8px;
        }

        .group-header {
          display: flex;
          align-items: center;
          gap: 8px;
          font-size: 11px;
          text-transform: uppercase;
          letter-spacing: 0.3px;
          color: rgba(255, 255, 255, 0.6);
        }

        .group-icon {
          font-size: 14px;
        }

        .group-header h4 {
          margin: 0;
          flex: 1;
        }

        .group-count {
          background: rgba(255, 255, 255, 0.1);
          padding: 2px 6px;
          border-radius: 3px;
          font-size: 10px;
        }

        .alert-list {
          display: flex;
          flex-direction: column;
          gap: 6px;
        }

        .alert-item {
          background: rgba(255, 255, 255, 0.02);
          border-left: 3px solid;
          border-radius: 4px;
          padding: 8px 12px;
          display: flex;
          gap: 12px;
        }

        .alert-time {
          font-size: 10px;
          color: rgba(255, 255, 255, 0.4);
          white-space: nowrap;
          min-width: 60px;
        }

        .alert-message {
          font-size: 12px;
          color: rgba(255, 255, 255, 0.7);
          flex: 1;
        }

        .no-alerts {
          font-size: 11px;
          color: rgba(255, 255, 255, 0.3);
          text-align: center;
          padding: 8px;
          font-style: italic;
        }
      `}</style>
    </div>
  );
};
