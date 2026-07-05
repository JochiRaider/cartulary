const checkRootTargets = new Set([
  "frontend-unit",
  "frontend-import-boundary-check",
  "generated-artifact-policy-check",
  "generate-drift",
  "phase-ledger-drift",
  "browser-e2e-webserver-backed",
  "browser-e2e-stateful",
]);

const explicitTargetInputs = new Map([
  ["browser-e2e-support", "BROWSER_SUPPORT_RESULTS_DIR"],
  ["browser-e2e-visual", "BROWSER_VISUAL_RESULTS_DIR"],
  ["browser-e2e-a11y", "BROWSER_A11Y_RESULTS_DIR"],
  ["browser-e2e-a11y-preflight", "BROWSER_A11Y_PREFLIGHT_RESULTS_DIR"],
  ["browser-e2e-measurement", "BROWSER_MEASUREMENT_RESULTS_DIR"],
]);

export const frontendEvidenceAuditRootInputNames = Object.freeze([
  "CHECK_RESULTS_DIR",
  "BROWSER_SUPPORT_RESULTS_DIR",
  "BROWSER_VISUAL_RESULTS_DIR",
  "BROWSER_A11Y_RESULTS_DIR",
  "BROWSER_A11Y_PREFLIGHT_RESULTS_DIR",
  "BROWSER_MEASUREMENT_RESULTS_DIR",
]);

export function frontendEvidenceAuditInputForTarget(targetName) {
  if (checkRootTargets.has(targetName)) {
    return "CHECK_RESULTS_DIR";
  }
  return explicitTargetInputs.get(targetName) ?? "";
}

export function frontendEvidenceAuditRootForTarget(targetName, roots) {
  const inputName = frontendEvidenceAuditInputForTarget(targetName);
  return inputName ? roots[inputName] ?? "" : "";
}
