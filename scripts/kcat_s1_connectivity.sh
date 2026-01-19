#!/bin/sh
set -e

BROKER="${BROKER:-127.0.0.1:39092}"

echo "S1 connectivity check (kcat)"
echo "Broker: ${BROKER}"

if ! command -v kcat >/dev/null 2>&1; then
  echo "kcat not found. Install with: brew install kcat (macOS) or apt-get install kcat (Linux)."
  exit 1
fi

echo "Fetching broker metadata..."
kcat -b "${BROKER}" -X broker.address.family=v4 -L >/dev/null
echo "OK: metadata fetched"
