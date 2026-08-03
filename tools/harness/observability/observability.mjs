import { createHash } from "node:crypto";
import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import { repoRoot } from "../contract/index.mjs";
import { validateCanonicalRun } from "./canonical-evidence.mjs";

export const harnessScope = "cartulary.harness.execution";
function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function sha256(value) {
  return createHash("sha256").update(value).digest("hex");
}

function policy() {
  const owner = readJSON(path.join(repoRoot, "tools/task_surface_owner.json"));
  return owner.observability_policy;
}

export function observabilityRequiredTarget(target) {
  return policy().required_targets.includes(target);
}

export function executionProfileDigests() {
  const host = JSON.stringify({
    architecture: process.arch,
    available_parallelism: os.availableParallelism(),
    logical_cpus: os.cpus().length,
    platform: process.platform,
  });
  return {
    host: sha256(host),
    capacity: sha256(
      readFileSync(path.join(repoRoot, "tools/scheduler_resource_registry.json")),
    ),
    workload: sha256(
      readFileSync(path.join(repoRoot, "tools/harness_work_graph_owner.json")),
    ),
    toolchain: sha256(readFileSync(path.join(repoRoot, "tools/toolchain_pins.json"))),
  };
}

function isCanonicalRun(dir) {
  return ["run-manifest.json", "run-summary.json", "unit-events.ndjson"].every(
    (name) => existsSync(path.join(dir, name)),
  );
}

export function resolveRunDir(resultsDir, runID = "", { allowNewest = true } = {}) {
  const selected = path.resolve(repoRoot, resultsDir);
  if (!existsSync(selected)) throw new Error("RESULTS_DIR does not exist");
  if (runID) {
    const candidate = path.basename(selected) === runID ? selected : path.join(selected, runID);
    if (!isCanonicalRun(candidate)) {
      throw new Error("RUN_ID does not identify canonical retained evidence");
    }
    return candidate;
  }
  if (isCanonicalRun(selected)) return selected;
  if (!allowNewest) {
    throw new Error("RUN_ID is required when RESULTS_DIR names a result root");
  }
  const candidates = readdirSync(selected, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => path.join(selected, entry.name))
    .filter(isCanonicalRun)
    .sort(
      (left, right) =>
        statSync(right).mtimeMs - statSync(left).mtimeMs || right.localeCompare(left),
    );
  if (candidates.length === 0) {
    throw new Error("RESULTS_DIR contains no canonical retained harness run");
  }
  return candidates[0];
}

export function resolveExactRunDir(resultsDir, runID = "") {
  return resolveRunDir(resultsDir, runID, { allowNewest: false });
}

function eventIntervals(events) {
  const starts = new Map();
  const intervals = [];
  for (const event of events) {
    if (event.event === "started") starts.set(event.unit_id, event.monotonic_ms);
    if (["completed", "failed", "cancelled"].includes(event.event)) {
      const start = starts.get(event.unit_id);
      if (start !== undefined) {
        intervals.push({
          unit_id: event.unit_id,
          start_ms: start,
          end_ms: event.monotonic_ms,
          status: event.status,
        });
      }
    }
  }
  return intervals.sort(
    (left, right) =>
      left.start_ms - right.start_ms || left.unit_id.localeCompare(right.unit_id),
  );
}

function otlpProjection(run) {
  const intervals = eventIntervals(run.events);
  const startedNs = BigInt(Date.parse(run.manifest.started_at)) * 1_000_000n;
  const attributes = [
    { key: "cartulary.run.id", value: { stringValue: run.manifest.run_id } },
    { key: "cartulary.target", value: { stringValue: run.manifest.target } },
    { key: "cartulary.graph.digest", value: { stringValue: run.manifest.graph_digest } },
  ];
  const spans = intervals.map((interval) => ({
    traceId: sha256(`${run.manifest.run_id}:trace`).slice(0, 32),
    spanId: sha256(`${run.manifest.run_id}:${interval.unit_id}`).slice(0, 16),
    name: interval.unit_id,
    kind: 1,
    startTimeUnixNano: (startedNs + BigInt(interval.start_ms) * 1_000_000n).toString(),
    endTimeUnixNano: (startedNs + BigInt(interval.end_ms) * 1_000_000n).toString(),
    attributes,
    status: { code: interval.status === "passed" ? 1 : 2 },
  }));
  return {
    traceOTLP: {
      resourceSpans: [{ resource: { attributes }, scopeSpans: [{ scope: { name: harnessScope }, spans }] }],
    },
    metricsOTLP: {
      resourceMetrics: [
        {
          resource: { attributes },
          scopeMetrics: [
            {
              scope: { name: harnessScope },
              metrics: [
                {
                  name: "cartulary.harness.run.wall_duration_ms",
                  gauge: {
                    dataPoints: [
                      {
                        asInt: String(run.summary.wall_duration_ms),
                        timeUnixNano: (
                          startedNs + BigInt(run.summary.wall_duration_ms) * 1_000_000n
                        ).toString(),
                        attributes,
                      },
                    ],
                  },
                },
              ],
            },
          ],
        },
      ],
    },
  };
}

function retainedProjection(runDir) {
  const run = validateCanonicalRun(runDir);
  const sourceDigests = [
    run.manifest.source_digest,
    run.manifest.toolchain_digest,
    run.manifest.system_digest,
    run.manifest.graph_digest,
  ];
  const result = otlpProjection(run);
  return {
    run,
    index: {
      schema_id: "cartulary.harness_canonical_observability.v1",
      status: "complete",
      invocations: [
        {
          target: run.manifest.target,
          run_id: run.manifest.run_id,
          source_digests: sourceDigests,
        },
      ],
    },
    built: [{ result }],
  };
}

export function captureExecutionContext(runDir, metadata = {}) {
  const retained = retainedProjection(runDir);
  return { manifest: retained.run.manifest, metadata };
}

export function loadRetainedExecutionContext(runDir) {
  return captureExecutionContext(runDir);
}

export function reconstructObservability(runDir) {
  return retainedProjection(runDir);
}

export function loadRetainedObservability(runDir) {
  return retainedProjection(runDir);
}

export function writePartialObservability(_runDir, error) {
  return {
    status: "partial",
    diagnostic: `${error instanceof Error ? error.name : "Error"}:canonical-evidence-unavailable`,
  };
}

export function finalizeObservabilitySafely(runDir) {
  try {
    const retained = retainedProjection(runDir);
    return { status: "complete", retained };
  } catch (error) {
    // Step and target wrappers can finish before the graph scheduler publishes
    // its atomic evidence set. They must not create a second timing authority.
    return {
      status: "skipped",
      diagnostic: `${error instanceof Error ? error.name : "Error"}:canonical-evidence-pending`,
    };
  }
}

export function deterministicBytes(result) {
  return `${JSON.stringify(result, null, 2)}\n`;
}

export function printObservabilityPerformance(runDir, target = "") {
  const retained = retainedProjection(runDir);
  const summary = target
    ? retained.run.targetSummaries.get(target)
    : retained.run.targetSummaries.get(retained.run.manifest.target);
  if (!summary) throw new Error("selected target has no canonical target summary");
  const line = [
    `target=${summary.target}`,
    `status=${summary.status}`,
    `inclusive_wall_ms=${summary.inclusive_wall_ms}`,
    `exclusive_wall_ms=${summary.exclusive_wall_ms}`,
    `units=${summary.unit_ids.length}`,
  ].join(" ");
  process.stdout.write(`${line}\n`);
  return summary;
}
