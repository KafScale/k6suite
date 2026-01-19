# KAF6 Plan: Franz-go Test Orchestrator (K6-Style)

This document plans a new `kaf6` module: a K6-style test orchestrator built on Franz-go to validate KafScale behavior without consumer-group protocol mismatches.

## Goals

- Deterministic Kafka/KafScale smoke and validation tests.
- K6-like scenario definitions with reproducible outputs.
- Partition-based consumption by default (no consumer groups).
- CI-friendly execution with machine-readable and human-readable reports.

## Non-Goals (v1)

- Full Kafka group semantics coverage.
- Interactive dashboards or live streaming UI.
- Embedded k6 JavaScript execution.

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
└── examples/              # sample scenarios
```

Execution flow:
1. Load scenario file (YAML/JSON).
2. Validate schema and resolve defaults.
3. Start scenario runners (producers/consumers) with explicit time bounds.
4. Collect metrics + checks.
5. Emit results (JSON + HTML).

## Scenario Planning

All scenarios must be partition-based by default and avoid consumer groups.

### S1: Connectivity

**Purpose:** Validate broker connectivity and metadata fetch.

**Checks:**
- Broker reachable.
- Metadata for topic can be fetched (if topic exists or auto-created).

### S2: Metrics Health

**Purpose:** Verify metrics endpoint for KafScale targets.

**Checks:**
- HTTP 200.
- Non-empty payload.

### S3: Single Produce/Consume

**Purpose:** Produce one message and consume by partition offset.

**Checks:**
- Produced messages count = 1.
- Consumed message UUID matches.

### S3: Multi-Producer, Single-Consumer

**Purpose:** Concurrent produce with single partition consumer.

**Checks:**
- Produced count == expected.
- Consumed count >= expected.
- No duplicate UUIDs.

### S3: Shared Connection (Single VU)

**Purpose:** Reuse writer/reader in one client for quick correctness.

**Checks:**
- Produce/consume loop succeeds.
- Latency threshold enforced.

### S4: Durability Smoke (Optional)

**Purpose:** Produce, restart broker, then consume.

**Checks:**
- Messages consumed after restart.
- Offsets monotonic.

### S5: Storage Throttle (Optional)

**Purpose:** Validate backpressure behavior with injected latency.

**Checks:**
- Throughput drops, no corruption.

### S6: Scale-Out Smoke (Optional)

**Purpose:** Validate metadata stability during scaling.

**Checks:**
- Produce/consume continuity.

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
    topic: smoke
    partition: 0
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

### JSON Summary (machine-readable)

```json
{
  "run_id": "20260116-1454",
  "profile": "local-service",
  "brokers": ["localhost:39092"],
  "scenarios": {
    "producer": {"sent": 1000, "errors": 0, "p95_ms": 12},
    "consumer": {"received": 1000, "errors": 0, "p95_ms": 18}
  },
  "checks": {
    "delivered_all": "pass",
    "no_duplicates": "pass"
  },
  "status": "pass",
  "duration_ms": 1842
}
```

### HTML Report (human-readable)

Sections:
- Header: run ID, timestamp, profile, brokers.
- Scenario results: sent/received, errors, latency percentiles.
- Checks table: pass/fail with expected vs actual.
- Timeline graph: produce/consume rate over time.
- Notes: warnings or protocol compatibility hints.

Implementation approach:
- Use a small embedded HTML template with inline CSS.
- Render from the JSON summary + time-series buffers.
- Keep it single-file for easy CI artifact retention.

## Next Steps

1. Create `kaf6/` skeleton with CLI and scenario schema.
2. Implement produce/consume with Franz-go (partition mode).
3. Add JSON summary output.
4. Add HTML report generator.
