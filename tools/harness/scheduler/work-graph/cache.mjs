import { createHash, randomUUID } from "node:crypto";
import {
  chmodSync,
  existsSync,
  lstatSync,
  mkdirSync,
  readFileSync,
  readdirSync,
  renameSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";

import { semanticJSONDigest, validateSchemaSync } from "../../contract/index.mjs";
import { resolveCacheDependencyClosure } from "./cache-dependencies.mjs";

const modes = new Set(["normal", "cold", "off"]);
const cacheRecordSchemaID = "cartulary.harness_cache_record.v2";
export const workGraphCacheRootRelative = ".cache/cartulary/graph-v2";

class CacheEntryError extends Error {}

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function sha256(bytes) {
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}

function safeSegment(value) {
  return value.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-");
}

function normalizedRelative(relative, label = "cache artifact path") {
  if (
    typeof relative !== "string" ||
    relative === "" ||
    path.posix.isAbsolute(relative) ||
    path.win32.isAbsolute(relative) ||
    relative.includes("\\") ||
    path.posix.normalize(relative) !== relative ||
    relative.split("/").some((segment) => segment === "" || segment === "." || segment === "..")
  ) {
    throw new CacheEntryError(`${label} is not a normalized relative path: ${relative}`);
  }
  return relative;
}

function ensureDirectory(directory, label) {
  if (!existsSync(directory)) throw new CacheEntryError(`${label} is missing`);
  const info = lstatSync(directory);
  if (!info.isDirectory() || info.isSymbolicLink()) {
    throw new CacheEntryError(`${label} must be a non-symlink directory`);
  }
  return directory;
}

function ensureContainedDirectory(anchor, directory, label) {
  const resolvedAnchor = path.resolve(anchor);
  const resolvedDirectory = path.resolve(directory);
  ensureDirectory(resolvedAnchor, `${label} anchor`);
  if (
    resolvedDirectory !== resolvedAnchor &&
    !resolvedDirectory.startsWith(`${resolvedAnchor}${path.sep}`)
  ) {
    throw new CacheEntryError(`${label} escapes its anchor`);
  }
  let current = resolvedAnchor;
  for (const segment of path.relative(resolvedAnchor, resolvedDirectory).split(path.sep).filter(Boolean)) {
    current = path.join(current, segment);
    if (!existsSync(current)) {
      mkdirSync(current, { mode: 0o700 });
      chmodSync(current, 0o700);
    }
    const info = lstatSync(current);
    if (!info.isDirectory() || info.isSymbolicLink()) {
      throw new CacheEntryError(`${label} ancestry is unsafe`);
    }
  }
  return resolvedDirectory;
}

function assertContainedDirectory(anchor, directory, label) {
  const resolvedAnchor = path.resolve(anchor);
  const resolvedDirectory = path.resolve(directory);
  ensureDirectory(resolvedAnchor, `${label} anchor`);
  if (
    resolvedDirectory !== resolvedAnchor &&
    !resolvedDirectory.startsWith(`${resolvedAnchor}${path.sep}`)
  ) {
    throw new CacheEntryError(`${label} escapes its anchor`);
  }
  let current = resolvedAnchor;
  for (const segment of path.relative(resolvedAnchor, resolvedDirectory).split(path.sep).filter(Boolean)) {
    current = path.join(current, segment);
    const info = lstatSync(current);
    if (!info.isDirectory() || info.isSymbolicLink()) {
      throw new CacheEntryError(`${label} ancestry is unsafe`);
    }
  }
  return resolvedDirectory;
}

function containedPath(root, relative, { createParents = false } = {}) {
  normalizedRelative(relative);
  const resolvedRoot = path.resolve(root);
  ensureDirectory(resolvedRoot, "cache destination root");
  const segments = relative.split("/");
  let current = resolvedRoot;
  for (const segment of segments.slice(0, -1)) {
    current = path.join(current, segment);
    if (!existsSync(current) && createParents) mkdirSync(current, { mode: 0o700 });
    if (existsSync(current)) {
      const info = lstatSync(current);
      if (!info.isDirectory() || info.isSymbolicLink()) {
        throw new CacheEntryError(`cache destination ancestry is unsafe: ${relative}`);
      }
    } else {
      break;
    }
  }
  const resolved = path.resolve(resolvedRoot, relative);
  if (!resolved.startsWith(`${resolvedRoot}${path.sep}`)) {
    throw new CacheEntryError(`cache artifact path escapes its root: ${relative}`);
  }
  return resolved;
}

function walkInputFiles(root, relative, output) {
  if ([".md", ".markdown", ".mdown", ".mkd"].includes(path.posix.extname(relative).toLowerCase())) {
    return;
  }
  const absolute = containedPath(root, normalizedRelative(relative, "cache input path"));
  if (!existsSync(absolute)) {
    output.push({ path: relative, missing: true });
    return;
  }
  const info = lstatSync(absolute);
  if (info.isSymbolicLink()) throw new Error(`cache input must not be a symlink: ${relative}`);
  if (info.isFile()) {
    output.push({ path: relative, digest: sha256(readFileSync(absolute)) });
    return;
  }
  if (!info.isDirectory()) throw new Error(`cache input has unsupported type: ${relative}`);
  for (const name of readdirSync(absolute).sort(compareASCII)) {
    if (new Set([".cache", ".cartulary", "coverage", "dist", "node_modules", "test-results", "tmp"]).has(name)) continue;
    walkInputFiles(root, path.posix.join(relative, name), output);
  }
}

export function cacheInputRootDigest(root, inputRoots) {
  const entries = [];
  for (const inputRoot of [...inputRoots].sort(compareASCII)) walkInputFiles(root, inputRoot, entries);
  return semanticJSONDigest(entries);
}

function artifactMode(info) {
  return `0${(info.mode & 0o777).toString(8).padStart(3, "0")}`;
}

function assertRegularFile(info, label) {
  if (!info.isFile() || info.isSymbolicLink() || info.nlink !== 1) {
    throw new CacheEntryError(`${label} must be a non-linked regular file`);
  }
}

function directoryMemberDigest(member) {
  return semanticJSONDigest({
    artifact_type: member.artifact_type,
    relative_path: member.relative_path,
    mode: member.mode,
  });
}

function snapshotDirectory(directory) {
  const members = [];
  function visit(current, relative) {
    for (const name of readdirSync(current).sort(compareASCII)) {
      const memberRelative = relative ? `${relative}/${name}` : name;
      normalizedRelative(memberRelative, "cache directory member");
      const absolute = path.join(current, name);
      const info = lstatSync(absolute);
      if (info.isSymbolicLink()) throw new CacheEntryError(`cache output contains a symlink: ${memberRelative}`);
      if (info.isDirectory()) {
        const member = {
          artifact_type: "directory",
          relative_path: memberRelative,
          mode: artifactMode(info),
        };
        members.push({ ...member, digest: directoryMemberDigest(member) });
        visit(absolute, memberRelative);
      } else if (info.isFile()) {
        assertRegularFile(info, `cache output member ${memberRelative}`);
        const bytes = readFileSync(absolute);
        members.push({
          artifact_type: "file",
          relative_path: memberRelative,
          mode: artifactMode(info),
          digest: sha256(bytes),
          bytes,
        });
      } else {
        throw new CacheEntryError(`cache output contains a special file: ${memberRelative}`);
      }
    }
  }
  visit(directory, "");
  return members.sort((left, right) => compareASCII(left.relative_path, right.relative_path));
}

function snapshotArtifacts(unit, root, runRoot, inputDigest) {
  return unit.reusable_artifact_outputs.map((declared) => {
    if (declared.producer_identity !== unit.unit_id) {
      throw new CacheEntryError(`cache output producer mismatch: ${declared.relative_path}`);
    }
    const destinationRoot = declared.destination_class === "run_root" ? runRoot : root;
    const file = containedPath(destinationRoot, declared.relative_path);
    if (!existsSync(file)) throw new CacheEntryError(`cache output is missing: ${declared.relative_path}`);
    const info = lstatSync(file);
    if (info.isSymbolicLink()) throw new CacheEntryError(`cache output is a symlink: ${declared.relative_path}`);
    if (artifactMode(info) !== declared.mode) throw new CacheEntryError(`cache output mode mismatch: ${declared.relative_path}`);
    if (declared.artifact_type === "file") {
      assertRegularFile(info, `cache output ${declared.relative_path}`);
      const bytes = readFileSync(file);
      return {
        manifest: {
          ...declared,
          digest: sha256(bytes),
          semantic_input_digest: inputDigest,
          members: [],
        },
        bytes,
        members: [],
      };
    }
    if (declared.artifact_type !== "directory" || !info.isDirectory()) {
      throw new CacheEntryError(`cache output type mismatch: ${declared.relative_path}`);
    }
    const members = snapshotDirectory(file);
    const manifestMembers = members.map(({ bytes: _bytes, ...member }) => member);
    return {
      manifest: {
        ...declared,
        digest: semanticJSONDigest(manifestMembers),
        semantic_input_digest: inputDigest,
        members: manifestMembers,
      },
      members,
    };
  });
}

function outputDigest(snapshots, semanticResult) {
  return semanticJSONDigest({
    semantic_result: semanticResult,
    artifacts: snapshots.map((snapshot) => snapshot.manifest),
  });
}

function semanticResultForUnit(unit, runRoot) {
  const rows = unit.current_run_evidence_outputs
    .filter((relative) => relative.startsWith("rows/") && relative.endsWith(".json"))
    .map((relative) => {
      const file = containedPath(runRoot, relative);
      const info = lstatSync(file);
      assertRegularFile(info, `cache semantic row result ${relative}`);
      const result = JSON.parse(readFileSync(file, "utf8"));
      validateSchemaSync("cartulary.harness_row_result.v2", result);
      if (result.terminal_state !== "passed") {
        throw new CacheEntryError(`cache semantic row result did not pass: ${result.row_id}`);
      }
      return { row_id: result.row_id, runner: result.runner };
    })
    .sort((left, right) => compareASCII(left.row_id, right.row_id));
  return { status: "passed", rows };
}

function targetForUnit(unit) {
  if (unit.unit_id.startsWith("target:")) return unit.unit_id.slice("target:".length);
  return unit.command.environment.CARTULARY_TEST_TARGET ?? "";
}

function expectedRowIDs(unit) {
  return unit.current_run_evidence_outputs
    .filter((relative) => relative.startsWith("rows/") && relative.endsWith(".json"))
    .map((relative) => relative.slice("rows/".length, -".json".length))
    .sort(compareASCII);
}

function validateSemanticResult(unit, semanticResult) {
  const actual = semanticResult.rows.map((row) => row.row_id).sort(compareASCII);
  if (semanticResult.status !== "passed" || JSON.stringify(actual) !== JSON.stringify(expectedRowIDs(unit))) {
    throw new CacheEntryError("cache semantic result does not match the selected row closure");
  }
}

function validateManifest(unit, inputDigest, snapshots) {
  if (snapshots.length !== unit.reusable_artifact_outputs.length) {
    throw new CacheEntryError("cache artifact manifest is partial or surplus");
  }
  for (const [index, snapshot] of snapshots.entries()) {
    const declared = unit.reusable_artifact_outputs[index];
    const manifest = snapshot.manifest;
    for (const field of ["artifact_type", "relative_path", "destination_class", "mode", "producer_identity"]) {
      if (manifest[field] !== declared[field]) {
        throw new CacheEntryError(`cache artifact ${field} mismatch: ${declared.relative_path}`);
      }
    }
    normalizedRelative(manifest.relative_path);
    if (manifest.producer_identity !== unit.unit_id || manifest.semantic_input_digest !== inputDigest) {
      throw new CacheEntryError(`cache artifact identity mismatch: ${manifest.relative_path}`);
    }
    if (manifest.artifact_type === "file") {
      if (manifest.members.length !== 0 || sha256(snapshot.bytes) !== manifest.digest) {
        throw new CacheEntryError(`cache file artifact mismatch: ${manifest.relative_path}`);
      }
      continue;
    }
    const paths = manifest.members.map((member) => member.relative_path);
    if (new Set(paths).size !== paths.length || JSON.stringify(paths) !== JSON.stringify([...paths].sort(compareASCII))) {
      throw new CacheEntryError(`cache directory manifest is not canonical: ${manifest.relative_path}`);
    }
    const byPath = new Map(manifest.members.map((member) => [member.relative_path, member]));
    for (const [memberIndex, member] of manifest.members.entries()) {
      normalizedRelative(member.relative_path, "cache directory member");
      const segments = member.relative_path.split("/");
      for (let depth = 1; depth < segments.length; depth += 1) {
        if (byPath.get(segments.slice(0, depth).join("/"))?.artifact_type !== "directory") {
          throw new CacheEntryError(`cache directory member has an unproved parent: ${member.relative_path}`);
        }
      }
      if (member.artifact_type === "directory") {
        if (member.digest !== directoryMemberDigest(member)) {
          throw new CacheEntryError(`cache directory member digest mismatch: ${member.relative_path}`);
        }
      } else if (sha256(snapshot.members[memberIndex].bytes) !== member.digest) {
        throw new CacheEntryError(`cache member digest mismatch: ${member.relative_path}`);
      }
    }
    if (semanticJSONDigest(manifest.members) !== manifest.digest) {
      throw new CacheEntryError(`cache directory artifact digest mismatch: ${manifest.relative_path}`);
    }
  }
}

function listEntryTree(root) {
  const files = [];
  const directories = [];
  function visit(current, relative) {
    for (const name of readdirSync(current).sort(compareASCII)) {
      const childRelative = relative ? `${relative}/${name}` : name;
      const absolute = path.join(current, name);
      const info = lstatSync(absolute);
      if (info.isSymbolicLink()) throw new CacheEntryError(`cache entry contains a symlink: ${childRelative}`);
      if (info.isDirectory()) {
        if (artifactMode(info) !== "0700") {
          throw new CacheEntryError(`cache entry directory mode mismatch: ${childRelative}`);
        }
        directories.push(childRelative);
        visit(absolute, childRelative);
      } else if (info.isFile()) {
        assertRegularFile(info, `cache entry payload ${childRelative}`);
        if (artifactMode(info) !== "0600") {
          throw new CacheEntryError(`cache entry payload mode mismatch: ${childRelative}`);
        }
        files.push(childRelative);
      } else {
        throw new CacheEntryError(`cache entry contains a special file: ${childRelative}`);
      }
    }
  }
  visit(root, "");
  return { files, directories };
}

function semanticEvidenceItems(unit, semanticResult, clock) {
  validateSemanticResult(unit, semanticResult);
  const timestamp = clock().toISOString();
  const items = semanticResult.rows.map((row) => {
    const payload = {
      schema_id: "cartulary.harness_row_result.v2",
      row_id: row.row_id,
      terminal_state: "passed",
      duration_ms: 0,
      exit_code: 0,
      failure_class: null,
      failure_reason: null,
      failure_diagnostic: null,
      runner: row.runner,
      started_at: timestamp,
      finished_at: timestamp,
      wall_duration_ms: 0,
    };
    validateSchemaSync(payload.schema_id, payload);
    return {
      artifact_type: "file",
      destination_class: "run_root",
      relative_path: `rows/${row.row_id}.json`,
      mode: "0600",
      bytes: Buffer.from(`${JSON.stringify(payload, null, 2)}\n`, "utf8"),
    };
  });
  const unitResultPath = unit.current_run_evidence_outputs.find((relative) => relative.startsWith("unit-results/"));
  if (!unitResultPath) throw new CacheEntryError(`${unit.unit_id} has no unit-result output`);
  const unitResult = {
    schema_id: "cartulary.harness_unit_result.v1",
    unit_id: unit.unit_id,
    semantic_digest: unit.semantic_digest,
    status: "passed",
    exit_code: null,
    signal: null,
    failure_class: null,
    failure_reason: null,
    evidence_outputs: unit.current_run_evidence_outputs,
    missing_outputs: [],
  };
  validateSchemaSync(unitResult.schema_id, unitResult);
  items.push({
    artifact_type: "file",
    destination_class: "run_root",
    relative_path: unitResultPath,
    mode: "0600",
    bytes: Buffer.from(`${JSON.stringify(unitResult, null, 2)}\n`, "utf8"),
  });
  return items;
}

function artifactPublishItems(snapshots) {
  return snapshots.map((snapshot) => ({
    ...snapshot.manifest,
    ...(snapshot.manifest.artifact_type === "file" ? { bytes: snapshot.bytes } : { members: snapshot.members }),
  }));
}

function assertSafeExistingOutput(output, label) {
  if (!existsSync(output)) return;
  const info = lstatSync(output);
  if (info.isSymbolicLink()) throw new CacheEntryError(`${label} is a symlink`);
  if (info.isFile()) {
    assertRegularFile(info, label);
    return;
  }
  if (!info.isDirectory()) throw new CacheEntryError(`${label} is a special file`);
  snapshotDirectory(output);
}

function materializeItem(stage, item) {
  if (item.artifact_type === "file") {
    writeFileSync(stage, item.bytes, { mode: Number.parseInt(item.mode.slice(1), 8) });
    chmodSync(stage, Number.parseInt(item.mode.slice(1), 8));
    return;
  }
  mkdirSync(stage, { mode: 0o700 });
  for (const member of item.members.filter((entry) => entry.artifact_type === "directory")) {
    mkdirSync(path.join(stage, member.relative_path), { recursive: true, mode: 0o700 });
  }
  for (const member of item.members.filter((entry) => entry.artifact_type === "file")) {
    const output = path.join(stage, member.relative_path);
    mkdirSync(path.dirname(output), { recursive: true, mode: 0o700 });
    writeFileSync(output, member.bytes, { mode: Number.parseInt(member.mode.slice(1), 8) });
    chmodSync(output, Number.parseInt(member.mode.slice(1), 8));
  }
  for (const member of [...item.members].reverse()) {
    if (member.artifact_type === "directory") {
      chmodSync(path.join(stage, member.relative_path), Number.parseInt(member.mode.slice(1), 8));
    }
  }
  chmodSync(stage, Number.parseInt(item.mode.slice(1), 8));
}

function underPath(candidate, root) {
  return candidate === root || candidate.startsWith(`${root}${path.sep}`);
}

function publishTransaction(items, root, runRoot) {
  const transactionID = randomUUID();
  const staged = [];
  const published = [];
  const backups = [];
  const destinations = new Set();
  try {
    for (const item of items) {
      const destinationRoot = item.destination_class === "run_root" ? runRoot : root;
      const output = containedPath(destinationRoot, item.relative_path, { createParents: true });
      if ([...destinations].some((other) => underPath(output, other) || underPath(other, output))) {
        throw new CacheEntryError(`cache outputs overlap: ${item.relative_path}`);
      }
      destinations.add(output);
      assertSafeExistingOutput(output, `cache destination ${item.relative_path}`);
      const stage = path.join(path.dirname(output), `.cartulary-cache-stage-${path.basename(output)}-${transactionID}`);
      if (existsSync(stage)) throw new CacheEntryError("cache staging collision");
      materializeItem(stage, item);
      staged.push({ output, stage });
    }
    for (const entry of staged) {
      if (existsSync(entry.output)) {
        const backup = `${entry.output}.cartulary-cache-backup-${transactionID}`;
        renameSync(entry.output, backup);
        backups.push({ output: entry.output, backup });
      }
      renameSync(entry.stage, entry.output);
      published.push(entry.output);
    }
    for (const { backup } of backups) rmSync(backup, { recursive: true, force: true });
  } catch (error) {
    for (const output of [...published].reverse()) rmSync(output, { recursive: true, force: true });
    for (const { output, backup } of [...backups].reverse()) {
      if (existsSync(backup)) renameSync(backup, output);
    }
    for (const { stage } of staged) rmSync(stage, { recursive: true, force: true });
    throw error;
  }
}

export function loadCacheRegistry(root) {
  const registry = JSON.parse(readFileSync(path.join(root, "tools/harness_cache_registry.json"), "utf8"));
  validateSchemaSync(registry.schema_id, registry);
  const profiles = new Map();
  const targetProfiles = new Map();
  for (const profile of registry.profiles) {
    if (profiles.has(profile.profile_id)) throw new Error(`duplicate cache profile ${profile.profile_id}`);
    profiles.set(profile.profile_id, profile);
    for (const target of profile.targets) {
      if (targetProfiles.has(target)) throw new Error(`duplicate cache target ${target}`);
      targetProfiles.set(target, profile);
    }
  }
  return { registry, profiles, targetProfiles };
}

export class WorkGraphCache {
  constructor({ root, runRoot, cacheRoot, mode = "normal", registry, toolchainDigest, helperDigest, vulnerabilityDatabaseRevision = "", sourceEntries, clock = () => new Date() }) {
    if (!modes.has(mode)) throw new Error(`unsupported graph cache mode ${mode}`);
    this.root = path.resolve(root);
    this.runRoot = path.resolve(runRoot);
    this.cacheRoot = path.resolve(cacheRoot);
    this.mode = mode;
    this.toolchainDigest = toolchainDigest;
    this.helperDigest = helperDigest;
    this.vulnerabilityDatabaseRevision = vulnerabilityDatabaseRevision;
    this.sourceEntries = sourceEntries;
    this.clock = clock;
    const loaded = registry
      ? (() => {
          validateSchemaSync(registry.schema_id, registry);
          const profiles = new Map(registry.profiles.map((entry) => [entry.profile_id, entry]));
          const targetProfiles = new Map(registry.profiles.flatMap((entry) => entry.targets.map((target) => [target, entry])));
          return { profiles, targetProfiles };
        })()
      : loadCacheRegistry(this.root);
    this.profiles = loaded.profiles;
    this.targetProfiles = loaded.targetProfiles;
    this.sameRun = new Map();
    this.dependencyClosures = new Map();
    this.dependencyResolverCache = new Map();
    this.contexts = new Map();
  }

  dependencyClosure(profile, unit) {
    const selector = unit.command.environment.CARTULARY_TEST_ROWS ?? unit.unit_id;
    const key = `${profile.profile_id}:${selector}`;
    if (!this.dependencyClosures.has(key)) {
      this.dependencyClosures.set(
        key,
        this.sourceEntries
          ? resolveCacheDependencyClosure({
              root: this.root,
              entries: this.sourceEntries,
              profile,
              unit,
              resolverCache: this.dependencyResolverCache,
            })
          : { strategy: "filesystem_broad", digest: cacheInputRootDigest(this.root, profile.input_roots) },
      );
    }
    return this.dependencyClosures.get(key);
  }

  context(unit) {
    if (unit.cache_policy === "none") return { eligible: false, reason: "policy_none" };
    const hasSemanticRows = expectedRowIDs(unit).length > 0;
    if (unit.reusable_artifact_outputs.length === 0 && !hasSemanticRows) {
      return { eligible: false, reason: "reusable_outputs_incomplete" };
    }
    if (!unit.current_run_evidence_outputs.every((relative) => relative.startsWith("rows/") || relative.startsWith("unit-results/"))) {
      return { eligible: false, reason: "fresh_outputs_incomplete" };
    }
    if (unit.reusable_artifact_outputs.some((output) => output.producer_identity !== unit.unit_id)) {
      return { eligible: false, reason: "artifact_producer_incomplete" };
    }
    const profile = this.targetProfiles.get(targetForUnit(unit));
    if (!profile || profile.policy !== unit.cache_policy) return { eligible: false, reason: "profile_missing" };
    if (profile.requires_vulnerability_database_revision && !this.vulnerabilityDatabaseRevision) {
      return { eligible: false, reason: "freshness_unknown", profile };
    }
    const contextKey = `${profile.profile_id}:${unit.unit_id}:${unit.semantic_digest}`;
    if (this.contexts.has(contextKey)) return this.contexts.get(contextKey);
    const dependencyClosure = this.dependencyClosure(profile, unit);
    const inputDigest = semanticJSONDigest({
      cache_schema_id: cacheRecordSchemaID,
      profile_id: profile.profile_id,
      dependency_strategy: dependencyClosure.strategy,
      unit_digest: unit.semantic_digest,
      inputs: dependencyClosure.digest,
      platform: `${process.platform}/${process.arch}`,
      toolchain_digest: this.toolchainDigest,
      helper_digest: this.helperDigest,
      vulnerability_database_revision: profile.requires_vulnerability_database_revision ? this.vulnerabilityDatabaseRevision : "",
    });
    const context = { eligible: true, profile, inputDigest, dependencyClosure };
    this.contexts.set(contextKey, context);
    return context;
  }

  validateGraph(graph) {
    for (const unit of graph.units) {
      if (unit.cache_policy === "none") continue;
      const profile = this.targetProfiles.get(targetForUnit(unit));
      if (!profile || profile.policy !== unit.cache_policy) {
        throw new Error(`work unit ${unit.unit_id} has no matching registered cache profile`);
      }
    }
    return graph;
  }

  entryDirectory(profile, inputDigest) {
    return path.join(this.cacheRoot, safeSegment(profile.profile_id), inputDigest.slice("sha256:".length));
  }

  readEntry(directory, unit, context) {
    assertContainedDirectory(this.root, directory, "cache entry");
    const directoryInfo = lstatSync(directory);
    if (!directoryInfo.isDirectory() || directoryInfo.isSymbolicLink()) throw new CacheEntryError("cache entry is not a regular directory");
    if (artifactMode(directoryInfo) !== "0700") throw new CacheEntryError("cache entry mode mismatch");
    const recordPath = path.join(directory, "record.json");
    const recordInfo = lstatSync(recordPath);
    assertRegularFile(recordInfo, "cache record");
    if (artifactMode(recordInfo) !== "0600") throw new CacheEntryError("cache record mode mismatch");
    const record = JSON.parse(readFileSync(recordPath, "utf8"));
    validateSchemaSync(record.schema_id, record);
    if (
      record.schema_id !== cacheRecordSchemaID ||
      record.profile_id !== context.profile.profile_id ||
      record.policy !== unit.cache_policy ||
      record.producer_identity !== unit.unit_id ||
      record.unit_id !== unit.unit_id ||
      record.unit_digest !== unit.semantic_digest ||
      record.input_digest !== context.inputDigest ||
      (context.profile.requires_vulnerability_database_revision && record.vulnerability_database_revision !== this.vulnerabilityDatabaseRevision)
    ) {
      throw new CacheEntryError("cache record identity mismatch");
    }
    validateSemanticResult(unit, record.semantic_result);
    const artifactsRoot = path.join(directory, "artifacts");
    const artifactsInfo = lstatSync(artifactsRoot);
    if (!artifactsInfo.isDirectory() || artifactsInfo.isSymbolicLink() || artifactMode(artifactsInfo) !== "0700") {
      throw new CacheEntryError("cache artifact root is unsafe");
    }
    const tree = listEntryTree(artifactsRoot);
    const expectedFiles = [];
    const expectedDirectories = record.artifacts.map((_artifact, index) => String(index));
    const snapshots = record.artifacts.map((manifest, artifactIndex) => {
      if (manifest.artifact_type === "file") {
        const payloadRelative = `${artifactIndex}/0`;
        expectedFiles.push(payloadRelative);
        return { manifest, bytes: readFileSync(path.join(artifactsRoot, payloadRelative)), members: [] };
      }
      const members = manifest.members.map((member, memberIndex) => {
        if (member.artifact_type === "directory") return member;
        const payloadRelative = `${artifactIndex}/${memberIndex}`;
        expectedFiles.push(payloadRelative);
        return { ...member, bytes: readFileSync(path.join(artifactsRoot, payloadRelative)) };
      });
      return { manifest, members };
    });
    if (
      JSON.stringify(tree.files) !== JSON.stringify(expectedFiles.sort(compareASCII)) ||
      JSON.stringify(tree.directories) !== JSON.stringify(expectedDirectories.sort(compareASCII))
    ) {
      throw new CacheEntryError("cache entry payload set is partial or surplus");
    }
    validateManifest(unit, context.inputDigest, snapshots);
    if (outputDigest(snapshots, record.semantic_result) !== record.output_digest) {
      throw new CacheEntryError("cache output digest mismatch");
    }
    return { record, snapshots };
  }

  quarantine(directory, profile) {
    if (!existsSync(directory)) return;
    try {
      assertContainedDirectory(this.root, path.dirname(directory), "cache entry parent");
    } catch {
      return;
    }
    const quarantineRoot = ensureContainedDirectory(
      this.root,
      path.join(this.cacheRoot, ".quarantine", safeSegment(profile.profile_id)),
      "cache quarantine root",
    );
    const destination = path.join(quarantineRoot, `${path.basename(directory)}-${randomUUID()}`);
    try {
      renameSync(directory, destination);
    } catch (error) {
      if (!new Set(["ENOENT", "EEXIST", "ENOTEMPTY"]).has(error?.code)) throw error;
    }
  }

  restore(unit, snapshots, semanticResult) {
    publishTransaction(
      [...artifactPublishItems(snapshots), ...semanticEvidenceItems(unit, semanticResult, this.clock)],
      this.root,
      this.runRoot,
    );
  }

  async lookup(unit) {
    const context = this.context(unit);
    const profileID = context.profile?.profile_id;
    if (!context.eligible) return { outcome: "bypass", reason: context.reason, profile_id: profileID };
    if (this.mode === "off") return { outcome: "bypass", reason: "mode_off", profile_id: profileID };
    if (this.mode === "cold") return { outcome: "miss", reason: "cold_read_bypass", profile_id: profileID, write_eligible: true };
    if (unit.cache_policy === "same_run") {
      const snapshot = this.sameRun.get(context.inputDigest);
      if (!snapshot) return { outcome: "miss", reason: "record_missing", profile_id: profileID, write_eligible: true };
      try {
        validateManifest(unit, context.inputDigest, snapshot.snapshots);
        this.restore(unit, snapshot.snapshots, snapshot.semantic_result);
        return { outcome: "hit", reason: "record_valid", profile_id: profileID, output_digest: snapshot.output_digest };
      } catch {
        return { outcome: "miss", reason: "record_invalid", profile_id: profileID, write_eligible: true };
      }
    }
    const directory = this.entryDirectory(context.profile, context.inputDigest);
    if (!existsSync(directory)) return { outcome: "miss", reason: "record_missing", profile_id: profileID, write_eligible: true };
    let entry;
    try {
      entry = this.readEntry(directory, unit, context);
    } catch {
      this.quarantine(directory, context.profile);
      return { outcome: "miss", reason: "record_invalid", profile_id: profileID, write_eligible: true };
    }
    try {
      this.restore(unit, entry.snapshots, entry.record.semantic_result);
      return { outcome: "hit", reason: "record_valid", profile_id: profileID, output_digest: entry.record.output_digest };
    } catch {
      return { outcome: "miss", reason: "restore_rejected", profile_id: profileID, write_eligible: true };
    }
  }

  writeStagingEntry(staging, record, snapshots) {
    mkdirSync(path.join(staging, "artifacts"), { recursive: true, mode: 0o700 });
    chmodSync(path.join(staging, "artifacts"), 0o700);
    snapshots.forEach((snapshot, artifactIndex) => {
      const artifactRoot = path.join(staging, "artifacts", String(artifactIndex));
      mkdirSync(artifactRoot, { mode: 0o700 });
      chmodSync(artifactRoot, 0o700);
      if (snapshot.manifest.artifact_type === "file") {
        const payload = path.join(artifactRoot, "0");
        writeFileSync(payload, snapshot.bytes, { mode: 0o600 });
        chmodSync(payload, 0o600);
        return;
      }
      snapshot.members.forEach((member, memberIndex) => {
        if (member.artifact_type === "file") {
          const payload = path.join(artifactRoot, String(memberIndex));
          writeFileSync(payload, member.bytes, { mode: 0o600 });
          chmodSync(payload, 0o600);
        }
      });
    });
    writeFileSync(path.join(staging, "record.json"), `${JSON.stringify(record, null, 2)}\n`, { mode: 0o600 });
    chmodSync(path.join(staging, "record.json"), 0o600);
  }

  publishEntry(directory, staging, unit, context) {
    ensureContainedDirectory(this.root, path.dirname(directory), "cache profile root");
    for (let attempt = 0; attempt < 2; attempt += 1) {
      try {
        renameSync(staging, directory);
        return { stored: true, reason: "stored" };
      } catch (error) {
        if (!new Set(["EEXIST", "ENOTEMPTY"]).has(error?.code)) throw error;
        try {
          const existing = this.readEntry(directory, unit, context);
          rmSync(staging, { recursive: true, force: true });
          return {
            stored: true,
            reason: "concurrent_entry",
            output_digest: existing.record.output_digest,
          };
        } catch {
          this.quarantine(directory, context.profile);
        }
      }
    }
    throw new Error("cache entry publication did not converge");
  }

  async store(unit) {
    const context = this.context(unit);
    if (!context.eligible || this.mode === "off") return { stored: false, reason: context.reason ?? "mode_off" };
    let snapshots;
    let semanticResult;
    try {
      snapshots = snapshotArtifacts(unit, this.root, this.runRoot, context.inputDigest);
      semanticResult = semanticResultForUnit(unit, this.runRoot);
      validateSemanticResult(unit, semanticResult);
      validateManifest(unit, context.inputDigest, snapshots);
    } catch {
      return { stored: false, reason: "output_missing" };
    }
    const digest = outputDigest(snapshots, semanticResult);
    if (unit.cache_policy === "same_run") {
      this.sameRun.set(context.inputDigest, { snapshots, semantic_result: semanticResult, output_digest: digest });
      return { stored: true, output_digest: digest };
    }
    const directory = this.entryDirectory(context.profile, context.inputDigest);
    const stagingRoot = ensureContainedDirectory(
      this.root,
      path.join(this.cacheRoot, ".staging", safeSegment(context.profile.profile_id)),
      "cache staging root",
    );
    const staging = path.join(stagingRoot, `${context.inputDigest.slice("sha256:".length)}-${randomUUID()}`);
    mkdirSync(staging, { mode: 0o700 });
    chmodSync(staging, 0o700);
    const record = {
      schema_id: cacheRecordSchemaID,
      profile_id: context.profile.profile_id,
      policy: unit.cache_policy,
      producer_identity: unit.unit_id,
      unit_id: unit.unit_id,
      unit_digest: unit.semantic_digest,
      input_digest: context.inputDigest,
      semantic_result: semanticResult,
      output_digest: digest,
      artifacts: snapshots.map((snapshot) => snapshot.manifest),
      ...(context.profile.requires_vulnerability_database_revision ? { vulnerability_database_revision: this.vulnerabilityDatabaseRevision } : {}),
      created_at: this.clock().toISOString(),
    };
    try {
      validateSchemaSync(record.schema_id, record);
      this.writeStagingEntry(staging, record, snapshots);
      this.readEntry(staging, unit, context);
      const published = this.publishEntry(directory, staging, unit, context);
      return {
        ...published,
        output_digest: published.output_digest ?? digest,
      };
    } catch (error) {
      rmSync(staging, { recursive: true, force: true });
      throw error;
    }
  }
}
