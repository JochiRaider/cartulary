import { parseExtensionContractArtifact } from "../services/extensionContractAdapter";

const supportRegistryPath =
  "contracts/extensions/generated/client-support-registry.json";
const tokenPattern = /^[a-z][a-z0-9_]{0,63}$/u;
const schemaIDPattern = /^[a-z][a-z0-9_.-]{0,159}$/u;
const digestPattern = /^[0-9a-f]{64}$/u;
const maximumGeneration = 18_446_744_073_709_551_615n;

export type ExtensionAvailabilityTag = {
  readonly epochId: string;
  readonly generation: bigint;
};

export type ExtensionWorkspaceIdentity = {
  readonly extensionProfileId: string;
  readonly workspaceKey: string;
};

export type ExtensionDiscoveryProfile = {
  readonly profile_id: string;
  readonly claimed: boolean;
  readonly contract_major: number | null;
  readonly route_families: readonly string[];
  readonly workspace_keys: readonly string[];
  readonly capabilities: readonly string[];
};

export type ClientExtensionSupportRegistry = {
  readonly schema_id: "cartulary.client_extension_support_registry.v1";
  readonly client_build_id: string;
  readonly client_build_class: "standard";
  readonly asset_set_sha256: string;
  readonly profiles: readonly {
    readonly profile_id: string;
    readonly supported_contract_majors: readonly [number];
    readonly workspace_keys: readonly string[];
    readonly capability_ids: readonly [];
    readonly public_schema_ids: readonly string[];
  }[];
};

export class ExtensionAvailabilityUnavailableError extends Error {
  readonly code = "extension_workspace_unavailable";

  constructor() {
    super("Extension workspace availability is not current.");
    this.name = "AbortError";
  }
}

function exactKeys(value: Record<string, unknown>, keys: readonly string[]) {
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  return (
    actual.length === expected.length &&
    actual.every((key, index) => key === expected[index])
  );
}

function sortedUniqueTokens(value: unknown): string[] | null {
  if (!Array.isArray(value)) {
    return null;
  }
  const tokens: string[] = [];
  for (const item of value) {
    if (
      typeof item !== "string" ||
      !tokenPattern.test(item) ||
      (tokens.length > 0 && (tokens[tokens.length - 1] ?? "") >= item)
    ) {
      return null;
    }
    tokens.push(item);
  }
  return tokens;
}

function sortedUniqueSchemaIDs(value: unknown): string[] | null {
  if (!Array.isArray(value)) {
    return null;
  }
  const schemaIDs: string[] = [];
  for (const item of value) {
    if (
      typeof item !== "string" ||
      !schemaIDPattern.test(item) ||
      (schemaIDs.length > 0 && (schemaIDs[schemaIDs.length - 1] ?? "") >= item)
    ) {
      return null;
    }
    schemaIDs.push(item);
  }
  return schemaIDs;
}

export function decodeClientExtensionSupportRegistry(
  value: unknown,
): ClientExtensionSupportRegistry | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return null;
  }
  const registry = value as Record<string, unknown>;
  if (
    !exactKeys(registry, [
      "schema_id",
      "client_build_id",
      "client_build_class",
      "asset_set_sha256",
      "profiles",
    ]) ||
    registry.schema_id !== "cartulary.client_extension_support_registry.v1" ||
    typeof registry.client_build_id !== "string" ||
    !/^[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}$/u.test(registry.client_build_id) ||
    registry.client_build_class !== "standard" ||
    typeof registry.asset_set_sha256 !== "string" ||
    !digestPattern.test(registry.asset_set_sha256) ||
    !Array.isArray(registry.profiles)
  ) {
    return null;
  }
  let previousProfileId = "";
  for (const rawProfile of registry.profiles) {
    if (
      !rawProfile ||
      typeof rawProfile !== "object" ||
      Array.isArray(rawProfile)
    ) {
      return null;
    }
    const profile = rawProfile as Record<string, unknown>;
    const workspaceKeys = sortedUniqueTokens(profile.workspace_keys);
    const publicSchemaIds = sortedUniqueSchemaIDs(profile.public_schema_ids);
    if (
      !exactKeys(profile, [
        "profile_id",
        "supported_contract_majors",
        "workspace_keys",
        "capability_ids",
        "public_schema_ids",
      ]) ||
      typeof profile.profile_id !== "string" ||
      !tokenPattern.test(profile.profile_id) ||
      previousProfileId >= profile.profile_id ||
      !Array.isArray(profile.supported_contract_majors) ||
      profile.supported_contract_majors.length !== 1 ||
      !Number.isSafeInteger(profile.supported_contract_majors[0]) ||
      Number(profile.supported_contract_majors[0]) < 1 ||
      workspaceKeys === null ||
      publicSchemaIds === null ||
      !Array.isArray(profile.capability_ids) ||
      profile.capability_ids.length !== 0
    ) {
      return null;
    }
    previousProfileId = profile.profile_id;
  }
  return registry as ClientExtensionSupportRegistry;
}

export function packagedClientExtensionSupportRegistry(): ClientExtensionSupportRegistry | null {
  const bootstrap = browserClientSupportRegistry();
  if (bootstrap.present) {
    return decodeClientExtensionSupportRegistry(bootstrap.value);
  }
  return decodeClientExtensionSupportRegistry(
    parseExtensionContractArtifact<unknown>(supportRegistryPath),
  );
}

function browserClientSupportRegistry(): {
  readonly present: boolean;
  readonly value: unknown;
} {
  if (typeof document === "undefined") {
    return { present: false, value: null };
  }
  const element = document.getElementById(
    "cartulary-client-extension-support-registry",
  );
  if (element === null) {
    return { present: false, value: null };
  }
  try {
    return { present: true, value: JSON.parse(element.textContent ?? "") };
  } catch {
    return { present: true, value: null };
  }
}

function workspaceIdentityKey(identity: ExtensionWorkspaceIdentity) {
  return `${identity.extensionProfileId}\u0000${identity.workspaceKey}`;
}

function discoveryProfilesEqual(
  left: readonly ExtensionDiscoveryProfile[],
  right: readonly ExtensionDiscoveryProfile[],
) {
  if (left.length !== right.length) {
    return false;
  }
  return left.every((profile, index) => {
    const candidate = right[index];
    return (
      candidate !== undefined &&
      profile.profile_id === candidate.profile_id &&
      profile.claimed === candidate.claimed &&
      profile.contract_major === candidate.contract_major &&
      profile.route_families.length === candidate.route_families.length &&
      profile.route_families.every(
        (route, routeIndex) => route === candidate.route_families[routeIndex],
      ) &&
      profile.workspace_keys.length === candidate.workspace_keys.length &&
      profile.workspace_keys.every(
        (workspace, workspaceIndex) =>
          workspace === candidate.workspace_keys[workspaceIndex],
      ) &&
      profile.capabilities.length === candidate.capabilities.length &&
      profile.capabilities.every(
        (capability, capabilityIndex) =>
          capability === candidate.capabilities[capabilityIndex],
      )
    );
  });
}

function randomEpochId(randomValues: (bytes: Uint8Array) => Uint8Array) {
  const bytes = randomValues(new Uint8Array(32));
  if (bytes.byteLength !== 32) {
    throw new Error("secure random source returned the wrong byte count");
  }
  return Array.from(bytes, (value) => value.toString(16).padStart(2, "0")).join(
    "",
  );
}

export function decodeExtensionWorkspaceAvailability(
  value: unknown,
  incidentId: string,
): readonly ExtensionWorkspaceIdentity[] | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return null;
  }
  const availability = value as Record<string, unknown>;
  if (
    !exactKeys(availability, ["schema_id", "incident_id", "workspaces"]) ||
    availability.schema_id !==
      "cartulary.extension_workspace_availability.v1" ||
    availability.incident_id !== incidentId ||
    !Array.isArray(availability.workspaces) ||
    availability.workspaces.length > 64
  ) {
    return null;
  }
  const rows: ExtensionWorkspaceIdentity[] = [];
  let previous = "";
  for (const rawRow of availability.workspaces) {
    if (!rawRow || typeof rawRow !== "object" || Array.isArray(rawRow)) {
      return null;
    }
    const row = rawRow as Record<string, unknown>;
    if (
      !exactKeys(row, ["extension_profile_id", "workspace_key"]) ||
      typeof row.extension_profile_id !== "string" ||
      !tokenPattern.test(row.extension_profile_id) ||
      typeof row.workspace_key !== "string" ||
      !tokenPattern.test(row.workspace_key)
    ) {
      return null;
    }
    const identity = {
      extensionProfileId: row.extension_profile_id,
      workspaceKey: row.workspace_key,
    };
    const key = workspaceIdentityKey(identity);
    if (previous >= key) {
      return null;
    }
    previous = key;
    rows.push(identity);
  }
  return rows;
}

export class ExtensionAvailabilityController {
  readonly clientInstanceId: string;
  readonly incidentId: string;
  readonly support: ClientExtensionSupportRegistry | null;
  #availability = new Set<string>();
  #discovery: readonly ExtensionDiscoveryProfile[] = [];
  #epochId = "";
  #generation = 0n;
  #enabled = true;
  #requestTail: Promise<void> = Promise.resolve();
  readonly #randomValues: (bytes: Uint8Array) => Uint8Array;

  constructor(options: {
    readonly clientInstanceId?: string;
    readonly incidentId: string;
    readonly support?: ClientExtensionSupportRegistry | null;
    readonly randomValues?: (bytes: Uint8Array) => Uint8Array;
    readonly initialGeneration?: bigint;
  }) {
    this.clientInstanceId = options.clientInstanceId ?? "unbound-client";
    this.incidentId = options.incidentId;
    this.support =
      options.support === undefined
        ? packagedClientExtensionSupportRegistry()
        : options.support;
    this.#randomValues =
      options.randomValues ??
      ((bytes) => {
        const generated = globalThis.crypto.getRandomValues(
          new Uint8Array(new ArrayBuffer(bytes.byteLength)),
        );
        bytes.set(generated);
        return bytes;
      });
    this.#generation = options.initialGeneration ?? 0n;
    try {
      this.#epochId = randomEpochId(this.#randomValues);
    } catch {
      this.#enabled = false;
    }
    if (this.support === null) {
      this.#enabled = false;
    }
  }

  setDiscovery(profiles: readonly ExtensionDiscoveryProfile[]) {
    const next = profiles.map((profile) => ({
      ...profile,
      route_families: [...profile.route_families],
      workspace_keys: [...profile.workspace_keys],
      capabilities: [...profile.capabilities],
    }));
    const changed = !discoveryProfilesEqual(this.#discovery, next);
    this.#discovery = next;
    if (changed) {
      this.invalidate();
    }
  }

  reserve(): ExtensionAvailabilityTag | null {
    if (!this.#enabled) {
      return null;
    }
    if (this.#generation === maximumGeneration) {
      try {
        this.#epochId = randomEpochId(this.#randomValues);
      } catch {
        this.#enabled = false;
        this.#availability.clear();
        return null;
      }
      this.#generation = 1n;
      this.#availability.clear();
    } else {
      this.#generation += 1n;
    }
    return { epochId: this.#epochId, generation: this.#generation };
  }

  isCurrent(tag: ExtensionAvailabilityTag): boolean {
    return (
      this.#enabled &&
      tag.epochId === this.#epochId &&
      tag.generation === this.#generation
    );
  }

  acceptWorkbookStartup(
    tag: ExtensionAvailabilityTag,
    availability: unknown,
  ): boolean {
    if (!this.isCurrent(tag)) {
      return false;
    }
    const rows = decodeExtensionWorkspaceAvailability(
      availability,
      this.incidentId,
    );
    this.#availability.clear();
    if (rows === null) {
      return false;
    }
    for (const row of rows) {
      this.#availability.add(workspaceIdentityKey(row));
    }
    return true;
  }

  acceptWorkbookStartupWorkspaces(
    tag: ExtensionAvailabilityTag,
    workspaces: readonly ExtensionWorkspaceIdentity[],
  ): boolean {
    if (!this.isCurrent(tag)) {
      return false;
    }
    this.#availability.clear();
    for (const workspace of workspaces) {
      this.#availability.add(workspaceIdentityKey(workspace));
    }
    return true;
  }

  invalidate(): ExtensionAvailabilityTag | null {
    this.#availability.clear();
    return this.reserve();
  }

  isRenderable(identity: ExtensionWorkspaceIdentity): boolean {
    if (!this.#enabled || this.support === null) {
      return false;
    }
    const discovery = this.#discovery.find(
      (profile) => profile.profile_id === identity.extensionProfileId,
    );
    const support = this.support.profiles.find(
      (profile) => profile.profile_id === identity.extensionProfileId,
    );
    return Boolean(
      discovery?.claimed &&
        discovery.capabilities.length === 0 &&
        discovery.workspace_keys.includes(identity.workspaceKey) &&
        support &&
        support.capability_ids.length === 0 &&
        discovery.contract_major === support.supported_contract_majors[0] &&
        support.workspace_keys.includes(identity.workspaceKey) &&
        this.#availability.has(workspaceIdentityKey(identity)),
    );
  }

  renderableWorkspaces(): readonly ExtensionWorkspaceIdentity[] {
    if (this.support === null) {
      return [];
    }
    const rows: ExtensionWorkspaceIdentity[] = [];
    for (const profile of this.support.profiles) {
      for (const workspaceKey of profile.workspace_keys) {
        const identity = {
          extensionProfileId: profile.profile_id,
          workspaceKey,
        };
        if (this.isRenderable(identity)) {
          rows.push(identity);
        }
      }
    }
    return rows;
  }

  isRouteAvailable(extensionProfileId: string, routeFamily: string): boolean {
    if (!this.#enabled || this.support === null) {
      return false;
    }
    const discovery = this.#discovery.find(
      (profile) => profile.profile_id === extensionProfileId,
    );
    const support = this.support.profiles.find(
      (profile) => profile.profile_id === extensionProfileId,
    );
    return Boolean(
      discovery?.claimed &&
        discovery.capabilities.length === 0 &&
        discovery.route_families.includes(routeFamily) &&
        support &&
        support.capability_ids.length === 0 &&
        discovery.contract_major === support.supported_contract_majors[0],
    );
  }

  async runProfileRequest<T>(
    extensionProfileId: string,
    routeFamily: string,
    request: () => Promise<T>,
  ): Promise<T> {
    let release: () => void = () => undefined;
    const previous = this.#requestTail;
    this.#requestTail = new Promise<void>((resolve) => {
      release = resolve;
    });
    await previous;
    try {
      if (!this.isRouteAvailable(extensionProfileId, routeFamily)) {
        throw new ExtensionAvailabilityUnavailableError();
      }
      const tag = this.reserve();
      if (tag === null) {
        throw new ExtensionAvailabilityUnavailableError();
      }
      const result = await request();
      if (
        !this.isCurrent(tag) ||
        !this.isRouteAvailable(extensionProfileId, routeFamily)
      ) {
        throw new ExtensionAvailabilityUnavailableError();
      }
      return result;
    } finally {
      release();
    }
  }

  async runRequest<T>(request: () => Promise<T>): Promise<T> {
    let release: () => void = () => undefined;
    const previous = this.#requestTail;
    this.#requestTail = new Promise<void>((resolve) => {
      release = resolve;
    });
    await previous;
    try {
      if (this.#availability.size === 0) {
        throw new ExtensionAvailabilityUnavailableError();
      }
      const tag = this.reserve();
      if (tag === null) {
        throw new ExtensionAvailabilityUnavailableError();
      }
      const result = await request();
      if (!this.isCurrent(tag)) {
        throw new ExtensionAvailabilityUnavailableError();
      }
      return result;
    } finally {
      release();
    }
  }

  currentTag(): ExtensionAvailabilityTag | null {
    if (!this.#enabled) {
      return null;
    }
    return { epochId: this.#epochId, generation: this.#generation };
  }
}

export function extensionCapabilityActivationFailure() {
  return Object.freeze({ code: "extension_capability_not_supported" });
}
