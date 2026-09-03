import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

import {
  repoRoot as defaultRepoRoot,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";

export const fallowReachabilityOwnerSchemaID =
  "cartulary.fallow_reachability_owner.v1";
export const defaultFallowReachabilityOwnerPath =
  "tools/fallow/reachability_owner.json";

const fallowConfigSchema =
  "https://raw.githubusercontent.com/fallow-rs/fallow/main/schema.json";
const pathPattern = /^(?!\/)(?!.*(?:^|\/)\.\.(?:\/|$))(?!.*\/\/).+$/u;
const assetReferencePattern =
  /\b(?:href|src)=["'](\/[^"']+)["']/giu;

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function repoPath(root, relativePath) {
  return path.join(root, relativePath);
}

function repoRel(root, file) {
  return path.relative(root, file).replaceAll("\\", "/") || ".";
}

function assertRepoRelativePath(value, label) {
  if (typeof value !== "string" || !pathPattern.test(value)) {
    throw new Error(`${label} must be a repository-relative path`);
  }
}

function assertExistingFile(root, relativePath, label) {
  assertRepoRelativePath(relativePath, label);
  const file = repoPath(root, relativePath);
  if (!existsSync(file)) {
    throw new Error(`${label} references missing file ${relativePath}`);
  }
  return file;
}

function uniqueSorted(values) {
  return [...new Set(values)].sort((left, right) => left.localeCompare(right));
}

function mergeIgnoreExportRules(rules) {
  const byFile = new Map();
  for (const rule of rules ?? []) {
    if (!rule?.file) {
      continue;
    }
    const current = byFile.get(rule.file) ?? new Set();
    for (const exportName of rule.exports ?? []) {
      current.add(exportName);
    }
    byFile.set(rule.file, current);
  }
  return Array.from(byFile.entries())
    .map(([file, exports]) => ({
      file,
      exports: exports.has("*") ? ["*"] : uniqueSorted(exports),
    }))
    .sort((left, right) => left.file.localeCompare(right.file));
}

function extensionOf(file) {
  return path.extname(file).toLowerCase();
}

function isScriptPath(file, extensions) {
  return extensions.has(extensionOf(file));
}

function collectStringValues(value, result = []) {
  if (typeof value === "string") {
    result.push(value);
    return result;
  }
  if (Array.isArray(value)) {
    for (const item of value) {
      collectStringValues(item, result);
    }
    return result;
  }
  if (value && typeof value === "object") {
    for (const item of Object.values(value)) {
      collectStringValues(item, result);
    }
  }
  return result;
}

function normalizeScriptPath(value) {
  const normalized = String(value ?? "").replace(/^\.\//u, "");
  return pathPattern.test(normalized) ? normalized : null;
}

function collectBackingScripts(taskSurfaceOwner, extensions) {
  const entries = [
    ...(taskSurfaceOwner.targets ?? []),
    ...(taskSurfaceOwner.harness_checks ?? []),
  ];
  return entries.flatMap((entry) =>
    (entry.backing_scripts ?? [])
      .map(normalizeScriptPath)
      .filter((script) => script && isScriptPath(script, extensions)),
  );
}

function collectCommandScripts(taskSurfaceOwner, extensions) {
  return collectStringValues(taskSurfaceOwner)
    .map(normalizeScriptPath)
    .filter((script) => script && isScriptPath(script, extensions));
}

function validateNodeToolBackingScripts(taskSurfaceOwner) {
  const targets = new Map(
    (taskSurfaceOwner.targets ?? []).map((entry) => [entry.name, entry]),
  );
  for (const [name, recipe] of Object.entries(taskSurfaceOwner.make_recipes ?? {})) {
    if (recipe?.type !== "node_tool") {
      continue;
    }
    const target = targets.get(name);
    if (!Array.isArray(target?.backing_scripts) || target.backing_scripts.length === 0) {
      throw new Error(`${name} node_tool recipe must declare target backing_scripts`);
    }
  }
}

function collectTaskSurfaceScripts(root, reachabilityOwner) {
  const ownerPath = reachabilityOwner.task_surface.owner_path;
  const ownerFile = assertExistingFile(root, ownerPath, "task_surface.owner_path");
  const taskSurfaceOwner = readJSON(ownerFile);
  const extensions = new Set(reachabilityOwner.task_surface.script_extensions);
  if (reachabilityOwner.task_surface.required_node_tool_backing_scripts) {
    validateNodeToolBackingScripts(taskSurfaceOwner);
  }
  const scripts = uniqueSorted([
    ...collectBackingScripts(taskSurfaceOwner, extensions),
    ...collectCommandScripts(taskSurfaceOwner, extensions),
  ]);
  for (const script of scripts) {
    assertExistingFile(root, script, "task-surface script reachability");
  }
  return scripts;
}

function collectHarnessEntrypoints(root, reachabilityOwner) {
  const files = uniqueSorted(reachabilityOwner.harness_entrypoints.files ?? []);
  for (const file of files) {
    assertExistingFile(root, file, "harness_entrypoints.files");
  }
  return files;
}

function collectHarnessDynamicExports(root, reachabilityOwner) {
  const entries = reachabilityOwner.harness_dynamic_exports ?? [];
  for (const entry of entries) {
    assertExistingFile(root, entry.file, "harness_dynamic_exports.file");
  }
  return mergeIgnoreExportRules(entries);
}

function validateVitest(root, vitest) {
  assertExistingFile(root, vitest.config_file, "vitest.config_file");
  for (const setupFile of vitest.setup_files) {
    assertExistingFile(root, setupFile, "vitest.setup_files");
  }
}

function publicAssetPath(publicRoot, url) {
  return `${publicRoot}/${url.replace(/^\/+/u, "")}`;
}

function validateVitePublicAssets(root, vitePublicAssets) {
  assertExistingFile(root, vitePublicAssets.public_root, "vite_public_assets.public_root");
  for (const htmlFile of vitePublicAssets.html_entry_files) {
    const resolved = assertExistingFile(root, htmlFile, "vite_public_assets.html_entry_files");
    const source = readFileSync(resolved, "utf8");
    for (const match of source.matchAll(assetReferencePattern)) {
      const url = match[1];
      if (!vitePublicAssets.url_prefixes.some((prefix) => url.startsWith(prefix))) {
        continue;
      }
      const assetPath = publicAssetPath(vitePublicAssets.public_root, url);
      assertExistingFile(root, assetPath, `${htmlFile} public asset ${url}`);
    }
  }
  for (const file of vitePublicAssets.always_used_files) {
    assertExistingFile(root, file, "vite_public_assets.always_used_files");
  }
}

function packageJSONHasDependency(root, packageName) {
  const packageJSON = readJSON(repoPath(root, "package.json"));
  return [
    packageJSON.dependencies,
    packageJSON.devDependencies,
    packageJSON.optionalDependencies,
  ].some((dependencies) => Object.hasOwn(dependencies ?? {}, packageName));
}

function validateExecutableToolingDependencies(root, executableToolingDependencies) {
  for (const dependency of executableToolingDependencies) {
    const ownerScript = assertExistingFile(
      root,
      dependency.owner_script,
      `${dependency.package_name} owner_script`,
    );
    if (!packageJSONHasDependency(root, dependency.package_name)) {
      throw new Error(`${dependency.package_name} must be declared in the root package.json`);
    }
    const source = readFileSync(ownerScript, "utf8");
    for (const token of dependency.command.slice(1)) {
      if (!source.includes(token)) {
        throw new Error(
          `${dependency.owner_script} must reference executable command token ${token}`,
        );
      }
    }
  }
}

function validateBlockingPackageSurfaces(root, policy) {
  const packageNames = new Set();
  const entrypoints = new Set();
  for (const packageSurface of policy.packages) {
    assertExistingFile(
      root,
      packageSurface.entrypoint,
      "blocking_package_surfaces.packages.entrypoint",
    );
    if (packageNames.has(packageSurface.package_name)) {
      throw new Error(
        `blocking_package_surfaces contains duplicate package ${packageSurface.package_name}`,
      );
    }
    if (entrypoints.has(packageSurface.entrypoint)) {
      throw new Error(
        `blocking_package_surfaces contains duplicate entrypoint ${packageSurface.entrypoint}`,
      );
    }
    packageNames.add(packageSurface.package_name);
    entrypoints.add(packageSurface.entrypoint);
  }
}

export function loadFallowReachabilityOwner({
  root = defaultRepoRoot,
  ownerPath = defaultFallowReachabilityOwnerPath,
} = {}) {
  const owner = readJSON(assertExistingFile(root, ownerPath, "Fallow reachability owner"));
  validateSchemaSync(fallowReachabilityOwnerSchemaID, owner);
  assertExistingFile(root, owner.base_config.path, "base_config.path");
  validateVitest(root, owner.vitest);
  collectHarnessEntrypoints(root, owner);
  collectHarnessDynamicExports(root, owner);
  validateVitePublicAssets(root, owner.vite_public_assets);
  validateExecutableToolingDependencies(root, owner.executable_tooling_dependencies);
  validateBlockingPackageSurfaces(root, owner.blocking_package_surfaces);
  return owner;
}

export function buildBlockingPackageSurfaceConfig({ config, owner }) {
  const policy = owner.blocking_package_surfaces;
  const entrypoints = new Set(policy.packages.map((entry) => entry.entrypoint));
  const packageNames = new Set(
    policy.packages.map((entry) => entry.package_name),
  );
  return {
    ...config,
    entry: (config.entry ?? []).filter((entry) => !entrypoints.has(entry)),
    ignorePatterns: uniqueSorted([
      ...(config.ignorePatterns ?? []),
      ...policy.packages.flatMap((entry) => entry.own_test_globs),
    ]),
    ignoreExports: (config.ignoreExports ?? []).filter(
      (entry) => !entrypoints.has(entry.file),
    ),
    publicPackages: (config.publicPackages ?? []).filter(
      (packageName) => !packageNames.has(packageName),
    ),
  };
}

export function buildResolvedFallowConfig({
  root = defaultRepoRoot,
  ownerPath = defaultFallowReachabilityOwnerPath,
  outputFile = null,
} = {}) {
  const owner = loadFallowReachabilityOwner({ root, ownerPath });
  const baseConfig = readJSON(repoPath(root, owner.base_config.path));
  const taskSurfaceScripts = collectTaskSurfaceScripts(root, owner);
  const harnessEntrypoints = collectHarnessEntrypoints(root, owner);
  const harnessDynamicExports = collectHarnessDynamicExports(root, owner);
  const executableToolingDependencies = uniqueSorted(
    owner.executable_tooling_dependencies.map((entry) => entry.package_name),
  );
  const config = {
    ...baseConfig,
    $schema: baseConfig.$schema ?? fallowConfigSchema,
    entry: uniqueSorted([
      ...(baseConfig.entry ?? []),
      ...taskSurfaceScripts,
      ...harnessEntrypoints,
    ]),
    ignoreExports: mergeIgnoreExportRules([
      ...(baseConfig.ignoreExports ?? []),
      ...harnessDynamicExports,
    ]),
    ignoreUnresolvedImports: uniqueSorted([
      ...(baseConfig.ignoreUnresolvedImports ?? []),
      ...owner.vite_public_assets.url_prefixes.map((prefix) => `${prefix}**`),
    ]),
    framework: [
      ...(baseConfig.framework ?? []),
      {
        name: "cartulary-vitest-setup",
        detection: { type: "fileExists", pattern: owner.vitest.config_file },
        entryPointRole: "test",
        entryPoints: owner.vitest.test_entry_globs,
        alwaysUsed: owner.vitest.setup_files,
      },
      {
        name: "cartulary-vite-public-assets",
        detection: {
          type: "fileExists",
          pattern: owner.vite_public_assets.html_entry_files[0],
        },
        alwaysUsed: owner.vite_public_assets.always_used_files,
      },
      {
        name: "cartulary-executable-tooling-dependencies",
        detection: {
          type: "fileExists",
          pattern: owner.executable_tooling_dependencies[0]?.owner_script ??
            owner.base_config.path,
        },
        toolingDependencies: executableToolingDependencies,
      },
    ],
  };
  const result = {
    config,
    owner,
    stats: {
      task_surface_entry_points: taskSurfaceScripts.length,
      harness_entry_points: harnessEntrypoints.length,
      harness_dynamic_export_files: harnessDynamicExports.length,
      harness_dynamic_exports: harnessDynamicExports.reduce(
        (count, rule) => count + rule.exports.length,
        0,
      ),
      vitest_setup_files: owner.vitest.setup_files.length,
      vite_public_assets: owner.vite_public_assets.always_used_files.length,
      executable_tooling_dependencies: executableToolingDependencies.length,
    },
  };
  if (outputFile) {
    secureWriteFile(outputFile, `${JSON.stringify(config, null, 2)}\n`);
    result.outputFile = outputFile;
    result.outputPath = repoRel(root, outputFile);
  }
  return result;
}
