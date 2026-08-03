#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import {
  copyFileSync,
  existsSync,
  mkdirSync,
  readFileSync,
  statSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { validateSchemaSync } from "../harness/contract/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../..");

const legacyStem = "min" + "io";
const occurrenceTokens = Object.freeze([
  "Min" + "IO",
  legacyStem,
  "MIN" + "IO",
  `${legacyStem}_bucket`,
  `${legacyStem}_endpoint`,
]);
const sdkModule = `github.com/${legacyStem}/${legacyStem}-go/v7`;
const sdkCredentialModule = `${sdkModule}/pkg/credentials`;
const serverImageNeedle = `${legacyStem}/${legacyStem}`;
const generatedPolicyPath = "tools/generated_artifact_policy.json";
const executableInputPolicyPath = "tools/executable_input_policy.json";
const defaultClassificationPath =
  "tools/seaweedfs_occurrence_classifications.json";
const defaultReleaseArtifactDir = ".cartulary/release-artifacts";
const releaseArtifactSubdir = "seaweedfs";
const seaweedfsCompatibilityTarget = "seaweedfs-compatibility";
const seaweedfsCompatibilityReportName = "object-store-compatibility-report.json";
const compatibilityUnitResultName = "target-seaweedfs-compatibility.json";
const releaseGateRows = Object.freeze([
  "SWFS-AC-001",
  "SWFS-AC-002",
  "SWFS-AC-006",
  "SWFS-AC-021",
  "SWFS-AC-022",
  "SWFS-AC-023",
  "SWFS-AC-024",
  "SWFS-AC-025",
]);
const seaweedfsCompatibilityCaseIds = Object.freeze(
  Array.from({ length: 14 }, (_, index) => `SWFS-COMP-${String(index + 1).padStart(3, "0")}`),
);
const objectStoreThreatPolicyPath =
  "contracts/object-store/release-threat-policy.v1.json";
const recoveryContractRegistryPath = "contracts/recovery/index.json";
const allowedClassifications = new Set([
  "sdk_only",
  "legacy_external_endpoint",
  "historical_changelog",
  "invalid",
  "unclassified",
]);
const restrictedInputPrefixes = JSON.parse(
  readFileSync(path.join(repoRoot, executableInputPolicyPath), "utf8"),
).restricted_roots.map((root) => `${root}/`);
const alwaysExcludedPrefixes = Object.freeze([
  ".git/",
  ".cartulary/",
  "tmp/",
  "node_modules/",
  ".pnpm-store/",
  "apps/web/dist/",
  "coverage/",
  "playwright-report/",
  "test-results/",
  "dist/",
  "build/",
  "out/",
  "target/",
  ...restrictedInputPrefixes,
]);
const releaseManifestPathPatterns = Object.freeze([
  /^docker-compose(?:\.[^.]+)?\.ya?ml$/,
  /^compose(?:\.[^.]+)?\.ya?ml$/,
  /^configs\/.+\.(?:json|ya?ml|toml)$/,
  /^\.github\/workflows\/.+\.ya?ml$/,
  /^Dockerfile(?:\..*)?$/,
  /^docker\/.+/,
  /^deploy(?:ment)?s?\//,
  /^k8s\//,
  /^helm\//,
  /^tools\/(?:execution_topology_manifest|task_surface_manifest|scheduler_manifest|browser_e2e_batch_manifest)\.json$/,
]);
const runtimeCodePrefixes = Object.freeze([
  "cmd/",
  "internal/app/",
  "internal/modules/",
  "internal/platform/",
]);
const runtimeCodeExcludedPrefixes = Object.freeze([
  "internal/testutil/",
  "internal/gen/",
]);
const allowedSdkImportPrefixes = Object.freeze([
  "internal/platform/objectstore/",
  "internal/testutil/s3test/",
  "tools/objectstoreprobe/",
]);
const secretNeedles = Object.freeze([
  "seaweedfs-secret",
  "min" + "io-secret",
  "supersecret",
  "secret-access-key",
  "secret_key=",
  "access_key=",
  "aws_secret_access_key=",
  "aws_access_key_id=",
  "CARTULARY_S3_SECRET_KEY=",
  "CARTULARY_S3_ACCESS_KEY=",
]);
const publicLeakNeedles = Object.freeze([
  "s3://",
  "private-bucket",
  "source.internal",
  "target.internal",
  "AWS_SECRET_ACCESS_KEY",
  "AWS_ACCESS_KEY_ID",
  "CARTULARY_S3_SECRET_KEY",
  "CARTULARY_S3_ACCESS_KEY",
]);
const publicRawStorageFieldNames = new Set([
  "bucket",
  "source_bucket",
  "target_bucket",
  "s3_bucket",
  "object_key",
  "source_key",
  "target_key",
  "storage_ref",
  "endpoint",
  "source_endpoint",
  "target_endpoint",
]);

function repoPath(...segments) {
  return path.join(repoRoot, ...segments);
}

function inputPath(value) {
  return path.isAbsolute(value) ? value : repoPath(value);
}

function displayPath(value) {
  const resolved = inputPath(value);
  return resolved.startsWith(`${repoRoot}${path.sep}`) ? relPath(resolved) : resolved;
}

function relPath(absPath) {
  return path.relative(repoRoot, absPath).replaceAll(path.sep, "/");
}

function readJSON(rel) {
  return JSON.parse(readFileSync(repoPath(rel), "utf8"));
}

function writeJSON(absPath, value) {
  mkdirSync(path.dirname(absPath), { recursive: true });
  writeFileSync(absPath, `${JSON.stringify(value, null, 2)}\n`);
}

function runGit(args, options = {}) {
  const result = spawnSync("git", args, {
    cwd: repoRoot,
    encoding: options.encoding ?? "utf8",
    maxBuffer: options.maxBuffer ?? 64 * 1024 * 1024,
  });
  if (result.status !== 0) {
    throw new Error(`git ${args.join(" ")} failed: ${result.stderr || result.stdout}`);
  }
  return result.stdout;
}

function repoCommit() {
  return runGit(["rev-parse", "HEAD"]).trim();
}

function trackedFiles() {
  return runGit(["ls-files", "-z"], { encoding: "buffer" })
    .toString("utf8")
    .split("\0")
    .filter(Boolean)
    .filter((rel) => {
      const absolute = repoPath(rel);
      return existsSync(absolute) && statSync(absolute).isFile();
    })
    .sort();
}

function loadGeneratedPolicy(policyRel = generatedPolicyPath) {
  if (!existsSync(repoPath(policyRel))) {
    return { roots: [], files: [] };
  }
  const policy = readJSON(policyRel);
  return {
    roots: (policy.generated_roots ?? []).map((entry) => normalizePath(entry.path)),
    files: (policy.generated_files ?? []).map((entry) => normalizePath(entry.path)),
  };
}

function normalizePath(value) {
  return String(value ?? "").replaceAll("\\", "/").replace(/^\.\//, "");
}

function isPolicyGenerated(rel, policy) {
  const normalized = normalizePath(rel);
  if (policy.files.includes(normalized)) {
    return true;
  }
  return policy.roots.some((root) => normalized === root || normalized.startsWith(`${root}/`));
}

function isDefaultExcluded(rel, policy) {
  const normalized = normalizePath(rel);
  if (alwaysExcludedPrefixes.some((prefix) => normalized.startsWith(prefix))) {
    return true;
  }
  return isPolicyGenerated(normalized, policy);
}

function isTextBuffer(buffer) {
  if (buffer.includes(0)) {
    return false;
  }
  return true;
}

function globToRegExp(glob) {
  let source = "^";
  for (let index = 0; index < glob.length; index += 1) {
    const char = glob[index];
    const next = glob[index + 1];
    if (char === "*" && next === "*") {
      source += ".*";
      index += 1;
    } else if (char === "*") {
      source += "[^/]*";
    } else if (char === "?") {
      source += "[^/]";
    } else {
      source += char.replace(/[|\\{}()[\]^$+?.]/g, "\\$&");
    }
  }
  source += "$";
  return new RegExp(source);
}

function loadClassificationRules(classificationRel = defaultClassificationPath) {
  const manifest = readJSON(classificationRel);
  if (manifest.schema_id !== "cartulary.seaweedfs_occurrence_classifications.v1") {
    throw new Error(`${classificationRel} has unexpected schema_id ${manifest.schema_id}`);
  }
  return (manifest.rules ?? []).map((rule, index) => {
    if (!allowedClassifications.has(rule.classification)) {
      throw new Error(`classification rule ${rule.id ?? index} has invalid classification`);
    }
    if (rule.classification !== "invalid") {
      for (const key of ["owner", "rationale"]) {
        if (typeof rule[key] !== "string" || rule[key].trim() === "") {
          throw new Error(`classification rule ${rule.id ?? index} missing ${key}`);
        }
      }
    }
    return {
      ...rule,
      pathRegexps: (rule.path_globs ?? []).map(globToRegExp),
      lineRegex: rule.line_regex ? new RegExp(rule.line_regex) : null,
      contains: rule.contains ?? null,
    };
  });
}

function findOccurrencesInText(rel, text) {
  const occurrences = [];
  const lines = text.split(/\n/);
  for (let lineIndex = 0; lineIndex < lines.length; lineIndex += 1) {
    const line = lines[lineIndex];
    for (const token of occurrenceTokens) {
      let start = line.indexOf(token);
      while (start !== -1) {
        occurrences.push({
          path: rel,
          line: lineIndex + 1,
          column: start + 1,
          token,
          line_excerpt: line.trim(),
        });
        start = line.indexOf(token, start + token.length);
      }
    }
  }
  return occurrences;
}

function classifyOccurrence(occurrence, rules) {
  for (const rule of rules) {
    const pathMatches =
      rule.pathRegexps.length === 0 ||
      rule.pathRegexps.some((regexp) => regexp.test(occurrence.path));
    if (!pathMatches) {
      continue;
    }
    if (rule.contains && !occurrence.line_excerpt.includes(rule.contains)) {
      continue;
    }
    if (rule.lineRegex && !rule.lineRegex.test(occurrence.line_excerpt)) {
      continue;
    }
    return {
      classification: rule.classification,
      owner: rule.owner ?? null,
      rationale: rule.rationale ?? null,
      classification_rule_id: rule.id ?? null,
    };
  }
  return {
    classification: "unclassified",
    owner: null,
    rationale: null,
    classification_rule_id: null,
  };
}

function sortOccurrences(occurrences) {
  return occurrences.sort((a, b) => {
    if (a.path !== b.path) return a.path.localeCompare(b.path);
    if (a.line !== b.line) return a.line - b.line;
    if (a.column !== b.column) return a.column - b.column;
    return a.token.localeCompare(b.token);
  });
}

function occurrenceResult(occurrences) {
  const invalid = occurrences.filter((entry) => entry.classification === "invalid").length;
  const unclassified = occurrences.filter((entry) => entry.classification === "unclassified").length;
  const missingOwner = occurrences.filter(
    (entry) =>
      entry.classification !== "invalid" &&
      (!entry.owner || !entry.rationale),
  ).length;
  return invalid === 0 && unclassified === 0 && missingOwner === 0 ? "pass" : "fail";
}

export function buildOccurrenceInventoryFromEntries({
  entries,
  rules,
  repoCommitValue = "test-commit",
  scannedAt = new Date().toISOString(),
  includedPaths = null,
  excludedPaths = [],
}) {
  const occurrences = [];
  for (const entry of entries) {
    const text = typeof entry.text === "string" ? entry.text : entry.buffer.toString("utf8");
    for (const occurrence of findOccurrencesInText(entry.path, text)) {
      const classified = {
        ...occurrence,
        ...classifyOccurrence(occurrence, rules),
      };
      delete classified.line_excerpt;
      occurrences.push({
        ...classified,
      });
    }
  }
  sortOccurrences(occurrences);
  return {
    schema_id: "cartulary.seaweedfs_occurrence_inventory.v1",
    scanned_at: scannedAt,
    repo_commit: repoCommitValue,
    scan_scope: {
      included_paths: includedPaths ?? entries.map((entry) => entry.path).sort(),
      excluded_paths: excludedPaths.sort(),
    },
    tokens: [...occurrenceTokens],
    occurrences,
    result: occurrenceResult(occurrences),
  };
}

function buildOccurrenceInventory({
  files,
  policy,
  rules,
  repoCommitValue,
  scannedAt,
}) {
  const entries = [];
  const excluded = [];
  for (const rel of files) {
    if (isDefaultExcluded(rel, policy)) {
      excluded.push(rel);
      continue;
    }
    const abs = repoPath(rel);
    if (!existsSync(abs) || !statSync(abs).isFile()) {
      excluded.push(rel);
      continue;
    }
    const buffer = readFileSync(abs);
    if (!isTextBuffer(buffer)) {
      excluded.push(rel);
      continue;
    }
    entries.push({ path: rel, buffer });
  }
  return buildOccurrenceInventoryFromEntries({
    entries,
    rules,
    repoCommitValue,
    scannedAt,
    includedPaths: entries.map((entry) => entry.path),
    excludedPaths: excluded,
  });
}

function lineFindings(rel, text, checks) {
  const findings = [];
  const lines = text.split(/\n/);
  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index];
    for (const check of checks) {
      if (check.pattern.test(line)) {
        findings.push({
          check_id: check.id,
          path: rel,
          line: index + 1,
          column: Math.max(1, line.search(check.pattern) + 1),
          severity: check.severity ?? "blocking",
          message: check.message,
          excerpt: line.trim(),
        });
      }
    }
  }
  return findings;
}

function buildReleaseManifestExposure({
  files,
  repoCommitValue,
  generatedAt,
  readFile = (rel) => readFileSync(repoPath(rel), "utf8"),
}) {
  const manifestPaths = files.filter((file) =>
    releaseManifestPathPatterns.some((pattern) => pattern.test(file)),
  );
  const serverImagePattern = new RegExp(
    `(?:docker\\.io/|quay\\.io/|ghcr\\.io/)?${serverImageNeedle.replace("/", "\\/")}(?::|@|\\b)`,
    "i",
  );
  const manifestChecks = [
    {
      id: "legacy-server-image",
      pattern: serverImagePattern,
      message: "release/default manifest references the retired object-store server image",
    },
    {
      id: "legacy-server-command",
      pattern: new RegExp(`\\b${legacyStem}\\s+server\\b`, "i"),
      message: "release/default manifest starts the retired object-store server command",
    },
    {
      id: "forbidden-seaweedfs-port",
      pattern:
        /(?:^|[\s"':])(?:9333|19333|8888|18888|8080|18080|7333|23646|6060)\s*:\s*(?:9333|19333|8888|18888|8080|18080|7333|23646|6060)(?:$|[\s"',])/,
      message: "manifest publishes a SeaweedFS admin, master, filer, volume, WebDAV, console, or debug surface",
    },
    {
      id: "wildcard-direct-upload-cors",
      pattern:
        /(?:allowedOrigins|allowed_origins|CORS_ALLOWED_ORIGINS|cors_allowed_origins|Access-Control-Allow-Origin)\s*[:=]\s*["']?\*/i,
      message: "direct-upload CORS allows wildcard origin",
    },
  ];
  const runtimeChecks = [
    {
      id: "runtime-seaweedfs-admin-api",
      pattern:
        /(?:\/dir\/status|\/cluster\/status|\/vol\/status|\/stats\/counter|\/admin\/|master\.port|filer\.port|volume\.port)/i,
      message: "runtime product code references a SeaweedFS admin API or admin port control",
    },
  ];
  const manifestFindings = manifestPaths.flatMap((rel) => lineFindings(rel, readFile(rel), manifestChecks));
  const runtimePaths = files.filter(
    (file) =>
      runtimeCodePrefixes.some((prefix) => file.startsWith(prefix)) &&
      !runtimeCodeExcludedPrefixes.some((prefix) => file.startsWith(prefix)) &&
      /\.(?:go|ts|tsx|js|mjs|sh|toml|json|ya?ml)$/.test(file),
  );
  const runtimeFindings = runtimePaths.flatMap((rel) => lineFindings(rel, readFile(rel), runtimeChecks));
  const checks = [
    {
      check_id: "release-manifest-forbidden-surfaces",
      result: manifestFindings.filter((finding) => finding.check_id === "forbidden-seaweedfs-port").length === 0 ? "pass" : "fail",
    },
    {
      check_id: "release-manifest-wildcard-cors",
      result: manifestFindings.filter((finding) => finding.check_id === "wildcard-direct-upload-cors").length === 0 ? "pass" : "fail",
    },
    {
      check_id: "release-manifest-retired-server-absence",
      result: manifestFindings.filter((finding) => finding.check_id.startsWith("legacy-server")).length === 0 ? "pass" : "fail",
    },
    {
      check_id: "runtime-admin-api-absence",
      result: runtimeFindings.length === 0 ? "pass" : "fail",
    },
  ];
  const findings = [...manifestFindings, ...runtimeFindings];
  return {
    schema_id: "cartulary.seaweedfs_release_manifest_exposure.v1",
    generated_at: generatedAt,
    repo_commit: repoCommitValue,
    scanned_paths: {
      release_manifests: manifestPaths.sort(),
      runtime_product_code: runtimePaths.sort(),
    },
    checks,
    findings,
    result: checks.every((check) => check.result === "pass") && findings.length === 0 ? "pass" : "fail",
  };
}

function findImportSites(files, readFile = (rel) => readFileSync(repoPath(rel), "utf8")) {
  const importSites = [];
  for (const rel of files.filter((file) => file.endsWith(".go"))) {
    const text = readFile(rel);
    const lines = text.split(/\n/);
    for (let index = 0; index < lines.length; index += 1) {
      const line = lines[index];
      for (const moduleName of [sdkModule, sdkCredentialModule]) {
        const column = line.indexOf(moduleName);
        if (column !== -1) {
          const allowed = allowedSdkImportPrefixes.some((prefix) => rel.startsWith(prefix));
          importSites.push({
            path: rel,
            line: index + 1,
            column: column + 1,
            module: moduleName,
            boundary: allowed ? allowedSdkImportPrefixes.find((prefix) => rel.startsWith(prefix)) : null,
            allowed,
          });
        }
      }
    }
  }
  return importSites.sort((a, b) => a.path.localeCompare(b.path) || a.line - b.line || a.column - b.column);
}

function buildDependencyBoundary({
  files,
  repoCommitValue,
  generatedAt,
  sbom = null,
  licenseReport = null,
  readFile = (rel) => readFileSync(repoPath(rel), "utf8"),
  goModText = existsSync(repoPath("go.mod")) ? readFileSync(repoPath("go.mod"), "utf8") : "",
}) {
  const importSites = findImportSites(files, readFile);
  const modulePresent = goModText.includes(sdkModule);
  const disallowedImports = importSites.filter((site) => !site.allowed);
  const releaseServerComponents = [];
  for (const source of [sbom, licenseReport]) {
    if (!source) continue;
    for (const match of findServerArtifactStrings(source)) {
      releaseServerComponents.push(match);
    }
  }
  const result =
    (!modulePresent || importSites.length > 0) &&
    disallowedImports.length === 0 &&
    releaseServerComponents.length === 0
      ? "pass"
      : "fail";
  return {
    schema_id: "cartulary.seaweedfs_dependency_boundary.v1",
    generated_at: generatedAt,
    repo_commit: repoCommitValue,
    allowed_boundaries: [...allowedSdkImportPrefixes],
    sdk_dependency: {
      module: sdkModule,
      present_in_go_mod: modulePresent,
      classification: modulePresent ? "sdk_only" : "absent",
      import_sites: importSites,
      disallowed_imports: disallowedImports,
    },
    release_server_artifacts: releaseServerComponents,
    result,
  };
}

function findServerArtifactStrings(value, trail = []) {
  const findings = [];
  if (typeof value === "string") {
    const lower = value.toLowerCase();
    const serverImageCoordinate = new RegExp(
      `(?:^|[/:@])${legacyStem}\\/${legacyStem}(?::|@|$|[?#])`,
      "i",
    );
    if (
      serverImageCoordinate.test(lower) ||
      lower.includes(`quay.io/${legacyStem}`) ||
      lower.includes(`docker.io/${legacyStem}`) ||
      lower.includes(`pkg:docker/${legacyStem}`)
    ) {
      findings.push({ json_path: trail.join(".") || "$", value });
    }
    return findings;
  }
  if (Array.isArray(value)) {
    for (let index = 0; index < value.length; index += 1) {
      findings.push(...findServerArtifactStrings(value[index], [...trail, String(index)]));
    }
    return findings;
  }
  if (value && typeof value === "object") {
    for (const [key, child] of Object.entries(value)) {
      findings.push(...findServerArtifactStrings(child, [...trail, key]));
    }
  }
  return findings;
}

function buildSbomLicenseClassification({
  repoCommitValue,
  generatedAt,
  sbomPath,
  licensePath,
  sbom = existsSync(inputPath(sbomPath)) ? JSON.parse(readFileSync(inputPath(sbomPath), "utf8")) : null,
  licenseReport = existsSync(inputPath(licensePath)) ? JSON.parse(readFileSync(inputPath(licensePath), "utf8")) : null,
}) {
  const missing = [];
  if (!sbom) missing.push(sbomPath);
  if (!licenseReport) missing.push(licensePath);
  const sdkStrings = [];
  for (const source of [sbom, licenseReport]) {
    if (!source) continue;
    collectStrings(source, [], sdkModule, sdkStrings);
  }
  const serverArtifacts = [
    ...(sbom ? findServerArtifactStrings(sbom) : []),
    ...(licenseReport ? findServerArtifactStrings(licenseReport) : []),
  ];
  const result = missing.length === 0 && serverArtifacts.length === 0 ? "pass" : "fail";
  return {
    schema_id: "cartulary.seaweedfs_sbom_license_classification.v1",
    generated_at: generatedAt,
    repo_commit: repoCommitValue,
    inputs: {
      sbom: displayPath(sbomPath),
      license_report: displayPath(licensePath),
      missing: missing.map(displayPath),
    },
    sdk_dependency: {
      module: sdkModule,
      classification: sdkStrings.length > 0 ? "sdk_only" : "absent",
      references: sdkStrings,
    },
    release_server_artifacts: serverArtifacts,
    result,
  };
}

function collectStrings(value, trail, needle, out) {
  if (typeof value === "string") {
    if (value.includes(needle)) {
      out.push({ json_path: trail.join(".") || "$", value });
    }
    return;
  }
  if (Array.isArray(value)) {
    value.forEach((child, index) => {
      collectStrings(child, [...trail, String(index)], needle, out);
    });
    return;
  }
  if (value && typeof value === "object") {
    for (const [key, child] of Object.entries(value)) {
      collectStrings(child, [...trail, key], needle, out);
    }
  }
}

function buildThreatModelCoverage({ repoCommitValue, generatedAt }) {
  const policy = readJSON(objectStoreThreatPolicyPath);
  validateSchemaSync("cartulary.object_store_release_threat_policy.v1", policy);
  const rows = policy.threat_model.map((row) => ({
    stride: row.stride,
    covered: row.controls.length > 0 && row.verification_hooks.length > 0,
    controls: row.controls,
    verification_hooks: row.verification_hooks,
    owner_verifications: [policy.verification_id],
  }));
  return {
    schema_id: "cartulary.seaweedfs_threat_model_coverage.v1",
    generated_at: generatedAt,
    repo_commit: repoCommitValue,
    owner_id: policy.owner_id,
    owner_verification: policy.verification_id,
    rows,
    result: rows.every((row) => row.covered) ? "pass" : "fail",
  };
}

function buildStorageRefOwnerCoverage({
  repoCommitValue = "test-commit",
  generatedAt = new Date().toISOString(),
} = {}) {
  const policy = readJSON(objectStoreThreatPolicyPath);
  validateSchemaSync("cartulary.object_store_release_threat_policy.v1", policy);
  const checks = policy.storage_ref_controls.map((control) => ({
    check_id: control.check_id,
    assertion: control.assertion,
    result: control.status === "active" ? "pass" : "fail",
    owner_verifications: [policy.verification_id],
  }));
  return {
    schema_id: "cartulary.seaweedfs_storage_ref_owner_coverage.v1",
    generated_at: generatedAt,
    repo_commit: repoCommitValue,
    owner_id: policy.owner_id,
    owner_verification: policy.verification_id,
    checks,
    result: checks.every((check) => check.result === "pass") ? "pass" : "fail",
  };
}

function readArtifactJSON({ artifactPath, schemaID, findings }) {
  if (!artifactPath) {
    findings.push({
      check_id: "artifact-path-present",
      severity: "blocking",
      message: "artifact path is required",
    });
    return null;
  }
  const abs = inputPath(artifactPath);
  if (!existsSync(abs)) {
    findings.push({
      check_id: "artifact-present",
      severity: "blocking",
      path: displayPath(artifactPath),
      message: "artifact is missing",
    });
    return null;
  }
  try {
    const parsed = JSON.parse(readFileSync(abs, "utf8"));
    if (schemaID && parsed?.schema_id !== schemaID) {
      findings.push({
        check_id: "schema-id",
        severity: "blocking",
        path: displayPath(artifactPath),
        message: `expected schema_id ${schemaID}`,
        actual: parsed?.schema_id ?? null,
      });
    }
    return parsed;
  } catch (error) {
    findings.push({
      check_id: "json-parse",
      severity: "blocking",
      path: displayPath(artifactPath),
      message: error instanceof Error ? error.message : String(error),
    });
    return null;
  }
}

function defaultCompatibilityReportPath() {
  if (process.env.SEAWEEDFS_COMPATIBILITY_REPORT) {
    return process.env.SEAWEEDFS_COMPATIBILITY_REPORT;
  }
  if (process.env.CARTULARY_TEST_RESULTS_DIR && process.env.CARTULARY_TEST_RUN_ID) {
    return path.join(
      process.env.CARTULARY_TEST_RESULTS_DIR,
      process.env.CARTULARY_TEST_RUN_ID,
      seaweedfsCompatibilityTarget,
      seaweedfsCompatibilityReportName,
    );
  }
  return path.join(defaultReleaseArtifactDir, releaseArtifactSubdir, seaweedfsCompatibilityReportName);
}

function expectedCurrentCompatibilityReportPath({ currentResultsDir, currentRunId }) {
  if (!currentResultsDir || !currentRunId) {
    return null;
  }
  return path.join(
    currentResultsDir,
    currentRunId,
    seaweedfsCompatibilityTarget,
    seaweedfsCompatibilityReportName,
  );
}

function expectedCurrentUnitResultPath({ currentResultsDir, currentRunId }) {
  if (!currentResultsDir || !currentRunId) {
    return null;
  }
  return path.join(currentResultsDir, currentRunId, "unit-results", compatibilityUnitResultName);
}

function sameResolvedPath(left, right) {
  if (!left || !right) {
    return false;
  }
  return path.resolve(inputPath(left)) === path.resolve(inputPath(right));
}

function parseUnitResult({ resultPath, findings }) {
  if (!resultPath) {
    findings.push({
      check_id: "compatibility-unit-result-path",
      severity: "blocking",
      message: "compatibility unit result path is required",
    });
    return null;
  }
  if (!existsSync(inputPath(resultPath))) {
    findings.push({
      check_id: "compatibility-unit-result-present",
      severity: "blocking",
      path: displayPath(resultPath),
      message: "current canonical seaweedfs-compatibility unit result is missing",
    });
    return null;
  }
  try {
    return JSON.parse(readFileSync(inputPath(resultPath), "utf8"));
  } catch (error) {
    findings.push({
      check_id: "compatibility-unit-result-parse",
      severity: "blocking",
      path: displayPath(resultPath),
      message: error instanceof Error ? error.message : String(error),
    });
    return null;
  }
}

function currentRunPrerequisitesSkipped() {
  return (
    process.env.CARTULARY_HARNESS_SKIP_PREREQUISITES === "1" &&
    process.env.CARTULARY_HARNESS_GRAPH_CHILD !== "1"
  );
}

function validateCurrentCompatibilitySource({
  findings,
  reportPath,
  unitResult = null,
  unitResultPath = null,
  currentResultsDir = process.env.CARTULARY_TEST_RESULTS_DIR ?? null,
  currentRunId = process.env.CARTULARY_TEST_RUN_ID ?? null,
  prerequisitesSkipped = currentRunPrerequisitesSkipped(),
}) {
  const displayReportPath = reportPath ? displayPath(reportPath) : null;
  if (prerequisitesSkipped) {
    findings.push({
      check_id: "compatibility-current-run-prerequisites",
      severity: "blocking",
      path: displayReportPath,
      message: "strict compatibility evidence requires running the current seaweedfs-compatibility prerequisite",
    });
  }
  const expectedReportPath = expectedCurrentCompatibilityReportPath({
    currentResultsDir,
    currentRunId,
  });
  if (!expectedReportPath) {
    findings.push({
      check_id: "compatibility-current-run-context",
      severity: "blocking",
      path: displayReportPath,
      message: "current compatibility evidence requires CARTULARY_TEST_RESULTS_DIR and CARTULARY_TEST_RUN_ID",
    });
  } else if (!sameResolvedPath(reportPath, expectedReportPath)) {
    findings.push({
      check_id: "compatibility-current-target-source",
      severity: "blocking",
      path: displayReportPath,
      message: "compatibility evidence must come from the current seaweedfs-compatibility target run",
      expected: displayPath(expectedReportPath),
    });
  }
  if (displayReportPath?.includes("/services-up/")) {
    findings.push({
      check_id: "compatibility-services-up-source",
      severity: "blocking",
      path: displayReportPath,
      message: "compatibility evidence must come from seaweedfs-compatibility, not retained services-up evidence",
    });
  }
  if (
    displayReportPath === path.join(defaultReleaseArtifactDir, releaseArtifactSubdir, seaweedfsCompatibilityReportName)
  ) {
    findings.push({
      check_id: "compatibility-stable-report-source",
      severity: "blocking",
      path: displayReportPath,
      message: "stable copied release-artifact compatibility reports are not strict release-gate evidence",
    });
  }

  const resultPath =
    unitResultPath ??
    expectedCurrentUnitResultPath({ currentResultsDir, currentRunId });
  const result =
    unitResult ??
    parseUnitResult({
      resultPath,
      findings,
    });
  if (!result) {
    return;
  }
  if (result.unit_id !== `target:${seaweedfsCompatibilityTarget}`) {
    findings.push({
      check_id: "compatibility-unit-result-identity",
      severity: "blocking",
      path: resultPath ? displayPath(resultPath) : null,
      message: "current canonical compatibility unit result has the wrong unit",
      actual: result.unit_id ?? null,
      expected: `target:${seaweedfsCompatibilityTarget}`,
    });
  }
  if (result.status !== "passed") {
    findings.push({
      check_id: "compatibility-unit-result-status",
      severity: "blocking",
      path: resultPath ? displayPath(resultPath) : null,
      message: "current canonical seaweedfs-compatibility unit result is not passed",
      actual: result.status ?? null,
    });
  }
}

function buildSeaweedFSCompatibilityEvidence({
  repoCommitValue = "test-commit",
  generatedAt = new Date().toISOString(),
  reportPath = null,
  report = null,
  requireCurrentRun = false,
  currentResultsDir = process.env.CARTULARY_TEST_RESULTS_DIR ?? null,
  currentRunId = process.env.CARTULARY_TEST_RUN_ID ?? null,
  unitResult = null,
  unitResultPath = null,
  prerequisitesSkipped = currentRunPrerequisitesSkipped(),
} = {}) {
  const findings = [];
  const loaded =
    report ??
    readArtifactJSON({
      artifactPath: reportPath,
      schemaID: "cartulary.seaweedfs_s3_compatibility_report.v1",
      findings,
    });
  const cases = Array.isArray(loaded?.cases) ? loaded.cases : [];
  const caseIDCounts = new Map();
  for (const item of cases) {
    caseIDCounts.set(item.case_id, (caseIDCounts.get(item.case_id) ?? 0) + 1);
  }
  const missingCaseIDs = seaweedfsCompatibilityCaseIds.filter((caseID) => !caseIDCounts.has(caseID));
  const unexpectedCaseIDs = [...caseIDCounts.keys()].filter(
    (caseID) => !seaweedfsCompatibilityCaseIds.includes(caseID),
  );
  const duplicateCaseIDs = [...caseIDCounts.entries()]
    .filter(([, count]) => count > 1)
    .map(([caseID]) => caseID);
  const failingCaseIDs = cases
    .filter((item) => seaweedfsCompatibilityCaseIds.includes(item.case_id) && item.status !== "pass")
    .map((item) => item.case_id);
  const forbiddenSkipRows = Array.isArray(loaded?.forbidden_skip_rows)
    ? loaded.forbidden_skip_rows
    : [];
  const displayReportPath = reportPath ? displayPath(reportPath) : null;
  if (loaded && loaded.result !== "pass") {
    findings.push({
      check_id: "compatibility-result",
      severity: "blocking",
      path: displayReportPath,
      message: "compatibility report result is not pass",
      actual: loaded.result ?? null,
    });
  }
  if (missingCaseIDs.length > 0) {
    findings.push({
      check_id: "compatibility-case-completeness",
      severity: "blocking",
      path: displayReportPath,
      message: "compatibility report is missing required cases",
      case_ids: missingCaseIDs,
    });
  }
  if (unexpectedCaseIDs.length > 0 || duplicateCaseIDs.length > 0) {
    findings.push({
      check_id: "compatibility-case-identity",
      severity: "blocking",
      path: displayReportPath,
      message: "compatibility report has unexpected or duplicate cases",
      unexpected_case_ids: unexpectedCaseIDs,
      duplicate_case_ids: duplicateCaseIDs,
    });
  }
  if (failingCaseIDs.length > 0) {
    findings.push({
      check_id: "compatibility-case-status",
      severity: "blocking",
      path: displayReportPath,
      message: "compatibility report contains non-passing cases",
      case_ids: failingCaseIDs,
    });
  }
  if (forbiddenSkipRows.length > 0) {
    findings.push({
      check_id: "compatibility-forbidden-skip",
      severity: "blocking",
      path: displayReportPath,
      message: "compatibility report contains forbidden skip rows",
      rows: forbiddenSkipRows,
    });
  }
  if (requireCurrentRun) {
    validateCurrentCompatibilitySource({
      findings,
      reportPath,
      unitResult,
      unitResultPath,
      currentResultsDir,
      currentRunId,
      prerequisitesSkipped,
    });
  }
  return {
    schema_id: "cartulary.seaweedfs_compatibility_evidence.v1",
    generated_at: generatedAt,
    repo_commit: repoCommitValue,
    compatibility_report_path: displayReportPath,
    expected_case_ids: [...seaweedfsCompatibilityCaseIds],
    observed_case_statuses: cases.map((item) => ({
      case_id: item.case_id,
      status: item.status ?? null,
      reason_code: item.reason_code ?? null,
    })),
    forbidden_skip_rows: forbiddenSkipRows,
    findings,
    result: findings.length === 0 ? "pass" : "fail",
  };
}

function buildBackupIntegrityCoverage({
  repoCommitValue = "test-commit",
  generatedAt = new Date().toISOString(),
} = {}) {
  const findings = [];
  const registry = readJSON(recoveryContractRegistryPath);
  const currentSchemaIDs = new Set(registry.current_schema_ids ?? []);
  const requiredSchemaIDs = [
    "cartulary.backup_artifact_envelope.v2",
    "cartulary.backup_integrity_manifest.v3",
    "cartulary.object_store_backup_manifest.v2",
    "cartulary.object_store_backup_summary.v2",
  ];
  const missingSchemaIDs = requiredSchemaIDs.filter((schemaID) => !currentSchemaIDs.has(schemaID));
  if (missingSchemaIDs.length > 0) {
    findings.push({
      check_id: "recovery-backup-integrity-contracts",
      severity: "blocking",
      path: recoveryContractRegistryPath,
      message: "Recovery contract registry is missing current backup-integrity schemas",
      missing_schema_ids: missingSchemaIDs,
    });
  }
  const threatPolicy = readJSON(objectStoreThreatPolicyPath);
  const activeControls = new Set(
    (threatPolicy.storage_ref_controls ?? [])
      .filter((control) => control.status === "active")
      .map((control) => control.check_id),
  );
  if (!activeControls.has("backup-preserves-storage-ref-ownership")) {
    findings.push({
      check_id: "backup-storage-ref-control",
      severity: "blocking",
      path: objectStoreThreatPolicyPath,
      message: "Object-store threat policy lacks the active backup storage-reference ownership control",
    });
  }

  return {
    schema_id: "cartulary.seaweedfs_backup_integrity_coverage.v1",
    generated_at: generatedAt,
    repo_commit: repoCommitValue,
    recovery_contract_registry_path: recoveryContractRegistryPath,
    threat_policy_path: objectStoreThreatPolicyPath,
    required_schema_ids: requiredSchemaIDs,
    active_storage_ref_control: "backup-preserves-storage-ref-ownership",
    findings,
    result: findings.length === 0 ? "pass" : "fail",
  };
}

function classifyArtifactBoundary(rel) {
  if (rel.includes("object-store-backup-manifest.json")) {
    return "operator_private";
  }
  return "public_or_shareable";
}

function redactionScanPaths({
  selectedArtifactPaths = [],
  compatibilityReportPath = null,
}) {
  const paths = [
    ...selectedArtifactPaths,
    ...(compatibilityReportPath ? [compatibilityReportPath] : []),
  ].map(displayPath);
  return paths.filter((entry, index, all) => all.indexOf(entry) === index);
}

function findPublicRawStorageFieldFindings({ rel, value, trail = [] }) {
  const findings = [];
  if (Array.isArray(value)) {
    value.forEach((child, index) => {
      findings.push(...findPublicRawStorageFieldFindings({ rel, value: child, trail: [...trail, String(index)] }));
    });
    return findings;
  }
  if (!value || typeof value !== "object") {
    return findings;
  }
  for (const [key, child] of Object.entries(value)) {
    const normalizedKey = key.toLowerCase();
    const jsonPath = [...trail, key].join(".") || "$";
    if (
      publicRawStorageFieldNames.has(normalizedKey) &&
      typeof child === "string" &&
      child.trim() !== ""
    ) {
      findings.push({
        check_id: "public-raw-storage-field",
        path: rel,
        severity: "blocking",
        message: "public/shareable artifact contains a raw object-store field",
        json_path: jsonPath,
      });
    }
    findings.push(...findPublicRawStorageFieldFindings({ rel, value: child, trail: [...trail, key] }));
  }
  return findings;
}

function buildRedactionLeakageScan({
  generatedAt,
  repoCommitValue,
  selectedArtifactPaths,
  compatibilityReportPath = null,
}) {
  const paths = redactionScanPaths({
    selectedArtifactPaths,
    compatibilityReportPath,
  });
  const artifacts = [];
  const findings = [];
  for (const rel of paths) {
    const abs = inputPath(rel);
    if (!existsSync(abs)) {
      artifacts.push({ path: rel, boundary: "missing", result: "missing" });
      findings.push({
        check_id: "current-artifact-present",
        path: rel,
        severity: "blocking",
        message: "required current evidence artifact path is missing",
      });
      continue;
    }
    const boundary = classifyArtifactBoundary(rel);
    const text = readFileSync(abs, "utf8");
    const artifactFindings = [];
    for (const needle of secretNeedles) {
      if (text.includes(needle)) {
        artifactFindings.push({
          check_id: "raw-secret-value",
          path: rel,
          severity: "blocking",
          message: "artifact contains a raw secret marker",
          marker: needle,
        });
      }
    }
    if (boundary !== "operator_private") {
      for (const needle of publicLeakNeedles) {
        if (text.includes(needle)) {
          artifactFindings.push({
            check_id: "public-storage-identifier",
            path: rel,
            severity: "blocking",
            message: "public/shareable artifact contains a raw object-store identifier",
            marker: needle,
          });
        }
      }
    }
    if (boundary !== "operator_private") {
      try {
        artifactFindings.push(
          ...findPublicRawStorageFieldFindings({
            rel,
            value: JSON.parse(text),
          }),
        );
      } catch {
        // Non-JSON shareable artifacts still receive raw text scanning above.
      }
    }
    artifacts.push({
      path: rel,
      boundary,
      result: artifactFindings.length === 0 ? "pass" : "fail",
    });
    findings.push(...artifactFindings);
  }
  return {
    schema_id: "cartulary.seaweedfs_release_redaction_scan.v1",
    generated_at: generatedAt,
    repo_commit: repoCommitValue,
    scanned_artifacts: artifacts,
    findings,
    result: findings.length === 0 ? "pass" : "fail",
  };
}

function artifactRef(pathMap, name) {
  return pathMap[name] ?? null;
}

function buildReleaseGateSummary({
  generatedAt,
  repoCommitValue,
  artifacts,
  pathMap,
}) {
  const claims = [];
  function add(row, status, predicate, evidence) {
    claims.push({ row, status, predicate, evidence_paths: evidence.filter(Boolean) });
  }
  add("SWFS-AC-001", artifacts.occurrence.result === "pass" && artifacts.exposure.result === "pass" ? "claimable" : "unclaimed", "occurrence inventory passes and release/default manifests contain zero retired server defaults", [
    artifactRef(pathMap, "occurrence-inventory.json"),
    artifactRef(pathMap, "release-manifest-exposure.json"),
  ]);
  add("SWFS-AC-002", artifacts.dependency.result === "pass" ? "claimable" : "unclaimed", "SDK dependency imports are only in allowed adapter, probe, and test-fixture boundaries", [
    artifactRef(pathMap, "dependency-boundary.json"),
  ]);
  add("SWFS-AC-006", artifacts.occurrence.result === "pass" && artifacts.exposure.result === "pass" ? "claimable" : "unclaimed", "occurrence inventory and release exposure scan prove forbidden SeaweedFS admin surfaces are absent", [
    artifactRef(pathMap, "occurrence-inventory.json"),
    artifactRef(pathMap, "release-manifest-exposure.json"),
  ]);
  add("SWFS-AC-021", artifacts.threat.result === "pass" ? "claimable" : "unclaimed", "retained STRIDE coverage covers every Section 15 row", [
    artifactRef(pathMap, "threat-model-coverage.json"),
  ]);
  add("SWFS-AC-022", artifacts.dependency.result === "pass" && artifacts.sbomLicense.result === "pass" ? "claimable" : "unclaimed", "SBOM/license evidence contains no retired server artifact and classifies the SDK dependency as client-only", [
    artifactRef(pathMap, "dependency-boundary.json"),
    artifactRef(pathMap, "sbom-license-classification.json"),
  ]);
  add("SWFS-AC-023", artifacts.occurrence.result === "pass" ? "claimable" : "unclaimed", "occurrence inventory has zero invalid default-document references", [
    artifactRef(pathMap, "occurrence-inventory.json"),
  ]);
  add("SWFS-AC-025", artifacts.occurrence.result === "pass" ? "claimable" : "unclaimed", "exact occurrence inventory artifact has result='pass'", [
    artifactRef(pathMap, "occurrence-inventory.json"),
  ]);
  add("SWFS-AC-015", artifacts.compatibility.result === "pass" ? "claimable" : "unclaimed", "full SeaweedFS compatibility report passes every SWFS-COMP row with no forbidden skip rows", [
    artifactRef(pathMap, "seaweedfs-compatibility-evidence.json"),
    artifacts.compatibility.compatibility_report_path,
  ]);
  add("SWFS-AC-018", artifacts.storageRefOwner.result === "pass" && artifacts.backupIntegrity.result === "pass" && artifacts.compatibility.result === "pass" ? "claimable" : "blocked", "Core 01 owns logical storage refs and physical keys, current backup contracts preserve that ownership, and the current SeaweedFS backend is compatible", [
    artifactRef(pathMap, "storage-ref-owner-coverage.json"),
    artifactRef(pathMap, "backup-integrity-coverage.json"),
    artifactRef(pathMap, "seaweedfs-compatibility-evidence.json"),
  ]);
  const blockingRows = claims
    .filter((claim) => claim.row !== "SWFS-AC-024" && claim.status !== "claimable")
    .map((claim) => claim.row);
  const aggregateChecks = [
    {
      check_id: "redaction-leakage-scan",
      result: artifacts.redaction.result,
      predicate: "redaction leakage scan passes for every release-gate artifact boundary",
      evidence_paths: [artifactRef(pathMap, "redaction-leakage-scan.json")].filter(Boolean),
    },
  ];
  const blockingChecks = aggregateChecks.filter((check) => check.result !== "pass");
  const aggregateChecksPass = blockingChecks.length === 0;
  add("SWFS-AC-024", blockingRows.length === 0 && aggregateChecksPass ? "claimable" : "blocked", "full release gate passes only when no release-blocking row or aggregate check remains unresolved", [
    artifactRef(pathMap, "release-gate-summary.json"),
    artifactRef(pathMap, "redaction-leakage-scan.json"),
  ]);
  const releaseRowsPass = releaseGateRows
    .filter((row) => row !== "SWFS-AC-024")
    .every((row) => claims.find((claim) => claim.row === row)?.status === "claimable");
  return {
    schema_id: "cartulary.seaweedfs_release_gate_summary.v2",
    generated_at: generatedAt,
    repo_commit: repoCommitValue,
    evidence_result: releaseRowsPass && aggregateChecksPass ? "pass" : "fail",
    release_gate_result: blockingRows.length === 0 && releaseRowsPass && aggregateChecksPass ? "pass" : "blocked",
    blocking_rows: blockingRows,
    blocking_checks: blockingChecks,
    claims,
    artifacts: pathMap,
    result: releaseRowsPass && aggregateChecksPass ? "pass" : "fail",
  };
}

function parseArgs(argv) {
  const args = {
    enforceReleaseGate: false,
    classificationPath: process.env.SEAWEEDFS_OCCURRENCE_CLASSIFICATIONS ?? defaultClassificationPath,
    outputDir:
      process.env.SEAWEEDFS_RELEASE_ARTIFACT_DIR ??
      path.join(process.env.RELEASE_ARTIFACT_DIR ?? defaultReleaseArtifactDir, releaseArtifactSubdir),
    sbomPath: process.env.SBOM_ARTIFACT ?? path.join(defaultReleaseArtifactDir, "sbom.cyclonedx.json"),
    licensePath: process.env.LICENSE_REPORT_ARTIFACT ?? path.join(defaultReleaseArtifactDir, "license-report.json"),
    compatibilityReportPath: defaultCompatibilityReportPath(),
    runId: process.env.CARTULARY_SEAWEEDFS_RELEASE_RUN_ID ?? null,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--enforce-release-gate") {
      args.enforceReleaseGate = true;
    } else if (arg === "--classification") {
      index += 1;
      args.classificationPath = argv[index];
    } else if (arg === "--output-dir") {
      index += 1;
      args.outputDir = argv[index];
    } else if (arg === "--sbom") {
      index += 1;
      args.sbomPath = argv[index];
    } else if (arg === "--license-report") {
      index += 1;
      args.licensePath = argv[index];
    } else if (arg === "--compatibility-report") {
      index += 1;
      args.compatibilityReportPath = argv[index];
    } else if (arg === "--run-id") {
      index += 1;
      args.runId = argv[index];
    } else {
      throw new Error(`unknown argument ${arg}`);
    }
  }
  return args;
}

function defaultRunId(now = new Date()) {
  const stamp = now.toISOString().replace(/[-:]/g, "").replace(/\.\d{3}Z$/, "Z");
  return `${stamp}-p${process.pid}`;
}

function materializeArtifacts({ outputDir, runId, artifacts }) {
  const outputAbs = path.isAbsolute(outputDir) ? outputDir : repoPath(outputDir);
  const runDir = path.join(outputAbs, runId);
  mkdirSync(runDir, { recursive: true });
  const pathMap = {};
  for (const [filename, artifact] of Object.entries(artifacts)) {
    const runPath = path.join(runDir, filename);
    const stablePath = path.join(outputAbs, filename);
    writeJSON(runPath, artifact);
    copyFileSync(runPath, stablePath);
    pathMap[filename] = relPath(runPath);
  }
  return { runDir: relPath(runDir), pathMap };
}

export function generateReleaseEvidence({
  classificationPath = defaultClassificationPath,
  outputDir = path.join(defaultReleaseArtifactDir, releaseArtifactSubdir),
  sbomPath = path.join(defaultReleaseArtifactDir, "sbom.cyclonedx.json"),
  licensePath = path.join(defaultReleaseArtifactDir, "license-report.json"),
  compatibilityReportPath = defaultCompatibilityReportPath(),
  enforceReleaseGate = false,
  runId = defaultRunId(),
  now = new Date(),
} = {}) {
  const generatedAt = now.toISOString();
  const commit = repoCommit();
  const files = trackedFiles();
  const policy = loadGeneratedPolicy();
  const rules = loadClassificationRules(classificationPath);
  const occurrence = buildOccurrenceInventory({
    files,
    policy,
    rules,
    repoCommitValue: commit,
    scannedAt: generatedAt,
  });
  const exposure = buildReleaseManifestExposure({
    files,
    repoCommitValue: commit,
    generatedAt,
  });
  const sbom = existsSync(inputPath(sbomPath)) ? JSON.parse(readFileSync(inputPath(sbomPath), "utf8")) : null;
  const licenseReport = existsSync(inputPath(licensePath))
    ? JSON.parse(readFileSync(inputPath(licensePath), "utf8"))
    : null;
  const dependency = buildDependencyBoundary({
    files,
    repoCommitValue: commit,
    generatedAt,
    sbom,
    licenseReport,
  });
  const sbomLicense = buildSbomLicenseClassification({
    repoCommitValue: commit,
    generatedAt,
    sbomPath,
    licensePath,
  });
  const threat = buildThreatModelCoverage({
    repoCommitValue: commit,
    generatedAt,
  });
  const storageRefOwner = buildStorageRefOwnerCoverage({
    repoCommitValue: commit,
    generatedAt,
  });
  const compatibility = buildSeaweedFSCompatibilityEvidence({
    repoCommitValue: commit,
    generatedAt,
    reportPath: compatibilityReportPath,
    requireCurrentRun: true,
  });
  const backupIntegrity = buildBackupIntegrityCoverage({
    repoCommitValue: commit,
    generatedAt,
  });
  const firstArtifacts = {
    "occurrence-inventory.json": occurrence,
    "release-manifest-exposure.json": exposure,
    "dependency-boundary.json": dependency,
    "sbom-license-classification.json": sbomLicense,
    "threat-model-coverage.json": threat,
    "storage-ref-owner-coverage.json": storageRefOwner,
    "seaweedfs-compatibility-evidence.json": compatibility,
    "backup-integrity-coverage.json": backupIntegrity,
  };
  const materialized = materializeArtifacts({
    outputDir,
    runId,
    artifacts: firstArtifacts,
  });
  const redaction = buildRedactionLeakageScan({
    generatedAt,
    repoCommitValue: commit,
    selectedArtifactPaths: Object.values(materialized.pathMap),
    compatibilityReportPath,
  });
  const secondMaterialized = materializeArtifacts({
    outputDir,
    runId,
    artifacts: {
      "redaction-leakage-scan.json": redaction,
    },
  });
  const pathMap = { ...materialized.pathMap, ...secondMaterialized.pathMap };
  const summary = buildReleaseGateSummary({
    generatedAt,
    repoCommitValue: commit,
    artifacts: {
      occurrence,
      exposure,
      dependency,
      sbomLicense,
      threat,
      storageRefOwner,
      compatibility,
      backupIntegrity,
      redaction,
    },
    pathMap,
  });
  const finalMaterialized = materializeArtifacts({
    outputDir,
    runId,
    artifacts: {
      "release-gate-summary.json": summary,
    },
  });
  pathMap["release-gate-summary.json"] = finalMaterialized.pathMap["release-gate-summary.json"];
  summary.artifacts = pathMap;
  for (const claim of summary.claims) {
    if (
      claim.row === "SWFS-AC-024" &&
      pathMap["release-gate-summary.json"] &&
      !claim.evidence_paths.includes(pathMap["release-gate-summary.json"])
    ) {
      claim.evidence_paths = [pathMap["release-gate-summary.json"], ...claim.evidence_paths];
    }
  }
  writeJSON(repoPath(pathMap["release-gate-summary.json"]), summary);
  copyFileSync(repoPath(pathMap["release-gate-summary.json"]), path.join(path.isAbsolute(outputDir) ? outputDir : repoPath(outputDir), "release-gate-summary.json"));
  return {
    run_dir: materialized.runDir,
    artifacts: {
      occurrence,
      exposure,
      dependency,
      sbomLicense,
      threat,
      storageRefOwner,
      compatibility,
      backupIntegrity,
      redaction,
      summary,
    },
    path_map: pathMap,
  };
}

function main() {
  try {
    const args = parseArgs(process.argv.slice(2));
    const result = generateReleaseEvidence({
      classificationPath: args.classificationPath,
      outputDir: args.outputDir,
      sbomPath: args.sbomPath,
      licensePath: args.licensePath,
      compatibilityReportPath: args.compatibilityReportPath,
      enforceReleaseGate: args.enforceReleaseGate,
      runId: args.runId ?? defaultRunId(),
    });
    const summary = result.artifacts.summary;
    console.log(
      `seaweedfs release evidence: evidence=${summary.evidence_result} release_gate=${summary.release_gate_result} run_dir=${result.run_dir}`,
    );
    if (summary.evidence_result !== "pass") {
      process.exitCode = 1;
      return;
    }
    if (args.enforceReleaseGate && summary.release_gate_result !== "pass") {
      process.exitCode = 1;
    }
  } catch (error) {
    console.error(error instanceof Error ? error.message : String(error));
    process.exitCode = 1;
  }
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main();
}

export {
  buildDependencyBoundary,
  buildOccurrenceInventory,
  buildReleaseGateSummary,
  buildReleaseManifestExposure,
  buildBackupIntegrityCoverage,
  buildSeaweedFSCompatibilityEvidence,
  buildSbomLicenseClassification,
  buildStorageRefOwnerCoverage,
  buildThreatModelCoverage,
  buildRedactionLeakageScan,
  classifyOccurrence,
  findOccurrencesInText,
  globToRegExp,
  isDefaultExcluded,
  isTextBuffer,
  loadClassificationRules,
  occurrenceResult,
  occurrenceTokens,
  sortOccurrences,
};
