import type { ViewRow } from "@cartulary/protocol-ts/http";

import {
  coordinationWorkflowTestId,
  dataTestIdSelector,
  genericEditActionSelectTestId,
  genericEditValueTestId,
  gridScrollportSelector,
  gridShellTestId,
  rowCellTestId,
  workbookFocusAnchorTestId,
} from "@cartulary/ui-contracts";
import {
  commLogViewSchemaId,
  evidenceViewSchemaId,
  partiesViewSchemaId,
  taskRequestsViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page, Request } from "@playwright/test";
import { expect, test } from "./fixtures";

import { createIncident } from "./support/incidents/fixtures";

import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  activateSemanticGridCell,
  expectWorkbookSaved,
  openGenericSurface,
} from "./support/workbook/authoring";
import {
  createViewRow,
  queryViewRows,
  waitForViewRowByCell,
} from "./support/workbook/query";
import { selectReferenceCandidates } from "./support/workbook/references";
import {
  editGenericCell,
  submitGenericEditAndWait,
} from "./support/workbook/rowMutations";

async function setGenericGridScroll(page: Page, viewSchemaId: string) {
  return page
    .locator(
      `${dataTestIdSelector(gridShellTestId(viewSchemaId))} ${gridScrollportSelector()}`,
    )
    .evaluate((element) => {
      element.scrollTop = 16;
      element.scrollLeft = 32;
      return { top: element.scrollTop, left: element.scrollLeft };
    });
}

async function expectGenericGridScroll(
  page: Page,
  viewSchemaId: string,
  expected: { top: number; left: number },
) {
  await expect
    .poll(async () =>
      page
        .locator(
          `${dataTestIdSelector(gridShellTestId(viewSchemaId))} ${gridScrollportSelector()}`,
        )
        .evaluate((element) => ({
          top: element.scrollTop,
          left: element.scrollLeft,
        })),
    )
    .toEqual(expected);
}

async function applyPartyPatchAndWait(
  page: Page,
  expectation: PartyPatchExpectation,
): Promise<ViewRow> {
  const isTargetPatchRequest = (request: Request) =>
    request.method() === "PATCH" &&
    request.url().endsWith(`/api/v1/records/${expectation.recordId}`);
  const patchRequests: Request[] = [];
  const capturePatchRequest = (request: Request) => {
    if (isTargetPatchRequest(request)) {
      patchRequests.push(request);
    }
  };
  page.on("request", capturePatchRequest);
  try {
    const [response] = await Promise.all([
      page.waitForResponse((candidate) =>
        isTargetPatchRequest(candidate.request()),
      ),
      page.getByTestId(expectation.clickTestId).click(),
    ]);

    expect(patchRequests).toHaveLength(1);
    expect(patchRequests[0]?.postDataJSON()).toMatchObject({
      view_schema_id: expectation.viewSchemaId,
      changes: expectation.changes,
    });
    expect(response.ok()).toBeTruthy();
    const body = (await response.json()) as { data: { row: ViewRow } };
    expectPartyPatchCells(body.data.row, expectation.expectedCells);
    const row = await waitForAuthoritativeViewRowCells(
      page,
      expectation.incidentId,
      expectation.viewSchemaId,
      expectation.recordId,
      expectation.expectedCells,
    );
    await expectWorkbookSaved(page);
    return row;
  } finally {
    page.off("request", capturePatchRequest);
  }
}

type PartyPatchExpectation = {
  changes: PartyPatchChange[];
  clickTestId: string;
  expectedCells: Record<string, unknown>;
  incidentId: string;
  recordId: string;
  viewSchemaId: string;
};

type PartyPatchChange = {
  field_key: string;
  value: unknown;
};

function expectPartyPatchCells(
  row: ViewRow,
  expectedCells: Record<string, unknown>,
) {
  for (const [fieldKey, expected] of Object.entries(expectedCells)) {
    expect(row.cells[fieldKey]?.value).toEqual(expected);
  }
}

async function waitForAuthoritativeViewRowCells(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  recordId: string,
  expectedCells: Record<string, unknown>,
): Promise<ViewRow> {
  let matched: ViewRow | null = null;
  await expect
    .poll(
      async () => {
        const rows = await queryViewRows(page, incidentId, viewSchemaId);
        matched = rows.find((row) => row.record_id === recordId) ?? null;
        if (matched === null) {
          return null;
        }
        return Object.fromEntries(
          Object.keys(expectedCells).map((fieldKey) => [
            fieldKey,
            matched?.cells[fieldKey]?.value,
          ]),
        );
      },
      {
        message: `${viewSchemaId} row ${recordId} should round-trip party fields`,
      },
    )
    .toEqual(expectedCells);
  if (matched === null) {
    throw new Error(`missing ${viewSchemaId} row ${recordId}`);
  }
  return matched;
}

function collectionItemRefs(value: unknown): string[] {
  if (!value || typeof value !== "object" || !("items" in value)) {
    return [];
  }
  const items = (value as { items?: unknown }).items;
  if (!Array.isArray(items)) {
    return [];
  }
  return items.flatMap((item) => {
    if (!item || typeof item !== "object") {
      return [];
    }
    const record = item as Record<string, unknown>;
    return typeof record.item_ref === "string" ? [record.item_ref] : [];
  });
}

test("Party create and link preserve raw text on the workbook surface", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("PARTY-LINKING"),
    "Workbook inspector party-linking party create link",
  );
  const evidence = await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("e904-evidence"),
    "evidence.title": "Workbook inspector party evidence",
    "evidence.collector_party_text": "Browser Collector Raw",
    "evidence.source_party_text": "Browser Source Raw",
  });
  const existingParty = await createViewRow(
    page,
    incidentId,
    partiesViewSchemaId,
    {
      client_txn_id: uniqueTxn("e904-existing-party"),
      "party.display_name": "Existing Source Party",
      "party.party_kind": "organization",
    },
  );

  await openGenericSurface(page, incidentId, evidenceViewSchemaId);
  const typedCollectorText = "Browser Collector Raw <collector@example.test>";
  await editGenericCell(
    page,
    evidenceViewSchemaId,
    evidence.record_id as string,
    "evidence.collector_party_text",
    typedCollectorText,
  );
  await expectWorkbookSaved(page);
  let rows = await queryViewRows(page, incidentId, evidenceViewSchemaId);
  let refreshedEvidence = rows.find(
    (row) => row.record_id === evidence.record_id,
  );
  expect(refreshedEvidence?.cells["evidence.collector_party_text"]?.value).toBe(
    typedCollectorText,
  );
  expect(
    refreshedEvidence?.cells["evidence.collector_party_id"]?.value,
  ).toBeNull();
  await activateSemanticGridCell(
    page.getByTestId(
      rowCellTestId(evidence.record_id as string, "evidence.title"),
    ),
  );
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${evidenceViewSchemaId}:${evidence.record_id}:evidence.title`,
  );
  const preservedScroll = await setGenericGridScroll(
    page,
    evidenceViewSchemaId,
  );
  const assertEvidenceContextStable = async () => {
    await expect(
      page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
    ).toBeVisible();
    await expect(page).toHaveURL(
      new RegExp(`view_schema_id=${encodeURIComponent(evidenceViewSchemaId)}`),
    );
    await expect(
      page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
    ).toBeVisible();
    await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
      `${evidenceViewSchemaId}:${evidence.record_id}:evidence.title`,
    );
    await expectGenericGridScroll(page, evidenceViewSchemaId, preservedScroll);
  };
  await page
    .getByTestId(coordinationWorkflowTestId("party-pair"))
    .selectOption("evidence.collector_party_text:evidence.collector_party_id");
  await page
    .getByTestId(coordinationWorkflowTestId("party-create-from-text"))
    .click();
  await page
    .getByRole("combobox", { name: "Kind value", exact: true })
    .selectOption("team");
  await page
    .getByRole("button", { name: "Save Party and link", exact: true })
    .click();
  const createdParty = await waitForViewRowByCell(
    page,
    incidentId,
    partiesViewSchemaId,
    "party.display_name",
    typedCollectorText,
  );
  await expectWorkbookSaved(page);
  await assertEvidenceContextStable();

  rows = await queryViewRows(page, incidentId, evidenceViewSchemaId);
  refreshedEvidence = rows.find((row) => row.record_id === evidence.record_id);
  expect(refreshedEvidence?.cells["evidence.collector_party_text"]?.value).toBe(
    typedCollectorText,
  );
  expect(refreshedEvidence?.cells["evidence.collector_party_id"]?.value).toBe(
    createdParty.record_id,
  );

  await page
    .getByTestId(coordinationWorkflowTestId("party-pair"))
    .selectOption("evidence.source_party_text:evidence.source_party_id");
  await page
    .getByTestId(coordinationWorkflowTestId("party-existing"))
    .locator(`input[type="radio"][value="${existingParty.record_id}"]`)
    .check();
  refreshedEvidence = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: evidenceViewSchemaId,
    recordId: evidence.record_id as string,
    clickTestId: coordinationWorkflowTestId("party-link-existing"),
    changes: [
      {
        field_key: "evidence.source_party_id",
        value: existingParty.record_id,
      },
    ],
    expectedCells: {
      "evidence.source_party_text": "Browser Source Raw",
      "evidence.source_party_id": existingParty.record_id,
    },
  });
  await assertEvidenceContextStable();
  expect(refreshedEvidence?.cells["evidence.source_party_text"]?.value).toBe(
    "Browser Source Raw",
  );
  expect(refreshedEvidence?.cells["evidence.source_party_id"]?.value).toBe(
    existingParty.record_id,
  );

  refreshedEvidence = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: evidenceViewSchemaId,
    recordId: evidence.record_id as string,
    clickTestId: coordinationWorkflowTestId("party-clear-link"),
    changes: [{ field_key: "evidence.source_party_id", value: null }],
    expectedCells: {
      "evidence.source_party_text": "Browser Source Raw",
      "evidence.source_party_id": null,
    },
  });
  await assertEvidenceContextStable();
  expect(refreshedEvidence?.cells["evidence.source_party_text"]?.value).toBe(
    "Browser Source Raw",
  );
  expect(
    refreshedEvidence?.cells["evidence.source_party_id"]?.value,
  ).toBeNull();

  await page
    .getByTestId(coordinationWorkflowTestId("party-existing"))
    .locator(`input[type="radio"][value="${existingParty.record_id}"]`)
    .check();
  refreshedEvidence = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: evidenceViewSchemaId,
    recordId: evidence.record_id as string,
    clickTestId: coordinationWorkflowTestId("party-link-existing"),
    changes: [
      {
        field_key: "evidence.source_party_id",
        value: existingParty.record_id,
      },
    ],
    expectedCells: {
      "evidence.source_party_text": "Browser Source Raw",
      "evidence.source_party_id": existingParty.record_id,
    },
  });
  await assertEvidenceContextStable();
  refreshedEvidence = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: evidenceViewSchemaId,
    recordId: evidence.record_id as string,
    clickTestId: coordinationWorkflowTestId("party-clear-text"),
    changes: [{ field_key: "evidence.source_party_text", value: null }],
    expectedCells: {
      "evidence.source_party_text": null,
      "evidence.source_party_id": existingParty.record_id,
    },
  });
  await assertEvidenceContextStable();
  expect(
    refreshedEvidence?.cells["evidence.source_party_text"]?.value,
  ).toBeNull();
  expect(refreshedEvidence?.cells["evidence.source_party_id"]?.value).toBe(
    existingParty.record_id,
  );

  await page
    .getByTestId(coordinationWorkflowTestId("party-existing"))
    .locator(`input[type="radio"][value="${existingParty.record_id}"]`)
    .check();
  refreshedEvidence = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: evidenceViewSchemaId,
    recordId: evidence.record_id as string,
    clickTestId: coordinationWorkflowTestId("party-clear-both"),
    changes: [
      { field_key: "evidence.source_party_text", value: null },
      { field_key: "evidence.source_party_id", value: null },
    ],
    expectedCells: {
      "evidence.source_party_text": null,
      "evidence.source_party_id": null,
    },
  });
  await assertEvidenceContextStable();
  expect(
    refreshedEvidence?.cells["evidence.source_party_text"]?.value,
  ).toBeNull();
  expect(
    refreshedEvidence?.cells["evidence.source_party_id"]?.value,
  ).toBeNull();

  await expectWorkbookSaved(page);
  await assertEvidenceContextStable();

  const task = await createViewRow(page, incidentId, taskRequestsViewSchemaId, {
    client_txn_id: uniqueTxn("e904-task"),
    "task.title": "Workbook inspector party task",
    "task.task_kind": "request",
    "task.requester_party_text": "Browser Requester Raw",
  });
  await openGenericSurface(page, incidentId, taskRequestsViewSchemaId);
  const typedRequesterText = "Browser Requester Raw <requester@example.test>";
  await editGenericCell(
    page,
    taskRequestsViewSchemaId,
    task.record_id as string,
    "task.requester_party_text",
    typedRequesterText,
  );
  await expectWorkbookSaved(page);
  rows = await queryViewRows(page, incidentId, taskRequestsViewSchemaId);
  let refreshedTask = rows.find((row) => row.record_id === task.record_id);
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBe(
    typedRequesterText,
  );
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBeNull();
  await activateSemanticGridCell(
    page.getByTestId(rowCellTestId(task.record_id as string, "task.title")),
  );
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${taskRequestsViewSchemaId}:${task.record_id}:task.title`,
  );
  const taskScroll = await setGenericGridScroll(page, taskRequestsViewSchemaId);
  const assertTaskContextStable = async () => {
    await expect(
      page.getByTestId(gridShellTestId(taskRequestsViewSchemaId)),
    ).toBeVisible();
    await expect(page).toHaveURL(
      new RegExp(
        `view_schema_id=${encodeURIComponent(taskRequestsViewSchemaId)}`,
      ),
    );
    await expect(
      page.getByTestId(gridShellTestId(taskRequestsViewSchemaId)),
    ).toBeVisible();
    await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
      `${taskRequestsViewSchemaId}:${task.record_id}:task.title`,
    );
    await expectGenericGridScroll(page, taskRequestsViewSchemaId, taskScroll);
  };

  await page
    .getByTestId(coordinationWorkflowTestId("party-pair"))
    .selectOption("task.requester_party_text:task.requester_party_id");
  await page
    .getByTestId(coordinationWorkflowTestId("party-create-from-text"))
    .click();
  await page
    .getByRole("combobox", { name: "Kind value", exact: true })
    .selectOption("team");
  await page
    .getByRole("button", { name: "Save Party and link", exact: true })
    .click();
  const createdRequester = await waitForViewRowByCell(
    page,
    incidentId,
    partiesViewSchemaId,
    "party.display_name",
    typedRequesterText,
  );
  await expectWorkbookSaved(page);
  await assertTaskContextStable();
  rows = await queryViewRows(page, incidentId, taskRequestsViewSchemaId);
  refreshedTask = rows.find((row) => row.record_id === task.record_id);
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBe(
    typedRequesterText,
  );
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBe(
    createdRequester.record_id,
  );

  await page
    .getByTestId(coordinationWorkflowTestId("party-existing"))
    .locator(`input[type="radio"][value="${existingParty.record_id}"]`)
    .check();
  refreshedTask = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: taskRequestsViewSchemaId,
    recordId: task.record_id as string,
    clickTestId: coordinationWorkflowTestId("party-link-existing"),
    changes: [
      {
        field_key: "task.requester_party_id",
        value: existingParty.record_id,
      },
    ],
    expectedCells: {
      "task.requester_party_text": typedRequesterText,
      "task.requester_party_id": existingParty.record_id,
    },
  });
  await assertTaskContextStable();
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBe(
    typedRequesterText,
  );
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBe(
    existingParty.record_id,
  );

  refreshedTask = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: taskRequestsViewSchemaId,
    recordId: task.record_id as string,
    clickTestId: coordinationWorkflowTestId("party-clear-link"),
    changes: [{ field_key: "task.requester_party_id", value: null }],
    expectedCells: {
      "task.requester_party_text": typedRequesterText,
      "task.requester_party_id": null,
    },
  });
  await assertTaskContextStable();
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBe(
    typedRequesterText,
  );
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBeNull();

  await page
    .getByTestId(coordinationWorkflowTestId("party-existing"))
    .locator(`input[type="radio"][value="${existingParty.record_id}"]`)
    .check();
  refreshedTask = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: taskRequestsViewSchemaId,
    recordId: task.record_id as string,
    clickTestId: coordinationWorkflowTestId("party-link-existing"),
    changes: [
      {
        field_key: "task.requester_party_id",
        value: existingParty.record_id,
      },
    ],
    expectedCells: {
      "task.requester_party_text": typedRequesterText,
      "task.requester_party_id": existingParty.record_id,
    },
  });
  await assertTaskContextStable();
  refreshedTask = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: taskRequestsViewSchemaId,
    recordId: task.record_id as string,
    clickTestId: coordinationWorkflowTestId("party-clear-text"),
    changes: [{ field_key: "task.requester_party_text", value: null }],
    expectedCells: {
      "task.requester_party_text": null,
      "task.requester_party_id": existingParty.record_id,
    },
  });
  await assertTaskContextStable();
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBeNull();
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBe(
    existingParty.record_id,
  );

  await page
    .getByTestId(coordinationWorkflowTestId("party-existing"))
    .locator(`input[type="radio"][value="${existingParty.record_id}"]`)
    .check();
  refreshedTask = await applyPartyPatchAndWait(page, {
    incidentId,
    viewSchemaId: taskRequestsViewSchemaId,
    recordId: task.record_id as string,
    clickTestId: coordinationWorkflowTestId("party-clear-both"),
    changes: [
      { field_key: "task.requester_party_text", value: null },
      { field_key: "task.requester_party_id", value: null },
    ],
    expectedCells: {
      "task.requester_party_text": null,
      "task.requester_party_id": null,
    },
  });
  await assertTaskContextStable();
  expect(refreshedTask?.cells["task.requester_party_text"]?.value).toBeNull();
  expect(refreshedTask?.cells["task.requester_party_id"]?.value).toBeNull();

  const commLog = await createViewRow(page, incidentId, commLogViewSchemaId, {
    client_txn_id: uniqueTxn("e904-comm"),
    "comm_log.comm_type": "briefing",
    "comm_log.audience": "Browser Leadership Audience",
    "comm_log.channel_or_meeting": "Bridge",
    "comm_log.summary": "Workbook inspector party comm log",
  });
  await openGenericSurface(page, incidentId, commLogViewSchemaId);
  await activateSemanticGridCell(
    page.getByTestId(
      rowCellTestId(commLog.record_id as string, "comm_log.summary"),
    ),
  );
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${commLogViewSchemaId}:${commLog.record_id}:comm_log.summary`,
  );
  const commScroll = await setGenericGridScroll(page, commLogViewSchemaId);
  const assertCommContextStable = async () => {
    await expect(
      page.getByTestId(gridShellTestId(commLogViewSchemaId)),
    ).toBeVisible();
    await expect(page).toHaveURL(
      new RegExp(`view_schema_id=${encodeURIComponent(commLogViewSchemaId)}`),
    );
    await expect(
      page.getByTestId(gridShellTestId(commLogViewSchemaId)),
    ).toBeVisible();
    await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
      `${commLogViewSchemaId}:${commLog.record_id}:comm_log.summary`,
    );
    await expectGenericGridScroll(page, commLogViewSchemaId, commScroll);
  };
  const addAndRemoveCommPartyRef = async (fieldKey: string) => {
    await page.locator(`[data-inspector-edit-field="${fieldKey}"]`).click();
    await page
      .getByTestId(genericEditActionSelectTestId(commLogViewSchemaId))
      .selectOption("add");
    await selectReferenceCandidates(
      page,
      genericEditValueTestId(commLogViewSchemaId),
      existingParty.record_id as string,
    );
    await submitGenericEditAndWait(
      page,
      commLogViewSchemaId,
      commLog.record_id as string,
    );
    await assertCommContextStable();

    rows = await queryViewRows(page, incidentId, commLogViewSchemaId);
    let refreshedComm = rows.find((row) => row.record_id === commLog.record_id);
    expect(refreshedComm?.cells["comm_log.audience"]?.value).toBe(
      "Browser Leadership Audience",
    );
    expect(collectionItemRefs(refreshedComm?.cells[fieldKey]?.value)).toContain(
      `party_ref:${existingParty.record_id}`,
    );

    await page
      .getByTestId(genericEditActionSelectTestId(commLogViewSchemaId))
      .selectOption("remove");
    await page
      .getByTestId(genericEditValueTestId(commLogViewSchemaId))
      .selectOption(`party_ref:${existingParty.record_id}`);
    await submitGenericEditAndWait(
      page,
      commLogViewSchemaId,
      commLog.record_id as string,
    );
    await assertCommContextStable();

    rows = await queryViewRows(page, incidentId, commLogViewSchemaId);
    refreshedComm = rows.find((row) => row.record_id === commLog.record_id);
    expect(refreshedComm?.cells["comm_log.audience"]?.value).toBe(
      "Browser Leadership Audience",
    );
    expect(
      collectionItemRefs(refreshedComm?.cells[fieldKey]?.value),
    ).not.toContain(`party_ref:${existingParty.record_id}`);
  };

  await addAndRemoveCommPartyRef("comm_log.audience_party_ids");
  await addAndRemoveCommPartyRef("comm_log.attendee_party_ids");
});
