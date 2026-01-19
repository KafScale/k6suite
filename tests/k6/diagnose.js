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
 * diagnose.js - Connection Diagnostic Test
 *
 * Scenario: S1 (Connectivity)
 * Features: F1.1, F2.1, F2.2
 *
 * Purpose: Validates basic connectivity to KafScale
 * - Can we create a connection?
 * - Can we create a writer (producer)?
 * - Can we create a reader (consumer)?
 *
 * Pass: All three resources created successfully
 * Fail: Any connection or resource creation fails
 *
 * Usage:
 *   K6_PROFILE=local-docker ./k6 run tests/k6/diagnose.js
 *   K6_PROFILE=local-service ./k6 run tests/k6/diagnose.js
 */

import * as kafka from "k6/x/kafka";
import { getProfile } from "../../config/profiles.js";
import { ensureTopic } from "./lib/topics.js";

const config = getProfile();
const brokers = config.brokers;
const topic = "diagnostic-test";

export const options = {
  vus: 1,
  iterations: 1,
};

export function setup() {
  ensureTopic(brokers, topic, 1, 1);
}

export default function () {
  console.log("=== KafScale Diagnostics ===");
  console.log(`Connecting to: ${brokers.join(", ")}`);

  try {
    // Try to create a connection
    const connection = new kafka.Connection({
      address: brokers[0],
    });

    console.log("✓ Connection created successfully");

    // Try to create a writer to see if we can connect
    const writer = new kafka.Writer({
      brokers,
      topic,
      requiredAcks: 1,
      autoCreateTopic: true,
    });

    console.log("✓ Writer created successfully");
    writer.close();

    // Try to create a reader
    const reader = new kafka.Reader({
      brokers,
      topic,
      partition: 0,
      offset: 0,
      maxWait: "1s",
    });

    console.log("✓ Reader created successfully");
    reader.close();

    console.log("\n=== All connections successful ===");
  } catch (e) {
    console.log(`✗ Connection failed: ${e}`);
  }
}
