export const serviceDependencyIDs = Object.freeze(["object_store", "postgres"]);

export function requiredServicesForFixture(fixtureCapability) {
  if (fixtureCapability.startsWith("postgres_")) return ["postgres"];
  if (fixtureCapability === "object_store_namespace") return ["object_store"];
  if (new Set(["browser_stack", "managed_process"]).has(fixtureCapability)) {
    return ["object_store", "postgres"];
  }
  return [];
}

export function assertFixtureServiceDependencies(fixtureCapability, dependencies, label) {
  for (const service of requiredServicesForFixture(fixtureCapability)) {
    if (!dependencies.includes(service)) {
      throw new Error(`${label} fixture ${fixtureCapability} requires service dependency ${service}`);
    }
  }
}
