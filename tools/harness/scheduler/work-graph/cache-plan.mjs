import { readFileSync } from "node:fs";
import path from "node:path";
import { semanticJSONDigest, validateSchemaSync } from "../../contract/index.mjs";
import { resolveCacheDependencyClosure } from "./cache-dependencies.mjs";
const modes = new Set(["normal", "cold", "off"]);
const cacheRecordSchemaID = "cartulary.harness_cache_record.v2";
const compareASCII = (a,b) => a < b ? -1 : a > b ? 1 : 0;

// Pure preparation facts. Go exclusively owns cache storage and transactions.
function expectedRowIDs(unit) {
  return unit.current_run_evidence_outputs
    .filter((relative) => relative.startsWith("rows/") && relative.endsWith(".json"))
    .map((relative) => relative.slice("rows/".length, -".json".length))
    .sort(compareASCII);
}

export function loadCacheRegistry(root) {
  const registry = JSON.parse(readFileSync(path.join(root, "tools/harness_cache_registry.json"), "utf8"));
  return indexCacheRegistry(registry);
}

function indexCacheRegistry(registry) {
  validateSchemaSync(registry.schema_id, registry);
  const profiles = new Map();
  const operationProfiles = new Map();
  const runnerProfiles = new Map();
  for (const profile of registry.profiles) {
    if (profiles.has(profile.profile_id)) throw new Error(`duplicate cache profile ${profile.profile_id}`);
    profiles.set(profile.profile_id, profile);
    for (const [keys, index] of [[profile.operation_ids, operationProfiles], [profile.runner_types, runnerProfiles]]) {
      for (const key of keys) {
        if (index.has(key)) throw new Error(`duplicate cache binding ${key}`);
        index.set(key, profile);
      }
    }
  }
  return { registry, profiles, operationProfiles, runnerProfiles };
}

export class CachePlan {
  constructor({ root, mode = "normal", registry, toolchainDigest, helperDigest, vulnerabilityDatabaseRevision = "", sourceEntries, getCatalog }) {
    if (!modes.has(mode)) throw new Error(`unsupported graph cache mode ${mode}`);
    this.root = path.resolve(root);
    this.mode = mode;
    this.toolchainDigest = toolchainDigest;
    this.helperDigest = helperDigest;
    this.vulnerabilityDatabaseRevision = vulnerabilityDatabaseRevision;
    this.sourceEntries = sourceEntries;
    let catalog;
    this.getCatalog = () => (catalog ??= getCatalog());
    const loaded = registry
      ? indexCacheRegistry(registry)
      : loadCacheRegistry(this.root);
    this.profiles = loaded.profiles;
    this.operationProfiles = loaded.operationProfiles;
    this.runnerProfiles = loaded.runnerProfiles;
    this.dependencyClosures = new Map();
    this.dependencyResolverCache = new Map();
    this.contexts = new Map();
  }

  profileForUnit(unit) {
    if (unit.unit_id.startsWith("operation:")) return this.operationProfiles.get(unit.unit_id.slice("operation:".length));
    const rows = expectedRowIDs(unit);
    if (rows.length === 0) return undefined;
    const catalog = this.getCatalog();
    const runners = new Set(rows.map((id) => catalog.rowByID.get(id)?.runner));
    if (runners.size !== 1 || runners.has(undefined)) throw new Error("cache unit has inconsistent row runners");
    return this.runnerProfiles.get([...runners][0]);
  }

  dependencyClosure(profile, unit) {
    const selector = unit.command.environment.CARTULARY_TEST_ROWS ?? unit.unit_id;
    const key = `${profile.profile_id}:${selector}`;
    if (!this.dependencyClosures.has(key)) {
      this.dependencyClosures.set(
        key,
        resolveCacheDependencyClosure({
              root: this.root,
              entries: this.sourceEntries,
              profile,
              unit,
              resolverCache: this.dependencyResolverCache,
              getCatalog: this.getCatalog,
            }),
      );
    }
    return this.dependencyClosures.get(key);
  }

  eligibility(unit) {
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
    const profile = this.profileForUnit(unit);
    if (!profile || profile.policy !== unit.cache_policy) return { eligible: false, reason: "profile_missing" };
    if (profile.requires_vulnerability_database_revision && !this.vulnerabilityDatabaseRevision) {
      return { eligible: false, reason: "freshness_unknown", profile };
    }
    return { eligible: true, profile };
  }

  context(unit) {
    return this.#keyedContext(unit, this.eligibility(unit));
  }

  #keyedContext(unit, eligibility) {
    if (!eligibility.eligible) return eligibility;
    const { profile } = eligibility;
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
      cryptographic_execution: { selector: process.env.GOFIPS140 ?? "", mode: process.env.GODEBUG ?? "", toolchain: process.env.GOTOOLCHAIN ?? "" },
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
      const profile = this.profileForUnit(unit);
      if (!profile || profile.policy !== unit.cache_policy) {
        throw new Error(`work unit ${unit.unit_id} has no matching registered cache profile`);
      }
    }
    return graph;
  }

}
