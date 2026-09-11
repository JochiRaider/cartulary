import {
  timelineCaptureActionTestId,
  timelineRowSupersedeButtonTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, type Page } from "@playwright/test";
import { createIncident } from "../incidents/fixtures";
import { uniqueIncidentKey, uniqueTxn } from "../runtime/fixtureIdentity";
import { createViewRow } from "./query";
import { openTimelineRowActions } from "./rowMutations";

export async function authorTimelineSupersession(
  page: Page,
  recordId: string,
  reason: string,
  replacementId: string | null,
) {
  await page
    .getByTestId(timelineCaptureActionTestId("reason", recordId))
    .fill(reason);
  const select = page.getByTestId(
    timelineCaptureActionTestId("replacement", recordId),
  );
  if (replacementId) {
    await expect(
      select.locator(`option[value="${replacementId}"]`),
    ).toHaveCount(1);
    await select.selectOption(replacementId);
  } else await select.selectOption("");
  await page
    .getByTestId(timelineCaptureActionTestId("review", recordId))
    .focus();
  await page.keyboard.press("Enter");
  await expect(
    page.getByTestId(timelineCaptureActionTestId("confirm", recordId)),
  ).toBeFocused();
}

export async function openTimelineSupersessionFixture(
  page: Page,
  navigate: (url: string) => Promise<unknown> = (url) => page.goto(url),
) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TIMELINE-CAPTURE"),
    "Timeline action review",
  );
  const target = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("target"),
    "timeline.activity_synopsis_text": "Investigated observation",
    "timeline.device_object_text": "Workstation A",
    "timeline.activity_utc_text": "2025-02-17T11:00:00Z",
  });
  const replacement = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("replacement"),
      "timeline.activity_synopsis_text": "Investigated observation",
      "timeline.device_object_text": "Workstation B",
      "timeline.activity_utc_text": "2025-02-17T11:05:00Z",
    },
  );
  await navigate(`/?incident_id=${incidentId}`);
  await openTimelineRowActions(page, target.record_id);
  await page
    .getByTestId(timelineRowSupersedeButtonTestId(target.record_id))
    .press("Enter");
  await authorTimelineSupersession(
    page,
    target.record_id,
    "The verified source corrects this observation.\nPreserve the original analyst attribution.",
    replacement.record_id,
  );
  return { incidentId, target, replacement };
}
