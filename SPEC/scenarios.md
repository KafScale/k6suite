# Test Scenarios

Scenarios group features and NFRs into executable tests. Each scenario answers a specific question about KafScale.

---

## S1: Connectivity

**Question:** Can clients connect to KafScale?

### Scope
| Features | NFRs |
|----------|------|
| F1.1, F1.2, F2.1, F2.2, F2.3 | NFR5.2 |

### Test Files
| File | Purpose |
|------|---------|
| `tests/k6/diagnose.js` | Connection, writer, reader creation |

### Procedure
1. Create connection to KafScale broker
2. Create a writer (producer)
3. Create a reader (consumer)
4. Verify no errors

### Pass Criteria
```
✓ Connection created successfully
✓ Writer created successfully
✓ Reader created successfully
```

### Status
**Working**

---

## S2: Connection Storm

**Question:** Can KafScale handle many concurrent connections?

### Scope
| Features | NFRs |
|----------|------|
| F1.1, F1.3, F1.4 | NFR3.4, NFR5.4 |

### Test File
`tests/k6/connection_storm.js` (planned)

### Procedure
1. Open 1,000+ concurrent Kafka connections
2. Randomly close and reopen connections
3. Measure connection latency and errors
4. Verify load distribution across brokers

### Pass Criteria
```
connection_errors: 0
handshake_failures: 0
load_distribution: balanced
```

### Status
**Planned**

---

## S3: Produce/Consume Correctness

**Question:** Does every message arrive correctly?

### Scope
| Features | NFRs |
|----------|------|
| F2.1, F2.2, F2.4, F2.5, F3.1, F3.4, F3.5, F6.1 | NFR2.1, NFR2.2, NFR2.3, NFR2.4, NFR5.1 |

### Test Files
| File | Purpose |
|------|---------|
| `tests/k6/smoke_single.js` | Single message, minimal validation |
| `tests/k6/smoke_concurrent.js` | Multi-VU with separate producer/consumer phases |
| `tests/k6/smoke_shared.js` | Shared connections, UUID verification |
| `tests/k6/produce_consume.js` | Full correctness test (planned) |

### Procedure
1. Produce N messages with unique UUIDs
2. Consume all messages
3. Verify UUID set matches exactly
4. (Optional) Verify UUIDs exist in object storage

### Pass Criteria
```
produced_uuids == consumed_uuids
duplicates: 0
missing: 0
```

### Variables
| Parameter | Smoke | Full |
|-----------|-------|------|
| Producers | 5 | 100 |
| Consumers | 5 | 100 |
| Messages | 500 | 100,000 |
| Partitions | 1 | 10 |

### Status
**Partial** (producer works, consumer has issues)

---

## S4: Broker Chaos

**Question:** Does data survive broker death?

### Scope
| Features | NFRs |
|----------|------|
| F3.2, F3.3, F4.1, F4.2, F4.3, F4.4, F4.5 | NFR1.1, NFR1.2, NFR1.3, NFR2.1, NFR2.2, NFR4.2, NFR4.4 |

### Test File
`tests/k6/chaos_broker.js` (planned)

### Procedure
1. Start continuous produce/consume traffic
2. Kill broker pod: `kubectl delete pod broker-N`
3. Continue traffic for 30 seconds
4. Restart broker
5. Verify zero message loss

### Pass Criteria
```
message_loss: 0
client_errors: 0
consumer_rebalances: 0
```

### Chaos Actions
```bash
# Docker
docker kill kafscale-broker-2
sleep 10
docker start kafscale-broker-2

# Kubernetes
kubectl delete pod kafscale-broker-2
```

### Status
**Planned**

---

## S5: Storage Throttle

**Question:** Does data survive object storage slowness?

### Scope
| Features | NFRs |
|----------|------|
| F6.3, F6.4 | NFR1.4, NFR1.5, NFR2.1, NFR2.2 |

### Test File
`tests/k6/objectstore_slow.js` (planned)

### Procedure
1. Start continuous produce/consume traffic
2. Inject 200ms latency into MinIO
3. Continue traffic for 60 seconds
4. Remove latency injection
5. Verify consumers catch up

### Pass Criteria
```
data_corruption: 0
message_loss: 0
recovery: complete
```

### Throttle Injection
```bash
# Linux tc
tc qdisc add dev eth0 root netem delay 200ms 50ms

# Remove
tc qdisc del dev eth0 root
```

### Status
**Planned**

---

## S6: Scale-Out

**Question:** Can we add brokers without client impact?

### Scope
| Features | NFRs |
|----------|------|
| F5.1, F5.2, F5.3, F5.4, F5.5 | NFR3.1, NFR3.2, NFR3.3, NFR4.3 |

### Test File
`tests/k6/scaleout.js` (planned)

### Procedure
1. Start with 1 broker
2. Begin continuous produce/consume traffic
3. Scale to 3 brokers: `docker-compose scale broker=3`
4. Continue traffic for 60 seconds
5. Verify no client reconnections

### Pass Criteria
```
client_reconnections: 0
metadata_changes: 0
throughput: increased
errors: 0
```

### Status
**Planned**

---

## S7: Rolling Upgrade

**Question:** Can we replace brokers without downtime?

### Scope
| Features | NFRs |
|----------|------|
| F4.1, F4.2, F4.5 | NFR4.1, NFR4.4 |

### Test File
`tests/k6/rolling_upgrade.js` (planned)

### Procedure
1. Start continuous produce/consume traffic
2. Restart brokers one by one: `kubectl rollout restart`
3. Verify zero errors during rollout
4. Verify all messages delivered

### Pass Criteria
```
produce_errors: 0
consume_errors: 0
reconnect_storms: 0
message_loss: 0
```

### Status
**Planned**

---

## S8: Permission Boundary

**Question:** Is access control enforced correctly?

### Scope
| Features | NFRs |
|----------|------|
| F7.1, F7.2, F7.3, F7.4 | - |

### Test File
`tests/k6/permission_boundary.js` (planned)

### Procedure
1. Create clients with different S3 credentials
2. Attempt to read/write outside assigned scope
3. Verify access denied with proper error
4. Verify no cross-tenant data leakage

### Pass Criteria
```
unauthorized_reads: blocked
unauthorized_writes: blocked
cross_tenant_leakage: 0
```

### Status
**Planned**

---

## Scenario Summary

| ID | Scenario | Test Files | Status | CI Stage |
|----|----------|------------|--------|----------|
| S1 | Connectivity | `diagnose.js` | Working | PR |
| S2 | Connection Storm | `connection_storm.js` | Planned | Release |
| S3 | Produce/Consume | `smoke_single.js`, `smoke_concurrent.js`, `smoke_shared.js` | Partial | PR |
| S4 | Broker Chaos | `chaos_broker.js` | Planned | Main |
| S5 | Storage Throttle | `objectstore_slow.js` | Planned | Main |
| S6 | Scale-Out | `scaleout.js` | Planned | Release |
| S7 | Rolling Upgrade | `rolling_upgrade.js` | Planned | Release |
| S8 | Permission Boundary | `permission_boundary.js` | Planned | Release |

---

## CI/CD Mapping

| Stage | Scenarios | Purpose |
|-------|-----------|---------|
| **PR** | S1, S3 | "We didn't break Kafka semantics" |
| **Main** | S4, S5 | "We didn't break durability" |
| **Release** | S2, S6, S7, S8 | "We didn't break the architecture" |
