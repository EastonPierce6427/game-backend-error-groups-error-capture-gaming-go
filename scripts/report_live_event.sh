#!/bin/sh
set -eu

curl --request POST http://127.0.0.1:8080 \
  --header 'Content-Type: application/json' \
  --data '{"workload":"live_event","operation":"close_tournament","occurrence_id":"round-2026-08-20-17","attempt":1,"exception":"round close transaction failed"}'
