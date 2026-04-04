# Frontend Alerts Module

Реализация long polling для мониторинга алертов локомотивов на фронте.

## Структура

```
src/
├── api/
│   └── alerts.ts              # API client для работы с alerts endpoints
├── hooks/
│   └── useAlertsPolling.ts    # Custom hook для long polling
├── components/
│   ├── AlertsPanel.tsx        # Main компонент-панель с статистикой
│   ├── AlertsPanel.css
│   ├── AlertsList.tsx         # Компонент со списком алертов
│   └── AlertsList.css
├── types/
│   └── alerts.ts              # TypeScript типы для alerts
└── App.tsx                    # Главное приложение
```

## Использование

### 1. Базовое использование в компоненте

```tsx
import { AlertsPanel } from './components/AlertsPanel'

function App() {
  return (
    <AlertsPanel 
      locomotiveId="loco-01" 
      pollingInterval={3000}  // опционально, дефолт 3000мс
    />
  )
}
```

### 2. Прямое использование hook'а

```tsx
import { useAlertsPolling } from './hooks/useAlertsPolling'

function MyComponent() {
  const { active, loading, error, acknowledgeAlert } = useAlertsPolling(
    'loco-01',
    {
      interval: 3000,      //Interval между запросами (мс)
      timeout: 5000,       // Timeout einzelного запроса
      maxRetries: 5        // Макс попыток до остановки
    }
  )

  return (
    <>
      {active.map(alert => (
        <div key={alert.id}>
          {alert.message}
          <button onClick={() => acknowledgeAlert(alert.id)}>
            Подтвердить
          </button>
        </div>
      ))}
    </>
  )
}
```

### 3. Через API напрямую

```tsx
import { alertsApi } from './api/alerts'

// Получить активные алерты
const response = await alertsApi.getActiveAlerts('loco-01')

// Получить историю алертов
const history = await alertsApi.getAlertsHistory(
  'loco-01',
  new Date('2024-01-01'),
  new Date('2024-01-02')
)

// Подтвердить алерт
const result = await alertsApi.acknowledgeAlert(123)
```

## API Endpoints

### GET `/api/v1/locomotives/:id/alerts`
Получить активные (неразрешенные) алерты для локомотива.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "locomotive_id": "loco-01",
      "triggered_at": "2024-01-15T10:30:00Z",
      "resolved_at": null,
      "severity": "critical",
      "code": "TEMP_HIGH",
      "metric_name": "engine_temp_c",
      "metric_value": 95.5,
      "threshold": 90.0,
      "message": "Engine temperature critical",
      "recommendation": "Stop and cool down",
      "acknowledged": false
    }
  ],
  "total": 1
}
```

### GET `/api/v1/locomotives/:id/alerts/history`
Получить историю алертов в диапазоне времени.

**Query Parameters:**
- `from` - RFC3339 start time (опционально)
- `to` - RFC3339 end time (опционально)
- `limit` - максимум результатов (опционально)

### POST `/api/v1/alerts/:alertId/acknowledge`
Подтвердить алерт.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "acknowledged": true,
    ...
  }
}
```

## Конфигурация

### Переменные окружения

Создайте файл `.env` в корне frontend проекта:

```env
VITE_API_URL=http://localhost:8081
```

## Long Polling Механика

Hook `useAlertsPolling` реализует следующую логику:

1. **Initial Fetch** - при монтировании компонента или смене `locomotiveId`
2. **Polling Loop** - запросы повторяются с интервалом `interval`
3. **Error Handling** - при ошибке увеличивается счетчик retries
4. **Auto-stop** - polling останавливается после `maxRetries` ошибок подряд
5. **Auto-recover** - счетчик обнуляется при успешном запросе
6. **Request Timeout** - каждый запрос прерывается через `timeout` мс (AbortController)

## Компоненты

### AlertsPanel
Главный компонент-панель с:
- Статистикой (всего алертов, неподтвержденных, критичных)
- Временем последнего обновления
- Список алертов с возможностью подтверждения

### AlertsList
Компонент со списком алертов включает:
- Отображение severity (warning/critical) с цветовой кодировкой
- Метрика с текущим значением и threshold
- Рекомендации по устранению
- Кнопка подтверждения (disable после подтверждения)
- "Относительное время" (2м назад, 1ч назад, итд)

## Стили

Компоненты включают полный набор CSS стилей:
- **Responsive design** - адаптация для мобильных, планшетов, десктопа
- **Color scheme** - warning (оранжевый) и critical (красный)
- **Animations** - плавные переходы и состояния
- **Accessibility** - правильная фокусировка и контраст

## Обработка ошибок

- Сетевые ошибки автоматически перехватываются
- Timeout запросов - 5 сек по умолчанию
- Graceful degradation - отображение сообщения об ошибке
- Автоматический retry после заданного интервала

## Performance

- **Debouncing**: Используется single polling interval
- **Abort signals**: Прерывание старых запросов при смене locomotive
- **State updates**: Minimal re-renders благодаря React hooks
- **Memory**: Cleanup на unmount, отмена pending запросов
