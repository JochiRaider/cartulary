#!/usr/bin/env node
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptPath = fileURLToPath(import.meta.url);
const scriptDir = path.dirname(scriptPath);
const repoRoot = path.resolve(scriptDir, "../../..");
const schemaID = "cartulary.test_support_inventory.v1";

function usage() {
  throw new Error(
    [
      "usage: support-inventory-profiles.mjs <profile>",
      "profiles:",
      "  gosec-runtime-flags --base <flags>",
      "  gosec-support-patterns",
      "  backend-runtime-excludes",
      "options:",
      "  --inventory <path>",
      "  --root <path>",
    ].join("\n"),
  );
}

function parseArgs(argv) {
  const options = {
    command: argv[0] ?? "",
    base: "",
    inventory:
      process.env.TEST_SUPPORT_INVENTORY ??
      process.env.CARTULARY_TEST_SUPPORT_INVENTORY ??
      "tools/test_support_inventory.json",
    root: process.env.CARTULARY_REPO_ROOT ?? repoRoot,
  };
  for (let index = 1; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--base") {
      options.base = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--inventory") {
      options.inventory = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--root") {
      options.root = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    usage();
  }
  if (
    ![
      "gosec-runtime-flags",
      "gosec-support-patterns",
      "backend-runtime-excludes",
    ].includes(options.command)
  ) {
    usage();
  }
  options.root = path.resolve(options.root);
  options.inventory = resolvePath(options.root, options.inventory);
  return options;
}

function resolvePath(root, value) {
  if (!value) {
    throw new Error("inventory path must be non-empty");
  }
  return path.isAbsolute(value) ? value : path.join(root, value);
}

function normalizePath(value) {
  return value.split(path.sep).join("/");
}

function requireString(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  const normalized = normalizePath(value.trim());
  if (
    normalized.startsWith("/") ||
    normalized === "." ||
    normalized.startsWith("../") ||
    normalized.includes("/../") ||
    normalized.includes("//")
  ) {
    throw new Error(`${label} must be a normalized repository-relative path`);
  }
  return normalized;
}

function requireEnum(value, label, allowed) {
  if (!allowed.has(value)) {
    throw new Error(`${label} must be one of ${[...allowed].join(", ")}`);
  }
  return value;
}

export function readSupportInventory(root, inventoryPath) {
  const absolutePath = resolvePath(root, inventoryPath);
  if (!existsSync(absolutePath)) {
    throw new Error(`test support inventory not found at ${absolutePath}`);
  }
  const raw = readFileSync(absolutePath, "utf8");
  const parsed = JSON.parse(raw);
  if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
    throw new Error(`${absolutePath} must be an object`);
  }
  if (parsed.schema_id !== schemaID) {
    throw new Error(`${absolutePath} must declare schema_id ${schemaID}`);
  }
  if (!Array.isArray(parsed.go_support_roots) || parsed.go_support_roots.length === 0) {
    throw new Error(`${absolutePath}.go_support_roots must be a non-empty array`);
  }
  const seen = new Set();
  const roots = parsed.go_support_roots.map((entry, index) => {
    const label = `${absolutePath}.go_support_roots[${index + 1}]`;
    const supportRoot = {
      path: requireString(entry?.path, `${label}.path`),
      runtimeScan: requireEnum(
        entry?.runtime_scan,
        `${label}.runtime_scan`,
        new Set(["included", "excluded"]),
      ),
      supportScan: requireEnum(
        entry?.support_scan,
        `${label}.support_scan`,
        new Set(["included", "not_applicable"]),
      ),
      serviceStarting: entry?.service_starting === true,
    };
    if (seen.has(supportRoot.path)) {
      throw new Error(`${absolutePath}.go_support_roots contains duplicate path ${supportRoot.path}`);
    }
    seen.add(supportRoot.path);
    return supportRoot;
  });
  roots.sort((left, right) => left.path.localeCompare(right.path));
  return { raw, path: absolutePath, roots };
}

export function runtimeExcludeFlags(inventory) {
  return inventory.roots
    .filter((root) => root.runtimeScan === "excluded")
    .map((root) => `-exclude-dir=${root.path}`);
}

export function supportPackagePatterns(inventory) {
  return inventory.roots
    .filter((root) => root.supportScan === "included")
    .map((root) => `./${root.path}/...`);
}

export function backendRuntimeExcludePatterns(inventory) {
  return inventory.roots
    .filter((root) => root.runtimeScan === "excluded")
    .map((root) => `${root.path}/**`);
}

export function renderGosecRuntimeFlags(inventory, baseFlags) {
  return [baseFlags.trim(), ...runtimeExcludeFlags(inventory)].filter(Boolean).join(" ");
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const inventory = readSupportInventory(options.root, options.inventory);
  switch (options.command) {
    case "gosec-runtime-flags":
      process.stdout.write(renderGosecRuntimeFlags(inventory, options.base));
      break;
    case "gosec-support-patterns":
      process.stdout.write(supportPackagePatterns(inventory).join(" "));
      break;
    case "backend-runtime-excludes":
      process.stdout.write(backendRuntimeExcludePatterns(inventory).join("\n"));
      break;
    default:
      usage();
  }
}

if (process.argv[1] === scriptPath) {
  try {
    main();
  } catch (error) {
    console.error(`support inventory profile rendering failed: ${error instanceof Error ? error.message : String(error)}`);
    process.exit(1);
  }
}
