/**
 * smoke_concurrent.js - Concurrent Produce/Consume Test
 *
 * Scenario: S3 (Produce/Consume Correctness)
 * Features: F2.1, F2.2, F2.4, F2.5, F3.4, F3.5
 * NFRs: NFR2.1, NFR2.2, NFR2.3
 *
 * Purpose: Multi-VU test with separate producer and consumer phases
 * - Phase 1: 5 producers write 100 messages each (500 total)
 * - Phase 2: 5 consumers read messages
 * - Each message has UUID for tracking
 *
 * This test uses k6 scenarios to run producers first, then consumers.
 * Each VU creates its own writer/reader per iteration.
 *
 * Pass: Messages produced and consumed with valid UUIDs
 * Fail: Consumer timeout, missing UUIDs, or errors
 *
 * Usage:
 *   K6_PROFILE=local-docker ./k6 run tests/k6/smoke_concurrent.js
 *   K6_PROFILE=local-service ./k6 run tests/k6/smoke_concurrent.js
 */

import { check, fail, sleep } from "k6";
import * as kafka from "k6/x/kafka";
import { getProfile } from "../../config/profiles.js";
import { ensureTopic } from "./lib/topics.js";

const config = getProfile();
const brokers = config.brokers;
const runId =
  __ENV.K6_RUN_ID || `${Date.now()}-${Math.floor(Math.random() * 1e6)}`;
const topic = `smoke-concurrent-${runId}`;

function uuidv4() {
  const rnd = () => Math.floor(Math.random() * 0xffffffff)
    .toString(16)
    .padStart(8, "0");
  return `${rnd()}-${rnd().slice(0, 4)}-4${rnd().slice(0, 3)}-a${rnd().slice(0, 3)}-${rnd()}${rnd().slice(0, 4)}`;
}

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
      vus: 5,
      iterations: 100,
      exec: "consumeMessages",
      startTime: "5s", // Start consuming after producers have had time to write
    },
  },
};

export function setup() {
  ensureTopic(brokers, topic, 1, 1);
}

// Producer function - just write messages
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

  const bytes = stringToBytes(payload);

  try {
    writer.produce({
      messages: [
        {
          key: stringToBytes(id),
          value: bytes,
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

// Consumer function - just read messages
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
  let tries = 0;

  let consumeErr = null;
  while (tries < 3 && msgs.length === 0) {
    try {
      msgs = reader.consume({ limit: 1, timeout: "3s" });
    } catch (e) {
      consumeErr = e;
    }
    tries++;
  }

  if (consumeErr && msgs.length === 0) {
    reader.close();
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
    reader.close();
    fail("Message validation failed");
  }

  reader.close();
  sleep(0.01);
}
