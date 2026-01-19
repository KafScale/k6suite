/**
 * smoke_shared.js - Shared Connection Produce/Consume Test
 *
 * Scenario: S3 (Produce/Consume Correctness)
 * Features: F2.1, F2.2, F2.5, F3.4, F3.5
 * NFRs: NFR2.1, NFR2.2, NFR2.3, NFR2.5
 *
 * Purpose: Test with shared writer/reader across all VUs
 * - Single writer instance shared by all VUs
 * - Single reader instance shared by all VUs
 * - Each iteration produces then immediately consumes
 * - UUID tracking for message verification
 *
 * This approach tests connection reuse and ordering within
 * a shared connection. Different from smoke_concurrent.js
 * which creates new connections per iteration.
 *
 * Pass: UUID in consumed message matches produced UUID
 * Fail: UUID mismatch, timeout, or errors
 *
 * Usage:
 *   K6_PROFILE=local-docker ./k6 run tests/k6/smoke_shared.js
 *   K6_PROFILE=local-service ./k6 run tests/k6/smoke_shared.js
 */

import { check, fail, sleep } from "k6";
import * as kafka from "k6/x/kafka";
import { getProfile } from "../../config/profiles.js";
import { ensureTopic } from "./lib/topics.js";

const config = getProfile();
const brokers = config.brokers;
const runId =
  __ENV.K6_RUN_ID || `${Date.now()}-${Math.floor(Math.random() * 1e6)}`;
const topic = `smoke-shared-${runId}`;

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

let writer;
let reader;

export const options = {
  vus: 1,
  iterations: 200,
};

export function setup() {
  ensureTopic(brokers, topic, 1, 1);
}

export default function () {
  if (!writer) {
    writer = new kafka.Writer({
      brokers,
      topic,
      requiredAcks: 1,
      autoCreateTopic: true,
    });
  }
  if (!reader) {
    reader = new kafka.Reader({
      brokers,
      topic,
      partition: 0,
      offset: 0,
      minBytes: 1,
      maxBytes: 10e6,
      maxWait: "5s",
    });
  }
  const id = uuidv4();

  const payload = JSON.stringify({
    uuid: id,
    ts: Date.now(),
  });

  // PRODUCE
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
  }

  // CONSUME (retry-safe)
  let msgs = [];
  let tries = 0;

  let consumeErr = null;
  while (tries < 5 && msgs.length === 0) {
    try {
      msgs = reader.consume({ limit: 1, timeout: "10s" });
    } catch (e) {
      consumeErr = e;
    }
    tries++;
  }

  if (consumeErr && msgs.length === 0) {
    fail(`Consumer error: ${consumeErr}`);
  }

  const ok = check(msgs, {
    "got message": (m) => m.length > 0,
    "uuid preserved": (m) => JSON.parse(bytesToString(m[0].value)).uuid === id,
  });
  if (!ok) {
    fail("Message validation failed");
  }

  sleep(0.05);
}

export function teardown() {
  if (writer) {
    writer.close();
  }
  if (reader) {
    reader.close();
  }
}
