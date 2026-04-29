import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..");
const defaultRegistryPath = path.join(repoRoot, "tools", "scheduler_resource_registry.json");
const registrySchemaID = "cartulary.scheduler_resource_registry.v1";

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

function loadRawRegistry(file = defaultRegistryPath) {
  const registry = JSON.parse(readFileSync(file, "utf8"));
  if (registry.schema_id !== registrySchemaID) {
    throw new Error(`${file} must declare schema_id ${registrySchemaID}`);
  }
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
      autoLimit: resource.auto_limit === true,
      envVariable: typeof resource.env_variable === "string" && resource.env_variable ? resource.env_variable : null,
      makeVariable: typeof resource.make_variable === "string" && resource.make_variable ? resource.make_variable : null,
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
  const retiredAliases = new Map();
  for (const alias of raw.retired_aliases ?? []) {
    retiredAliases.set(requireString(alias.name, "retired_aliases[].name"), {
      replacement: typeof alias.replacement === "string" && alias.replacement ? alias.replacement : null,
      reason: typeof alias.reason === "string" ? alias.reason : "",
    });
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
  return { resources, templates, retiredAliases, forwardingProfiles };
}

export function schedulerResourceRegistry() {
  cachedRegistry ??= buildRegistry();
  return cachedRegistry;
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

export function retiredResourceAlias(resource) {
  return schedulerResourceRegistry().retiredAliases.get(resource) ?? null;
}

export function resourceDescriptor(resource) {
  const registry = schedulerResourceRegistry();
  const descriptor = registry.resources.get(resource);
  if (descriptor) {
    return descriptor;
  }
  if (isBrowserStageResource(resource)) {
    const template = browserStageTemplate();
    return {
      name: resource,
      displayName: `${template.displayName} ${resource.slice(template.prefix.length)}`,
      schedulers: template.schedulers,
      displayOrder: template.displayOrder,
      autoLimit: false,
      envVariable: null,
      makeVariable: null,
      template: template.name,
    };
  }
  return null;
}

export function assertKnownResource(resource, label, { scheduler = null } = {}) {
  const normalized = requireString(resource, label);
  const retired = retiredResourceAlias(normalized);
  if (retired) {
    const replacement = retired.replacement ? `; use ${retired.replacement}` : "";
    throw new Error(`${label} uses retired resource ${normalized}${replacement}`);
  }
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
  return resourceDescriptor(resource)?.autoLimit === true;
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

export function normalizeResourceLimits(value, label, { scheduler, overrides = new Map(), allowAuto = false } = {}) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} resource_limits must be an object`);
  }
  const limits = new Map();
  const sources = new Map();
  for (const [resource, rawLimit] of Object.entries(value)) {
    const normalizedResource = assertKnownResource(resource.trim(), `${label} resource_limits.${resource}`, { scheduler });
    if (rawLimit === "auto") {
      if (!allowAuto || !isAutoLimitResource(normalizedResource)) {
        throw new Error(`${label} resource_limits.${normalizedResource} may not use "auto"`);
      }
      limits.set(normalizedResource, rawLimit);
      sources.set(normalizedResource, "manifest:auto");
      continue;
    }
    if (!Number.isInteger(rawLimit) || rawLimit < 1) {
      throw new Error(`${label} resource_limits.${normalizedResource} must be a positive integer${allowAuto ? ' or "auto"' : ""}`);
    }
    limits.set(normalizedResource, rawLimit);
    sources.set(normalizedResource, "manifest");
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
    sources.set(normalizedResource, "override");
  }
  return { limits, sources };
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
    if (amount > resourceLimits.get(normalizedResource)) {
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
