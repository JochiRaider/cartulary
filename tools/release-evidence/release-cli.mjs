#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import { closeSync, openSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { resolveRetainedArtifactIdentity, secureMkdir } from "../harness/contract/index.mjs";
import { atomicLocalFile } from "../harness/runtime/secure-local-files.mjs";
import { releaseRoots } from "./verify-release-evidence.mjs";
import { runContextEnvironment } from "../harness/contract/run-context.mjs";

const repoRoot = path.resolve(import.meta.dirname, "../..");

// The caller may name an existing root for any stage, by the roots-file key.
export const stageInputs = Object.freeze({ ci: "CI_RESULTS_DIR", release_check: "RELEASE_CHECK_RESULTS_DIR", release_measure: "RELEASE_MEASURE_RESULTS_DIR" });

/**
 * Compose a release (R5). A stage root the caller names is reused as given; a
 * missing stage runs as its own top-level command with an explicit run ID.
 * Stages run one at a time, so release-measure's quiet campaign never
 * overlaps other work, and no root is chosen by recency. release-measure
 * consumes the package of this release's release-check root. The decision
 * is release-evidence-verify's, over the roots file this writes.
 */
export function composeRelease({ supplied, runID, resultsDir, releaseRoot, runMake, report = () => {} }) {
  const roots = {};
  for (const [key, target] of Object.entries(releaseRoots)) {
    if (supplied[key]) {
      roots[key] = path.resolve(repoRoot, supplied[key]);
      report({ stage: target, root: roots[key], status: "reused" });
      continue;
    }
    const stageRunID = `${runID}.${target}`;
    roots[key] = path.join(resultsDir, stageRunID);
    // A release executes every row: its ci never reuses cached results (D12).
    const inputs = key === "release_measure" ? { RELEASE_RESULTS_DIR: roots.release_check }
      : key === "ci" ? { CARTULARY_HARNESS_CACHE_MODE: "off" } : {};
    const status = runMake(target, { runID: stageRunID, inputs });
    report({ stage: target, root: roots[key], status: status === 0 ? "pass" : "fail" });
    if (status !== 0) return { status, failed: target, roots };
  }
  const rootsFile = path.join(releaseRoot, "release/roots.json");
  atomicLocalFile(rootsFile, `${JSON.stringify({ schema_id: "cartulary.release_evidence_roots.v2", ...roots }, null, 2)}\n`);
  const verifyRunID = `${runID}.release-evidence-verify`;
  const status = runMake("release-evidence-verify", { runID: verifyRunID, inputs: { EVIDENCE_ROOTS_FILE: rootsFile } });
  report({ stage: "release-evidence-verify", root: path.join(resultsDir, verifyRunID), status: status === 0 ? "pass" : "fail" });
  return { status, failed: status === 0 ? null : "release-evidence-verify", roots, rootsFile };
}

// A stage is a fresh top-level Make invocation: no inherited Make state,
// harness identity, service attachment or release input; its output goes to
// a log beside the roots file.
function makeRunner(resultsDir, logRoot) {
  return (target, { runID, inputs }) => {
    const environment = Object.fromEntries(Object.entries(process.env).filter(([name]) =>
      !/^(?:MAKEFLAGS|MFLAGS|GNUMAKEFLAGS|MAKEOVERRIDES|MAKELEVEL)$/u.test(name) &&
      !/^CARTULARY_(?:TEST|HARNESS|MAKE)_/u.test(name) && !Object.values(stageInputs).includes(name) && name !== "RELEASE_RESULTS_DIR"));
    Object.assign(environment, runContextEnvironment({ runID, resultsDir }));
    const log = openSync(path.join(logRoot, `${target}.log`), "wx", 0o600);
    try {
      const result = spawnSync(process.env.MAKE || "make", ["--no-print-directory", target, ...Object.entries(inputs).map(([name, value]) => `${name}=${value}`)],
        { cwd: repoRoot, env: environment, stdio: ["ignore", log, log] });
      return result.status ?? 1;
    } finally { closeSync(log); }
  };
}

// A root inside the repository prints relative to it; any other as given.
function display(file) {
  const relative = path.relative(repoRoot, file);
  return relative.startsWith("..") ? file : relative;
}

function parseArgs(argv) {
  const supplied = {};
  const flags = Object.fromEntries(Object.keys(stageInputs).map(key => [`--${key.replaceAll("_", "-")}`, key]));
  for (let index = 0; index < argv.length; index += 2) {
    const key = flags[argv[index]];
    if (!key || !argv[index + 1] || supplied[key]) throw new Error("usage: make release [CI_RESULTS_DIR=<root>] [RELEASE_CHECK_RESULTS_DIR=<root>] [RELEASE_MEASURE_RESULTS_DIR=<root>]");
    supplied[key] = argv[index + 1];
  }
  return supplied;
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    const supplied = parseArgs(process.argv.slice(2));
    const identity = resolveRetainedArtifactIdentity("release", process.env);
    const resultsDir = identity.result_root;
    const logRoot = path.join(identity.run_root, "release");
    secureMkdir(logRoot);
    const outcome = composeRelease({ supplied, runID: identity.run_id, resultsDir, releaseRoot: identity.run_root,
      runMake: makeRunner(resultsDir, logRoot),
      report: ({ stage, root, status }) => process.stdout.write(`[RELEASE] stage=${stage} status=${status} root=${display(root)}\n`) });
    if (outcome.status !== 0) {
      process.stderr.write(`release failed at ${outcome.failed}; rerun it alone or name the passing roots (log: ${display(path.join(logRoot, `${outcome.failed}.log`))})\n`);
      process.exitCode = outcome.status;
    }
  } catch (error) {
    process.stderr.write(`release failed: ${error.message}\n`);
    process.exitCode = 2;
  }
}
