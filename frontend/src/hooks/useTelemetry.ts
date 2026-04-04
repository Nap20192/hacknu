import { useState, useCallback } from 'react'
import type { LocoUpdate, MetricHistory } from '../types/telemetry'
import { useWebSocket } from './useWebSocket'

const MAX_HISTORY = 300 // ~5 min at 1Hz

export function useTelemetry(locoId: string | null) {
  const [latest, setLatest] = useState<LocoUpdate | null>(null)
  const [history, setHistory] = useState<MetricHistory>({})
  const wsUrl = locoId ? '/ws/live' : null

  const onMessage = useCallback((raw: string) => {
    let update: LocoUpdate
    try {
      update = JSON.parse(raw) as LocoUpdate
    } catch {
      return
    }
    if (locoId && update.loco_id !== locoId) return

    setLatest(update)

    const ts = new Date(update.ts).getTime()
    setHistory((prev) => {
      const next = { ...prev }
      for (const [name, value] of Object.entries(update.metrics)) {
        const arr = prev[name] ? [...prev[name]] : []
        arr.push({ ts, value })
        if (arr.length > MAX_HISTORY) arr.shift()
        next[name] = arr
      }
      return next
    })
  }, [locoId])

  const { status } = useWebSocket(wsUrl, { onMessage })

  return { latest, history, wsStatus: status }
}
