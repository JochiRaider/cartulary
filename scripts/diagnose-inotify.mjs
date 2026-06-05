#!/usr/bin/env node

import fs from "node:fs";

const args = new Set(process.argv.slice(2));
const mode = args.has("--json")
  ? "json"
  : args.has("--require-dev-watch-capacity")
    ? "require"
    : "advisory";

const minimumWatches = positiveIntEnv("CARTULARY_INOTIFY_MIN_USER_WATCHES", 524288);
const minimumInstances = positiveIntEnv("CARTULARY_INOTIFY_MIN_USER_INSTANCES", 1024);
const recommendedWatches = positiveIntEnv(
  "CARTULARY_INOTIFY_RECOMMENDED_USER_WATCHES",
  1048576,
);
const recommendedInstances = positiveIntEnv(
  "CARTULARY_INOTIFY_RECOMMENDED_USER_INSTANCES",
  1024,
);

const diagnostics = collectDiagnostics();
const failures = [];
const warnings = [];

if (diagnostics.platform === "linux") {
  if (
    Number.isInteger(diagnostics.max_user_watches) &&
    diagnostics.max_user_watches < minimumWatches
  ) {
    failures.push(
      `fs.inotify.max_user_watches=${diagnostics.max_user_watches} is below ${minimumWatches}`,
    );
  }
  if (
    Number.isInteger(diagnostics.max_user_instances) &&
    diagnostics.max_user_instances < minimumInstances
  ) {
    failures.push(
      `fs.inotify.max_user_instances=${diagnostics.max_user_instances} is below ${minimumInstances}`,
    );
  }
  if (
    Number.isInteger(diagnostics.current_watches) &&
    Number.isInteger(diagnostics.max_user_watches) &&
    diagnostics.max_user_watches > 0 &&
    diagnostics.current_watches / diagnostics.max_user_watches >= 0.9
  ) {
    failures.push(
      `current inotify watches ${diagnostics.current_watches} are at least 90% of max_user_watches`,
    );
  }
  if (
    Number.isInteger(diagnostics.current_instances) &&
    Number.isInteger(diagnostics.max_user_instances) &&
    diagnostics.max_user_instances > 0 &&
    diagnostics.current_instances / diagnostics.max_user_instances >= 0.9
  ) {
    failures.push(
      `current inotify instances ${diagnostics.current_instances} are at least 90% of max_user_instances`,
    );
  }
  if (diagnostics.current_usage_available === false) {
    warnings.push("current inotify usage could not be fully inspected from /proc");
  }
}

diagnostics.minimum_user_watches = minimumWatches;
diagnostics.minimum_user_instances = minimumInstances;
diagnostics.recommended_user_watches = recommendedWatches;
diagnostics.recommended_user_instances = recommendedInstances;
diagnostics.remediation =
  `raise Linux inotify limits, for example: sudo sysctl fs.inotify.max_user_watches=${recommendedWatches}; sudo sysctl fs.inotify.max_user_instances=${recommendedInstances}`;
diagnostics.status = failures.length > 0 ? "fail" : warnings.length > 0 ? "warn" : "pass";
diagnostics.failures = failures;
diagnostics.warnings = warnings;

if (mode === "json") {
  process.stdout.write(`${JSON.stringify(diagnostics, null, 2)}\n`);
  process.exit(0);
}

printHumanSummary(diagnostics);
process.exit(mode === "require" && failures.length > 0 ? 1 : 0);

function collectDiagnostics() {
  const diagnostics = {
    schema_id: "cartulary.inotify_diagnostics.v1",
    platform: process.platform,
    max_user_watches: readIntFile("/proc/sys/fs/inotify/max_user_watches"),
    max_user_instances: readIntFile("/proc/sys/fs/inotify/max_user_instances"),
  };
  if (process.platform !== "linux") {
    return diagnostics;
  }

  let currentInstances = 0;
  let currentWatches = 0;
  let usageAvailable = true;
  try {
    for (const pid of fs.readdirSync("/proc")) {
      if (!/^\d+$/.test(pid)) {
        continue;
      }
      const fdinfoDir = `/proc/${pid}/fdinfo`;
      let fdinfos = [];
      try {
        fdinfos = fs.readdirSync(fdinfoDir);
      } catch {
        continue;
      }
      for (const fdinfo of fdinfos) {
        let text = "";
        try {
          text = fs.readFileSync(`${fdinfoDir}/${fdinfo}`, "utf8");
        } catch {
          usageAvailable = false;
          continue;
        }
        const matches = text.match(/^inotify\b/gm);
        if (matches) {
          currentInstances += 1;
          currentWatches += matches.length;
        }
      }
    }
  } catch {
    usageAvailable = false;
  }

  diagnostics.current_instances = currentInstances;
  diagnostics.current_watches = currentWatches;
  diagnostics.current_usage_available = usageAvailable;
  return diagnostics;
}

function printHumanSummary(diagnostics) {
  if (diagnostics.platform !== "linux") {
    process.stdout.write(
      `ok inotify: platform ${diagnostics.platform} does not use Linux inotify limits\n`,
    );
    return;
  }

  const prefix = diagnostics.status === "fail" ? "missing" : diagnostics.status;
  process.stdout.write(
    `${prefix} inotify: max_user_watches=${valueOrUnknown(
      diagnostics.max_user_watches,
    )} max_user_instances=${valueOrUnknown(
      diagnostics.max_user_instances,
    )} current_watches=${valueOrUnknown(
      diagnostics.current_watches,
    )} current_instances=${valueOrUnknown(diagnostics.current_instances)}\n`,
  );
  for (const warning of diagnostics.warnings) {
    process.stdout.write(`warning inotify: ${warning}\n`);
  }
  for (const failure of diagnostics.failures) {
    process.stdout.write(`missing inotify: ${failure}\n`);
  }
  if (diagnostics.status === "fail") {
    process.stdout.write(`remediation inotify: ${diagnostics.remediation}\n`);
  }
}

function readIntFile(file) {
  try {
    const parsed = Number.parseInt(fs.readFileSync(file, "utf8").trim(), 10);
    return Number.isInteger(parsed) ? parsed : undefined;
  } catch {
    return undefined;
  }
}

function positiveIntEnv(name, fallback) {
  const parsed = Number.parseInt(process.env[name] ?? "", 10);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
}

function valueOrUnknown(value) {
  return Number.isInteger(value) ? String(value) : "unknown";
}
