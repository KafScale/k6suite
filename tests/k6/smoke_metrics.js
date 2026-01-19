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
 * smoke_metrics.js - Metrics Endpoint Smoke Test
 *
 * Scenario: S2 (Observability)
 *
 * Purpose: Ensure the broker metrics endpoint is reachable
 * and returns a non-empty payload.
 *
 * Usage:
 *   K6_PROFILE=local-docker ./k6 run tests/k6/smoke_metrics.js
 *   K6_PROFILE=local-service ./k6 run tests/k6/smoke_metrics.js
 */

import http from "k6/http";
import { check, fail } from "k6";
import { getProfile } from "../../config/profiles.js";

const config = getProfile();
const metricsUrl = config.kafscale && config.kafscale.metricsUrl;
const target = __ENV.K6_TARGET || "kafscale";

export const options = {
  vus: 1,
  iterations: 1,
};

export default function () {
  if (target !== "kafscale") {
    console.log("Skipping metrics smoke: K6_TARGET != kafscale");
    return;
  }

  if (!metricsUrl) {
    fail("Metrics URL is not configured in the selected profile");
  }

  const res = http.get(metricsUrl, { timeout: "5s" });
  const ok = check(res, {
    "metrics status 200": (r) => r.status === 200,
    "metrics body non-empty": (r) => r.body && r.body.length > 0,
  });

  if (!ok) {
    fail(`Metrics check failed: status=${res.status}`);
  }
}
