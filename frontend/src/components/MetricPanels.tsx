import React from 'react';
import type { TelemetryData } from '../types/telemetry';

interface MetricPanelProps {
  title: string;
  metrics: Array<{
    label: string;
    value: number | string;
    unit?: string;
    warning?: { min?: number; max?: number };
    critical?: { min?: number; max?: number };
    icon?: string;
  }>;
}

export const MetricPanel: React.FC<MetricPanelProps> = ({ title, metrics }) => {
  const getMetricColor = (
    value: number,
    warning?: { min?: number; max?: number },
    critical?: { min?: number; max?: number }
  ): string => {
    if (critical) {
      if (
        (critical.min !== undefined && value < critical.min) ||
        (critical.max !== undefined && value > critical.max)
      ) {
        return '#ff5252';
      }
    }
    if (warning) {
      if (
        (warning.min !== undefined && value < warning.min) ||
        (warning.max !== undefined && value > warning.max)
      ) {
        return '#ffd54f';
      }
    }
    return '#5dd6a8';
  };

  return (
    <div className="metric-panel">
      <h3>{title}</h3>
      <div className="metrics-grid">
        {metrics.map((metric, i) => {
          const numValue = typeof metric.value === 'number' ? metric.value : 0;
          const color = typeof metric.value === 'number'
            ? getMetricColor(numValue, metric.warning, metric.critical)
            : '#5dd6a8';

          return (
            <div key={i} className="metric-item">
              {metric.icon && <span className="metric-icon">{metric.icon}</span>}
              <div className="metric-label">{metric.label}</div>
              <div className="metric-value" style={{ color }}>
                {metric.value}
                {metric.unit && <span className="metric-unit">{metric.unit}</span>}
              </div>
            </div>
          );
        })}
      </div>

      <style>{`
        .metric-panel {
          background: #1a1a2e;
          border: 1px solid rgba(255, 255, 255, 0.1);
          border-radius: 8px;
          padding: 16px;
        }

        .metric-panel h3 {
          margin: 0 0 12px 0;
          font-size: 12px;
          text-transform: uppercase;
          color: rgba(255, 255, 255, 0.5);
          letter-spacing: 0.5px;
        }

        .metrics-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
          gap: 12px;
        }

        .metric-item {
          display: flex;
          flex-direction: column;
          gap: 4px;
          padding: 8px;
          background: rgba(255, 255, 255, 0.02);
          border-radius: 6px;
          border: 1px solid rgba(255, 255, 255, 0.05);
        }

        .metric-icon {
          font-size: 20px;
        }

        .metric-label {
          font-size: 11px;
          color: rgba(255, 255, 255, 0.4);
          text-transform: uppercase;
          letter-spacing: 0.3px;
        }

        .metric-value {
          font-size: 18px;
          font-weight: bold;
          color: #5dd6a8;
          transition: color 0.2s;
        }

        .metric-unit {
          font-size: 12px;
          margin-left: 4px;
          opacity: 0.7;
        }
      `}</style>
    </div>
  );
};

export const SpeedPanel: React.FC<{ data: TelemetryData }> = ({ data }) => (
  <MetricPanel
    title="Speed"
    metrics={[
      {
        icon: '🚄',
        label: 'Current',
        value: data.speed.toFixed(1),
        unit: 'km/h',
        warning: { min: 0, max: 150 },
        critical: { max: 160 },
      },
    ]}
  />
);

export const FuelPanel: React.FC<{ data: TelemetryData }> = ({ data }) => (
  <MetricPanel
    title="Fuel & Energy"
    metrics={[
      {
        icon: '⛽',
        label: 'Level',
        value: data.fuel.level.toFixed(1),
        unit: '%',
        warning: { min: 20 },
        critical: { min: 10 },
      },
      {
        icon: '💨',
        label: 'Consumption',
        value: data.fuel.consumption.toFixed(1),
        unit: 'L/h',
        warning: { max: 60 },
        critical: { max: 80 },
      },
    ]}
  />
);

export const PressurePanel: React.FC<{ data: TelemetryData }> = ({ data }) => (
  <MetricPanel
    title="Pressures & Temperatures"
    metrics={[
      {
        icon: '🌡️',
        label: 'Engine Temp',
        value: data.temperature.engine.toFixed(1),
        unit: '°C',
        warning: { min: 70, max: 105 },
        critical: { max: 120 },
      },
      {
        icon: '💧',
        label: 'Hydraulic Temp',
        value: data.temperature.hydraulic.toFixed(1),
        unit: '°C',
        warning: { max: 85 },
        critical: { max: 100 },
      },
      {
        icon: '📊',
        label: 'Air Pressure',
        value: data.pressure.air.toFixed(1),
        unit: 'kPa',
        warning: { min: 80, max: 120 },
        critical: { min: 70, max: 130 },
      },
      {
        icon: '🛑',
        label: 'Brake Press',
        value: data.pressure.brake.toFixed(1),
        unit: 'kPa',
        warning: { min: 20, max: 110 },
        critical: { min: 10, max: 120 },
      },
    ]}
  />
);

export const ElectricPanel: React.FC<{ data: TelemetryData }> = ({ data }) => (
  <MetricPanel
    title="Electrical System"
    metrics={[
      {
        icon: '⚡',
        label: 'Voltage',
        value: data.electrical.voltage.toFixed(2),
        unit: 'V',
        warning: { min: 23, max: 27 },
        critical: { min: 20, max: 30 },
      },
      {
        icon: '🔌',
        label: 'Current',
        value: data.electrical.current.toFixed(0),
        unit: 'A',
        warning: { max: 300 },
        critical: { max: 350 },
      },
      {
        icon: '🔋',
        label: 'Battery',
        value: data.electrical.batteryHealth.toFixed(1),
        unit: '%',
        warning: { min: 60 },
        critical: { min: 40 },
      },
    ]}
  />
);
