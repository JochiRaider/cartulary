const defaultSchedulerPriority = 0;

export function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

export function resourceClaimsObject(value) {
  return Object.fromEntries(
    Object.entries(value).sort(([left], [right]) => left.localeCompare(right)),
  );
}

export function addClaim(claims, resource, amount) {
  if (amount === "limit") {
    claims.set(resource, amount);
    return;
  }
  if (!Number.isInteger(amount) || amount < 1) {
    throw new Error(`resource claim ${resource} must be a positive integer or "limit"`);
  }
  if (claims.get(resource) === "limit") {
    return;
  }
  claims.set(resource, (claims.get(resource) ?? 0) + amount);
}

export function sortedUnique(values) {
  return Array.from(new Set(values.filter((value) => value !== ""))).sort((left, right) =>
    String(left).localeCompare(String(right)),
  );
}

export function priority(value) {
  return Number.isInteger(value) && value > 0 ? value : defaultSchedulerPriority;
}

export function command(type, extra = {}) {
  return { type, ...extra };
}

export function mergeEnv(...parts) {
  const entries = new Map();
  for (const part of parts) {
    for (const [name, value] of Object.entries(part ?? {})) {
      entries.set(name, value);
    }
  }
  return Object.fromEntries([...entries.entries()].sort(([left], [right]) => left.localeCompare(right)));
}

export function unionStrings(left, right) {
  return Array.from(new Set([...(left ?? []), ...(right ?? [])]));
}
