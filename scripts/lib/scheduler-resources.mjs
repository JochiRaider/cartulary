import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..");
const defaultRegistryPath = path.join(repoRoot, "tools", "scheduler_resource_registry.json");
const registrySchemaID = "cartulary.scheduler_resource_registry.v3";
const envVariablePattern = /^[A-Z][A-Z0-9_]*$/;
const registryTopLevelKeys = new Set([
  "schema_id",
  "resources",
  "templates",
  "capacity_profiles",
  "forwarding_profiles",
]);

let cachedRegistry = null;

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function requireString(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  return value.trim();
}

function requireStringArray(value, label) {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  const seen = new Set();
  return value.map((entry, index) => {
    const normalized = requireString(entry, `${label}[${index + 1}]`);
    if (seen.has(normalized)) {
      throw new Error(`${label} contains duplicate ${normalized}`);
    }
    seen.add(normalized);
    return normalized;
  });
}

function requirePositiveInteger(value, label) {
  if (!Number.isInteger(value) || value < 1) {
    throw new Error(`${label} must be a positive integer`);
  }
  return value;
}

function requirePlainObject(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value;
}

function assertKnownObjectKeys(value, label, allowedKeys) {
  for (const key of Object.keys(value)) {
    if (!allowedKeys.has(key)) {
      throw new Error(`${label} has unknown key ${key}`);
    }
  }
}

function normalizeOverrideEnv(value, label) {
  if (value === undefined) {
    return null;
  }
  const normalized = requireString(value, label);
  if (!envVariablePattern.test(normalized)) {
    throw new Error(`${label} must be a safe environment variable name`);
  }
  return normalized;
}

function normalizeCapacity(value, label) {
  if (value === undefined) {
    return null;
  }
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label}.capacity must be an object`);
  }
  const allowedKeys = new Set(["default_limit", "auto_policy", "override_env"]);
  for (const key of Object.keys(value)) {
    if (!allowedKeys.has(key)) {
      throw new Error(`${label}.capacity has unknown key ${key}`);
    }
  }
  const hasDefault = Object.hasOwn(value, "default_limit");
  const hasAuto = Object.hasOwn(value, "auto_policy");
  if (hasDefault === hasAuto) {
    throw new Error(`${label}.capacity must declare exactly one of default_limit or auto_policy`);
  }
  return {
    defaultLimit: hasDefault
      ? requirePositiveInteger(value.default_limit, `${label}.capacity.default_limit`)
      : null,
    autoPolicy: hasAuto ? requireString(value.auto_policy, `${label}.capacity.auto_policy`) : null,
    overrideEnv: normalizeOverrideEnv(value.override_env, `${label}.capacity.override_env`),
  };
}

function loadRawRegistry(file = defaultRegistryPath) {
  const registry = requirePlainObject(JSON.parse(readFileSync(file, "utf8")), "scheduler resource registry");
  if (registry.schema_id !== registrySchemaID) {
    throw new Error(`${file} must declare schema_id ${registrySchemaID}`);
  }
  assertKnownObjectKeys(registry, "scheduler resource registry", registryTopLevelKeys);
  return registry;
}

function buildRegistry(file = defaultRegistryPath) {
  const raw = loadRawRegistry(file);
  const resources = new Map();
  for (const [index, resource] of (raw.resources ?? []).entries()) {
    const label = `resources[${index + 1}]`;
    const name = requireString(resource?.name, `${label}.name`);
    if (resources.has(name)) {
      throw new Error(`scheduler resource registry declares duplicate resource ${name}`);
    }
    resources.set(name, {
      name,
      displayName: requireString(resource.display_name, `${label}.display_name`),
      schedulers: new Set((resource.schedulers ?? []).map((entry) => requireString(entry, `${label}.schedulers[]`))),
      displayOrder: Number.isFinite(resource.display_order) ? resource.display_order : 1000,
      capacity: normalizeCapacity(resource.capacity, label),
    });
  }
  const templates = new Map();
  for (const [index, template] of (raw.templates ?? []).entries()) {
    const label = `templates[${index + 1}]`;
    const name = requireString(template?.name, `${label}.name`);
    templates.set(name, {
      name,
      prefix: requireString(template.prefix, `${label}.prefix`),
      displayName: requireString(template.display_name, `${label}.display_name`),
      schedulers: new Set((template.schedulers ?? []).map((entry) => requireString(entry, `${label}.schedulers[]`))),
      displayOrder: Number.isFinite(template.display_order) ? template.display_order : 1000,
    });
  }
  const capacityProfiles = new Map();
  for (const [index, profile] of (raw.capacity_profiles ?? []).entries()) {
    const label = `capacity_profiles[${index + 1}]`;
    const name = requireString(profile?.name, `${label}.name`);
    if (capacityProfiles.has(name)) {
      throw new Error(`scheduler resource registry declares duplicate capacity profile ${name}`);
    }
    const scheduler = requireString(profile.scheduler, `${label}.scheduler`);
    const profileResources = requireStringArray(profile.resources, `${label}.resources`);
    for (const resource of profileResources) {
      const descriptor = resources.get(resource);
      if (!descriptor) {
        throw new Error(`${label}.resources references unknown resource ${resource}`);
      }
      if (!descriptor.schedulers.has(scheduler)) {
        throw new Error(`${label}.resources ${resource} is not valid for ${scheduler} scheduler`);
      }
      if (!descriptor.capacity) {
        throw new Error(`${label}.resources ${resource} does not declare capacity metadata`);
      }
    }
    capacityProfiles.set(name, { name, scheduler, resources: profileResources });
  }
  const forwardingProfiles = new Map();
  for (const [index, profile] of (raw.forwarding_profiles ?? []).entries()) {
    const label = `forwarding_profiles[${index + 1}]`;
    const name = requireString(profile?.name, `${label}.name`);
    const mappings = (profile.mappings ?? []).map((mapping, mappingIndex) => ({
      sourceResource: requireString(mapping.source_resource, `${label}.mappings[${mappingIndex + 1}].source_resource`),
      targetResource: requireString(mapping.target_resource, `${label}.mappings[${mappingIndex + 1}].target_resource`),
      envVariable: requireString(mapping.env_variable, `${label}.mappings[${mappingIndex + 1}].env_variable`),
    }));
    forwardingProfiles.set(name, { name, mappings });
  }
  return { resources, templates, capacityProfiles, forwardingProfiles };
}

export function schedulerResourceRegistry() {
  cachedRegistry ??= buildRegistry();
  return cachedRegistry;
}

export function loadSchedulerResourceRegistry(file = defaultRegistryPath) {
  return buildRegistry(file);
}

function browserStageTemplate() {
  return schedulerResourceRegistry().templates.get("browser_stage");
}

export function browserStageResource(stageName) {
  const normalized = requireString(stageName, "browser stage name").replaceAll("-", "_");
  return `${browserStageTemplate().prefix}${normalized}`;
}

export function isBrowserStageResource(resource) {
  const template = browserStageTemplate();
  return typeof resource === "string" && resource.startsWith(template.prefix) && resource.length > template.prefix.length;
}

export function resourceDescriptor(resource) {
  const registry = schedulerResourceRegistry();
  const descriptor = registry.resources.get(resource);
  if (descriptor) {
    return {
      ...descriptor,
      autoLimit: descriptor.capacity?.autoPolicy !== null && descriptor.capacity?.autoPolicy !== undefined,
      autoPolicy: descriptor.capacity?.autoPolicy ?? null,
      defaultLimit: descriptor.capacity?.defaultLimit ?? null,
      overrideEnv: descriptor.capacity?.overrideEnv ?? null,
      envVariable: descriptor.capacity?.overrideEnv ?? null,
      makeVariable: null,
    };
  }
  if (isBrowserStageResource(resource)) {
    const template = browserStageTemplate();
    return {
      name: resource,
      displayName: `${template.displayName} ${resource.slice(template.prefix.length)}`,
      schedulers: template.schedulers,
      displayOrder: template.displayOrder,
      capacity: null,
      autoLimit: false,
      autoPolicy: null,
      defaultLimit: null,
      overrideEnv: null,
      envVariable: null,
      makeVariable: null,
      template: template.name,
    };
  }
  return null;
}

export function assertKnownResource(resource, label, { scheduler = null } = {}) {
  const normalized = requireString(resource, label);
  const descriptor = resourceDescriptor(normalized);
  if (!descriptor) {
    throw new Error(`${label} uses undeclared scheduler resource ${normalized}`);
  }
  if (scheduler && !descriptor.schedulers.has(scheduler)) {
    throw new Error(`${label} resource ${normalized} is not valid for ${scheduler} scheduler`);
  }
  return normalized;
}

export function isAutoLimitResource(resource) {
  const descriptor = resourceDescriptor(resource);
  return descriptor ? descriptor.autoPolicy !== null : false;
}

function resourceSortKey(resource) {
  const descriptor = resourceDescriptor(resource);
  return [descriptor?.displayOrder ?? 1000, resource];
}

export function compareResources(left, right) {
  const [leftOrder, leftName] = resourceSortKey(left);
  const [rightOrder, rightName] = resourceSortKey(right);
  return leftOrder - rightOrder || compareStrings(leftName, rightName);
}

export function resourceMapToObject(values) {
  return Object.fromEntries(Array.from(values.entries()).sort((left, right) => compareResources(left[0], right[0])));
}

export function formatResourceMap(values) {
  const entries = Array.from(values.entries()).sort((left, right) => compareResources(left[0], right[0]));
  if (entries.length === 0) {
    return "{}";
  }
  return `{${entries.map(([key, value]) => `${key}:${value}`).join(",")}}`;
}

export function resourceLimitSummary(resourceLimits, preferred = []) {
  const seen = new Set();
  const entries = [];
  for (const resource of preferred) {
    if (resourceLimits.has(resource)) {
      entries.push(`${resource}:${resourceLimits.get(resource)}`);
      seen.add(resource);
    }
  }
  for (const [resource, value] of Array.from(resourceLimits.entries()).sort((left, right) => compareResources(left[0], right[0]))) {
    if (!seen.has(resource)) {
      entries.push(`${resource}:${value}`);
    }
  }
  return entries.join(",");
}

export function preferredResourcesForScheduler(scheduler) {
  const resources = [];
  for (const descriptor of schedulerResourceRegistry().resources.values()) {
    if (descriptor.schedulers.has(scheduler)) {
      resources.push(descriptor.name);
    }
  }
  return resources.sort(compareResources);
}

function validateRawLimit(rawLimit, label, resource, { allowAuto }) {
  const descriptor = resourceDescriptor(resource);
  if (rawLimit === "auto") {
    if (!allowAuto || !descriptor?.autoPolicy) {
      throw new Error(`${label} resource_limits.${resource} may not use "auto"`);
    }
    return rawLimit;
  }
  if (!Number.isInteger(rawLimit) || rawLimit < 1) {
    throw new Error(`${label} resource_limits.${resource} must be a positive integer${allowAuto ? ' or "auto"' : ""}`);
  }
  return rawLimit;
}

function limitFromCapacityDescriptor(descriptor, label, { allowAuto }) {
  if (!descriptor.capacity) {
    throw new Error(`${label} resource ${descriptor.name} does not declare capacity metadata`);
  }
  if (descriptor.capacity.autoPolicy) {
    return validateRawLimit("auto", label, descriptor.name, { allowAuto });
  }
  return validateRawLimit(descriptor.capacity.defaultLimit, label, descriptor.name, { allowAuto });
}

export function resourceLimitsForCapacityProfile(name, label, { scheduler, allowAuto = false } = {}) {
  const profileName = requireString(name, `${label}.capacity_profile`);
  const profile = schedulerResourceRegistry().capacityProfiles.get(profileName);
  if (!profile) {
    throw new Error(`${label}.capacity_profile references unknown capacity profile ${profileName}`);
  }
  if (scheduler && profile.scheduler !== scheduler) {
    throw new Error(`${label}.capacity_profile ${profileName} is not valid for ${scheduler} scheduler`);
  }
  const limits = new Map();
  const sources = new Map();
  for (const resource of profile.resources) {
    const descriptor = resourceDescriptor(resource);
    limits.set(resource, limitFromCapacityDescriptor(descriptor, label, { allowAuto }));
    sources.set(resource, `registry:${profileName}`);
  }
  return { limits, sources, profile };
}

function parsePositiveIntegerEnv(env, name) {
  if (!env || !name) {
    return null;
  }
  const raw = env[name];
  if (raw === undefined || raw === "") {
    return null;
  }
  const amount = Number(raw);
  if (!Number.isInteger(amount) || amount < 1) {
    throw new Error(`${name} must be a positive integer`);
  }
  return amount;
}

export function resourceOverrideEnvVariablesForScheduler(scheduler) {
  const envVariables = [];
  for (const descriptor of schedulerResourceRegistry().resources.values()) {
    if (descriptor.schedulers.has(scheduler) && descriptor.capacity?.overrideEnv) {
      envVariables.push([descriptor.name, descriptor.capacity.overrideEnv]);
    }
  }
  return envVariables
    .sort((left, right) => compareResources(left[0], right[0]))
    .map((entry) => entry[1]);
}

export function normalizeResourceLimits(value, label, { scheduler, capacityProfile = null, overrides = new Map(), allowAuto = false, env = null } = {}) {
  const limits = new Map();
  const sources = new Map();
  if (capacityProfile) {
    const profileLimits = resourceLimitsForCapacityProfile(capacityProfile, label, { scheduler, allowAuto });
    for (const [resource, limit] of profileLimits.limits.entries()) {
      limits.set(resource, limit);
    }
    for (const [resource, source] of profileLimits.sources.entries()) {
      sources.set(resource, source);
    }
  }
  if (!capacityProfile && (!value || typeof value !== "object" || Array.isArray(value))) {
    throw new Error(`${label} resource_limits must be an object`);
  }
  if (value !== undefined) {
    if (!value || typeof value !== "object" || Array.isArray(value)) {
      throw new Error(`${label} resource_limits must be an object`);
    }
    for (const [resource, rawLimit] of Object.entries(value)) {
      const normalizedResource = assertKnownResource(resource.trim(), `${label} resource_limits.${resource}`, { scheduler });
      const limit = validateRawLimit(rawLimit, label, normalizedResource, { allowAuto });
      if (sources.get(normalizedResource)?.startsWith("registry:")) {
        if (limits.get(normalizedResource) !== limit) {
          throw new Error(`${label} resource_limits.${normalizedResource} must match ${sources.get(normalizedResource)}`);
        }
        continue;
      }
      limits.set(normalizedResource, limit);
      sources.set(normalizedResource, "manifest");
    }
    if (capacityProfile) {
      for (const resource of limits.keys()) {
        assertKnownResource(resource, `${label} resource_limits.${resource}`, { scheduler });
      }
    }
  }
  for (const [resource, limit] of limits.entries()) {
    const descriptor = resourceDescriptor(resource);
    const envOverride = parsePositiveIntegerEnv(env, descriptor?.overrideEnv);
    if (envOverride !== null) {
      limits.set(resource, envOverride);
      sources.set(resource, `env:${descriptor.overrideEnv}`);
    } else if (limit === "auto" && !descriptor?.autoPolicy) {
      throw new Error(`${label} resource_limits.${resource} may not use "auto"`);
    }
  }
  for (const [resource, amount] of overrides.entries()) {
    const normalizedResource = assertKnownResource(resource.trim(), `${label} resource limit override ${resource}`, { scheduler });
    if (!limits.has(normalizedResource)) {
      throw new Error(`${label} resource limit override ${normalizedResource} is not declared`);
    }
    if (!Number.isInteger(amount) || amount < 1) {
      throw new Error(`${label} resource limit override ${normalizedResource} must be a positive integer`);
    }
    limits.set(normalizedResource, amount);
    sources.set(normalizedResource, "cli");
  }
  return { limits, sources };
}

export function resolveAutoResourceLimits(resourceLimits, resourceLimitSources, label, autoResolvers) {
  const resolved = new Map(resourceLimits.entries());
  const sources = new Map(resourceLimitSources.entries());
  const entries = Array.from(resolved.entries()).sort((left, right) => compareResources(left[0], right[0]));
  for (const [resource, limit] of entries) {
    if (limit !== "auto") {
      continue;
    }
    const descriptor = resourceDescriptor(resource);
    const policy = descriptor?.autoPolicy;
    if (!policy) {
      throw new Error(`${label} resource_limits.${resource} may not use "auto"`);
    }
    const resolver = autoResolvers?.[policy];
    if (typeof resolver !== "function") {
      throw new Error(`${label} resource_limits.${resource} references unknown auto policy ${policy}`);
    }
    const amount = resolver({ resource, resourceLimits: resolved });
    if (!Number.isInteger(amount) || amount < 1) {
      throw new Error(`${label} auto policy ${policy} for ${resource} must return a positive integer`);
    }
    resolved.set(resource, amount);
    sources.set(resource, `auto:${policy}`);
  }
  return { resourceLimits: resolved, resourceLimitSources: sources };
}

export function normalizeBoundedLimitClaim(value, label, resource, resourceLimit) {
  const allowedKeys = new Set(["mode", "reserve", "min", "max"]);
  for (const key of Object.keys(value)) {
    if (!allowedKeys.has(key)) {
      throw new Error(`${label} resource_claims.${resource} bounded_limit has unknown key ${key}`);
    }
  }
  if (value.mode !== "bounded_limit") {
    throw new Error(`${label} resource_claims.${resource}.mode must be bounded_limit`);
  }
  for (const key of ["reserve", "min", "max"]) {
    if (!Object.hasOwn(value, key)) {
      throw new Error(`${label} resource_claims.${resource} bounded_limit must declare ${key}`);
    }
  }
  if (!Number.isInteger(value.reserve) || value.reserve < 0) {
    throw new Error(`${label} resource_claims.${resource}.reserve must be a non-negative integer`);
  }
  if (!Number.isInteger(value.min) || value.min < 1) {
    throw new Error(`${label} resource_claims.${resource}.min must be a positive integer`);
  }
  if (!Number.isInteger(value.max) || value.max < 1) {
    throw new Error(`${label} resource_claims.${resource}.max must be a positive integer`);
  }
  if (value.max < value.min) {
    throw new Error(`${label} resource_claims.${resource}.max must be greater than or equal to min`);
  }
  return Math.min(resourceLimit, value.max, Math.max(value.min, resourceLimit - value.reserve));
}

export function normalizeResourceClaimAmount(rawAmount, label, resource, resourceLimit, { allowBounded = false } = {}) {
  if (rawAmount === "limit") {
    return resourceLimit;
  }
  if (Number.isInteger(rawAmount)) {
    return rawAmount;
  }
  if (allowBounded && rawAmount && typeof rawAmount === "object" && !Array.isArray(rawAmount)) {
    return normalizeBoundedLimitClaim(rawAmount, label, resource, resourceLimit);
  }
  return rawAmount;
}

export function normalizeResourceClaims(value, label, resourceLimits, { scheduler, allowBounded = false } = {}) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} resource_claims must be an object`);
  }
  const claims = new Map();
  for (const [resource, rawAmount] of Object.entries(value)) {
    const normalizedResource = assertKnownResource(resource.trim(), `${label} resource_claims.${resource}`, { scheduler });
    if (!resourceLimits.has(normalizedResource)) {
      throw new Error(`${label} resource_claims entry ${normalizedResource} is not declared in resource_limits`);
    }
    const amount = normalizeResourceClaimAmount(
      rawAmount,
      label,
      normalizedResource,
      resourceLimits.get(normalizedResource),
      { allowBounded },
    );
    if (!Number.isInteger(amount) || amount < 1) {
      throw new Error(`${label} resource_claims.${normalizedResource} must be a positive integer, "limit"${allowBounded ? ", or bounded_limit object" : ""}`);
    }
    const resourceLimit = resourceLimits.get(normalizedResource);
    if (Number.isInteger(resourceLimit) && amount > resourceLimit) {
      throw new Error(`${label} resource_claims.${normalizedResource} exceeds resource limit`);
    }
    claims.set(normalizedResource, amount);
  }
  return claims;
}

export function resolveForwardingProfile(name, resourceClaims, label) {
  const profileName = requireString(name, `${label}.forwarding`);
  const profile = schedulerResourceRegistry().forwardingProfiles.get(profileName);
  if (!profile) {
    throw new Error(`${label}.forwarding references unknown forwarding profile ${profileName}`);
  }
  const resourceLimitEnv = new Map();
  const forwardedResourceLimits = new Map();
  const forwardingMappings = [];
  const envNames = new Set();
  for (const mapping of profile.mappings) {
    if (!resourceClaims.has(mapping.sourceResource)) {
      throw new Error(`${label}.forwarding ${profileName} source ${mapping.sourceResource} must be claimed by work unit`);
    }
    if (envNames.has(mapping.envVariable)) {
      throw new Error(`${label}.forwarding ${profileName} maps multiple resources to ${mapping.envVariable}`);
    }
    envNames.add(mapping.envVariable);
    const amount = resourceClaims.get(mapping.sourceResource);
    resourceLimitEnv.set(mapping.envVariable, amount);
    forwardedResourceLimits.set(mapping.targetResource, amount);
    forwardingMappings.push({
      source_resource: mapping.sourceResource,
      target_resource: mapping.targetResource,
      env_variable: mapping.envVariable,
      amount,
    });
  }
  return {
    profile: profileName,
    resourceLimitEnv,
    forwardedResourceLimits,
    forwardingMappings,
  };
}

export function resourceLimitSourcesToObject(values) {
  return Object.fromEntries(Array.from(values.entries()).sort((left, right) => compareResources(left[0], right[0])));
}
