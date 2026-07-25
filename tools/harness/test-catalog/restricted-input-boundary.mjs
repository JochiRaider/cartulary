import { spawnSync } from "node:child_process";
import { existsSync, readFileSync, readdirSync } from "node:fs";
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

function trackedExecutableSources(root) {
  const result = spawnSync(
    "git",
    ["ls-files", "--cached", "--others", "--exclude-standard", "-z"],
    { cwd: root, encoding: "buffer" },
  );
  if (result.status !== 0) {
    throw new Error(`git ls-files failed: ${String(result.stderr)}`);
  }
  return result.stdout
    .toString("utf8")
    .split("\0")
    .filter(Boolean)
    .filter((file) => existsSync(path.join(root, file)))
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

export function scanRestrictedReadSource(
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
  for (const file of trackedExecutableSources(root)) {
    if (documentationOnly.has(file)) {
      continue;
    }
    const source = readFileSync(path.join(root, file), "utf8");
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
