import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import path from "node:path";

import { semanticJSONDigest } from "../../contract/index.mjs";
import { loadTestCatalog } from "../../test-catalog/index.mjs";

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function sha256(bytes) {
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}

function normalizedRoot(value) {
  return value.replaceAll("\\", "/").replace(/\/$/u, "");
}

function under(relative, root) {
  return relative === root || relative.startsWith(`${root}/`);
}

function entriesForRoots(entries, roots) {
  const normalized = [...new Set(roots.map(normalizedRoot))].sort(compareASCII);
  return entries.filter((entry) => normalized.some((root) => under(entry.path, root)));
}

function digestClosure(strategy, entries, metadata = {}) {
  const sorted = [...new Map(entries.map((entry) => [entry.path, entry])).values()]
    .sort((left, right) => compareASCII(left.path, right.path));
  return {
    strategy,
    metadata,
    entries: sorted,
    digest: semanticJSONDigest({ strategy, metadata, entries: sorted }),
  };
}

function broadClosure(entries, profile, reason = "registered_broad") {
  return digestClosure(
    reason === "registered_broad" ? "broad" : "broad_fallback",
    entriesForRoots(entries, profile.input_roots),
    { reason, registered_roots: [...profile.input_roots].sort(compareASCII) },
  );
}

function cached(cache, key, build) {
  if (!cache) return build();
  if (!cache.has(key)) cache.set(key, build());
  return cache.get(key);
}

function closureCacheKey(kind, value) {
  return `closure:${kind}:${semanticJSONDigest(value)}`;
}

function checkedSourceText(root, entry) {
  if (!entry || entry.kind !== "file") throw new Error("dependency source is not a regular file");
  const bytes = readFileSync(path.join(root, entry.path));
  if (sha256(bytes) !== entry.byte_digest) {
    throw new Error("dependency source changed after source snapshot capture");
  }
  return bytes.toString("utf8");
}

function commonDependencyPaths(catalog, rows, entries) {
  const selectedOwners = new Set(
    rows.flatMap((row) => [row.owner_id, ...(row.collaborator_ids ?? [])]),
  );
  const ownerManifestByID = new Map(
    catalog.registry.owners.map((owner) => [owner.owner_id, owner.manifest_path]),
  );
  const exact = new Set([
    "Makefile",
    "go.mod",
    "go.sum",
    "package.json",
    "pnpm-lock.yaml",
    "pnpm-workspace.yaml",
    "tsconfig.base.json",
    "tsconfig.json",
    "tools/execution_topology_manifest.json",
    "tools/generated_artifact_policy.json",
    "tools/harness_cache_registry.json",
    "tools/harness_work_graph_owner.json",
    "tools/scheduler_resource_registry.json",
    "tools/task_surface.generated.mk",
    "tools/task_surface.runtime.generated.mk",
    "tools/task_surface_manifest.json",
    "tools/task_surface_owner.json",
    "tools/test_catalog_owner.json",
    "tools/toolchain_pins.json",
  ]);
  for (const ownerID of selectedOwners) {
    const manifest = ownerManifestByID.get(ownerID);
    if (manifest) exact.add(manifest);
    const contractRoot = `contracts/${ownerID.replace(/^[^.]+\./u, "").replaceAll("_", "-")}`;
    if (entries.some((entry) => under(entry.path, contractRoot))) exact.add(contractRoot);
  }
  const prefixes = [
    "configs/",
    "scripts/",
    "tools/harness",
    "tools/schemas",
    "tools/test_families/harness.",
  ];
  return entries.filter((entry) =>
    exact.has(entry.path) ||
    prefixes.some((prefix) => entry.path.startsWith(prefix)) ||
    /^tools\/(?:generate|render)/u.test(entry.path),
  );
}

function selectedOwnerIDs(rows) {
  return [...new Set(rows.flatMap((row) => [row.owner_id, ...(row.collaborator_ids ?? [])]))]
    .sort(compareASCII);
}

function moduleIdentity(root, entryByPath) {
  const goMod = entryByPath.get("go.mod");
  const match = checkedSourceText(root, goMod).match(/^module\s+(\S+)\s*$/mu);
  if (!match) throw new Error("go.mod has no module identity");
  return match[1];
}

function goImports(source) {
  const imports = [];
  const blocks = source.matchAll(/\bimport\s*(?:\(([^)]*)\)|"([^"]+)")/gu);
  for (const match of blocks) {
    if (match[2]) imports.push(match[2]);
    if (match[1]) {
      for (const value of match[1].matchAll(/"([^"]+)"/gu)) imports.push(value[1]);
    }
  }
  return [...new Set(imports)].sort(compareASCII);
}

function goPackageClosure(root, entries, rows, catalog, resolverCache) {
  const initialPackages = [...new Set(rows.map((row) => normalizedRoot(row.selector.package).replace(/^\.\//u, "")))]
    .sort(compareASCII);
  const cacheKey = closureCacheKey("go", {
    owners: selectedOwnerIDs(rows),
    packages: initialPackages,
  });
  if (resolverCache?.has(cacheKey)) return resolverCache.get(cacheKey);
  if (initialPackages.some((packagePath) => packagePath === "" || packagePath.startsWith("../"))) {
    throw new Error("selected Go package is unsafe");
  }
  const selectedDirectories = cached(
    resolverCache,
    closureCacheKey("go-package-set", { packages: initialPackages }),
    () => {
      const entryByPath = new Map(entries.map((entry) => [entry.path, entry]));
      const modulePath = moduleIdentity(root, entryByPath);
      const pending = [...initialPackages];
      const selected = new Set();
      const parsed = new Set();
      while (pending.length > 0) {
        const packageDirectory = pending.shift();
        if (parsed.has(packageDirectory)) continue;
        const goFiles = entries.filter((entry) =>
          entry.path.startsWith(`${packageDirectory}/`) &&
          !entry.path.slice(packageDirectory.length + 1).includes("/") &&
          entry.path.endsWith(".go"),
        );
        if (goFiles.length === 0 || goFiles.some((entry) => entry.kind !== "file")) {
          throw new Error("selected Go package is absent or unsafe");
        }
        parsed.add(packageDirectory);
        selected.add(packageDirectory);
        for (const file of goFiles) {
          for (const imported of goImports(checkedSourceText(root, file))) {
            if (imported === modulePath) {
              throw new Error("repository module root cannot be resolved as a package");
            }
            if (!imported.startsWith(`${modulePath}/`)) continue;
            const importedDirectory = imported.slice(modulePath.length + 1);
            if (!parsed.has(importedDirectory)) pending.push(importedDirectory);
          }
        }
        pending.sort(compareASCII);
      }
      return [...selected].sort(compareASCII);
    },
  );
  const packageEntries = entries.filter((entry) =>
    selectedDirectories.some((directory) => under(entry.path, directory)),
  );
  const closure = digestClosure(
    "go_packages",
    [...commonDependencyPaths(catalog, rows, entries), ...packageEntries],
    { packages: selectedDirectories },
  );
  resolverCache?.set(cacheKey, closure);
  return closure;
}

function workspaceManifests(root, entries, resolverCache) {
  if (resolverCache?.has("workspace-manifests")) {
    return resolverCache.get("workspace-manifests");
  }
  const manifests = entries
    .filter((entry) => /^(?:apps|packages)\/[^/]+\/package\.json$/u.test(entry.path));
  const workspaces = [];
  for (const entry of manifests) {
    const manifest = JSON.parse(checkedSourceText(root, entry));
    if (typeof manifest.name !== "string" || manifest.name.trim() === "") {
      throw new Error("workspace package has no name");
    }
    workspaces.push({
      name: manifest.name,
      root: path.posix.dirname(entry.path),
      manifest,
    });
  }
  const sorted = workspaces.sort((left, right) => compareASCII(left.root, right.root));
  resolverCache?.set("workspace-manifests", sorted);
  return sorted;
}

function declaredWorkspaceDependencies(workspace, byName) {
  const sections = ["dependencies", "devDependencies", "optionalDependencies", "peerDependencies"];
  return sections
    .flatMap((section) => Object.keys(workspace.manifest[section] ?? {}))
    .filter((name) => byName.has(name));
}

function typescriptSpecifiers(source) {
  const values = [];
  for (const match of source.matchAll(/\b(?:import|export)\s+(?:type\s+)?(?:[^"'()]*?\s+from\s+)?["']([^"']+)["']/gu)) {
    values.push(match[1]);
  }
  for (const match of source.matchAll(/\b(?:import|require)\(\s*["']([^"']+)["']\s*\)/gu)) {
    values.push(match[1]);
  }
  return [...new Set(values)].sort(compareASCII);
}

function referencedWorkspaceDependencies(root, workspace, workspaces, entries, byName) {
  const names = [];
  const sourceEntries = entries.filter((entry) =>
    under(entry.path, workspace.root) &&
    entry.kind === "file" &&
    /\.(?:[cm]?[jt]sx?)$/u.test(entry.path),
  );
  for (const entry of sourceEntries) {
    const source = checkedSourceText(root, entry);
    for (const specifier of typescriptSpecifiers(source)) {
      if (specifier.startsWith(".")) {
        const resolved = path.posix.normalize(path.posix.join(path.posix.dirname(entry.path), specifier));
        if (resolved.startsWith("../")) {
          throw new Error("relative TypeScript dependency escapes the repository");
        }
        const dependency = workspaces.find((candidate) => under(resolved, candidate.root));
        if (!dependency) {
          throw new Error("relative TypeScript dependency has no proved workspace");
        }
        if (dependency.name !== workspace.name) names.push(dependency.name);
        continue;
      }
      const dependency = [...byName.keys()].find((name) =>
        specifier === name || specifier.startsWith(`${name}/`),
      );
      if (dependency && dependency !== workspace.name) names.push(dependency);
    }
  }
  return [...new Set(names)].sort(compareASCII);
}

function typescriptWorkspaceClosure(root, entries, rows, catalog, resolverCache) {
  const workspaces = workspaceManifests(root, entries, resolverCache);
  const byName = new Map(workspaces.map((workspace) => [workspace.name, workspace]));
  const selected = new Map();
  for (const row of rows) {
    const selectedFile = normalizedRoot(row.selector.file);
    const workspace = workspaces.find((candidate) => under(selectedFile, candidate.root));
    if (!workspace || !entries.some((entry) => entry.path === selectedFile)) {
      throw new Error("selected TypeScript file has no proved workspace");
    }
    selected.set(workspace.name, workspace);
  }
  const cacheKey = closureCacheKey("typescript", {
    owners: selectedOwnerIDs(rows),
    workspaces: [...selected.values()].map((workspace) => workspace.root).sort(compareASCII),
  });
  if (resolverCache?.has(cacheKey)) return resolverCache.get(cacheKey);
  const selectedWorkspaceNames = cached(
    resolverCache,
    closureCacheKey("typescript-workspace-set", {
      workspaces: [...selected.values()].map((workspace) => workspace.root).sort(compareASCII),
    }),
    () => {
      const transitive = new Map(selected);
      const pending = [...transitive.values()];
      while (pending.length > 0) {
        const workspace = pending.shift();
        const dependencies = [
          ...declaredWorkspaceDependencies(workspace, byName),
          ...referencedWorkspaceDependencies(root, workspace, workspaces, entries, byName),
        ];
        for (const name of dependencies.sort(compareASCII)) {
          if (transitive.has(name)) continue;
          const dependency = byName.get(name);
          if (!dependency) throw new Error("workspace dependency cannot be resolved");
          transitive.set(name, dependency);
          pending.push(dependency);
        }
      }
      return [...transitive.keys()].sort(compareASCII);
    },
  );
  const selectedWorkspaces = selectedWorkspaceNames.map((name) => byName.get(name));
  const workspaceEntries = entries.filter((entry) =>
    selectedWorkspaces.some((workspace) => under(entry.path, workspace.root)),
  );
  const closure = digestClosure(
    "typescript_workspaces",
    [...commonDependencyPaths(catalog, rows, entries), ...workspaceEntries],
    { workspaces: selectedWorkspaces.map((workspace) => workspace.root).sort(compareASCII) },
  );
  resolverCache?.set(cacheKey, closure);
  return closure;
}

export function resolveCacheDependencyClosure({ root, entries, profile, unit, resolverCache }) {
  if (!entries || profile.dependency_strategy !== "unit_aware") {
    return cached(
      resolverCache,
      closureCacheKey("broad", { profile_id: profile.profile_id, roots: profile.input_roots }),
      () => broadClosure(entries ?? [], profile),
    );
  }
  try {
    const rowIDs = (unit.command.environment.CARTULARY_TEST_ROWS ?? "")
      .split(",")
      .map((value) => value.trim())
      .filter(Boolean);
    if (rowIDs.length === 0 || new Set(rowIDs).size !== rowIDs.length) {
      throw new Error("unit has no exact selected row set");
    }
    const catalog = cached(resolverCache, "test-catalog", () => loadTestCatalog(root));
    const rows = rowIDs.map((rowID) => {
      const row = catalog.rowByID.get(rowID);
      if (!row) throw new Error("selected row is unknown");
      return row;
    });
    const runners = new Set(rows.map((row) => row.runner));
    if (runners.size !== 1) throw new Error("selected rows mix dependency models");
    if (runners.has("go")) {
      return goPackageClosure(root, entries, rows, catalog, resolverCache);
    }
    if (runners.has("vitest")) {
      return typescriptWorkspaceClosure(root, entries, rows, catalog, resolverCache);
    }
    throw new Error("selected runner has no unit-aware dependency model");
  } catch {
    return cached(
      resolverCache,
      closureCacheKey("broad-fallback", {
        profile_id: profile.profile_id,
        roots: profile.input_roots,
      }),
      () => broadClosure(entries, profile, "resolution_incomplete"),
    );
  }
}
