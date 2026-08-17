#!/usr/bin/env node

import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { validateSchemaSync } from "../contract/index.mjs";
import { runPrivateCapturedProcess } from "../runtime/private-child-process.mjs";
import {
  loadPerformanceFixtureSnapshotRegistry,
  postgresMigrationDigest,
  snapshotKey,
} from "./index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");

function parseArgs(argv) {
  const values = new Map();
  for (let index = 0; index < argv.length; index += 2) {
    const flag = argv[index];
    const value = argv[index + 1];
    if (!new Set(["--fixture-profile", "--snapshot-key"]).has(flag) || !value) {
      throw new Error("usage: snapshot-builder-cli.mjs --fixture-profile <id> --snapshot-key <key>");
    }
    if (values.has(flag)) throw new Error(`duplicate ${flag}`);
    values.set(flag, value);
  }
  if (values.size !== 2) {
    throw new Error("usage: snapshot-builder-cli.mjs --fixture-profile <id> --snapshot-key <key>");
  }
  return {
    fixtureProfileID: values.get("--fixture-profile"),
    snapshotKey: values.get("--snapshot-key"),
  };
}

function requiredEnv(name) {
  const value = String(process.env[name] ?? "").trim();
  if (!value) throw new Error(`${name} is required`);
  return value;
}

function runRoot() {
  const root = requiredEnv("CARTULARY_TEST_RESULTS_DIR");
  const runID = requiredEnv("CARTULARY_TEST_RUN_ID");
  if (!/^[A-Za-z0-9_.-]+$/u.test(runID)) throw new Error("unsafe test run identity");
  return path.resolve(repoRoot, root, runID);
}

function validateExisting(file, expected) {
  const artifact = JSON.parse(readFileSync(file, "utf8"));
  validateSchemaSync(artifact.schema_id, artifact);
  for (const [key, value] of Object.entries(expected)) {
    if (artifact[key] !== value) throw new Error(`existing snapshot build ${key} mismatch`);
  }
  if (artifact.state !== "sealed") throw new Error("existing snapshot build is not sealed");
}

function validateDiagnostics(file, expected) {
  const artifact = JSON.parse(readFileSync(file, "utf8"));
  validateSchemaSync("cartulary.performance_fixture_build_diagnostics.v1", artifact);
  for (const [key, value] of Object.entries(expected)) {
    if (artifact[key] !== value) throw new Error(`existing fixture build diagnostics ${key} mismatch`);
  }
  if (artifact.state !== "sealed") throw new Error("existing fixture build diagnostics are not sealed");
}

export async function runSnapshotBuilder(argv) {
  const args = parseArgs(argv);
  const registry = loadPerformanceFixtureSnapshotRegistry(repoRoot);
  const profile = registry.profiles.get(args.fixtureProfileID);
  if (!profile || profile.status !== "active") throw new Error("fixture profile is not active");
  const migrationDigest = postgresMigrationDigest(repoRoot);
  const expectedKey = snapshotKey(profile, migrationDigest);
  if (args.snapshotKey !== expectedKey) throw new Error("snapshot key does not match canonical profile input");
  const builderUnitID = requiredEnv("CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID");
  const expected = {
    fixture_profile_id: profile.fixture_profile_id,
    snapshot_key: expectedKey,
    migration_digest: migrationDigest,
    source_contract_digest: profile.source_contract_digest,
    builder_unit_id: builderUnitID,
  };
  const artifact = path.join(runRoot(), "performance-fixtures", expectedKey, "snapshot-build.json");
  const diagnostics = path.join(runRoot(), "performance-fixtures", expectedKey, "build-diagnostics.json");
  if (existsSync(artifact)) {
    validateExisting(artifact, expected);
    validateDiagnostics(diagnostics, {
      fixture_profile_id: profile.fixture_profile_id,
      snapshot_key: expectedKey,
      builder_unit_id: builderUnitID,
    });
    return artifact;
  }
  const executable = requiredEnv("CARTULARY_TEST_SERVICES_BIN");
  const result = await runPrivateCapturedProcess(executable, [
    "build-performance-fixture",
    "--fixture-profile", profile.fixture_profile_id,
    "--snapshot-key", expectedKey,
    "--migration-digest", migrationDigest,
    "--source-contract-digest", profile.source_contract_digest,
    "--builder-unit-id", builderUnitID,
    "--artifact-file", artifact,
  ], {
    captureID: `snapshot-builder-${expectedKey.slice(0, 16)}`,
    cwd: repoRoot,
    env: process.env,
    repoRoot,
    runRoot: runRoot(),
  });
  try {
    if (result.status !== 0) {
      throw new Error((result.stderr || result.stdout || `snapshot builder exited ${result.status}`).trim());
    }
  } finally {
    result.cleanup();
  }
  validateExisting(artifact, expected);
  validateDiagnostics(diagnostics, {
    fixture_profile_id: profile.fixture_profile_id,
    snapshot_key: expectedKey,
    builder_unit_id: builderUnitID,
  });
  return artifact;
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    process.stdout.write(`${await runSnapshotBuilder(process.argv.slice(2))}\n`);
  } catch (error) {
    process.stderr.write(`${error.message}\n`);
    process.exitCode = 1;
  }
}
