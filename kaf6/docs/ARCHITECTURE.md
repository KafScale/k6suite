# KAF6 Architecture Overview

## Why KAF6 exists

KAF6 exists to provide a deterministic, Kafka-protocol-correct test runner for KafScale and Kafka environments without the protocol mismatches seen in k6/xk6-kafka. The goal is to keep a k6-like scenario experience while using a modern, explicit Kafka client that can select protocol versions and avoid false negatives.

### Problem KAF6 solves

k6/xk6-kafka uses a client stack with older protocol behavior (for example, OffsetFetch v1 in consumer-group flows). KafScale supports newer protocol versions and intentionally does not emulate every legacy behavior. This mismatch causes timeouts and misleading failures in smoke tests. KAF6 replaces the client layer with Franz-go to get:

- Explicit protocol version handling (OffsetFetch v5+).
- Reliable consumer group coordination where supported.
- Deterministic, short-running smoke scenarios.
- A single binary with no JavaScript runtime constraints.

## How KAF6 behaves relative to k6

- k6 runs JavaScript in a VM and cannot load arbitrary Go libraries at runtime.
- xk6-kafka embeds kafka-go with limited protocol control.
- KAF6 is a Go CLI that runs scenarios defined in JSON and uses Franz-go directly.
- KAF6 avoids false negatives caused by client/protocol mismatches and provides stable smoke behavior.

KAF6 does not replace k6. It complements it by covering Kafka protocol correctness with a modern client and by producing deterministic reports for CI.

## Architecture at a glance

```
+-------------------------+       +------------------------+
| Scenario JSON (suite/)  | ----> | KAF6 Runner (Go)        |
+-------------------------+       |  - Scenario loader      |
                                  |  - Profile resolver     |
                                  |  - Producer/Consumer    |
                                  |  - Metrics checks       |
                                  |  - Reports (HTML/JSON)  |
                                  +-----------+------------+
                                              |
                                              v
                                     +------------------+
                                     | Kafka / KafScale |
                                     +------------------+
```

## Core components

- `cmd/kaf6`: CLI entry point. Supports `run`, `run-suite`, `select`, and report rendering.
- `internal/scenario`: JSON scenario model and profile resolution.
- `internal/engine`: Test execution (produce, consume, metrics, topic management).
- `internal/report`: Unified HTML + JSON reports, tabs per profile.
- `config/profiles.json` and `suite/profiles.json`: Profile sources.

## Execution flow

1. Load scenario JSON.
2. Resolve profile (suite-local or fallback config).
3. Preflight connectivity.
4. Create or recreate topics (if configured).
5. Run producer, consumer, and optional metrics scenarios.
6. Evaluate checks and mark pass/fail.
7. Write JSON data and HTML report.

## Reports and re-rendering

KAF6 separates report data from rendering:

- `report.json` stores data.
- `report.html` is rendered from data.
- `kaf6 render-report report.json` regenerates HTML without re-running tests.

## Summary

KAF6 exists because k6’s Kafka extension is limited by protocol versioning and JS execution constraints. By switching to Franz-go and a Go-native runner, KAF6 delivers deterministic, protocol-correct smoke testing with clear reporting, while still keeping a k6-like scenario UX.
