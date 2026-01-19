# K6SUITE Architecture

Understanding how K6SUITE validates KafScale's architectural claims.

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         K6SUITE                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    K6 Test Runner                         │  │
│  │  • Orchestrates VUs (virtual users)                       │  │
│  │  • Collects metrics                                       │  │
│  │  • Runs scenarios                                         │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                  │
│                              ▼                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                   xk6-kafka Extension                     │  │
│  │  • Kafka protocol implementation                          │  │
│  │  • Producer/Consumer APIs                                 │  │
│  │  • Connection management                                  │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                               │
                               │ Kafka Protocol
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                        KafScale                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Single IP Entry Point (:39092)               │  │
│  │  • Load balances across stateless brokers                 │  │
│  │  • No client awareness of individual brokers              │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                  │
│                              ▼                                  │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                │
│  │  Broker 1  │  │  Broker 2  │  │  Broker N  │  (Stateless)   │
│  └────────────┘  └────────────┘  └────────────┘                │
│                              │                                  │
└──────────────────────────────│──────────────────────────────────┘
                               │
                               │ S3 API
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Object Storage (MinIO/S3/GCS)                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                   kafka-data bucket                       │  │
│  │  • Source of truth for all messages                       │  │
│  │  • Durable storage layer                                  │  │
│  │  • ACL-based access control                               │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Key Architectural Principles

### 1. Single IP Entry Point

Traditional Kafka requires clients to know about all brokers. KafScale provides one IP:

```
Classic Kafka:          KafScale:
┌─────────┐             ┌─────────┐
│ Client  │             │ Client  │
└────┬────┘             └────┬────┘
     │                       │
     ├──► Broker 1           │
     ├──► Broker 2           ▼
     └──► Broker 3      Single IP
                             │
                        ┌────┴────┐
                        │ Brokers │
                        └─────────┘
```

**What K6SUITE tests:** Clients never need to know about individual brokers.

### 2. Stateless Brokers

KafScale brokers hold no persistent state:

```
Classic Kafka:           KafScale:
┌──────────────┐         ┌──────────────┐
│   Broker     │         │   Broker     │
│ ┌──────────┐ │         │              │
│ │  State   │ │         │  (nothing)   │
│ │  Logs    │ │         │              │
│ │  Indexes │ │         │              │
│ └──────────┘ │         └──────────────┘
└──────────────┘                │
                                ▼
                         ┌──────────────┐
                         │Object Storage│
                         │  (all data)  │
                         └──────────────┘
```

**What K6SUITE tests:** Kill any broker, no data is lost.

### 3. Object Storage as Source of Truth

Every message exists in object storage:

```
Producer ──► KafScale ──► MinIO/S3
                             │
                             ▼
                         ┌───────────┐
                         │  Bucket   │
                         │ ┌───────┐ │
                         │ │ msg-1 │ │
                         │ │ msg-2 │ │
                         │ │ msg-n │ │
                         │ └───────┘ │
                         └───────────┘
                             │
Consumer ◄── KafScale ◄──────┘
```

**What K6SUITE tests:** Messages in object storage match messages via Kafka.

## Test Architecture Mapping

Each test validates a specific architectural claim:

| Test | Architecture Validated | Mechanism |
|------|----------------------|-----------|
| `diagnose.js` | Single IP works | Connect through one endpoint |
| `smoke_single.js` | Kafka semantics | Produce/consume through KafScale |
| `chaos_broker.js` | Stateless brokers | Kill brokers, no impact |
| `objectstore_slow.js` | Object storage durability | Slow storage, no data loss |
| `scaleout.js` | Transparent scaling | Add brokers, no reconnects |

## Three-Layer Validation

K6SUITE validates at three levels:

### Layer 1: Kafka Protocol
```javascript
// Can we speak Kafka through KafScale?
const writer = new kafka.Writer({ brokers: ["localhost:39092"] });
writer.produce({ topic: "test", messages: [...] });
```

### Layer 2: Data Consistency
```javascript
// Does consume return what produce sent?
const produced = producer.send({ uuid: "abc123" });
const consumed = consumer.read();
check(consumed, { "uuid matches": (m) => m.uuid === "abc123" });
```

### Layer 3: Ground Truth (Object Storage)
```javascript
// Is the data actually in object storage?
const s3Object = http.get(`${MINIO}/${BUCKET}/?prefix=${uuid}`);
check(s3Object, { "exists in object store": (r) => r.body.includes(uuid) });
```

## Side-Channel Validation

The most important tests don't just use Kafka - they verify object storage directly:

```
                    ┌─────────────┐
                    │   k6 Test   │
                    └──────┬──────┘
                           │
            ┌──────────────┼──────────────┐
            │              │              │
            ▼              ▼              ▼
       ┌────────┐    ┌──────────┐    ┌────────┐
       │Produce │    │ Consume  │    │ Verify │
       │via Kafka│    │via Kafka │    │via S3  │
       └────────┘    └──────────┘    └────────┘
            │              │              │
            └──────────────┴──────────────┘
                           │
                           ▼
                    All three agree?
                    ───────────────
                    YES → KafScale works
                    NO  → Architecture broken
```

## Chaos Testing Architecture

For broker chaos tests:

```
┌─────────────────────────────────────────────────────────────┐
│                      Chaos Test Flow                        │
│                                                             │
│  ┌──────────┐                                               │
│  │ k6 Test  │───────────────────────────────────────────┐   │
│  └──────────┘                                           │   │
│       │                                                 │   │
│       │ Continuous traffic                              │   │
│       ▼                                                 │   │
│  ┌──────────────────────────────────────────────────┐   │   │
│  │                  KafScale                         │   │   │
│  │  ┌────────┐  ┌────────┐  ┌────────┐              │   │   │
│  │  │Broker 1│  │Broker 2│  │Broker 3│              │   │   │
│  │  └────────┘  └───┬────┘  └────────┘              │   │   │
│  │                  │                                │   │   │
│  │                  │ ◄─── docker kill / kubectl    │   │   │
│  │                  │       delete pod               │   │   │
│  │                  ▼                                │   │   │
│  │              [DEAD]                               │   │   │
│  │                                                   │   │   │
│  │  Traffic redistributes, no errors                 │   │   │
│  └──────────────────────────────────────────────────┘   │   │
│                                                         │   │
│  Assert: errors == 0, messages_lost == 0                │   │
└─────────────────────────────────────────────────────────────┘
```

## CI/CD Pipeline Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    GitHub Actions Pipeline                  │
│                                                             │
│  PR Created                                                 │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────────────────────────────────────┐           │
│  │ pr.yml                                       │           │
│  │  • smoke_single.js                           │           │
│  │  • smoke_concurrent.js                       │           │
│  │  ─────────────────────                       │           │
│  │  "We didn't break Kafka semantics"           │           │
│  └─────────────────────────────────────────────┘           │
│       │                                                     │
│       │ Pass                                                │
│       ▼                                                     │
│  Merge to main                                              │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────────────────────────────────────┐           │
│  │ main.yml                                     │           │
│  │  • chaos_broker.js                           │           │
│  │  • objectstore_slow.js                       │           │
│  │  ─────────────────────                       │           │
│  │  "We didn't break durability"                │           │
│  └─────────────────────────────────────────────┘           │
│       │                                                     │
│       │ Pass                                                │
│       ▼                                                     │
│  Release Tag                                                │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────────────────────────────────────┐           │
│  │ release.yml                                  │           │
│  │  • scaleout.js                               │           │
│  │  • Full chaos suite                          │           │
│  │  • Generate reports                          │           │
│  │  ─────────────────────                       │           │
│  │  "We didn't break the architecture"          │           │
│  └─────────────────────────────────────────────┘           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Local Development Environment

```
┌─────────────────────────────────────────────────────────────┐
│                docker-compose.yml                           │
│                                                             │
│  ┌──────────────────┐    ┌──────────────────────────────┐  │
│  │      MinIO       │    │          KafScale            │  │
│  │                  │    │                              │  │
│  │  Port 9000: API  │◄───│  Port 39092: Kafka protocol  │  │
│  │  Port 9001: UI   │    │                              │  │
│  │                  │    │  S3_ENDPOINT: minio:9000     │  │
│  │  Bucket:         │    │  S3_BUCKET: kafka-data       │  │
│  │  kafka-data      │    │                              │  │
│  └──────────────────┘    └──────────────────────────────┘  │
│          ▲                           ▲                      │
│          │                           │                      │
│          │     Health Checks         │                      │
│          │                           │                      │
└──────────┼───────────────────────────┼──────────────────────┘
           │                           │
           │                           │
           │                           │
      ┌────┴────────────────────────────┴────┐
      │              ./k6 run                 │
      │           tests/k6/*.js               │
      └──────────────────────────────────────┘
```

## Why This Architecture Matters

Traditional "Kafka on S3" implementations fail because:
1. They keep broker state, just persist to S3
2. They require rebalancing when brokers change
3. They don't validate object storage is actually the source of truth

K6SUITE doesn't benchmark performance. It proves the architecture is correct.

When these tests pass, you have evidence that:
- Brokers can be killed without client impact
- Data durability comes from object storage, not brokers
- Scaling is truly transparent
- The single-IP claim is real

That's not benchmarking. That's architectural validation.
