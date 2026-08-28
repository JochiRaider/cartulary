#!/usr/bin/env node

import { existsSync, readFileSync, renameSync, writeFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";

import { validateSchemaSync } from "../contract/index.mjs";

function parseArgs(argv) {
  if (argv.length % 2 !== 0) throw new Error("browser reset attempt flags require values");
  const values = {};
  for (let index = 0; index < argv.length; index += 2) {
    const flag = argv[index];
    if (!new Set([
      "--label",
      "--status",
      "--exit-code",
      "--attempt-file",
      "--database-file",
      "--object-store-marker-file",
      "--state-marker-file",
      "--backend-ready-marker-file",
      "--lease-file",
      "--generation-before",
      "--duration-ms",
    ]).has(flag)) throw new Error(`unsupported browser reset attempt flag ${flag}`);
    values[flag] = argv[index + 1];
  }
  return values;
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function relativeToRun(file) {
  const results = process.env.CARTULARY_TEST_RESULTS_DIR;
  const runID = process.env.CARTULARY_TEST_RUN_ID;
  if (!results || !runID) throw new Error("browser reset evidence requires run identity");
  const root = path.resolve(results, runID);
  const relative = path.relative(root, path.resolve(file)).replaceAll("\\", "/");
  if (!relative || relative.startsWith("../") || path.isAbsolute(relative)) {
    throw new Error("browser reset evidence path escapes the run root");
  }
  return relative;
}

function classification(exitCode) {
  if (exitCode === 13) return ["timing", "timeout_failure"];
  if (exitCode === 2) return ["config", "configuration_error"];
  if (exitCode === 11) return ["artifact", "artifact_error"];
  return ["harness", "fixture_error"];
}

function main(argv) {
  const args = parseArgs(argv);
  const label = String(args["--label"] ?? "");
  const status = args["--status"];
  const exitCode = Number.parseInt(args["--exit-code"], 10);
  const attemptFile = args["--attempt-file"];
  const databaseFile = args["--database-file"];
  const objectStoreMarkerFile = args["--object-store-marker-file"];
  const stateMarkerFile = args["--state-marker-file"];
  const backendReadyMarkerFile = args["--backend-ready-marker-file"];
  const leaseFile = args["--lease-file"];
  const generationBefore = Number.parseInt(args["--generation-before"], 10);
  const durationMS = Number.parseInt(args["--duration-ms"], 10);
  if (!/^[A-Za-z0-9_.-]+$/u.test(label) || !["pass", "fail"].includes(status)) {
    throw new Error("browser reset attempt has invalid identity or status");
  }
  const database = existsSync(databaseFile) ? readJSON(databaseFile) : null;
  if (database) validateSchemaSync(database.schema_id, database);
  const lease = readJSON(leaseFile);
  const generationAfter = Number.isSafeInteger(lease.backend_generation)
    ? lease.backend_generation
    : generationBefore;
  const head = existsSync(lease.backend_generation_head ?? "")
    ? readJSON(lease.backend_generation_head)
    : null;
  const [failureClass, failureReason] = classification(exitCode);
  const databasePassed = database?.status === "pass";
  const objectStoreReset = existsSync(objectStoreMarkerFile);
  const browserStateCleared = existsSync(stateMarkerFile);
  const backendReady = existsSync(backendReadyMarkerFile);
  const generationPublished = generationAfter === generationBefore + 1 && head !== null;
	const stages = [
		{ stage: "allocation_validated", status: "pass" },
		{ stage: "backend_stopped", status: database ? "pass" : "fail" },
		{
			stage: "database_reset",
			status: !database ? "skipped" : databasePassed ? "pass" : "fail",
		},
    {
      stage: "object_store_reset",
      status: !databasePassed ? "skipped" : objectStoreReset ? "pass" : "fail",
    },
    {
      stage: "browser_state_cleared",
      status: !objectStoreReset ? "skipped" : browserStateCleared ? "pass" : "fail",
		},
		{
			stage: "replacement_backend_ready",
			status: !browserStateCleared ? "skipped" : backendReady ? "pass" : "fail",
		},
		{
      stage: "generation_published",
      status: !backendReady ? "skipped" : generationPublished ? "pass" : "fail",
		},
	];
  const failureStage = status === "pass"
    ? null
    : !database
      ? "backend_stop_or_preflight"
      : database.status === "fail"
        ? "database"
        : !objectStoreReset
          ? "object_store"
          : !browserStateCleared
            ? "browser_state"
            : !backendReady
              ? "replacement_backend"
              : "generation_publication";
  const payload = {
    schema_id: "cartulary.browser_reset_attempt.v1",
    reset_id: label,
    status,
    attempt: 1,
    duration_ms: Math.max(durationMS, 0),
    runtime_profile_id: String(process.env.CARTULARY_BROWSER_RUNTIME_PROFILE_ID ?? ""),
		stages,
    backend_generation_before: generationBefore,
    backend_generation_after: generationAfter,
    failure_stage: failureStage,
    database_diagnostic_ref: database ? relativeToRun(databaseFile) : null,
    database_stage: database?.stage ?? null,
    database_sqlstate: database?.sqlstate ?? null,
    persistent_state_reset: databasePassed,
    browser_state_cleared: browserStateCleared,
    backend_ready: backendReady,
    backend_generation_ref: generationPublished ? String(head.artifact_ref) : null,
    tainted: status === "fail",
    failure_class: status === "fail" ? failureClass : null,
    failure_reason: status === "fail" ? failureReason : null,
  };
  validateSchemaSync(payload.schema_id, payload);
  const temporary = `${attemptFile}.tmp-${process.pid}`;
  writeFileSync(temporary, `${JSON.stringify(payload, null, 2)}\n`, { mode: 0o600 });
  renameSync(temporary, attemptFile);
}

try {
  main(process.argv.slice(2));
} catch (error) {
  process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
  process.exitCode = 1;
}
