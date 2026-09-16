import type { CreateViewRowResponse } from "@cartulary/protocol-ts/http";
import {
  assessmentCreateControlTestId,
  assessmentCreatePanelTestId,
  gridShellTestId,
  surfaceTabTestId,
  workbookAddRowButtonTestId,
  workbookShellReadyTestId,
  workbookSurfacesMenuOptionTestId,
  workbookSurfacesMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page, TestInfo } from "@playwright/test";
import { expect, test } from "./fixtures";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { fetchFullRecordHistory } from "./support/workbook/history";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { openRecoveryItem, recoveryEntry } from "./support/workbook/recovery";

test("Lost Assessment response replays exactly one append and support set after navigation", async ({
  page,
}, info) => {
  await recovery(page, info, "lost");
});
test("Accepted Assessment refresh recovery sends reads only", async ({
  page,
}, info) => {
  await recovery(page, info, "refresh");
});
test("Late Assessment acceptance preserves navigation focus and retained receipt", async ({
  page,
}, info) => {
  await recovery(page, info, "late");
});

async function recovery(
  page: Page,
  info: TestInfo,
  mode: "lost" | "refresh" | "late",
) {
  await page.setViewportSize({
    width: mode === "lost" ? 768 : 1280,
    height: 800,
  });
  const incident = await createIncident(
    page,
    uniqueIncidentKey("ASSESSMENT-RECOVERY"),
    "Assessment append recovery",
  );
  const subject = await createViewRow(page, incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("subject"),
    "host.display_name": "Reviewed subject",
  });
  const support = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("support"),
    "timeline.activity_synopsis_text": "Reviewed Timeline support",
  });
  const subjectHistory = await fetchFullRecordHistory(page, subject.record_id);
  const supportHistory = await fetchFullRecordHistory(page, support.record_id);
  const requests: string[] = [],
    receipts: CreateViewRowResponse[] = [];
  let failedReads = false,
    release: (() => void) | null = null;
  await page.route(
    `**/views/${assessmentsViewSchemaId}/query`,
    async (route) => {
      if (failedReads) await route.abort("failed");
      else await route.continue();
    },
  );
  await page.route(
    `**/views/${assessmentsViewSchemaId}/rows`,
    async (route) => {
      requests.push(route.request().postData() ?? "");
      const response = await route.fetch();
      expect(response.ok()).toBeTruthy();
      receipts.push(await response.json());
      if (mode === "late")
        await new Promise<void>((resolve) => {
          release = resolve;
        });
      if (mode === "lost" && requests.length === 1) await route.abort("failed");
      else {
        if (mode === "refresh") failedReads = true;
        await route.fulfill({ response });
      }
    },
  );
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${assessmentsViewSchemaId}`,
  );
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await page
    .getByTestId(workbookAddRowButtonTestId(assessmentsViewSchemaId))
    .click();
  await expect(
    page.getByTestId(assessmentCreateControlTestId("subject")),
  ).toHaveValue("");
  await page
    .getByTestId(assessmentCreateControlTestId("subject"))
    .selectOption(subject.record_id);
  await page
    .getByTestId(assessmentCreateControlTestId("rationale"))
    .fill("Reviewed judgment with omitted defaults.");
  await page
    .getByRole("button", { name: "Choose support", exact: true })
    .click();
  await page
    .getByTestId(assessmentCreateControlTestId("support-refs"))
    .selectOption(support.record_id);
  await page
    .getByRole("button", { name: "Apply support selection", exact: true })
    .click();
  const submit = page.getByTestId(assessmentCreateControlTestId("submit"));
  await submit.scrollIntoViewIfNeeded();
  await info.attach("assessment-authoring", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await submit.focus();
  await submit.press("Enter");
  await expect.poll(() => receipts.length).toBe(1);
  const committed = receipts[0];
  if (!committed) throw new Error("Authoritative receipt required.");
  const committedHistory = await fetchFullRecordHistory(
    page,
    committed.data.row.record_id,
  );
  const committedSupportHistory = await fetchFullRecordHistory(
    page,
    support.record_id,
  );
  const newSupportItems = committedSupportHistory.items.filter(
    (item) =>
      !supportHistory.items.some(
        (prior) => prior.history_item_ref === item.history_item_ref,
      ),
  );
  expect(newSupportItems).toHaveLength(1);
  expect(newSupportItems[0]?.change_set_id).toBe(committed.data.change_set_id);
  if (mode !== "refresh") {
    const tab = page.getByTestId(surfaceTabTestId(hostsViewSchemaId));
    if (await tab.isVisible()) await tab.click();
    else {
      await page.getByTestId(workbookSurfacesMenuTriggerTestId()).click();
      await page
        .getByTestId(workbookSurfacesMenuOptionTestId(hostsViewSchemaId))
        .click();
    }
    await expect(
      page.getByTestId(gridShellTestId(hostsViewSchemaId)),
    ).toBeVisible();
    if (mode === "late") await tab.focus();
    if (mode === "late") {
      const complete = release as (() => void) | null;
      complete?.();
      await expect(recoveryEntry(page)).toBeVisible();
      await expect(tab).toBeFocused();
    }
  }
  const trigger = recoveryEntry(page);
  await openRecoveryItem(page, /^Assessment append ·/);
  const retained = page.getByRole("region", {
    name: "Retained Assessment appends",
    exact: true,
  });
  if (mode === "lost") {
    await expect(retained).toContainText("Append result unconfirmed.");
    await info.attach("assessment-uncertain", {
      body: await page.screenshot(),
      contentType: "image/png",
    });
    await retained
      .getByRole("button", { name: "Recover assessment append", exact: true })
      .click();
    await expect.poll(() => receipts.length).toBe(2);
    expect(requests[1]).toBe(requests[0]);
    expect(receipts[1]?.data).toEqual(committed.data);
  }
  await expect(retained).toContainText("Assessment created.");
  await expect(retained).toContainText(committed.data.change_set_id);
  if (mode === "refresh") {
    const retry = retained.getByRole("button", {
      name: "Retry assessment refresh",
      exact: true,
    });
    await expect(retry).toBeEnabled();
    await info.attach("assessment-refresh-required", {
      body: await page.screenshot(),
      contentType: "image/png",
    });
    const writesBefore = [...requests];
    failedReads = false;
    await retry.click();
    await expect(retry).toHaveCount(0);
    expect(requests).toEqual(writesBefore);
  }
  const bounds = await retained.boundingBox();
  expect(bounds?.x).toBeGreaterThanOrEqual(0);
  expect((bounds?.x ?? 0) + (bounds?.width ?? 0)).toBeLessThanOrEqual(
    page.viewportSize()?.width ?? 0,
  );
  await retained.press("Escape");
  await expect(retained).not.toBeVisible();
  await expect(trigger).toBeFocused();
  const rows = await queryViewRows(page, incident, assessmentsViewSchemaId);
  expect(rows).toHaveLength(1);
  expect(rows[0]?.record_id).toBe(committed.data.row.record_id);
  expect(rows[0]?.cells["assessment.supporting_link_count"]?.value).toBe(1);
  expect(rows[0]?.cells["assessment.support_refs"]?.value).toEqual(
    committed.data.row.cells["assessment.support_refs"]?.value,
  );
  expect(
    JSON.stringify(rows[0]?.cells["assessment.support_refs"]?.value),
  ).toContain(support.record_id);
  expect(
    await fetchFullRecordHistory(page, committed.data.row.record_id),
  ).toEqual(committedHistory);
  expect(
    new Set(committedHistory.items.map((item) => item.change_set_id)),
  ).toEqual(new Set([committed.data.change_set_id]));
  expect(await fetchFullRecordHistory(page, subject.record_id)).toEqual(
    subjectHistory,
  );
  expect(await fetchFullRecordHistory(page, support.record_id)).toEqual(
    committedSupportHistory,
  );
  expect(requests).toHaveLength(mode === "lost" ? 2 : 1);
  await info.attach("assessment-recovery-evidence", {
    body: JSON.stringify({
      requests,
      receipts,
      committedHistory,
      committedSupportHistory,
      newSupportItems,
      rows,
    }),
    contentType: "application/json",
  });
}

test("Assessment support rejection preserves the editable draft with no partial append", async ({
  page,
}, info) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("ASSESSMENT-REJECT"),
    "Assessment atomic rejection",
  );
  const subject = await createViewRow(page, incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("subject"),
    "host.display_name": "Rejection subject",
  });
  const support = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("support"),
    "timeline.activity_synopsis_text": "Selected Timeline support",
  });
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${assessmentsViewSchemaId}`,
  );
  await page
    .getByTestId(workbookAddRowButtonTestId(assessmentsViewSchemaId))
    .click();
  await page
    .getByTestId(assessmentCreateControlTestId("subject"))
    .selectOption(subject.record_id);
  await page
    .getByTestId(assessmentCreateControlTestId("rationale"))
    .fill("Keep rejection draft.");
  await page
    .getByRole("button", { name: "Choose support", exact: true })
    .click();
  await page
    .getByTestId(assessmentCreateControlTestId("support-refs"))
    .selectOption(support.record_id);
  await page
    .getByRole("button", { name: "Apply support selection", exact: true })
    .click();
  // Rejection comes from the real service: the retained selection points to another incident.
  const other = await createIncident(
    page,
    uniqueIncidentKey("ASSESSMENT-OTHER"),
    "Other fixture incident",
  );
  const unavailable = await createViewRow(page, other, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("foreign-support"),
    "timeline.activity_synopsis_text": "Foreign support",
  });
  await page.route(
    `**/views/${assessmentsViewSchemaId}/rows`,
    async (route) => {
      const request = route.request().postDataJSON();
      request["assessment.support_refs"].actions.push({
        op: "add_record_ref",
        linked_record_id: unavailable.record_id,
      });
      const response = await route.fetch({ postData: JSON.stringify(request) });
      expect(response.status()).toBe(400);
      await route.fulfill({ response });
    },
  );
  const before = await fetchFullRecordHistory(page, support.record_id);
  await page.getByTestId(assessmentCreateControlTestId("submit")).click();
  await expect(
    page.getByTestId(assessmentCreateControlTestId("message")),
  ).toContainText("invalid_mutation_payload");
  await expect(
    page.getByTestId(assessmentCreateControlTestId("rationale")),
  ).toHaveValue("Keep rejection draft.");
  await expect(
    page.getByRole("region", { name: "Assessment supporting records" }),
  ).toContainText("Selected Timeline support");
  expect(await queryViewRows(page, incident, assessmentsViewSchemaId)).toEqual(
    [],
  );
  expect(await fetchFullRecordHistory(page, support.record_id)).toEqual(before);
  await info.attach("assessment-rejected", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await page.getByTestId(assessmentCreatePanelTestId()).press("Escape");
});
