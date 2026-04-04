interface Props {
  status: 'connecting' | 'connected' | 'disconnected' | 'error'
}

const labels: Record<string, string> = {
  connecting: 'Подключение...',
  connected: 'Онлайн',
  disconnected: 'Нет связи',
  error: 'Ошибка',
}

const colors: Record<string, string> = {
  connecting: '#f59e0b',
  connected: '#22c55e',
  disconnected: '#94a3b8',
  error: '#ef4444',
}

export function ConnectionBadge({ status }: Props) {
  return (
    <span className="connection-badge" style={{ color: colors[status] }}>
      <span className="badge-dot" style={{ background: colors[status] }} />
      {labels[status]}
    </span>
  )
}
