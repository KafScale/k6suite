# KafScale K6 Test Suite - Troubleshooting

## Current Status

### ✅ Working
- K6 with xk6-kafka extension installed
- Connection to KafScale broker (localhost:39092)
- Writer API - messages are produced successfully
- Topic auto-creation

### ❌ Not Working
- Consumer group protocol with xk6-kafka (OffsetFetch v1 mismatch)
- Coordinator metadata invariants in group flows

## Test Results

### Diagnostic Test
```bash
./k6 run tests/k6/diagnose.js
```
**Status:** Expected to pass when KafScale is reachable.

### Simple Smoke Test
```bash
./k6 run tests/k6/smoke_single.js
```
**Status:** Expected to pass for direct partition consumption.

## Root Cause Analysis

### Producer Side
- `kafka_writer_message_count: 1` ✅
- `kafka_writer_error_count: 0` ✅
- Messages are successfully written to KafScale

### Consumer Side
- `kafka_reader_message_count: 0` ❌
- `kafka_reader_error_count: 2` ❌
- `kafka_reader_rebalance_count: 0` ℹ️
- Consumer groups timing out on fetch requests

## Possible Issues

### 1. KafScale Consumer Group Support
KafScale might not fully implement Kafka's consumer group protocol. Check:
- Does KafScale support consumer groups?
- Does it require specific consumer configuration?
- Is there a different API for consumption?

### 2. Object Storage Integration
- Are messages being written to object storage (MinIO/S3)?
- Is KafScale properly reading from object storage during consumer requests?
- Check KafScale logs for object storage errors

### 3. Topic/Partition Metadata
- Is topic metadata properly exposed to consumers?
- Are partition assignments working?
- Check: `kafka_reader_rebalance_count: 0` suggests no rebalancing happened

## Diagnostic Commands

### Check if KafScale is running
```bash
lsof -i :39092
```

### Check KafScale logs
```bash
# Add command to view your KafScale logs
docker logs kafscale-broker-1  # if running in Docker
```

### Test with standard Kafka tools
```bash
# Install kafkacat if not already installed
brew install kafkacat  # macOS
# or
apt-get install kafkacat  # Linux

# Produce a message
echo "test" | kafkacat -P -b localhost:39092 -t test-topic

# Try to consume
kafkacat -C -b localhost:39092 -t test-topic -o beginning -e
```

### Check object storage (MinIO)
```bash
# If using MinIO, check if messages are in the bucket
mc ls myminio/kafka-data/
```

## Next Steps

1. **Verify KafScale Configuration**
   - Check KafScale documentation for consumer group support
   - Verify object storage configuration
   - Review KafScale logs for errors

2. **Test Alternative Consumer Patterns**
   - Try partition-specific consumption (without consumer groups)
   - Test with different consumer configurations
   - Check if KafScale has its own consumer API

3. **Verify Object Storage**
   - Confirm messages are being written to MinIO/S3
   - Check if KafScale can read from object storage
   - Verify bucket permissions

4. **Contact KafScale Team**
   - Share test results showing producer works but consumer doesn't
   - Ask about consumer group protocol support
   - Request example consumer configuration

## Questions for KafScale Team

1. Does KafScale fully support Kafka's consumer group protocol?
2. Are there specific consumer configurations required?
3. How does KafScale handle offset management?
4. Are there known issues with xk6-kafka and KafScale?
5. Can you provide a working consumer example?

## Workarounds

If consumer groups don't work, consider:
- Using partition-specific reads (if supported)
- Directly reading from object storage to verify data
- Testing with official Kafka consumer libraries
- Implementing custom consumer logic

## Files

- [diagnose.js](tests/k6/diagnose.js) - Basic connectivity test
- [smoke_single.js](tests/k6/smoke_single.js) - Simple produce/consume test
- [smoke_concurrent.js](tests/k6/smoke_concurrent.js) - Full smoke test
