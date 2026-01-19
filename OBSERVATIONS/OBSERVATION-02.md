# OBSERVATION-02: Coordinator Broker ID Missing from Metadata

## Short Answer

The client and KafScale disagree about the coordinator broker identity, and the client keeps retrying because it cannot reconcile the response with its metadata model.

## What the Log Is Saying (Step by Step)

1. **Client asks for the group coordinator**

`FindCoordinator v3` for group `smoke-group-20260116-162303`

KafScale replies: coordinator is **broker 1**.

2. **Client checks broker metadata**

Immediately after, the client reports:

```
broker replied that group ... has broker coordinator 1,
but did not reply with that broker in the broker list
```

This is the core problem: the coordinator broker ID is not present in Metadata.

3. **Client retries forever (correct behavior)**

The client assumes metadata is stale and retries:
- Refresh Metadata (Metadata v12)
- Re-issue FindCoordinator
- Gets same answer
- Fails again

## Why This Happens with KafScale

This is not a client bug. It is a semantic mismatch caused by KafScale’s stateless architecture.

In classic Kafka:
- Brokers have stable numeric IDs.
- Coordinator ID must appear in Metadata.
- Broker identity is a hard invariant.

In KafScale:
- Brokers are stateless endpoints.
- Coordinator identity can be logical/virtual.
- Broker IDs may be remapped or not advertised as expected.

So KafScale says: coordinator is broker 1, but broker 1 is missing from Metadata.

## Why This Shows Up Now

You are using:
- Franz-go (strict protocol validation)
- Consumer groups (coordinator is mandatory)
- Debug logging

Producer-only or partition-based consumers do not hit this path.

## What Is Not the Problem

- Not a timeout
- Not network
- Not MinIO
- Not etcd
- Not consumer lag
- Not k6 load

This is pure protocol-level incompatibility.

## Options

### Option 1 — Fix in KafScale (correct, long-term)

Ensure **any broker ID returned by FindCoordinator appears in Metadata**. Options:
- Advertise a stable virtual coordinator broker.
- Map coordinator ID to an existing advertised broker.
- Pin coordinator to a fixed broker ID that is always present.

### Option 2 — Client-side workaround (acceptable for tests)

Avoid consumer groups:
- Manual partition assignment
- No group.id
- No OffsetFetch

This bypasses coordinator logic and works for KafScale validation.

### Option 3 — Accept as Known Limitation

KafScale is Kafka-protocol compatible, not behavior-identical. Consumer group semantics are one of the hardest APIs to emulate without stateful brokers.

## Practical Recommendation (Now)

For test suites and tutorials:
1. Rule: **KafScale smoke tests use manual partition assignment**.
2. Split test types:
   - ✅ Producer-only
   - ✅ Stateless consumers (no group)
   - ⚠️ Group consumers = experimental / known limitation
3. Checklist item:
   - “Does this test require consumer groups?”
     - If yes → Kafka
     - If no → KafScale

## One-Sentence Diagnosis

The client is correct: KafScale returns a coordinator broker ID that is not present in the Metadata response, violating a core Kafka protocol invariant.
