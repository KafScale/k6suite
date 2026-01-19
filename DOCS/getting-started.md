# Getting Started with K6SUITE

This guide walks you through setting up K6SUITE and running your first tests against KafScale.

## Prerequisites

### Required Software

| Software | Version | Purpose |
|----------|---------|---------|
| Go | 1.21+ | Building k6 with extensions |
| Docker | 20.10+ | Running local environment |
| Docker Compose | 2.0+ | Orchestrating services |

### Optional

- `kafkacat` / `kcat` - CLI tool for Kafka debugging
- MinIO Client (`mc`) - For inspecting object storage

## Installation

### Step 1: Clone the Repository

```bash
git clone https://github.com/your-org/k6suite.git
cd k6suite
```

### Step 2: Build K6 with Kafka Extension

```bash
# Install xk6 (k6 extension builder)
go install go.k6.io/xk6/cmd/xk6@latest

# Build k6 with xk6-kafka extension
xk6 build --with github.com/mostafa/xk6-kafka

# Verify the binary was created
./k6 version
```

The compiled `k6` binary with Kafka support will be in your current directory.

### Step 3: Start or Connect to a KafScale Environment

K6SUITE expects KafScale and object storage to be running. This repo does not
ship a Docker Compose stack, so use your existing environment or bring up your
own compose stack.

If you have a compose stack:

```bash
docker-compose up -d
```

Typical local setup:
- **MinIO** on ports 9000 (API) and 9001 (console)
- **KafScale** on port 39092

### Step 4: Verify Services are Running

```bash
# If using Docker Compose, check containers
docker-compose ps

# Check MinIO console (optional)
open http://localhost:9001
# Login: minioadmin / minioadmin

# Check KafScale is listening
nc -zv localhost 39092
```

## Running Your First Test

### Diagnostic Test

Start with the diagnostic test to verify connectivity:

```bash
./k6 run tests/k6/diagnose.js
```

**Expected output:**
```
✓ Connection created successfully
✓ Writer created successfully
✓ Reader created successfully

checks.........................: 100.00% ✓ 3  ✗ 0
```

If all checks pass, KafScale is accessible.

### Simple Smoke Test

Run the basic produce/consume test:

```bash
./k6 run tests/k6/smoke_single.js
```

**What this tests:**
- Create a Kafka producer
- Write a message with UUID
- Create a Kafka consumer
- Attempt to read the message back

**Current status:** Producer works, consumer has known issues (see [TROUBLESHOOTING.md](TROUBLESHOOTING.md))

### Full Smoke Test

Run the complete smoke test with multiple VUs:

```bash
./k6 run tests/k6/smoke_concurrent.js
```

**Configuration:**
- 5 producer VUs, 100 iterations each
- 5 consumer VUs, 100 iterations each
- Staggered start (consumers start after producers)

## Customizing Test Runs

### Change Virtual Users (VUs)

```bash
./k6 run --vus 10 tests/k6/smoke_concurrent.js
```

### Change Iterations

```bash
./k6 run --iterations 500 tests/k6/smoke_concurrent.js
```

### Use Duration Instead

```bash
./k6 run --duration 30s tests/k6/smoke_concurrent.js
```

### Export Results to JSON

```bash
./k6 run --out json=results.json tests/k6/smoke_concurrent.js
```

### Verbose Output

```bash
./k6 run --verbose tests/k6/smoke_concurrent.js
```

## Understanding Test Output

K6 provides detailed metrics after each run:

### Key Metrics

```
kafka_writer_message_count.....: 1000   # Messages produced
kafka_writer_error_count.......: 0      # Producer errors
kafka_reader_message_count.....: 1000   # Messages consumed
kafka_reader_error_count.......: 0      # Consumer errors
```

### Success Criteria

- `kafka_writer_error_count: 0` - No producer errors
- `kafka_reader_error_count: 0` - No consumer errors
- Message counts match between producer and consumer

## Connecting to Your Own KafScale

To test against a different KafScale instance:

### 1. Edit the Broker Address

In your test file, change:

```javascript
const brokers = ["localhost:39092"];
```

to:

```javascript
const brokers = ["your-kafscale-host:9092"];
```

### 2. Handle TLS (if needed)

```javascript
const kafka = new Kafka({
  brokers: ["kafscale.example.com:9092"],
  tls: {
    enabled: true,
  },
});
```

### 3. Add Authentication (if needed)

```javascript
const kafka = new Kafka({
  brokers: ["kafscale.example.com:9092"],
  sasl: {
    mechanism: "plain",
    username: "user",
    password: "password",
  },
});
```

## Next Steps

1. **Read the test reference** - [test-reference.md](test-reference.md)
2. **Understand the architecture** - [architecture.md](architecture.md)
3. **Check specs for advanced tests** - [SPEC/](../SPEC/)
4. **Troubleshoot issues** - [TROUBLESHOOTING.md](TROUBLESHOOTING.md)

## Common Issues

### "Connection refused" on port 39092

KafScale is not running or is bound to a different host/port. Start your
KafScale environment or update `K6_PROFILE` to match your deployment.

### "xk6: command not found"

Install xk6:

```bash
go install go.k6.io/xk6/cmd/xk6@latest
```

Make sure `$GOPATH/bin` is in your PATH.

### Tests timeout

Check that KafScale and MinIO are healthy:

```bash
# If using Docker Compose
docker-compose ps
docker-compose logs kafscale
```
