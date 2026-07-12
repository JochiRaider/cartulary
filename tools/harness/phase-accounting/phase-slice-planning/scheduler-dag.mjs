import {
  loadRuntimeBinaryRegistry,
  runtimeBinaryProducerTargetsForIDs,
} from "../../runtime-binary-registry.mjs";
import { readinessAttributionForMakeTarget } from "../../scheduler/scheduler-manifest.mjs";

const serviceSessionWeightMs = 4000;
const readinessWeights = new Map([
  ["frontend-install", 1500],
  ["build-server-harness", 5000],
  ["build-migrate", 3000],
  ["build-web", 8000],
  ["test-service-images", 5000],
]);

function unique(values) {
  return Array.from(new Set(values));
}

function addNeed(unit, need) {
  unit.needs = unique([...(unit.needs ?? []), need]);
}

function completionKeys(unit) {
  return unit.completionKeys ?? unit.completion_keys ?? [unit.id];
}

function setupTargets(plan, root) {
  const classes = new Set(plan.workUnits.map((unit) => unit.class));
  const hasBrowser = classes.has("browser");
  const hasFrontend = classes.has("frontend");
  const hasBackendProcess = plan.workUnits.some(
    (unit) => unit.target === "backend-process",
  );
  const hasService =
    plan.service_requirements.includes("postgres") ||
    plan.service_requirements.includes("object_store");
  const targets = [];
  if (hasFrontend || hasBrowser) {
    targets.push("frontend-install");
  }
  if (hasBackendProcess || hasBrowser) {
    targets.push("build-server-harness");
  }
  targets.push(
    ...runtimeBinaryProducerTargetsForIDs(
      loadRuntimeBinaryRegistry({ repoRoot: root }),
      plan.runtime_binaries ?? [],
      "phase-slice",
    ),
  );
  if (hasBrowser) {
    targets.push("build-migrate", "build-web");
  }
  if (hasService) {
    targets.push("test-service-images");
  }
  return unique(targets);
}

function readinessUnit(plan, target) {
  const readinessAttribution = readinessAttributionForMakeTarget(target);
  return {
    id: `readiness:${target}`,
    label: target,
    kind: "make_target",
    type: "make_target",
    class: "readiness",
    target,
    aggregateTarget: target,
    group: "readiness",
    needs: target === "build-web" ? ["frontend-install"] : [],
    completionKeys: [target],
    failureKeys: [target],
    weightMs: readinessWeights.get(target) ?? 1000,
    make_prerequisite_policy: "run",
    ...(readinessAttribution
      ? { readiness_attribution: readinessAttribution }
      : {}),
    resourceClaims: new Map([["process", 1]]),
    order: plan.nextOrder++,
  };
}

function runtimeProducerTargetsForUnit(unit, root) {
  return runtimeBinaryProducerTargetsForIDs(
    loadRuntimeBinaryRegistry({ repoRoot: root }),
    unit.runtime_binaries ?? unit.runtimeBinaries ?? [],
    "phase-slice",
  );
}

export function normalizePhaseSliceSchedulerDAG(plan, root) {
  const selectedUnits = [...plan.workUnits];
  const targets = setupTargets(plan, root);
  const targetSet = new Set(targets);
  const hasService = targetSet.has("test-service-images");
  const serviceSessionKey = `service_session:${plan.target}`;

  for (const unit of selectedUnits) {
    if (["frontend", "browser"].includes(unit.class) && targetSet.has("frontend-install")) {
      addNeed(unit, "frontend-install");
    }
    if (
      (unit.class === "browser" || unit.target === "backend-process") &&
      targetSet.has("build-server-harness")
    ) {
      addNeed(unit, "build-server-harness");
    }
    if (unit.class === "browser") {
      for (const target of ["build-migrate", "build-web"]) {
        if (targetSet.has(target)) {
          addNeed(unit, target);
        }
      }
    }
    for (const target of runtimeProducerTargetsForUnit(unit, root)) {
      if (targetSet.has(target)) {
        addNeed(unit, target);
      }
    }
    if (hasService) {
      addNeed(unit, serviceSessionKey);
      unit.serviceSession = { target: plan.target };
    }
  }

  const readinessUnits = targets.map((target) => readinessUnit(plan, target));
  const serviceUnits = [];
  if (hasService) {
    serviceUnits.push({
      id: `${plan.target}:service-session`,
      label: `${plan.target}/service-session`,
      kind: "service_session",
      type: "service_session_start",
      class: "service",
      target: plan.target,
      aggregateTarget: plan.target,
      group: "service-session",
      needs: ["test-service-images"],
      completionKeys: [serviceSessionKey],
      failureKeys: [serviceSessionKey],
      weightMs: serviceSessionWeightMs,
      resourceClaims: new Map([["process", 1]]),
      serviceSession: { target: plan.target },
      order: plan.nextOrder++,
    });
    serviceUnits.push({
      id: `${plan.target}:service-complete`,
      label: `${plan.target}/service-complete`,
      kind: "service_complete",
      type: "service_complete",
      class: "service",
      target: plan.target,
      aggregateTarget: plan.target,
      group: "service-session",
      needs: unique(selectedUnits.flatMap(completionKeys)),
      completionKeys: [`service_complete:${plan.target}`],
      failureKeys: [`service_complete:${plan.target}`],
      weightMs: 1,
      resourceClaims: new Map(),
      serviceSession: { target: plan.target },
      countInTotal: false,
      order: plan.nextOrder++,
    });
  }

  plan.workUnits = [...readinessUnits, ...serviceUnits, ...selectedUnits];
  return plan;
}
