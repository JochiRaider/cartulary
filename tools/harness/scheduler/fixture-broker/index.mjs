import { validateSchemaSync } from "../../contract/index.mjs";

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
    this.closed = false;
  }

  replenish() {
    if (this.closed) return;
    while (this.ready.length + this.pending.size < this.targetSize) {
      const pending = Promise.resolve()
        .then(() => this.create())
        .then(async (resource) => {
          if (this.closed) await this.destroy(resource);
          else this.ready.push(resource);
        })
        .finally(() => this.pending.delete(pending));
      this.pending.add(pending);
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
    this.replenish();
    return resource;
  }

  async release(resource, { healthy = true } = {}) {
    if (!this.leased.delete(resource)) return;
    let reusable = healthy && !this.closed;
    if (reusable) {
      try {
        reusable = (await this.healthy(resource)) === true;
        if (reusable) await this.reset(resource);
      } catch {
        reusable = false;
      }
    }
    if (reusable && this.ready.length < this.targetSize) this.ready.push(resource);
    else await this.destroy(resource);
    this.replenish();
  }

  async close() {
    if (this.closed) return;
    this.closed = true;
    await Promise.all(this.pending);
    const ready = this.ready.splice(0);
    await Promise.all(ready.map((resource) => this.destroy(resource)));
  }
}

export class DigestPoolRegistry {
  constructor(createPool) {
    if (typeof createPool !== "function") throw new Error("digest pool registry requires a factory");
    this.createPool = createPool;
    this.pools = new Map();
  }

  pool(digest) {
    if (!/^sha256:[a-f0-9]{64}$/u.test(digest)) {
      throw new Error("migrated template digest must be sha256");
    }
    if (!this.pools.has(digest)) this.pools.set(digest, this.createPool(digest));
    return this.pools.get(digest);
  }

  async close() {
    await Promise.all([...this.pools.values()].map((pool) => pool.close()));
    this.pools.clear();
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
        release: () => pool.release(resource, { healthy: true }),
        quarantine: () => pool.release(resource, { healthy: false }),
      };
    },
    close: () => pool.close(),
  };
}

class FixtureLease {
  constructor(broker, record, allocation, sharedKey) {
    this.broker = broker;
    this.record = record;
    this.allocation = allocation;
    this.sharedKey = sharedKey;
    this.resource = allocation?.resource;
    this.released = false;
  }

  async release({ healthy = true, retainWarm = false } = {}) {
    if (this.released) return { retained: false };
    this.released = true;
    return this.broker.release(this, { healthy, retainWarm });
  }

  async quarantine() {
    this.record.state = "quarantined";
    validateSchemaSync(this.record.schema_id, this.record);
    await this.broker.recordSink(this.record);
    await this.release({ healthy: false });
  }
}

export class FixtureBroker {
  constructor({ providers = {}, clock = () => new Date(), idFactory, recordSink = () => {} } = {}) {
    this.providers = providers;
    this.clock = clock;
    this.nextID = 1;
    this.idFactory = idFactory ?? (() => `lease-${String(this.nextID++).padStart(6, "0")}`);
    if (typeof recordSink !== "function") throw new Error("fixture broker recordSink must be a function");
    this.recordSink = recordSink;
    this.active = [];
    this.shared = new Map();
    this.closed = false;
  }

  async acquire(
    capability,
    {
      affinityKey,
      unitID = "unit",
      digest,
      runtimeProfileID,
      fixtureProfileID,
      snapshotKey,
      builderUnitID,
      rowID,
      predicateID,
    } = {},
  ) {
    if (this.closed) throw new Error("fixture broker is closed");
    if (!capabilities.has(capability)) throw new Error(`unknown fixture capability ${capability}`);
    const sharedKey = capability === "browser_stack"
        ? `${capability}:${affinityKey ?? unitID}:${fixtureProfileID ?? "none"}:${snapshotKey ?? "none"}`
        : "";
    const leaseID = this.idFactory();
    let shared = sharedKey ? this.shared.get(sharedKey) : null;
    let allocation;
    if (shared) {
      shared.references += 1;
      allocation = shared.allocation;
    } else if (capability === "none") {
      allocation = {
        ownership: "borrowed",
        resource_ids: [],
        resource: null,
      };
    } else {
      const provider = this.providers[capability];
      if (!provider?.acquire) throw new Error(`no provider for fixture capability ${capability}`);
      allocation = await provider.acquire({
        affinityKey,
        unitID,
        digest,
        runtimeProfileID,
        fixtureProfileID,
        snapshotKey,
        builderUnitID,
        rowID,
        predicateID,
        leaseID,
      });
      if (!allocation || !["owned", "borrowed"].includes(allocation.ownership)) {
        throw new Error(`${capability} provider returned invalid ownership`);
      }
      allocation.resource_ids = [...(allocation.resource_ids ?? [])].sort(compareASCII);
      if (new Set(allocation.resource_ids).size !== allocation.resource_ids.length) {
        throw new Error(`${capability} provider returned duplicate resource IDs`);
      }
      if (sharedKey) {
        shared = { allocation, references: 1 };
        this.shared.set(sharedKey, shared);
      }
    }
    const record = {
      schema_id: "cartulary.harness_fixture_lease.v3",
      lease_id: leaseID,
      capability,
      ownership: allocation.ownership,
      state: "leased",
      resource_ids: allocation.resource_ids,
      ...(affinityKey ? { affinity_key: affinityKey } : {}),
      ...(allocation.fixture_profile_id
        ? { fixture_profile_id: allocation.fixture_profile_id }
        : {}),
      ...(allocation.snapshot_key ? { snapshot_key: allocation.snapshot_key } : {}),
      ...(allocation.builder_unit_id
        ? { builder_unit_id: allocation.builder_unit_id }
        : {}),
      ...(allocation.clone_ordinal
        ? { clone_ordinal: allocation.clone_ordinal }
        : {}),
      created_at: this.clock().toISOString(),
    };
    validateSchemaSync(record.schema_id, record);
    await this.recordSink(record);
    const lease = new FixtureLease(this, record, allocation, sharedKey);
    this.active.push(lease);
    return lease;
  }

  async release(lease, { healthy, retainWarm = false }) {
    const index = this.active.indexOf(lease);
    if (index >= 0) this.active.splice(index, 1);
    let releaseAllocation = true;
    let retained = false;
    if (lease.sharedKey) {
      const shared = this.shared.get(lease.sharedKey);
      if (shared) {
        shared.references -= 1;
        releaseAllocation = shared.references === 0;
        if (
          releaseAllocation &&
          healthy &&
          retainWarm &&
          lease.record.capability === "browser_stack" &&
          !this.closed
        ) {
          // Browser stacks are run-scoped affinity resources. Keep a healthy
          // zero-reference allocation warm for later units in the same chain.
          releaseAllocation = false;
          retained = true;
        } else if (releaseAllocation) {
          this.shared.delete(lease.sharedKey);
        }
      }
    }
    if (releaseAllocation && lease.record.capability !== "none") {
      if (lease.allocation.ownership === "borrowed") {
        await lease.allocation.detach?.();
      } else if (healthy) {
        await lease.allocation.release?.();
      } else {
        await (lease.allocation.quarantine?.() ?? lease.allocation.destroy?.());
      }
    }
    lease.record.state = healthy
      ? "released"
      : lease.record.state === "quarantined"
        ? "quarantined"
        : "destroyed";
    validateSchemaSync(lease.record.schema_id, lease.record);
    await this.recordSink(lease.record);
    return { retained };
  }

  async close() {
    if (this.closed) return;
    this.closed = true;
    for (const lease of [...this.active].reverse()) await lease.release();
    for (const shared of this.shared.values()) {
      if (shared.allocation.ownership === "borrowed") {
        await shared.allocation.detach?.();
      } else {
        await shared.allocation.release?.();
      }
    }
    this.shared.clear();
    for (const provider of Object.values(this.providers)) await provider.close?.();
  }
}
