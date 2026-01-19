# K6SUITE for KafScale

**Trust Contract Validation System for Stateless Kafka over Object Storage**

---

## What is K6SUITE?

K6SUITE is not a benchmark. It is a **truth machine** that proves KafScale delivers on its architectural promises:

- **Single IP Access** - Clients connect through one endpoint, not individual brokers
- **Stateless Brokers** - Brokers are disposable with no persistent state
- **Object Storage Durability** - All data lives in S3-compatible storage
- **Zero-Downtime Operations** - Kill, replace, scale brokers without client impact
- **Transparent Scaling** - Add brokers without metadata changes or reconnections

If these tests pass, KafScale is real. If they fail, the architecture is broken.

KafScale is Kafka-compatible but not a Kafka replacement. The suite targets KafScale's durability and operational guarantees rather than full Kafka parity.

---

## Why This Exists

Classic Apache Kafka has fundamental limitations:
- Brokers hold state, making them precious
- Scaling requires rebalancing and leader elections
- Broker failures cause consumer group churn
- Rolling upgrades require careful orchestration

KafScale eliminates these by putting object storage (MinIO/S3/GCS) at the center. K6SUITE proves this actually works.

---

## Compatibility & Ecosystem Intent

K6SUITE is the compatibility contract for the KafScale ecosystem. It makes Kafka-protocol support explicit and testable:

- KafScale is Kafka-compatible, not a Kafka replacement.
- We validate the behaviors KafScale promises: single endpoint, stateless brokers, and object-storage durability.
- Every gap is documented with a test or an observation so the ecosystem can build on known truths.

See [DOCS/compatibility.md](DOCS/compatibility.md) and [OBSERVATIONS/](OBSERVATIONS/) for current coverage and known gaps.

---

## Quick Start

### Prerequisites

1. **K6 with xk6-kafka extension**
   ```bash
   # Install xk6
   go install go.k6.io/xk6/cmd/xk6@latest

   # Build k6 with kafka extension
   xk6 build --with github.com/mostafa/xk6-kafka

   # This creates ./k6 binary
   ```

2. **KafScale + object storage** (local or remote)
   - This repo does not ship a Docker Compose stack.
   - Use your existing KafScale deployment or bring up your own compose stack.

### 1. Start the Environment

```bash
# Ensure KafScale and object storage are running
# Example (from your own compose stack):
docker-compose up -d
```

If you use a compose stack, this typically launches:
- **MinIO** - S3-compatible object storage (ports 9000/9001)
- **KafScale** - Kafka-compatible broker (port 39092)

### 2. Verify Connectivity

```bash
./k6 run tests/k6/diagnose.js
```

Expected output:
```
✓ Connection created successfully
✓ Writer created successfully
✓ Reader created successfully
```

### 3. Run Smoke Test

```bash
./k6 run tests/k6/smoke_single.js
```

If producers write successfully, basic KafScale connectivity is working.

## KAF6 (Franz-go Runner)

Planned K6-style test orchestrator using Franz-go with per-run HTML reports. See `DOCS/kaf6-plan.md`.

### Kafka Compatibility Mode (Optional)

To run the lowest-common-denominator suite against Apache Kafka:

```bash
K6_PROFILE=kafka-local K6_TARGET=kafka ./k6 run tests/k6/smoke_single.js
```

---

## The Five Critical Questions

K6SUITE answers these questions with binary yes/no results:

| # | Question | Test | Status |
|---|----------|------|--------|
| 1 | Can clients connect through a single IP? | `diagnose.js` | Working |
| 2 | Can they produce and consume correctly? | `smoke_*.js` | Partial |
| 3 | Does data survive broker death? | `chaos_broker.js` | Planned |
| 4 | Does data survive object-store slowness? | `objectstore_slow.js` | Planned |
| 5 | Can we scale brokers without touching clients? | `scaleout.js` | Planned |

---

## Test Suite Overview

### Scenario S1: Connectivity

#### `diagnose.js` - Connection Diagnostic
Validates basic Kafka protocol connectivity through KafScale's single IP.

```bash
./k6 run tests/k6/diagnose.js
```

### Scenario S3: Produce/Consume Correctness

#### `smoke_single.js` - Single Message Test
Minimal test: produce one message, consume it back. Use this first.

```bash
./k6 run tests/k6/smoke_single.js
```

#### `smoke_concurrent.js` - Concurrent Test
5 producers and 5 consumers in separate phases. Tests multi-VU behavior.

```bash
./k6 run tests/k6/smoke_concurrent.js
```

#### `smoke_shared.js` - Shared Connection Test
Shared writer/reader across VUs. Tests connection reuse and UUID verification.

```bash
./k6 run tests/k6/smoke_shared.js
```

### Scenario S2: Observability

#### `smoke_metrics.js` - Metrics Endpoint
Validates KafScale metrics endpoint is accessible.

```bash
./k6 run tests/k6/smoke_metrics.js
```

### Additional S3 Tests

#### `smoke_topic_autocreate.js` - Topic Auto-Creation
Validates topic auto-creation and immediate consume.

```bash
./k6 run tests/k6/smoke_topic_autocreate.js
```

#### `smoke_multi_producer_single_consumer.js` - Multi-Producer, Single-Consumer
Validates multi-producer, single-consumer behavior with direct partition reads.

```bash
./k6 run tests/k6/smoke_multi_producer_single_consumer.js
```

#### `smoke_consumer_group.js` - Direct Partition Consume
Validates direct partition consumption without consumer-group coordination.

```bash
./k6 run tests/k6/smoke_consumer_group.js
```

### Run All Tests

```bash
# Individual tests
./k6 run tests/k6/diagnose.js
./k6 run tests/k6/smoke_metrics.js
./k6 run tests/k6/smoke_single.js
./k6 run tests/k6/smoke_concurrent.js
./k6 run tests/k6/smoke_shared.js
./k6 run tests/k6/smoke_topic_autocreate.js
./k6 run tests/k6/smoke_multi_producer_single_consumer.js
./k6 run tests/k6/smoke_consumer_group.js

# Run all sequentially (stops on first failure)
./k6 run tests/k6/diagnose.js && \
./k6 run tests/k6/smoke_metrics.js && \
./k6 run tests/k6/smoke_single.js && \
./k6 run tests/k6/smoke_concurrent.js && \
./k6 run tests/k6/smoke_shared.js && \
./k6 run tests/k6/smoke_topic_autocreate.js && \
./k6 run tests/k6/smoke_multi_producer_single_consumer.js && \
./k6 run tests/k6/smoke_consumer_group.js

# With specific profile
K6_PROFILE=local-docker ./k6 run tests/k6/diagnose.js && \
K6_PROFILE=local-docker ./k6 run tests/k6/smoke_single.js
```

### Planned Tests

#### `chaos_broker.js` - Broker Kill Test (S4)
Continuous traffic while randomly killing broker pods. No client errors allowed.

#### `objectstore_slow.js` - Storage Throttle Test (S5)
Inject latency into MinIO. Verify backpressure works and no data is lost.

#### `scaleout.js` - Scale-Out Test (S6)
Start with 1 broker, scale to 3 under load. No client reconnections allowed.

---

## Project Structure

```
k6suite/
├── README.md                 # This file
├── CLAUDE.md                 # Developer guide
├── OBSERVATIONS/             # Protocol gaps and behavior notes
├── k6                        # Compiled k6 binary with xk6-kafka
├── kaf6/                     # Franz-go runner (planned/experimental)
│
├── tests/
│   └── k6/
│       ├── diagnose.js              # S1: Connection diagnostic
│       ├── smoke_metrics.js         # S2: Metrics endpoint
│       ├── smoke_single.js          # S3: Single message test
│       ├── smoke_concurrent.js      # S3: Multi-VU concurrent test
│       ├── smoke_shared.js          # S3: Shared connection test
│       ├── smoke_topic_autocreate.js # S3: Topic auto-creation
│       ├── smoke_multi_producer_single_consumer.js # S3: Multi-producer/single-consumer
│       └── smoke_consumer_group.js  # S3: Direct partition consume
│
├── SPEC/                     # Requirements & specifications
│   ├── features.md           # Functional requirements (F1-F7)
│   ├── nfr.md                # Non-functional requirements (NFR1-NFR5)
│   └── scenarios.md          # Test scenarios (S1-S8)
│
├── config/
│   ├── profiles.js           # Profile loader
│   └── profiles.json         # Execution profile configuration
│
├── DOCS/                     # User documentation
│   ├── getting-started.md    # Quick start guide
│   ├── compatibility.md      # Compatibility charter and status
│   ├── release-checklist.md  # Release readiness checklist
│   ├── test-reference.md     # Test documentation
│   ├── TROUBLESHOOTING.md    # Known issues & fixes
│   ├── profiles.md           # Execution profiles guide
│   └── architecture.md       # System architecture
│
├── .github/workflows/        # CI/CD pipelines
└── scripts/                  # Utility scripts
```

---

## Configuration

### Execution Profiles

Tests use execution profiles to connect to different environments. Set the `K6_PROFILE` environment variable:

```bash
# Run against Docker environment (port 9092)
K6_PROFILE=local-docker ./k6 run tests/k6/diagnose.js

# Run against local service (port 39092) - this is the default
K6_PROFILE=local-service ./k6 run tests/k6/diagnose.js

# Run against Kubernetes
K6_PROFILE=k8s-local ./k6 run tests/k6/diagnose.js
```

| Profile | Port | Description |
|---------|------|-------------|
| `local-docker` | 9092 | KafScale in Docker via docker-compose |
| `local-service` | 39092 | KafScale as local service (default) |
| `k8s-local` | 39092 | KafScale in local Kubernetes |
| `kafka-local` | 9092 | Apache Kafka running locally on port 9092 |

See [DOCS/profiles.md](DOCS/profiles.md) for details on creating custom profiles.

### Docker Environment (Example)

If you use a docker-compose stack, it typically includes:

| Service | Port | Purpose |
|---------|------|---------|
| MinIO API | 9000 | S3-compatible storage |
| MinIO Console | 9001 | Web UI (admin/minioadmin) |
| KafScale | 39092 | Kafka protocol endpoint |

### Test Parameters

```javascript
export const options = {
  vus: 5,              // Concurrent virtual users
  iterations: 100,     // Total test iterations
  duration: "30s",     // Or use duration instead
};
```

---

## Metrics

K6 captures Kafka-specific metrics:

### Producer Metrics
- `kafka_writer_message_count` - Messages produced
- `kafka_writer_error_count` - Producer errors
- `kafka_writer_write_seconds` - Write latency

### Consumer Metrics
- `kafka_reader_message_count` - Messages consumed
- `kafka_reader_error_count` - Consumer errors
- `kafka_reader_rebalance_count` - Group rebalances

---

## CI/CD Integration (Planned)

| Stage | Tests | Purpose |
|-------|-------|---------|
| PR | `smoke_single.js`, `smoke_concurrent.js` | "We didn't break Kafka semantics" |
| Main | `chaos_broker.js`, `objectstore_slow.js` | "We didn't break durability" |
| Release | `scaleout.js`, full suite | "We didn't break the architecture" |

---

## Current Status

### Working
- K6 with xk6-kafka extension
- Connection to KafScale (localhost:39092)
- Producer API - messages write successfully
- Topic auto-creation

### Known Issues
- Consumer groups timeout on fetch requests
- See [DOCS/TROUBLESHOOTING.md](DOCS/TROUBLESHOOTING.md) for details

---

## Documentation

- [SPEC/features.md](SPEC/features.md) - Functional requirements
- [SPEC/nfr.md](SPEC/nfr.md) - Non-functional requirements
- [SPEC/scenarios.md](SPEC/scenarios.md) - Test scenarios
- [DOCS/getting-started.md](DOCS/getting-started.md) - Detailed setup guide
- [DOCS/compatibility.md](DOCS/compatibility.md) - Compatibility charter and status
- [DOCS/release-checklist.md](DOCS/release-checklist.md) - Release readiness checklist
- [DOCS/test-reference.md](DOCS/test-reference.md) - Test documentation
- [DOCS/TROUBLESHOOTING.md](DOCS/TROUBLESHOOTING.md) - Known issues & fixes
- [DOCS/profiles.md](DOCS/profiles.md) - Execution profiles guide
- [CLAUDE.md](CLAUDE.md) - Developer guide

---

## Contributing

When adding tests:
1. Map to a requirement in `SPEC/features.md` or `SPEC/nfr.md`
2. Use UUID tracking for message verification
3. Enable `autoCreateTopic: true` on writers
4. Add binary assertions with `check()`
5. See [CLAUDE.md](CLAUDE.md) for code patterns and guidelines

---

## License

See [LICENSE](LICENSE) file.

---

## References

- [xk6-kafka](https://github.com/mostafa/xk6-kafka) - K6 Kafka extension
- [K6 Documentation](https://k6.io/docs/) - Load testing framework
- [MinIO](https://min.io/) - S3-compatible object storage
