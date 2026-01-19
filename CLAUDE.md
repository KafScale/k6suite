# CLAUDE.md - Developer Guide for K6SUITE

This file provides guidance for developers (human or AI) working on K6SUITE.

---

## Project Purpose

K6SUITE validates KafScale's architectural claims. It is NOT a benchmark.

**Benchmarks ask:** How fast?
**K6SUITE asks:** Does it work correctly?

---

## Quick Reference

### Run Tests
```bash
# Default profile (local-service on port 39092)
./k6 run tests/k6/diagnose.js         # S1: Connectivity
./k6 run tests/k6/smoke_metrics.js    # S2: Metrics endpoint
./k6 run tests/k6/smoke_single.js     # S3: Single message
./k6 run tests/k6/smoke_concurrent.js # S3: Multi-VU concurrent
./k6 run tests/k6/smoke_shared.js     # S3: Shared connection
./k6 run tests/k6/smoke_topic_autocreate.js  # S3: Auto-create topic
./k6 run tests/k6/smoke_multi_producer_single_consumer.js # S3: Multi-producer/single-consumer
./k6 run tests/k6/smoke_consumer_group.js    # S3: Direct partition consume

# With specific profile
K6_PROFILE=local-docker ./k6 run tests/k6/diagnose.js
K6_PROFILE=local-service ./k6 run tests/k6/smoke_single.js
```

### Execution Profiles
| Profile | Port | Use Case |
|---------|------|----------|
| `local-docker` | 9092 | Docker Compose environment |
| `local-service` | 39092 | Local native service (default) |
| `k8s-local` | 39092 | Local Kubernetes cluster |
| `kafka-local` | 9092 | Apache Kafka local |

### Start Environment
```bash
# Use your own KafScale + object storage environment
docker-compose up -d
```

### Build K6
```bash
xk6 build --with github.com/mostafa/xk6-kafka
```

---

## Project Structure

```
k6suite/
├── tests/k6/           # Test implementations
├── config/             # Configuration files
│   ├── profiles.js     # Execution profile loader
│   └── profiles.json   # Execution profiles
├── SPEC/               # Requirements (features.md, nfr.md, scenarios.md)
├── DOCS/               # User documentation
├── OBSERVATIONS/       # Protocol gaps and behavior notes
├── k6                  # Compiled k6 binary
├── CLAUDE.md           # This file (developer guidance)
└── DOCS/TROUBLESHOOTING.md  # Known issues
```

---

## Specifications

All tests trace to requirements in `SPEC/`:

| File | Contains |
|------|----------|
| `SPEC/features.md` | Functional requirements (F1-F7) |
| `SPEC/nfr.md` | Non-functional requirements (NFR1-NFR5) |
| `SPEC/scenarios.md` | Test scenarios (S1-S8) |

**Traceability:** Feature → Scenario → Test File → Pass/Fail

---

## Writing Tests

### 1. Map to Requirements

Every test must validate at least one requirement from `SPEC/features.md` or `SPEC/nfr.md`.

```javascript
// tests/k6/my_test.js
// Validates: F3.1 (every message exists in object storage)
// Scenario: S3 (Produce/Consume Correctness)
```

### 2. Use UUID Tracking

Every message must have a unique identifier:

```javascript
import { uuidv4 } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

const id = uuidv4();
const msg = JSON.stringify({
  uuid: id,
  payload: "test",
  ts: Date.now(),
});
```

### 3. Use Binary Assertions

Tests pass or fail. No "partial success":

```javascript
import { check, fail } from "k6";

const ok = check(messages, {
  "got message": (m) => m.length === 1,
  "uuid matches": (m) => JSON.parse(m[0].value).uuid === expectedId,
});
if (!ok) {
  fail("Message validation failed");
}
```

### 4. Clean Up Resources

Always close writers and readers:

```javascript
const writer = new kafka.Writer({ ... });
try {
  // test logic
} finally {
  writer.close();
}
```

### 5. Use Unique Topics and Group IDs

Avoid cross-run interference by generating per-run names:

```javascript
const runId =
  __ENV.K6_RUN_ID || `${Date.now()}-${Math.floor(Math.random() * 1e6)}`;
const topic = `smoke-${runId}`;
const groupId = `smoke-group-${runId}`;
```

---

## Code Patterns

### Using Profiles in Tests

```javascript
import { getProfile } from "../../config/profiles.js";

const config = getProfile();
const brokers = config.brokers;

// Access profile settings
// config.minio.endpoint, config.minio.bucket
// config.kafscale.metricsUrl
// config.tls.enabled
```

### Producer (Writer)

```javascript
import * as kafka from "k6/x/kafka";
import { getProfile } from "../../config/profiles.js";

const config = getProfile();
const writer = new kafka.Writer({
  brokers: config.brokers,
  topic: "test-topic",
  autoCreateTopic: true,
});

function produceMessage(uuid, payload) {
  const messages = [{
    key: kafka.encode({ codec: kafka.CODEC_STRING }, uuid),
    value: kafka.encode({ codec: kafka.CODEC_STRING }, JSON.stringify({
      uuid: uuid,
      payload: payload,
      ts: Date.now(),
    })),
  }];
  writer.produce({ messages });
}
```

### Consumer (Reader)

```javascript
import { getProfile } from "../../config/profiles.js";

const config = getProfile();
const reader = new kafka.Reader({
  brokers: config.brokers,
  groupId: "test-group",
  groupTopics: ["test-topic"],
  startOffset: kafka.OFFSET_BEGINNING,
});

const messages = reader.consume({ limit: 10 });
reader.close();
```

### Test Structure

```javascript
import { check, sleep } from "k6";
import * as kafka from "k6/x/kafka";
import { getProfile } from "../../config/profiles.js";

const config = getProfile();
const brokers = config.brokers;
const topic = "test-topic";

export const options = {
  vus: 5,
  iterations: 100,
};

export default function () {
  // Setup
  const writer = new kafka.Writer({ brokers, topic, autoCreateTopic: true });

  // Test
  // ...

  // Verify
  check(result, {
    "condition": (r) => r === expected,
  });

  // Cleanup
  writer.close();
}
```

---

## Current Status

### Working
- Connection to KafScale (localhost:39092)
- Producer API (messages write successfully)
- Topic auto-creation
- `diagnose.js` passes

### Known Issues
- **Consumer groups timeout** - xk6-kafka consumer may be incompatible with KafScale
- See `DOCS/TROUBLESHOOTING.md` for details

Note: `smoke_consumer_group.js` uses direct partition consumption and avoids group coordination.

### Priority
1. Fix consumer issue (blocks all S3+ scenarios)
2. Implement S3 (Produce/Consume with UUID verification)
3. Implement S4 (Broker Chaos)
4. Implement S5 (Storage Throttle)

---

## Design Principles

1. **Binary Results** - Pass or fail, no gray areas
2. **Requirement Traceability** - Every test maps to SPEC
3. **UUID Tracking** - Every message is uniquely identifiable
4. **Side-Channel Verification** - Verify object storage directly
5. **No Mocks** - Test real KafScale only

---

## What NOT to Do

- Do not add performance benchmarks (use KafScale benchmarks for that)
- Do not add tests without requirement mapping
- Do not use mocks or simulators
- Do not ignore consumer errors (they indicate real problems)
- Do not skip UUID tracking (it's how we prove correctness)

---

## File Naming

| File | Scenario | Purpose |
|------|----------|---------|
| `diagnose.js` | S1 | Connectivity validation |
| `smoke_single.js` | S3 | Single message produce/consume |
| `smoke_concurrent.js` | S3 | Multi-VU with separate phases |
| `smoke_shared.js` | S3 | Shared connection across VUs |
| `smoke_multi_producer_single_consumer.js` | S3 | Multi-producer/single-consumer |
| `chaos_broker.js` | S4 | Broker kill resilience (planned) |
| `objectstore_slow.js` | S5 | Storage throttle test (planned) |
| `scaleout.js` | S6 | Scale-out test (planned) |
| `rolling_upgrade.js` | S7 | Rolling upgrade test (planned) |
| `permission_boundary.js` | S8 | ACL/security test (planned) |

---

## CI/CD Mapping

| Stage | Scenarios | When |
|-------|-----------|------|
| PR | S1, S3 | Every pull request |
| Main | S4, S5 | After merge to main |
| Release | S2, S6, S7, S8 | Before release |

---

## Useful Commands

```bash
# Run single test (default profile: local-service)
./k6 run tests/k6/diagnose.js

# Run with specific profile
K6_PROFILE=local-docker ./k6 run tests/k6/diagnose.js
K6_PROFILE=local-service ./k6 run tests/k6/smoke_single.js
K6_PROFILE=k8s-local ./k6 run tests/k6/smoke_concurrent.js

# Run with more VUs
./k6 run --vus 10 tests/k6/smoke_concurrent.js

# Run with duration
./k6 run --duration 30s tests/k6/smoke_concurrent.js

# Output JSON results
./k6 run --out json=results.json tests/k6/smoke_concurrent.js

# Check Docker environment (if using a compose stack)
docker-compose ps

# View KafScale logs (if using a compose stack)
docker-compose logs kafscale

# Rebuild k6 with latest extension
xk6 build --with github.com/mostafa/xk6-kafka@latest
```

---

## Questions?

1. **What requirement does this test validate?** → Check `SPEC/features.md` and `SPEC/nfr.md`
2. **What scenario does this belong to?** → Check `SPEC/scenarios.md`
3. **How do I run this locally?** → See Quick Reference above
4. **Something is broken** → Check `DOCS/TROUBLESHOOTING.md`
