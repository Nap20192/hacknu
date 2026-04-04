export interface IssueWire {
  code: string
  level: 'Info' | 'Warning' | 'Critical'
  target: string
  message: string
  health_weight: number
}

export interface LocoUpdate {
  loco_id: string
  ts: string
  state: 'Operational' | 'Degraded' | 'Emergency' | 'Maintenance'
  score: number
  category: 'Normal' | 'Warning' | 'Critical' | 'Maintenance'
  issues: IssueWire[]
  metrics: Record<string, number>
}

export interface Locomotive {
  id: string
  display_name: string
  loco_type: string
  registered_at: string
  last_seen_at?: string
  active: boolean
}

export interface MetricPoint {
  ts: number
  value: number
}

export interface MetricHistory {
  [metricName: string]: MetricPoint[]
}
