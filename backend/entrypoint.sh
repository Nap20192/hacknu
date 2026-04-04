#!/bin/sh
set -e

echo "[entrypoint] Running migrations..."
./migrate \
  -path  ./db/migrations \
  -database "${DATABASE_URL}" \
  up

echo "[entrypoint] Starting server..."
exec ./server
