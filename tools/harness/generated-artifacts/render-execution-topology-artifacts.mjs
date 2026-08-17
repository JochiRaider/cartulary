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
  renderTaskSurfaceManifest,
} from "./execution-topology.mjs";
import {
  collectTaskSurfaceManifestErrors,
  renderTaskSurfaceMake,
  renderTaskSurfaceMakeRuntime,
  collectTaskSurfaceMakeDensityErrors,
} from "./task-surface/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");
export const renderIndexSchemaID = "cartulary.execution_topology_render_index.v1";
const generatorVersion = 3;
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

function collectCatalogInputs(inputs, seen, catalogRoot) {
  const ownerRegistryPath = path.join(catalogRoot, "tools", "test_catalog_owner.json");
  addFileInput(inputs, seen, "test_owner_registry", ownerRegistryPath);
  const ownerRegistry = readJSON(ownerRegistryPath);
  for (const owner of ownerRegistry.owners ?? []) {
    addFileInput(
      inputs,
      seen,
      "test_family_manifest",
      path.join(catalogRoot, owner.manifest_path),
    );
  }
  addFileInput(
    inputs,
    seen,
    "test_runner_registry",
    path.join(catalogRoot, "tools", "test_runner_registry.json"),
  );
  const verificationRegistryPath = path.join(
    catalogRoot,
    "contracts",
    "verification",
    "registry.json",
  );
  addFileInput(inputs, seen, "verification_registry", verificationRegistryPath);
  const verificationRegistry = readJSON(verificationRegistryPath);
  for (const owner of verificationRegistry.owners ?? []) {
    addFileInput(
      inputs,
      seen,
      "verification_owner_contract",
      path.join(catalogRoot, owner.contract_path),
    );
  }
  const fixtureRegistryPath = path.join(
    catalogRoot,
    "tools",
    "performance_fixture_snapshot_owner.json",
  );
  addFileInput(inputs, seen, "performance_fixture_snapshot_owner", fixtureRegistryPath);
  addFileInput(
    inputs,
    seen,
    "performance_fixture_builder_policy",
    path.join(catalogRoot, "tools", "performance_fixture_builder_policy.json"),
  );
  addFileInput(
    inputs,
    seen,
    "postgres_fixture_policy_registry",
    path.join(catalogRoot, "tools", "postgres_fixture_policy_registry.json"),
  );
  const fixtureRegistry = readJSON(fixtureRegistryPath);
  for (const profile of fixtureRegistry.profiles ?? []) {
    for (const ref of profile.source_contract_refs ?? []) {
      addFileInput(
        inputs,
        seen,
        "performance_fixture_source_contract",
        path.join(catalogRoot, ref.path),
      );
    }
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
    "tools/harness/contract",
    "tools/harness/test-catalog",
    "scripts/lib",
  ];
  const isProductionModule = (candidate) =>
    candidate.endsWith(".mjs") &&
    !candidate.split(path.sep).includes("tests");
  for (const root of rendererSourceRoots) {
    for (const file of collectFiles(path.join(repoRoot, root), isProductionModule)) {
      addFileInput(inputs, seen, "renderer_source", file);
    }
  }
  addRequiredRepoInput(inputs, seen, "renderer_source", "tools/harness/runtime-binary-registry.mjs");
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
  collectCatalogInputs(inputs, seen, path.resolve(options.catalogRoot ?? repoRoot));
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
    throw new Error(`${repoDisplayPath(indexPath)} generator_version is stale; run make generate`);
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
    throw new Error(`${repoDisplayPath(indexPath)} missing; run make generate`);
  }
  const index = readJSON(indexPath);
  validateRenderIndex(index, indexPath);
  const expectedOutputs = expectedRenderedOutputFiles(inputInfo.topologyRaw).sort();
  const indexedOutputs = Object.keys(index.outputs).sort();
  if (JSON.stringify(indexedOutputs) !== JSON.stringify(expectedOutputs)) {
    throw new Error(`${repoDisplayPath(indexPath)} output set is stale; run make generate`);
  }
  if (index.input_digest !== inputInfo.input_digest) {
    const changes = changedInputs(inputInfo.inputs, index.inputs).slice(0, 5);
    const suffix = changes.length > 0 ? ` (${changes.join(", ")})` : "";
    throw new Error(`generated topology inputs are stale${suffix}; run make generate`);
  }
  const staleOutputs = [];
  for (const [file, expectedHash] of Object.entries(index.outputs)) {
    const outputPath = resolveRepoPath(file);
    if (!existsSync(outputPath) || hashFile(outputPath) !== expectedHash) {
      staleOutputs.push(file);
    }
  }
  if (staleOutputs.length > 0) {
    throw new Error(`${staleOutputs.sort().join(", ")} stale; run make generate`);
  }
}

function renderArtifacts(options) {
  const topology = loadExecutionTopology({ manifestPath: options.topology });
  const taskSurfaceManifest = renderTaskSurfaceManifest(topology);
  const browserBatchManifest = renderBrowserBatchManifest(topology);
  const schedulerManifest = renderSchedulerManifest({ topology });
  const taskSurfaceErrors = collectTaskSurfaceManifestErrors(taskSurfaceManifest, {
    browserBatchManifest,
    serviceBackedScheduleManifest: { schedules: [] },
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

function renderSchedulerManifest({ topology }) {
  return {
    schema_id: "cartulary.scheduler_manifest.v3",
    generated: {
      generator: "tools/harness/generated-artifacts/render-execution-topology-artifacts.mjs",
      topology: repoDisplayPath(topology.manifestPath),
      source_authoring: { work_graph: "tools/harness_work_graph_owner.json" },
    },
    schedules: [],
  };
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
    throw new Error(`${stale.join(", ")} stale; run make generate`);
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
    console.log("generated-topology: unchanged");
    return;
  }
  const displayed = changed.slice(0, 5);
  const suffix = changed.length > displayed.length ? `, ... +${changed.length - displayed.length} more` : "";
  console.log(`generated-topology: updated ${changed.length} files (${displayed.join(", ")}${suffix})`);
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  if (options.check) {
    quickCheckRenderIndex(options);
    return;
  }
  const inputInfo = collectRenderInputs(options);
  const artifacts = renderArtifacts(options);
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
