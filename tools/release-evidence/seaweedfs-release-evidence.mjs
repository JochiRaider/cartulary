#!/usr/bin/env node
import path from "node:path";
import { fileURLToPath } from "node:url";

import { parseStrictJSON, readRetainedArtifactBytes, validateSchemaSync } from "../harness/contract/index.mjs";
import { executableRepositoryPaths, inspectExecutableRepositoryPath, isDocumentationOnlyPath, restrictedExecutableInputRoots } from "../harness/test-catalog/index.mjs";

import { buildSeaweedfsSummary, publishSeaweedfsSummary, readCurrentSeaweedfsCompatibility, readCurrentSeaweedfsRecovery,
  readSeaweedfsInventories, scanSeaweedfsPrivacy, seaweedfsRunContext } from "./seaweedfs-artifacts.mjs";

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
const defaultClassificationPath =
  "tools/seaweedfs_occurrence_classifications.json";
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
const restrictedInputRoots = restrictedExecutableInputRoots(repoRoot);
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
  return parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(readRepositorySource(rel)));
}

export function repositorySourceFiles(root = repoRoot, restrictedRoots = restrictedInputRoots) {
  return executableRepositoryPaths(root, { restrictedRoots, untracked: true })
    .filter(rel => !isDocumentationOnlyPath(rel, restrictedRoots) &&
      !alwaysExcludedPrefixes.some(prefix => rel.startsWith(prefix)))
    .filter(rel => {
      const admitted = inspectExecutableRepositoryPath(root, rel);
      if (!admitted) return false; // A tracked deletion is absent from current source.
      if (!admitted.stat.isFile()) throw new Error("release source must be a regular file");
      return true;
    });
}

export function readRepositorySource(relative, root = repoRoot, restrictedRoots = restrictedInputRoots) {
  if (isDocumentationOnlyPath(relative, restrictedRoots)) throw new Error("release source is outside executable scope");
  return readRetainedArtifactBytes(root, relative, 64 * 1024 * 1024);
}

function loadGeneratedPolicy(policyRel = generatedPolicyPath) {
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
  if (isDocumentationOnlyPath(normalized, restrictedInputRoots)) return true;
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
    const buffer = readRepositorySource(rel);
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
  readFile = (rel) => readRepositorySource(rel).toString("utf8"),
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

function findImportSites(files, readFile = (rel) => readRepositorySource(rel).toString("utf8")) {
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
  readFile = (rel) => readRepositorySource(rel).toString("utf8"),
  goModText = readRepositorySource("go.mod").toString("utf8"),
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
      findings.push({ json_path: trail.join(".") || "$" });
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
  sbom,
  licenseReport,
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
      out.push({ json_path: trail.join(".") || "$" });
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

function buildBackupIntegrityCoverage({
  repoCommitValue = "test-commit",
  generatedAt = new Date().toISOString(),
} = {}) {
  const findings = [];
  const registry = readJSON(recoveryContractRegistryPath);
  const currentSchemaIDs = new Set(registry.current_schema_ids ?? []);
  const requiredSchemaIDs = [
    "cartulary.backup_artifact_envelope.v3",
    "cartulary.backup_integrity_manifest.v4",
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

// Pure report assembly keeps source policy, metadata and current execution
// observations distinct. The caller acquires those inputs once before projection.
export function buildSeaweedfsReleaseReport({ manifest, target }, { source, inventory, compatible, recovery }) {
  const { occurrence, exposure, dependency, policies } = source;
  const generatedAt = new Date().toISOString(), commit = manifest.source_commit;
  const findings = [];
  const outcomes = {};
  const snapshots = [];
  const status = (id, value, evidence_refs = []) => { outcomes[id] = { status: value ? "pass" : "fail", evidence_refs }; };
  const add = (check_id, code, fields = {}) => findings.push({ check_id, code, ...fields });
  if (!inventory) add("inventory_classification", "inventory-admission-failed");
  status("occurrence_inventory", occurrence.result === "pass", [defaultClassificationPath]);
  status("release_surface", exposure.result === "pass");
  status("dependency_boundary", dependency.result === "pass", ["go.mod"]);
  for (const entry of occurrence.occurrences) if (entry.classification === "invalid" || entry.classification === "unclassified" || !entry.owner || !entry.rationale) {
    add("occurrence_inventory", "unadmitted-occurrence", { path: entry.path, line: entry.line, column: entry.column });
  }
  for (const entry of exposure.findings) add("release_surface", entry.check_id, { path: entry.path, line: entry.line, column: entry.column });
  for (const entry of dependency.sdk_dependency.disallowed_imports) add("dependency_boundary", "sdk-import-boundary", { path: entry.path, line: entry.line, column: entry.column });
  if (dependency.release_server_artifacts.length) add("dependency_boundary", "retired-server-artifact");
  if (inventory) {
    const classification = buildSbomLicenseClassification({ repoCommitValue: commit, generatedAt,
      sbomPath: inventory.evidence_refs[0], licensePath: inventory.evidence_refs[1], ...inventory });
    status("inventory_classification", classification.result === "pass", inventory.evidence_refs);
    if (classification.result !== "pass") add("inventory_classification", "retired-server-artifact");
  } else status("inventory_classification", false);
  for (const [id, result, refs] of [
    ["threat_policy", policies.threat, [objectStoreThreatPolicyPath]],
    ["storage_ref_policy", policies.storageRefOwner, [objectStoreThreatPolicyPath]],
    ["backup_contracts", policies.backupIntegrity, [recoveryContractRegistryPath, objectStoreThreatPolicyPath]],
  ]) {
    status(id, result.result === "pass", refs);
    if (result.result !== "pass") add(id, "contract-inventory-incomplete");
  }
  if (compatible) {
    status("compatibility", true, compatible.evidence_refs);
    snapshots.push({ path: compatible.evidence_refs[0], boundary: "public_or_shareable", value: compatible.unit },
      { path: compatible.evidence_refs[1], boundary: "public_or_shareable", value: compatible.report });
  } else {
    status("compatibility", false);
    add("compatibility", "current-compatibility-unqualified");
  }
  if (recovery) {
    status("recovery", true, recovery.evidence_refs);
    snapshots.push({ path: recovery.evidence_refs[0], boundary: "public_or_shareable", value: recovery.unit },
      { path: recovery.evidence_refs[1], boundary: "public_or_shareable", value: recovery.row });
  } else {
    status("recovery", false);
    add("recovery", "current-recovery-unqualified");
  }
  const summaryRef = `${target}/artifacts/release-gate-summary.json`;
  const draft = buildSeaweedfsSummary({ manifest, target, outcomes, findings, occurrences: occurrence.occurrences });
  const privacy = scanSeaweedfsPrivacy([...snapshots, { path: summaryRef, boundary: "public_or_shareable", value: draft }]);
  outcomes.redaction = { status: privacy.status, evidence_refs: privacy.evidence_refs };
  findings.push(...privacy.findings);
  const summary = buildSeaweedfsSummary({ manifest, target, outcomes, findings, occurrences: occurrence.occurrences });
  return summary;
}

export function generateReleaseEvidence(context = seaweedfsRunContext()) {
  const generatedAt = new Date().toISOString(), commit = context.manifest.source_commit;
  const files = repositorySourceFiles(), policy = loadGeneratedPolicy();
  let inventory = null, compatible = null, recovery = null;
  try { inventory = readSeaweedfsInventories(repoRoot); } catch { /* Fixed failure projection below. */ }
  try { compatible = readCurrentSeaweedfsCompatibility(context); } catch { /* Fixed failure projection below. */ }
  try { recovery = readCurrentSeaweedfsRecovery(context); } catch { /* Fixed failure projection below. */ }
  const source = {
    occurrence: buildOccurrenceInventory({ files, policy, rules: loadClassificationRules(), repoCommitValue: commit, scannedAt: generatedAt }),
    exposure: buildReleaseManifestExposure({ files, repoCommitValue: commit, generatedAt }),
    dependency: buildDependencyBoundary({ files, repoCommitValue: commit, generatedAt, sbom: inventory?.sbom, licenseReport: inventory?.licenseReport }),
    policies: {
      threat: buildThreatModelCoverage({ repoCommitValue: commit, generatedAt }),
      storageRefOwner: buildStorageRefOwnerCoverage({ repoCommitValue: commit, generatedAt }),
      backupIntegrity: buildBackupIntegrityCoverage({ repoCommitValue: commit, generatedAt }),
    },
  };
  const summary = buildSeaweedfsReleaseReport(context, { source, inventory, compatible, recovery });
  publishSeaweedfsSummary(context, summary);
  return summary;
}

function main() {
  if (process.argv.length !== 2) throw new Error("SeaweedFS evidence accepts only graph-owned context");
  const summary = generateReleaseEvidence();
  console.log(`seaweedfs release evidence: ${summary.status}`);
  if (summary.status !== "pass") process.exitCode = 1;
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try { main(); }
  catch { console.error("SeaweedFS release evidence unavailable"); process.exitCode = 11; }
}

export {
  buildDependencyBoundary,
  buildOccurrenceInventory,
  buildReleaseManifestExposure,
  buildBackupIntegrityCoverage,
  buildSbomLicenseClassification,
  buildStorageRefOwnerCoverage,
  buildThreatModelCoverage,
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
