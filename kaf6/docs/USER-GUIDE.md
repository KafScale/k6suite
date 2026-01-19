# KAF6 User Guide

KAF6 is a Franz-go based, K6-style test runner for Kafka/KafScale.

## Build

```bash
cd kaf6
go build -o ../kaf6-runner ./cmd/kaf6
```

Makefile alternative:

```bash
cd kaf6
make build
```

## Run a Scenario

```bash
./kaf6-runner run kaf6/suite/smoke.json
```

Makefile alternative:

```bash
cd kaf6
make run
```

Note: `make run` uses `suite/smoke.json` relative to the `kaf6/` directory.

## Run the Suite

```bash
cd kaf6
make run-suite
```

`run-suite` performs a dry-run validation first, writes `suite/status.md` with SHA-256 hashes, then re-verifies the hashes before executing scenarios. If `profiles.json` is present, it runs all scenarios once per profile.

## Interactive Selection

```bash
cd kaf6
./kaf6-runner select suite
```

`select` shows a list of scenarios (with an "All scenarios" option) and a list of profiles (with an "All profiles" option). Use space to select one or more items, then run the suite selectively.

To keep the report even when scenarios fail without exiting non-zero:

```bash
KAF6_ALLOW_FAIL=1 ./kaf6-runner select suite
```

## K6 Scenario Selector

Run legacy k6 scripts with the same profile selector:

```bash
./kaf6-runner k6-select tests/k6
```

Set `K6_BIN` if your k6 binary is not at `./k6` or `../k6`.

Outputs:
- `reports/<run-id>/summary.json`
- `reports/<run-id>/report.html`
- `reports/<run-id>/report.json` (renderable data)

To re-render without re-running:

```bash
./kaf6-runner render-report reports/<run-id>/report.json
```

## Scenario Format (JSON)

Required fields:
- `brokers` or `profile` (profiles supply brokers)
- `scenarios.producer` and/or `scenarios.consumer`

Example: `kaf6/suite/smoke.json`

## Profiles

KAF6 uses a JSON profile registry instead of the old k6-style JS config.

Resolution order:
1. `profiles.json` in the suite directory (next to the scenario JSON).
2. `config/profiles.json` in the current working directory.

If `profile` is omitted in a scenario, the `default_profile` from the profile file is used.
If `brokers` is set in the scenario, it overrides the profile.

Example `profiles.json`:

```json
{
  "default_profile": "local-service",
  "profiles": {
    "local-service": {
      "name": "Local Service",
      "description": "KafScale running as a local service",
      "brokers": ["127.0.0.1:39092"],
      "metrics_url": "http://127.0.0.1:9093/metrics"
    }
  }
}
```

## Consumer Groups (Default)

KAF6 uses consumer groups by default to match real client behavior.
Set `scenarios.consumer.group.id` to control the group ID.

## Partition Mode (Opt-in)

Partition-based consumption can be added later for KafScale-specific verification.

## Notes

- KAF6 uses Franz-go for protocol correctness.
- OffsetFetch is handled at v5+ by Franz-go.
