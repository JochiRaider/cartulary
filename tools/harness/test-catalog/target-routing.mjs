const playwrightTargets = Object.freeze({
  accessibility: "browser-e2e-a11y",
  measurement: "browser-e2e-measurement",
  stateful: "browser-e2e-stateful",
  support: "browser-e2e-support",
  visual: "browser-e2e-visual",
  webserver_backed: "browser-e2e-webserver-backed",
});

export function goTargetForFamily(familyID) {
  const family = String(familyID).split(".").at(-1);
  if (["engine", "fixtures", "support_unit", "unit"].includes(family)) {
    return "backend-unit";
  }
  if (["storage", "store"].includes(family)) return "backend-store";
  if (family === "support_integration") return "backend-integration-support";
  if (family === "process") return "backend-process";
  return "backend-integration";
}

export function commandTargetForEvidenceTarget(targetID) {
  if (targetID === "backend-integration-support") return "backend-integration";
  return targetID;
}

export function targetForCatalogRow(row) {
  if (row.runner === "go") {
    if (row.evidence_class === "unit" && row.runtime_profile_id === "none") {
      return "backend-unit";
    }
    return goTargetForFamily(row.family_id);
  }
  if (row.runner === "vitest") return "frontend-unit";
  if (row.runner === "playwright") {
    const target = playwrightTargets[row.selector.stage];
    if (!target) {
      throw new Error(
        `catalog row ${row.row_id} has unsupported Playwright stage ${row.selector.stage}`,
      );
    }
    return target;
  }
  if (row.runner === "shell") {
    if (!/^harness\.[a-z][a-z0-9_]*$/u.test(row.selector.operation_id ?? "")) throw new Error(`catalog row ${row.row_id} has no canonical operation`);
    return row.selector.operation_id;
  }

  throw new Error(`catalog row ${row.row_id} has unsupported runner ${row.runner}`);
}
