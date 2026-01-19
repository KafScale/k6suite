# KAF6 Plan

KAF6 is a K6-style Kafka/KafScale test orchestrator built on Franz-go. It provides deterministic scenarios, consumer-group based consumption by default, and per-run HTML reports.

## Goals

- Deterministic Kafka/KafScale smoke and validation tests.
- Declarative scenarios with reproducible outputs.
- Consumer groups by default to match real client behavior.
- CI-friendly execution with machine-readable and human-readable reports.

## Non-Goals (v1)

- Advanced Kafka group semantics (beyond simple group consumption).
- Embedded k6 JavaScript execution.
- Live dashboards or streaming UI.

## Architecture Overview

```
kaf6/
├── cmd/kaf6/              # CLI entrypoint
├── internal/
│   ├── engine/            # scenario runner, VU/clients, schedules
│   ├── kafka/             # Franz-go adapters (producer/consumer)
│   ├── scenario/          # schema, parsing, validation
│   ├── metrics/           # counters, histograms, summaries
│   └── report/            # JSON + HTML report generation
├── schemas/               # JSON schema for scenarios
└── suite/                 # sample scenarios
```

Execution flow:
1. Load scenario file (YAML/JSON).
2. Validate schema and resolve defaults.
3. Start scenario runners with explicit time bounds.
4. Collect metrics + checks.
5. Emit results (JSON + HTML).

## Scenario Planning

S1: Connectivity
- Verify broker reachable and metadata fetchable.

S2: Metrics Health (KafScale targets)
- Verify metrics endpoint returns HTTP 200 and non-empty payload.

S3: Single Produce/Consume
- Produce one message, consume via consumer group, validate UUID.

S3: Multi-Producer Single-Consumer
- Concurrent producers, one group consumer; validate counts and UUIDs.

S3: Shared Connection (Single VU)
- Reuse writer/reader in one client; validate loop success (group-based).

S4: Durability Smoke (Optional)
- Produce, restart broker, consume via consumer group.

S5: Storage Throttle (Optional)
- Inject latency; verify backpressure and no corruption.

S6: Scale-Out Smoke (Optional)
- Validate produce/consume continuity during scale changes.

## Scenario Schema (Draft)

```yaml
profile: local-service
brokers:
  - localhost:39092

topics:
  - name: smoke
    partitions: 1

scenarios:
  producer:
    type: produce
    clients: 5
    messages: 1000
    rate: 50/s
    value:
      json:
        uuid: "{{uuid}}"
        ts: "{{now}}"

  consumer:
    type: consume
    clients: 1
    group:
      id: smoke-group
    topic: smoke
    offset: earliest
    limit: 1000

checks:
  - name: delivered_all
    type: count_equals
    expected: 1000
```

## Reporting Plan (Per Run)

Outputs:
- `reports/<run-id>/summary.json`
- `reports/<run-id>/report.html`

HTML report sections:
- Run header (timestamp, profile, brokers).
- Scenario metrics (sent/received, errors, latency percentiles).
- Checks table (pass/fail with expected vs actual).
- Timeline graph (produce/consume rate).

Implementation approach:
- Single-file HTML template with inline CSS.
- Render from JSON summary + time-series buffers.
