import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  partiesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, type Page } from "@playwright/test";
import { createIncident } from "../incidents/fixtures";
import { uniqueIncidentKey, uniqueTxn } from "../runtime/fixtureIdentity";
import { createViewRow } from "./query";
import { openRecoveryItem, recoveryEntry } from "./recovery";
import { openTimelineInspector } from "./rowMutations";

export async function openTimelineEvidenceFixture(
  page: Page,
  navigate: (url: string) => Promise<unknown> = (url) => page.goto(url),
) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("TRE-PRESENTATION"),
    "Timeline Evidence review",
  );
  const source = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("source"),
    "timeline.activity_synopsis_text": "Preserved investigation source",
    "timeline.raw_activity_text": "Original source text remains unchanged.",
  });
  const party = await createViewRow(page, incident, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("party"),
    "party.display_name": "Response collection team",
    "party.party_kind": "team",
  });
  await navigate(
    `/?incident_id=${incident}&view_schema_id=${timelineViewSchemaId}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await openTimelineInspector(page, source.record_id);
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        timelineViewSchemaId,
        "create_related.evidence",
      ),
    )
    .click();
  const form = page.getByRole("region", {
    name: "Create Related Evidence",
    exact: true,
  });
  await expect(form).toBeVisible();
  await form
    .getByTestId(genericCreateFieldTestId("evidence.collector_party_text"))
    .fill("Response team collection log");
  return { incident, source, party, form };
}

/** Commit the real create, then detach presentation before returning its response. */
export async function retainTimelineEvidencePartialResult(page: Page) {
  let release = () => {},
    committed = false;
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  await page.route(`**/views/${evidenceViewSchemaId}/rows`, async (route) => {
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    committed = true;
    await gate;
    await route.fulfill({ response });
  });
  await page
    .getByTestId(genericCreateSubmitTestId(evidenceViewSchemaId))
    .click();
  await expect.poll(() => committed).toBe(true);
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  release();
  const summary = recoveryEntry(page);
  await openRecoveryItem(page, /^Timeline Evidence creation ·/);
  const recovery = page.getByRole("region", {
    name: "Retained Timeline Evidence creation",
    exact: true,
  });
  await expect(
    recovery
      .getByText("Evidence created; Timeline link incomplete.", { exact: true })
      .first(),
  ).toBeVisible();
  await expect(
    recovery.getByText("Creation: saved; views refreshed.", { exact: true }),
  ).toBeVisible();
  return { recovery, summary };
}
