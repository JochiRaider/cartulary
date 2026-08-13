import {
  assertFixtureServiceDependencies,
  requiredServicesForFixture,
} from "../../test-catalog/service-dependencies.mjs";

export { assertFixtureServiceDependencies, requiredServicesForFixture };

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

export function assertServiceDependencies(topology, runtimeProfileID, dependencies, label) {
  const sorted = [...dependencies].sort(compareASCII);
  if (JSON.stringify(sorted) !== JSON.stringify(dependencies)) {
    throw new Error(`${label}.service_dependencies must be sorted`);
  }
  if (new Set(dependencies).size !== dependencies.length) {
    throw new Error(`${label}.service_dependencies contains a duplicate`);
  }
  const runtime = topology.runtime_profiles.find((entry) => entry.id === runtimeProfileID);
  if (!runtime) throw new Error(`${label} has unknown runtime profile ${runtimeProfileID}`);
  for (const dependency of dependencies) {
    if (!Object.hasOwn(topology.service_resource_minimums, dependency)) {
      throw new Error(`${label} has unknown service dependency ${dependency}`);
    }
    if (!runtime.managed_service_ids.includes(dependency)) {
      throw new Error(`${label} service dependency ${dependency} is unavailable in ${runtimeProfileID}`);
    }
  }
}

export function topologyResourceClaims(topology, resourceProfileID, dependencies, label) {
  const profile = topology.resource_profiles.find((entry) => entry.id === resourceProfileID);
  if (!profile) throw new Error(`${label} has unknown resource profile ${resourceProfileID}`);
  const claims = { ...profile.resource_claims };
  for (const dependency of dependencies) {
    const minimums = topology.service_resource_minimums[dependency];
    if (!minimums) throw new Error(`${label} has unknown service dependency ${dependency}`);
    for (const [resource, amount] of Object.entries(minimums)) {
      claims[resource] = Math.max(claims[resource] ?? 0, amount);
    }
  }
  return Object.fromEntries(
    Object.entries(claims).sort(([left], [right]) => compareASCII(left, right)),
  );
}
