# KAF6 Solution Design

## Problem

k6/xk6-kafka uses older Kafka protocol behaviors that conflict with KafScale (for example, OffsetFetch v1). This creates false negatives in smoke tests. KAF6 solves this by using Franz-go directly and using consumer groups with OffsetFetch v5 by default.

## Design Principles

- Explicit protocol control (Franz-go).
- Deterministic scenarios (bounded counts and time).
- Consumer groups by default (OffsetFetch v5).
- Clear pass/fail checks with reproducible outputs.

## Components

### CLI (`cmd/kaf6`)

Commands:
- `kaf6 run <scenario.yaml>`
- `kaf6 validate <scenario.yaml>`

Responsibilities:
- Parse arguments and load scenarios.
- Configure output paths.
- Exit with non-zero status on failed checks.

### Scenario Loader (`internal/scenario`)

- YAML/JSON parsing.
- Schema validation.
- Default value resolution.

### Engine (`internal/engine`)

- Schedules producer/consumer clients.
- Enforces rate limits and run duration.
- Aggregates metrics and checks.

### Kafka Client Layer (`internal/kafka`)

Producer:
- Franz-go `kgo` client with explicit options.
- Produce UUID-tagged payloads.

Consumer:
- Group-based consumption (normal client behavior).
- Explicit offsets and timeouts.

### Metrics (`internal/metrics`)

Track:
- Produced/consumed counts.
- Latency percentiles (p50/p95/p99).
- Errors and retries.
- Lag (best-effort).

### Reports (`internal/report`)

Artifacts:
- `summary.json` (machine readable)
- `report.html` (human readable)

## Default Behavior

- Consumer groups enabled by default.
- OffsetFetch pinned to v5.
- Deterministic message counts per scenario.

## Optional Modes

Partition mode (opt-in):
- Explicitly disable group semantics.
- Use partition-based consumption for KafScale-specific verification.

## Failure Policy

- Any failed check fails the run.
- Any client error increments `errors_total`.
- Report includes root-cause hints (protocol mismatch, broker unreachable).
