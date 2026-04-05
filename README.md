# Locomotive Digital Twin

Платформа мониторинга и диагностики локомотивов в реальном времени. Система принимает телеметрию с локомотивов по WebSocket, вычисляет индекс здоровья, детектирует аномалии и выдаёт предупреждения/критические алерты через REST API и WebSocket-пуш.

## Стек

| Слой | Технологии |
|------|-----------|
| Backend | Go 1.25, Fiber v2, TimescaleDB (PostgreSQL 15) |
| Frontend | React 19, TypeScript, Vite, Recharts |
| Инфраструктура | Docker Compose |

## Архитектура

```
simulation → WebSocket → pipeline (buffer → worker) → TimescaleDB
                                   ↓
                            spec/engine (правила)
                                   ↓
                        REST API + WebSocket push → frontend
```

- **simulation** — генератор телеметрии, имитирует работу локомотива
- **pipeline** — буферизует входящие события и пакетно записывает в БД
- **spec/engine** — движок правил: вычисляет индекс здоровья и формирует алерты
- **hub** — WebSocket-менеджер рассылки событий клиентам

## Быстрый старт

### Предварительные требования

- [Docker](https://docs.docker.com/get-docker/) + Docker Compose v2
- [Node.js](https://nodejs.org/) 20+ и npm (для фронтенда)

### 1. Backend (Docker Compose)

```bash
cd backend
```
# Создайте файл с переменными окружения
```bash
cp .env.example .env 
```

# Поднимите все сервисы (БД + сервер + симулятор)
```bash
docker compose up --build
```

Сервисы после запуска:
- API сервер: http://localhost:8080
- Swagger UI: http://localhost:8080/swagger/index.html
- PostgreSQL: localhost:5433

### 2. Frontend

```bash
cd frontend
npm install
npm run dev
```

Приложение откроется на http://localhost:5173

## API

Основные эндпоинты (`/api/v1`):

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/locomotives` | Список всех локомотивов |
| `GET` | `/locomotives/:id` | Данные локомотива |
| `GET` | `/locomotives/:id/health` | Текущий индекс здоровья |
| `GET` | `/locomotives/:id/health/history` | История индекса здоровья |
| `GET` | `/locomotives/:id/telemetry` | История телеметрии |
| `GET` | `/locomotives/:id/alerts` | Активные алерты |
| `GET` | `/locomotives/:id/alerts/history` | История алертов |
| `POST` | `/alerts/:alertId/acknowledge` | Подтвердить алерт |
| `WS` | `/ws/telemetry` | WebSocket для приёма телеметрии |

Полная документация: Swagger UI после запуска backend.

## Переменные окружения (backend)

| Переменная | По умолчанию | Описание |
|------------|-------------|----------|
| `DATABASE_URL` | — | PostgreSQL DSN |
| `PORT` | `8081` | Порт HTTP-сервера |
| `LOG_LEVEL` | `INFO` | Уровень логирования |
| `BUFFER_CAP` | `50` | Размер буфера телеметрии |
| `FLUSH_INTERVAL_MS` | `500` | Интервал сброса буфера (мс) |
| `TELEMETRY_RETENTION` | `72h` | Время хранения телеметрии |
| `EMA_RECALC_INTERVAL` | `1h` | Интервал пересчёта EMA |

