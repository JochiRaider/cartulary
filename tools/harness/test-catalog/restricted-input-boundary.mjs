import { spawnSync } from "node:child_process";
import { closeSync, constants, fstatSync, lstatSync, openSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";

const policySchemaID = "cartulary.executable_input_policy.v1";
const policyPath = "tools/executable_input_policy.json";
const sourceExtensions = new Set([
  ".cjs",
  ".go",
  ".js",
  ".mjs",
  ".sh",
  ".ts",
  ".tsx",
]);

function normalize(value) {
  return value.replaceAll("\\", "/").replace(/^\.\//u, "");
}

function readPolicy(root) {
  const policy = JSON.parse(readFileSync(path.join(root, policyPath), "utf8"));
  validateSchemaSync(policySchemaID, policy);
  for (const key of [
    "restricted_roots",
    "documentation_only_sources",
    "machine_evidence_roots",
  ]) {
    const sorted = [...policy[key]].sort();
    if (JSON.stringify(policy[key]) !== JSON.stringify(sorted)) {
      throw new Error(`${policyPath}.${key} must be ASCII-sorted`);
    }
  }
  return policy;
}

export function restrictedExecutableInputRoots(root) {
  return [...readPolicy(root).restricted_roots];
}

// Git must prune human material before inspecting untracked directories. A
// filter applied only to Git's output is too late to enforce this boundary.
export function executableInputPathspec(restrictedRoots, includes = ["."]) {
  return [
    ...includes,
    ...restrictedRoots.map((entry) => `:(top,exclude,literal)${entry}`),
    ...["*.md", "*.markdown", "README", "README.*", "**/README", "**/README.*"]
      .map((pattern) => `:(exclude,icase)${pattern}`),
  ];
}

export function executableRepositoryPaths(root, {
  restrictedRoots = restrictedExecutableInputRoots(root),
  untracked = false,
  includes = ["."],
} = {}) {
  const result = spawnSync("git", [
    "ls-files", "--cached", ...(untracked ? ["--others", "--exclude-standard"] : []),
    "-z", "--", ...executableInputPathspec(restrictedRoots, includes),
  ], { cwd: root, encoding: "utf8" });
  if (result.status !== 0) throw new Error("cannot enumerate admitted repository inputs");
  return [...new Set(result.stdout.split("\0").filter(Boolean))].sort();
}

// Git can retain descendants of a directory replaced by a link. Inspect every
// ancestor before probing a descendant; lexical admission alone is insufficient.
// A final link is returned as data for snapshot consumers, never dereferenced.
export function inspectExecutableRepositoryPath(root, relative) {
  const segments = relative.split("/");
  if (path.isAbsolute(relative) || relative.includes("\\") || relative.includes("\0") ||
      segments.some((segment) => !segment || segment === "." || segment === "..")) {
    throw new Error("executable input path is not normalized");
  }
  let absolute = path.resolve(root);
  for (let index = 0; index < segments.length; index++) {
    absolute = path.join(absolute, segments[index]);
    let stat;
    try { stat = lstatSync(absolute); }
    catch (error) { if (error.code === "ENOENT") return null; throw error; }
    if (index === segments.length - 1) return { absolute, stat };
    if (!stat.isDirectory()) throw new Error("executable input ancestor is not a directory (symbolic links are forbidden)");
  }
}

function readExecutableSource(root, relative) {
  const admitted = inspectExecutableRepositoryPath(root, relative);
  if (!admitted) return null;
  if (!admitted.stat.isFile()) throw new Error("executable source must be a regular file (symbolic links are forbidden)");
  const descriptor = openSync(admitted.absolute, constants.O_RDONLY | constants.O_NOFOLLOW | constants.O_NONBLOCK);
  try {
    const opened = fstatSync(descriptor);
    if (!opened.isFile() || opened.dev !== admitted.stat.dev || opened.ino !== admitted.stat.ino)
      throw new Error("executable source changed during admission");
    return readFileSync(descriptor, "utf8");
  } finally { closeSync(descriptor); }
}

function trackedExecutableSources(root, restrictedRoots) {
  const result = spawnSync(
    "git",
    ["ls-files", "--cached", "--others", "--exclude-standard", "-z", "--", ...executableInputPathspec(restrictedRoots)],
    { cwd: root, encoding: "buffer" },
  );
  if (result.status !== 0) {
    throw new Error(`git ls-files failed: ${String(result.stderr)}`);
  }
  return result.stdout
    .toString("utf8")
    .split("\0")
    .filter(Boolean)
    .filter((file) => {
      const normalized = normalize(file);
      const basename = path.posix.basename(normalized);
      return (
        basename === "Makefile" ||
        normalized.endsWith(".mk") ||
        basename === "package.json" ||
        (sourceExtensions.has(path.extname(normalized)) &&
          /^(?:cmd|internal|apps|packages|scripts|tools)\//u.test(normalized))
      );
    })
    .sort();
}

function lineForIndex(source, index) {
  return source.slice(0, index).split("\n").length;
}

function scanRestrictedReadSource(
  consumerPath,
  source,
  restrictedRoots,
) {
  const findings = [];
  for (const root of restrictedRoots) {
    const escaped = root.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&");
    const patterns = [
      new RegExp(`(?:^|[\"'\\x60/])${escaped}/`, "gmu"),
      new RegExp(
        `(?:filepath|path)\\.(?:Join|join|resolve)\\([^;\\n]{0,500}["'\x60]${escaped}["'\x60]`,
        "gmu",
      ),
    ];
    for (const pattern of patterns) {
      for (const match of source.matchAll(pattern)) {
        findings.push({
          consumer_path: normalize(consumerPath),
          restricted_root: root,
          line: lineForIndex(source, match.index),
        });
      }
    }
  }
  return findings;
}

function walkJSONFiles(root, relativeRoot) {
  const result = [];
  const fullRoot = path.join(root, relativeRoot);
  for (const entry of readdirSync(fullRoot, { withFileTypes: true })) {
    const relative = `${relativeRoot}/${entry.name}`;
    if (entry.isDirectory()) {
      result.push(...walkJSONFiles(root, relative));
    } else if (entry.isFile() && entry.name.endsWith(".json")) {
      result.push(relative);
    }
  }
  return result;
}

export function validateExecutableInputPolicy(root) {
  const policy = readPolicy(root);
  const documentationOnly = new Set(policy.documentation_only_sources);
  const findings = [];
  for (const file of trackedExecutableSources(root, policy.restricted_roots)) {
    if (documentationOnly.has(file)) {
      continue;
    }
    const source = readExecutableSource(root, file);
    if (source === null) continue;
    findings.push(
      ...scanRestrictedReadSource(file, source, policy.restricted_roots).map(
        (finding) =>
          `${finding.consumer_path}:${finding.line} references restricted root ${finding.restricted_root}`,
      ),
    );
  }
  for (const evidenceRoot of policy.machine_evidence_roots) {
    for (const file of walkJSONFiles(root, evidenceRoot)) {
      const source = readFileSync(path.join(root, file), "utf8");
      for (const restrictedRoot of policy.restricted_roots) {
        if (
          source.includes(`"${restrictedRoot}"`) ||
          source.includes(`"${restrictedRoot}/`)
        ) {
          findings.push(`${file} references restricted root ${restrictedRoot}`);
        }
      }
    }
  }
  if (findings.length > 0) {
    findings.sort();
    throw new Error(`unauthorized executable inputs: ${findings.join("; ")}`);
  }
  return [];
}
