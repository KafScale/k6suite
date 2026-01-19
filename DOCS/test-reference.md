# K6SUITE Test Reference

Complete documentation for all tests in the K6SUITE framework.

## Test Overview

| Test | File | Scenario | Purpose | Status |
|------|------|----------|---------|--------|
| Diagnostic | `diagnose.js` | S1 | Verify connectivity | Working |
| Metrics Smoke | `smoke_metrics.js` | S2 | Metrics endpoint health | New |
| Single Smoke | `smoke_single.js` | S3 | Single message test | Partial |
| Concurrent Smoke | `smoke_concurrent.js` | S3 | Multi-VU phased test | Partial |
| Shared Smoke | `smoke_shared.js` | S3 | Shared connection test | Partial |
| Multi-Producer Single-Consumer | `smoke_multi_producer_single_consumer.js` | S3 | Multi-producer single consumer | New |
| Topic Auto-Create | `smoke_topic_autocreate.js` | S3 | Auto-create topic and consume | New |
| Partition Consume | `smoke_consumer_group.js` | S3 | Direct partition consumption | New |
| Correctness | `produce_consume.js` | S3 | UUID validation at scale | Planned |
| Broker Chaos | `chaos_broker.js` | S4 | Survive broker kills | Planned |
| Storage Throttle | `objectstore_slow.js` | S5 | Survive slow object store | Planned |
| Scale-Out | `scaleout.js` | S6 | Transparent scaling | Planned |

---

## Scenario S1: Connectivity

### diagnose.js

**Purpose:** Validates basic Kafka protocol connectivity through KafScale.

**Features:** F1.1, F2.1, F2.2

**What It Tests:**
1. Can we create a connection to the broker?
2. Can we create a writer (producer)?
3. Can we create a reader (consumer)?

**Run Command:**
```bash
./k6 run tests/k6/diagnose.js
```

**Configuration:**
```javascript
export const options = {
  vus: 1,
  iterations: 1,
};
```

**Expected Output:**
```
✓ Connection created successfully
✓ Writer created successfully
✓ Reader created successfully
```

**Pass Criteria:** All three checks pass.

**Status:** Working

---

## Scenario S2: Observability

### smoke_metrics.js

**Purpose:** Validate broker metrics endpoint availability.

**What It Tests:**
1. Metrics endpoint returns HTTP 200
2. Payload is non-empty

**Target:** KafScale only (`K6_TARGET=kafscale`). Skipped when `K6_TARGET=kafka`.

**Run Command:**
```bash
./k6 run tests/k6/smoke_metrics.js
```

**Configuration:**
```javascript
export const options = {
  vus: 1,
  iterations: 1,
};
```

**Pass Criteria:** Metrics endpoint responds with non-empty payload.

**Status:** New

---

## Scenario S3: Produce/Consume Correctness

### smoke_single.js

**Purpose:** Minimal end-to-end test with one message.

**Features:** F2.1, F2.2, F3.4
**NFRs:** NFR2.1, NFR2.2

**What It Tests:**
1. Producer can write a single message
2. Consumer can read the message
3. Message arrives correctly

**Run Command:**
```bash
./k6 run tests/k6/smoke_single.js
```

**Configuration:**
```javascript
export const options = {
  vus: 1,
  iterations: 1,
};
```

**When to Use:** Start here. If this fails, more complex tests will also fail.

**Pass Criteria:** Message produced and consumed successfully.

**Status:** Partial (producer works, consumer has timeout issues)

---

### smoke_concurrent.js

**Purpose:** Multi-VU test with separate producer and consumer phases.

**Features:** F2.1, F2.2, F2.4, F2.5, F3.4, F3.5
**NFRs:** NFR2.1, NFR2.2, NFR2.3

**What It Tests:**
1. Multiple producers can write concurrently
2. Multiple consumers can read concurrently
3. UUID tracking across VUs
4. No message loss under concurrent load

**Run Command:**
```bash
./k6 run tests/k6/smoke_concurrent.js
```

**Configuration:**
```javascript
export const options = {
  scenarios: {
    producer: {
      executor: "shared-iterations",
      vus: 5,
      iterations: 100,
      exec: "produceMessages",
      startTime: "0s",
    },
    consumer: {
      executor: "shared-iterations",
      vus: 5,
      iterations: 100,
      exec: "consumeMessages",
      startTime: "5s",
    },
  },
};
```

**Flow:**
1. **Phase 1 (0-5s):** 5 producers write 100 messages each = 500 total
2. **Phase 2 (5s+):** 5 consumers read messages

**Key Difference from smoke_shared.js:** Creates new writer/reader per iteration.

**Pass Criteria:** Messages produced and consumed with valid UUIDs.

**Status:** Partial (producer works, consumer has timeout issues)

---

### smoke_shared.js

**Purpose:** Test with shared writer/reader instances across all VUs.

**Features:** F2.1, F2.2, F2.5, F3.4, F3.5
**NFRs:** NFR2.1, NFR2.2, NFR2.3, NFR2.5

**What It Tests:**
1. Connection reuse across VUs
2. Produce then immediately consume pattern
3. UUID matches between produce and consume
4. Message ordering within shared connection

**Run Command:**
```bash
./k6 run tests/k6/smoke_shared.js
```

**Configuration:**
```javascript
export const options = {
  vus: 5,
  iterations: 200,
};
```

**Key Difference from smoke_concurrent.js:** Single shared writer/reader, not per-iteration.

**Pass Criteria:** UUID in consumed message matches produced UUID.

**Status:** Partial (producer works, consumer has timeout issues)

---

### smoke_multi_producer_single_consumer.js

**Purpose:** Multiple producers write while a single consumer reads.

**Features:** F2.1, F2.2, F2.4, F3.4

**What It Tests:**
1. Multiple producers can write concurrently
2. Single consumer can read produced messages
3. UUID payloads are valid

**Run Command:**
```bash
./k6 run tests/k6/smoke_multi_producer_single_consumer.js
```

**Configuration:**
```javascript
export const options = {
  scenarios: {
    producer: {
      executor: "shared-iterations",
      vus: 5,
      iterations: 100,
      exec: "produceMessages",
      startTime: "0s",
    },
    consumer: {
      executor: "shared-iterations",
      vus: 1,
      iterations: 100,
      exec: "consumeMessages",
      startTime: "5s",
    },
  },
};
```

**Pass Criteria:** Messages are produced and a single consumer reads valid UUIDs.

**Status:** New

---

### smoke_topic_autocreate.js

**Purpose:** Validate auto-creation of a brand-new topic and immediate consume.

**Features:** F2.1, F2.2, F3.4

**What It Tests:**
1. Producer can auto-create a topic
2. Producer can write a message
3. Consumer can read the message

**Run Command:**
```bash
./k6 run tests/k6/smoke_topic_autocreate.js
```

**Configuration:**
```javascript
export const options = {
  vus: 1,
  iterations: 1,
};
```

**Pass Criteria:** Topic auto-creates and message is consumed.

**Status:** New

---

### smoke_consumer_group.js

**Purpose:** Validate direct partition consumption without consumer groups.

**Features:** F2.1, F2.2, F2.4, F3.4

**What It Tests:**
1. Partition reader can read produced messages
2. UUID payloads are valid

**Run Command:**
```bash
./k6 run tests/k6/smoke_consumer_group.js
```

**Configuration:**
```javascript
export const options = {
  vus: 1,
  iterations: 1,
};
```

**Pass Criteria:** Partition reader consumes expected messages.

**Status:** New

---

## Planned Tests

### produce_consume.js (S3)

**Purpose:** Validate message correctness at scale with UUID tracking.

**What It Will Test:**
1. 100k messages with unique UUIDs
2. Zero duplicates on consume
3. Zero missing messages
4. Order preservation within partitions

**Planned Configuration:**
```javascript
export const options = {
  scenarios: {
    producer: {
      executor: "shared-iterations",
      vus: 100,
      iterations: 1000,  // 100k total messages
    },
    consumer: {
      executor: "shared-iterations",
      vus: 100,
      iterations: 1000,
    },
  },
};
```

**Pass Criteria:**
```
produced_uuids == consumed_uuids
duplicate_count == 0
missing_count == 0
```

---

### chaos_broker.js (S4)

**Purpose:** Prove that brokers are truly disposable.

**What It Will Test:**
1. Continuous produce/consume traffic
2. Random broker kills
3. No client-visible errors
4. No message loss

**Chaos Actions:**
```bash
# Docker
docker kill kafscale-broker-2
sleep 10
docker start kafscale-broker-2

# Kubernetes
kubectl delete pod kafscale-broker-2
```

**Pass Criteria:**
```
errors == 0
uuid_loss == 0
rebalance_count == 0  # Clients should not notice
```

---

### objectstore_slow.js (S5)

**Purpose:** Test backpressure handling when object storage is slow.

**What It Will Test:**
1. Inject latency into MinIO/S3
2. Monitor produce latency increase
3. Verify no data corruption
4. Verify consumers catch up after throttle removed

**Throttle Injection:**
```bash
tc qdisc add dev eth0 root netem delay 200ms 50ms
```

**Pass Criteria:**
```
latency: increased (expected)
throughput: decreased (expected)
errors: 0
data_corruption: 0
```

---

### scaleout.js (S6)

**Purpose:** Prove transparent scaling without client impact.

**What It Will Test:**
1. Start with 1 broker under load
2. Scale to 3 brokers while traffic continues
3. No client reconnections
4. Throughput increases

**Scale Actions:**
```bash
docker-compose scale broker=3
# or
kubectl scale deployment kafscale --replicas=3
```

**Pass Criteria:**
```
client_reconnections: 0
metadata_refreshes: 0
throughput: increased
```

---

## Test Selection Guide

| Situation | Use This Test |
|-----------|---------------|
| First time setup | `diagnose.js` |
| Ops/metrics health | `smoke_metrics.js` |
| Basic validation | `smoke_single.js` |
| Concurrent behavior | `smoke_concurrent.js` |
| Connection reuse | `smoke_shared.js` |
| Multi-producer single consumer | `smoke_multi_producer_single_consumer.js` |
| Topic auto-create | `smoke_topic_autocreate.js` |
| Partition consumption | `smoke_consumer_group.js` |
| Full correctness | `produce_consume.js` |
| Resilience testing | `chaos_broker.js` |
| Storage behavior | `objectstore_slow.js` |
| Scaling behavior | `scaleout.js` |

---

## Metrics Reference

### Writer Metrics
| Metric | Type | Description |
|--------|------|-------------|
| `kafka_writer_message_count` | Counter | Messages produced |
| `kafka_writer_error_count` | Counter | Producer errors |
| `kafka_writer_write_seconds` | Histogram | Write latency |

### Reader Metrics
| Metric | Type | Description |
|--------|------|-------------|
| `kafka_reader_message_count` | Counter | Messages consumed |
| `kafka_reader_error_count` | Counter | Consumer errors |
| `kafka_reader_rebalance_count` | Counter | Group rebalances |
| `kafka_reader_lag` | Gauge | Consumer lag |
