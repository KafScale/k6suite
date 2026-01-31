# Release Checklist

This repo is the release gate for KafScale compatibility. A release is ready
only when the following are satisfied.

## Required Steps

1. **Environment**: KafScale and object storage are reachable for the target
   profile (`K6_PROFILE`).
2. **Tooling**: `k6` is built with the `xk6-kafka` extension.
3. **Smoke Coverage**: Run the core suite:
   - `tests/k6/diagnose.js`
   - `tests/k6/smoke_single.js`
   - `tests/k6/smoke_concurrent.js`
   - `tests/k6/smoke_shared.js`
   - `tests/k6/smoke_topic_autocreate.js`
   - `tests/k6/smoke_multi_producer_single_consumer.js`
   - `tests/k6/smoke_consumer_group.js`
   - `tests/k6/smoke_metrics.js` (KafScale targets only)
   - `tests/k6/smoke_acl_basic.js` (v1.5.0+ ACL validation)
4. **Transparency**: Any failures are documented in `OBSERVATIONS/` and
   reflected in `DOCS/test-coverage-report.md`.
5. **Acceptance Criteria**: All critical NFRs pass per
   `SPEC/nfr.md` (no message loss, zero client errors, 100% UUID match).
6. **Compatibility Charter**: Update `DOCS/compatibility.md` if scope or status
   changes.

## v1.5.0 Release Smoketest

For KafScale v1.5.0 releases, use the automated smoketest script:

```bash
# Run full v1.5.0 smoketest suite
./scripts/smoketest_v1.5.0.sh

# Or via Makefile
make smoketest-v1.5.0
```

This runs all tests in sequence and produces a JSON report. See
[v1.5.0 Smoketest Guide](smoketest-v1.5.0.md) for details.

**v1.5.0-specific validations:**
- ACL enforcement (positive case)
- Per-group authorization

**Document results** using `DOCS/smoketest-report-template.md`.

## Optional (But Recommended)

- Run the Kafka baseline profile (`K6_PROFILE=kafka-local`, `K6_TARGET=kafka`)
  for a lowest-common-denominator check.
