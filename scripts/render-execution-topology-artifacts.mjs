#!/usr/bin/env node
import { readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  defaultExecutionTopologyManifestPath,
  loadExecutionTopology,
  renderBrowserBatchManifest,
  renderCheckScheduleManifest,
  renderTaskSurfaceManifest,
} from "./lib/execution-topology.mjs";
import { collectTaskSurfaceManifestErrors, renderTaskSurfaceMake } from "./lib/task-surface.mjs";
import { renderServiceBackedScheduleManifest } from "./render-service-backed-schedule-manifest.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");

function usage() {
  throw new Error("usage: render-execution-topology-artifacts.mjs [--check] [--topology <path>]");
}

function parseArgs(argv) {
  const options = {
    check: false,
    topology: process.env.CARTULARY_EXECUTION_TOPOLOGY_MANIFEST ?? defaultExecutionTopologyManifestPath,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--check") {
      options.check = true;
      continue;
    }
    if (arg === "--topology") {
      options.topology = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.topology) {
    usage();
  }
  return options;
}

function resolveRepoPath(value) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function serializeJSON(value) {
  return `${JSON.stringify(value, null, 2)}\n`;
}

function outputEntry(file, content) {
  const outputPath = resolveRepoPath(file);
  return {
    file,
    outputPath,
    content,
  };
}

function renderArtifacts(options) {
  const topology = loadExecutionTopology({ manifestPath: options.topology });
  const taskSurfaceManifest = renderTaskSurfaceManifest(topology);
  const browserBatchManifest = renderBrowserBatchManifest(topology);
  const serviceBackedScheduleManifest = renderServiceBackedScheduleManifest({
    topology: options.topology,
    topologyObject: topology,
  });
  const taskSurfaceErrors = collectTaskSurfaceManifestErrors(taskSurfaceManifest, {
    browserBatchManifest,
    serviceBackedScheduleManifest,
  });
  if (taskSurfaceErrors.length > 0) {
    throw new Error(
      `rendered task surface is invalid:\n${taskSurfaceErrors.map((error) => `  - ${error}`).join("\n")}`,
    );
  }
  return [
    outputEntry(topology.generatedOutputs.task_surface_manifest, serializeJSON(taskSurfaceManifest)),
    outputEntry(
      topology.generatedOutputs.check_schedule_manifest,
      serializeJSON(renderCheckScheduleManifest(topology)),
    ),
    outputEntry(
      topology.generatedOutputs.browser_e2e_batch_manifest,
      serializeJSON(browserBatchManifest),
    ),
    outputEntry(
      topology.generatedOutputs.service_backed_schedule_manifest,
      serializeJSON(serviceBackedScheduleManifest),
    ),
    outputEntry(topology.generatedOutputs.task_surface_make, renderTaskSurfaceMake(taskSurfaceManifest)),
  ];
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const artifacts = renderArtifacts(options);
  const stale = [];
  for (const artifact of artifacts) {
    if (options.check) {
      const existing = readFileSync(artifact.outputPath, "utf8");
      if (existing !== artifact.content) {
        stale.push(artifact.file);
      }
      continue;
    }
    writeFileSync(artifact.outputPath, artifact.content);
  }
  if (stale.length > 0) {
    throw new Error(`${stale.join(", ")} stale; run make phase-schedules`);
  }
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  console.error(`execution topology render failed: ${message}`);
  process.exit(1);
}
