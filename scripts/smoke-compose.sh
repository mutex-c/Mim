#!/usr/bin/env bash
#
# End-to-end smoke test for the Mim microservice cluster.
#
# Brings up the Docker Compose stack (redis / etcd / kafka / logic / gateway / transfer),
# waits for it to be healthy, then sends one real chat message through the gateway and
# asserts that the receiver actually gets pushed the frame.
#
# This is the only check that catches deployment-level failures which unit tests cannot
# see: wrong advertised addresses, a consumer group that joined before its topic existed,
# or a Kafka pipeline that silently drops every message.
#
# Usage:
#   scripts/smoke-compose.sh                # bring the stack up, test, leave it running
#   scripts/smoke-compose.sh --no-up        # test an already running stack
#   scripts/smoke-compose.sh --down         # tear the stack down afterwards
#   scripts/smoke-compose.sh --build        # rebuild images first
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEPLOY="$ROOT/backend/deploy"
GATEWAY_URL="http://127.0.0.1:8081"
WS_URL="ws://127.0.0.1:8081/ws"

DO_UP=1
DO_BUILD=0
DO_DOWN=0
UID_A=101
UID_B=202
WAIT=20

usage() {
  sed -n '2,20p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
  exit "${1:-0}"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-up) DO_UP=0 ;;
    --build) DO_BUILD=1 ;;
    --down) DO_DOWN=1 ;;
    --gw) WS_URL="$2"; shift ;;
    --uid-a) UID_A="$2"; shift ;;
    --uid-b) UID_B="$2"; shift ;;
    --timeout) WAIT="$2"; shift ;;
    -h|--help) usage 0 ;;
    *) echo "unknown option: $1" >&2; usage 1 ;;
  esac
  shift
done

compose() { (cd "$DEPLOY" && docker compose "$@"); }

log() { printf '\n\033[1m==> %s\033[0m\n' "$*"; }

wait_for() {
  local label="$1" timeout="$2"
  shift 2
  local deadline=$((SECONDS + timeout))
  while (( SECONDS < deadline )); do
    if "$@" >/dev/null 2>&1; then
      echo "  ok: $label"
      return 0
    fi
    sleep 1
  done
  echo "  FAILED: $label (timeout ${timeout}s)" >&2
  return 1
}

kafka_ready() { docker exec mim-kafka kafka-topics --bootstrap-server localhost:9092 --list; }
gateway_ready() { curl -sf "$GATEWAY_URL/health"; }

# A consumer that has joined but holds zero partitions consumes nothing, so the smoke
# message would be produced into a pipeline that nobody is reading yet.
consumer_ready() {
  docker exec mim-kafka kafka-consumer-groups --bootstrap-server localhost:9092 \
    --describe --group mim_transfer_group --members 2>/dev/null \
    | awk 'NR > 1 && $NF + 0 > 0 { found = 1 } END { exit(found ? 0 : 1) }'
}

cleanup() {
  if [[ "$DO_DOWN" == "1" ]]; then
    log "Tearing down the stack"
    compose down
  fi
}
trap cleanup EXIT

if ! docker info >/dev/null 2>&1; then
  echo "Docker daemon is not reachable." >&2
  exit 1
fi

if [[ "$DO_UP" == "1" ]]; then
  if [[ "$DO_BUILD" == "1" ]]; then
    log "Building images"
    compose build
  fi

  log "Starting infrastructure (redis, etcd, kafka)"
  compose up -d redis etcd kafka

  log "Waiting for Kafka"
  wait_for "kafka broker" 90 kafka_ready

  log "Starting services (logic, gateway, transfer)"
  compose up -d logic gateway transfer

  log "Waiting for gateway"
  wait_for "gateway /health" 60 gateway_ready

  log "Waiting for transfer consumer assignment"
  wait_for "transfer consumer partitions" 60 consumer_ready
fi

log "Running cluster smoke test (user $UID_A -> user $UID_B)"
cd "$ROOT/backend"
go run ./cmd/tools/cluster_smoke \
  -gw "$WS_URL" \
  -uid-a "$UID_A" \
  -uid-b "$UID_B" \
  -timeout "${WAIT}s"

log "Verifying consumer group assignment"
(cd "$DEPLOY" && docker exec mim-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --describe --group mim_transfer_group --members 2>/dev/null || true)

log "PASSED"
