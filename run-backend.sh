#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Crawler runtime defaults (can be overridden via env vars)
MAX_CONCURRENCY="${MAX_CONCURRENCY:-5}"
MAX_PAGES="${MAX_PAGES:-5}"

# Query engine DB defaults (can be overridden via env vars)
export MONGO_HOST="${MONGO_HOST:-localhost}"
export MONGO_PORT="${MONGO_PORT:-27017}"
export MONGO_USERNAME="${MONGO_USERNAME:-admin}"
export MONGO_PASSWORD="${MONGO_PASSWORD:-pass123}"
export MONGO_DB="${MONGO_DB:-mongo-test}"

pids=()

start_service() {
  local name="$1"
  local dir="$2"
  local cmd="$3"

  echo "[backend] starting ${name}..."
  (
    cd "${dir}"
    eval "${cmd}"
  ) &
  pids+=("$!")
}

cleanup() {
  echo
  echo "[backend] stopping all services..."
  for pid in "${pids[@]:-}"; do
    if kill -0 "${pid}" 2>/dev/null; then
      kill "${pid}" 2>/dev/null || true
    fi
  done
  wait || true
}

trap cleanup EXIT INT TERM

start_service "indexer" "${SCRIPT_DIR}/src/indexer" "go run main.go"
start_service "crawler" "${SCRIPT_DIR}/src/crawler" "go run main.go --max-concurrency ${MAX_CONCURRENCY} --max-pages ${MAX_PAGES}"
start_service "query_engine" "${SCRIPT_DIR}/src/query_engine" "go run main.go"

echo "[backend] all services launched"
echo "[backend] MAX_CONCURRENCY=${MAX_CONCURRENCY}, MAX_PAGES=${MAX_PAGES}, MONGO_DB=${MONGO_DB}"
echo "[backend] press Ctrl+C to stop all services"

wait
