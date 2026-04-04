import { useEffect, useState, useRef } from 'react';
import type { TelemetryData, Alert } from '../types/telemetry';

export function useTelemetrySimulator(updateInterval: number = 1000) {
  const [data, setData] = useState<TelemetryData | null>(null);
  const [history, setHistory] = useState<TelemetryData[]>([]);
  const [isConnected, setIsConnected] = useState(true);
  const dataRef = useRef<TelemetryData>({
    timestamp: Date.now(),
    speed: 50,
    fuel: { level: 75, consumption: 35 },
    pressure: { air: 100, hydraulic: 65, brake: 50 },
    temperature: { engine: 85, hydraulic: 55, ambient: 20 },
    electrical: { voltage: 25, current: 150, batteryHealth: 90 },
    alerts: [],
    gps: { lat: 51.5074, lon: -0.1278 },
  });

  useEffect(() => {
    const interval = setInterval(() => {
      const current = dataRef.current;
      const timestamp = Date.now();

      // Simulate realistic changes with trend
      const speedTrend = Math.sin(timestamp / 30000) * 30 + 60;
      const speed = Math.max(0, Math.min(160, speedTrend + (Math.random() - 0.5) * 10));

      const fuelConsumption = 25 + (speed / 160) * 40;
      const fuel = Math.max(5, current.fuel.level - fuelConsumption / 3600);

      const engineTemp = 80 + (speed / 160) * 20 + (Math.random() - 0.5) * 5;
      const airPressure = 100 + Math.sin(timestamp / 20000) * 15 + (Math.random() - 0.5) * 5;
      const brakePressure = current.pressure.brake + (Math.random() - 0.5) * 15;

      const voltage = 25 + (Math.random() - 0.5) * 2;
      const batteryHealth = Math.max(40, current.electrical.batteryHealth - 0.0001);
      const current_current = 100 + (speed / 160) * 100 + (Math.random() - 0.5) * 20;

      // Simulate occasional alerts
      const alerts: Alert[] = [];
      if (Math.random() > 0.98 && fuel < 30) {
        alerts.push({
          id: `alert_${timestamp}_fuel`,
          type: 'warning',
          message: 'Low fuel level',
          timestamp,
          severity: 'warning',
        });
      }
      if (Math.random() > 0.99) {
        alerts.push({
          id: `alert_${timestamp}_temp`,
          type: 'warning',
          message: 'High engine temperature',
          timestamp,
          severity: 'warning',
        });
      }
      if (Math.random() > 0.995) {
        alerts.push({
          id: `alert_${timestamp}_error`,
          type: 'error',
          message: 'System error detected',
          timestamp,
          severity: 'critical',
        });
      }

      const newData: TelemetryData = {
        timestamp,
        speed: Math.round(speed * 10) / 10,
        fuel: {
          level: Math.round(fuel * 10) / 10,
          consumption: Math.round(fuelConsumption * 10) / 10,
        },
        pressure: {
          air: Math.round(airPressure * 10) / 10,
          hydraulic: current.pressure.hydraulic + (Math.random() - 0.5) * 2,
          brake: Math.max(0, Math.min(120, brakePressure)),
        },
        temperature: {
          engine: Math.round(engineTemp * 10) / 10,
          hydraulic: current.temperature.hydraulic + (Math.random() - 0.5) * 1,
          ambient: 20 + (Math.random() - 0.5) * 3,
        },
        electrical: {
          voltage: Math.round(voltage * 100) / 100,
          current: Math.round(current_current),
          batteryHealth: Math.round(batteryHealth * 10) / 10,
        },
        alerts: [...current.alerts.slice(-5), ...alerts].slice(-10),
        gps: {
          lat: 51.5074 + Math.sin(timestamp / 50000) * 0.01,
          lon: -0.1278 + Math.cos(timestamp / 50000) * 0.01,
        },
      };

      dataRef.current = newData;
      setData(newData);
      setHistory(prev => [...prev, newData].slice(-900)); // Keep last 15 minutes at 1Hz

      // Simulate occasional disconnection
      if (Math.random() > 0.999) {
        setIsConnected(false);
        setTimeout(() => setIsConnected(true), 2000);
      }
    }, updateInterval);

    return () => clearInterval(interval);
  }, [updateInterval]);

  return { data, history, isConnected };
}
