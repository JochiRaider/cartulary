export const sectionDefinitions = [
  ["unit", "U-"],
  ["integration", "I-"],
  ["e2e", "E-"],
  ["visual", "V-"],
];

export const validCoverage = new Set(["authoritative", "supplemental"]);
export const validGoSections = new Set(["unit", "integration", "e2e"]);
export const validClaimStatuses = new Set(["implemented", "blocked", "not_applicable"]);
export const defaultReasonRequiredLayers = new Set([
  "browser_functional",
  "browser_stateful",
  "browser_support",
  "browser_visual",
]);
export const validRuntimeBinaries = new Set(["operator"]);
export const supportTargetSections = new Map([
  ["backend_unit", "unit"],
  ["backend_integration_support", "integration"],
]);
