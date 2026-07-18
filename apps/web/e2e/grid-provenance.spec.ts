import {
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  gridSortHeaderTestId,
  rowCellTestId,
  surfaceTabTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  notesViewSchemaId,
  timelineViewSchemaId,
} from "./support/contracts/workbookSurfaces";
import {
  aliasCollectionActionsPayload,
  collectionItems,
  findRow,
  hostRefsFieldKey,
  identityRefsFieldKey,
  type ViewRow,
} from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { editGenericCell } from "./support/workbook/rowMutations";

const exactScenarioTitle =
  "Verify Hosts, Identities, and Notes grids render contract-derived columns and preserve mention/entity provenance through edit and refresh.";

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

async function expectVisibleContractFieldsReachable(
  page: Page,
  viewSchemaId: string,
  fieldKeys: readonly string[],
) {
  for (const fieldKey of fieldKeys) {
    const headerTestId = gridSortHeaderTestId(viewSchemaId, fieldKey);
    await scrollGridTargetIntoView({
      page,
      surface: viewSchemaId,
      targetTestId: headerTestId,
    });
    await expect(page.getByTestId(headerTestId)).toHaveAttribute(
      "data-grid-field-key",
      fieldKey,
    );
  }
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
    uniqueIncidentKey("GRIDPROVENANCE"),
    "integration.entity-linking.row-01 grid provenance",
  );
  const host = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("grid-provenance-host"),
    "host.display_name": "integration.entity-linking Gateway",
    "host.hostname": "grid-provenance-gateway.example.test",
    "host.aliases": aliasCollectionActionsPayload(["GRIDPROVENANCE Gateway"]),
  })) as ViewRow;
  const identity = (await createViewRow(
    page,
    incidentId,
    identitiesViewSchemaId,
    {
      client_txn_id: uniqueTxn("grid-provenance-identity"),
      "identity.display_name": "integration.entity-linking Analyst",
      "identity.upn": "grid-provenance.analyst@example.test",
      "identity.email": "grid-provenance.analyst@example.test",
      "identity.sam_account_name": "grid-provenance",
      "identity.aliases": aliasCollectionActionsPayload([
        "GRIDPROVENANCE Analyst",
      ]),
    },
  )) as ViewRow;
  const note = (await createViewRow(page, incidentId, notesViewSchemaId, {
    client_txn_id: uniqueTxn("grid-provenance-note"),
    "note.title": "integration.entity-linking Note",
    "note.body": "Initial provenance note",
  })) as ViewRow;
  const timeline = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("grid-provenance-timeline"),
      "timeline.activity_synopsis_text":
        "integration.entity-linking Gateway login by analyst",
      [hostRefsFieldKey]: mixedRefPayload(
        " GRIDPROVENANCE Gateway ",
        host.record_id,
        "GRIDPROVENANCE Unresolved Host",
      ),
      [identityRefsFieldKey]: mixedRefPayload(
        " GRIDPROVENANCE Analyst ",
        identity.record_id,
        "GRIDPROVENANCE Unresolved Identity",
      ),
    },
  )) as ViewRow;

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
    " GRIDPROVENANCE Gateway ",
  );
  const identityMentionBefore = mentionFingerprint(
    timelineBefore,
    identityRefsFieldKey,
    " GRIDPROVENANCE Analyst ",
  );
  const unresolvedHostMentionBefore = mentionFingerprint(
    timelineBefore,
    hostRefsFieldKey,
    "GRIDPROVENANCE Unresolved Host",
  );
  const unresolvedIdentityMentionBefore = mentionFingerprint(
    timelineBefore,
    identityRefsFieldKey,
    "GRIDPROVENANCE Unresolved Identity",
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
  await expectVisibleContractFieldsReachable(
    page,
    hostsViewSchemaId,
    hostVisibleFields,
  );
  await scrollGridCellIntoView({
    cellKey: "host.aliases",
    page,
    recordId: host.record_id,
    surface: hostsViewSchemaId,
  });
  await expect(
    page.getByTestId(rowCellTestId(host.record_id, "host.aliases")),
  ).toContainText("GRIDPROVENANCE Gateway");
  const hostPatch = await patchBodyForEdit(
    page,
    hostsViewSchemaId,
    host.record_id,
    "host.display_name",
    "integration.entity-linking Gateway edited",
  );
  expect(hostPatch).toMatchObject({
    view_schema_id: hostsViewSchemaId,
    base_row_version: host.row_version,
    changes: [
      {
        field_key: "host.display_name",
        value: "integration.entity-linking Gateway edited",
      },
    ],
  });
  expect(String(hostPatch.client_txn_id)).toMatch(
    /^entity-patch-cartulary\.view\.hosts\.v1-[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u,
  );
  await scrollGridCellIntoView({
    cellKey: "host.display_name",
    page,
    recordId: host.record_id,
    surface: hostsViewSchemaId,
  });
  await expect(
    page.getByTestId(rowCellTestId(host.record_id, "host.display_name")),
  ).toHaveText("integration.entity-linking Gateway edited");

  await page.getByTestId(surfaceTabTestId(identitiesViewSchemaId)).click();
  await expectVisibleContractFieldsReachable(
    page,
    identitiesViewSchemaId,
    identityVisibleFields,
  );
  await scrollGridCellIntoView({
    cellKey: "identity.aliases",
    page,
    recordId: identity.record_id,
    surface: identitiesViewSchemaId,
  });
  await expect(
    page.getByTestId(rowCellTestId(identity.record_id, "identity.aliases")),
  ).toContainText("GRIDPROVENANCE Analyst");
  const identityPatch = await patchBodyForEdit(
    page,
    identitiesViewSchemaId,
    identity.record_id,
    "identity.display_name",
    "integration.entity-linking Analyst edited",
  );
  expect(identityPatch).toMatchObject({
    view_schema_id: identitiesViewSchemaId,
    base_row_version: identity.row_version,
    changes: [
      {
        field_key: "identity.display_name",
        value: "integration.entity-linking Analyst edited",
      },
    ],
  });
  expect(String(identityPatch.client_txn_id)).toMatch(
    /^entity-patch-cartulary\.view\.identities\.v1-[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u,
  );
  await scrollGridCellIntoView({
    cellKey: "identity.display_name",
    page,
    recordId: identity.record_id,
    surface: identitiesViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(identity.record_id, "identity.display_name"),
    ),
  ).toHaveText("integration.entity-linking Analyst edited");

  await page.getByTestId(surfaceTabTestId(notesViewSchemaId)).click();
  await expectVisibleContractFieldsReachable(
    page,
    notesViewSchemaId,
    noteVisibleFields,
  );
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
  expect(String(notePatch.client_txn_id)).toMatch(
    /^generic-patch-cartulary\.view\.notes\.v1-[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u,
  );
  await scrollGridCellIntoView({
    cellKey: "note.body",
    page,
    recordId: note.record_id,
    surface: notesViewSchemaId,
  });
  await expect(
    page.getByTestId(rowCellTestId(note.record_id, "note.body")),
  ).toHaveText("Edited provenance note");

  await page.reload();
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await page.getByTestId(surfaceTabTestId(hostsViewSchemaId)).click();
  await scrollGridCellIntoView({
    cellKey: "host.display_name",
    page,
    recordId: host.record_id,
    surface: hostsViewSchemaId,
  });
  await expect(
    page.getByTestId(rowCellTestId(host.record_id, "host.display_name")),
  ).toHaveText("integration.entity-linking Gateway edited");
  await page.getByTestId(surfaceTabTestId(identitiesViewSchemaId)).click();
  await scrollGridCellIntoView({
    cellKey: "identity.display_name",
    page,
    recordId: identity.record_id,
    surface: identitiesViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(identity.record_id, "identity.display_name"),
    ),
  ).toHaveText("integration.entity-linking Analyst edited");
  await page.getByTestId(surfaceTabTestId(notesViewSchemaId)).click();
  await scrollGridCellIntoView({
    cellKey: "note.body",
    page,
    recordId: note.record_id,
    surface: notesViewSchemaId,
  });
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
    mentionFingerprint(
      timelineAfter,
      hostRefsFieldKey,
      " GRIDPROVENANCE Gateway ",
    ),
  ).toEqual(hostMentionBefore);
  expect(
    mentionFingerprint(
      timelineAfter,
      hostRefsFieldKey,
      "GRIDPROVENANCE Unresolved Host",
    ),
  ).toEqual(unresolvedHostMentionBefore);
  expect(
    mentionFingerprint(
      timelineAfter,
      identityRefsFieldKey,
      " GRIDPROVENANCE Analyst ",
    ),
  ).toEqual(identityMentionBefore);
  expect(
    mentionFingerprint(
      timelineAfter,
      identityRefsFieldKey,
      "GRIDPROVENANCE Unresolved Identity",
    ),
  ).toEqual(unresolvedIdentityMentionBefore);
});
