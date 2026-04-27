import path from "node:path";

export const schedulerProgressIntervalMs = 10_000;

export function verboseSchedulerOutput() {
  return process.env.VERBOSE === "1" || process.env.CI_VERBOSE === "1";
}

export function normalizePath(value) {
  return value.replaceAll("\\", "/");
}

export function relToRepo(repoRoot, value) {
  if (!value) {
    return "";
  }
  const normalized = normalizePath(value);
  if (!path.isAbsolute(value)) {
    return normalized;
  }
  const relative = normalizePath(path.relative(repoRoot, value));
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return normalized;
}

export function schedulerTargetDir(repoRoot, target) {
  const runID = process.env.CARTULARY_TEST_RUN_ID || "adhoc";
  const configured = process.env.CARTULARY_TEST_RESULTS_DIR;
  const resultsRoot = configured
    ? path.isAbsolute(configured)
      ? configured
      : path.join(repoRoot, configured)
    : path.join(repoRoot, ".cartulary", "test-results");
  return path.join(resultsRoot, runID, target);
}

export function formatResourceMap(values) {
  const entries = Array.from(values.entries()).sort((left, right) => left[0].localeCompare(right[0]));
  if (entries.length === 0) {
    return "{}";
  }
  return `{${entries.map(([key, value]) => `${key}:${value}`).join(",")}}`;
}

export function resourceMapToObject(values) {
  return Object.fromEntries(
    Array.from(values.entries()).sort((left, right) => left[0].localeCompare(right[0])),
  );
}

export function formatResourceList(values) {
  if (values.length === 0) {
    return "none";
  }
  return values.join(",");
}

export function formatLabelList(values, limit = 3) {
  if (values.length === 0) {
    return "none";
  }
  const displayed = values.slice(0, limit);
  const suffix = values.length > limit ? `,+${values.length - limit}` : "";
  return `${displayed.join(",")}${suffix}`;
}

export function formatDurationMs(value) {
  if (!Number.isFinite(value)) {
    return "0.00s";
  }
  return `${(value / 1000).toFixed(2)}s`;
}

export function formatSlowestWork(work) {
  if (work.length === 0) {
    return "none";
  }
  return work.map((entry) => `${entry.label}:${formatDurationMs(entry.duration_ms)}`).join(",");
}

export function resourceLimitSummary(resourceLimits, preferred = []) {
  const seen = new Set();
  const entries = [];
  for (const resource of preferred) {
    if (resourceLimits.has(resource)) {
      entries.push(`${resource}:${resourceLimits.get(resource)}`);
      seen.add(resource);
    }
  }
  for (const [resource, value] of Array.from(resourceLimits.entries()).sort((left, right) =>
    left[0].localeCompare(right[0]),
  )) {
    if (!seen.has(resource)) {
      entries.push(`${resource}:${value}`);
    }
  }
  return entries.join(",");
}

export function countBy(values, field) {
  const counts = new Map();
  for (const value of values) {
    counts.set(value[field], (counts.get(value[field]) ?? 0) + 1);
  }
  return Array.from(counts.entries())
    .sort((left, right) => String(left[0]).localeCompare(String(right[0])))
    .map(([key, count]) => `${key}:${count}`)
    .join(",");
}

export function topWeightedUnits(workUnits, limit = 5) {
  return [...workUnits]
    .sort((left, right) => right.weight - left.weight || left.label.localeCompare(right.label))
    .slice(0, limit)
    .map((unit) => `${unit.label}:${unit.weight}`)
    .join(",");
}
