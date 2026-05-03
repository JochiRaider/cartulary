import { existsSync, readFileSync } from "node:fs";

import { relToRepo, resolveRepoPath } from "./repo-paths.mjs";

export function readJSON(repoRoot, file) {
  return JSON.parse(readFileSync(resolveRepoPath(repoRoot, file), "utf8"));
}

export function sortedObjectByKey(entriesOrObject) {
  const entries =
    entriesOrObject && typeof entriesOrObject[Symbol.iterator] === "function"
      ? entriesOrObject
      : Object.entries(entriesOrObject ?? {});
  return Object.fromEntries([...entries].sort(([left], [right]) => left.localeCompare(right)));
}

export function assertPositiveTargetWeights(baseline, label) {
  if (!baseline.targets || typeof baseline.targets !== "object" || Array.isArray(baseline.targets)) {
    throw new Error(`${label} targets must be an object`);
  }
  for (const [target, weight] of Object.entries(baseline.targets)) {
    if (!Number.isInteger(weight) || weight <= 0) {
      throw new Error(`${label} targets.${target} must be positive integer weight ms`);
    }
  }
}

export function readPositiveTargetBaseline({
  repoRoot,
  file,
  schemaID,
  missingDocument,
  allowMissing = false,
}) {
  const baselineFile = resolveRepoPath(repoRoot, file);
  const label = relToRepo(repoRoot, baselineFile);
  if (!existsSync(baselineFile)) {
    if (allowMissing) {
      return { ...missingDocument, targets: { ...(missingDocument.targets ?? {}) } };
    }
    throw new Error(`${label} is missing`);
  }
  const baseline = readJSON(repoRoot, baselineFile);
  if (baseline.schema_id !== schemaID) {
    throw new Error(`${label} must declare schema_id ${schemaID}`);
  }
  assertPositiveTargetWeights(baseline, label);
  return baseline;
}
