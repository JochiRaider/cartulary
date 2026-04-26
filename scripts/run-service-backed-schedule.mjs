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

function usage() {
  process.stderr.write(
    "usage: run-service-backed-schedule.mjs --target <target> --jobs <n> [--manifest <path>]\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = {
    manifest: defaultManifestPath,
    target: "",
    jobs: "",
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
  if (manifest.schema_id !== "cartulary.service_backed_schedule.v1") {
    throw new Error(`${manifestPath} must declare schema_id=cartulary.service_backed_schedule.v1`);
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

function validateChild(scheduleTarget, child, index) {
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
  const resourceTags = normalizeStringArray(child.resource_tags ?? [], `${label} resource_tags`);
  const exclusiveTags = normalizeStringArray(child.exclusive_tags ?? [], `${label} exclusive_tags`);
  const normalized = {
    target: child.target.trim(),
    kind: child.kind,
    weight: child.weight,
    resourceTags,
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
  if (!normalized.resourceTags.includes("browser")) {
    throw new Error(`${label} browser target ${normalized.target} must declare browser resource tag`);
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
  const children = schedule.children.map((child, index) => validateChild(target, child, index));
  const duplicates = children
    .map((child) => child.target)
    .filter((targetName, index, targets) => targets.indexOf(targetName) !== index);
  if (duplicates.length > 0) {
    throw new Error(`schedule ${target} contains duplicate child targets: ${duplicates.join(", ")}`);
  }
  return {
    target,
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

function hasActiveResource(tag, activeResourceTags) {
  return (activeResourceTags.get(tag) ?? 0) > 0;
}

function canStart(child, activeExclusiveTags, activeResourceTags) {
  const exclusiveConflict = child.exclusiveTags.some(
    (tag) => activeExclusiveTags.has(tag) || hasActiveResource(tag, activeResourceTags),
  );
  if (exclusiveConflict) {
    return false;
  }
  return child.resourceTags.every((tag) => !activeExclusiveTags.has(tag));
}

function addResourceTags(child, activeResourceTags) {
  for (const tag of child.resourceTags) {
    activeResourceTags.set(tag, (activeResourceTags.get(tag) ?? 0) + 1);
  }
}

function removeResourceTags(child, activeResourceTags) {
  for (const tag of child.resourceTags) {
    const next = (activeResourceTags.get(tag) ?? 0) - 1;
    if (next <= 0) {
      activeResourceTags.delete(tag);
    } else {
      activeResourceTags.set(tag, next);
    }
  }
}

async function runSchedule({ schedule, jobs, makeBin, testOutputScript }) {
  const childrenCsv = schedule.children.map((child) => child.target).join(",");
  const tempDir = await mkdtemp(path.join(os.tmpdir(), "cartulary-service-backed-schedule-"));
  const pending = [...schedule.executionChildren];
  const running = new Map();
  const activeExclusiveTags = new Set();
  const activeResourceTags = new Map();
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
      addResourceTags(child, activeResourceTags);
      const logFile = path.join(tempDir, `${String(started).padStart(2, "0")}-${child.target}.log`);
      const promise = runChild(makeBin, child.target, logFile);
      running.set(promise, child);
      promise.then(() => {
        for (const tag of child.exclusiveTags) {
          activeExclusiveTags.delete(tag);
        }
        removeResourceTags(child, activeResourceTags);
      });
    };

    while (pending.length > 0 || running.size > 0) {
      while (running.size < jobs) {
        const nextIndex = pending.findIndex((candidate) =>
          canStart(candidate, activeExclusiveTags, activeResourceTags),
        );
        if (nextIndex === -1) {
          break;
        }
        const [child] = pending.splice(nextIndex, 1);
        await startChild(child);
      }

      if (running.size === 0) {
        throw new Error(`scheduler deadlock for ${schedule.target}; check exclusive_tags`);
      }

      const result = await Promise.race(running.keys());
      for (const [promise, candidate] of running.entries()) {
        if (candidate.target === result.target) {
          running.delete(promise);
          break;
        }
      }
      await replayLog(result.logFile, result.status === 0 ? process.stdout : process.stderr);
      if (result.status !== 0 && firstFailure === 0) {
        firstFailure = result.status;
      }
    }

    const requestedStatus = firstFailure === 0 ? "pass" : "fail";
    await runLifecycle(testOutputScript, [
      "target-summary",
      schedule.target,
      requestedStatus,
      "--children",
      childrenCsv,
    ], requestedStatus === "pass" ? process.stdout : process.stderr);
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
  });
  process.exitCode = status;
}

main().catch((error) => {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
});
