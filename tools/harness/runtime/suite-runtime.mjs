import { createHash, randomUUID } from "node:crypto";
import {
  chmodSync,
  closeSync,
  createReadStream,
  existsSync,
  lstatSync,
  mkdirSync,
  mkdtempSync,
  openSync,
  readFileSync,
  readdirSync,
  realpathSync,
  renameSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";

const ownerSchemaID = "cartulary.harness_suite_runtime_owner.v1";
const ownerFilename = "runtime-owner.json";
const runtimePrefix = "suite-";
const stagingPrefix = ".creating-suite-";
const staleAgeMS = 24 * 60 * 60 * 1000;
const identityPattern = /^[A-Za-z0-9_.-]+$/u;
const sensitiveEnvironmentName =
  /(?:PASSWORD|SECRET|TOKEN(?!S)|COOKIE|DSN|ACCESS_KEY|KEY_RING|PRIVATE_KEY)/u;
const forbiddenBasenames = [
  /^(?:stack|suite|test-services).*\.env$/iu,
  /^postgres\.recovery\.dsn$/u,
  /^performance-fixture-runtime\.json$/u,
  /^suite-environment\.json$/u,
  /^suite-lease\.json$/u,
  /^stack\.lease$/u,
  /^test-route-token$/u,
  /key[-_.]?ring/iu,
];
const secretSyntaxPatterns = [
  /"[^"]*(?:password|secret|token(?!s)|cookie|dsn|access_key|private_key)[^"]*"\s*:\s*"(?!<redacted>)[^"]+"/iu,
  /(?:PASSWORD|SECRET|TOKEN(?!S)|COOKIE|DSN|ACCESS_KEY|PRIVATE_KEY)=(?!<redacted>)[^\s]+/u,
  /(?:postgres|postgresql):\/\/[^\s/:]+:(?!<redacted>)[^\s/@]+@/iu,
];

function contained(parent, child) {
  const relative = path.relative(parent, child);
  return relative === "" || (!relative.startsWith("..") && !path.isAbsolute(relative));
}

function assertNoSymlinkComponents(candidate) {
  const resolved = path.resolve(candidate);
  const parsed = path.parse(resolved);
  let current = parsed.root;
  for (const component of resolved.slice(parsed.root.length).split(path.sep).filter(Boolean)) {
    current = path.join(current, component);
    if (!existsSync(current)) break;
    if (lstatSync(current).isSymbolicLink()) {
      throw new Error(`suite runtime path contains a symlink component: ${current}`);
    }
  }
}

function assertPrivateDirectory(directory, label) {
  assertNoSymlinkComponents(directory);
  const info = lstatSync(directory);
  if (!info.isDirectory() || info.isSymbolicLink() || (info.mode & 0o777) !== 0o700) {
    throw new Error(`${label} must be a non-symlink owner-only 0700 directory`);
  }
  if (typeof process.getuid === "function" && info.uid !== process.getuid()) {
    throw new Error(`${label} must be owned by the current user`);
  }
}

function ensurePrivateDirectory(directory, label) {
  assertNoSymlinkComponents(path.dirname(directory));
  mkdirSync(directory, { recursive: true, mode: 0o700 });
  chmodSync(directory, 0o700);
  assertPrivateDirectory(directory, label);
}

function readOwner(directory) {
  const ownerPath = path.join(directory, ownerFilename);
  const info = lstatSync(ownerPath);
  if (!info.isFile() || info.isSymbolicLink() || (info.mode & 0o777) !== 0o600) {
    throw new Error(`suite runtime owner marker is not a private regular file: ${ownerPath}`);
  }
  const owner = JSON.parse(readFileSync(ownerPath, "utf8"));
  if (
    owner.schema_id !== ownerSchemaID ||
    !identityPattern.test(owner.run_id) ||
    owner.run_id.length > 96 ||
    !/^[0-9a-f-]{36}$/u.test(owner.lease_id)
  ) {
    throw new Error(`suite runtime owner marker is invalid: ${ownerPath}`);
  }
  return owner;
}

function removeOwnedRuntime(directory, runtimeBase, expectedLeaseID = "") {
  const resolved = path.resolve(directory);
  if (
    path.dirname(resolved) !== runtimeBase ||
    !path.basename(resolved).startsWith(runtimePrefix)
  ) {
    throw new Error(`suite runtime cleanup target is outside its owned namespace: ${resolved}`);
  }
  assertPrivateDirectory(resolved, "suite runtime root");
  const owner = readOwner(resolved);
  if (expectedLeaseID && owner.lease_id !== expectedLeaseID) {
    throw new Error("suite runtime cleanup lease identity does not match ownership proof");
  }
  rmSync(resolved, { recursive: true, force: false });
}

export function closeSuiteRuntimeRoot({ root, expectedLeaseID }) {
  const resolved = path.resolve(root);
  const runtimeBase = path.dirname(resolved);
  assertPrivateDirectory(runtimeBase, "suite runtime base");
  removeOwnedRuntime(resolved, runtimeBase, expectedLeaseID);
}

export function cleanupStaleSuiteRuntimeRoots({
  repoRoot,
  runRoot,
  scratchRoot = process.env.CARTULARY_HARNESS_SCRATCH_ROOT,
  now = Date.now(),
  maxEntries = 64,
  beforeCandidateInspection,
} = {}) {
  if (
    beforeCandidateInspection !== undefined &&
    typeof beforeCandidateInspection !== "function"
  ) {
    throw new TypeError("suite runtime janitor inspection callback must be a function");
  }
  const base = runtimeBasePath({ repoRoot, runRoot, scratchRoot, create: false });
  if (!existsSync(base)) return { scanned: 0, removed: 0 };
  assertPrivateDirectory(base, "suite runtime base");
  let scanned = 0;
  let removed = 0;
  for (const entry of readdirSync(base, { withFileTypes: true })
    .filter(
      (item) => item.name.startsWith(runtimePrefix) || item.name.startsWith(stagingPrefix),
    )
    .sort((left, right) => left.name.localeCompare(right.name))
    .slice(0, maxEntries)) {
    scanned += 1;
    const candidate = path.join(base, entry.name);
    if (!entry.isDirectory() || entry.isSymbolicLink()) {
      throw new Error(`suite runtime janitor found an unowned entry: ${candidate}`);
    }
    beforeCandidateInspection?.(candidate);
    try {
      assertPrivateDirectory(candidate, "suite runtime janitor candidate");
      if (entry.name.startsWith(stagingPrefix)) {
        const info = lstatSync(candidate);
        if (now - info.mtimeMs < staleAgeMS) continue;
        const contents = readdirSync(candidate).sort((left, right) =>
          left.localeCompare(right),
        );
        if (contents.length === 1 && contents[0] === ownerFilename) {
          readOwner(candidate);
        } else if (contents.length !== 0) {
          throw new Error(`suite runtime janitor found an invalid staging root: ${candidate}`);
        }
        rmSync(candidate, { recursive: true, force: false });
        removed += 1;
        continue;
      }
      const owner = readOwner(candidate);
      const createdAt = Date.parse(owner.created_at);
      if (!Number.isFinite(createdAt)) {
        throw new Error(`suite runtime janitor found an invalid creation time: ${candidate}`);
      }
      if (now - createdAt >= staleAgeMS) {
        removeOwnedRuntime(candidate, base, owner.lease_id);
        removed += 1;
      }
    } catch (error) {
      if (error?.code === "ENOENT") continue;
      throw error;
    }
  }
  return { scanned, removed };
}

function runtimeBasePath({ repoRoot, runRoot, scratchRoot, create }) {
  const repository = realpathSync(path.resolve(repoRoot));
  const retained = realpathSync(path.resolve(runRoot));
  const scratch = path.resolve(
    scratchRoot || path.join(tmpdir(), "cartulary-harness-scratch"),
  );
  if (contained(repository, scratch) || contained(retained, scratch)) {
    throw new Error("suite runtime scratch root must be outside repository and result roots");
  }
  const base = path.join(scratch, "suite-runtime");
  if (create) ensurePrivateDirectory(base, "suite runtime base");
  if (existsSync(base)) {
    const canonical = realpathSync(base);
    if (contained(repository, canonical) || contained(retained, canonical)) {
      throw new Error("suite runtime base resolves inside repository or result roots");
    }
  }
  return base;
}

export function createSuiteRuntime({ repoRoot, runRoot, runID, scratchRoot } = {}) {
  if (!identityPattern.test(String(runID ?? "")) || String(runID).length > 96) {
    throw new Error("suite runtime requires a safe run identity");
  }
  const runtimeBase = runtimeBasePath({ repoRoot, runRoot, scratchRoot, create: true });
  cleanupStaleSuiteRuntimeRoots({ repoRoot, runRoot, scratchRoot });
  const staging = mkdtempSync(path.join(runtimeBase, `${stagingPrefix}${runID}-`));
  chmodSync(staging, 0o700);
  assertPrivateDirectory(staging, "suite runtime staging root");
  const root = path.join(
    runtimeBase,
    path.basename(staging).replace(stagingPrefix, runtimePrefix),
  );
  const leaseID = randomUUID();
  const owner = {
    schema_id: ownerSchemaID,
    lease_id: leaseID,
    run_id: runID,
    owner_uid: typeof process.getuid === "function" ? process.getuid() : -1,
    created_at: new Date().toISOString(),
  };
  try {
    const ownerPath = path.join(staging, ownerFilename);
    const descriptor = openSync(ownerPath, "wx", 0o600);
    try {
      writeFileSync(descriptor, `${JSON.stringify(owner)}\n`, "utf8");
    } finally {
      closeSync(descriptor);
    }
    readOwner(staging);
    renameSync(staging, root);
    assertPrivateDirectory(root, "suite runtime root");
    readOwner(root);
  } catch (error) {
    if (existsSync(staging)) rmSync(staging, { recursive: true, force: true });
    throw error;
  }
  const secrets = new Set();
  let closed = false;
  return {
    root,
    leaseID,
    runID,
    privatePath(...parts) {
      if (parts.length === 0 || parts.some((part) => !identityPattern.test(String(part)))) {
        throw new Error("suite runtime child path requires safe identity components");
      }
      const child = path.join(root, ...parts.map(String));
      if (!contained(root, child)) throw new Error("suite runtime child path escapes root");
      return child;
    },
    registerSecret(value) {
      const secret = String(value ?? "");
      if (secret.length >= 8) secrets.add(secret);
    },
    registerEnvironment(environment) {
      for (const [name, value] of Object.entries(environment ?? {})) {
        if (sensitiveEnvironmentName.test(name)) this.registerSecret(value);
      }
    },
    forbiddenValues() {
      return [...secrets];
    },
    close() {
      if (closed) return;
      removeOwnedRuntime(root, runtimeBase, leaseID);
      closed = true;
    },
  };
}

export function borrowSuiteRuntime({ repoRoot, runRoot, environment = process.env } = {}) {
  const root = path.resolve(String(environment.CARTULARY_HARNESS_SUITE_RUNTIME_ROOT ?? ""));
  const leaseID = String(environment.CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID ?? "");
  const runID = String(environment.CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID ?? "");
  if (!root || !leaseID || !runID) throw new Error("suite runtime environment is incomplete");
  assertPrivateDirectory(root, "borrowed suite runtime root");
  const repository = realpathSync(path.resolve(repoRoot));
  const retained = realpathSync(path.resolve(runRoot));
  const resolved = realpathSync(root);
  if (contained(repository, resolved) || contained(retained, resolved)) {
    throw new Error("borrowed suite runtime resolves inside repository or result roots");
  }
  const owner = readOwner(resolved);
  if (owner.lease_id !== leaseID || owner.run_id !== runID) {
    throw new Error("borrowed suite runtime identity does not match ownership proof");
  }
  return {
    leaseID,
    root: resolved,
    runID,
    privatePath(...parts) {
      if (parts.length === 0 || parts.some((part) => !identityPattern.test(String(part)))) {
        throw new Error("suite runtime child path requires safe identity components");
      }
      const child = path.join(resolved, ...parts.map(String));
      if (!contained(resolved, child)) throw new Error("suite runtime child path escapes root");
      return child;
    },
  };
}

async function scanFile(file, forbiddenValues, removeUnsafe) {
  const maxNeedle = Math.max(
    256,
    ...forbiddenValues.map((value) => Buffer.byteLength(value, "utf8")),
  );
  let overlap = "";
  for await (const chunk of createReadStream(file, { highWaterMark: 64 * 1024 })) {
    const text = overlap + chunk.toString("utf8");
    if (secretSyntaxPatterns.some((pattern) => pattern.test(text))) {
      if (removeUnsafe) rmSync(file, { force: true });
      throw new Error(`retained evidence contains secret-capable syntax: ${file}`);
    }
    if (forbiddenValues.some((value) => text.includes(value))) {
      if (removeUnsafe) rmSync(file, { force: true });
      throw new Error(`retained evidence contains a registered runtime secret: ${file}`);
    }
    overlap = text.slice(-maxNeedle);
  }
}

export async function scanRetainedRoot(
  runRoot,
  { forbiddenValues = [], removeUnsafe = false } = {},
) {
  const configuredRoot = path.resolve(runRoot);
  const configuredInfo = lstatSync(configuredRoot);
  if (configuredInfo.isSymbolicLink() || !configuredInfo.isDirectory()) {
    throw new Error("retained run root must be a non-symlink directory");
  }
  if ((configuredInfo.mode & 0o077) !== 0) {
    if (removeUnsafe) chmodSync(configuredRoot, 0o700);
    throw new Error("retained run root must be owner-only");
  }
  if (typeof process.getuid === "function" && configuredInfo.uid !== process.getuid()) {
    throw new Error("retained run root must be owned by the current user");
  }
  const root = realpathSync(configuredRoot);
  const normalizedSecrets = [...new Set(forbiddenValues.map(String).filter((value) => value.length >= 8))];
  let files = 0;
  const visit = async (directory) => {
    for (const entry of readdirSync(directory, { withFileTypes: true }).sort((left, right) =>
      left.name.localeCompare(right.name),
    )) {
      const candidate = path.join(directory, entry.name);
      const info = lstatSync(candidate);
      if (info.isSymbolicLink()) {
        if (removeUnsafe) rmSync(candidate, { force: true });
        throw new Error(`retained evidence contains a symlink: ${candidate}`);
      }
      if (info.isDirectory()) {
        if ((info.mode & 0o077) !== 0) {
          if (removeUnsafe) chmodSync(candidate, 0o700);
          throw new Error(`retained evidence directory is not owner-only: ${candidate}`);
        }
        await visit(candidate);
        continue;
      }
      if (!info.isFile()) {
        if (removeUnsafe) rmSync(candidate, { force: true });
        throw new Error(`retained evidence contains a non-regular entry: ${candidate}`);
      }
      if (forbiddenBasenames.some((pattern) => pattern.test(entry.name))) {
        if (removeUnsafe) rmSync(candidate, { force: true });
        throw new Error(`retained evidence contains a forbidden runtime filename: ${candidate}`);
      }
      if ((info.mode & 0o077) !== 0) {
        if (removeUnsafe) chmodSync(candidate, 0o600);
        throw new Error(`retained evidence file is not owner-only: ${candidate}`);
      }
      files += 1;
      await scanFile(candidate, normalizedSecrets, removeUnsafe);
    }
  };
  await visit(root);
  return {
    schema_id: "cartulary.harness_retained_secret_scan.v1",
    status: "pass",
    scanned_files: files,
    policy_digest: `sha256:${createHash("sha256")
      .update(forbiddenBasenames.map(String).join("\n"))
      .update(secretSyntaxPatterns.map(String).join("\n"))
      .digest("hex")}`,
  };
}
