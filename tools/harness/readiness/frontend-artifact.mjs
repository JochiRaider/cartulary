import { createHash, randomUUID } from "node:crypto";
import { execFileSync, spawn } from "node:child_process";
import {
  chmodSync, cpSync, existsSync, lstatSync, mkdirSync, mkdtempSync,
  readFileSync, readdirSync, renameSync, rmSync, writeFileSync,
} from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { validateSchemaSync } from "../contract/index.mjs";
import { loadExecutionTopology } from "../generated-artifacts/execution-topology.mjs";
import { borrowSuiteRuntime, createSuiteRuntime } from "../runtime/suite-runtime.mjs";
import { buildSourceSnapshot } from "../test-catalog/index.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
const digest = (bytes) => `sha256:${createHash("sha256").update(bytes).digest("hex")}`;

export function frontendRunRoot(repoRoot = root, environment = process.env) {
  const runID = environment.CARTULARY_TEST_RUN_ID;
  if (!/^[A-Za-z0-9_.-]+$/u.test(runID ?? "")) throw new Error("frontend artifact requires run identity");
  return path.resolve(repoRoot, environment.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results", runID);
}

function artifactProfile(repoRoot, target) {
  const entries = Object.entries(loadExecutionTopology({ root: repoRoot }).browserBatch.frontend_artifacts);
  const match = entries.find(([, value]) => value.producer_target === target);
  if (!match) throw new Error("unknown frontend artifact producer");
  return { id: match[0], ...match[1] };
}

export function frontendContentDigest(directory, { seal = false } = {}) {
  const hash = createHash("sha256");
  const walk = (current) => {
    const info = lstatSync(current);
    if (info.uid !== process.getuid() || info.isSymbolicLink() || !info.isDirectory()) throw new Error("frontend artifact directory must not be a symlink");
    if (seal) chmodSync(current, 0o700);
    else if ((info.mode & 0o777) !== 0o700) throw new Error("frontend artifact directory must be owner-only");
    for (const name of readdirSync(current).sort()) {
      const file = path.join(current, name);
      const relative = path.relative(directory, file).replaceAll(path.sep, "/");
      const entry = lstatSync(file);
      if (entry.uid !== process.getuid() || entry.isSymbolicLink()) throw new Error("frontend artifact contains an unowned entry or symlink");
      if (entry.isDirectory()) {
        hash.update(`${JSON.stringify(["directory", relative])}\n`);
        walk(file);
      } else if (entry.isFile()) {
        if (entry.nlink !== 1) throw new Error("frontend artifact must not share a file inode");
        if (seal) chmodSync(file, 0o600);
        else if ((entry.mode & 0o777) !== 0o600) throw new Error("frontend artifact file must be owner-only");
        hash.update(`${JSON.stringify(["file", relative, digest(readFileSync(file))])}\n`);
      } else throw new Error("frontend artifact contains an unsupported entry");
    }
  };
  walk(directory);
  return `sha256:${hash.digest("hex")}`;
}

function provenance(repoRoot, runRoot) {
  const manifest = path.join(runRoot, "run-manifest.json");
  if (existsSync(manifest)) {
    const value = JSON.parse(readFileSync(manifest, "utf8"));
    return { source_digest: value.source_digest, toolchain_digest: value.toolchain_digest };
  }
  return {
    source_digest: buildSourceSnapshot(repoRoot).digest,
    toolchain_digest: digest(readFileSync(path.join(repoRoot, "tools/toolchain_pins.json"))),
  };
}

export function resolveFrontendArtifact(repoRoot, target, environment = process.env) {
  const runRoot = frontendRunRoot(repoRoot, environment);
  const runtime = borrowSuiteRuntime({ repoRoot, runRoot, environment });
  const profile = artifactProfile(repoRoot, target);
  const receiptRef = `${target}/frontend-artifact.json`;
  const receiptPath = path.join(runRoot, receiptRef);
  const receiptParent = lstatSync(path.dirname(receiptPath));
  if (!receiptParent.isDirectory() || receiptParent.isSymbolicLink()) throw new Error("frontend receipt directory must not be a symlink");
  const info = lstatSync(receiptPath);
  if (!info.isFile() || info.isSymbolicLink()) throw new Error("frontend artifact receipt must be regular");
  const receiptBytes = readFileSync(receiptPath);
  const receipt = JSON.parse(receiptBytes);
  validateSchemaSync(receipt.schema_id, receipt);
  const expected = {
    schema_id: "cartulary.frontend_build_artifact.v1",
    run_id: runtime.runID,
    artifact_id: profile.id,
    producer_unit_id: `target:${target}`,
    ...provenance(repoRoot, runRoot),
  };
  for (const [key, value] of Object.entries(expected)) {
    if (receipt[key] !== value) throw new Error(`frontend artifact ${key} mismatch`);
  }
  const directory = runtime.privatePath(`frontend-${profile.id}`);
  for (const entry of profile.entries) {
    const file = lstatSync(path.join(directory, entry));
    if (!file.isFile() || file.isSymbolicLink()) throw new Error("frontend artifact required entry is missing or invalid");
  }
  if (frontendContentDigest(directory) !== receipt.content_digest) throw new Error("frontend build digest mismatch");
  return { directory, receipt, receiptRef, receiptDigest: digest(receiptBytes) };
}

// The completed private directory is the source for conventional packaging
// outputs. Browser readers never resolve this mutable publication path.
export function publishFrontendOutput(directory, destination) {
  if (lstatSync(destination, { throwIfNoEntry: false })?.isSymbolicLink()) throw new Error("frontend publication must not replace a symlink");
  const parent = path.dirname(destination);
  const staging = mkdtempSync(path.join(parent, ".frontend-publish-"));
  const previous = `${staging}-previous`;
  let published = false;
  try {
    cpSync(directory, staging, { recursive: true });
    if (existsSync(destination)) renameSync(destination, previous);
    try { renameSync(staging, destination); published = true; }
    catch (error) {
      if (existsSync(previous) && !existsSync(destination)) renameSync(previous, destination);
      throw error;
    }
  } finally {
    rmSync(staging, { recursive: true, force: true });
    if (published) rmSync(previous, { recursive: true, force: true });
  }
}

export function sealFrontendArtifact({ repoRoot, runRoot, runtime, profile, staging }) {
  if (path.dirname(path.resolve(staging)) !== runtime.root) throw new Error("frontend staging must belong to its private suite root");
  for (const entry of profile.entries) {
    const info = lstatSync(path.join(staging, entry));
    if (!info.isFile() || info.isSymbolicLink()) throw new Error("frontend artifact required entry is invalid");
  }
  const receipt = {
    schema_id: "cartulary.frontend_build_artifact.v1",
    run_id: runtime.runID,
    artifact_id: profile.id,
    producer_unit_id: `target:${profile.producer_target}`,
    ...provenance(repoRoot, runRoot),
    content_digest: frontendContentDigest(staging, { seal: true }),
  };
  validateSchemaSync(receipt.schema_id, receipt);
  const directory = runtime.privatePath(`frontend-${profile.id}`);
  renameSync(staging, directory);
  const receiptDir = path.join(runRoot, profile.producer_target);
  mkdirSync(receiptDir, { recursive: true, mode: 0o700 });
  const temporary = path.join(receiptDir, `.frontend-artifact-${randomUUID()}.json`);
  try {
    writeFileSync(temporary, `${JSON.stringify(receipt, null, 2)}\n`, { flag: "wx", mode: 0o600 });
    renameSync(temporary, path.join(receiptDir, "frontend-artifact.json"));
  } finally { rmSync(temporary, { force: true }); }
  return directory;
}

async function build(target, command, args) {
  const runRoot = frontendRunRoot();
  mkdirSync(runRoot, { recursive: true, mode: 0o700 });
  const owned = !process.env.CARTULARY_HARNESS_SUITE_RUNTIME_ROOT;
  const runtime = owned
    ? createSuiteRuntime({ repoRoot: root, runRoot, runID: process.env.CARTULARY_TEST_RUN_ID })
    : borrowSuiteRuntime({ repoRoot: root, runRoot });
  const profile = artifactProfile(root, target);
  const completed = runtime.privatePath(`frontend-${profile.id}`);
  if (!owned && existsSync(completed)) {
    resolveFrontendArtifact(root, target);
    return;
  }
  const expectedProvenance = provenance(root, runRoot);
  const validateBuildInputs = () => {
    if (buildSourceSnapshot(root).digest !== expectedProvenance.source_digest ||
      digest(readFileSync(path.join(root, "tools/toolchain_pins.json"))) !== expectedProvenance.toolchain_digest) {
      throw new Error("frontend build inputs changed after the run source snapshot");
    }
  };
  const staging = mkdtempSync(runtime.privatePath(`frontend-staging-${profile.id}-`));
  try {
    validateBuildInputs();
    const status = await new Promise((resolve, reject) => {
      const child = spawn(command, [...args, "--outDir", staging, "--emptyOutDir"], { cwd: root, stdio: "inherit" });
      const terminate = (signal) => child.kill(signal);
      const interrupt = () => terminate("SIGINT");
      const cancel = () => terminate("SIGTERM");
      process.once("SIGINT", interrupt);
      process.once("SIGTERM", cancel);
      const removeHandlers = () => {
        process.off("SIGINT", interrupt);
        process.off("SIGTERM", cancel);
      };
      child.once("error", (error) => { removeHandlers(); reject(error); });
      child.once("close", (code, signal) => { removeHandlers(); resolve(signal ? 1 : code); });
    });
    if (status !== 0) throw new Error(`frontend build failed with status ${status}`);
    validateBuildInputs();
    const directory = sealFrontendArtifact({ repoRoot: root, runRoot, runtime, profile, staging });
    const lockRoot = path.join(root, ".cache/cartulary");
    mkdirSync(lockRoot, { recursive: true, mode: 0o700 });
    // Only the brief conventional-output publication is serialized. Builds and
    // every browser consumer continue independently on their private artifacts.
    execFileSync("flock", ["-x", path.join(lockRoot, "frontend-publication.lock"),
      process.execPath, fileURLToPath(import.meta.url), "publish", target, directory], { stdio: "inherit" });
  } finally {
    rmSync(staging, { recursive: true, force: true });
    if (owned) runtime.close();
  }
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const [operation, target, ...args] = process.argv.slice(2);
  try {
    if (operation === "build") await build(target, args[0], args.slice(1));
    else if (operation === "publish") publishFrontendOutput(args[0], path.join(root, artifactProfile(root, target).path));
    else if (operation === "resolve") process.stdout.write(`${resolveFrontendArtifact(root, target).directory}\n`);
    else throw new Error("usage: frontend-artifact.mjs build <target> <command> [args...] | resolve <target>");
  } catch (error) {
    process.stderr.write(`frontend artifact: ${error.message}\n`);
    process.exitCode = 11;
  }
}
