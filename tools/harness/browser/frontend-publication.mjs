import { randomUUID } from "node:crypto";
import path from "node:path";
import { spawn } from "../workspace/child-process.mjs";
import { parseStrictJSON, redactString } from "../contract/index.mjs";
import { ownedProcess, stopOwnedProcess } from "../runtime/owned-process.mjs";
import { borrowRuntime } from "../runtime/session-client.mjs";

const repository = path.resolve(import.meta.dirname, "../../..");
const retainedInputs = new Set(["CARTULARY_HOST_ADMISSION_LEASE", "CARTULARY_UNIT_CPU_TOKENS", "CARTULARY_HARNESS_CACHE_MODE"]);

export function frontendPublicationEnvironment(environment) {
  return Object.fromEntries(Object.entries(environment).filter(([key]) =>
    (!key.startsWith("CARTULARY_") || retainedInputs.has(key)) &&
    !["MAKEFLAGS", "MAKEOVERRIDES", "MFLAGS", "MAKELEVEL"].includes(key)));
}

// Fresh public builds borrow the waiting browser unit's admission envelope.
// The existing native runtime retains exact process authority if its test dies.
export function startFrontendPublication(measurement, { environment = process.env } = {}) {
  if (typeof measurement !== "boolean") throw new Error("frontend publication requires an artifact selection");
  if (!environment.CARTULARY_TEST_RUN_ID || !environment.CARTULARY_TEST_RESULTS_DIR) throw new Error("frontend publication requires its owning run");
  const parentRoot = path.resolve(repository, environment.CARTULARY_TEST_RESULTS_DIR, environment.CARTULARY_TEST_RUN_ID);
  const runtime = borrowRuntime({ repoRoot: repository, runRoot: parentRoot, environment });
  const runID = `frontend-publication-${randomUUID()}`;
  const results = path.join(parentRoot, "frontend-publications");
  const child = spawn("make", ["build", measurement ? "COMPONENT=web-measurement" : "COMPONENT=web"], {
    cwd: repository, detached: true, stdio: ["ignore", "pipe", "pipe"],
    env: { ...frontendPublicationEnvironment(environment), CARTULARY_TEST_RUN_ID: runID,
      CARTULARY_TEST_RESULTS_DIR: results, CARTULARY_OUTPUT_MODE: "ci",
      CARTULARY_HARNESS_CACHE_MODE: environment.CARTULARY_HARNESS_CACHE_MODE || "normal" },
  });
  let output = "", outputBytes = 0, outputTruncated = false;
  const retain = (data) => {
    outputBytes += data.length;
    if (outputBytes <= 1024 * 1024) output += data.toString();
    else outputTruncated = true;
  };
  child.stdout.on("data", retain); child.stderr.on("data", retain);
  const closed = new Promise((resolve, reject) => {
    child.once("error", reject);
    child.once("close", (code, signal) => resolve({ code, signal }));
  });
  closed.catch(() => {});
  let proof, registered = false, registrationError, stopping;
  try {
    if (child.pid) {
      proof = ownedProcess(child.pid);
      runtime.record({ kind: "helper_process", target: proof,
        depends_on: parseStrictJSON(environment.CARTULARY_HARNESS_RESOURCE_DEPENDENCIES || "[]") });
      registered = true;
    }
  } catch (error) { registrationError = error; }
  const stop = () => stopping ??= (async () => {
    if (proof) await stopOwnedProcess(proof, { environment });
    await closed.catch(() => {});
    if (registered) {
      runtime.record({ kind: "helper_process", target: proof, state: "released" });
      registered = false;
    }
  })();
  const finished = (async () => {
    try {
      if (registrationError) throw registrationError;
      const outcome = await closed;
      if (outcome.code !== 0 || outcome.signal || outputTruncated) {
        throw new Error(`Independent frontend Make build failed: code=${outcome.code} signal=${outcome.signal ?? "none"} output_truncated=${outputTruncated}\n${redactString(output).slice(-8192)}`);
      }
      return path.join(results, runID);
    } finally { await stop(); }
  })();
  finished.catch(() => {});
  return { finished, stop };
}
