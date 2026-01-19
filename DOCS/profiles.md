# Execution Profiles

K6SUITE supports multiple execution profiles to run tests against different environments.

## Available Profiles

| Profile | Port | Description |
|---------|------|-------------|
| `local-docker` | 9092 | KafScale in Docker via docker-compose (external) |
| `local-service` | 39092 | KafScale as local service (default) |
| `k8s-local` | 39092 | KafScale in local Kubernetes |
| `kafka-local` | 9092 | Apache Kafka running locally |

## Usage

Set the `K6_PROFILE` environment variable before running tests:

```bash
# Run against Docker environment (port 9092)
K6_PROFILE=local-docker ./k6 run tests/k6/diagnose.js

# Run against local service (port 39092) - this is the default
K6_PROFILE=local-service ./k6 run tests/k6/diagnose.js

# Run against Kubernetes
K6_PROFILE=k8s-local ./k6 run tests/k6/diagnose.js
```

If no profile is specified, `local-service` is used by default.

## Profile Details

### local-docker

For testing against KafScale running in Docker Compose. This repo does not
include a compose stack; use your own stack or the KafScale repo.

```bash
# Start your compose stack
docker-compose up -d

# Run tests
K6_PROFILE=local-docker ./k6 run tests/k6/diagnose.js
```

**Configuration:**
- Brokers: `localhost:9092`
- MinIO: `http://localhost:9000`
- Bucket: `kafka-data`

### local-service

For testing against KafScale running as a native service or external process.

```bash
# Assuming KafScale is already running on port 39092
K6_PROFILE=local-service ./k6 run tests/k6/diagnose.js
```

**Configuration:**
- Brokers: `localhost:39092`
- MinIO: `http://localhost:9000`
- Bucket: `kafka-data`

### k8s-local

For testing against KafScale in a local Kubernetes cluster (kind, minikube).

```bash
# Port-forward KafScale
kubectl port-forward svc/kafscale 39092:9092 -n kafscale-demo &

# Run tests
K6_PROFILE=k8s-local ./k6 run tests/k6/diagnose.js
```

### kafka-local

For testing against Apache Kafka on localhost:9092.

```bash
K6_PROFILE=kafka-local K6_TARGET=kafka ./k6 run tests/k6/smoke_single.js
```

## Creating Custom Profiles

Edit `config/profiles.json` to add new profiles:

```json
{
  "profiles": {
    "my-custom-profile": {
      "name": "My Custom Environment",
      "description": "Description of this environment",
      "brokers": ["custom-host:9092"],
      "kafscale": {
        "metricsUrl": "https://custom-host/metrics"
      }
    }
  }
}
```

## Profile Configuration Structure

Each profile contains:

```json
{
  "name": "Human-readable name",
  "description": "Description",
  "brokers": ["host:port"],
  "kafscale": {
    "metricsUrl": "http://..."
  }
}
```

## Using Profiles in Tests

Tests import the profile configuration:

```javascript
import { getProfile } from "../../config/profiles.js";

const config = getProfile();
const brokers = config.brokers;

// Use brokers in writer/reader
const writer = new kafka.Writer({
  brokers,
  topic: "my-topic",
});
```

## Convenience Scripts

You can create shell aliases or scripts for common profiles:

```bash
# ~/.bashrc or ~/.zshrc
alias k6-docker='K6_PROFILE=local-docker ./k6'
alias k6-local='K6_PROFILE=local-service ./k6'
alias k6-k8s='K6_PROFILE=k8s-local ./k6'

# Usage
k6-docker run tests/k6/diagnose.js
k6-local run tests/k6/smoke_single.js
```

Or create a wrapper script:

```bash
#!/bin/bash
# scripts/run-tests.sh
PROFILE=${1:-local-service}
TEST=${2:-tests/k6/diagnose.js}

echo "Running $TEST with profile: $PROFILE"
K6_PROFILE=$PROFILE ./k6 run $TEST
```

## CI/CD Integration

In CI/CD pipelines, set the profile based on the target environment:

```yaml
# GitHub Actions example
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        profile: [local-docker, k8s-local]
    steps:
      - name: Run tests
        env:
          K6_PROFILE: ${{ matrix.profile }}
        run: ./k6 run tests/k6/smoke_single.js
```
