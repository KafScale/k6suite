# KafScale v1.5.0 Smoketest Guide

This guide explains how to run the v1.5.0 release smoketest suite for KafScale.

## Overview

The v1.5.0 smoketest validates KafScale release readiness across four phases:

| Phase | Focus | Tests |
|-------|-------|-------|
| 1 | Pre-flight | Connectivity, metrics endpoint |
| 2 | Core functionality | Single message produce/consume |
| 3 | Concurrency | Multi-VU, shared connections, consumer groups |
| 4 | v1.5.0 features | ACL enforcement |

## v1.5.0 Features Validated

The smoketest covers these v1.5.0-specific features:

- **ACL Enforcement** - Broker-side access control lists
- **Per-group Authorization** - Independent group-level authz decisions

Features requiring special setup (not covered by default):
- PROXY Protocol v1/v2 support
- Auth denial logging (verify manually via broker logs)

## Prerequisites

1. **k6 binary** built with xk6-kafka extension (see [Getting Started](getting-started.md))
2. **KafScale v1.5.0+** running and accessible
3. **Object storage** (MinIO/S3) running

## Running the Smoketest

### Option 1: Shell Script (Recommended)

Run the full smoketest suite with the orchestration script:

```bash
./scripts/smoketest_v1.5.0.sh
```

This runs all tests in sequence and produces:
- Colored pass/fail output in the terminal
- JSON results in `smoketest-v1.5.0-results.json`
- Summary table at completion

### Option 2: Makefile Target

```bash
make smoketest-v1.5.0
```

### Option 3: Individual Tests

Run specific tests manually:

```bash
# Run ACL test only
./k6 run tests/k6/smoke_acl_basic.js

# Via Makefile
make k6-smoke-acl-basic
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `K6_BIN` | `./k6` | Path to k6 binary |
| `K6_PROFILE` | `local-service` | Execution profile (see below) |
| `K6_TARGET` | `kafscale` | Target system |
| `REPORT_FILE` | `smoketest-v1.5.0-results.json` | JSON output file |

### Execution Profiles

| Profile | Broker Port | Use Case |
|---------|-------------|----------|
| `local-service` | 39092 | Local native KafScale (default) |
| `local-docker` | 9092 | Docker Compose environment |
| `k8s-local` | 39092 | Local Kubernetes cluster |

### Examples

```bash
# Run against Docker Compose environment
K6_PROFILE=local-docker ./scripts/smoketest_v1.5.0.sh

# Run against custom k6 binary
K6_BIN=/usr/local/bin/k6 ./scripts/smoketest_v1.5.0.sh

# Run with custom output file
REPORT_FILE=results/smoketest-$(date +%Y%m%d).json ./scripts/smoketest_v1.5.0.sh
```

## Understanding the Output

### Terminal Output

The script uses colored output:
- **[PASS]** (green) - Test succeeded
- **[FAIL]** (red) - Test failed
- **[INFO]** (blue) - Informational message

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All tests passed |
| 1 | One or more tests failed |
| 2 | Setup/configuration error |

### JSON Results

The script writes results to `smoketest-v1.5.0-results.json`:

```json
{
  "version": "v1.5.0",
  "timestamp": "2024-01-15T10:30:00Z",
  "profile": "local-service",
  "target": "kafscale",
  "summary": {
    "total": 9,
    "passed": 9,
    "failed": 0
  },
  "tests": [
    {"name": "S1: Connectivity (diagnose)", "result": "pass", "duration": "5s"},
    {"name": "S2: Metrics endpoint", "result": "pass", "duration": "3s"},
    ...
  ]
}
```

## Test Descriptions

### Phase 1: Pre-flight

| Test | File | Purpose |
|------|------|---------|
| S1: Connectivity | `diagnose.js` | Verify broker is reachable |
| S2: Metrics | `smoke_metrics.js` | Verify metrics endpoint |

**Critical:** If Phase 1 fails, the script aborts. Fix connectivity before proceeding.

### Phase 2: Core Functionality

| Test | File | Purpose |
|------|------|---------|
| S3: Single message | `smoke_single.js` | Basic produce/consume |
| S3: Topic auto-create | `smoke_topic_autocreate.js` | Topic creation |

### Phase 3: Concurrency

| Test | File | Purpose |
|------|------|---------|
| S3: Concurrent VUs | `smoke_concurrent.js` | Multi-VU parallel produce/consume |
| S3: Shared connection | `smoke_shared.js` | Shared writer/reader |
| S3: Multi-producer | `smoke_multi_producer_single_consumer.js` | Fan-in pattern |
| S3: Consumer group | `smoke_consumer_group.js` | Direct partition consume |

### Phase 4: v1.5.0 Features

| Test | File | Purpose |
|------|------|---------|
| S8: ACL basic | `smoke_acl_basic.js` | Authorized produce/consume |

**Note:** The ACL test validates positive behavior (authorized access works). Negative ACL testing (denied access) requires specific KafScale configuration.

## Documenting Results

Use the report template to document smoketest results:

1. Copy `DOCS/smoketest-report-template.md` to your results location
2. Fill in the test results from the JSON output
3. Add environment details and any issues encountered
4. Obtain sign-off for release approval

## Troubleshooting

### Phase 1 Fails (Connectivity)

```
[FAIL] S1: Connectivity (diagnose)
Pre-flight failed - connectivity issue. Aborting.
```

**Solutions:**
- Verify KafScale is running: `docker-compose ps` or check your deployment
- Check the port: `nc -zv localhost 39092`
- Verify profile matches your environment: `K6_PROFILE=local-docker`

### ACL Test Fails

```
[FAIL] S8: ACL basic (v1.5.0)
```

**Possible causes:**
- KafScale not configured with ACL enabled
- Principal lacks produce/consume permissions
- Check broker logs for auth denial messages

**Solutions:**
- Verify ACL is enabled in KafScale configuration
- Check that test principal has appropriate permissions
- Review `DOCS/TROUBLESHOOTING.md` for detailed guidance

### JSON Results Not Written

If `smoketest-v1.5.0-results.json` is not created:
- Check write permissions in current directory
- Specify alternate path: `REPORT_FILE=/tmp/results.json ./scripts/smoketest_v1.5.0.sh`

## CI/CD Integration

### GitHub Actions Example

```yaml
- name: Run v1.5.0 Smoketest
  run: |
    ./scripts/smoketest_v1.5.0.sh
  env:
    K6_PROFILE: local-docker

- name: Upload Results
  uses: actions/upload-artifact@v3
  with:
    name: smoketest-results
    path: smoketest-v1.5.0-results.json
```

### Jenkins Example

```groovy
stage('Smoketest') {
    steps {
        sh 'K6_PROFILE=local-docker ./scripts/smoketest_v1.5.0.sh'
    }
    post {
        always {
            archiveArtifacts artifacts: 'smoketest-v1.5.0-results.json'
        }
    }
}
```

## Related Documentation

- [Getting Started](getting-started.md) - Installation and setup
- [Test Reference](test-reference.md) - Complete test documentation
- [Execution Profiles](profiles.md) - Profile configuration
- [Release Checklist](release-checklist.md) - Full release process
- [Troubleshooting](TROUBLESHOOTING.md) - Common issues
