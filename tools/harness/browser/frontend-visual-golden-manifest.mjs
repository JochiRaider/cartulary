#!/usr/bin/env node

import { createHash } from "node:crypto";
import {
  existsSync,
  lstatSync,
  readdirSync,
  readFileSync,
  renameSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

export const goldenManifestSchemaID =
  "cartulary.frontend_visual_golden_manifest.v1";
export const goldenManifestPath = "tools/frontend_visual_golden_manifest.json";
export const rendererProfilePath = "tools/frontend_visual_renderer_profile.json";
export const visualSnapshotRoot =
  "apps/web/e2e/workbook.visual.spec.ts-snapshots";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRoot = path.resolve(scriptDir, "../../..");

function normalizePath(value) {
  return value.replaceAll("\\", "/");
}

function sha256File(file) {
  return createHash("sha256").update(readFileSync(file)).digest("hex");
}

function committedPNGs(directory) {
  if (!existsSync(directory)) {
    throw new Error(`visual snapshot root is missing: ${directory}`);
  }
  const stat = lstatSync(directory);
  if (!stat.isDirectory() || stat.isSymbolicLink()) {
    throw new Error(`visual snapshot root must be a real directory: ${directory}`);
  }
  return readdirSync(directory, { withFileTypes: true })
    .filter((entry) => entry.isFile() && entry.name.endsWith(".png"))
    .map((entry) => path.join(directory, entry.name))
    .sort((left, right) => left.localeCompare(right));
}

export function buildFrontendVisualGoldenManifest({
  root = defaultRoot,
  snapshotRoot = path.join(root, visualSnapshotRoot),
} = {}) {
  const renderer = JSON.parse(
    readFileSync(path.join(root, rendererProfilePath), "utf8"),
  );
  if (
    renderer.schema_id !== "cartulary.frontend_visual_renderer_profile.v1" ||
    typeof renderer.profile_id !== "string"
  ) {
    throw new Error("frontend visual renderer profile is invalid");
  }
  return {
    schema_id: goldenManifestSchemaID,
    renderer_profile_id: renderer.profile_id,
    goldens: committedPNGs(snapshotRoot).map((file) => ({
      path: normalizePath(path.join(visualSnapshotRoot, path.basename(file))),
      sha256: sha256File(file),
    })),
  };
}

export function validateFrontendVisualGoldenManifest(
  manifest,
  { root = defaultRoot, snapshotRoot = path.join(root, visualSnapshotRoot) } = {},
) {
  const expected = buildFrontendVisualGoldenManifest({ root, snapshotRoot });
  const errors = [];
  if (manifest?.schema_id !== expected.schema_id) {
    errors.push(`unexpected golden manifest schema ${manifest?.schema_id}`);
  }
  if (manifest?.renderer_profile_id !== expected.renderer_profile_id) {
    errors.push("golden manifest renderer profile does not match the active profile");
  }
  if (JSON.stringify(manifest?.goldens ?? null) !== JSON.stringify(expected.goldens)) {
    errors.push("golden manifest paths or SHA-256 identities do not match committed PNGs");
  }
  return errors;
}

export function writeFrontendVisualGoldenManifest({
  root = defaultRoot,
  snapshotRoot = path.join(root, visualSnapshotRoot),
  output = path.join(root, goldenManifestPath),
} = {}) {
  const manifest = buildFrontendVisualGoldenManifest({ root, snapshotRoot });
  const temporary = `${output}.tmp-${process.pid}`;
  try {
    writeFileSync(temporary, `${JSON.stringify(manifest, null, 2)}\n`, {
      encoding: "utf8",
      mode: 0o600,
      flag: "wx",
    });
    renameSync(temporary, output);
  } finally {
    rmSync(temporary, { force: true });
  }
  return manifest;
}

function parseArgs(argv) {
  const options = { root: defaultRoot, snapshotRoot: "", output: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (argument === "--root") options.root = path.resolve(argv[++index] ?? "");
    else if (argument === "--snapshot-root") options.snapshotRoot = path.resolve(argv[++index] ?? "");
    else if (argument === "--output") options.output = path.resolve(argv[++index] ?? "");
    else throw new Error("usage: frontend-visual-golden-manifest.mjs [--root <path>] [--snapshot-root <path>] [--output <path>]");
  }
  return options;
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  try {
    const options = parseArgs(process.argv.slice(2));
    writeFrontendVisualGoldenManifest({
      root: options.root,
      ...(options.snapshotRoot ? { snapshotRoot: options.snapshotRoot } : {}),
      ...(options.output ? { output: options.output } : {}),
    });
  } catch (error) {
    process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
    process.exitCode = 1;
  }
}
