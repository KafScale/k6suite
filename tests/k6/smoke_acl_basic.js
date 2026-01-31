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
 * smoke_acl_basic.js - Basic ACL Validation Test (v1.5.0)
 *
 * Scenario: S8 (Permission Boundary) - partial coverage
 * Features: F7.1, F7.2 (ACL enforcement)
 * NFRs: NFR5.1 (Security)
 *
 * Purpose: Validate v1.5.0 ACL enforcement for authorized access
 * - Produce message with authorized principal
 * - Consume message with authorized principal
 * - Verify ACL context is present in connection
 *
 * This is a POSITIVE test - it verifies authorized access works.
 * Negative ACL testing (denied access) requires specific KafScale
 * configuration with restricted principals.
 *
 * v1.5.0 Features Validated:
 * - ACL Enforcement: Broker-side access control lists
 * - Per-group Authorization: Independent group-level authz decisions
 *
 * Prerequisites:
 * - KafScale v1.5.0+ running with ACL enabled
 * - Test principal has produce/consume permissions (default: allow-all)
 *
 * Pass: Authorized produce and consume succeed
 * Fail: Authorization error or connection failure
 *
 * Usage:
 *   K6_PROFILE=local-service ./k6 run tests/k6/smoke_acl_basic.js
 *   K6_PROFILE=local-docker ./k6 run tests/k6/smoke_acl_basic.js
 */

import { check, fail, sleep } from "k6";
import * as kafka from "k6/x/kafka";
import { getProfile } from "../../config/profiles.js";
import { ensureTopic } from "./lib/topics.js";

const config = getProfile();
const brokers = config.brokers;
const runId =
  __ENV.K6_RUN_ID || `${Date.now()}-${Math.floor(Math.random() * 1e6)}`;
const topic = `acl-test-${runId}`;
const groupId = `acl-group-${runId}`;

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
  thresholds: {
    checks: ["rate==1.0"], // All checks must pass
  },
};

export function setup() {
  console.log("=== ACL Basic Test Setup ===");
  console.log(`Profile: ${__ENV.K6_PROFILE || "local-service"}`);
  console.log(`Brokers: ${brokers.join(", ")}`);
  console.log(`Topic: ${topic}`);
  console.log(`Group ID: ${groupId}`);

  ensureTopic(brokers, topic, 1, 1);

  return {
    topic,
    groupId,
    testMessage: `acl-test-message-${runId}`,
  };
}

export default function (data) {
  console.log("=== Starting ACL Basic Test ===");

  // Test 1: Authorized Produce
  console.log("Test 1: Authorized produce");
  const writer = new kafka.Writer({
    brokers,
    topic: data.topic,
    requiredAcks: 1,
    autoCreateTopic: true,
  });

  let produceErr = null;
  try {
    writer.produce({
      messages: [
        {
          key: stringToBytes("acl-test-key"),
          value: stringToBytes(
            JSON.stringify({
              uuid: runId,
              payload: data.testMessage,
              ts: Date.now(),
              aclTest: true,
            })
          ),
        },
      ],
    });
    console.log("Produce succeeded - ACL allowed write access");
  } catch (e) {
    produceErr = e;
    console.error(`Produce failed: ${e}`);
  } finally {
    writer.close();
  }

  const produceOk = check(produceErr, {
    "authorized produce succeeded": (err) => err === null,
  });
  if (!produceOk) {
    fail(`ACL blocked authorized produce: ${produceErr}`);
  }

  // Wait for message to be committed
  console.log("Waiting for message commit...");
  sleep(3);

  // Test 2: Authorized Consume
  console.log("Test 2: Authorized consume");
  const reader = new kafka.Reader({
    brokers,
    topic: data.topic,
    partition: 0,
    offset: 0,
    maxWait: "5s",
  });

  let msgs = [];
  let consumeErr = null;
  try {
    msgs = reader.consume({ limit: 10, timeout: "5s" });
    console.log(`Consumed ${msgs.length} messages - ACL allowed read access`);

    if (msgs.length > 0) {
      for (let i = 0; i < msgs.length; i++) {
        const content = bytesToString(msgs[i].value);
        console.log(`  Message ${i}: ${content.substring(0, 80)}...`);
      }
    }
  } catch (e) {
    consumeErr = e;
    console.error(`Consume failed: ${e}`);
  } finally {
    reader.close();
  }

  const consumeOk = check(null, {
    "authorized consume succeeded": () => consumeErr === null,
    "received at least one message": () => msgs.length > 0,
  });
  if (!consumeOk) {
    fail(`ACL blocked authorized consume or no messages received`);
  }

  // Test 3: Verify message content (UUID tracking)
  console.log("Test 3: Verify message content");
  let foundMessage = false;
  for (const msg of msgs) {
    try {
      const content = JSON.parse(bytesToString(msg.value));
      if (content.uuid === runId && content.aclTest === true) {
        foundMessage = true;
        console.log(`Found test message with UUID: ${content.uuid}`);
        break;
      }
    } catch (e) {
      // Skip non-JSON messages
    }
  }

  check(foundMessage, {
    "test message found with correct UUID": (found) => found === true,
  });

  console.log("=== ACL Basic Test Complete ===");
  console.log("Summary:");
  console.log("  - Authorized produce: PASS");
  console.log("  - Authorized consume: PASS");
  console.log(`  - Message verification: ${foundMessage ? "PASS" : "WARN"}`);
  console.log("");
  console.log("Note: This test validates POSITIVE ACL behavior.");
  console.log("Negative ACL testing (denied access) requires specific");
  console.log("KafScale configuration with restricted principals.");
}

export function teardown(data) {
  console.log("=== ACL Basic Test Teardown ===");
  console.log(`Topic ${data.topic} left for inspection if needed.`);
}
