/**
 * smoke_topic_autocreate.js - Topic Auto-Create Smoke Test
 *
 * Scenario: S3 (Produce/Consume Correctness)
 *
 * Purpose: Validate that a brand-new topic can be auto-created,
 * produced to, and consumed from in a single run.
 *
 * Usage:
 *   K6_PROFILE=local-docker ./k6 run tests/k6/smoke_topic_autocreate.js
 *   K6_PROFILE=local-service ./k6 run tests/k6/smoke_topic_autocreate.js
 */

import { check, fail, sleep } from "k6";
import * as kafka from "k6/x/kafka";
import { getProfile } from "../../config/profiles.js";

const config = getProfile();
const brokers = config.brokers;
const runId =
  __ENV.K6_RUN_ID || `${Date.now()}-${Math.floor(Math.random() * 1e6)}`;
const topic = `smoke-autocreate-${runId}`;

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

export default function () {
  const writer = new kafka.Writer({
    brokers,
    topic,
    requiredAcks: 1,
    autoCreateTopic: true,
  });

  const payload = `auto-${Date.now()}`;
  try {
    writer.produce({
      messages: [{ value: stringToBytes(payload) }],
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

  let msgs = [];
  let consumeErr = null;
  let tries = 0;
  while (tries < 3 && msgs.length === 0) {
    try {
      msgs = reader.consume({ limit: 5, timeout: "3s" });
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
    "received at least one message": (m) => m.length > 0,
    "payload matches": (m) => bytesToString(m[0].value) === payload,
  });
  if (!ok) {
    fail("Auto-create smoke validation failed");
  }
}
