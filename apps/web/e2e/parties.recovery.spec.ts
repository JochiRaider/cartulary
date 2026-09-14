import type {
  CreateViewRowResponse,
  PatchRecordResponse,
} from "@cartulary/protocol-ts/http";
import {
  coordinationWorkflowTestId,
  gridShellTestId,
  surfaceTabTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  workbookConflictResolverTestId,
  workbookShellReadyTestId,
  workbookSurfacesMenuOptionTestId,
  workbookSurfacesMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  partiesViewSchemaId,
  taskRequestsViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page, TestInfo } from "@playwright/test";
import { expect, test } from "./fixtures";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { fetchFullRecordHistory } from "./support/workbook/history";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import { openGenericInspectorForRecord } from "./support/workbook/rowMutations";

test("Party creation remains saved after a real source conflict and renewed link review", async ({
  page,
}, info) => {
  await recovery("rejected", page, info);
});
test("Lost Party creation response replays the original result after sheet navigation", async ({
  page,
}, info) => {
  await recovery("create_lost", page, info);
});
test("Lost Evidence source link response replays the exact patch after sheet navigation", async ({
  page,
}, info) => {
  await recovery("link_lost", page, info);
});
test("Accepted Task requester linking recovers refresh failure through reads only", async ({
  page,
}, info) => {
  await recovery("refresh", page, info);
});
test("Party creation returned after row pair and sheet changes requires original source review", async ({
  page,
}, info) => {
  await recovery("navigation", page, info);
});

async function sheet(page: Page, view: string) {
  const tab = page.getByTestId(surfaceTabTestId(view));
  if (await tab.isVisible()) await tab.click();
  else if (
    await page.getByTestId(systemViewSwitcherTriggerTestId()).isVisible()
  ) {
    await page.getByTestId(systemViewSwitcherTriggerTestId()).click();
    await page
      .getByTestId(systemViewSwitcherOptionTestId("coordination", view))
      .click();
  } else {
    await page.getByTestId(workbookSurfacesMenuTriggerTestId()).click();
    await page.getByTestId(workbookSurfacesMenuOptionTestId(view)).click();
  }
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
}

test("Reviewed contextual Party creation reuses unchanged exact matches and reports safe cross-key conflicts", async ({
  page,
}, info) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("PARTY-MATCH"),
    "Reviewed Party matching",
  );
  const first = await createViewRow(page, incident, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("email-party"),
    "party.display_name": "Existing email Party",
    "party.party_kind": "person",
    "party.primary_email": "exact@example.test",
  });
  const second = await createViewRow(page, incident, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("external-party"),
    "party.display_name": "Existing reference Party",
    "party.party_kind": "organization",
    "party.external_ref": "external-second",
  });
  const source = await createViewRow(page, incident, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("source"),
    "evidence.title": "Match review source",
    "evidence.source_party_text": "Source <exact@example.test>",
  });
  const requests: string[] = [],
    receipts: CreateViewRowResponse[] = [];
  await page.route(`**/views/${partiesViewSchemaId}/rows`, async (route) => {
    requests.push(route.request().postData() ?? "");
    const response = await route.fetch();
    if (requests.length === 1) {
      expect(response.status()).toBe(409);
      const error = (await response.json()).error;
      expect(error.code).toBe("party_match_conflict");
      expect(Object.keys(error.details).sort()).toEqual([
        "conflicting_field_keys",
        "reason_code",
      ]);
    } else {
      expect(response.ok()).toBeTruthy();
      receipts.push(await response.json());
    }
    await route.fulfill({ response });
  });
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${evidenceViewSchemaId}`,
  );
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await openGenericInspectorForRecord(
    page,
    evidenceViewSchemaId,
    source.record_id,
  );
  await page
    .getByTestId(coordinationWorkflowTestId("party-pair"))
    .selectOption("evidence.source_party_text:evidence.source_party_id");
  await page
    .getByRole("button", { name: "Create party from text", exact: true })
    .click();
  const form = page.getByRole("form", { name: "Review Party creation" });
  await form
    .getByRole("textbox", { name: "Display Name value", exact: true })
    .fill("Reviewed replacement name");
  await form
    .getByRole("combobox", { name: "Kind value", exact: true })
    .selectOption("team");
  await form.getByRole("checkbox", { name: /Include email proposal/ }).check();
  await expect(
    form.getByRole("textbox", { name: "Email value", exact: true }),
  ).toHaveValue("exact@example.test");
  await form.getByText("Optional Party details", { exact: true }).click();
  await form
    .getByRole("textbox", { name: "External Ref value", exact: true })
    .fill("external-second");
  await form
    .getByRole("button", { name: "Save Party and link", exact: true })
    .click();
  const result = page.getByRole("region", { name: "Party creation result" });
  await expect(result).toContainText("party_match_conflict");
  await expect(result).not.toContainText(first.record_id);
  await expect(result).not.toContainText(second.record_id);
  await expect(
    form.getByRole("textbox", { name: "Display Name value", exact: true }),
  ).toHaveValue("Reviewed replacement name");
  await form
    .getByRole("textbox", { name: "External Ref value", exact: true })
    .fill("");
  await form
    .getByRole("button", { name: "Save Party and link", exact: true })
    .click();
  await expect.poll(() => receipts.length).toBe(1);
  expect(receipts[0]?.data.row).toEqual(first);
  expect(JSON.parse(requests[0] ?? "{}").client_txn_id).not.toBe(
    JSON.parse(requests[1] ?? "{}").client_txn_id,
  );
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, incident, evidenceViewSchemaId))[0]?.cells[
          "evidence.source_party_id"
        ]?.value,
    )
    .toBe(first.record_id);
  expect(
    (await queryViewRows(page, incident, evidenceViewSchemaId))[0]?.cells[
      "evidence.source_party_text"
    ]?.value,
  ).toBe(source.cells["evidence.source_party_text"]?.value);
  expect(
    (await queryViewRows(page, incident, partiesViewSchemaId)).find(
      (row) => row.record_id === first.record_id,
    ),
  ).toEqual(first);
  await info.attach("unchanged-party-reuse", {
    body: JSON.stringify({ requests, receipts }),
    contentType: "application/json",
  });
});

test("Party picker reaches the next authorized page and retains loaded candidates after read failure", async ({
  page,
}, info) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("PARTY-PAGING"),
    "Party candidate paging",
  );
  for (let index = 0; index < 101; index++)
    await createViewRow(page, incident, partiesViewSchemaId, {
      client_txn_id: uniqueTxn("candidate"),
      "party.display_name": `Candidate ${String(index).padStart(3, "0")}`,
      "party.party_kind": "team",
    });
  const source = await createViewRow(page, incident, taskRequestsViewSchemaId, {
    client_txn_id: uniqueTxn("task-source"),
    "task.title": "Paged requester",
    "task.task_kind": "request",
    "task.requester_party_text": "Original requester",
  });
  const cursors: string[] = [];
  let failed = false;
  await page.route(`**/views/${partiesViewSchemaId}/query`, async (route) => {
    const request = route.request().postDataJSON();
    if (request.cursor_token) {
      expect(request.limit).toBe(100);
      cursors.push(request.cursor_token);
      if (!failed) {
        failed = true;
        await route.abort("failed");
        return;
      }
    }
    await route.continue();
  });
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${taskRequestsViewSchemaId}`,
  );
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await openGenericInspectorForRecord(
    page,
    taskRequestsViewSchemaId,
    source.record_id,
  );
  const region = page.getByRole("region", {
    name: "Requester Party",
    exact: true,
  });
  await expect(region).toContainText(
    "100 Parties loaded. More Parties are available.",
  );
  await region
    .getByRole("button", { name: "Load more Parties", exact: true })
    .click();
  await expect(region).toContainText("Previously loaded Parties may be stale.");
  await expect(
    region
      .getByRole("combobox", { name: "Existing party", exact: true })
      .locator("option"),
  ).toHaveCount(101);
  await region
    .getByRole("button", { name: "Retry Party read", exact: true })
    .click();
  await expect(region).toContainText("101 Parties loaded.");
  expect(cursors).toHaveLength(2);
  expect(cursors[0]).toBe(cursors[1]);
  await region
    .getByRole("textbox", { name: "Filter loaded Parties", exact: true })
    .fill("Candidate 100");
  const picker = region.getByRole("combobox", {
    name: "Existing party",
    exact: true,
  });
  await picker.selectOption({ label: "Candidate 100" });
  const target = await picker.inputValue();
  await region
    .getByRole("button", { name: "Link existing party", exact: true })
    .click();
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, incident, taskRequestsViewSchemaId))[0]
          ?.cells["task.requester_party_id"]?.value,
    )
    .toBe(target);
  expect(
    (await queryViewRows(page, incident, taskRequestsViewSchemaId))[0]?.cells[
      "task.requester_party_text"
    ]?.value,
  ).toBe("Original requester");
  await info.attach("party-candidate-pages", {
    body: JSON.stringify({ cursors, target }),
    contentType: "application/json",
  });
});
async function recovery(
  mode: "rejected" | "create_lost" | "link_lost" | "refresh" | "navigation",
  page: Page,
  info: TestInfo,
) {
  await page.setViewportSize({
    width: mode === "create_lost" ? 768 : 1280,
    height: 800,
  });
  const incident = await createIncident(
    page,
    uniqueIncidentKey("PARTY-RECOVERY"),
    "Party source recovery",
  );
  const task = mode === "create_lost" || mode === "refresh";
  const view = task ? taskRequestsViewSchemaId : evidenceViewSchemaId;
  const textField = task
    ? "task.requester_party_text"
    : mode === "link_lost"
      ? "evidence.source_party_text"
      : "evidence.collector_party_text";
  const refField = task
    ? "task.requester_party_id"
    : mode === "link_lost"
      ? "evidence.source_party_id"
      : "evidence.collector_party_id";
  const titleField = task ? "task.title" : "evidence.title";
  const seedText = "  Preserved source <proposal@example.test>  ";
  const source = await createViewRow(page, incident, view, {
    client_txn_id: uniqueTxn("source"),
    [titleField]: "Origin source",
    [textField]: seedText,
    ...(task ? { "task.task_kind": "request" } : {}),
  });
  const sourceText = String(source.cells[textField]?.value);
  const other = await createViewRow(page, incident, view, {
    client_txn_id: uniqueTxn("other-source"),
    [titleField]: "Other source",
    [textField]: "Do not change this row",
    ...(task ? { "task.task_kind": "request" } : {}),
  });
  const competing =
    mode === "rejected"
      ? await createViewRow(page, incident, partiesViewSchemaId, {
          client_txn_id: uniqueTxn("competing-party"),
          "party.display_name": "Remote selected Party",
          "party.party_kind": "organization",
        })
      : null;
  const creates: string[] = [],
    links: string[] = [],
    createReceipts: CreateViewRowResponse[] = [],
    linkReceipts: PatchRecordResponse[] = [];
  let failRefresh = false,
    releaseCreation: (() => void) | null = null;
  await page.route(`**/views/${view}/query`, async (route) => {
    if (failRefresh) await route.abort("failed");
    else await route.continue();
  });
  await page.route(`**/views/${partiesViewSchemaId}/rows`, async (route) => {
    creates.push(route.request().postData() ?? "");
    const response = await route.fetch();
    expect(response.ok()).toBeTruthy();
    createReceipts.push(await response.json());
    if (mode === "navigation")
      await new Promise<void>((resolve) => {
        releaseCreation = resolve;
      });
    if (mode === "create_lost" && creates.length === 1)
      await route.abort("failed");
    else await route.fulfill({ response });
  });
  await page.route(`**/records/${source.record_id}`, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    links.push(route.request().postData() ?? "");
    if (mode === "rejected" && links.length === 1 && competing) {
      await patchRecord(page, source.record_id, {
        base_row_version: source.row_version,
        view_schema_id: view,
        client_txn_id: uniqueTxn("competing-source-write"),
        changes: [{ field_key: refField, value: competing.record_id }],
      });
    }
    const response = await route.fetch();
    if (mode === "rejected" && links.length === 1)
      expect(response.status()).toBe(409);
    else {
      expect(response.status()).toBe(200);
      linkReceipts.push(await response.json());
    }
    if (mode === "link_lost" && links.length === 1) await route.abort("failed");
    else {
      if (mode === "refresh") failRefresh = true;
      await route.fulfill({ response });
    }
  });
  await page.goto(`/?incident_id=${incident}&view_schema_id=${view}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await openGenericInspectorForRecord(page, view, source.record_id);
  await page
    .getByTestId(coordinationWorkflowTestId("party-pair"))
    .selectOption(`${textField}:${refField}`);
  await page
    .getByRole("button", { name: "Create party from text", exact: true })
    .click();
  const form = page.getByRole("form", { name: "Review Party creation" });
  await expect(
    form.getByRole("textbox", { name: "Display Name value", exact: true }),
  ).toHaveValue(sourceText);
  await expect(
    form.getByRole("combobox", { name: "Kind value", exact: true }),
  ).toHaveValue("");
  await expect(form.getByRole("checkbox")).not.toBeChecked();
  await expect(
    form.getByRole("button", { name: "Save Party and link" }),
  ).toBeDisabled();
  await form
    .getByRole("textbox", { name: "Display Name value", exact: true })
    .fill("Reviewed Party");
  await form
    .getByRole("combobox", { name: "Kind value", exact: true })
    .selectOption("team");
  await form
    .getByRole("button", { name: "Save Party and link" })
    .scrollIntoViewIfNeeded();
  await info.attach("reviewed-party-authoring", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await form.getByRole("button", { name: "Save Party and link" }).focus();
  await form
    .getByRole("button", { name: "Save Party and link" })
    .press("Enter");
  await expect.poll(() => createReceipts.length).toBe(1);
  if (mode === "navigation") {
    await expect.poll(() => releaseCreation !== null).toBeTruthy();
    await openGenericInspectorForRecord(page, view, other.record_id);
    await page
      .getByTestId(coordinationWorkflowTestId("party-pair"))
      .selectOption("evidence.source_party_text:evidence.source_party_id");
  } else if (mode !== "create_lost")
    await expect.poll(() => links.length).toBe(1);
  if (mode === "rejected") {
    const resolver = page.getByTestId(workbookConflictResolverTestId());
    await expect(resolver).toBeVisible();
    await resolver
      .getByRole("button", { name: "Keep saved", exact: true })
      .click();
    await expect(resolver).not.toBeVisible();
  }
  const createdId = createReceipts[0]?.data.row.record_id;
  if (!createdId) throw new Error("Authoritative creation receipt required.");
  const createdHistory = await fetchFullRecordHistory(page, createdId);
  const sourceHistory = await fetchFullRecordHistory(page, source.record_id);
  await sheet(page, partiesViewSchemaId);
  if (mode === "navigation") {
    const release = releaseCreation as (() => void) | null;
    release?.();
  }
  const trigger = page.getByText(/^Party operations \(/);
  await trigger.click();
  const retained = page.getByRole("region", {
    name: "Retained Party operations",
  });
  if (mode === "create_lost") {
    await expect(retained).toContainText(
      "Party creation outcome is uncertain.",
    );
    await retained
      .getByRole("button", { name: "Replay Party creation", exact: true })
      .click();
    await expect.poll(() => createReceipts.length).toBe(2);
    expect(creates[1]).toBe(creates[0]);
    expect(createReceipts[1]?.data).toEqual(createReceipts[0]?.data);
  }
  await expect(retained).toContainText("Party saved: Reviewed Party");
  if (mode === "link_lost") {
    await expect(retained).toContainText("Source change outcome is uncertain.");
    await retained
      .getByRole("button", {
        name: "Replay original source change",
        exact: true,
      })
      .click();
    await expect.poll(() => linkReceipts.length).toBe(2);
    expect(links[1]).toBe(links[0]);
    expect(linkReceipts[1]?.data).toEqual(linkReceipts[0]?.data);
  }
  if (mode === "refresh") {
    await expect(retained).toContainText(
      "Source change saved; refresh is still required.",
    );
    failRefresh = false;
    await retained
      .getByRole("button", { name: "Refresh source result", exact: true })
      .click();
    await expect(retained).toContainText("Source change saved.");
  }
  expect(await fetchFullRecordHistory(page, createdId)).toEqual(createdHistory);
  if (mode === "link_lost" || mode === "refresh")
    expect(await fetchFullRecordHistory(page, source.record_id)).toEqual(
      sourceHistory,
    );
  await info.attach("retained-party-recovery", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  const bounds = await retained.boundingBox();
  expect(bounds?.x).toBeGreaterThanOrEqual(0);
  expect((bounds?.x ?? 0) + (bounds?.width ?? 0)).toBeLessThanOrEqual(
    page.viewportSize()?.width ?? 0,
  );
  await retained.press("Escape");
  await expect(retained).not.toBeVisible();
  await expect(trigger).toBeFocused();
  if (["create_lost", "navigation", "rejected"].includes(mode)) {
    if (mode !== "rejected") expect(links).toHaveLength(0);
    await sheet(page, view);
    await openGenericInspectorForRecord(page, view, source.record_id);
    await page
      .getByTestId(coordinationWorkflowTestId("party-pair"))
      .selectOption(`${textField}:${refField}`);
    await page
      .getByRole("button", {
        name: "Link saved Party to this source",
        exact: true,
      })
      .click();
    await expect.poll(() => linkReceipts.length).toBe(1);
  }
  const sourceRows = await queryViewRows(page, incident, view);
  expect(
    sourceRows.find((row) => row.record_id === source.record_id)?.cells[
      textField
    ]?.value,
  ).toBe(sourceText);
  expect(
    sourceRows.find((row) => row.record_id === source.record_id)?.cells[
      refField
    ]?.value,
  ).toBe(createReceipts[0]?.data.row.record_id);
  expect(
    sourceRows.find((row) => row.record_id === other.record_id)?.cells[refField]
      ?.value,
  ).toBeNull();
  expect(
    sourceRows.find((row) => row.record_id === other.record_id)?.row_version,
  ).toBe(other.row_version);
  const parties = await queryViewRows(page, incident, partiesViewSchemaId);
  expect(parties).toHaveLength(competing ? 2 : 1);
  const saved = parties.find(
    (row) => row.record_id === createReceipts[0]?.data.row.record_id,
  );
  expect(saved?.cells["party.primary_email"]?.value).toBeNull();
  expect(saved?.row_version).toBe(1);
  expect(creates).toHaveLength(mode === "create_lost" ? 2 : 1);
  expect(links).toHaveLength(
    mode === "link_lost" || mode === "rejected" ? 2 : 1,
  );
  await info.attach("party-recovery-requests-and-receipts", {
    body: JSON.stringify({ creates, links, createReceipts, linkReceipts }),
    contentType: "application/json",
  });
}
