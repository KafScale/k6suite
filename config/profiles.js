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
