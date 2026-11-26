import { readFileSync } from "node:fs";
import path from "node:path";

// The authored producer registry (P5-3): each producer publishes its output at
// a binding's path. The store, through `go run ./tools/harnessbuild produce`,
// decides whether to build; Make and the operation runner only name producers.
export const producerRegistryPath = "tools/harness_producers.json";
export const producerRegistrySchemaID = "cartulary.harness_producers.v1";
export const producerCommand = ["run", "./tools/harnessbuild", "produce"];

const bindingTemplate = /^\$\(([A-Z][A-Z0-9_]*)\)$/u;

function readJSON(root, file) {
  return JSON.parse(readFileSync(path.join(root, file), "utf8"));
}

export function readProducerRegistry(root) {
  const registry = readJSON(root, producerRegistryPath);
  if (registry.schema_id !== producerRegistrySchemaID || !Array.isArray(registry.producers)) {
    throw new Error(`${producerRegistryPath} must declare ${producerRegistrySchemaID} producers`);
  }
  return registry;
}

/** Maps each producer binding to its producer ID; a web bundle has none. */
export function producerBindings(registry) {
  return new Map(registry.producers.filter((producer) => producer.binding).map((producer) => [producer.binding, producer.id]));
}

/**
 * Splits prerequisites into the producers that publish them and the files
 * Make still produces. A prerequisite is produced when it is exactly a
 * registered binding, such as `$(SQLC_BIN)`.
 */
export function splitProducedPrerequisites(prerequisites, bindings) {
  const producers = [];
  const files = [];
  for (const prerequisite of prerequisites) {
    const id = bindings.get(bindingTemplate.exec(prerequisite)?.[1]);
    if (id === undefined) files.push(prerequisite);
    else if (!producers.includes(id)) producers.push(id);
  }
  return { producers, files };
}

/**
 * The repository-relative path each published producer makes available at
 * when its binding is unset: the file under its layout directory. A Go
 * tool's file is its binary and pinned version, ShellCheck's carries its
 * pinned version, embedded assets are named by their archive, the Node
 * runtime by its executable, the other installations and the report renderer
 * by their receipts, and the release inventory by its SBOM.
 * The store's Workspace.DefaultName applies the same rule.
 */
export function producerDefaultPaths(root, registry = readProducerRegistry(root)) {
  const pins = readJSON(root, "tools/toolchain_pins.json");
  const layout = new Map(readJSON(root, "tools/workspace_layout.json").directories.map((entry) => [entry.id, entry.path]));
  const defaults = new Map();
  for (const producer of registry.producers.filter((entry) => entry.binding)) {
    const directory = layout.get(producer.directory);
    if (directory === undefined) throw new Error(`${producer.id}: the workspace layout declares no ${producer.directory} directory`);
    let name;
    if (producer.kind === "go_binary") {
      name = producer.name;
    } else if (producer.kind === "go_tool") {
      const module = pins.tools?.[producer.tool];
      if (typeof module !== "string" || !module.includes("@")) throw new Error(`${producer.id}: the toolchain pins name no tool ${producer.tool}`);
      const [modulePath, version] = module.split("@");
      name = `${path.posix.basename(modulePath)}-${version}`;
    } else if (producer.kind === "shellcheck") {
      name = `shellcheck-v${pins.shellcheck_version}`;
    } else if (producer.kind === "embedded_assets") {
      name = "web-assets.zip";
    } else if (producer.kind === "node_runtime") {
      name = "bin/node";
    } else if (producer.kind === "frontend_install") {
      name = "proof/installed.v1.json";
    } else if (["playwright", "test_service_images", "report_renderer"].includes(producer.kind)) {
      name = "installation.json";
    } else if (producer.kind === "release_inventory") {
      name = "sbom.cyclonedx.json";
    } else {
      throw new Error(`${producer.id}: unknown producer kind ${producer.kind}`);
    }
    defaults.set(producer.binding, path.posix.join(directory, name));
  }
  return defaults;
}

/** The producer context Make rendering needs: each binding's producer and default path. */
export function producerContext(root) {
  const registry = readProducerRegistry(root);
  return { bindings: producerBindings(registry), defaults: producerDefaultPaths(root, registry) };
}
