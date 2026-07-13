import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";

const defaultDurationDriftThresholds = {
  underRatio: 2.5,
  underDeltaMs: 15000,
  overRatio: 4,
  overDeltaMs: 30000,
};

const serviceScopeFileName = "service-scope.json";
const contaminationReasonLimit = 8;

export function formatRatio(actual, planned) {
  if (planned <= 0) {
    return "inf";
  }
  return (actual / planned).toFixed(2);
}

export function formatSignedMs(value) {
  return value > 0 ? `+${value}` : String(value);
}

export function durationDriftKind(actual, planned, thresholds = defaultDurationDriftThresholds) {
  if (
    actual > planned * thresholds.underRatio &&
    actual - planned > thresholds.underDeltaMs
  ) {
    return "underplanned";
  }
  if (
    planned > actual * thresholds.overRatio &&
    planned - actual > thresholds.overDeltaMs
  ) {
    return "overplanned";
  }
  return "";
}

export function durationDriftDescription(kind, fields) {
  return [
    kind,
    fields.subject,
    `planned_ms=${fields.plannedMs}`,
    `actual_ms=${fields.actualMs}`,
    `ratio=${formatRatio(fields.actualMs, fields.plannedMs)}`,
    fields.details,
  ]
    .filter(Boolean)
    .join(" ");
}

export function collectServiceTimingContamination(repoRoot, resultsDir) {
  const root = path.resolve(resultsDir);
  if (!existsSync(root) || !statSync(root).isDirectory()) {
    return { contaminated: false, reasons: [] };
  }

  const reasons = new Map();
  for (const file of walkFiles(root)) {
    if (path.basename(file) === serviceScopeFileName) {
      recordServiceScopeContamination(reasons, repoRoot, file);
      continue;
    }
    if (!file.endsWith(".json") || !file.split(path.sep).includes("events")) {
      continue;
    }
    recordServiceEventContamination(reasons, repoRoot, file);
  }

  const values = [...reasons.values()].sort();
  return {
    contaminated: values.length > 0,
    reasons: values,
  };
}

export function formatContaminationReasons(contamination) {
  const reasons = contamination?.reasons ?? [];
  const selected = reasons.slice(0, contaminationReasonLimit);
  const omitted = reasons.length - selected.length;
  return `${selected.join("; ")}${omitted > 0 ? `; omitted=${omitted}` : ""}`;
}

export function printContaminationReasons(stream, contamination) {
  for (const reason of contamination?.reasons ?? []) {
    stream.write(`- ${reason}\n`);
  }
}

function walkFiles(root) {
  const files = [];
  const stack = [root];
  while (stack.length > 0) {
    const current = stack.pop();
    let entries = [];
    try {
      entries = readdirSync(current, { withFileTypes: true });
    } catch {
      continue;
    }
    for (const entry of entries) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (entry.isFile()) {
        files.push(next);
      }
    }
  }
  return files.sort();
}

function readJSONIfPossible(file) {
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch {
    return null;
  }
}

function rel(repoRoot, file) {
  const relative = path.relative(repoRoot, file).replaceAll("\\", "/");
  return relative.startsWith("../") ? file.replaceAll("\\", "/") : relative;
}

function reasonText(value) {
  return String(value ?? "")
    .trim()
    .replace(/\s+/g, "_");
}

function intValue(value) {
  return Number.isInteger(value) ? value : Number.isFinite(value) ? Math.trunc(value) : 0;
}

function retryCount(startup) {
  const explicit = intValue(startup?.retry_count);
  if (explicit > 0) {
    return explicit;
  }
  return (startup?.attempts ?? []).filter((attempt) => attempt?.retry_scheduled === true).length;
}

function recordServiceScopeContamination(reasons, repoRoot, file) {
  const scope = readJSONIfPossible(file);
  if (!scope || typeof scope !== "object") {
    return;
  }
  const artifact = rel(repoRoot, path.dirname(file));
  for (const service of ["postgres", "object_store"]) {
    const startup = scope?.[service]?.startup;
    const retries = retryCount(startup);
    if (retries > 0) {
      addReason(
        reasons,
        `service_startup_retry service=${service} retries=${retries} artifact=${artifact}`,
      );
    }
    const finalStatus = reasonText(startup?.final_status);
    if (finalStatus && finalStatus !== "pass") {
      addReason(
        reasons,
        `service_startup_failed service=${service} status=${finalStatus} artifact=${artifact}`,
      );
    }
  }
  if (scope.failure) {
    addReason(
      reasons,
      [
        "suite_service_failure",
        `service=${reasonText(scope.failure.service) || "-"}`,
        `stage=${reasonText(scope.failure.stage) || "-"}`,
        `operation=${reasonText(scope.failure.operation) || "-"}`,
        `artifact=${artifact}`,
      ].join(" "),
    );
  }
}

function recordServiceEventContamination(reasons, repoRoot, file) {
  const event = readJSONIfPossible(file);
  if (!event || event.type !== "timing-span") {
    return;
  }
  const details = event.details ?? {};
  const status = reasonText(details.status ?? event.status);
  const artifact = rel(repoRoot, path.dirname(path.dirname(file)));
  if (details.startup_attempt === true && details.retry_scheduled === true) {
    addReason(
      reasons,
      [
        "service_startup_retry",
        `service=${reasonText(details.service) || "-"}`,
        `attempt=${intValue(details.attempt)}`,
        `artifact=${artifact}`,
      ].join(" "),
    );
  }
  if (status && status !== "pass" && details.janitorial !== true) {
    addReason(
      reasons,
      [
        "failed_timing_span",
        `label=${reasonText(details.label) || "-"}`,
        `target=${reasonText(details.target) || "-"}`,
        `artifact=${artifact}`,
      ].join(" "),
    );
  }
}

function addReason(reasons, reason) {
  reasons.set(reason, reason);
}
