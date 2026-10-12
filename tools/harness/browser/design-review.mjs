import path from "node:path";
import { fileURLToPath } from "node:url";
import { generateRunId, validateResultRoot, validateRunId } from "../contract/harness-contract.mjs";
import { nativeReview } from "./ui-review/native-client.mjs";
import { toolProfile } from "./ui-review/toolchain.mjs";
import { repoRoot } from "./ui-review/policy.mjs";

export function reviewProfile(value = "network_flow_claimed") {
  const profile = value.trim();
  if (!["default", "network_flow_claimed"].includes(profile)) throw Object.assign(new Error("invalid review profile"), { exitCode: 2 });
  return profile;
}
export async function runDesignReview({ smoke = false, environment = process.env, writeOutput = (value) => process.stdout.write(value) } = {}) {
  const profile = reviewProfile(environment.REVIEW_PROFILE);
  const results = validateResultRoot(environment.CARTULARY_TEST_RESULTS_DIR || undefined, { root: repoRoot, create: false });
  const runID = environment.CARTULARY_TEST_RUN_ID ? validateRunId(environment.CARTULARY_TEST_RUN_ID) : generateRunId();
  let locator, outcome;
  try {
    outcome = await nativeReview({ command: smoke ? "browser-design-review-smoke" : "browser-design-review", mode: "seeded", origin: null, profile,
      tool_profile: toolProfile(), results_root: results, run_id: runID, preparation_policy: "ensure" }, { environment, onProgress: ({ kind, value }) => {
      if (kind === "preparing") { locator = value; writeOutput(`Preparing browser review. Diagnostics: ${path.join(results, runID)}\n`); }
      if (kind === "design_ready") {
        writeOutput(`Browser review ready: ${value.public_origin}\nPrivate login instructions: ${path.join(value.private_directory, "access.json")}\nSample files: ${value.samples}\nScenario index: ${path.join(results, runID, "review-session.json")}\n`);
        for (const scenario of value.scenarios) writeOutput(`  ${scenario.name}: ${scenario.url}\n`);
        writeOutput(smoke ? "Smoke complete; cleaning up owned resources.\n" : "Press Ctrl-C to stop and remove this session's data.\n");
      }
    } });
  } catch (error) {
    if (!locator) throw error;
    outcome = await nativeReview({ command: "ui-review-stop", locator, request: null, bundle_id: null }, { environment });
  }
  if (outcome.exit_code) throw Object.assign(new Error(outcome.failures[0]?.diagnostic_code ?? "review failed"), { exitCode: outcome.exit_code });
  return path.join(results, runID);
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    if (process.argv.slice(2).some((arg) => arg !== "--smoke")) throw Object.assign(new Error("invalid review arguments"), { exitCode: 2 });
    await runDesignReview({ smoke: process.argv.includes("--smoke") });
  }
  catch (error) {
    process.stderr.write("Browser review failed; inspect the exact run's review-session.json and cleanup-results.json.\n");
    process.exitCode = error.exitCode ?? 1;
  }
}
