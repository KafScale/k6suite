# OBSERVATION-01: OffsetFetch Version Mismatch in Group Consumers

Summary: KafScale supports `OffsetFetch` response version **5** only. xk6-kafka requests version **1**, which is incompatible. The solution is to avoid consumer groups in KafScale verification tests and use direct partition consumption instead.

## Summary

The log line below indicates a Kafka protocol mismatch, not a KafScale runtime failure:

```
2026/01/16 14:25:28 handle request: offset fetch response version 1 not supported
```

This happens when a client (xk6-kafka) uses consumer groups and sends an `OffsetFetch` request expecting response version **1**, while KafScale only supports `OffsetFetch` response version **5**.

## Root Cause

- Consumer group readers trigger `OffsetFetch` automatically.
- xk6-kafka requests `OffsetFetch` response version 1.
- KafScale supports `OffsetFetch` response version 5 only (per `docs/protocol.md` in the KafScale repo).
- The broker rejects the response version, and the client later surfaces a generic timeout (`context deadline exceeded`).

## Why This Is Expected (And OK)

KafScale is Kafka-compatible but not a full Kafka replacement. Its design prioritizes:
- Produce/fetch correctness
- Stateless brokers
- Object-store durability

Consumer group semantics are not a primary target surface and can fail on older protocol versions.

## Recommended Action

Do not use consumer groups for KafScale smoke verification. Use direct partition consumption instead:

```javascript
const reader = new kafka.Reader({
  brokers,
  topic,
  partition: 0,
  offset: 0,
  maxWait: "10s",
});
```

This avoids `OffsetFetch` entirely and aligns with KafScale’s architectural contract.

## Recommendations

You have three viable options. Only one aligns with KafScale verification goals.

### Option 1 — Patch or fork xk6-kafka (not recommended)

**What it would take**
- Modify xk6-kafka to advertise higher `OffsetFetch` versions or force v5.
- Rebuild k6 with the fork.

**Why this is a bad idea**
- High maintenance burden; you now own a Kafka client fork.
- k6 is not meant to be a protocol-conformance testbed.
- Distracts from KafScale verification.

**Verdict:** Do not do this.

### Option 2 — Disable OffsetFetch in xk6-kafka (not feasible today)

In Kafka clients, `OffsetFetch` is invoked when:
- `groupId` is set
- `groupTopics` is set
- auto-commit is enabled
- group coordination is enabled

xk6-kafka does not expose flags to disable `OffsetFetch` while still using groups.

**Verdict:** Not feasible with current xk6-kafka.

### Option 3 — Avoid OffsetFetch entirely (correct approach)

Do not use consumer groups in KafScale verification tests. Instead:
- Consume by `topic` + `partition`
- Track offsets inside the test
- Treat Kafka as a log, not a coordination system

This aligns with stateless brokers, object-store-backed offsets, and documented scope.

**Recommended consumer**
```javascript
const reader = new kafka.Reader({
  brokers: ["localhost:39092"],
  topic: "smoke",
  partition: 0,
  offset: 0,
  maxWait: "5s",
});
```

**Do not use**
```javascript
groupId: "test-group"
groupTopics: ["smoke"]
```

### How to prove correctness without consumer groups

- Produce messages with UUIDs
- Track IDs in-memory
- Validate no duplicates, no missing IDs, and ordering per partition

Example:
```javascript
check(seen.size, {
  "all messages consumed": (n) => n === expectedCount,
});
```

This is more deterministic than group-based validation.

### Recommended policy

| Test Type | Consumer Style |
|-----------|----------------|
| Smoke | Partition-based |
| Produce/consume | Partition-based |
| Chaos | Partition-based |
| Object-store durability | Partition-based |
| Scale-out | Partition-based |
| Group semantics | Out of scope |

If group semantics are ever tested, treat them as a separate compatibility suite.
