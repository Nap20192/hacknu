import { useState, useEffect } from 'react'
import type { Locomotive } from '../types/telemetry'

interface PagedResponse<T> {
  success: boolean
  data: T[]
  total: number
}

export function useLocomotives() {
  const [locomotives, setLocomotives] = useState<Locomotive[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetch('/api/v1/locomotives')
      .then((r) => r.json() as Promise<PagedResponse<Locomotive>>)
      .then((body) => {
        if (body.success) setLocomotives(body.data)
        else setError('Failed to load locomotives')
      })
      .catch((e: unknown) => setError(e instanceof Error ? e.message : 'Network error'))
      .finally(() => setLoading(false))
  }, [])

  return { locomotives, loading, error }
}
