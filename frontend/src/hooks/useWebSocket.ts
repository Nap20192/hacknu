import { useEffect, useRef, useState, useCallback } from 'react'

type WsStatus = 'connecting' | 'connected' | 'disconnected' | 'error'

interface UseWebSocketOptions {
  onMessage: (data: string) => void
  reconnectDelay?: number
}

export function useWebSocket(url: string | null, { onMessage, reconnectDelay = 3000 }: UseWebSocketOptions) {
  const [status, setStatus] = useState<WsStatus>('disconnected')
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const onMessageRef = useRef(onMessage)
  onMessageRef.current = onMessage

  const connect = useCallback(() => {
    if (!url) return
    if (wsRef.current?.readyState === WebSocket.OPEN) return

    setStatus('connecting')
    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => setStatus('connected')

    ws.onmessage = (evt) => onMessageRef.current(evt.data as string)

    ws.onclose = () => {
      setStatus('disconnected')
      reconnectTimer.current = setTimeout(connect, reconnectDelay)
    }

    ws.onerror = () => {
      setStatus('error')
      ws.close()
    }
  }, [url, reconnectDelay])

  useEffect(() => {
    connect()
    return () => {
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [connect])

  return { status }
}
