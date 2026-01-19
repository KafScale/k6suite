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
 * smoke_multi_producer_single_consumer.js - Multi-Producer, Single-Consumer Smoke Test
 *
 * Scenario: S3 (Produce/Consume Correctness)
 * Features: F2.1, F2.2, F2.4, F3.4
 *
 * Purpose: Multiple producers write while a single consumer reads.
 *
 * Usage:
 *   K6_PROFILE=local-docker ./k6 run tests/k6/smoke_multi_producer_single_consumer.js
 *   K6_PROFILE=local-service ./k6 run tests/k6/smoke_multi_producer_single_consumer.js
 */

import { check, fail, sleep } from "k6";
import * as kafka from "k6/x/kafka";
import { getProfile } from "../../config/profiles.js";
import { ensureTopic } from "./lib/topics.js";

const config = getProfile();
const brokers = config.brokers;
const runId =
  __ENV.K6_RUN_ID || `${Date.now()}-${Math.floor(Math.random() * 1e6)}`;
const topic = `smoke-multi-prod-single-cons-${runId}`;

function uuidv4() {
  const rnd = () => Math.floor(Math.random() * 0xffffffff)
    .toString(16)
    .padStart(8, "0");
  return `${rnd()}-${rnd().slice(0, 4)}-4${rnd().slice(0, 3)}-a${rnd().slice(0, 3)}-${rnd()}${rnd().slice(0, 4)}`;
}

function stringToBytes(str) {
  const bytes = new Uint8Array(str.length);
  for (let i = 0; i < str.length; i++) {
    bytes[i] = str.charCodeAt(i);
  }
  return bytes;
}

function bytesToString(bytes) {
  return String.fromCharCode.apply(null, new Uint8Array(bytes));
}

export const options = {
  scenarios: {
    producer: {
      executor: "shared-iterations",
      vus: 5,
      iterations: 100,
      exec: "produceMessages",
      startTime: "0s",
    },
    consumer: {
      executor: "shared-iterations",
      vus: 1,
      iterations: 100,
      exec: "consumeMessages",
      startTime: "5s",
    },
  },
};

export function setup() {
  ensureTopic(brokers, topic, 1, 1);
}

export function produceMessages() {
  const writer = new kafka.Writer({
    brokers,
    topic,
    requiredAcks: 1,
    autoCreateTopic: true,
  });

  const id = uuidv4();
  const payload = JSON.stringify({
    uuid: id,
    ts: Date.now(),
  });

  try {
    writer.produce({
      messages: [
        {
          key: stringToBytes(id),
          value: stringToBytes(payload),
        },
      ],
    });
  } catch (e) {
    fail(`Producer error: ${e}`);
  } finally {
    writer.close();
  }

  sleep(0.01);
}

export function consumeMessages() {
  const reader = new kafka.Reader({
    brokers,
    topic,
    partition: 0,
    offset: 0,
    minBytes: 1,
    maxBytes: 10e6,
    maxWait: "1s",
  });

  let msgs = [];
  let consumeErr = null;
  let tries = 0;

  while (tries < 5 && msgs.length === 0) {
    try {
      msgs = reader.consume({ limit: 1, timeout: "10s" });
    } catch (e) {
      consumeErr = e;
    }
    tries++;
  }

  reader.close();

  if (consumeErr && msgs.length === 0) {
    fail(`Consumer error: ${consumeErr}`);
  }

  const ok = check(msgs, {
    "got message": (m) => m.length > 0,
    "message has content": (m) => {
      if (m.length === 0) return false;
      try {
        const msg = JSON.parse(bytesToString(m[0].value));
        return msg.uuid !== undefined;
      } catch (e) {
        return false;
      }
    },
  });
  if (!ok) {
    fail("Message validation failed");
  }

  sleep(0.01);
}
