export interface TelemetryData {
  timestamp: number;
  speed: number; // км/ч
  fuel: {
    level: number; // %
    consumption: number; // л/ч
  };
  pressure: {
    air: number; // кПа
    hydraulic: number; // кПа
    brake: number; // кПа
  };
  temperature: {
    engine: number; // °C
    hydraulic: number; // °C
    ambient: number; // °C
  };
  electrical: {
    voltage: number; // V
    current: number; // A
    batteryHealth: number; // %
  };
  alerts: Alert[];
  gps?: {
    lat: number;
    lon: number;
  };
}

export interface Alert {
  id: string;
  type: 'error' | 'warning' | 'info';
  message: string;
  timestamp: number;
  severity: 'critical' | 'warning' | 'info';
}

export interface HealthIndexBreakdown {
  score: number; // 0-100
  category: 'normal' | 'warning' | 'critical';
  factors: {
    name: string;
    weight: number;
    value: number;
    impact: number;
  }[];
  recommendations: string[];
}

export interface HistoryEntry extends TelemetryData {
  healthIndex: number;
}
