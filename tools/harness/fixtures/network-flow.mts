import {
  closeSync,
  fstatSync,
  mkdirSync,
  mkdtempSync,
  readdirSync,
  readFileSync,
  readSync,
  writeFileSync,
} from "node:fs";
import { createRequire } from "node:module";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import {
  copyVerified,
  digest,
  openRegular,
  safePath,
  seal,
  secureDirectory,
  type FileRecord,
} from "./fixture-artifacts.mjs";

// Only file admission and evidence belong here. Owner tests interpret product data.
const repoRoot = fileURLToPath(new URL("../../../", import.meta.url));
const require = createRequire(import.meta.url);
const { foundationSchemaValidators } = require(
  path.join(
    repoRoot,
    "tools/harness/contract/generated/foundation-schema-validators.cjs",
  ),
) as {
  foundationSchemaValidators: Record<string, (value: unknown) => boolean>;
};
const manifestSchema = "cartulary.network_flow_fixture_manifest.v2";
const scenarioSchema = "cartulary.network_flow_fixture_scenario.v2";
const summarySchema = "cartulary.network_flow_fixture_execution.v1";
const fixturePattern = /^NF-FIX-[0-9]{3}-[a-z0-9][a-z0-9-]*$/u;
const segmentPattern = /^[A-Za-z0-9._@+-]+$/u;

type Manifest = {
  fixture_id: string;
  freeze: { status: string; revision: number };
  source_files: FileRecord[];
  expected_artifacts: FileRecord[];
  transcript_files: FileRecord[];
  source_bundle_sha256: string;
  expected_bundle_sha256: string;
};
export type FixtureExecutionSummary = {
  schema_id: string;
  fixture_id: string;
  freeze_revision: number | null;
  manifest_sha256: string | null;
  source_bundle_sha256: string | null;
  expected_bundle_sha256: string | null;
  materialized_input_root: string | null;
  produced_artifacts: FileRecord[];
  comparison_status: "passed" | "failed" | "blocked";
  failure_phase: "admission" | "comparison" | null;
};
export type NetworkFlowFixture = {
  read(logicalPath: string): Buffer;
  record(name: string, data: Uint8Array | string): void;
};
export class FixtureExecutionError extends Error {
  constructor(
    readonly summaryPath: string,
    cause: unknown,
  ) {
    super("Fixture execution failed; retained evidence: " + summaryPath, {
      cause,
    });
    this.name = "FixtureExecutionError";
  }
}
class DraftFixtureError extends Error {}

export async function decodeFixtureJSON(data: Uint8Array): Promise<unknown> {
  const { parseStrictJSON } = (await import(
    pathToFileURL(path.join(repoRoot, "tools/harness/contract/strict-json.mjs"))
      .href
  )) as { parseStrictJSON(source: string, label?: string): unknown };
  return parseStrictJSON(
    new TextDecoder("utf-8", { fatal: true, ignoreBOM: true }).decode(data),
  );
}
function validate(schema: string, value: unknown): void {
  const validator = foundationSchemaValidators[schema];
  if (!validator?.(value))
    throw new Error("Fixture schema validation failed: " + schema);
}
export function selectedFixtureRunRoot(): string {
  const id = process.env.CARTULARY_TEST_RUN_ID ?? "";
  const results = process.env.CARTULARY_TEST_RESULTS_DIR;
  if (!safePath(id) || id.includes("/") || !results)
    throw new Error("Fixture execution requires a selected harness run");
  const root = path.resolve(repoRoot, results, id);
  secureDirectory(root, true);
  return root;
}

function inventory(root: string, prefix = ""): string[] {
  const result: string[] = [];
  for (const entry of readdirSync(path.join(root, prefix), {
    withFileTypes: true,
  })) {
    const logicalPath = prefix ? prefix + "/" + entry.name : entry.name;
    if (!safePath(logicalPath) || entry.isSymbolicLink())
      throw new Error("Unsafe fixture inventory");
    if (entry.isDirectory()) result.push(...inventory(root, logicalPath));
    else if (entry.isFile()) result.push(logicalPath);
    else throw new Error("Nonregular fixture inventory entry");
  }
  return result.sort();
}
function bundleDigest(files: FileRecord[]): string {
  return digest(
    files
      .map(
        (file) =>
          file.logical_path +
          "\0" +
          file.sha256 +
          "\0" +
          file.size_bytes +
          "\n",
      )
      .join(""),
  );
}


async function admit(
  corpus: string,
  fixtureID: string,
  revision: number,
  workspace: string,
  runRoot: string,
  summary: FixtureExecutionSummary,
): Promise<NetworkFlowFixture> {
  const root = path.join(corpus, fixtureID);
  secureDirectory(root);
  const descriptor = openRegular(root, "manifest.json");
  let data: Buffer;
  try {
    if (fstatSync(descriptor).size > 1024 * 1024)
      throw new Error("Excessive fixture manifest");
    const buffer = Buffer.alloc(1024 * 1024 + 1);
    let size = 0;
    while (size < buffer.length) {
      const n = readSync(descriptor, buffer, size, buffer.length - size, null);
      if (!n) break;
      size += n;
    }
    if (size === buffer.length) throw new Error("Excessive fixture manifest");
    data = buffer.subarray(0, size);
  } finally {
    closeSync(descriptor);
  }
  summary.manifest_sha256 = digest(data);
  const decoded = await decodeFixtureJSON(data);
  validate(manifestSchema, decoded);
  const manifest = decoded as Manifest;
  if (manifest.fixture_id !== fixtureID)
    throw new Error("Fixture identity mismatch");
  if (manifest.freeze.status !== "frozen")
    throw new DraftFixtureError("Draft fixture cannot qualify behavior");
  if (manifest.freeze.revision !== revision)
    throw new Error("Frozen fixture revision mismatch");
  summary.freeze_revision = manifest.freeze.revision;
  summary.source_bundle_sha256 = manifest.source_bundle_sha256;
  summary.expected_bundle_sha256 = manifest.expected_bundle_sha256;
  const listed = new Map<string, FileRecord>();
  for (const files of [
    manifest.source_files,
    manifest.expected_artifacts,
    manifest.transcript_files,
  ]) {
    let previous = "";
    for (const file of files) {
      if (
        !safePath(file.logical_path) ||
        file.logical_path === "manifest.json" ||
        file.logical_path <= previous ||
        listed.has(file.logical_path)
      ) {
        throw new Error("Invalid, duplicate or unordered fixture path");
      }
      listed.set(file.logical_path, file);
      previous = file.logical_path;
    }
  }
  if (
    bundleDigest(manifest.source_files) !== manifest.source_bundle_sha256 ||
    bundleDigest([
      ...manifest.expected_artifacts,
      ...manifest.transcript_files,
    ]) !== manifest.expected_bundle_sha256
  ) {
    throw new Error("Fixture bundle digest mismatch");
  }
  if (
    JSON.stringify(inventory(root)) !==
    JSON.stringify(["manifest.json", ...listed.keys()].sort())
  ) {
    throw new Error("Fixture inventory mismatch");
  }
  const inputs = path.join(workspace, "inputs");
  const outputs = path.join(workspace, "produced");
  mkdirSync(inputs, { mode: 0o700 });
  mkdirSync(outputs, { mode: 0o700 });
  for (const file of listed.values()) {
    copyVerified(root, file, inputs);
    if (file.logical_path.endsWith(".scenario.json")) {
      const scenario = await decodeFixtureJSON(
        readFileSync(path.join(inputs, file.logical_path)),
      );
      validate(scenarioSchema, scenario);
      if ((scenario as { fixture_id: string }).fixture_id !== fixtureID)
        throw new Error("Scenario identity mismatch");
    }
  }
  seal(inputs);
  summary.materialized_input_root = path
    .relative(runRoot, inputs)
    .split(path.sep)
    .join("/");
  return {
    read(logicalPath) {
      if (!listed.has(logicalPath)) throw new Error("Unlisted fixture input");
      return readFileSync(path.join(inputs, logicalPath));
    },
    record(name, value) {
      if (
        !safePath(name) ||
        !segmentPattern.test(name) ||
        name.length > 120 ||
        summary.produced_artifacts.length >= 64
      ) {
        throw new Error("Invalid or excessive produced artifacts");
      }
      const bytes =
        typeof value === "string"
          ? Buffer.from(value, "utf8")
          : Buffer.from(value);
      const filename = path.join(outputs, name);
      writeFileSync(filename, bytes, { flag: "wx", mode: 0o600 });
      summary.produced_artifacts.push({
        logical_path: path
          .relative(runRoot, filename)
          .split(path.sep)
          .join("/"),
        size_bytes: bytes.length,
        sha256: digest(bytes),
      });
    },
  };
}

export async function runNetworkFlowFixture(
  config: { fixtureID: string; revision: number; corpusRoot?: string },
  compare: (fixture: NetworkFlowFixture) => void | Promise<void>,
): Promise<{ summaryPath: string; summary: FixtureExecutionSummary }> {
  if (
    !fixturePattern.test(config.fixtureID) ||
    config.fixtureID.length > 120 ||
    !Number.isSafeInteger(config.revision) ||
    config.revision < 1
  )
    throw new Error("Invalid fixture configuration");
  const runRoot = selectedFixtureRunRoot();
  const workspace = mkdtempSync(
    path.join(runRoot, config.fixtureID + "-node-"),
  );
  const summaryPath = path.join(workspace, "comparison-summary.json");
  const summary: FixtureExecutionSummary = {
    schema_id: summarySchema,
    fixture_id: config.fixtureID,
    freeze_revision: null,
    manifest_sha256: null,
    source_bundle_sha256: null,
    expected_bundle_sha256: null,
    materialized_input_root: null,
    produced_artifacts: [],
    comparison_status: "failed",
    failure_phase: "admission",
  };
  let failure: unknown;
  let failed = false;
  try {
    const fixture = await admit(
      config.corpusRoot ?? path.join(repoRoot, "fixtures/network-flow"),
      config.fixtureID,
      config.revision,
      workspace,
      runRoot,
      summary,
    );
    summary.failure_phase = "comparison";
    await compare(fixture);
    summary.comparison_status = "passed";
    summary.failure_phase = null;
  } catch (error) {
    failed = true;
    failure = error;
    if (error instanceof DraftFixtureError)
      summary.comparison_status = "blocked";
  }
  try {
    validate(summarySchema, summary);
    writeFileSync(summaryPath, JSON.stringify(summary, null, 2) + "\n", {
      flag: "wx",
      mode: 0o600,
    });
  } catch (error) {
    throw new FixtureExecutionError(
      summaryPath,
      failed ? new AggregateError([failure, error]) : error,
    );
  }
  if (failed) throw new FixtureExecutionError(summaryPath, failure);
  return { summaryPath, summary };
}
