import {
  gridShellTestId,
  rowCellTestId,
  surfaceTabTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  queryViewRows,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
  collectionActionsPayload,
  collectionItems,
  editGenericCell,
  findRow,
  hostRefsFieldKey,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  identityRefsFieldKey,
  notesViewSchemaId,
  timelineViewSchemaId,
  type ViewRow,
} from "./phase4Helpers";

const exactScenarioTitle =
  "FE-I-P5-01 Verify Hosts, Identities, and Notes grids render contract-derived columns and preserve mention/entity provenance through edit and refresh.";

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);
const notesContract = requireViewContract(notesViewSchemaId);
const hostVisibleFields = hostsContract.defaultVisibleFields;
const identityVisibleFields = identitiesContract.defaultVisibleFields;
const noteVisibleFields = notesContract.defaultVisibleFields;
const hostAllFieldKeys = hostsContract.fields.map((field) => field.fieldKey);
const identityAllFieldKeys = identitiesContract.fields.map(
  (field) => field.fieldKey,
);
const noteAllFieldKeys = notesContract.fields.map((field) => field.fieldKey);

function visibleHeaderFieldKeys(page: Page, viewSchemaId: string) {
  return page
    .getByTestId(gridShellTestId(viewSchemaId))
    .locator('[role="columnheader"] [data-grid-field-key]')
    .evaluateAll((headers) =>
      headers.map((header) => header.getAttribute("data-grid-field-key") ?? ""),
    );
}

function assertFullCells(row: ViewRow, expectedFieldKeys: readonly string[]) {
  expect(Object.keys(row.cells).sort()).toEqual([...expectedFieldKeys].sort());
  expect(row.cells).not.toHaveProperty("record_id");
  expect(row.cells).not.toHaveProperty("row_version");
  expect(row).toHaveProperty("record_id");
  expect(row).toHaveProperty("row_version");
}

function mentionFingerprint(row: ViewRow, fieldKey: string, rawText: string) {
  const item = collectionItems(row, fieldKey).find(
    (candidate) => candidate.raw_text === rawText,
  );
  if (!item) {
    throw new Error(`missing ${fieldKey} item raw_text=${rawText}`);
  }
  return {
    item_ref: item.item_ref,
    raw_text: item.raw_text,
    resolved_record_id: item.resolved_record_id,
    provenance: item.provenance,
    confidence: item.confidence,
    matched_alias_text: item.matched_alias_text,
    item_kind: item.item_kind,
  };
}

function mixedRefPayload(
  rawText: string,
  resolvedRecordId: string,
  unresolvedRawText: string,
) {
  return {
    kind: "collection_actions_v1",
    actions: [
      {
        op: "add_resolved_ref",
        raw_text: rawText,
        resolved_record_id: resolvedRecordId,
      },
      {
        op: "add_token",
        raw_text: unresolvedRawText,
      },
    ],
  };
}

async function patchBodyForEdit(
  page: Page,
  viewSchemaId: string,
  recordId: string,
  fieldKey: string,
  value: string,
) {
  const responsePromise = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${recordId}`),
  );
  await editGenericCell(page, viewSchemaId, recordId, fieldKey, value);
  const response = await responsePromise;
  expect(response.ok()).toBeTruthy();
  return response.request().postDataJSON() as Record<string, unknown>;
}

test(exactScenarioTitle, async ({ page }) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEIP501"),
    "FE-I-P5-01 grid provenance",
  );
  const host = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("feip501-host"),
    "host.display_name": "FE-I-P5 Gateway",
    "host.hostname": "feip501-gateway.example.test",
    "host.aliases": collectionActionsPayload(["FEIP501 Gateway"]),
  })) as ViewRow;
  const identity = (await createViewRow(
    page,
    incidentId,
    identitiesViewSchemaId,
    {
      client_txn_id: uniqueTxn("feip501-identity"),
      "identity.display_name": "FE-I-P5 Analyst",
      "identity.upn": "feip501.analyst@example.test",
      "identity.email": "feip501.analyst@example.test",
      "identity.sam_account_name": "feip501",
      "identity.aliases": collectionActionsPayload(["FEIP501 Analyst"]),
    },
  )) as ViewRow;
  const note = (await createViewRow(page, incidentId, notesViewSchemaId, {
    client_txn_id: uniqueTxn("feip501-note"),
    "note.title": "FE-I-P5 Note",
    "note.body": "Initial provenance note",
  })) as ViewRow;
  const timeline = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("feip501-timeline"),
    "timeline.summary": "FE-I-P5 Gateway login by analyst",
    [hostRefsFieldKey]: mixedRefPayload(
      " FEIP501 Gateway ",
      host.record_id,
      "FEIP501 Unresolved Host",
    ),
    [identityRefsFieldKey]: mixedRefPayload(
      " FEIP501 Analyst ",
      identity.record_id,
      "FEIP501 Unresolved Identity",
    ),
  })) as ViewRow;

  const hostRowsBefore = (await queryViewRows(
    page,
    incidentId,
    hostsViewSchemaId,
  )) as ViewRow[];
  const identityRowsBefore = (await queryViewRows(
    page,
    incidentId,
    identitiesViewSchemaId,
  )) as ViewRow[];
  const noteRowsBefore = (await queryViewRows(
    page,
    incidentId,
    notesViewSchemaId,
  )) as ViewRow[];
  assertFullCells(findRow(hostRowsBefore, host.record_id), hostAllFieldKeys);
  assertFullCells(
    findRow(identityRowsBefore, identity.record_id),
    identityAllFieldKeys,
  );
  assertFullCells(findRow(noteRowsBefore, note.record_id), noteAllFieldKeys);

  const timelineRowsBefore = (await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  )) as ViewRow[];
  const timelineBefore = findRow(timelineRowsBefore, timeline.record_id);
  const hostMentionBefore = mentionFingerprint(
    timelineBefore,
    hostRefsFieldKey,
    " FEIP501 Gateway ",
  );
  const identityMentionBefore = mentionFingerprint(
    timelineBefore,
    identityRefsFieldKey,
    " FEIP501 Analyst ",
  );
  const unresolvedHostMentionBefore = mentionFingerprint(
    timelineBefore,
    hostRefsFieldKey,
    "FEIP501 Unresolved Host",
  );
  const unresolvedIdentityMentionBefore = mentionFingerprint(
    timelineBefore,
    identityRefsFieldKey,
    "FEIP501 Unresolved Identity",
  );
  expect(hostMentionBefore.resolved_record_id).toBe(host.record_id);
  expect(identityMentionBefore.resolved_record_id).toBe(identity.record_id);
  expect(unresolvedHostMentionBefore.resolved_record_id ?? null).toBeNull();
  expect(unresolvedHostMentionBefore.item_kind).toBe("unresolved_mention");
  expect(unresolvedIdentityMentionBefore.resolved_record_id ?? null).toBeNull();
  expect(unresolvedIdentityMentionBefore.item_kind).toBe("unresolved_mention");

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      hostsViewSchemaId,
    )}`,
  );
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect
    .poll(async () => visibleHeaderFieldKeys(page, hostsViewSchemaId))
    .toEqual([...hostVisibleFields]);
  await expect(
    page.getByTestId(rowCellTestId(host.record_id, "host.aliases")),
  ).toContainText("FEIP501 Gateway");
  const hostPatch = await patchBodyForEdit(
    page,
    hostsViewSchemaId,
    host.record_id,
    "host.display_name",
    "FE-I-P5 Gateway edited",
  );
  expect(hostPatch).toMatchObject({
    view_schema_id: hostsViewSchemaId,
    base_row_version: host.row_version,
    changes: [
      {
        field_key: "host.display_name",
        value: "FE-I-P5 Gateway edited",
      },
    ],
  });
  await expect(
    page.getByTestId(rowCellTestId(host.record_id, "host.display_name")),
  ).toHaveText("FE-I-P5 Gateway edited");

  await page.getByTestId(surfaceTabTestId(identitiesViewSchemaId)).click();
  await expect
    .poll(async () => visibleHeaderFieldKeys(page, identitiesViewSchemaId))
    .toEqual([...identityVisibleFields]);
  await expect(
    page.getByTestId(rowCellTestId(identity.record_id, "identity.aliases")),
  ).toContainText("FEIP501 Analyst");
  const identityPatch = await patchBodyForEdit(
    page,
    identitiesViewSchemaId,
    identity.record_id,
    "identity.display_name",
    "FE-I-P5 Analyst edited",
  );
  expect(identityPatch).toMatchObject({
    view_schema_id: identitiesViewSchemaId,
    base_row_version: identity.row_version,
    changes: [
      {
        field_key: "identity.display_name",
        value: "FE-I-P5 Analyst edited",
      },
    ],
  });
  await expect(
    page.getByTestId(
      rowCellTestId(identity.record_id, "identity.display_name"),
    ),
  ).toHaveText("FE-I-P5 Analyst edited");

  await page.getByTestId(surfaceTabTestId(notesViewSchemaId)).click();
  await expect
    .poll(async () => visibleHeaderFieldKeys(page, notesViewSchemaId))
    .toEqual([...noteVisibleFields]);
  const notePatch = await patchBodyForEdit(
    page,
    notesViewSchemaId,
    note.record_id,
    "note.body",
    "Edited provenance note",
  );
  expect(notePatch).toMatchObject({
    view_schema_id: notesViewSchemaId,
    base_row_version: note.row_version,
    changes: [
      {
        field_key: "note.body",
        value: "Edited provenance note",
      },
    ],
  });
  await expect(
    page.getByTestId(rowCellTestId(note.record_id, "note.body")),
  ).toHaveText("Edited provenance note");

  await page.reload();
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await page.getByTestId(surfaceTabTestId(hostsViewSchemaId)).click();
  await expect(
    page.getByTestId(rowCellTestId(host.record_id, "host.display_name")),
  ).toHaveText("FE-I-P5 Gateway edited");
  await page.getByTestId(surfaceTabTestId(identitiesViewSchemaId)).click();
  await expect(
    page.getByTestId(
      rowCellTestId(identity.record_id, "identity.display_name"),
    ),
  ).toHaveText("FE-I-P5 Analyst edited");
  await page.getByTestId(surfaceTabTestId(notesViewSchemaId)).click();
  await expect(
    page.getByTestId(rowCellTestId(note.record_id, "note.body")),
  ).toHaveText("Edited provenance note");

  const timelineRowsAfter = (await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  )) as ViewRow[];
  const timelineAfter = findRow(timelineRowsAfter, timeline.record_id);
  expect(
    mentionFingerprint(timelineAfter, hostRefsFieldKey, " FEIP501 Gateway "),
  ).toEqual(hostMentionBefore);
  expect(
    mentionFingerprint(
      timelineAfter,
      hostRefsFieldKey,
      "FEIP501 Unresolved Host",
    ),
  ).toEqual(unresolvedHostMentionBefore);
  expect(
    mentionFingerprint(
      timelineAfter,
      identityRefsFieldKey,
      " FEIP501 Analyst ",
    ),
  ).toEqual(identityMentionBefore);
  expect(
    mentionFingerprint(
      timelineAfter,
      identityRefsFieldKey,
      "FEIP501 Unresolved Identity",
    ),
  ).toEqual(unresolvedIdentityMentionBefore);
});
