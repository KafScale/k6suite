/**
 * smoke_consumer_group.js - Partition Consumption Smoke Test
 *
 * Scenario: S3 (Produce/Consume Correctness)
 *
 * Purpose: Validate direct partition consumption without
 * consumer-group coordination.
 *
 * Usage:
 *   K6_PROFILE=local-docker ./k6 run tests/k6/smoke_consumer_group.js
 *   K6_PROFILE=local-service ./k6 run tests/k6/smoke_consumer_group.js
 */

import { check, fail, sleep } from "k6";
import * as kafka from "k6/x/kafka";
import { getProfile } from "../../config/profiles.js";
import { ensureTopic } from "./lib/topics.js";

const config = getProfile();
const brokers = config.brokers;
const runId =
  __ENV.K6_RUN_ID || `${Date.now()}-${Math.floor(Math.random() * 1e6)}`;
const topic = `smoke-group-${runId}`;

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
  vus: 1,
  iterations: 1,
};

export function setup() {
  ensureTopic(brokers, topic, 1, 1);
}

export default function () {
  const writer = new kafka.Writer({
    brokers,
    topic,
    requiredAcks: 1,
    autoCreateTopic: true,
  });

  try {
    writer.produce({
      messages: [
        { value: stringToBytes(`group-${Date.now()}`) },
        { value: stringToBytes(`group-${Date.now()}-2`) },
      ],
    });
  } catch (e) {
    fail(`Producer error: ${e}`);
  } finally {
    writer.close();
  }

  sleep(2);

  const reader = new kafka.Reader({
    brokers,
    topic,
    partition: 0,
    offset: 0,
    maxWait: "5s",
  });

  const expected = 2;
  const seen = new Set();
  let consumeErr = null;
  let tries = 0;

  while (tries < 5 && seen.size < expected) {
    let msgs = [];
    try {
      msgs = reader.consume({ limit: 2, timeout: "5s" });
    } catch (e) {
      consumeErr = e;
    }
    for (let i = 0; i < msgs.length; i++) {
      seen.add(bytesToString(msgs[i].value));
    }
    tries++;
  }
  reader.close();

  if (consumeErr && seen.size === 0) {
    fail(`Consumer error: ${consumeErr}`);
  }

  const ok = check(seen, {
    "consumed expected messages": (s) => s.size >= expected,
  });
  if (!ok) {
    fail("Did not consume expected messages");
  }
}
