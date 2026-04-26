#!/usr/bin/env node
import { spawn } from "node:child_process";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import { createReadStream, createWriteStream } from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { findTargetDescriptor } from "./lib/target-plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const defaultManifestPath = path.join(repoRoot, "tools", "service_backed_schedule_manifest.json");
const validKinds = new Set(["backend", "browser"]);
const supportedSchemaIDs = new Set([
  "cartulary.service_backed_schedule.v1",
  "cartulary.service_backed_schedule.v2",
]);

function usage() {
  process.stderr.write(
    "usage: run-service-backed-schedule.mjs --target <target> --jobs <n> [--manifest <path>] [--defer-summary]\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = {
    manifest: defaultManifestPath,
    target: "",
    jobs: "",
    deferSummary: false,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--target") {
      options.target = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--jobs") {
      options.jobs = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--manifest") {
      options.manifest = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--defer-summary") {
      options.deferSummary = true;
      continue;
    }
    usage();
  }
  if (!options.target || !options.manifest || !options.jobs) {
    usage();
  }
  const jobs = Number.parseInt(options.jobs, 10);
  if (!Number.isInteger(jobs) || jobs < 1) {
    throw new Error(`--jobs must be a positive integer, got ${JSON.stringify(options.jobs)}`);
  }
  return { ...options, jobs };
}

function isDryRun() {
  const flags = ` ${process.env.MAKEFLAGS ?? ""} `;
  return flags.includes(" n") || flags.includes(" --just-print") || flags.includes(" --dry-run");
}

async function loadManifest(file) {
  const manifestPath = path.isAbsolute(file) ? file : path.join(repoRoot, file);
  const manifest = JSON.parse(await readFile(manifestPath, "utf8"));
  if (!supportedSchemaIDs.has(manifest.schema_id)) {
    throw new Error(
      `${manifestPath} must declare schema_id cartulary.service_backed_schedule.v1 or cartulary.service_backed_schedule.v2`,
    );
  }
  if (!Array.isArray(manifest.schedules)) {
    throw new Error(`${manifestPath} must declare schedules[]`);
  }
  return { manifest, manifestPath };
}

function normalizeStringArray(value, label) {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  return value.map((entry) => {
    if (typeof entry !== "string" || entry.trim() === "") {
      throw new Error(`${label} entries must be non-empty strings`);
    }
    return entry.trim();
  });
}

function normalizeResourceLimits(value, label, schemaID) {
  if (value === undefined && schemaID === "cartulary.service_backed_schedule.v1") {
    return new Map();
  }
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} resource_limits must be an object`);
  }

  const resourceLimits = new Map();
  for (const [resource, limit] of Object.entries(value)) {
    const normalizedResource = resource.trim();
    if (normalizedResource === "") {
      throw new Error(`${label} resource_limits keys must be non-empty strings`);
    }
    if (!Number.isInteger(limit) || limit < 1) {
      throw new Error(`${label} resource_limits.${normalizedResource} must be a positive integer`);
    }
    resourceLimits.set(normalizedResource, limit);
  }
  return resourceLimits;
}

function normalizeUniqueStringArray(value, label) {
  const entries = normalizeStringArray(value, label);
  const duplicates = entries.filter((entry, index) => entries.indexOf(entry) !== index);
  if (duplicates.length > 0) {
    throw new Error(`${label} contains duplicate entries: ${[...new Set(duplicates)].join(", ")}`);
  }
  return entries;
}

function validateChild(scheduleTarget, child, index, resourceLimits, schemaID) {
  const label = `${scheduleTarget} child ${index + 1}`;
  if (!child || typeof child !== "object" || Array.isArray(child)) {
    throw new Error(`${label} must be an object`);
  }
  if (typeof child.target !== "string" || child.target.trim() === "") {
    throw new Error(`${label} must declare target`);
  }
  if (!validKinds.has(child.kind)) {
    throw new Error(`${label} ${child.target} must declare kind backend or browser`);
  }
  if (!Number.isFinite(child.weight) || child.weight < 0) {
    throw new Error(`${label} ${child.target} must declare non-negative weight`);
  }
  const legacyResourceTags = normalizeUniqueStringArray(
    child.resource_tags ?? [],
    `${label} resource_tags`,
  );
  const exclusiveTags = normalizeUniqueStringArray(
    child.exclusive_tags ?? [],
    `${label} exclusive_tags`,
  );
  const resourceClaims = normalizeUniqueStringArray(
    schemaID === "cartulary.service_backed_schedule.v2"
      ? (child.resource_claims ?? [])
      : legacyResourceTags,
    `${label} resource_claims`,
  );
  if (schemaID === "cartulary.service_backed_schedule.v2" && child.resource_tags !== undefined) {
    throw new Error(`${label} ${child.target} must use resource_claims, not resource_tags`);
  }
  if (schemaID === "cartulary.service_backed_schedule.v2" && child.exclusive_tags !== undefined) {
    throw new Error(`${label} ${child.target} must not declare legacy exclusive_tags`);
  }
  for (const resource of resourceClaims) {
    if (resourceLimits.size > 0 && !resourceLimits.has(resource)) {
      throw new Error(
        `${label} ${child.target} resource_claims entry ${resource} is not declared in resource_limits`,
      );
    }
  }
  const normalized = {
    target: child.target.trim(),
    kind: child.kind,
    weight: child.weight,
    resourceClaims,
    exclusiveTags,
    order: index,
  };

  if (normalized.kind === "backend") {
    const descriptor = findTargetDescriptor(normalized.target);
    if (!descriptor) {
      throw new Error(`${label} backend target ${normalized.target} is not in target-plan`);
    }
    if (!descriptor.serviceBacked) {
      throw new Error(`${label} backend target ${normalized.target} is not service-backed`);
    }
    if (scheduleTarget === "check-service-backed" && descriptor.checkServiceBackedSafe !== true) {
      throw new Error(
        `${label} backend target ${normalized.target} is not check-service-backed safe`,
      );
    }
    return normalized;
  }

  if (!normalized.target.startsWith("browser-e2e-")) {
    throw new Error(`${label} browser target ${normalized.target} must be a browser-e2e target`);
  }
  if (
    !normalized.resourceClaims.includes("browser") &&
    !normalized.resourceClaims.includes("browser-webserver")
  ) {
    throw new Error(
      `${label} browser target ${normalized.target} must declare browser or browser-webserver resource claim`,
    );
  }
  return normalized;
}

function findSchedule(manifest, target) {
  const matches = manifest.schedules.filter((schedule) => schedule?.target === target);
  if (matches.length !== 1) {
    throw new Error(`expected exactly one schedule for ${target}, found ${matches.length}`);
  }
  const [schedule] = matches;
  if (!Array.isArray(schedule.children) || schedule.children.length === 0) {
    throw new Error(`schedule ${target} must declare at least one child`);
  }
  const resourceLimits = normalizeResourceLimits(
    schedule.resource_limits,
    `schedule ${target}`,
    manifest.schema_id,
  );
  const children = schedule.children.map((child, index) =>
    validateChild(target, child, index, resourceLimits, manifest.schema_id),
  );
  const duplicates = children
    .map((child) => child.target)
    .filter((targetName, index, targets) => targets.indexOf(targetName) !== index);
  if (duplicates.length > 0) {
    throw new Error(`schedule ${target} contains duplicate child targets: ${duplicates.join(", ")}`);
  }
  return {
    target,
    resourceLimits,
    usesResourceCapacity: resourceLimits.size > 0,
    children,
    executionChildren: [...children].sort(
      (left, right) => right.weight - left.weight || left.target.localeCompare(right.target),
    ),
  };
}

function runLifecycle(testOutputScript, args, stream = process.stdout) {
  return new Promise((resolve, reject) => {
    const child = spawn(testOutputScript, args, {
      cwd: repoRoot,
      env: process.env,
      stdio: ["ignore", "pipe", "pipe"],
    });
    child.stdout.pipe(stream, { end: false });
    child.stderr.pipe(process.stderr, { end: false });
    child.on("error", reject);
    child.on("close", (status) => {
      if (status === 0) {
        resolve();
        return;
      }
      reject(new Error(`${testOutputScript} ${args.join(" ")} exited ${status}`));
    });
  });
}

function runPostgresFixtureBudgetCheck(targets) {
  return new Promise((resolve, reject) => {
    const child = spawn(
      process.execPath,
      [path.join(repoRoot, "scripts", "check-postgres-fixture-budget.mjs"), "--targets", targets.join(",")],
      {
        cwd: repoRoot,
        env: process.env,
        stdio: ["ignore", "pipe", "pipe"],
      },
    );
    child.stdout.pipe(process.stdout, { end: false });
    child.stderr.pipe(process.stderr, { end: false });
    child.on("error", reject);
    child.on("close", (status) => {
      if (status === 0) {
        resolve(0);
        return;
      }
      resolve(status ?? 1);
    });
  });
}

async function replayLog(file, stream) {
  await new Promise((resolve, reject) => {
    const reader = createReadStream(file);
    reader.on("error", reject);
    reader.on("end", resolve);
    reader.pipe(stream, { end: false });
  });
}

function childCommand(makeBin, target) {
  return [makeBin, ["--no-print-directory", "--output-sync=target", target]];
}

function runChild(makeBin, target, logFile) {
  const [command, args] = childCommand(makeBin, target);
  return new Promise((resolve) => {
    const log = createWriteStream(logFile);
    let settled = false;
    const finish = (status) => {
      if (settled) {
        return;
      }
      settled = true;
      log.end(() => resolve({ target, status }));
    };
    const child = spawn(command, args, {
      cwd: repoRoot,
      env: process.env,
      stdio: ["ignore", "pipe", "pipe"],
    });
    child.stdout.pipe(log, { end: false });
    child.stderr.pipe(log, { end: false });
    child.on("error", (error) => {
      log.write(`${error.message}\n`);
      finish(127);
    });
    child.on("close", (status) => {
      finish(status ?? 1);
    });
  }).then((result) => ({ target, status: result.status, logFile }));
}

function hasActiveResource(resource, activeResourceClaims) {
  return (activeResourceClaims.get(resource) ?? 0) > 0;
}

function hasResourceCapacity(child, resourceLimits, activeResourceClaims) {
  for (const resource of child.resourceClaims) {
    const limit = resourceLimits.get(resource);
    if (limit !== undefined && (activeResourceClaims.get(resource) ?? 0) + 1 > limit) {
      return false;
    }
  }
  return true;
}

function canStart(child, resourceLimits, activeExclusiveTags, activeResourceClaims) {
  const exclusiveConflict = child.exclusiveTags.some(
    (tag) => activeExclusiveTags.has(tag) || hasActiveResource(tag, activeResourceClaims),
  );
  if (exclusiveConflict) {
    return false;
  }
  if (child.resourceClaims.some((resource) => activeExclusiveTags.has(resource))) {
    return false;
  }
  return hasResourceCapacity(child, resourceLimits, activeResourceClaims);
}

function addResourceClaims(child, activeResourceClaims) {
  for (const resource of child.resourceClaims) {
    activeResourceClaims.set(resource, (activeResourceClaims.get(resource) ?? 0) + 1);
  }
}

function removeResourceClaims(child, activeResourceClaims) {
  for (const resource of child.resourceClaims) {
    const next = (activeResourceClaims.get(resource) ?? 0) - 1;
    if (next <= 0) {
      activeResourceClaims.delete(resource);
    } else {
      activeResourceClaims.set(resource, next);
    }
  }
}

function formatResourceMap(values) {
  const entries = Array.from(values.entries()).sort((left, right) => left[0].localeCompare(right[0]));
  if (entries.length === 0) {
    return "{}";
  }
  return `{${entries.map(([key, value]) => `${key}:${value}`).join(",")}}`;
}

function formatResourceLimits(resourceLimits) {
  return formatResourceMap(resourceLimits);
}

function formatActiveResourceClaims(activeResourceClaims) {
  return formatResourceMap(activeResourceClaims);
}

function formatActiveExclusiveTags(activeExclusiveTags) {
  const tags = Array.from(activeExclusiveTags).sort();
  return tags.length === 0 ? "[]" : `[${tags.join(",")}]`;
}

function formatBlockedChildren(children) {
  return children
    .map((child) => `${child.target} claims=[${child.resourceClaims.join(",")}] exclusive=[${child.exclusiveTags.join(",")}]`)
    .join("; ");
}

async function runSchedule({ schedule, jobs, makeBin, testOutputScript, deferSummary }) {
  const childrenCsv = schedule.children.map((child) => child.target).join(",");
  const backendBudgetTargets = schedule.children
    .filter((child) => child.kind === "backend")
    .map((child) => child.target);
  const tempDir = await mkdtemp(path.join(os.tmpdir(), "cartulary-service-backed-schedule-"));
  const pending = [...schedule.executionChildren];
  const running = new Map();
  const activeExclusiveTags = new Set();
  const activeResourceClaims = new Map();
  let started = 0;
  let firstFailure = 0;

  try {
    await runLifecycle(testOutputScript, [
      "target-start",
      schedule.target,
      "--children",
      childrenCsv,
      "--service-backed",
      "1",
    ]);

    const startChild = async (child) => {
      started += 1;
      await runLifecycle(testOutputScript, [
        "step-start",
        schedule.target,
        String(started),
        String(schedule.children.length),
        child.target,
        "--mode",
        "scheduler",
        "--jobs",
        String(jobs),
      ]);
      for (const tag of child.exclusiveTags) {
        activeExclusiveTags.add(tag);
      }
      addResourceClaims(child, activeResourceClaims);
      const logFile = path.join(tempDir, `${String(started).padStart(2, "0")}-${child.target}.log`);
      const promise = runChild(makeBin, child.target, logFile);
      running.set(promise, child);
    };

    while (pending.length > 0 || running.size > 0) {
      while (schedule.usesResourceCapacity || running.size < jobs) {
        const nextIndex = pending.findIndex((candidate) =>
          canStart(candidate, schedule.resourceLimits, activeExclusiveTags, activeResourceClaims),
        );
        if (nextIndex === -1) {
          break;
        }
        const [child] = pending.splice(nextIndex, 1);
        await startChild(child);
      }

      if (running.size === 0) {
        throw new Error(
          `scheduler deadlock for ${schedule.target}; pending=${formatBlockedChildren(pending)} active_resource_claims=${formatActiveResourceClaims(activeResourceClaims)} active_exclusive_tags=${formatActiveExclusiveTags(activeExclusiveTags)} resource_limits=${formatResourceLimits(schedule.resourceLimits)}`,
        );
      }

      const result = await Promise.race(running.keys());
      for (const [promise, candidate] of running.entries()) {
        if (candidate.target === result.target) {
          running.delete(promise);
          for (const tag of candidate.exclusiveTags) {
            activeExclusiveTags.delete(tag);
          }
          removeResourceClaims(candidate, activeResourceClaims);
          break;
        }
      }
      await replayLog(result.logFile, result.status === 0 ? process.stdout : process.stderr);
      if (result.status !== 0 && firstFailure === 0) {
        firstFailure = result.status;
      }
    }

    if (firstFailure === 0 && backendBudgetTargets.length > 0) {
      firstFailure = await runPostgresFixtureBudgetCheck(backendBudgetTargets);
    }

    const requestedStatus = firstFailure === 0 ? "pass" : "fail";
    if (!deferSummary) {
      await runLifecycle(testOutputScript, [
        "target-summary",
        schedule.target,
        requestedStatus,
        "--children",
        childrenCsv,
      ], requestedStatus === "pass" ? process.stdout : process.stderr);
    }
    return firstFailure;
  } finally {
    await rm(tempDir, { recursive: true, force: true });
  }
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const { manifest, manifestPath } = await loadManifest(options.manifest);
  const schedule = findSchedule(manifest, options.target);
  const makeBin = process.env.MAKE || "make";
  const testOutputScript = process.env.TEST_OUTPUT_SCRIPT || path.join(repoRoot, "scripts", "lib", "test-output.sh");

  if (isDryRun()) {
    process.stdout.write(
      `[DRY-RUN] ${options.target} jobs=${options.jobs} manifest=${path.relative(repoRoot, manifestPath)} children=${schedule.children.map((child) => child.target).join(",")}\n`,
    );
    return;
  }

  const status = await runSchedule({
    schedule,
    jobs: Math.min(options.jobs, schedule.children.length),
    makeBin,
    testOutputScript,
    deferSummary: options.deferSummary,
  });
  process.exitCode = status;
}

main().catch((error) => {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
});
