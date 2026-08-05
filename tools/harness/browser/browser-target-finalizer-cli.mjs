#!/usr/bin/env node

import { createHash } from "node:crypto";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import {
  secureMkdir,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import { buildFrontendVisualReconciliation } from "./frontend-visual-reconciliation.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../..");

function parse(argv) {
  const options = { target: "", groups: [], groupTargets: new Map(), children: [] };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--target") options.target = argv[++index] ?? "";
    else if (arg === "--groups") options.groups = (argv[++index] ?? "").split(",").filter(Boolean);
    else if (arg === "--group-targets") {
      for (const entry of (argv[++index] ?? "").split(",").filter(Boolean)) {
        const separator = entry.lastIndexOf("=");
        if (separator < 1) throw new Error("invalid browser group target mapping");
        options.groupTargets.set(entry.slice(0, separator), entry.slice(separator + 1));
      }
    }
    else if (arg === "--children") options.children = (argv[++index] ?? "").split(",").filter(Boolean);
    else throw new Error("invalid browser target finalizer arguments");
  }
  if (!options.target || options.groups.length === 0 || options.groupTargets.size !== options.groups.length) {
    throw new Error("browser target finalizer requires target and exact group targets");
  }
  return options;
}

function runRoot() {
  const results = process.env.CARTULARY_TEST_RESULTS_DIR;
  const runID = process.env.CARTULARY_TEST_RUN_ID;
  if (!results || !runID) throw new Error("browser target finalizer requires result-root identity");
  return path.resolve(root, results, runID);
}

function groupResult(base, groupID, target) {
  const file = path.join(base, target, "browser-groups", groupID, "browser-group-result.json");
  if (!existsSync(file)) return null;
  const bytes = readFileSync(file);
  const result = JSON.parse(bytes.toString("utf8"));
  validateSchemaSync("cartulary.browser_group_result.v3", result);
  return {
    file,
    bytes,
    result,
  };
}

function sha256(bytes) {
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}

function relativeToRun(base, file) {
  const relative = path.relative(base, file).replaceAll("\\", "/");
  if (!relative || relative.startsWith("../") || path.isAbsolute(relative)) {
    throw new Error(`browser target artifact escapes run root: ${file}`);
  }
  return relative;
}

function writeTargetResult(base, options, groups) {
  const targetDirectory = path.join(base, options.target);
  const output = path.join(targetDirectory, "browser-target-result.json");
  if (existsSync(output)) {
    throw new Error(`browser target result is immutable: ${output}`);
  }
  secureMkdir(targetDirectory);
  const artifacts = [];
  let visualReconciliation = null;
  if (options.target === "browser-e2e-visual") {
    const reconciliationOutput = path.join(
      targetDirectory,
      "frontend-visual-reconciliation.json",
    );
    if (existsSync(reconciliationOutput)) {
      throw new Error(
        `frontend visual reconciliation is immutable: ${reconciliationOutput}`,
      );
    }
    visualReconciliation = buildFrontendVisualReconciliation({
      root,
      reportPaths: groups.map((group) =>
        path.join(path.dirname(group.file), "playwright-report.json"),
      ),
      attemptPassed: groups.every((group) => group.result.status === "pass"),
    });
    validateSchemaSync(visualReconciliation.schema_id, visualReconciliation);
    const reconciliationBytes = Buffer.from(
      `${JSON.stringify(visualReconciliation, null, 2)}\n`,
      "utf8",
    );
    secureWriteFile(reconciliationOutput, reconciliationBytes);
    artifacts.push({
      kind: "frontend_visual_reconciliation_v1",
      ref: relativeToRun(base, reconciliationOutput),
      sha256: sha256(reconciliationBytes),
    });
  }
  const sessionsByID = new Map();
  for (const group of groups) {
    const existing = sessionsByID.get(group.result.browser_session_id);
    const session = {
      browser_session_id: group.result.browser_session_id,
      runtime_profile_id: group.result.runtime_profile_id,
      service_requirement: group.result.service_requirement,
      artifacts: group.result.session_artifacts,
    };
    if (existing && JSON.stringify(existing) !== JSON.stringify(session)) {
      throw new Error(
        `browser session ${group.result.browser_session_id} has conflicting target evidence`,
      );
    }
    sessionsByID.set(group.result.browser_session_id, session);
  }
  const payload = {
    schema_id: "cartulary.browser_target_result.v1",
    target_id: options.target,
    status:
      groups.every((group) => group.result.status === "pass") &&
      (visualReconciliation === null || visualReconciliation.status === "pass")
        ? "pass"
        : "fail",
    group_results: groups
      .map((group) => ({
        group_id: group.result.group_id,
        browser_session_id: group.result.browser_session_id,
        ref: relativeToRun(base, group.file),
        sha256: sha256(group.bytes),
      }))
      .sort((left, right) => left.group_id.localeCompare(right.group_id)),
    sessions: [...sessionsByID.values()].sort((left, right) =>
      left.browser_session_id.localeCompare(right.browser_session_id),
    ),
    ...(artifacts.length > 0 ? { artifacts } : {}),
    generated_at: new Date().toISOString(),
  };
  validateSchemaSync(payload.schema_id, payload);
  secureWriteFile(output, `${JSON.stringify(payload, null, 2)}\n`);
  return payload;
}

const options = parse(process.argv.slice(2));
const base = runRoot();
const groups = options.groups
  .map((group) =>
    groupResult(base, group, options.groupTargets.get(group)),
  )
  .filter((group) => group !== null);
const complete = groups.length === options.groups.length;
const targetResult = complete ? writeTargetResult(base, options, groups) : null;
if (!complete) {
  process.stderr.write(`browser target ${options.target} is missing group results\n`);
  process.exitCode = 11;
} else if (targetResult.status !== "pass") {
  process.stderr.write(
    `browser target ${options.target} contains failed groups or target artifacts\n`,
  );
  process.exitCode = 10;
}
