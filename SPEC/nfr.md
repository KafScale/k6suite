# Non-Functional Requirements (NFRs)

NFRs define how well KafScale must perform. These are quality attributes, not features.

---

## NFR1: Resilience

**Claim:** KafScale continues operating under failure conditions.

| ID | Requirement | Threshold |
|----|-------------|-----------|
| NFR1.1 | Broker failure causes zero message loss | 0 messages lost |
| NFR1.2 | Broker failure causes zero client errors | 0 errors (beyond TCP retry) |
| NFR1.3 | Recovery time after broker kill | < 5 seconds |
| NFR1.4 | Object storage slowdown causes no data corruption | 0 corrupted messages |
| NFR1.5 | System recovers after object storage returns to normal | Full recovery |

**Validated by:** Scenario S4 (Broker Chaos), Scenario S5 (Storage Throttle)

---

## NFR2: Consistency

**Claim:** Data integrity is maintained at all times.

| ID | Requirement | Threshold |
|----|-------------|-----------|
| NFR2.1 | No duplicate messages under normal operation | 0 duplicates |
| NFR2.2 | No missing messages under normal operation | 0 missing |
| NFR2.3 | UUID tracking matches across produce/consume | 100% match |
| NFR2.4 | Kafka data matches object storage data | 100% match |
| NFR2.5 | Message ordering preserved within partition | Strict ordering |

**Validated by:** Scenario S3 (Produce/Consume), Scenario S4 (Broker Chaos)

---

## NFR3: Scalability

**Claim:** KafScale scales horizontally without degradation.

| ID | Requirement | Threshold |
|----|-------------|-----------|
| NFR3.1 | Scale from 1 to N brokers with zero downtime | 0 errors during scale |
| NFR3.2 | Throughput increases with broker count | Linear or better |
| NFR3.3 | No client reconnections during scale-out | 0 reconnections |
| NFR3.4 | Handle 10k+ concurrent connections | No connection drops |
| NFR3.5 | Handle 100k+ messages in single test run | No degradation |

**Validated by:** Scenario S2 (Connection Storm), Scenario S6 (Scale-Out)

---

## NFR4: Availability

**Claim:** KafScale remains available during operations.

| ID | Requirement | Threshold |
|----|-------------|-----------|
| NFR4.1 | Zero downtime during rolling upgrade | 0 errors |
| NFR4.2 | Zero downtime during broker replacement | 0 errors |
| NFR4.3 | Zero downtime during scale-out | 0 errors |
| NFR4.4 | Continuous traffic during chaos events | No interruption |

**Validated by:** Scenario S4 (Broker Chaos), Scenario S6 (Scale-Out), Scenario S7 (Rolling Upgrade)

---

## NFR5: Observability

**Claim:** System state is measurable and verifiable.

| ID | Requirement | Threshold |
|----|-------------|-----------|
| NFR5.1 | Message count is trackable via UUID | 100% trackable |
| NFR5.2 | Errors are countable and classifiable | All errors captured |
| NFR5.3 | Object storage state is verifiable | Direct S3 verification |
| NFR5.4 | Connection distribution is measurable | Per-broker metrics |

**Validated by:** All scenarios (built into test framework)

---

## NFR Summary

| ID | NFR | Critical | Status |
|----|-----|----------|--------|
| NFR1 | Resilience | Yes | Planned |
| NFR2 | Consistency | Yes | Partial |
| NFR3 | Scalability | Yes | Planned |
| NFR4 | Availability | Yes | Planned |
| NFR5 | Observability | No | Partial |

---

## Acceptance Criteria

A KafScale release is considered validated when:

1. **All critical NFRs pass** with thresholds met
2. **Zero message loss** across all chaos scenarios
3. **Zero client errors** during scale/upgrade operations
4. **100% UUID match** between produce and consume

These are not negotiable. If any fail, the release is not ready.
