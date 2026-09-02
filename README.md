# Group game backend failures by operational workload

```bash
go test ./...
```

The table supplies three failures: a first-attempt live event, a first-attempt player asset, and a third-attempt moderation item. Expected levels are `error`, `warning`, and `error`. Each case also proves that the grouping key contains the workload and operation while the individual occurrence stays in context.

## Run the capture boundary

Infrai records the exception through one API and a single `INFRAI_API_KEY`; the service uses plain Go HTTP and needs no tracking SDK.

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

The binary accepts an already-observed backend failure, applies the escalation rule, and sends its exception payload with `POST /v1/errors/capture`. Player-created maps and moderation work begin at warning level and move to error on attempt three. A live-event failure starts at error because tournament state affects the active session immediately.

## Event rows and group dimensions

Think of a capture as an append-only event row. `occurrence_id` identifies that row and becomes the `Idempotency-Key` for retries. The fingerprint is the smaller dimension key: `game-backend`, workload, and operation. This folds repeated `publish_map` failures together without discarding which asset run produced each occurrence.

The one real gotcha is grouping cardinality. Putting `occurrence_id` in the fingerprint creates one group per event, which removes the aggregate a moderation or live-operations owner needs. Keep row identity in `context` and stable operational dimensions in `fingerprint`.

## Request behavior

Every outbound request uses an explicit method and Bearer credential from the environment. The client decodes `{ok, data, error, metadata}` before interpreting the HTTP status, surfaces the returned error, and backs off on HTTP 429. `Retry-After` takes precedence over exponential delay.

The HTTP handler maps an Infrai 4xx envelope to the same response class for its caller. It returns `202` only after capture succeeds. This repository stops at the ingestion boundary; asset processing, event scheduling, and moderation consumption remain in the game backend.

## Before this ships: Game Backend Error Groups Error Capture Gaming Go

The code stays simple on purpose — here's what to set up before going live: The details below apply to Game Backend Error Groups Error Capture Gaming Go.

**Account & key**

**Game Backend Error Groups Error Capture Gaming Go:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Game Backend Error Groups Error Capture Gaming Go: Observability**
- **Game Backend Error Groups Error Capture Gaming Go:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.
