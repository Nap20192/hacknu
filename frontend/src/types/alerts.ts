// API Response types
export interface Alert {
  id: number;
  locomotive_id: string;
  triggered_at: string;
  resolved_at: string | null;
  severity: 'warning' | 'critical';
  code: string;
  metric_name: string | null;
  metric_value: number | null;
  threshold: number | null;
  message: string;
  recommendation: string;
  acknowledged: boolean;
}

export interface PagedResponse<T> {
  success: boolean;
  data: T[];
  total: number;
}

export interface Response<T> {
  success: boolean;
  data?: T;
  error?: string;
}

// Frontend-specific types
export interface AlertsState {
  active: Alert[];
  loading: boolean;
  error: string | null;
  lastUpdated: Date | null;
}

export interface LongPollingConfig {
  interval?: number; // polling interval in ms (default: 3000)
  timeout?: number; // request timeout in ms (default: 5000)
  maxRetries?: number; // max consecutive failures before stopping (default: 5)
}
