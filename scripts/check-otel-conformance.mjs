#!/usr/bin/env node

import { createHash } from "node:crypto";
import { mkdirSync, readFileSync, statSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const snapshotPath = "contracts/otel/otel_source_snapshot.v1.json";
const otelDocPath = "docs/opentelemetry-instrumentation-nlspec.md";
const core01Path = "docs/spec/01_architecture_storage_and_view_contracts.md";
const scriptPath = "scripts/check-otel-conformance.mjs";
const expectedDigest = "3f8f80a2ed04521dfe29e50fcddd7f7de70145a6aee01959f985a65fbb4c8632";
const otelCommit = "d4a91bddb53b4c308df3e40171a60059183efd88";
const semconvCommit = "e018fe6f91862f5ed63c082f87697cddac596784";

const expectedTopLevelKeys = [
  "schema_id",
  "otel_spec_version",
  "otel_spec_ref",
  "otel_spec_commit_sha",
  "semconv_version",
  "semconv_ref",
  "semconv_commit_sha",
  "semconv_model_digest_algorithm",
  "semconv_model_digest",
  "semconv_generated_constants",
  "language_sdk_versions",
  "source_paths",
  "created_at",
  "created_by_tool",
];

const expectedSourcePaths = [
  ["Specification overview", "specification/overview.md", otelCommit],
  ["Configuration overview", "specification/configuration/README.md", otelCommit],
  ["Configuration data model", "specification/configuration/data-model.md", otelCommit],
  ["Configuration API", "specification/configuration/api.md", otelCommit],
  ["Configuration SDK", "specification/configuration/sdk.md", otelCommit],
  ["Common configuration parsing", "specification/configuration/common.md", otelCommit],
  ["SDK environment variables", "specification/configuration/sdk-environment-variables.md", otelCommit],
  ["Trace SDK", "specification/trace/sdk.md", otelCommit],
  ["Metrics API", "specification/metrics/api.md", otelCommit],
  ["Metrics SDK", "specification/metrics/sdk.md", otelCommit],
  ["Logs API", "specification/logs/api.md", otelCommit],
  ["Logs data model", "specification/logs/data-model.md", otelCommit],
  ["Resource SDK", "specification/resource/sdk.md", otelCommit],
  ["Protocol exporter", "specification/protocol/exporter.md", otelCommit],
  ["Common concepts", "specification/common/README.md", otelCommit],
  ["Versioning and stability", "specification/versioning-and-stability.md", otelCommit],
  ["Semantic conventions model", "semantic-conventions/model/**", semconvCommit],
  ["Semantic conventions docs", "semantic-conventions/docs/**", semconvCommit],
];

const requiredPackageFamilies = new Set([
  "API",
  "SDK",
  "Trace SDK",
  "Metrics SDK",
  "Logs SDK or bridge",
  "OTLP HTTP exporters",
  "OTLP gRPC exporters",
  "Semantic-convention constants",
  "Instrumentation adapters",
  "Resource-detector packages",
  "Autoconfiguration or declarative-config packages",
  "Browser-side OTel packages",
]);

function repoPath(relativePath) {
  return path.join(repoRoot, relativePath);
}

function readText(relativePath) {
  return readFileSync(repoPath(relativePath), "utf8");
}

function readJSON(relativePath) {
  return JSON.parse(readText(relativePath));
}

function gitBlobSHA(relativePath) {
  const content = readFileSync(repoPath(relativePath));
  return createHash("sha1")
    .update(`blob ${content.length}\0`)
    .update(content)
    .digest("hex");
}

function assert(condition, message, checks, id) {
  if (!condition) {
    checks.push({ id, status: "fail", message });
    return false;
  }
  checks.push({ id, status: "pass", message });
  return true;
}

function sectionBetween(text, start, end) {
  const startIndex = text.indexOf(start);
  if (startIndex < 0) {
    return "";
  }
  const endIndex = text.indexOf(end, startIndex + start.length);
  return endIndex < 0 ? text.slice(startIndex) : text.slice(startIndex, endIndex);
}

function publicErrorCodes(core01) {
  const section = sectionBetween(
    core01,
    "##### 3.3.6.1 Canonical public error-code registry",
    "##### 3.3.6.2",
  );
  return new Set(
    [...section.matchAll(/^\| `([^`]+)` \|/gm)]
      .map((match) => match[1])
      .filter((code) => code !== "error.code"),
  );
}

function mappedErrorCodes(otelDoc) {
  const section = sectionBetween(otelDoc, "**OTEL-REQ-142**", "**OTEL-REQ-143**");
  const counts = new Map();
  for (const line of section.split("\n")) {
    if (!line.startsWith("| `") || line.includes("---")) {
      continue;
    }
    const cells = line.split("|").map((cell) => cell.trim());
    if (cells.length < 4 || cells[1] === "`cartulary.error_class`") {
      continue;
    }
    for (const match of cells[2].matchAll(/`([^`]+)`/g)) {
      counts.set(match[1], (counts.get(match[1]) ?? 0) + 1);
    }
  }
  return counts;
}

function validateSnapshot(snapshot, checks) {
  const keys = Object.keys(snapshot).sort();
  const expectedKeys = [...expectedTopLevelKeys].sort();
  assert(
    JSON.stringify(keys) === JSON.stringify(expectedKeys),
    "source snapshot has exactly the adopted top-level keys",
    checks,
    "snapshot.top_level_keys",
  );
  assert(snapshot.schema_id === "cartulary.otel_source_snapshot.v1", "schema_id is adopted", checks, "snapshot.schema_id");
  assert(snapshot.otel_spec_version === "1.57.0", "OTel spec version is pinned", checks, "snapshot.otel_spec_version");
  assert(snapshot.otel_spec_ref === "v1.57.0", "OTel spec ref is immutable", checks, "snapshot.otel_spec_ref");
  assert(snapshot.otel_spec_commit_sha === otelCommit, "OTel spec commit SHA is full and pinned", checks, "snapshot.otel_spec_commit_sha");
  assert(snapshot.semconv_version === "1.41.0", "semantic-conventions version is pinned", checks, "snapshot.semconv_version");
  assert(snapshot.semconv_ref === "v1.41.0", "semantic-conventions ref is immutable", checks, "snapshot.semconv_ref");
  assert(snapshot.semconv_commit_sha === semconvCommit, "semantic-conventions commit SHA is full and pinned", checks, "snapshot.semconv_commit_sha");
  assert(snapshot.semconv_model_digest_algorithm === "semconv_model_digest_v1", "digest algorithm is adopted", checks, "snapshot.digest_algorithm");
  assert(snapshot.semconv_model_digest === expectedDigest, "semantic-conventions model digest is concrete", checks, "snapshot.digest");

  const constants = snapshot.semconv_generated_constants ?? {};
  assert(constants.source_kind === "repo_codegen", "generated constants source kind is closed", checks, "snapshot.constants.source_kind");
  assert(constants.input_model_digest === expectedDigest, "generated constants bind to model digest", checks, "snapshot.constants.digest_binding");
  assert(constants.generator_source_sha === gitBlobSHA(scriptPath), "generated constants source SHA matches checker script", checks, "snapshot.constants.source_sha");

  const paths = snapshot.source_paths ?? [];
  const seenPaths = new Set();
  const expectedByPath = new Map(expectedSourcePaths.map(([family, sourcePath, commit]) => [sourcePath, { family, commit }]));
  let sourcePathOK = Array.isArray(paths) && paths.length === expectedSourcePaths.length;
  for (const row of paths) {
    const expected = expectedByPath.get(row.path);
    if (!expected || seenPaths.has(row.path) || row.family !== expected.family || row.status !== "adopted" || row.source_commit_sha !== expected.commit) {
      sourcePathOK = false;
      break;
    }
    seenPaths.add(row.path);
  }
  assert(sourcePathOK, "source_paths exactly match the adopted registry", checks, "snapshot.source_paths");

  const families = new Set((snapshot.language_sdk_versions ?? []).map((row) => row.package_family));
  const missingFamilies = [...requiredPackageFamilies].filter((family) => !families.has(family));
  assert(missingFamilies.length === 0, `language SDK package-family rows are exhaustive${missingFamilies.length ? `; missing ${missingFamilies.join(", ")}` : ""}`, checks, "snapshot.package_families");
}

function validateDocs(checks) {
  const otelDoc = readText(otelDocPath);
  assert(otelDoc.includes("status: adopted/current"), "OpenTelemetry NLSpec front matter is adopted", checks, "docs.otel_status_frontmatter");
  assert(otelDoc.includes("Status: `adopted/current`."), "OpenTelemetry NLSpec status text is adopted", checks, "docs.otel_status_text");
  assert(!otelDoc.includes("## 16. Open decisions"), "Open decision section is absent", checks, "docs.no_open_decisions");
  const placeholderLines = otelDoc
    .split("\n")
    .filter((line) => /\bTODO\b|\bTBD\b/.test(line))
    .filter((line) => !line.includes("placeholder") && !line.includes("No `TODO`"));
  assert(placeholderLines.length === 0, "OpenTelemetry NLSpec contains no TODO/TBD placeholders", checks, "docs.no_todo_tbd");
  assert(otelDoc.includes(expectedDigest), "OpenTelemetry NLSpec contains the adopted semantic-convention digest", checks, "docs.digest_present");

  for (const dir of [
    "internal/testutil/golden/otel",
    "internal/testutil/golden/otel/source-snapshot",
    "internal/testutil/golden/otel/signals",
  ]) {
    assert(statSync(repoPath(dir)).isDirectory(), `${dir} exists`, checks, `golden.${dir}`);
  }
}

function validateErrorMapping(checks) {
  const publicCodes = publicErrorCodes(readText(core01Path));
  const mappedCounts = mappedErrorCodes(readText(otelDocPath));
  const missing = [...publicCodes].filter((code) => !mappedCounts.has(code));
  const duplicated = [...mappedCounts.entries()].filter(([, count]) => count !== 1).map(([code]) => code);
  const unknown = [...mappedCounts.keys()].filter((code) => !publicCodes.has(code));
  assert(missing.length === 0, `all Core 01 public error codes are mapped${missing.length ? `; missing ${missing.join(", ")}` : ""}`, checks, "errors.mapping_complete");
  assert(duplicated.length === 0, `no public error code is mapped more than once${duplicated.length ? `; duplicated ${duplicated.join(", ")}` : ""}`, checks, "errors.mapping_unique");
  assert(unknown.length === 0, `mapping contains no unknown public error code${unknown.length ? `; unknown ${unknown.join(", ")}` : ""}`, checks, "errors.mapping_known");
}

function targetDir() {
  const resultsRoot = process.env.CARTULARY_TEST_RESULTS_DIR || path.join(repoRoot, ".cartulary", "test-results");
  const runId = process.env.CARTULARY_TEST_RUN_ID || "manual";
  return path.join(resultsRoot, runId, "otel-conformance");
}

function writeSummary(checks, status) {
  const outDir = targetDir();
  mkdirSync(outDir, { recursive: true, mode: 0o700 });
  const summary = {
    schema_id: "cartulary.otel_conformance_summary.v1",
    target: "otel-conformance",
    status,
    checked_at: new Date().toISOString(),
    source_snapshot_path: snapshotPath,
    checks,
  };
  writeFileSync(path.join(outDir, "otel-conformance-summary.json"), `${JSON.stringify(summary, null, 2)}\n`);
}

function main() {
  const checks = [];
  try {
    validateSnapshot(readJSON(snapshotPath), checks);
    validateDocs(checks);
    validateErrorMapping(checks);
  } catch (error) {
    checks.push({ id: "otel_conformance.exception", status: "fail", message: error.message });
  }

  const status = checks.every((check) => check.status === "pass") ? "pass" : "fail";
  writeSummary(checks, status);
  if (status !== "pass") {
    for (const check of checks.filter((entry) => entry.status !== "pass")) {
      process.stderr.write(`${check.id}: ${check.message}\n`);
    }
    process.exit(1);
  }
}

main();
