import React, { useMemo } from 'react';
import {
  LineChart,
  Line,
  Area,
  AreaChart,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import type { TelemetryData } from '../types/telemetry';
import { format } from 'date-fns';

interface TrendsChartProps {
  history: TelemetryData[];
  timeWindow?: number; // minutes
}

export const TrendsChart: React.FC<TrendsChartProps> = ({ history, timeWindow = 10 }) => {
  const chartData = useMemo(() => {
    if (history.length === 0) return [];
    const now = Date.now();
    const cutoff = now - timeWindow * 60 * 1000;

    return history
      .filter(d => d.timestamp >= cutoff)
      .map(d => ({
        time: format(new Date(d.timestamp), 'HH:mm:ss'),
        speed: d.speed,
        fuel: d.fuel.level,
        engineTemp: d.temperature.engine,
        airPressure: d.pressure.air,
        voltage: d.electrical.voltage,
      }));
  }, [history, timeWindow]);

  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload[0]) {
      return (
        <div className="tooltip-content">
          <p>{payload[0].payload.time}</p>
          {payload.map((entry: any, i: number) => (
            <p key={i} style={{ color: entry.color }}>
              {entry.name}: {entry.value.toFixed(2)}
            </p>
          ))}
        </div>
      );
    }
    return null;
  };

  return (
    <div className="trends-container">
      <h3>Telemetry Trends ({timeWindow} min)</h3>

      <div className="chart-grid">
        <div className="chart-item">
          <h4>Speed Profile</h4>
          <ResponsiveContainer width="100%" height={200}>
            <AreaChart data={chartData}>
              <defs>
                <linearGradient id="speedGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#94e1d5" stopOpacity={0.8} />
                  <stop offset="95%" stopColor="#94e1d5" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
              <XAxis
                dataKey="time"
                tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 11 }}
                interval={Math.max(0, Math.floor(chartData.length / 6))}
              />
              <YAxis tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 11 }} />
              <Tooltip content={<CustomTooltip />} />
              <Area
                type="monotone"
                dataKey="speed"
                stroke="#94e1d5"
                fillOpacity={1}
                fill="url(#speedGradient)"
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        <div className="chart-item">
          <h4>Fuel Level</h4>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
              <XAxis
                dataKey="time"
                tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 11 }}
                interval={Math.max(0, Math.floor(chartData.length / 6))}
              />
              <YAxis tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 11 }} domain={[0, 100]} />
              <Tooltip content={<CustomTooltip />} />
              <Line
                type="monotone"
                dataKey="fuel"
                stroke="#ffd54f"
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        <div className="chart-item">
          <h4>Temperatures</h4>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
              <XAxis
                dataKey="time"
                tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 11 }}
                interval={Math.max(0, Math.floor(chartData.length / 6))}
              />
              <YAxis tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 11 }} />
              <Tooltip content={<CustomTooltip />} />
              <Legend wrapperStyle={{ color: 'rgba(255,255,255,0.7)' }} />
              <Line
                type="monotone"
                dataKey="engineTemp"
                stroke="#ff6b6b"
                strokeWidth={2}
                dot={false}
                name="Engine"
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        <div className="chart-item">
          <h4>Electrical System</h4>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
              <XAxis
                dataKey="time"
                tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 11 }}
                interval={Math.max(0, Math.floor(chartData.length / 6))}
              />
              <YAxis tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 11 }} />
              <Tooltip content={<CustomTooltip />} />
              <Line
                type="monotone"
                dataKey="voltage"
                stroke="#a8e6cf"
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>

      <style>{`
        .trends-container {
          background: #1a1a2e;
          border: 1px solid rgba(255, 255, 255, 0.1);
          border-radius: 8px;
          padding: 16px;
        }

        .trends-container h3 {
          margin: 0 0 16px 0;
          font-size: 12px;
          text-transform: uppercase;
          color: rgba(255, 255, 255, 0.5);
          letter-spacing: 0.5px;
        }

        .chart-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
          gap: 16px;
        }

        .chart-item {
          background: rgba(255, 255, 255, 0.02);
          border: 1px solid rgba(255, 255, 255, 0.05);
          border-radius: 6px;
          padding: 12px;
        }

        .chart-item h4 {
          margin: 0 0 8px 0;
          font-size: 11px;
          text-transform: uppercase;
          color: rgba(255, 255, 255, 0.6);
          letter-spacing: 0.3px;
        }

        .tooltip-content {
          background: rgba(0, 0, 0, 0.8);
          border: 1px solid rgba(255, 255, 255, 0.1);
          border-radius: 4px;
          padding: 8px 12px;
          font-size: 11px;
        }

        .tooltip-content p {
          margin: 2px 0;
          color: rgba(255, 255, 255, 0.8);
        }
      `}</style>
    </div>
  );
};
