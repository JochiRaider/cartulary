const executionDependencyDefinitions = Object.freeze([
  {
    id: "backend_unit",
    target: "backend-unit",
    category: "backend",
    order: 10,
    service_backed: false,
    support_target: true,
  },
  {
    id: "backend_store",
    target: "backend-store",
    category: "backend",
    order: 20,
    service_backed: true,
  },
  {
    id: "backend_integration",
    target: "backend-integration",
    category: "backend",
    order: 30,
    service_backed: true,
  },
  {
    id: "backend_integration_support",
    target: "backend-integration-support",
    category: "backend",
    order: 40,
    service_backed: true,
    support_target: true,
  },
  {
    id: "backend_process",
    target: "backend-process",
    category: "backend",
    order: 50,
    service_backed: true,
  },
  {
    id: "frontend_unit",
    target: "frontend-unit",
    category: "frontend",
    order: 60,
    service_backed: false,
  },
  {
    id: "browser_functional",
    target: "browser-e2e-webserver-backed",
    category: "browser",
    order: 70,
    service_backed: true,
  },
  {
    id: "browser_stateful",
    target: "browser-e2e-stateful",
    category: "browser",
    order: 80,
    service_backed: true,
  },
  {
    id: "browser_measurement",
    target: "browser-e2e-measurement",
    category: "browser",
    order: 90,
    service_backed: true,
  },
  {
    id: "browser_support",
    target: "browser-e2e-support",
    category: "browser",
    order: 100,
    service_backed: true,
  },
]);

export const executionDependencyMetadata = new Map(
  executionDependencyDefinitions.map((entry) => [entry.id, entry]),
);

export const validExecutionDependencies = new Set(executionDependencyMetadata.keys());

export const validSupportTargets = new Set(
  executionDependencyDefinitions
    .filter((entry) => entry.support_target)
    .map((entry) => entry.id),
);

export const serviceBackedGoExecutionDependencies = new Set(
  ["backend_store", "backend_integration", "backend_process"],
);

export const serviceBackedSupportTargets = new Set(
  executionDependencyDefinitions
    .filter((entry) => entry.support_target && entry.service_backed)
    .map((entry) => entry.id),
);

export function executionDependencyInfo(id) {
  return executionDependencyMetadata.get(id) ?? null;
}

export function compareExecutionDependencies(left, right) {
  const leftInfo = executionDependencyInfo(left);
  const rightInfo = executionDependencyInfo(right);
  return (
    (leftInfo?.order ?? Number.MAX_SAFE_INTEGER) -
      (rightInfo?.order ?? Number.MAX_SAFE_INTEGER) ||
    String(left).localeCompare(String(right))
  );
}

export function targetForExecutionDependency(id, label = "execution_dependency") {
  if (id === "") {
    return "";
  }
  const info = executionDependencyInfo(id);
  if (!info) {
    throw new Error(`${label} has no execution dependency metadata for ${id}`);
  }
  if (!info.target) {
    throw new Error(`${label} ${id} has no Make target mapping`);
  }
  return info.target;
}
