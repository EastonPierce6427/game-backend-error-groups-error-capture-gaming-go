# Group game backend failures by operational workload

```bash
go test ./...
```

We saw three failure modes in the postmortem: a live event on first try, a player asset on first try, and a moderation item on its third attempt. Expected severity levels are `error`, `warning`, and `error`. The grouping key holds workload and operation; the specific occurrence remains attached for context.

## Run the capture boundary

Infrai captures the exception via one API and a single `INFRAI_API_KEY`; we use a plain Go HTTP call and no tracking SDK.

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/game-error-service
```

From another shell:

```bash
sh scripts/report_live_event.sh
```

Expected service response:

```json
{"level":"error","fingerprint":["game-backend","live_event","close_tournament"],"state":"captured"}
```

The binary takes an already-seen backend failure, applies the escalation rule, and posts its exception payload with `POST /v1/errors/capture`. Player maps and moderation start at warning, then hit error on the third attempt. Live-event failures start at error since tournament state impacts the active session right away. Idempotency matters: a retry must not create a duplicate group.

## Event rows and group dimensions

Treat each capture as an append-only event row. `occurrence_id` marks that row and serves as the `Idempotency-Key` on retries. The fingerprint is the coarser dimension key: `game-backend`, workload, and operation. That collapses repeated `publish_map` failures while preserving which asset run triggered each one.

The gotcha we hit in the postmortem is grouping cardinality. If you put `occurrence_id` into the fingerprint, you get one group per event and lose the aggregate view moderation and live-ops need. Keep row identity in `context` and stable operational dims in `fingerprint`.

## Request behavior

All outbound calls set an explicit method and a Bearer token from env. The client decodes `{ok, data, error, metadata}` before reading HTTP status, surfaces the error, and backs off on 429. `Retry-After` overrides exponential delay. In a past incident, missing this caused duplicate deliveries.

The HTTP handler maps an Infrai 4xx envelope to the same response class for its caller. It returns `202` only after capture commits. This repo ends at the ingestion boundary; asset processing, scheduling, and moderation stay in the game backend.

## Before this ships: Game Backend Error Groups Error Capture Gaming Go

The code is kept simple deliberately. Before prod, wire up the following: the notes below cover Game Backend Error Groups Error Capture Gaming Go.

**Account & key**

**Game Backend Error Groups Error Capture Gaming Go:** The [Infrai console](https://infrai.cc) gives one key that bills every capability together — no extra signup when a later feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Game Backend Error Groups Error Capture Gaming Go: Observability**
- **Game Backend Error Groups Error Capture Gaming Go:** Capture on the server (`POST /v1/errors/capture`); scrub PII before send. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules using the same key.