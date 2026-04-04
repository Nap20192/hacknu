#!/bin/sh
set -e

# Загрузить .env если существует (для локального запуска без Docker)
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

echo "[entrypoint] Running migrations..."
./migrate \
  -path     ./db/migrations \
  -database "${DATABASE_URL}" \
  up

echo "[entrypoint] Starting server..."
exec ./server
