# KafScale Compatibility Charter

K6SUITE defines the compatibility contract for the KafScale ecosystem. It makes
Kafka protocol support explicit, testable, and transparent.

## Intent

- KafScale is Kafka-compatible, not a Kafka replacement.
- We validate the behaviors KafScale promises: single endpoint, stateless
  brokers, and object-storage durability.
- Every gap is documented with a test or an observation so the ecosystem can
  build on known truths.

## Compatibility Scope and Status

| Capability | Coverage (K6SUITE) | Known Status | Evidence |
|-----------|--------------------|--------------|----------|
| Basic connectivity (Metadata/ApiVersions) | `diagnose.js` | Validated when tests pass | `tests/k6/diagnose.js` |
| Produce + consume with direct partition reads | `smoke_single.js`, `smoke_shared.js` | Primary compatibility baseline | `tests/k6/` |
| Multi-producer single-consumer | `smoke_multi_producer_single_consumer.js` | Baseline behavior | `tests/k6/` |
| Topic auto-creation | `smoke_topic_autocreate.js` | Baseline behavior | `tests/k6/` |
| Metrics endpoint (KafScale only) | `smoke_metrics.js` | KafScale-specific | `tests/k6/smoke_metrics.js` |
| Consumer groups (OffsetFetch v1) | Not covered (unsupported) | Known incompatible with xk6-kafka | `OBSERVATIONS/OBSERVATION-01.md` |
| Group coordinator metadata invariants | Not covered | Known incompatible | `OBSERVATIONS/OBSERVATION-02.md` |
| Chaos, scale, upgrade scenarios | Planned | Not validated | `SPEC/scenarios.md` |

## How We Keep This Transparent

1. Add or update tests under `tests/k6/`.
2. Record protocol mismatches in `OBSERVATIONS/`.
3. Update `DOCS/test-coverage-report.md` when coverage changes.
4. Re-run the suite and publish pass/fail outcomes alongside releases.
