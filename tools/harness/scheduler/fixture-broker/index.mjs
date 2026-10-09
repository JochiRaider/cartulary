import { validateSchemaSync } from "../../contract/index.mjs";
import { CommandFailure } from "../../runtime/command-failure.mjs";
import { aggregateCleanup, CleanupResults } from "./cleanup-lifecycle.mjs";
export { aggregateCleanup, cleanupFailure, CleanupResults, createSuiteController } from "./cleanup-lifecycle.mjs";

export {
  productionFixtureProviders,
  startManagedSuite,
} from "./providers.mjs";

const capabilities = new Set([
  "none",
  "postgres_transaction",
  "postgres_dedicated",
  "postgres_migration",
  "object_store_namespace",
  "managed_process",
  "browser_stack",
]);

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

export class DedicatedResourcePool {
  constructor({ create, reset, healthy, destroy, targetSize = 1 }) {
    if (![create, reset, healthy, destroy].every((entry) => typeof entry === "function")) {
      throw new Error("dedicated pool requires create, reset, healthy, and destroy functions");
    }
    if (!Number.isInteger(targetSize) || targetSize < 1) {
      throw new Error("dedicated pool targetSize must be a positive integer");
    }
    this.create = create;
    this.reset = reset;
    this.healthy = healthy;
    this.destroy = destroy;
    this.targetSize = targetSize;
    this.ready = [];
    this.leased = new Set();
    this.pending = new Set();
    this.releasePromises = new Map();
    this.destroyPromises = new Map();
    this.failures = [];
    this.closePromise = null;
    this.closed = false;
  }

  replenish() {
    if (this.closed) return;
    while (this.ready.length + this.pending.size < this.targetSize) {
      const pending = Promise.resolve()
        .then(() => this.create())
        .then(async (resource) => {
          if (this.closed) await this.destroyResource(resource);
          else this.ready.push(resource);
        })
        .finally(() => this.pending.delete(pending));
      this.pending.add(pending);
      pending.catch((error) => this.failures.push(error));
    }
  }

  async warm() {
    this.replenish();
    await Promise.all(this.pending);
  }

  async acquire() {
    if (this.closed) throw new Error("dedicated pool is closed");
    if (this.ready.length === 0) await this.warm();
    const resource = this.ready.shift();
    if (resource === undefined) throw new Error("dedicated pool failed to replenish");
    this.leased.add(resource);
    this.releasePromises.delete(resource);
    this.replenish();
    return resource;
  }

  release(resource, { healthy = true } = {}) {
    if (this.releasePromises.has(resource)) return this.releasePromises.get(resource);
    const promise = this.releaseOnce(resource, healthy);
    this.releasePromises.set(resource, promise);
    return promise;
  }

  async releaseOnce(resource, healthy) {
    if (!this.leased.has(resource)) throw new Error("dedicated resource is not leased");
    let reusable = healthy && !this.closed;
    if (reusable) {
      try {
        reusable = (await this.healthy(resource)) === true;
        if (reusable) await this.reset(resource);
      } catch {
        reusable = false;
      }
    }
    if (reusable && !this.closed && this.ready.length < this.targetSize) this.ready.push(resource);
    else await this.destroyResource(resource);
    this.leased.delete(resource);
    this.replenish();
  }

  destroyResource(resource) {
    if (!this.destroyPromises.has(resource)) this.destroyPromises.set(resource, Promise.resolve().then(() => this.destroy(resource)));
    return this.destroyPromises.get(resource);
  }

  close() {
    if (this.closePromise) return this.closePromise;
    this.closed = true;
    this.closePromise = this.closeOnce();
    return this.closePromise;
  }

  async closeOnce() {
    const pending = await Promise.allSettled([...this.pending, ...this.releasePromises.values()]);
    const destroyed = await Promise.allSettled(this.ready.splice(0).map((resource) => this.destroyResource(resource)));
    const errors = [...this.failures, ...[...pending, ...destroyed].filter((result) => result.status === "rejected").map((result) => result.reason)];
    if (this.leased.size) errors.push(new Error("dedicated pool still has unresolved leased resources"));
    const failure = aggregateCleanup(errors);
    if (failure) throw failure;
  }
}

export class DigestPoolRegistry {
  constructor(createPool) {
    if (typeof createPool !== "function") throw new Error("digest pool registry requires a factory");
    this.createPool = createPool;
    this.pools = new Map();
    this.closePromise = null;
  }

  pool(digest) {
    if (this.closePromise) throw new Error("digest pool registry is closed");
    if (!/^sha256:[a-f0-9]{64}$/u.test(digest)) {
      throw new Error("migrated template digest must be sha256");
    }
    if (!this.pools.has(digest)) this.pools.set(digest, this.createPool(digest));
    return this.pools.get(digest);
  }

  close() {
    this.closePromise ??= Promise.resolve().then(async () => {
      const results = await Promise.allSettled([...this.pools.values()].map((pool) => pool.close()));
      const failure = aggregateCleanup(results.filter((result) => result.status === "rejected").map((result) => result.reason));
      if (failure) throw failure;
      this.pools.clear();
    });
    return this.closePromise;
  }
}

export function dedicatedPoolProvider(pool, resourceID = (resource) => String(resource.id)) {
  return {
    async acquire() {
      const resource = await pool.acquire();
      return {
        ownership: "owned",
        resource,
        resource_ids: [resourceID(resource)],
        release: ({ healthy }) => pool.release(resource, { healthy }),
      };
    },
    close: () => pool.close(),
  };
}

class FixtureLease {
  constructor(broker, record, entry, unitID) {
    this.broker = broker;
    this.record = record;
    this.entry = entry;
    this.resource = entry.allocation.resource;
    this.unitID = unitID;
    this.releasePromise = null;
  }

  release({ healthy = true, retainWarm = false } = {}) {
    if (typeof healthy !== "boolean" || typeof retainWarm !== "boolean") {
      return Promise.reject(new Error("fixture release disposition must be boolean"));
    }
    if (this.releasePromise) {
      if (this.disposition.healthy !== healthy || this.disposition.retainWarm !== retainWarm) {
        return Promise.reject(new Error("contradictory fixture release disposition"));
      }
      return this.releasePromise;
    }
    this.disposition = { healthy, retainWarm };
    this.releasePromise = this.broker.release(this, this.disposition);
    return this.releasePromise;
  }
}

export class FixtureBroker {
  constructor({ providers = {}, clock = () => new Date(), idFactory, recordSink = () => {}, observeLease = () => {}, cleanupResults = new CleanupResults() } = {}) {
    this.providers = providers;
    this.clock = clock;
    this.nextID = 1;
    this.idFactory = idFactory ?? (() => `lease-${String(this.nextID++).padStart(6, "0")}`);
    if (typeof recordSink !== "function") throw new Error("fixture broker recordSink must be a function");
    this.recordSink = recordSink;
    this.observeLease = observeLease;
    this.cleanupResults = cleanupResults;
    this.active = new Set();
    this.allocations = new Set();
    this.acquisitions = new Set();
    this.shared = new Map();
    this.closed = false;
    this.closePromise = null;
    this.providerCleanupFailed = false;
    this.acquisitionCleanupErrors = [];
  }

  acquire(capability, options = {}) {
    const promise = this.acquireOnce(capability, options);
    this.acquisitions.add(promise);
    promise.then(() => this.acquisitions.delete(promise), () => this.acquisitions.delete(promise));
    return promise;
  }

  async acquireOnce(capability, { affinityKey, unitID = "unit", digest, browserStage, runtimeProfileID,
    fixtureProfileID, snapshotKey, builderUnitID, rowID, predicateID } = {}) {
    if (this.closed) throw new Error("fixture broker is closed");
    if (!capabilities.has(capability)) throw new Error(`unknown fixture capability ${capability}`);
    const sharedKey = capability === "browser_stack"
      ? `${capability}:${affinityKey ?? unitID}:${fixtureProfileID ?? "none"}:${snapshotKey ?? "none"}` : "";
    const leaseID = this.idFactory();
    let entry = sharedKey ? this.shared.get(sharedKey) : null;
    let lease;
    try {
      if (!entry) {
        const provider = this.providers[capability];
        if (capability !== "none" && typeof provider?.acquire !== "function") {
          throw new Error(`no provider for fixture capability ${capability}`);
        }
        const allocation = capability === "none"
          ? { ownership: "borrowed", resource_ids: [], resource: null }
          : await provider.acquire({ affinityKey, unitID, digest, browserStage, runtimeProfileID,
              fixtureProfileID, snapshotKey, builderUnitID, rowID, predicateID, leaseID });
        if (!allocation || !["owned", "borrowed"].includes(allocation.ownership)) {
          throw new Error(`${capability} provider returned invalid ownership`);
        }
        // Register ownership before validation or publication can throw.
        entry = { allocation, capability, sharedKey, references: 0, tainted: false,
          cleanupPromise: null, cleanupOutcome: "not_required", releaseErrors: [], lastLease: null, leaseID, unitID };
        this.allocations.add(entry);
        if (capability !== "none") {
          const hook = allocation.ownership === "owned" ? "release" : "detach";
          if (typeof allocation[hook] !== "function") throw new Error(`${capability} provider requires ${hook}`);
        }
        allocation.resource_ids = [...(allocation.resource_ids ?? [])].sort(compareASCII);
        if (new Set(allocation.resource_ids).size !== allocation.resource_ids.length) {
          throw new Error(`${capability} provider returned duplicate resource IDs`);
        }
        if (sharedKey) this.shared.set(sharedKey, entry);
      }
      entry.references += 1;
      const allocation = entry.allocation;
      const record = {
        schema_id: "cartulary.harness_fixture_lease.v4", lease_id: leaseID, capability,
        ownership: allocation.ownership, state: "leased", cleanup_outcome: "not_required",
        cleanup_failure_reason: null, resource_ids: allocation.resource_ids,
        ...(affinityKey ? { affinity_key: affinityKey } : {}),
        ...(allocation.fixture_profile_id ? { fixture_profile_id: allocation.fixture_profile_id } : {}),
        ...(allocation.snapshot_key ? { snapshot_key: allocation.snapshot_key } : {}),
        ...(allocation.builder_unit_id ? { builder_unit_id: allocation.builder_unit_id } : {}),
        ...(allocation.clone_ordinal ? { clone_ordinal: allocation.clone_ordinal } : {}),
        created_at: this.clock().toISOString(),
      };
      lease = new FixtureLease(this, record, entry, unitID);
      entry.lastLease = lease;
      this.active.add(lease);
      await this.publish(lease);
      try { this.observeLease({ allocation_ref: `allocation:${entry.leaseID}`, lease_ref: leaseID,
        unit_id: unitID, capability, ownership: allocation.ownership }); }
      catch { /* Optional correlation never changes fixture ownership or outcome. */ }
      if (this.closed) throw new Error("fixture broker closed during acquisition");
      return lease;
    } catch (error) {
      if (entry) {
        try {
          if (lease) await lease.release({ healthy: false });
          else { entry.tainted = true; this.unshare(entry); await this.cleanupAllocation(entry); }
        } catch (cleanupError) { (error.cleanupFailures ??= []).push(cleanupError); }
      } else if (error.cleanupFailures?.length) {
        const failure = aggregateCleanup(error.cleanupFailures);
        this.acquisitionCleanupErrors.push(failure);
        this.providerCleanupFailed = true;
        this.cleanupResults.record("fixture_acquisition_cleanup", { unitID, leaseID, error: failure });
      }
      throw error;
    }
  }

  async publish(lease) {
    try {
      validateSchemaSync(lease.record.schema_id, lease.record);
      if (["released", "destroyed"].includes(lease.record.state) && lease.record.cleanup_outcome === "failed") {
        throw new Error("settled fixture cannot claim failed cleanup");
      }
      await this.recordSink(lease.record);
      lease.published = true;
    } catch (error) {
      throw new CommandFailure("fixture lease publication failed", {
        failure_class: "artifact", failure_reason: "artifact_error",
      }, { cause: error });
    }
  }

  unshare(entry) {
    if (this.shared.get(entry.sharedKey) === entry) this.shared.delete(entry.sharedKey);
  }

  cleanupAllocation(entry) {
    if (entry.cleanupPromise) return entry.cleanupPromise;
    entry.cleanupPromise = Promise.resolve().then(async () => {
      const lease = entry.lastLease;
      const errors = [...entry.releaseErrors];
      entry.cleanupOutcome = entry.capability === "none" ? "not_required" : "pending";
      if (lease && entry.capability !== "none") {
        lease.record.state = entry.tainted ? "quarantined" : "leased";
        lease.record.cleanup_outcome = "pending";
        try { await this.publish(lease); } catch (error) { errors.push(error); }
      }
      if (entry.capability !== "none") {
        let physicalFailed = false;
        try {
          if (entry.allocation.ownership === "borrowed") await entry.allocation.detach();
          else await entry.allocation.release({ healthy: !entry.tainted });
        } catch (error) { physicalFailed = true; errors.push(error); }
        entry.cleanupOutcome = physicalFailed ? "failed" : "completed";

      }
      if (lease) {
        lease.record.state = entry.cleanupOutcome === "failed"
          ? entry.tainted ? "quarantined" : "failed" : entry.tainted ? "destroyed" : "released";
        lease.record.cleanup_outcome = entry.cleanupOutcome;
        lease.record.cleanup_failure_reason = entry.cleanupOutcome === "failed" ? "cleanup_error" : null;
        try { await this.publish(lease); } catch (error) { errors.push(error); }
      }
      const failure = aggregateCleanup(errors);
      if (entry.capability !== "none") this.cleanupResults.record(entry.allocation.ownership === "borrowed" ? "fixture_detach" : "fixture_release", {
        unitID: lease?.unitID ?? entry.unitID, leaseID: lease?.record.lease_id ?? entry.leaseID, error: failure,
        artifactRefs: lease?.published ? [`_shared/fixture-leases/${lease.record.lease_id}.json`] : [],
      });
      if (failure) throw failure;
    });
    return entry.cleanupPromise;
  }

  async release(lease, { healthy, retainWarm }) {
    const entry = lease.entry;
    entry.references -= 1;
    entry.tainted ||= !healthy;
    if (entry.tainted) this.unshare(entry);
    try {
      const retained = entry.references === 0 && !entry.tainted && retainWarm &&
        entry.capability === "browser_stack" && !this.closed;
      if (entry.references === 0 && !retained) {
        this.unshare(entry);
        await this.cleanupAllocation(entry);
      } else {
        lease.record.state = entry.tainted ? "quarantined" : "released";
        lease.record.cleanup_outcome = "not_required";
        await this.publish(lease);
      }
      return { retained };
    } catch (error) {
      entry.tainted = true;
      this.unshare(entry);
      if (!entry.cleanupPromise) entry.releaseErrors.push(error);
      if (entry.references === 0 && !entry.cleanupPromise) {
        try { await this.cleanupAllocation(entry); } catch (cleanupError) { (error.cleanupFailures ??= []).push(cleanupError); }
      }
      throw error;
    } finally { this.active.delete(lease); }
  }

  hasUnresolvedCleanup() {
    return this.providerCleanupFailed || [...this.allocations].some((entry) =>
      entry.capability !== "none" && entry.cleanupOutcome !== "completed");
  }

  close() {
    if (this.closePromise) return this.closePromise;
    this.closed = true;
    this.closePromise = this.closeOnce();
    return this.closePromise;
  }

  async closeOnce() {
    await Promise.allSettled([...this.acquisitions]);
    const errors = [...this.acquisitionCleanupErrors];
    for (const lease of [...this.active].reverse()) {
      try { await (lease.releasePromise ?? lease.release()); } catch (error) { errors.push(error); }
    }
    for (const entry of [...this.allocations].reverse()) {
      this.unshare(entry);
      try { await this.cleanupAllocation(entry); } catch (error) { errors.push(error); }
    }
    for (const provider of new Set(Object.values(this.providers))) {
      if (typeof provider.close !== "function") continue;
      const error = await this.cleanupResults.attempt("provider_close", () => provider.close());
      if (error) { this.providerCleanupFailed = true; errors.push(error); }
    }
    const failure = aggregateCleanup(errors);
    if (failure) throw failure;
  }
}
