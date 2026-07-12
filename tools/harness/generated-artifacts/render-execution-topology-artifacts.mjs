#!/usr/bin/env node
import {
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { createHash } from "node:crypto";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  defaultExecutionTopologyManifestPath,
  loadExecutionTopology,
  renderBrowserBatchManifest,
  renderCheckScheduleManifest,
  renderTaskSurfaceManifest,
} from "./execution-topology.mjs";
import { validateAllPhaseSlicePlans } from "../phase-accounting/index.mjs";
import {
  collectTaskSurfaceManifestErrors,
  renderTaskSurfaceMake,
  renderTaskSurfaceMakeRuntime,
  collectTaskSurfaceMakeDensityErrors,
} from "./task-surface/index.mjs";
import {
  expandServiceBackedSchedule,
  expandServiceBackedScheduleForCheck,
} from "../execution/service-backed/index.mjs";
import { renderServiceBackedScheduleManifest } from "./render-service-backed-schedule-manifest.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");
export const renderIndexSchemaID = "cartulary.execution_topology_render_index.v1";
const renderCacheSchemaID = "cartulary.execution_topology_render_cache.v1";
const generatorVersion = 1;
const cacheDir = process.env.CARTULARY_EXECUTION_TOPOLOGY_RENDER_CACHE_DIR
  ? path.resolve(process.env.CARTULARY_EXECUTION_TOPOLOGY_RENDER_CACHE_DIR)
  : path.join(repoRoot, ".cache", "cartulary", "execution-topology-render");
const renderedOutputKeys = [
  "task_surface_manifest",
  "browser_e2e_batch_manifest",
  "scheduler_manifest",
  "task_surface_make",
  "task_surface_runtime_make",
];

function usage() {
  throw new Error(
    "usage: render-execution-topology-artifacts.mjs [--check|--full-check] [--topology <path>]",
  );
}

function parseArgs(argv) {
  const options = {
    check: false,
    fullCheck: false,
    topology: process.env.CARTULARY_EXECUTION_TOPOLOGY_MANIFEST ?? defaultExecutionTopologyManifestPath,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--check") {
      options.check = true;
      continue;
    }
    if (arg === "--full-check") {
      options.fullCheck = true;
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
  if (options.check && options.fullCheck) {
    usage();
  }
  return options;
}

function resolveRepoPath(value) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function repoDisplayPath(file) {
  const resolved = path.resolve(file);
  const relative = path.relative(repoRoot, resolved);
  if (!relative.startsWith("..") && !path.isAbsolute(relative)) {
    return relative.split(path.sep).join("/");
  }
  return resolved;
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function serializeJSON(value) {
  return `${JSON.stringify(value, null, 2)}\n`;
}

function hashContent(content) {
  return `sha256:${createHash("sha256").update(content).digest("hex")}`;
}

function hashFile(file) {
  return hashContent(readFileSync(file));
}

function hashStructured(value) {
  return hashContent(serializeJSON(value));
}

function outputEntry(file, content) {
  const outputPath = resolveRepoPath(file);
  return {
    file,
    outputPath,
    content,
  };
}

function collectFiles(root, predicate) {
  if (!existsSync(root)) {
    return [];
  }
  const files = [];
  for (const entry of readdirSync(root, { withFileTypes: true })) {
    const entryPath = path.join(root, entry.name);
    if (entry.isDirectory()) {
      files.push(...collectFiles(entryPath, predicate));
      continue;
    }
    if (entry.isFile() && predicate(entryPath)) {
      files.push(entryPath);
    }
  }
  return files.sort((left, right) => repoDisplayPath(left).localeCompare(repoDisplayPath(right)));
}

function phaseManifestRoot() {
  return process.env.CARTULARY_PHASE_MANIFEST_ROOT
    ? path.resolve(process.env.CARTULARY_PHASE_MANIFEST_ROOT)
    : repoRoot;
}

function addFileInput(inputs, seen, role, file) {
  const resolved = path.resolve(file);
  if (seen.has(resolved)) {
    return;
  }
  const stat = statSync(resolved);
  if (!stat.isFile()) {
    throw new Error(`${repoDisplayPath(resolved)} is not a file`);
  }
  seen.add(resolved);
  inputs.push({
    path: repoDisplayPath(resolved),
    role,
    hash: hashFile(resolved),
  });
}

function addRequiredRepoInput(inputs, seen, role, file) {
  addFileInput(inputs, seen, role, path.join(repoRoot, file));
}

function collectActivePhaseManifestInputs(inputs, seen) {
  const manifestRoot = phaseManifestRoot();
  const registryPath = path.join(manifestRoot, "tools", "phase_registry.json");
  addFileInput(inputs, seen, "phase_registry", registryPath);
  const registry = readJSON(registryPath);
  for (const entry of registry.phases ?? []) {
    if (entry?.status !== "active") {
      continue;
    }
    if (typeof entry.manifest_path !== "string" || entry.manifest_path.trim() === "") {
      throw new Error(`active phase ${entry?.phase ?? "<unknown>"} must declare manifest_path`);
    }
    addFileInput(inputs, seen, "active_phase_manifest", path.join(manifestRoot, entry.manifest_path));
  }
}

function collectRendererSourceInputs(inputs, seen) {
  // The rendered schedules depend on the planning modules, not only their two
  // top-level renderers.  Keep the complete production planning surface in the
  // render index so a helper change cannot leave an apparently current manifest
  // behind in the render cache.
  const rendererSourceRoots = [
    "tools/harness/generated-artifacts",
    "tools/harness/execution",
    "tools/harness/scheduler",
    "tools/harness/browser",
    "tools/harness/backend",
    "tools/harness/phase-accounting",
    "scripts/lib",
  ];
  const isProductionModule = (candidate) =>
    candidate.endsWith(".mjs") && !candidate.split(path.sep).includes("tests");
  for (const root of rendererSourceRoots) {
    for (const file of collectFiles(path.join(repoRoot, root), isProductionModule)) {
      addFileInput(inputs, seen, "renderer_source", file);
    }
  }
  addRequiredRepoInput(inputs, seen, "renderer_source", "tools/harness/runtime-binary-registry.mjs");
}

function collectDurationBaselineInputs(inputs, seen, topologyRaw, topologyPath) {
  for (const file of [
    "tools/browser_e2e_duration_baselines.json",
    "tools/go_test_duration_baselines.json",
    "tools/harness_smoke_duration_baselines.json",
    "tools/service_backed_make_target_duration_baselines.json",
  ]) {
    addRequiredRepoInput(inputs, seen, "duration_baseline", file);
  }
  const configuredServiceBaseline =
    topologyRaw?.service_backed_schedules?.defaults?.make_target_duration_baseline;
  if (typeof configuredServiceBaseline === "string" && configuredServiceBaseline.trim() !== "") {
    const baselinePath = path.isAbsolute(configuredServiceBaseline)
      ? configuredServiceBaseline
      : path.join(path.dirname(topologyPath), configuredServiceBaseline);
    addFileInput(inputs, seen, "duration_baseline", baselinePath);
  }
}

export function collectRenderInputs(options = {}) {
  const topologyPath = resolveRepoPath(
    options.topology ??
      process.env.CARTULARY_EXECUTION_TOPOLOGY_MANIFEST ??
      defaultExecutionTopologyManifestPath,
  );
  const topologyRaw = readJSON(topologyPath);
  const inputs = [];
  const seen = new Set();
  addFileInput(inputs, seen, "execution_topology", topologyPath);
  const taskSurfaceOwner = topologyRaw?.task_surface_owner;
  if (typeof taskSurfaceOwner !== "string" || taskSurfaceOwner.trim() === "") {
    throw new Error("task_surface_owner must be a non-empty repo-local path");
  }
  addRequiredRepoInput(inputs, seen, "task_surface_owner", taskSurfaceOwner);
  addRequiredRepoInput(inputs, seen, "scheduler_resource_registry", "tools/scheduler_resource_registry.json");
  collectActivePhaseManifestInputs(inputs, seen);
  collectDurationBaselineInputs(inputs, seen, topologyRaw, topologyPath);
  collectRendererSourceInputs(inputs, seen);
  inputs.sort((left, right) => left.path.localeCompare(right.path) || left.role.localeCompare(right.role));
  return {
    input_digest: hashStructured(inputs),
    inputs,
    topologyRaw,
    topologyPath,
  };
}

function renderIndexPath(topologyRaw) {
  const file = topologyRaw?.generated_outputs?.execution_topology_render_index;
  if (typeof file !== "string" || file.trim() === "") {
    throw new Error("generated_outputs.execution_topology_render_index must be a non-empty string");
  }
  return resolveRepoPath(file);
}

function expectedRenderedOutputFiles(topologyRaw) {
  const generatedOutputs = topologyRaw?.generated_outputs;
  if (!generatedOutputs || typeof generatedOutputs !== "object" || Array.isArray(generatedOutputs)) {
    throw new Error("generated_outputs must be an object");
  }
  return renderedOutputKeys.map((key) => {
    const file = generatedOutputs[key];
    if (typeof file !== "string" || file.trim() === "") {
      throw new Error(`generated_outputs.${key} must be a non-empty string`);
    }
    return file;
  });
}

function outputHashes(artifacts) {
  return Object.fromEntries(
    artifacts
      .map((artifact) => [artifact.file, hashContent(artifact.content)])
      .sort(([left], [right]) => left.localeCompare(right)),
  );
}

export function buildRenderIndex({ inputInfo, artifacts }) {
  return {
    schema_id: renderIndexSchemaID,
    generator: "tools/harness/generated-artifacts/render-execution-topology-artifacts.mjs",
    generator_version: generatorVersion,
    input_digest: inputInfo.input_digest,
    inputs: inputInfo.inputs,
    outputs: outputHashes(artifacts),
  };
}

function validateRenderIndex(index, indexPath) {
  if (!index || typeof index !== "object" || Array.isArray(index)) {
    throw new Error(`${repoDisplayPath(indexPath)} must be an object`);
  }
  if (index.schema_id !== renderIndexSchemaID) {
    throw new Error(`${repoDisplayPath(indexPath)} must declare schema_id ${renderIndexSchemaID}`);
  }
  if (index.generator_version !== generatorVersion) {
    throw new Error(`${repoDisplayPath(indexPath)} generator_version is stale; run make phase-schedules`);
  }
  if (typeof index.input_digest !== "string" || !index.input_digest.startsWith("sha256:")) {
    throw new Error(`${repoDisplayPath(indexPath)} input_digest must be a sha256 digest`);
  }
  if (!Array.isArray(index.inputs)) {
    throw new Error(`${repoDisplayPath(indexPath)} inputs must be an array`);
  }
  if (!index.outputs || typeof index.outputs !== "object" || Array.isArray(index.outputs)) {
    throw new Error(`${repoDisplayPath(indexPath)} outputs must be an object`);
  }
}

function changedInputs(currentInputs, indexedInputs) {
  const indexedByPath = new Map(indexedInputs.map((input) => [input.path, input]));
  const changes = [];
  for (const input of currentInputs) {
    const indexed = indexedByPath.get(input.path);
    if (!indexed) {
      changes.push(`${input.path} added`);
      continue;
    }
    if (indexed.hash !== input.hash || indexed.role !== input.role) {
      changes.push(`${input.path} changed`);
    }
    indexedByPath.delete(input.path);
  }
  for (const input of indexedByPath.values()) {
    changes.push(`${input.path} removed`);
  }
  return changes;
}

export function quickCheckRenderIndex(options = {}) {
  const inputInfo = collectRenderInputs(options);
  const indexPath = renderIndexPath(inputInfo.topologyRaw);
  if (!existsSync(indexPath)) {
    throw new Error(`${repoDisplayPath(indexPath)} missing; run make phase-schedules`);
  }
  const index = readJSON(indexPath);
  validateRenderIndex(index, indexPath);
  const expectedOutputs = expectedRenderedOutputFiles(inputInfo.topologyRaw).sort();
  const indexedOutputs = Object.keys(index.outputs).sort();
  if (JSON.stringify(indexedOutputs) !== JSON.stringify(expectedOutputs)) {
    throw new Error(`${repoDisplayPath(indexPath)} output set is stale; run make phase-schedules`);
  }
  if (index.input_digest !== inputInfo.input_digest) {
    const changes = changedInputs(inputInfo.inputs, index.inputs).slice(0, 5);
    const suffix = changes.length > 0 ? ` (${changes.join(", ")})` : "";
    throw new Error(`phase schedule inputs are stale${suffix}; run make phase-schedules`);
  }
  const staleOutputs = [];
  for (const [file, expectedHash] of Object.entries(index.outputs)) {
    const outputPath = resolveRepoPath(file);
    if (!existsSync(outputPath) || hashFile(outputPath) !== expectedHash) {
      staleOutputs.push(file);
    }
  }
  if (staleOutputs.length > 0) {
    throw new Error(`${staleOutputs.sort().join(", ")} stale; run make phase-schedules`);
  }
}

function renderArtifacts(options) {
  const topology = loadExecutionTopology({ manifestPath: options.topology });
  const taskSurfaceManifest = renderTaskSurfaceManifest(topology);
  validateAllPhaseSlicePlans({ root: repoRoot, taskSurfaceManifest });
  const browserBatchManifest = renderBrowserBatchManifest(topology);
  const serviceBackedScheduleManifest = renderServiceBackedScheduleManifest({
    topology: options.topology,
    topologyObject: topology,
  });
  const schedulerManifest = renderSchedulerManifest({
    topology,
    serviceBackedScheduleManifest,
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
  const densityErrors = collectTaskSurfaceMakeDensityErrors(taskSurfaceManifest);
  if (densityErrors.length > 0) {
    throw new Error(
      `rendered task surface exceeds generated Make budgets:\n${densityErrors.map((error) => `  - ${error}`).join("\n")}`,
    );
  }
  return [
    outputEntry(topology.generatedOutputs.task_surface_manifest, serializeJSON(taskSurfaceManifest)),
    outputEntry(
      topology.generatedOutputs.scheduler_manifest,
      serializeJSON(schedulerManifest),
    ),
    outputEntry(
      topology.generatedOutputs.browser_e2e_batch_manifest,
      serializeJSON(browserBatchManifest),
    ),
    outputEntry(topology.generatedOutputs.task_surface_make, renderTaskSurfaceMake(taskSurfaceManifest)),
    outputEntry(
      topology.generatedOutputs.task_surface_runtime_make,
      renderTaskSurfaceMakeRuntime(taskSurfaceManifest),
    ),
  ];
}

function renderSchedulerManifest({ topology, serviceBackedScheduleManifest }) {
  const checkManifest = renderCheckScheduleManifest(topology, {
    serviceBackedScheduleManifest,
    expandServiceBackedScheduleForCheck,
  });
  const serviceSchedules = serviceBackedScheduleManifest.schedules.map((schedule) => ({
    target: schedule.target,
    scheduler_kind: schedule.scheduler_kind,
    capacity_profile: schedule.capacity_profile,
    resource_limits: schedule.resource_limits,
    stop_on_first_failure: false,
    progress_tick_seconds: 30,
    validate_timing: true,
    summary_groups: [],
    work_units: expandServiceBackedSchedule({ repoRoot, serviceSchedule: schedule }),
    finalizers: [],
  }));
  const checkSchedules = checkManifest.schedules.map((schedule) => ({
    ...schedule,
    stop_on_first_failure: true,
    progress_tick_seconds: 30,
    validate_timing: true,
    finalizers: [],
  }));
  return {
    schema_id: "cartulary.scheduler_manifest.v2",
    generated: {
      generator: "tools/harness/generated-artifacts/render-execution-topology-artifacts.mjs",
      topology: repoDisplayPath(topology.manifestPath),
      source_authoring: {
        check_schedules: "tools/execution_topology_manifest.json",
        service_backed_schedules: "tools/execution_topology_manifest.json",
      },
    },
    schedules: [...checkSchedules, ...serviceSchedules],
  };
}

function cachePath(inputDigest) {
  return path.join(cacheDir, `${inputDigest.replace(/^sha256:/, "")}.json`);
}

function readCachedArtifacts(inputDigest) {
  if (process.env.CARTULARY_EXECUTION_TOPOLOGY_RENDER_DISABLE_CACHE === "1") {
    return null;
  }
  const file = cachePath(inputDigest);
  if (!existsSync(file)) {
    return null;
  }
  const cache = readJSON(file);
  if (
    cache?.schema_id !== renderCacheSchemaID ||
    cache.generator !== "tools/harness/generated-artifacts/render-execution-topology-artifacts.mjs" ||
    cache.generator_version !== generatorVersion ||
    cache.node_version !== process.version ||
    cache.input_digest !== inputDigest ||
    !Array.isArray(cache.artifacts)
  ) {
    return null;
  }
  const artifacts = [];
  for (const artifact of cache.artifacts) {
    if (
      !artifact ||
      typeof artifact.file !== "string" ||
      typeof artifact.content !== "string" ||
      hashContent(artifact.content) !== artifact.hash
    ) {
      return null;
    }
    artifacts.push(outputEntry(artifact.file, artifact.content));
  }
  return artifacts;
}

function writeCache(inputDigest, artifacts) {
  if (process.env.CARTULARY_EXECUTION_TOPOLOGY_RENDER_DISABLE_CACHE === "1") {
    return;
  }
  mkdirSync(cacheDir, { recursive: true });
  writeFileSync(
    cachePath(inputDigest),
    serializeJSON({
      schema_id: renderCacheSchemaID,
      generator: "tools/harness/generated-artifacts/render-execution-topology-artifacts.mjs",
      generator_version: generatorVersion,
      node_version: process.version,
      input_digest: inputDigest,
      artifacts: artifacts.map((artifact) => ({
        file: artifact.file,
        hash: hashContent(artifact.content),
        content: artifact.content,
      })),
      written_at: new Date().toISOString(),
    }),
  );
}

function compareArtifacts(artifacts) {
  const stale = [];
  for (const artifact of artifacts) {
    const existing = readFileSync(artifact.outputPath, "utf8");
    if (existing !== artifact.content) {
      stale.push(artifact.file);
    }
  }
  if (stale.length > 0) {
    throw new Error(`${stale.join(", ")} stale; run make phase-schedules`);
  }
}

function writeChangedArtifacts(artifacts) {
  const changed = [];
  for (const artifact of artifacts) {
    const existing = existsSync(artifact.outputPath)
      ? readFileSync(artifact.outputPath, "utf8")
      : null;
    if (existing === artifact.content) {
      continue;
    }
    mkdirSync(path.dirname(artifact.outputPath), { recursive: true });
    writeFileSync(artifact.outputPath, artifact.content);
    changed.push(artifact.file);
  }
  return changed;
}

function printRenderSummary(changed) {
  if (changed.length === 0) {
    console.log("phase-schedules: unchanged");
    return;
  }
  const displayed = changed.slice(0, 5);
  const suffix = changed.length > displayed.length ? `, ... +${changed.length - displayed.length} more` : "";
  console.log(`phase-schedules: updated ${changed.length} files (${displayed.join(", ")}${suffix})`);
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  if (options.check) {
    quickCheckRenderIndex(options);
    return;
  }
  const inputInfo = collectRenderInputs(options);
  let artifacts = null;
  if (options.fullCheck) {
    artifacts = renderArtifacts(options);
    writeCache(inputInfo.input_digest, artifacts);
  } else {
    artifacts = readCachedArtifacts(inputInfo.input_digest);
    if (!artifacts) {
      artifacts = renderArtifacts(options);
      writeCache(inputInfo.input_digest, artifacts);
    }
  }
  const index = buildRenderIndex({ inputInfo, artifacts });
  const allArtifacts = [
    ...artifacts,
    outputEntry(inputInfo.topologyRaw.generated_outputs.execution_topology_render_index, serializeJSON(index)),
  ];
  if (options.fullCheck) {
    compareArtifacts(allArtifacts);
    return;
  }
  printRenderSummary(writeChangedArtifacts(allArtifacts));
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    main();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(`execution topology render failed: ${message}`);
    if (process.env.CARTULARY_DEBUG_RENDER_STACK === "1" && error instanceof Error && error.stack) {
      console.error(error.stack);
    }
    process.exit(1);
  }
}
