/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/**
 * smoke_single.js - Single Message Produce/Consume Test
 *
 * Scenario: S3 (Produce/Consume Correctness)
 * Features: F2.1, F2.2, F3.4
 * NFRs: NFR2.1, NFR2.2
 *
 * Purpose: Minimal end-to-end test with one message
 * - Produce a single message
 * - Wait for commit
 * - Consume and verify
 *
 * Use this test to validate basic Kafka semantics work.
 * If this fails, more complex tests will also fail.
 *
 * Pass: Message produced and consumed successfully
 * Fail: Timeout, missing message, or consumer error
 *
 * Usage:
 *   K6_PROFILE=local-docker ./k6 run tests/k6/smoke_single.js
 *   K6_PROFILE=local-service ./k6 run tests/k6/smoke_single.js
 */

import { check, fail, sleep } from "k6";
import * as kafka from "k6/x/kafka";
import { getProfile } from "../../config/profiles.js";
import { ensureTopic } from "./lib/topics.js";

const config = getProfile();
const brokers = config.brokers;
const runId =
  __ENV.K6_RUN_ID || `${Date.now()}-${Math.floor(Math.random() * 1e6)}`;
const topic = `smoke-single-${runId}`;

// Helper to convert string to byte array (k6-compatible)
function stringToBytes(str) {
  const bytes = new Uint8Array(str.length);
  for (let i = 0; i < str.length; i++) {
    bytes[i] = str.charCodeAt(i);
  }
  return bytes;
}

// Helper to convert byte array to string
function bytesToString(bytes) {
  return String.fromCharCode.apply(null, new Uint8Array(bytes));
}

export const options = {
  vus: 1,
  iterations: 1,
};

export function setup() {
  ensureTopic(brokers, topic, 1, 1);
}

export default function () {
  console.log("=== Starting Simple Smoke Test ===");

  // Step 1: Produce a single message
  console.log("Step 1: Creating writer...");
  const writer = new kafka.Writer({
    brokers,
    topic,
    requiredAcks: 1,
    autoCreateTopic: true,
  });

  const message = "test-message-" + Date.now();
  console.log(`Step 2: Producing message: ${message}`);

  let produceErr = null;
  try {
    writer.produce({
      messages: [
        {
          value: stringToBytes(message),
        },
      ],
    });
    console.log("Step 3: Message produced successfully");
  } catch (e) {
    produceErr = e;
  } finally {
    writer.close();
  }
  if (produceErr) {
    fail(`Step 3: Producer error: ${produceErr}`);
  }

  // Wait a bit for message to be committed
  console.log("Step 4: Waiting 3 seconds for message to be committed...");
  sleep(3);

  // Step 2: Try to consume it
  console.log("Step 5: Creating reader...");
  const reader = new kafka.Reader({
    brokers,
    topic,
    partition: 0,
    offset: 0,
    maxWait: "5s",
  });

  console.log("Step 6: Attempting to consume messages...");
  let msgs = [];

  let consumeErr = null;
  try {
    msgs = reader.consume({ limit: 10, timeout: "5s" });
    console.log(`Step 7: Consumed ${msgs.length} messages`);

    if (msgs.length > 0) {
      for (let i = 0; i < msgs.length; i++) {
        const content = bytesToString(msgs[i].value);
        console.log(`  Message ${i}: ${content}`);
      }
    }
  } catch (e) {
    consumeErr = e;
  }

  reader.close();

  if (consumeErr) {
    fail(`Step 7: Consumer error: ${consumeErr}`);
  }

  const ok = check(msgs, {
    "received at least one message": (m) => m.length > 0,
  });
  if (!ok) {
    fail("No messages consumed");
  }

  console.log("=== Test Complete ===");
}
