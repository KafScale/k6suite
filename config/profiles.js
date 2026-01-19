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

import { fail } from "k6";

const raw = open("./profiles.json");
const data = JSON.parse(raw);

export const profiles = data.profiles || {};
export const defaultProfile = data.default_profile || "local-service";

export function getProfile() {
  const requested = __ENV.K6_PROFILE || defaultProfile;
  const profile = profiles[requested];
  if (!profile) {
    fail(`Unknown K6_PROFILE: ${requested}`);
  }
  console.log(`K6 profile: ${requested}`);
  if (profile.name) {
    console.log(`K6 profile name: ${profile.name}`);
  }
  if (profile.description) {
    console.log(`K6 profile description: ${profile.description}`);
  }
  if (profile.brokers) {
    console.log(`K6 brokers: ${profile.brokers.join(", ")}`);
  }
  if (profile.kafscale && profile.kafscale.metricsUrl) {
    console.log(`K6 metrics URL: ${profile.kafscale.metricsUrl}`);
  }
  return profile;
}
