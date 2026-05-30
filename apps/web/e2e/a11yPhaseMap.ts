import { existsSync, readdirSync, readFileSync } from "node:fs";
import path from "node:path";

type FrontendPhaseRow = {
  claim_status: string;
  id: string;
  scenario_titles: string[];
};

type FrontendPhaseMap = {
  rows: FrontendPhaseRow[];
};

function findRepoRoot(): string {
  let candidate = process.cwd();
  while (true) {
    if (
      existsSync(path.join(candidate, "tools", "frontend_phase_registry.json"))
    ) {
      return candidate;
    }
    const parent = path.dirname(candidate);
    if (parent === candidate) {
      throw new Error("could not find tools/frontend_phase_registry.json");
    }
    candidate = parent;
  }
}

const frontendPhaseMapDir = path.join(
  findRepoRoot(),
  "tools",
  "frontend_phase_maps",
);

function loadFrontendPhaseMaps(): FrontendPhaseMap[] {
  return readdirSync(frontendPhaseMapDir)
    .filter((name) => /^fe_p\d+_test_map\.json$/.test(name))
    .sort((left, right) => left.localeCompare(right, "en", { numeric: true }))
    .map(
      (name) =>
        JSON.parse(
          readFileSync(path.join(frontendPhaseMapDir, name), "utf8"),
        ) as FrontendPhaseMap,
    );
}

function accessibilityRows() {
  return loadFrontendPhaseMaps().flatMap((manifest) =>
    manifest.rows.filter((row) => row.id.startsWith("FE-A11Y-")),
  );
}

export function scenarioTitlesForAccessibilityRow(rowId: string): string[] {
  const row = accessibilityRows().find((candidate) => candidate.id === rowId);
  if (!row) {
    throw new Error(`missing frontend accessibility row ${rowId}`);
  }
  return row.scenario_titles;
}

export function blockedAccessibilityScenarioTitles(): string[] {
  return accessibilityRows().flatMap((row) =>
    row.claim_status === "blocked" ? row.scenario_titles : [],
  );
}

export const p1AccessibilityScenarioTitles = scenarioTitlesForAccessibilityRow(
  "FE-A11Y-P1-01",
) as [
  string,
  string,
  string,
  string,
  string,
  string,
  string,
  string,
  string,
];

if (p1AccessibilityScenarioTitles.length !== 9) {
  throw new Error(
    `FE-A11Y-P1-01 must declare exactly 9 scenarios; found ${p1AccessibilityScenarioTitles.length}`,
  );
}
