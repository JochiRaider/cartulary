import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  entityInspectButtonTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  notesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, type Page } from "@playwright/test";
import { createIncident } from "../incidents/fixtures";
import { uniqueIncidentKey, uniqueTxn } from "../runtime/fixtureIdentity";
import { createViewRow } from "./query";
import { openRecoveryItem, recoveryEntry } from "./recovery";
import {
  openGenericInspectorForRecord,
  openTimelineInspector,
} from "./rowMutations";

export const noteSourceFields = {
  [timelineViewSchemaId]: "timeline.activity_synopsis_text",
  [hostsViewSchemaId]: "host.display_name",
  [identitiesViewSchemaId]: "identity.display_name",
  [evidenceViewSchemaId]: "evidence.title",
} as const;
export async function openNoteFixture(
  page: Page,
  view: string = timelineViewSchemaId,
  navigate: (url: string) => Promise<unknown> = (url) => page.goto(url),
) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("LINKED-NOTE"),
    "Source-linked Note authoring",
  );
  const field = noteSourceFields[view as keyof typeof noteSourceFields];
  if (!field) throw new Error("Unsupported Note source fixture");
  const source = await createViewRow(page, incident, view, {
    client_txn_id: uniqueTxn("note-source"),
    [field]: "Reviewed investigation source",
  });
  await navigate(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
  );
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
  if (view === timelineViewSchemaId)
    await openTimelineInspector(page, source.record_id);
  else if (view === hostsViewSchemaId || view === identitiesViewSchemaId) {
    const targetTestId = entityInspectButtonTestId(
      view === hostsViewSchemaId ? "host" : "identity",
      source.record_id,
    );
    await scrollGridTargetIntoView({ page, surface: view, targetTestId });
    await page.getByTestId(targetTestId).click();
    await page.getByTestId(workbookInspectorToggleTestId(view)).click();
  } else await openGenericInspectorForRecord(page, view, source.record_id);
  const action = page.getByTestId(
    workbookInspectorFeatureActionTestId(view, "create_related.note"),
  );
  await action.focus();
  await action.press("Enter");
  const form = page.getByRole("region", { name: "Create Note", exact: true });
  await expect(form).toBeVisible();
  return { incident, source, view, form, action };
}

export async function retainNoteUncertainResult(
  page: Page,
  sourceId: string,
  view: string,
) {
  let reachedTransport: () => void = () => {};
  const captured = new Promise<void>((resolve) => {
    reachedTransport = resolve;
  });
  await page.route(`**/records/${sourceId}/linked-notes`, async (route) => {
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    await route.abort("failed");
    reachedTransport();
  });
  await page.getByTestId(genericCreateSubmitTestId(notesViewSchemaId)).click();
  await captured;
  const summary = recoveryEntry(page);
  await expect(summary).toBeVisible();
  await page.getByTestId(workbookInspectorCloseButtonTestId(view)).click();
  await openRecoveryItem(page, /^Note creation ·/);
  const recovery = page.getByRole("region", {
    name: "Retained Note authoring",
    exact: true,
  });
  await expect(
    recovery.getByRole("button", { name: "Recover submission", exact: true }),
  ).toBeEnabled();
  return { summary, recovery };
}
