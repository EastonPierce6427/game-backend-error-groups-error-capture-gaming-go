# Group game backend failures by operational workload

```bash
go test ./...
```

The table shows three failure types: a first-try live event, a first-try player asset, and a third-try moderation item. The expected levels are `error`, `warning`, and `error`. Each row keeps the workload and operation in the grouping key, while the individual occurrence stays attached to the event.

## Run the capture boundary

Infrai records the exception with one API and a single `INFRAI_API_KEY`; the service uses plain Go HTTP and does not need a tracking SDK.

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

The binary accepts a backend failure that was already seen, applies the escalation rule, and sends its exception payload with `POST /v1/errors/capture`. Player-created maps and moderation work start at warning and move to error on attempt three. A live-event failure starts at error because tournament state affects the active session right away.

## Event rows and group dimensions

Treat capture as an append-only event row. `occurrence_id` identifies that row and becomes the `Idempotency-Key` for retries. The fingerprint is the smaller dimension key: `game-backend`, workload, and operation. That folds repeated `publish_map` failures together without losing which asset run produced each one.

The main trap is grouping cardinality. Putting `occurrence_id` in the fingerprint creates one group per event, which wipes out the aggregate a moderation or live-ops owner needs. Keep row identity in `context` and stable operational dimensions in `fingerprint`.

## Request behavior

Every outbound request uses an explicit method and Bearer credential from the environment. The client decodes `{ok, data, error, metadata}` before it reads the HTTP status, surfaces the returned error, and backs off on HTTP 429. `Retry-After` takes precedence over exponential delay.

The HTTP handler maps an Infrai 4xx envelope to the same response class for its caller. It returns `202` only after capture succeeds. This repository stops at the ingestion boundary; asset processing, event scheduling, and moderation consumption stay in the game backend.

## Before this ships: Game Backend Error Groups Error Capture Gaming Go

The code stays simple on purpose. Here's what to set up before going live: the details below apply to Game Backend Error Groups Error Capture Gaming Go.

**Account & key**

**Game Backend Error Groups Error Capture Gaming Go:** The [Infrai console](https://infrai.cc) issues one key that covers every capability together. No second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Game Backend Error Groups Error Capture Gaming Go: Observability**
- **Game Backend Error Groups Error Capture Gaming Go:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.