# Functional Requirements (Features)

Features define what KafScale must do. Each feature is testable with a binary pass/fail result.

---

## F1: Single IP Access

**Claim:** Clients connect through one IP endpoint, not individual brokers.

| ID | Requirement |
|----|-------------|
| F1.1 | Clients connect using a single broker address |
| F1.2 | Clients do not receive individual broker IPs in metadata |
| F1.3 | Connection works with TLS enabled |
| F1.4 | Multiple concurrent connections are load-balanced |

**Validated by:** Scenario S1 (Connectivity), Scenario S2 (Connection Storm)

---

## F2: Kafka Protocol Compatibility

**Claim:** KafScale speaks standard Kafka protocol.

| ID | Requirement |
|----|-------------|
| F2.1 | Producers can write messages using Kafka protocol |
| F2.2 | Consumers can read messages using Kafka protocol |
| F2.3 | Topics can be created (manually or auto-create) |
| F2.4 | Consumer groups function correctly |
| F2.5 | Message ordering is preserved within partitions |

**Validated by:** Scenario S1 (Connectivity), Scenario S3 (Produce/Consume)

---

## F3: Message Durability

**Claim:** All messages are durably stored in object storage.

| ID | Requirement |
|----|-------------|
| F3.1 | Every produced message exists in object storage |
| F3.2 | Messages survive broker restarts |
| F3.3 | Messages survive broker termination |
| F3.4 | No message loss under normal operation |
| F3.5 | No message duplication under normal operation |

**Validated by:** Scenario S3 (Produce/Consume), Scenario S4 (Broker Chaos)

---

## F4: Stateless Brokers

**Claim:** Brokers hold no persistent state; they are disposable.

| ID | Requirement |
|----|-------------|
| F4.1 | Killing a broker causes no data loss |
| F4.2 | Killing a broker causes no client errors (beyond TCP retry) |
| F4.3 | Killing a broker causes no consumer group rebalance |
| F4.4 | Killing a broker causes no metadata refresh |
| F4.5 | New broker can serve traffic immediately |

**Validated by:** Scenario S4 (Broker Chaos), Scenario S7 (Rolling Upgrade)

---

## F5: Transparent Scaling

**Claim:** Brokers can be added without client awareness.

| ID | Requirement |
|----|-------------|
| F5.1 | Adding brokers requires no client reconnection |
| F5.2 | Adding brokers requires no client configuration change |
| F5.3 | Adding brokers causes no metadata change visible to clients |
| F5.4 | Traffic redistributes to new brokers automatically |
| F5.5 | Throughput increases when brokers are added |

**Validated by:** Scenario S6 (Scale-Out)

---

## F6: Object Storage Integration

**Claim:** Object storage (S3/MinIO/GCS) is the system of record.

| ID | Requirement |
|----|-------------|
| F6.1 | Data consumed via Kafka matches data in object storage |
| F6.2 | Object storage can be read directly to verify data |
| F6.3 | Slow object storage causes latency increase, not data loss |
| F6.4 | Object storage errors propagate as backpressure |

**Validated by:** Scenario S3 (Produce/Consume), Scenario S5 (Storage Throttle)

---

## F7: Access Control

**Claim:** Access control is enforced at the object storage layer.

| ID | Requirement |
|----|-------------|
| F7.1 | Clients cannot read outside their assigned scope |
| F7.2 | Clients cannot write outside their assigned scope |
| F7.3 | Unauthorized access fails cleanly with proper error |
| F7.4 | No cross-tenant data leakage |

**Validated by:** Scenario S8 (Permission Boundary)

---

## Feature Summary

| ID | Feature | Critical | Status |
|----|---------|----------|--------|
| F1 | Single IP Access | Yes | Partial |
| F2 | Kafka Protocol | Yes | Partial |
| F3 | Message Durability | Yes | Planned |
| F4 | Stateless Brokers | Yes | Planned |
| F5 | Transparent Scaling | Yes | Planned |
| F6 | Object Storage | Yes | Planned |
| F7 | Access Control | No | Planned |

**Status Legend:**
- Partial: Some requirements validated
- Planned: Tests not yet implemented
- Complete: All requirements validated
