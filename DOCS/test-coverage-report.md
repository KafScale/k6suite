# Kafscale Smoke Test Coverage Report

This report treats `/Users/kamir/GITHUB.kafscale/platform` as the system under test and evaluates the current K6SUITE smoke tests against the coverage claims and test expectations documented in that repository.

KafScale is Kafka-compatible but not a drop-in replacement. It targets a narrower operational scope (stateless brokers + object storage durability), so the smoke suite prioritizes those guarantees over full Kafka parity.

## Sources Reviewed

- Platform test expectations and make targets: `docs/development.md`
- Protocol coverage claims: `docs/protocol.md`
- Existing platform e2e harness and test suite: `test/e2e/`
- Current K6SUITE tests: `tests/k6/*.js`

## Current K6SUITE Coverage (And Challenges)

**Existing tests**
- `diagnose.js`: basic broker connection, writer, reader construction.
- `smoke_single.js`: single message produce/consume.
- `smoke_concurrent.js`: multi-VU produce then consume, UUID payload.
- `smoke_shared.js`: shared writer/reader produce/consume.

**Challenges / risks**
- Consumer timeouts are already noted in `DOCS/test-reference.md` (tests marked "Partial"). This means the suite is not yet reliable as a smoke gate.
- Most tests log errors without failing the run (exceptions are often swallowed or only printed), which can mask regressions.
- No coverage of topic management (CreateTopics/DeleteTopics), metadata discovery, or ApiVersions negotiation even though these are in the platform's "supported" list.
- No validation of consumer group functionality (JoinGroup/SyncGroup/Heartbeat/OffsetCommit/OffsetFetch), despite being core platform features and part of the claimed protocol support.
- No coverage of ops/admin APIs (`make test-ops-api`) or health/metrics endpoints, which are part of the platform's e2e posture.
- No durability smoke (restart and read-back) even though the platform positions S3 durability as a core property.
- Shared tests rely on auto-create and fixed group IDs, which increases flakiness and cross-run interference.

## Claimed Coverage in the Platform Repository

From `docs/development.md` and `docs/protocol.md`, the platform expects:
- Unit tests plus targeted e2e suites (`make test-produce-consume`, `make test-consumer-group`, `make test-ops-api`, `make test-multi-segment-durability`).
- Protocol support for core client workflows: Produce, Fetch, ListOffsets, Metadata, group membership, offset commit/fetch, topic management, and group listing.
- E2E validation against real clients (e.g., Franz-go) and broker process lifecycle.

Smoke coverage for K6SUITE should therefore verify the subset of these claims that most directly impact "can we safely deploy this build?".

## Required Smoke Coverage Going Forward

The smoke suite should be minimal but must reflect the platform's advertised surface area. The coverage below aligns with the protocol claims and e2e expectations and can be implemented with short-running K6 tests.

| Smoke Test | Platform Claim / Surface | Rationale |
|------------|---------------------------|-----------|
| Connectivity + ApiVersions/Metadata | Supported in `docs/protocol.md` | Validates basic protocol negotiation and broker discovery. |
| Topic lifecycle | CreateTopics/DeleteTopics | Ensures admin path works and auto-create is not the only viable path. |
| Produce + Fetch (single + small batch) | Produce/Fetch supported | Core data path must pass before anything else. |
| Consumer group commit/fetch | OffsetCommit/OffsetFetch | Verifies consumer tracking works, matching e2e expectations. |
| Group membership + rebalance smoke | JoinGroup/SyncGroup/Heartbeat/LeaveGroup | Ensures group coordination is functional. |
| ListOffsets / OffsetForLeaderEpoch | ListOffsets + safe recovery | Lightweight validation of offset APIs used by clients. |
| Ops/metrics health check | `make test-ops-api` + metrics docs | Confirms admin endpoints are reachable and broker is healthy. |
| Restart durability smoke | S3 durability claim | Prove minimal durability: produce, restart broker, consume. |

## Kafka Compatibility Mode (Lowest Common Denominator)

The suite can also run against Apache Kafka for baseline compatibility. This is not the primary target and should not be used as a release gate for KafScale. Use `K6_TARGET=kafka` together with the `kafka-local` profile to skip KafScale-only checks (for example, the broker metrics endpoint smoke test).

## Suggested Improvements to the Suite

- Add explicit checks/thresholds so a failed Kafka operation fails the test run.
- Use unique topic and group names per run to avoid cross-test interference.
- Add a small admin API test (topic create/delete, describe configs) to align with protocol claims.
- Add a short consumer-group test (commit + fetch + list groups) to match the platform's e2e coverage focus.
- Add a minimal restart durability smoke to validate the S3-backed design promise.
- Capture and report key metrics (latency, error counts) so smoke regressions are visible.

## Why This Is Not Necessarily a KafScale Failure

The observed `context deadline exceeded` errors do not automatically indicate a KafScale bug. Based on known behavior and prior context, likely causes include:

1. **Consumer group rebalancing pressure**
   - Group-based consumers are fragile in k6.
   - Multiple VUs joining/leaving causes constant rebalances.
2. **Auto-offset / group semantics mismatches**
   - KafScale does not support all Kafka group behaviors equally.
   - Rapid connect/disconnect patterns amplify gaps.
3. **k6/xk6-kafka limitations**
   - The Reader API is not designed for per-iteration group joins.
   - It is better suited for long-lived readers.

## Recommended Next Steps

1. Stabilize `smoke_single.js` and `smoke_concurrent.js` by failing fast on producer/consumer errors and using unique topics/groups.
2. Add new smoke tests for: topic lifecycle, group commit/fetch, and ops health.
3. Add a short restart durability smoke (single broker restart) once the base suite is stable.
