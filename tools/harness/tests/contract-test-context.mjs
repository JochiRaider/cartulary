import { readFileSync } from "node:fs";
import path from "node:path";
import { loadTestCatalog } from "../test-catalog/index.mjs";
import { WorkGraphCompiler, loadCacheRegistry } from "../scheduler/work-graph/index.mjs";
import { loadPerformanceFixtureBuilderPolicy } from "../performance-fixture/index.mjs";

/** Private, case-owned dependencies. Constructing a context performs no I/O. */
export function createContractTestContext(root, overrides = {}) {
  const readJSON = (relative) => JSON.parse(readFileSync(path.join(root, relative), "utf8"));
  const loaders = {
    taskSurface: () => readJSON("tools/task_surface_owner.json"),
    topology: () => readJSON("tools/execution_topology_manifest.json"),
    rowMigrations: () => readJSON("tools/test_catalog_row_migrations.json"),
    catalog: () => loadTestCatalog(root),
    compiler: () => new WorkGraphCompiler(root),
    cacheRegistry: () => loadCacheRegistry(root).registry,
    fixtureBuilderPolicy: () => loadPerformanceFixtureBuilderPolicy(root),
    ...overrides,
  };
  const context = { root, readJSON };
  for (const [name, load] of Object.entries(loaders)) {
    let loaded = false;
    let value;
    Object.defineProperty(context, name, {
      enumerable: true,
      get() {
        if (!loaded) {
          value = load();
          loaded = true;
        }
        return value;
      },
    });
  }
  return context;
}
