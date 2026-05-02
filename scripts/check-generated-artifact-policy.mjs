#!/usr/bin/env node
import { execFileSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const schemaID = "cartulary.generated_artifact_policy.v1";
const defaultPolicyPath = path.join(repoRoot, "tools", "generated_artifact_policy.json");

function usage() {
  throw new Error("usage: check-generated-artifact-policy.mjs [--root <path>] [--policy <path>]");
}

function parseArgs(argv) {
  const options = {
    root: repoRoot,
    policy: defaultPolicyPath,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--root") {
      options.root = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--policy") {
      options.policy = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.root || !options.policy) {
    usage();
  }
  options.root = path.resolve(options.root);
  options.policy = path.isAbsolute(options.policy) ? options.policy : path.join(options.root, options.policy);
  return options;
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function repoPath(value, label, { rootPath = false } = {}) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty repo path`);
  }
  if (value.includes("\0") || value.includes("\\") || value.includes("*") || path.posix.isAbsolute(value)) {
    throw new Error(`${label} must be an explicit repo-local path without globs`);
  }
  const normalized = path.posix.normalize(value);
  if (normalized === "." || normalized.startsWith("../") || normalized.includes("/../")) {
    throw new Error(`${label} must stay under the repository root`);
  }
  if (rootPath && normalized.split("/").length < 2) {
    throw new Error(`${label} is too broad for a generated root`);
  }
  return normalized;
}

function stringArray(value, label) {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  return value.map((entry, index) => {
    if (typeof entry !== "string" || entry.trim() === "") {
      throw new Error(`${label}[${index + 1}] must be a non-empty string`);
    }
    return entry;
  });
}

function extensionSet(value, label) {
  const extensions = stringArray(value, label);
  for (const extension of extensions) {
    if (!/^\.[A-Za-z0-9]+$/.test(extension)) {
      throw new Error(`${label} contains invalid extension ${extension}`);
    }
  }
  return new Set(extensions);
}

function markerRegex(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty regular expression`);
  }
  return new RegExp(value, "m");
}

function validatePolicy(policy) {
  if (policy.schema_id !== schemaID) {
    throw new Error(`policy must declare schema_id ${schemaID}`);
  }
  const ignoredSentinels = new Set(stringArray(policy.ignored_sentinel_filenames ?? [], "ignored_sentinel_filenames"));
  const roots = (policy.generated_roots ?? []).map((entry, index) => ({
    path: repoPath(entry?.path, `generated_roots[${index + 1}].path`, { rootPath: true }),
    allowedExtensions: extensionSet(entry?.allowed_extensions, `generated_roots[${index + 1}].allowed_extensions`),
    marker: markerRegex(entry?.required_marker, `generated_roots[${index + 1}].required_marker`),
  }));
  const files = (policy.generated_files ?? []).map((entry, index) => ({
    path: repoPath(entry?.path, `generated_files[${index + 1}].path`),
    allowedExtensions: extensionSet(entry?.allowed_extensions, `generated_files[${index + 1}].allowed_extensions`),
    marker: markerRegex(entry?.required_marker, `generated_files[${index + 1}].required_marker`),
  }));
  const seenRoots = new Set();
  for (const root of roots) {
    if (seenRoots.has(root.path)) {
      throw new Error(`duplicate generated root ${root.path}`);
    }
    seenRoots.add(root.path);
  }
  for (let left = 0; left < roots.length; left += 1) {
    for (let right = left + 1; right < roots.length; right += 1) {
      const a = roots[left].path;
      const b = roots[right].path;
      if (a.startsWith(`${b}/`) || b.startsWith(`${a}/`)) {
        throw new Error(`generated roots must not overlap: ${a} and ${b}`);
      }
    }
  }
  const seenFiles = new Set();
  for (const file of files) {
    if (seenFiles.has(file.path)) {
      throw new Error(`duplicate generated file ${file.path}`);
    }
    seenFiles.add(file.path);
    if ([...seenRoots].some((root) => file.path.startsWith(`${root}/`))) {
      throw new Error(`generated file ${file.path} is already covered by a generated root`);
    }
  }
  return {
    ignoredSentinels,
    roots,
    files,
    lintScopeChecks: policy.lint_scope_checks ?? {},
  };
}

function gitFiles(root, relPath, mode) {
  const args =
    mode === "tracked"
      ? ["-C", root, "ls-files", "-z", "--", relPath]
      : ["-C", root, "ls-files", "-z", "--others", "--exclude-standard", "--", relPath];
  const output = execFileSync("git", args, { encoding: "utf8" });
  return output.split("\0").filter(Boolean);
}

function generatedFilesForRoot(root, relPath) {
  return Array.from(new Set([...gitFiles(root, relPath, "tracked"), ...gitFiles(root, relPath, "untracked")])).sort();
}

function firstHeader(file) {
  return readFileSync(file, "utf8").split(/\r?\n/).slice(0, 20).join("\n");
}

function checkGeneratedFile({ root, relPath, allowedExtensions, marker, ignoredSentinels, failures }) {
  const base = path.posix.basename(relPath);
  if (ignoredSentinels.has(base)) {
    return false;
  }
  const fullPath = path.join(root, relPath);
  if (!existsSync(fullPath)) {
    failures.push(`${relPath}: tracked generated artifact is missing from the working tree`);
    return false;
  }
  const extension = path.posix.extname(relPath);
  if (!allowedExtensions.has(extension)) {
    failures.push(`${relPath}: unsupported generated extension ${extension || "(none)"}`);
    return false;
  }
  if (!marker.test(firstHeader(fullPath))) {
    failures.push(`${relPath}: missing required generated marker`);
    return false;
  }
  return true;
}

function requireStringInFile(root, relPath, needle, failures) {
  const fullPath = path.join(root, relPath);
  if (!existsSync(fullPath)) {
    failures.push(`${relPath}: lint scope source file is missing`);
    return "";
  }
  const content = readFileSync(fullPath, "utf8");
  if (!content.includes(needle)) {
    failures.push(`${relPath}: missing lint scope guard ${JSON.stringify(needle)}`);
  }
  return content;
}

function checkShellScopeSources(root, checks, failures) {
  for (const [index, check] of (checks ?? []).entries()) {
    const relPath = repoPath(check?.path, `lint_scope_checks.shell_sources[${index + 1}].path`);
    const fullPath = path.join(root, relPath);
    if (!existsSync(fullPath)) {
      failures.push(`${relPath}: lint scope source file is missing`);
      continue;
    }
    const content = readFileSync(fullPath, "utf8");
    for (const needle of stringArray(check.must_contain ?? [], `${relPath}.must_contain`)) {
      if (!content.includes(needle)) {
        failures.push(`${relPath}: missing lint scope guard ${JSON.stringify(needle)}`);
      }
    }
    for (const forbidden of stringArray(check.must_not_contain ?? [], `${relPath}.must_not_contain`)) {
      if (content.includes(forbidden)) {
        failures.push(`${relPath}: forbidden broad generated exclusion ${JSON.stringify(forbidden)}`);
      }
    }
  }
}

function checkBiomeScope(root, check, failures) {
  if (!check) {
    return;
  }
  const relPath = repoPath(check.path, "lint_scope_checks.biome.path");
  const config = readJSON(path.join(root, relPath));
  const includes = config.files?.includes;
  if (!Array.isArray(includes)) {
    failures.push(`${relPath}: files.includes must be declared`);
    return;
  }
  for (const required of stringArray(check.required_files_includes ?? [], `${relPath}.required_files_includes`)) {
    if (!includes.includes(required)) {
      failures.push(`${relPath}: files.includes must exclude ${required}`);
    }
  }
  for (const forbidden of stringArray(check.forbidden_files_includes ?? [], `${relPath}.forbidden_files_includes`)) {
    if (includes.includes(forbidden)) {
      failures.push(`${relPath}: files.includes uses non-recursive generated exclusion ${forbidden}`);
    }
  }
  const overrideIncludes = (config.overrides ?? []).flatMap((override) =>
    Array.isArray(override.includes) ? override.includes : [],
  );
  for (const required of stringArray(check.required_override_includes ?? [], `${relPath}.required_override_includes`)) {
    if (!overrideIncludes.includes(required)) {
      failures.push(`${relPath}: overrides must exclude ${required}`);
    }
  }
}

function checkFrontendBoundaryScope(root, check, failures) {
  if (!check) {
    return;
  }
  const relPath = repoPath(check.path, "lint_scope_checks.frontend_import_boundaries.path");
  const config = readJSON(path.join(root, relPath));
  const scanExcludes = config.scan_excludes ?? [];
  for (const required of stringArray(check.required_scan_excludes ?? [], `${relPath}.required_scan_excludes`)) {
    if (!scanExcludes.includes(required)) {
      failures.push(`${relPath}: scan_excludes must exclude ${required}`);
    }
  }
  const restrictedImports = JSON.stringify((config.rules ?? []).flatMap((rule) => rule.restricted_imports ?? []));
  for (const required of stringArray(check.required_restricted_paths ?? [], `${relPath}.required_restricted_paths`)) {
    if (!restrictedImports.includes(required)) {
      failures.push(`${relPath}: restricted imports must include ${required}`);
    }
  }
}

function checkLintScopes(root, checks, failures) {
  checkShellScopeSources(root, checks.shell_sources, failures);
  requireStringInFile(root, "scripts/lib/generated-artifacts.sh", "cartulary_is_generated_artifact_path", failures);
  checkBiomeScope(root, checks.biome, failures);
  checkFrontendBoundaryScope(root, checks.frontend_import_boundaries, failures);
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const policy = validatePolicy(readJSON(options.policy));
  const failures = [];
  let checked = 0;

  for (const generatedRoot of policy.roots) {
    for (const relPath of generatedFilesForRoot(options.root, generatedRoot.path)) {
      if (
        checkGeneratedFile({
          root: options.root,
          relPath,
          allowedExtensions: generatedRoot.allowedExtensions,
          marker: generatedRoot.marker,
          ignoredSentinels: policy.ignoredSentinels,
          failures,
        })
      ) {
        checked += 1;
      }
    }
  }

  for (const generatedFile of policy.files) {
    if (
      checkGeneratedFile({
        root: options.root,
        relPath: generatedFile.path,
        allowedExtensions: generatedFile.allowedExtensions,
        marker: generatedFile.marker,
        ignoredSentinels: policy.ignoredSentinels,
        failures,
      })
    ) {
      checked += 1;
    }
  }

  checkLintScopes(options.root, policy.lintScopeChecks, failures);

  if (failures.length > 0) {
    console.error("generated artifact policy check failed:");
    for (const failure of failures) {
      console.error(`  - ${failure}`);
    }
    process.exit(1);
  }

  console.log(`generated artifact policy check passed: ${checked} generated artifacts checked`);
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  console.error(`generated artifact policy check failed: ${message}`);
  process.exit(1);
}
