import type {
  CreateViewRowRequest,
  ViewRow,
} from "@cartulary/protocol-ts/http";
import {
  gridSavedRowsSelector,
  gridShellTestId,
  rowCellTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, type Locator, type Page } from "@playwright/test";
import {
  timelineExpectations,
  timelineRecipe,
  timelineRows,
  verifyTimelineInvestigation,
} from "../../../../../tools/harness/fixtures/timeline-investigation/index.mjs";
import { uniqueIncidentKey } from "../runtime/fixtureIdentity";
import { seedVisualTimelineInvestigation } from "../timeline/timelineInvestigation";
import { timelineScenarioFields } from "../timeline/timelineScenarioFields";
import { createViewRow } from "../workbook/query";
import {
  requireTimelineDataProfile,
  timelineDataProfiles,
} from "./timelineDataProfiles";

type Investigation = Awaited<
  ReturnType<typeof seedVisualTimelineInvestigation>
>;
const investigations = new Map<
  string,
  { fixture: Investigation; scenarioRows: Map<string, ViewRow> }
>();

export function registerTimelineVisualFixture(fixture: Investigation) {
  if (investigations.has(fixture.incidentId))
    throw new Error("Duplicate Timeline visual incident");
  investigations.set(fixture.incidentId, { fixture, scenarioRows: new Map() });
  return fixture;
}

export async function createRichTimelineIncident(
  page: Page,
  key: string,
  title: string,
) {
  const fixture = registerTimelineVisualFixture(
    await seedVisualTimelineInvestigation(page, {
      continuationCount: 0,
      incident: { key, title },
    }),
  );
  return fixture.incidentId;
}

export function visualTimelineRecord(incidentId: string, key: string) {
  const fixture = investigations.get(incidentId)?.fixture;
  const row = fixture?.rows.find(
    (row) => row.record_id === fixture.recordId(key),
  );
  if (!row) throw new Error(`Missing investigation record ${key}`);
  return row;
}

export async function createRichTimelineSource(
  page: Page,
  key: string,
  fields: CreateViewRowRequest,
) {
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUAL-SOURCE"),
    "Service-account investigation",
  );
  const source = await createVisualTimelineRow(page, incidentId, key, fields);
  return { incidentId, source };
}

export async function createVisualTimelineRow(
  page: Page,
  incidentId: string,
  semanticKey: string,
  payload: CreateViewRowRequest,
) {
  const state = investigations.get(incidentId);
  if (!state)
    throw new Error(
      "Timeline scenario row requires an admitted rich investigation",
    );
  if (
    state.scenarioRows.has(semanticKey) ||
    state.fixture.mapping.has(semanticKey)
  )
    throw new Error(`Duplicate visual semantic key ${semanticKey}`);
  const fields = timelineScenarioFields(payload, state.scenarioRows.size);
  const row = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    fields,
  );
  for (const field of timelineExpectations.fields.source)
    expect(row.cells[field]?.value, `scenario ${semanticKey}/${field}`).toBe(
      fields[field],
    );
  state.scenarioRows.set(semanticKey, row);
  return row;
}

/** Query before fault installation, and once after the scenario releases its faults. */
export async function verifyTimelineVisualCores() {
  try {
    for (const { fixture, scenarioRows } of investigations.values()) {
      const rows = await fixture.readRows();
      const ids = new Set(
        timelineRecipe.rows.map((row) => fixture.recordId(row.key)),
      );
      verifyTimelineInvestigation({
        authored: timelineRows(timelineRecipe, 0),
        rows: rows.filter((row) => ids.has(row.record_id)),
        mapping: fixture.mapping,
      });
      expect(rows).toHaveLength(
        fixture.outcomes.total_rows + scenarioRows.size,
      );
    }
  } finally {
    investigations.clear();
  }
}

/** Read only the retained UI here: intentional request failures must remain undisturbed. */
export async function verifyTimelineCaptureData(
  page: Page,
  name: string,
  scope?: Locator,
) {
  const grid = page.getByTestId(gridShellTestId(timelineViewSchemaId));
  const declared = timelineDataProfiles.captures[name];
  const gridPresent = (await grid.count()) > 0;
  let capturesGrid = gridPresent;
  if (scope && gridPresent) {
    const handle = await grid.elementHandle();
    capturesGrid =
      handle !== null &&
      (await scope.evaluate(
        (element, target) => element === target || element.contains(target),
        handle,
      ));
  }
  if (!declared && !capturesGrid) return null;
  const profile = requireTimelineDataProfile(name);
  const saved = grid.locator(gridSavedRowsSelector());
  if (profile.kind === "empty") {
    await expect(saved).toHaveCount(0);
    return { ...profile, core_rows: 0 };
  }
  if (profile.kind === "sparse") {
    // The shared sparse reading/geometry specimen intentionally has one saved row.
    await expect(saved).toHaveCount(1);
    return { ...profile, core_rows: 0 };
  }
  const incidentId = new URL(page.url()).searchParams.get("incident_id");
  const state = incidentId ? investigations.get(incidentId) : undefined;
  if (!state)
    throw new Error(`Capture ${name} has no verified rich Timeline fixture`);
  for (const row of state.fixture.rows) {
    for (const field of timelineExpectations.fields.source) {
      const cell = page.getByTestId(rowCellTestId(row.record_id, field));
      if (await cell.count())
        await expect(cell).toHaveText(String(row.cells[field]?.value));
    }
  }
  return {
    ...profile,
    dataset_id: state.fixture.outcomes.dataset_id,
    content_revision: state.fixture.outcomes.content_revision,
    core_rows: 12,
    seeded_rows: state.fixture.outcomes.total_rows + state.scenarioRows.size,
  };
}
