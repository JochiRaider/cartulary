import { createHash } from "node:crypto";
import {
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

import { validateSchemaSync } from "../../contract/index.mjs";
import { semanticJSONDigest } from "../../test-catalog/index.mjs";

const modes = new Set(["normal", "cold", "off"]);

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function sha256(bytes) {
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}

function safeSegment(value) {
  return value.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-");
}

function containedFile(root, relative) {
  if (path.isAbsolute(relative) || relative.split(/[\\/]/u).includes("..")) {
    throw new Error(`cache artifact path escapes its root: ${relative}`);
  }
  const resolvedRoot = path.resolve(root);
  const resolved = path.resolve(root, relative);
  if (resolved !== resolvedRoot && !resolved.startsWith(`${resolvedRoot}${path.sep}`)) {
    throw new Error(`cache artifact path escapes its root: ${relative}`);
  }
  return resolved;
}

function walkFiles(root, relative, output) {
  const absolute = containedFile(root, relative);
  if (!existsSync(absolute)) {
    output.push({ path: relative, missing: true });
    return;
  }
  const info = lstatSync(absolute);
  if (info.isSymbolicLink()) throw new Error(`cache input must not be a symlink: ${relative}`);
  if (info.isFile()) {
    output.push({ path: relative.replaceAll("\\", "/"), digest: sha256(readFileSync(absolute)) });
    return;
  }
  if (!info.isDirectory()) throw new Error(`cache input has unsupported type: ${relative}`);
  for (const name of readdirSync(absolute).sort(compareASCII)) {
    if (new Set([".cache", ".cartulary", "coverage", "dist", "node_modules", "test-results", "tmp"]).has(name)) {
      continue;
    }
    walkFiles(root, path.join(relative, name), output);
  }
}

export function cacheInputRootDigest(root, inputRoots) {
  const entries = [];
  for (const inputRoot of [...inputRoots].sort(compareASCII)) walkFiles(root, inputRoot, entries);
  return semanticJSONDigest(entries);
}

function artifactSnapshots(unit, runRoot) {
  return unit.evidence_outputs.map((relative) => {
    const file = containedFile(runRoot, relative);
    if (!existsSync(file)) throw new Error(`cache output is missing: ${relative}`);
    const info = lstatSync(file);
    if (!info.isFile() || info.isSymbolicLink()) {
      throw new Error(`cache output must be a regular file: ${relative}`);
    }
    const bytes = readFileSync(file);
    return { path: relative, digest: sha256(bytes), bytes };
  });
}

function outputDigest(snapshots) {
  return semanticJSONDigest(
    snapshots.map(({ path: artifactPath, digest }) => ({ path: artifactPath, digest })),
  );
}

function targetForUnit(unit) {
  if (unit.unit_id.startsWith("target:")) return unit.unit_id.slice("target:".length);
  return unit.command.environment.CARTULARY_TEST_TARGET ?? "";
}

export function loadCacheRegistry(root) {
  const registry = JSON.parse(
    readFileSync(path.join(root, "tools/harness_cache_registry.json"), "utf8"),
  );
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
  constructor({
    root,
    runRoot,
    cacheRoot,
    mode = "normal",
    registry,
    toolchainDigest,
    helperDigest,
    vulnerabilityDatabaseRevision = "",
    clock = () => new Date(),
  }) {
    if (!modes.has(mode)) throw new Error(`unsupported graph cache mode ${mode}`);
    this.root = path.resolve(root);
    this.runRoot = path.resolve(runRoot);
    this.cacheRoot = path.resolve(cacheRoot);
    this.mode = mode;
    this.toolchainDigest = toolchainDigest;
    this.helperDigest = helperDigest;
    this.vulnerabilityDatabaseRevision = vulnerabilityDatabaseRevision;
    this.clock = clock;
    const loaded = registry
      ? (() => {
          validateSchemaSync(registry.schema_id, registry);
          const profiles = new Map(registry.profiles.map((entry) => [entry.profile_id, entry]));
          const targetProfiles = new Map(
            registry.profiles.flatMap((entry) => entry.targets.map((target) => [target, entry])),
          );
          return { profiles, targetProfiles };
        })()
      : loadCacheRegistry(this.root);
    this.profiles = loaded.profiles;
    this.targetProfiles = loaded.targetProfiles;
    this.sameRun = new Map();
  }

  context(unit) {
    if (unit.cache_policy === "none") return { eligible: false, reason: "policy_none" };
    const profile = this.targetProfiles.get(targetForUnit(unit));
    if (!profile || profile.policy !== unit.cache_policy) {
      return { eligible: false, reason: "profile_missing" };
    }
    if (profile.requires_vulnerability_database_revision && !this.vulnerabilityDatabaseRevision) {
      return { eligible: false, reason: "freshness_unknown", profile };
    }
    const inputDigest = semanticJSONDigest({
      unit_digest: unit.semantic_digest,
      inputs: cacheInputRootDigest(this.root, profile.input_roots),
      toolchain_digest: this.toolchainDigest,
      helper_digest: this.helperDigest,
      vulnerability_database_revision: profile.requires_vulnerability_database_revision
        ? this.vulnerabilityDatabaseRevision
        : "",
    });
    return { eligible: true, profile, inputDigest };
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

  async lookup(unit) {
    const context = this.context(unit);
    const profileID = context.profile?.profile_id;
    if (!context.eligible) return { outcome: "bypass", reason: context.reason, profile_id: profileID };
    if (this.mode === "off") return { outcome: "bypass", reason: "mode_off", profile_id: profileID };
    if (this.mode === "cold") {
      return { outcome: "miss", reason: "cold_read_bypass", profile_id: profileID, write_eligible: true };
    }
    if (unit.cache_policy === "same_run") {
      const snapshot = this.sameRun.get(context.inputDigest);
      if (!snapshot) return { outcome: "miss", reason: "record_missing", profile_id: profileID, write_eligible: true };
      this.publish(snapshot.artifacts);
      return { outcome: "hit", reason: "record_valid", profile_id: profileID, output_digest: snapshot.output_digest };
    }
    const directory = this.entryDirectory(context.profile, context.inputDigest);
    const recordPath = path.join(directory, "record.json");
    if (!existsSync(recordPath)) {
      return { outcome: "miss", reason: "record_missing", profile_id: profileID, write_eligible: true };
    }
    try {
      const record = JSON.parse(readFileSync(recordPath, "utf8"));
      validateSchemaSync(record.schema_id, record);
      if (
        record.profile_id !== profileID ||
        record.policy !== unit.cache_policy ||
        record.unit_id !== unit.unit_id ||
        record.unit_digest !== unit.semantic_digest ||
        record.input_digest !== context.inputDigest ||
        (context.profile.requires_vulnerability_database_revision &&
          record.vulnerability_database_revision !== this.vulnerabilityDatabaseRevision)
      ) {
        throw new Error("cache record identity mismatch");
      }
      const snapshots = record.artifacts.map((artifact, index) => {
        const bytes = readFileSync(path.join(directory, "artifacts", String(index)));
        if (sha256(bytes) !== artifact.digest) throw new Error("cache artifact digest mismatch");
        return { ...artifact, bytes };
      });
      if (outputDigest(snapshots) !== record.output_digest) throw new Error("cache output digest mismatch");
      this.publish(snapshots);
      return { outcome: "hit", reason: "record_valid", profile_id: profileID, output_digest: record.output_digest };
    } catch {
      return { outcome: "miss", reason: "record_invalid", profile_id: profileID, write_eligible: true };
    }
  }

  publish(snapshots) {
    for (const snapshot of snapshots) {
      const output = containedFile(this.runRoot, snapshot.path);
      mkdirSync(path.dirname(output), { recursive: true });
      const temporary = `${output}.cache-${process.pid}`;
      writeFileSync(temporary, snapshot.bytes, { mode: 0o600 });
      renameSync(temporary, output);
    }
  }

  async store(unit) {
    const context = this.context(unit);
    if (!context.eligible || this.mode === "off") return { stored: false, reason: context.reason ?? "mode_off" };
    let snapshots;
    try {
      snapshots = artifactSnapshots(unit, this.runRoot);
    } catch {
      return { stored: false, reason: "output_missing" };
    }
    const digest = outputDigest(snapshots);
    if (unit.cache_policy === "same_run") {
      this.sameRun.set(context.inputDigest, { artifacts: snapshots, output_digest: digest });
      return { stored: true, output_digest: digest };
    }
    const directory = this.entryDirectory(context.profile, context.inputDigest);
    const temporary = `${directory}.tmp-${process.pid}-${Date.now()}`;
    rmSync(temporary, { recursive: true, force: true });
    mkdirSync(path.join(temporary, "artifacts"), { recursive: true });
    snapshots.forEach((snapshot, index) => {
      writeFileSync(path.join(temporary, "artifacts", String(index)), snapshot.bytes, { mode: 0o600 });
    });
    const record = {
      schema_id: "cartulary.harness_cache_record.v1",
      profile_id: context.profile.profile_id,
      policy: unit.cache_policy,
      unit_id: unit.unit_id,
      unit_digest: unit.semantic_digest,
      input_digest: context.inputDigest,
      output_digest: digest,
      artifacts: snapshots.map(({ path: artifactPath, digest: artifactDigest }) => ({
        path: artifactPath,
        digest: artifactDigest,
      })),
      ...(context.profile.requires_vulnerability_database_revision
        ? { vulnerability_database_revision: this.vulnerabilityDatabaseRevision }
        : {}),
      created_at: this.clock().toISOString(),
    };
    validateSchemaSync(record.schema_id, record);
    writeFileSync(path.join(temporary, "record.json"), `${JSON.stringify(record, null, 2)}\n`, { mode: 0o600 });
    mkdirSync(path.dirname(directory), { recursive: true });
    rmSync(directory, { recursive: true, force: true });
    renameSync(temporary, directory);
    return { stored: true, output_digest: digest };
  }
}
