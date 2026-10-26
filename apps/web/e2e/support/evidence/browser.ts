import { Buffer } from "node:buffer";

import { gridShellTestId } from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, type Locator, type Page } from "@playwright/test";
import { createIncident } from "../incidents/fixtures";
import { uniqueIncidentKey, uniqueTxn } from "../runtime/fixtureIdentity";
import { createViewRow, patchRecord } from "../workbook/query";
import { createUploadedEvidenceFixture } from "./fixtures";

export async function openTimelineSurface(page: Page, incidentId: string) {
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      timelineViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
}

export function tinyPNG() {
  return Buffer.from(
    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=",
    "base64",
  );
}

export async function openEvidenceSurface(page: Page, incidentId: string) {
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      evidenceViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
  ).toBeVisible();
}

export async function linkedTextReviewFixture(page: Page) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TIMELINE-LINKED-REVIEW"),
    "Timeline linked Evidence review",
  );
  const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("timeline-linked-source"),
    "timeline.activity_synopsis_text": "Linked text review source",
  });
  const title =
    "Synthetic acquisition notes with a long but meaningful original Evidence title";
  const evidence = await createUploadedEvidenceFixture(page, incidentId, {
    txnPrefix: "e5-preview",
    collectorPartyText: "Browser evidence",
    title,
    filename: "acquisition-notes.txt",
    contentType: "text/plain",
    body: Buffer.from("linked timeline preview body", "utf8"),
  });
  await linkEvidenceToTimeline(page, source.record_id, source.row_version, [
    evidence.record_id,
  ]);

  return { incidentId, source, title, evidence };
}

export async function linkEvidenceToTimeline(
  page: Page,
  sourceRecordId: string,
  rowVersion: number,
  evidenceRecordIds: readonly [string, ...string[]],
) {
  const [firstRecordId, ...restRecordIds] = evidenceRecordIds;
  return patchRecord(page, sourceRecordId, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: rowVersion,
    client_txn_id: uniqueTxn("timeline-linked-evidence"),
    changes: [
      {
        field_key: "timeline.attached_evidence_ids",
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            { op: "add_record_ref", linked_record_id: firstRecordId },
            ...restRecordIds.map((recordId) => ({
              op: "add_record_ref" as const,
              linked_record_id: recordId,
            })),
          ],
        },
      },
    ],
  });
}

export async function expectRevealedInInspectorBody(target: Locator) {
  await expect
    .poll(() =>
      target.evaluate((element) => {
        const body = element.closest("[data-inspector-scroll-body]");
        if (!(body instanceof HTMLElement)) return false;
        const bodyRect = body.getBoundingClientRect();
        const targetRect = element.getBoundingClientRect();
        return (
          targetRect.bottom > bodyRect.top + 1 &&
          targetRect.top < bodyRect.bottom - 1
        );
      }),
    )
    .toBe(true);
}
