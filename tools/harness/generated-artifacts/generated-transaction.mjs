#!/usr/bin/env node
import {
  cpSync,
  existsSync,
  lstatSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  readdirSync,
  renameSync,
  rmSync,
} from "node:fs";
import path from "node:path";
import process from "node:process";

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function validateRelative(relative) {
  if (
    typeof relative !== "string" ||
    !relative ||
    path.isAbsolute(relative) ||
    relative.split(/[\\/]/u).includes("..")
  ) {
    throw new Error(`unsafe generated path ${relative}`);
  }
  return relative;
}

function equalPath(left, right) {
  if (!existsSync(left) || !existsSync(right)) return existsSync(left) === existsSync(right);
  const leftInfo = lstatSync(left);
  const rightInfo = lstatSync(right);
  if (leftInfo.isSymbolicLink() || rightInfo.isSymbolicLink()) return false;
  if (leftInfo.isFile() && rightInfo.isFile()) {
    return leftInfo.mode === rightInfo.mode && readFileSync(left).equals(readFileSync(right));
  }
  if (!leftInfo.isDirectory() || !rightInfo.isDirectory()) return false;
  const leftNames = readdirSync(left).sort(compareASCII);
  const rightNames = readdirSync(right).sort(compareASCII);
  return (
    JSON.stringify(leftNames) === JSON.stringify(rightNames) &&
    leftNames.every((name) => equalPath(path.join(left, name), path.join(right, name)))
  );
}

export function changedGeneratedPaths(repoRoot, renderedRoot, generatedPaths) {
  return [...generatedPaths]
    .map(validateRelative)
    .filter((relative) =>
      !equalPath(path.join(repoRoot, relative), path.join(renderedRoot, relative)),
    )
    .sort(compareASCII);
}

export function publishGeneratedTransaction({
  repoRoot,
  renderedRoot,
  generatedPaths,
  transactionRoot,
  afterMove = () => {},
}) {
  const changed = changedGeneratedPaths(repoRoot, renderedRoot, generatedPaths);
  if (changed.length === 0) return { status: "unchanged", changed: [] };
  const root = transactionRoot ?? path.join(repoRoot, "tmp");
  mkdirSync(root, { recursive: true });
  const transaction = mkdtempSync(path.join(root, "generated-publish."));
  const staging = path.join(transaction, "staging");
  const backup = path.join(transaction, "backup");
  const moved = [];
  try {
    for (const relative of changed) {
      const source = path.join(renderedRoot, relative);
      if (!existsSync(source)) throw new Error(`render omitted generated path ${relative}`);
      const staged = path.join(staging, relative);
      mkdirSync(path.dirname(staged), { recursive: true });
      cpSync(source, staged, { recursive: true, preserveTimestamps: true });
    }
    for (const relative of changed) {
      const destination = path.join(repoRoot, relative);
      const prior = path.join(backup, relative);
      const staged = path.join(staging, relative);
      mkdirSync(path.dirname(destination), { recursive: true });
      if (existsSync(destination)) {
        mkdirSync(path.dirname(prior), { recursive: true });
        renameSync(destination, prior);
      }
      renameSync(staged, destination);
      moved.push(relative);
      afterMove(relative, moved.length);
    }
    rmSync(transaction, { recursive: true, force: true });
    return { status: "published", changed };
  } catch (error) {
    for (const relative of [...moved].reverse()) {
      const destination = path.join(repoRoot, relative);
      const prior = path.join(backup, relative);
      rmSync(destination, { recursive: true, force: true });
      if (existsSync(prior)) {
        mkdirSync(path.dirname(destination), { recursive: true });
        renameSync(prior, destination);
      }
    }
    rmSync(transaction, { recursive: true, force: true });
    throw error;
  }
}

function parseCLI(argv) {
  const options = { repo: "", rendered: "", manifest: "", mode: "check" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--repo") options.repo = argv[++index] ?? "";
    else if (arg === "--rendered") options.rendered = argv[++index] ?? "";
    else if (arg === "--manifest") options.manifest = argv[++index] ?? "";
    else if (arg === "--mode") options.mode = argv[++index] ?? "";
    else throw new Error("invalid generated transaction arguments");
  }
  if (!options.repo || !options.rendered || !options.manifest || !["check", "refresh"].includes(options.mode)) {
    throw new Error("generated transaction requires repo, rendered, manifest, and check|refresh mode");
  }
  return options;
}

function main() {
  const options = parseCLI(process.argv.slice(2));
  const manifest = JSON.parse(readFileSync(options.manifest, "utf8"));
  const generatedPaths = manifest.generated_paths;
  if (!Array.isArray(generatedPaths) || generatedPaths.length === 0) {
    throw new Error("generated transaction manifest has no generated_paths");
  }
  const changed = changedGeneratedPaths(options.repo, options.rendered, generatedPaths);
  if (options.mode === "check" && changed.length > 0) {
    throw new Error(`generated artifact drift: ${changed.join(", ")}`);
  }
  const result =
    options.mode === "refresh"
      ? publishGeneratedTransaction({
          repoRoot: options.repo,
          renderedRoot: options.rendered,
          generatedPaths,
        })
      : { status: "unchanged", changed: [] };
  process.stdout.write(`${JSON.stringify(result)}\n`);
}

if (import.meta.url === new URL(`file://${process.argv[1]}`).href) {
  try {
    main();
  } catch (error) {
    process.stderr.write(`${error.message}\n`);
    process.exitCode = 1;
  }
}
