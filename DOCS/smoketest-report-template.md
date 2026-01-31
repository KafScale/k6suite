# KafScale v{VERSION} Smoketest Report

| Field | Value |
|-------|-------|
| Version | v{VERSION} |
| Date | YYYY-MM-DD |
| Profile | local-service |
| Target | kafscale |
| Tester | {name} |
| K6 Binary | ./k6 |

---

## Summary

| Metric | Value |
|--------|-------|
| Total Tests | {N} |
| Passed | {N} |
| Failed | {N} |
| Pass Rate | {N}% |

**Overall Result:** PASS / FAIL

---

## Phase Results

### Phase 1: Pre-flight (Connectivity & Metrics)

| Test | Scenario | Status | Duration | Notes |
|------|----------|--------|----------|-------|
| diagnose.js | S1 | PASS/FAIL | Ns | |
| smoke_metrics.js | S2 | PASS/FAIL | Ns | |

### Phase 2: Core Functionality (Basic Produce/Consume)

| Test | Scenario | Status | Duration | Notes |
|------|----------|--------|----------|-------|
| smoke_single.js | S3 | PASS/FAIL | Ns | |
| smoke_topic_autocreate.js | S3 | PASS/FAIL | Ns | |

### Phase 3: Concurrency Tests

| Test | Scenario | Status | Duration | Notes |
|------|----------|--------|----------|-------|
| smoke_concurrent.js | S3 | PASS/FAIL | Ns | |
| smoke_shared.js | S3 | PASS/FAIL | Ns | |
| smoke_multi_producer_single_consumer.js | S3 | PASS/FAIL | Ns | |
| smoke_consumer_group.js | S3 | PASS/FAIL | Ns | |

### Phase 4: v1.5.0 Features

| Test | Scenario | Status | Duration | Notes |
|------|----------|--------|----------|-------|
| smoke_acl_basic.js | S8 | PASS/FAIL | Ns | ACL enforcement validation |

---

## v1.5.0 Feature Validation

| Feature | Test Coverage | Status | Notes |
|---------|---------------|--------|-------|
| ACL Enforcement | smoke_acl_basic.js | PASS/FAIL | Positive case only |
| PROXY Protocol | Not covered | N/A | Requires special setup |
| Per-group Authorization | smoke_acl_basic.js | Partial | Via group-based consume |
| Auth Denial Logging | Manual | N/A | Check broker logs |

---

## Environment Details

```
KafScale Version: v1.5.0
Broker Endpoint: {broker:port}
Object Storage: {minio/s3}
Test Machine: {os/arch}
```

---

## Known Issues

| Issue | Impact | Workaround |
|-------|--------|------------|
| | | |

---

## Test Output (JSON)

Attach or link to `smoketest-v1.5.0-results.json`:

```json
{
  "version": "v1.5.0",
  "timestamp": "YYYY-MM-DDTHH:MM:SSZ",
  "profile": "local-service",
  "target": "kafscale",
  "summary": {
    "total": N,
    "passed": N,
    "failed": N
  },
  "tests": [...]
}
```

---

## Sign-off

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Tester | | | |
| Reviewer | | | |
| Release Manager | | | |

---

## Appendix: How to Run

```bash
# Run full smoketest suite
./scripts/smoketest_v1.5.0.sh

# Run with specific profile
K6_PROFILE=local-docker ./scripts/smoketest_v1.5.0.sh

# Run via Makefile
make smoketest-v1.5.0

# Run individual ACL test
./k6 run tests/k6/smoke_acl_basic.js
```

---

## Appendix: Troubleshooting

If tests fail, check:

1. **Connectivity issues (Phase 1 fails)**
   - Is KafScale running? `docker-compose ps`
   - Correct port? Check profile in `config/profiles.json`
   - Network accessible? `nc -zv localhost 39092`

2. **Produce/Consume issues (Phase 2 fails)**
   - Check KafScale logs: `docker-compose logs kafscale`
   - Object storage accessible? Check MinIO/S3 connectivity

3. **ACL issues (Phase 4 fails)**
   - ACL enabled in KafScale config?
   - Principal has required permissions?
   - Check auth denial logs in broker

See `DOCS/TROUBLESHOOTING.md` for detailed guidance.
