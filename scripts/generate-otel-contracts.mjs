#!/usr/bin/env node

import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { mkdirSync, readFileSync, statSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptFile = fileURLToPath(import.meta.url);
const repoRoot = path.resolve(path.dirname(scriptFile), "..");

export const otelContractPaths = {
  sourceSnapshot: "contracts/otel/otel_source_snapshot.v1.json",
  generatedConstantsManifest: "contracts/otel/generated_constants_manifest.json",
  importBoundary: "contracts/otel/import_boundary.json",
};

export const otelGeneratorSourceRef = "scripts/generate-otel-contracts.mjs";

export const otelGeneratedConstantsProvenance = {
  generatorName: "cartulary-otel-contract-generator",
  generatorVersion: "1.0.0",
  inputModelDigest: "3f8f80a2ed04521dfe29e50fcddd7f7de70145a6aee01959f985a65fbb4c8632",
  outputPackageOrPath: "contracts/otel/semantic_conventions_constants.v1.json",
};

export const expectedBrowserRuntimeProbe = {
  make_target: "frontend-unit",
  evidence: "apps/web/src/app/otelBoundary.test.ts::OpenTelemetry browser boundary",
  state_sources: ["localStorage", "sessionStorage", "DOM attribute", "URL parameter", "globalThis"],
  forbidden_effects: ["telemetry_export_global", "remote_telemetry_request", "browser_config_authority"],
};

const sourceSnapshotTopLevelKeys = [
  "schema_id",
  "otel_spec_version",
  "otel_spec_ref",
  "otel_spec_commit_sha",
  "semconv_version",
  "semconv_ref",
  "semconv_commit_sha",
  "semconv_model_digest_algorithm",
  "semconv_model_digest",
  "sampler_profile_review_after",
  "sampler_profile_current_fractional",
  "probability_sampler_status",
  "semconv_generated_constants",
  "language_sdk_versions",
  "source_paths",
  "created_at",
  "created_by_tool",
];

const generatedConstantsKeys = [
  "source_kind",
  "generator_name",
  "generator_version",
  "generator_source_ref",
  "generator_source_sha",
  "input_model_digest",
  "output_package_or_path",
];

const generatedConstantsManifestKeys = [
  "schema_id",
  "source_kind",
  "generator_name",
  "generator_version",
  "generator_source_ref",
  "input_model_digest_algorithm",
  "input_model_digest",
  "output_path",
  "standard_attribute_allowlist_exceptions",
  "standard_metric_allowlist_exceptions",
  "status",
];

const importBoundaryTopLevelKeys = [
  "schema_id",
  "ordinary_go_source_roots",
  "ordinary_go_test_support_roots",
  "telemetry_bootstrap_roots",
  "hostile_fixture_roots",
  "allowed_ordinary_otel_imports",
  "forbidden_ordinary_otel_import_prefixes",
  "telemetry_bootstrap_import_notes",
  "allowed_cartulary_telemetry_accessor_import",
  "browser_package_manifest_roots",
  "forbidden_browser_package_patterns",
  "browser_source_roots",
  "browser_runtime_source_roots",
  "browser_built_bundle_roots",
  "browser_text_bundle_extensions",
  "browser_source_exclude_suffixes",
  "browser_runtime_probe",
  "non_transfer_absence_rules",
  "forbidden_browser_runtime_patterns",
  "observed_module_graph_notes",
];

const repoRuntimeOrGeneratedPrefixes = [
  ".cartulary/",
  ".cache/",
  "apps/web/dist/",
  "internal/gen/",
  "node_modules/",
  "packages/protocol-ts/src/generated/",
  "packages/ui-contracts/src/generated/",
  "tmp/",
];

function repoPath(root, relativePath) {
  return path.join(root, relativePath);
}

function readJSON(root, relativePath) {
  return JSON.parse(readFileSync(repoPath(root, relativePath), "utf8"));
}

function prettyJSON(value) {
  return `${JSON.stringify(value, null, 2)}\n`;
}

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

function sorted(value) {
  return [...value].sort();
}

function objectKeyErrors(value, keys, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return [`${label} must be an object`];
  }
  const actual = sorted(Object.keys(value));
  const expected = sorted(keys);
  if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    return [`${label} keys must be exactly ${expected.join(", ")}; saw ${actual.join(", ")}`];
  }
  return [];
}

export function gitBlobSHA(root, relativePath) {
  const content = readFileSync(repoPath(root, relativePath));
  return createHash("sha1")
    .update(`blob ${content.length}\0`)
    .update(content)
    .digest("hex");
}

function isRegularFile(root, relativePath) {
  try {
    return statSync(repoPath(root, relativePath)).isFile();
  } catch {
    return false;
  }
}

function isGitTracked(root, relativePath) {
  try {
    execFileSync("git", ["-C", root, "ls-files", "--error-unmatch", "--", relativePath], { stdio: "ignore" });
    return true;
  } catch {
    try {
      const output = execFileSync("git", ["-C", root, "ls-files", "--others", "--exclude-standard", "--", relativePath], {
        encoding: "utf8",
      });
      return output.split("\n").includes(relativePath);
    } catch {
      return false;
    }
  }
}

function repoRelativePathErrors(relativePath, label) {
  const errors = [];
  if (typeof relativePath !== "string" || relativePath.length === 0) {
    return [`${label} must be a non-empty repo-relative path`];
  }
  if (path.isAbsolute(relativePath)) {
    errors.push(`${label} must not be absolute`);
  }
  if (relativePath.includes("\\") || relativePath.split("/").includes("..")) {
    errors.push(`${label} must be normalized POSIX repo-relative path without traversal`);
  }
  if (repoRuntimeOrGeneratedPrefixes.some((prefix) => relativePath === prefix.slice(0, -1) || relativePath.startsWith(prefix))) {
    errors.push(`${label} must not point into runtime, cache, or generated output directories`);
  }
  return errors;
}

export function validateOtelGeneratorReference({ root = repoRoot, generatorSourceRef, generatorSourceSHA, requireTracked = false }) {
  const errors = repoRelativePathErrors(generatorSourceRef, "semconv_generated_constants.generator_source_ref");
  if (errors.length > 0) {
    return errors;
  }
  if (!isRegularFile(root, generatorSourceRef)) {
    errors.push(`semconv_generated_constants.generator_source_ref file is missing: ${generatorSourceRef}`);
    return errors;
  }
  if (requireTracked && !isGitTracked(root, generatorSourceRef)) {
    errors.push(`semconv_generated_constants.generator_source_ref is not tracked by git: ${generatorSourceRef}`);
  }
  const actualSHA = gitBlobSHA(root, generatorSourceRef);
  if (generatorSourceSHA !== actualSHA) {
    errors.push(`semconv_generated_constants.generator_source_sha must match ${generatorSourceRef}; expected ${actualSHA}, saw ${generatorSourceSHA}`);
  }
  return errors;
}

export function parseEvidenceReference(evidence) {
  if (typeof evidence !== "string" || evidence.length === 0) {
    return { path: "", name: "", errors: ["browser_runtime_probe.evidence must be a non-empty path::test name reference"] };
  }
  const separator = evidence.indexOf("::");
  if (separator <= 0 || separator !== evidence.lastIndexOf("::") || separator === evidence.length - 2) {
    return { path: "", name: "", errors: [`browser_runtime_probe.evidence must use path::test name format: ${evidence}`] };
  }
  const evidencePath = evidence.slice(0, separator);
  const name = evidence.slice(separator + 2);
  const errors = [
    ...repoRelativePathErrors(evidencePath, "browser_runtime_probe.evidence path"),
  ];
  return { path: evidencePath, name, errors };
}

export function validateBrowserRuntimeProbe(root, probe, { requireEvidenceFile = true } = {}) {
  const errors = [];
  if (!probe || typeof probe !== "object" || Array.isArray(probe)) {
    return ["browser_runtime_probe must be an object"];
  }
  if (probe.make_target !== expectedBrowserRuntimeProbe.make_target) {
    errors.push(`browser_runtime_probe.make_target must be ${expectedBrowserRuntimeProbe.make_target}`);
  }
  for (const key of ["state_sources", "forbidden_effects"]) {
    if (JSON.stringify(probe[key] ?? []) !== JSON.stringify(expectedBrowserRuntimeProbe[key])) {
      errors.push(`browser_runtime_probe.${key} must match the adopted OTel browser-boundary registry`);
    }
  }
  const evidence = parseEvidenceReference(probe.evidence);
  errors.push(...evidence.errors);
  if (evidence.errors.length === 0 && requireEvidenceFile) {
    if (!isRegularFile(root, evidence.path)) {
      errors.push(`browser_runtime_probe.evidence file is missing: ${evidence.path}`);
    } else {
      const text = readFileSync(repoPath(root, evidence.path), "utf8");
      if (!text.includes(evidence.name)) {
        errors.push(`browser_runtime_probe.evidence test name is absent from ${evidence.path}: ${evidence.name}`);
      }
    }
  }
  return errors;
}

function stringArrayErrors(value, label) {
  if (!Array.isArray(value) || value.some((entry) => typeof entry !== "string" || entry.length === 0)) {
    return [`${label} must be an array of non-empty strings`];
  }
  return [];
}

function validateImportBoundaryShape(root, boundary, { requireEvidenceFile = true } = {}) {
  const errors = objectKeyErrors(boundary, importBoundaryTopLevelKeys, "import_boundary");
  if (errors.length > 0) {
    return errors;
  }
  if (boundary.schema_id !== "cartulary.otel_import_boundary.v1") {
    errors.push("import_boundary.schema_id must be cartulary.otel_import_boundary.v1");
  }
  for (const key of [
    "ordinary_go_source_roots",
    "ordinary_go_test_support_roots",
    "telemetry_bootstrap_roots",
    "hostile_fixture_roots",
    "allowed_ordinary_otel_imports",
    "forbidden_ordinary_otel_import_prefixes",
    "browser_package_manifest_roots",
    "forbidden_browser_package_patterns",
    "browser_source_roots",
    "browser_runtime_source_roots",
    "browser_built_bundle_roots",
    "browser_text_bundle_extensions",
    "browser_source_exclude_suffixes",
    "non_transfer_absence_rules",
  ]) {
    errors.push(...stringArrayErrors(boundary[key], `import_boundary.${key}`));
  }
  errors.push(...validateBrowserRuntimeProbe(root, boundary.browser_runtime_probe, { requireEvidenceFile }));
  return errors;
}

export function validateOtelImportBoundaryContractShape(root, boundary, options = {}) {
  return validateImportBoundaryShape(root, boundary, options);
}

export function buildOtelContracts(root = repoRoot) {
  const snapshot = clone(readJSON(root, otelContractPaths.sourceSnapshot));
  const manifest = clone(readJSON(root, otelContractPaths.generatedConstantsManifest));
  const boundary = clone(readJSON(root, otelContractPaths.importBoundary));
  if (!isRegularFile(root, otelGeneratorSourceRef)) {
    throw new Error(`OTel contract generator source is missing: ${otelGeneratorSourceRef}`);
  }
  const generatorSourceSHA = gitBlobSHA(root, otelGeneratorSourceRef);

  snapshot.semconv_generated_constants = {
    source_kind: "repo_codegen",
    generator_name: otelGeneratedConstantsProvenance.generatorName,
    generator_version: otelGeneratedConstantsProvenance.generatorVersion,
    generator_source_ref: otelGeneratorSourceRef,
    generator_source_sha: generatorSourceSHA,
    input_model_digest: otelGeneratedConstantsProvenance.inputModelDigest,
    output_package_or_path: otelGeneratedConstantsProvenance.outputPackageOrPath,
  };

  manifest.generator_name = otelGeneratedConstantsProvenance.generatorName;
  manifest.generator_version = otelGeneratedConstantsProvenance.generatorVersion;
  manifest.generator_source_ref = otelGeneratorSourceRef;

  boundary.browser_runtime_probe = clone(expectedBrowserRuntimeProbe);

  return {
    [otelContractPaths.sourceSnapshot]: snapshot,
    [otelContractPaths.generatedConstantsManifest]: manifest,
    [otelContractPaths.importBoundary]: boundary,
  };
}

export function validateBuiltOtelContracts(root, contracts, { requireTrackedGenerator = false } = {}) {
  const errors = [];
  const snapshot = contracts[otelContractPaths.sourceSnapshot];
  const manifest = contracts[otelContractPaths.generatedConstantsManifest];
  const boundary = contracts[otelContractPaths.importBoundary];

  errors.push(...objectKeyErrors(snapshot, sourceSnapshotTopLevelKeys, "source_snapshot"));
  if (snapshot?.schema_id !== "cartulary.otel_source_snapshot.v1") {
    errors.push("source_snapshot.schema_id must be cartulary.otel_source_snapshot.v1");
  }
  errors.push(...objectKeyErrors(snapshot?.semconv_generated_constants, generatedConstantsKeys, "source_snapshot.semconv_generated_constants"));
  errors.push(...validateOtelGeneratorReference({
    root,
    generatorSourceRef: snapshot?.semconv_generated_constants?.generator_source_ref,
    generatorSourceSHA: snapshot?.semconv_generated_constants?.generator_source_sha,
    requireTracked: requireTrackedGenerator,
  }));

  errors.push(...objectKeyErrors(manifest, generatedConstantsManifestKeys, "generated_constants_manifest"));
  if (manifest?.schema_id !== "cartulary.otel_generated_constants_manifest.v1") {
    errors.push("generated_constants_manifest.schema_id must be cartulary.otel_generated_constants_manifest.v1");
  }
  if (manifest?.generator_source_ref !== otelGeneratorSourceRef) {
    errors.push(`generated_constants_manifest.generator_source_ref must be ${otelGeneratorSourceRef}`);
  }

  errors.push(...validateImportBoundaryShape(root, boundary, { requireEvidenceFile: true }));
  return errors;
}

function expectedContractTextByPath(root) {
  const contracts = buildOtelContracts(root);
  const errors = validateBuiltOtelContracts(root, contracts);
  if (errors.length > 0) {
    throw new Error(`generated OTel contracts are invalid:\n${errors.join("\n")}`);
  }
  return Object.fromEntries(Object.entries(contracts).map(([relativePath, value]) => [relativePath, prettyJSON(value)]));
}

function writeContracts(root) {
  for (const [relativePath, text] of Object.entries(expectedContractTextByPath(root))) {
    mkdirSync(path.dirname(repoPath(root, relativePath)), { recursive: true });
    writeFileSync(repoPath(root, relativePath), text);
  }
}

function checkContracts(root) {
  const currentContracts = {
    [otelContractPaths.sourceSnapshot]: readJSON(root, otelContractPaths.sourceSnapshot),
    [otelContractPaths.generatedConstantsManifest]: readJSON(root, otelContractPaths.generatedConstantsManifest),
    [otelContractPaths.importBoundary]: readJSON(root, otelContractPaths.importBoundary),
  };
  const currentErrors = validateBuiltOtelContracts(root, currentContracts);
  if (currentErrors.length > 0) {
    throw new Error(`current OTel contracts are invalid:\n${currentErrors.join("\n")}`);
  }
  const stale = [];
  for (const [relativePath, expectedText] of Object.entries(expectedContractTextByPath(root))) {
    const actualText = readFileSync(repoPath(root, relativePath), "utf8");
    if (actualText !== expectedText) {
      stale.push(relativePath);
    }
  }
  if (stale.length > 0) {
    throw new Error(`OTel contract drift detected: ${stale.join(", ")}; run make generate`);
  }
}

function usage() {
  throw new Error("usage: generate-otel-contracts.mjs --write|--check [--root <path>]");
}

function parseArgs(argv) {
  const options = { mode: "", root: repoRoot };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--write" || arg === "--check") {
      if (options.mode) {
        usage();
      }
      options.mode = arg.slice(2);
      continue;
    }
    if (arg === "--root") {
      options.root = path.resolve(argv[index + 1] ?? "");
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.mode || !options.root) {
    usage();
  }
  return options;
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  if (options.mode === "write") {
    writeContracts(options.root);
    return;
  }
  checkContracts(options.root);
}

if (process.argv[1] && path.resolve(process.argv[1]) === scriptFile) {
  try {
    main();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(`generate-otel-contracts failed: ${message}`);
    process.exit(1);
  }
}
