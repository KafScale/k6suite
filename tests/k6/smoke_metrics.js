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
