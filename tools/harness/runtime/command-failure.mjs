import { randomUUID } from "node:crypto";
import { lstatSync, mkdirSync, readFileSync, renameSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";
import { normalizeFailureRecord, parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { borrowSuiteRuntime } from "./suite-runtime.mjs";

const schemaID = "cartulary.harness_command_failure.v1";
const contextKey = "CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT";
const accountingFailure = Object.freeze({ failure_class: "harness", failure_reason: "scheduler_accounting_error" });

export class CommandFailure extends Error {
  constructor(message, failure, options) {
    super(message, options);
    Object.assign(this, failure);
  }
}

function assertPrivate(file, directory = false) {
  const info = lstatSync(file);
  if (info.uid !== process.getuid() || info.isSymbolicLink() ||
      (directory ? !info.isDirectory() : !info.isFile() || info.nlink !== 1 || info.size > 4096) ||
      (info.mode & 0o777) !== (directory ? 0o700 : 0o600)) {
    throw new Error("unsafe private command failure channel");
  }
}

function channel(repoRoot, environment, expected) {
  const runRoot = path.resolve(repoRoot, environment.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results", environment.CARTULARY_TEST_RUN_ID);
  const runtime = borrowSuiteRuntime({ repoRoot, runRoot, environment });
  validateSchemaSync(schemaID, { ...expected, ...accountingFailure });
  if (runtime.runID !== expected.run_id) throw new Error("command failure run mismatch");
  const directory = runtime.privatePath(`command-failure-${expected.invocation_id}`);
  return { directory, file: path.join(directory, "failure.json") };
}

function readEnvelope(repoRoot, environment, expected) {
  const { directory, file } = channel(repoRoot, environment, expected);
  assertPrivate(directory, true);
  if (lstatSync(path.join(directory, "conflict"), { throwIfNoEntry: false })) throw new Error("conflicting command failures");
  if (!lstatSync(file, { throwIfNoEntry: false })) {
    if (lstatSync(path.join(directory, "writer"), { throwIfNoEntry: false })) throw new Error("incomplete command failure publication");
    return null;
  }
  assertPrivate(file);
  const envelope = parseStrictJSON(readFileSync(file, "utf8"));
  validateSchemaSync(schemaID, envelope);
  for (const [key, value] of Object.entries(expected)) {
    if (envelope[key] !== value) throw new Error("command failure invocation mismatch");
  }
  const normalized = normalizeFailureRecord({ failure_reason: envelope.failure_reason });
  if (normalized.failure_class !== envelope.failure_class) throw new Error("inconsistent command failure taxonomy");
  return { failure_class: envelope.failure_class, failure_reason: envelope.failure_reason };
}

export function createCommandFailureContext({ repoRoot, environment, unitID, commandID }) {
  const expected = Object.freeze({ schema_id: schemaID, run_id: environment.CARTULARY_TEST_RUN_ID, unit_id: unitID, command_id: commandID, invocation_id: randomUUID() });
  const { directory } = channel(repoRoot, environment, expected);
  mkdirSync(directory, { mode: 0o700 });
  return {
    environment: { [contextKey]: JSON.stringify(expected) },
    read() {
      try { return readEnvelope(repoRoot, environment, expected); }
      catch { return accountingFailure; }
    },
    close() { rmSync(directory, { recursive: true, force: true }); },
  };
}

export function readCommandFailure(repoRoot, environment = process.env) {
  if (!environment[contextKey]) return null;
  try { return readEnvelope(repoRoot, environment, parseStrictJSON(environment[contextKey])); }
  catch { return accountingFailure; }
}

export function publishCommandFailure(repoRoot, failure, environment = process.env) {
  if (!environment[contextKey]) return;
  const expected = parseStrictJSON(environment[contextKey]);
  const envelope = { ...expected, failure_class: failure.failure_class, failure_reason: failure.failure_reason };
  validateSchemaSync(schemaID, envelope);
  const { directory, file } = channel(repoRoot, environment, expected);
  assertPrivate(directory, true);
  try { mkdirSync(path.join(directory, "writer"), { mode: 0o700 }); }
  catch (error) {
    if (error.code !== "EEXIST") throw error;
    const existing = readEnvelope(repoRoot, environment, expected);
    if (existing?.failure_class === failure.failure_class && existing?.failure_reason === failure.failure_reason) return;
    writeFileSync(path.join(directory, "conflict"), "", { mode: 0o600 });
    throw new Error("conflicting command failures");
  }
  const temporary = path.join(directory, "pending.json");
  writeFileSync(temporary, `${JSON.stringify(envelope)}\n`, { flag: "wx", mode: 0o600 });
  renameSync(temporary, file);
}

export function reportCommandFailure(repoRoot, error, fallback, environment = process.env) {
  const failure = error instanceof CommandFailure ? error : fallback;
  try { publishCommandFailure(repoRoot, failure, environment); }
  catch { return accountingFailure; }
  return { failure_class: failure.failure_class, failure_reason: failure.failure_reason };
}
