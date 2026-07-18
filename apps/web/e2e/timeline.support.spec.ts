import {
  applyFilterChip,
  changeGrouping,
  collapseGridGroup,
  expandGridGroup,
  pasteGridMatrix,
  scrollGridCellIntoView,
  sortByHeader,
} from "@cartulary/test-utils/grid";
import {
  gridGroupRowTestId,
  gridRowTestId,
  rowCellTestId,
  saveStateTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowVersionTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { holdBrowserRequest as holdBrowserApiRequest } from "./support/transport/requestInterception";
import {
  assertRecordFieldMutationAnchor,
  fillDownGridCells,
} from "./support/workbook/mutationAnchors";
import { createViewRow } from "./support/workbook/query";
import { clickTimelineRowAction } from "./support/workbook/rowMutations";

const timelineViewSchemaId = "cartulary.view.timeline.v2";

test("Verify one-click focused editing, rejected-transition retention, sort, filter, group, paste, exact-range fill-down, scroll-to-cell, group expand/collapse, and anchor assertions through browser command helpers.", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("S301"),
    "Timeline support query controls",
  );
  const alphaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s301-alpha"),
    "timeline.activity_synopsis_text": "Alpha summary",
  });
  const betaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s301-beta"),
    "timeline.activity_synopsis_text": "Beta summary",
  });
  const clickEditRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("s301-click-edit"),
      "timeline.device_object_text": "Host Alpha",
    },
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: alphaRow.record_id,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(alphaRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("Alpha summary");

  await scrollGridCellIntoView({
    cellKey: "timeline.device_object_text",
    page,
    recordId: clickEditRow.record_id,
    surface: timelineViewSchemaId,
  });
  const deviceCell = page.getByTestId(
    rowCellTestId(clickEditRow.record_id, "timeline.device_object_text"),
  );
  await expect(deviceCell).toHaveText("Host Alpha");
  const clickEditPatches: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/api/v1/records/${clickEditRow.record_id}`)
    ) {
      clickEditPatches.push(request.postData() ?? "");
    }
  });
  await deviceCell.click();
  const deviceEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.device_object_text",
      recordId: clickEditRow.record_id,
      surface: "grid",
    }),
  );
  await expect(deviceEditor).toBeFocused();
  await expect(deviceEditor).toHaveValue("Host Alpha");
  expect(
    await deviceEditor.evaluate((element) => ({
      end: (element as HTMLInputElement).selectionEnd,
      start: (element as HTMLInputElement).selectionStart,
    })),
  ).toEqual({ end: 10, start: 10 });
  await page.keyboard.insertText(" edited");
  await expect(deviceEditor).toHaveValue("Host Alpha edited");
  await deviceEditor.press("Enter");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  expect(clickEditPatches).toHaveLength(1);
  expect(JSON.parse(clickEditPatches[0] ?? "{}")).toMatchObject({
    base_row_version: clickEditRow.row_version,
    changes: [
      {
        field_key: "timeline.device_object_text",
        value: "Host Alpha edited",
      },
    ],
  });

  await clickTimelineRowAction(
    page,
    betaRow.record_id,
    timelineRowMarkReviewedButtonTestId(betaRow.record_id),
  );
  await scrollGridCellIntoView({
    cellKey: "timeline.capture_state",
    page,
    recordId: betaRow.record_id,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(betaRow.record_id, "timeline.capture_state"),
    ),
  ).toHaveText("reviewed");

  const sortRequest = waitForTimelineQuery(page, incidentId);
  await sortByHeader(
    page,
    timelineViewSchemaId,
    "timeline.activity_synopsis_text",
  );
  expect(readPostBody(await sortRequest)).toEqual({
    sort: [{ direction: "asc", field_key: "timeline.activity_synopsis_text" }],
  });

  const filterRequest = waitForTimelineQuery(page, incidentId);
  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.capture_state",
    "reviewed",
  );
  expect(readPostBody(await filterRequest)).toEqual({
    filters: [
      {
        arg: { value: "reviewed" },
        field_key: "timeline.capture_state",
        op: "eq",
      },
    ],
    sort: [{ direction: "asc", field_key: "timeline.activity_synopsis_text" }],
  });
  await expect(
    page.getByTestId(
      rowCellTestId(betaRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("Beta summary");

  const groupRequest = waitForTimelineQuery(page, incidentId);
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
  expect(readPostBody(await groupRequest)).toEqual({
    filters: [
      {
        arg: { value: "reviewed" },
        field_key: "timeline.capture_state",
        op: "eq",
      },
    ],
    group_by: "timeline.capture_state",
    sort: [
      { direction: "asc", field_key: "timeline.capture_state" },
      { direction: "asc", field_key: "timeline.activity_synopsis_text" },
    ],
  });
  const reviewedGroupTestId = gridGroupRowTestId(
    timelineViewSchemaId,
    "timeline.capture_state",
    "reviewed",
  );
  await expect(page.getByTestId(reviewedGroupTestId)).toBeVisible();
  await collapseGridGroup({
    groupTestId: reviewedGroupTestId,
    page,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(betaRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).not.toBeVisible();
  await expandGridGroup({
    groupTestId: reviewedGroupTestId,
    page,
    surface: timelineViewSchemaId,
  });
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: betaRow.record_id,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(betaRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toBeVisible();

  const bulkMutationRequest = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/bulk-mutations`,
        ),
  );
  const bulkMutationResponse = await fillDownGridCells({
    apiBase,
    csrfHeaders: await csrfHeaders(page),
    fieldKey: "timeline.raw_activity_text",
    incidentId,
    page,
    surface: timelineViewSchemaId,
    targetRecords: [
      {
        baseRowVersion: alphaRow.row_version,
        recordId: alphaRow.record_id,
      },
    ],
    value: "Filled details through helper",
  });
  expect(bulkMutationResponse.ok()).toBeTruthy();
  expect(readPostBody(await bulkMutationRequest)).toMatchObject({
    view_schema_id: timelineViewSchemaId,
    kind: "fill_down_v1",
    field_key: "timeline.raw_activity_text",
    value: "Filled details through helper",
    targets: [
      {
        base_row_version: alphaRow.row_version,
        record_id: alphaRow.record_id,
      },
    ],
  });

  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: betaRow.record_id,
    surface: timelineViewSchemaId,
  });
  const summaryPatch = page.waitForRequest(
    (request) =>
      request.method() === "PATCH" &&
      request.url().endsWith(`/api/v1/records/${betaRow.record_id}`),
  );
  const betaSummary = page.getByTestId(
    rowCellTestId(betaRow.record_id, "timeline.activity_synopsis_text"),
  );
  await betaSummary.click();
  const betaSummaryEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId: betaRow.record_id,
      surface: "grid",
    }),
  );
  await betaSummaryEditor.fill("Beta summary anchored");
  await betaSummaryEditor.press("Enter");
  assertRecordFieldMutationAnchor({
    actualRecordId: betaRow.record_id,
    body: readPostBody(await summaryPatch),
    expectedRecordId: betaRow.record_id,
    expectedValue: "Beta summary anchored",
    fieldKey: "timeline.activity_synopsis_text",
  });
  await expect(
    page.getByTestId(
      rowCellTestId(betaRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("Beta summary anchored");

  const pasteRequest = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
        ),
  );
  const pasteResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
        ),
  );
  await pasteGridMatrix({
    fieldKey: "timeline.activity_synopsis_text",
    matrix: [["Beta pasted through helper", "host-token"]],
    page,
    recordId: betaRow.record_id,
    surface: timelineViewSchemaId,
  });
  expect(readPostBody(await pasteRequest)).toMatchObject({
    view_schema_id: timelineViewSchemaId,
    clipboard_text: "Beta pasted through helper\thost-token",
    format: "tsv",
    start_field_key: "timeline.activity_synopsis_text",
    columns: ["timeline.activity_synopsis_text", "timeline.data_source_text"],
    targets: [
      {
        kind: "record",
        record_id: betaRow.record_id,
      },
    ],
  });
  expect((await pasteResponse).ok()).toBeTruthy();
});

test("support keeps a pending edit anchored to its record under sort, filter, group, and live invalidation", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("S302"),
    "Timeline support anchor stability",
  );
  const alphaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s302-alpha"),
    "timeline.activity_synopsis_text": "Alpha summary",
  });
  const betaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s302-beta"),
    "timeline.activity_synopsis_text": "Beta summary",
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await clickTimelineRowAction(
    page,
    betaRow.record_id,
    timelineRowMarkReviewedButtonTestId(betaRow.record_id),
  );
  await expect(
    page.getByTestId(
      rowCellTestId(betaRow.record_id, "timeline.capture_state"),
    ),
  ).toHaveText("reviewed");

  await sortByHeader(
    page,
    timelineViewSchemaId,
    "timeline.activity_synopsis_text",
  );
  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.has_evidence",
    "false",
  );
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");

  const heldPatch = await holdBrowserApiRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${alphaRow.record_id}`,
  });

  try {
    const alphaSummary = page.getByTestId(
      rowCellTestId(alphaRow.record_id, "timeline.activity_synopsis_text"),
    );
    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page,
      recordId: alphaRow.record_id,
      surface: timelineViewSchemaId,
    });
    await alphaSummary.click();
    const alphaSummaryEditor = page.getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.activity_synopsis_text",
        recordId: alphaRow.record_id,
        surface: "grid",
      }),
    );
    await alphaSummaryEditor.fill("Zulu anchored");
    await alphaSummaryEditor.press("Enter");
    await heldPatch.waitForHit;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");

    const betaVersion = Number.parseInt(
      (await page
        .getByTestId(timelineRowVersionTestId(betaRow.record_id))
        .textContent()) ?? "0",
      10,
    );
    const invalidationResponse = await page.request.patch(
      `${apiBase}/api/v1/records/${betaRow.record_id}`,
      {
        headers: await csrfHeaders(page),
        data: {
          view_schema_id: timelineViewSchemaId,
          base_row_version: betaVersion,
          client_txn_id: uniqueTxn("s302-invalidation"),
          changes: [
            {
              field_key: "timeline.raw_activity_text",
              value: "Support invalidation",
            },
          ],
        },
      },
    );
    expect(invalidationResponse.ok()).toBeTruthy();

    expect(heldPatch.hitCount()).toBe(1);
    heldPatch.release();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  } finally {
    await heldPatch.dispose();
  }
  await expect(
    page.getByTestId(timelineRowVersionTestId(alphaRow.record_id)),
  ).toHaveText("2");
  await expect(
    page.getByTestId(
      rowCellTestId(alphaRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("Zulu anchored");
  const alphaCaptureState = (
    (await page
      .getByTestId(rowCellTestId(alphaRow.record_id, "timeline.capture_state"))
      .textContent()) ?? ""
  ).trim();
  expect(alphaCaptureState).not.toBe("");
  await expect(
    page.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        alphaCaptureState,
      ),
    ),
  ).toBeVisible();
});

test("support keeps repeated scalar grid edits out of the RDG measured-width crash path", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("S303"),
    "Timeline support RDG edit stability",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s303-row"),
    "timeline.activity_synopsis_text": "RDG edit row",
  });
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s303-peer"),
    "timeline.activity_synopsis_text": "Alpha RDG peer",
  });
  const recordId = row.record_id as string;

  await page.goto(`/?incident_id=${incidentId}`);
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("RDG edit row");
  await sortByHeader(
    page,
    timelineViewSchemaId,
    "timeline.activity_synopsis_text",
  );
  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.has_evidence",
    "false",
  );
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
  await expect(
    page.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "rough",
      ),
    ),
  ).toBeVisible();

  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId,
    surface: timelineViewSchemaId,
  });

  await expectNoPageCrashDuring(page, async () => {
    const summaryDisplay = page.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    );
    const summaryEditor = page.getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.activity_synopsis_text",
        recordId,
        surface: "grid",
      }),
    );
    await summaryDisplay.click();
    await expect(summaryEditor).toBeFocused();
    for (const value of [
      "RDG edit row patched 1",
      "RDG edit row patched 2",
      "RDG edit row patched 3",
      "RDG edit row final",
    ]) {
      await summaryEditor.fill(value);
      await expect(summaryEditor).toHaveValue(value);
      await expect(
        page.getByTestId(gridRowTestId(timelineViewSchemaId, recordId)),
      ).toBeAttached();
    }

    const patchResponse = page.waitForResponse(
      (response) =>
        response.request().method() === "PATCH" &&
        response.url().endsWith(`/api/v1/records/${recordId}`),
    );
    await summaryEditor.press("Enter");
    await patchResponse;
  });

  await expect(page.getByTestId(timelineRowVersionTestId(recordId))).toHaveText(
    "2",
  );
  await expect(
    page.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("RDG edit row final");
});

function waitForTimelineQuery(page: Page, incidentId: string) {
  return page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
        ),
  );
}

function readPostBody(request: { postData: () => string | null }) {
  return JSON.parse(request.postData() ?? "{}") as Record<string, unknown>;
}

type CrashMonitor = {
  readonly promise: Promise<never>;
  readonly stop: () => void;
};

async function expectNoPageCrashDuring(
  page: Page,
  action: () => Promise<void>,
) {
  const monitors = [rejectOnPageError(page), rejectOnAppRootEmpty(page)];
  try {
    await Promise.race([
      action(),
      ...monitors.map((monitor) => monitor.promise),
    ]);
  } finally {
    for (const monitor of monitors) {
      monitor.stop();
    }
  }
}

function rejectOnPageError(page: Page): CrashMonitor {
  let rejectCrash!: (error: Error) => void;
  const promise = new Promise<never>((_, reject) => {
    rejectCrash = reject;
  });
  const listener = (error: Error) => {
    rejectCrash(error);
  };
  page.on("pageerror", listener);
  return {
    promise,
    stop: () => {
      page.off("pageerror", listener);
    },
  };
}

function rejectOnAppRootEmpty(page: Page): CrashMonitor {
  let stopped = false;
  let timeout: ReturnType<typeof setTimeout> | null = null;
  const promise = new Promise<never>((_, reject) => {
    const checkRoot = () => {
      if (stopped) {
        return;
      }
      void page
        .evaluate(() => {
          const root = document.querySelector("#root");
          return root === null || root.childElementCount === 0;
        })
        .then(
          (isEmpty) => {
            if (stopped) {
              return;
            }
            if (isEmpty) {
              reject(
                new Error(
                  "React app root was removed or emptied during grid edit.",
                ),
              );
              return;
            }
            timeout = setTimeout(checkRoot, 50);
          },
          (error: unknown) => {
            if (!stopped) {
              reject(error instanceof Error ? error : new Error(String(error)));
            }
          },
        );
    };
    checkRoot();
  });

  return {
    promise,
    stop: () => {
      stopped = true;
      if (timeout !== null) {
        clearTimeout(timeout);
      }
    },
  };
}
