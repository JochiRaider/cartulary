import { Buffer } from "node:buffer";
import type { CreateManualIndicatorObservationResponse } from "@cartulary/protocol-ts/http";
import {
  indicatorObservationTestId,
  rowHistoryActionTestId,
  rowHistoryRollbackConfirmButtonTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import {
  indicatorsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { apiBase } from "./support/runtime/configuration";
import { uniqueTxn } from "./support/runtime/fixtureIdentity";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import {
  fetchFullRecordHistory,
  openHistoryEventDetails,
} from "./support/workbook/history";
import {
  createObservationFixture,
  listSourceObservations,
  observationPrefix,
  observationRawText,
  observationServiceSnapshot,
  openObservationEditor,
  selectRepeatedObservation,
} from "./support/workbook/indicatorObservations";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";

test("Indicator observations replay source capture and every transition after commit without duplicate history", async ({
  page,
}, testInfo) => {
  const fixture = await createObservationFixture(page),
    { incidentId, source, oldTarget, newTarget } = fixture;
  const requests: string[] = [],
    receipts: CreateManualIndicatorObservationResponse["data"][] = [];
  await page.route(
    `**/api/v1/records/${source.record_id}/indicator-observations`,
    async (route) => {
      if (route.request().method() !== "POST") return route.continue();
      requests.push(route.request().postData() ?? "");
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      receipts.push((await response.json()).data);
      if (requests.length === 1) await route.abort("failed");
      else await route.fulfill({ response });
    },
  );
  await page.goto(`/?incident_id=${incidentId}`);
  let editor = await openObservationEditor(page, source.record_id);
  await selectRepeatedObservation(page);
  await editor
    .getByLabel("Existing Indicator", { exact: true })
    .selectOption(oldTarget.record_id);
  await editor
    .getByRole("button", { name: "Create observation", exact: true })
    .click();
  await expect(
    editor.getByText(/The observation outcome is unknown/),
  ).toBeVisible();
  const before = await observationServiceSnapshot(page, fixture);
  expect(before.observations).toHaveLength(1);
  expect(before.oldTarget).toHaveLength(1);
  expect(before.newTarget).toHaveLength(0);
  expect(
    Buffer.from(
      String(before.source?.cells["timeline.raw_activity_text"]?.value),
    ),
  ).toEqual(Buffer.from(observationRawText));
  const request = JSON.parse(requests[0] ?? "{}");
  expect(request).toEqual({
    client_txn_id: expect.any(String),
    base_row_version: source.row_version,
    source_field_key: "timeline.raw_activity_text",
    span_start_byte: Buffer.byteLength(observationPrefix),
    span_end_byte: Buffer.byteLength(observationPrefix) + 13,
    resolved_indicator_record_id: oldTarget.record_id,
  });
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  let failRefresh = true;
  await page.route(`**/views/${timelineViewSchemaId}/query`, async (route) => {
    if (!failRefresh) return route.continue();
    failRefresh = false;
    await route.fulfill({
      status: 503,
      contentType: "application/json",
      body: JSON.stringify({
        error: {
          code: "service_unavailable",
          details: {},
          status: 503,
          retryable: true,
          message: "Refresh unavailable",
          request_id: "refresh-failure",
        },
      }),
    });
  });
  await openRecoveryItem(page, /^Indicator observation ·/);
  const recovery = page.getByTestId(indicatorObservationTestId("recovery"));
  await recovery
    .getByRole("button", { name: "Replay original observation request" })
    .click();
  await expect(
    recovery.getByText("Observation change saved. Refresh is still required.", {
      exact: true,
    }),
  ).toBeVisible();
  expect(requests).toHaveLength(2);
  expect(requests[1]).toBe(requests[0]);
  expect(receipts[1]).toEqual({ ...receipts[0], replayed: true });
  await recovery
    .getByRole("button", { name: "Retry observation refresh" })
    .click();
  await expect(
    recovery.getByText(
      "Observation change saved. Records and history refreshed.",
      { exact: true },
    ),
  ).toBeVisible();
  expect(requests).toHaveLength(2);
  expect(await observationServiceSnapshot(page, fixture)).toEqual(before);
  await page.getByRole("button", { name: "Close recovery" }).click();
  editor = await openObservationEditor(page, source.record_id);
  const observationId = before.observations[0]?.observation_id;
  if (!observationId) throw new Error("observation missing");
  const transitionEvidence = [];
  for (const action of ["resolve", "dismiss", "restore"] as const) {
    const bodies: string[] = [],
      results: CreateManualIndicatorObservationResponse["data"][] = [];
    await page.route(
      `**/api/v1/indicator-observations/${observationId}/${action}`,
      async (route) => {
        bodies.push(route.request().postData() ?? "");
        const response = await route.fetch();
        expect(response.status()).toBe(200);
        results.push((await response.json()).data);
        if (bodies.length === 1) await route.abort("failed");
        else await route.fulfill({ response });
      },
    );
    const previous = (await listSourceObservations(page, source.record_id)).data
      .observations[0];
    if (!previous) throw new Error("previous observation missing");
    const item = editor.getByRole("article", {
      name: "Observation: alpha.example",
      exact: true,
    });
    if (action === "resolve") {
      await item.getByRole("button", { name: "Reassign", exact: true }).click();
      await item
        .getByLabel("Existing Indicator", { exact: true })
        .selectOption(newTarget.record_id);
      await item
        .getByRole("button", { name: "Reassign observation", exact: true })
        .click();
    } else
      await item
        .getByRole("button", {
          name:
            action === "dismiss"
              ? "Dismiss observation"
              : "Restore observation",
          exact: true,
        })
        .click();
    await expect(
      editor.getByText(/The observation outcome is unknown/),
    ).toBeVisible();
    const committed = await observationServiceSnapshot(page, fixture);
    await editor
      .getByRole("button", { name: "Replay original observation request" })
      .click();
    await expect(
      editor.getByText(/The observation outcome is unknown/),
    ).toHaveCount(0);
    await expect
      .poll(
        async () =>
          (await observationServiceSnapshot(page, fixture)).observations[0]
            ?.row_version,
      )
      .toBe(previous.row_version + 1);
    expect(bodies).toHaveLength(2);
    expect(bodies[1]).toBe(bodies[0]);
    expect(results[1]).toEqual({ ...results[0], replayed: true });
    expect(JSON.parse(bodies[0] ?? "{}").base_row_version).toBe(
      previous.row_version,
    );
    expect(await observationServiceSnapshot(page, fixture)).toEqual(committed);
    expect(committed.observations).toHaveLength(1);
    expect(committed.oldTarget).toHaveLength(0);
    expect(committed.newTarget).toHaveLength(action === "resolve" ? 1 : 0);
    expect(committed.observations[0]?.resolution_status).toBe(
      action === "resolve"
        ? "resolved"
        : action === "dismiss"
          ? "dismissed"
          : "unresolved",
    );
    expect(committed.source?.cells["timeline.raw_activity_text"]?.value).toBe(
      observationRawText,
    );
    for (const indicator of committed.indicators)
      expect(indicator.cells["indicator.observation_count"]?.value).toBe(
        action === "resolve" && indicator.record_id === newTarget.record_id
          ? 1
          : 0,
      );
    transitionEvidence.push({ action, bodies, results, committed });
    // An intentional same-state request is a rejection and must have no effects.
    if (action === "resolve" || action === "restore") {
      const child = committed.observations[0];
      if (!child) throw new Error("child missing");
      const common = {
        pathParameters: { observation_id: child.observation_id },
        headers: await csrfHeaders(page),
        request: atJsonOrigin(page.request, apiBase),
      };
      const rejected =
        action === "resolve"
          ? await publicHttpOperation({
              ...common,
              operationID: "resolveIndicatorObservation",
              body: {
                client_txn_id: uniqueTxn("illegal"),
                base_row_version: child.row_version,
                resolved_indicator_record_id: newTarget.record_id,
              },
            })
          : await publicHttpOperation({
              ...common,
              operationID: "restoreIndicatorObservation",
              body: {
                client_txn_id: uniqueTxn("illegal"),
                base_row_version: child.row_version,
              },
            });
      expect(rejected.ok).toBe(false);
      expect(rejected.status).toBe(409);
      expect(await observationServiceSnapshot(page, fixture)).toEqual(
        committed,
      );
    }
  }
  // A distinct intentional observation can contain exactly the same text.
  await selectRepeatedObservation(page);
  await editor
    .getByLabel("Existing Indicator", { exact: true })
    .first()
    .selectOption("");
  await editor
    .getByRole("button", { name: "Create observation", exact: true })
    .click();
  await expect
    .poll(
      async () =>
        (await listSourceObservations(page, source.record_id)).data.observations
          .length,
    )
    .toBe(2);
  const second = receipts[2];
  if (!second) throw new Error("second receipt missing");
  expect(second.observation.observed_text).toBe("alpha.example");
  expect(second.observation.observation_id).not.toBe(observationId);
  const history = await fetchFullRecordHistory(page, source.record_id);
  const rollback = history.items.find(
    (item) =>
      item.change_set_id === second.change_set_id &&
      item.available_rollback_actions.includes("change_set"),
  );
  if (!rollback) throw new Error("rollback item missing");
  await page.getByRole("button", { name: "Open history", exact: true }).click();
  const anchor = {
    action: "change_set" as const,
    historyItemRef: rollback.history_item_ref,
  };
  await openHistoryEventDetails(page, rollback.history_item_ref);
  await page.getByTestId(rowHistoryActionTestId(anchor)).click();
  await page.getByTestId(rowHistoryRollbackConfirmButtonTestId(anchor)).click();
  await expect
    .poll(async () =>
      (
        await listSourceObservations(page, source.record_id)
      ).data.observations.map((item) => item.observation_id),
    )
    .toEqual([observationId]);
  const afterRollback = await observationServiceSnapshot(page, fixture);
  expect(afterRollback.source?.cells["timeline.raw_activity_text"]?.value).toBe(
    observationRawText,
  );
  await testInfo.attach("indicator-observation-recovery-evidence", {
    body: Buffer.from(
      JSON.stringify(
        { requests, receipts, before, transitionEvidence, afterRollback },
        null,
        2,
      ),
    ),
    contentType: "application/json",
  });
});

test("Indicator observation browsing retains failed pages and independent targets after source edits", async ({
  page,
}, testInfo) => {
  const fixture = await createObservationFixture(page),
    { incidentId, source, newTarget } = fixture;
  for (let index = 0; index < 101; index++)
    await createViewRow(page, incidentId, indicatorsViewSchemaId, {
      client_txn_id: uniqueTxn("candidate"),
      "indicator.indicator_type": "domain_name",
      "indicator.value_kind": "atomic",
      "indicator.display_value": `a${String(index).padStart(3, "0")}.example`,
    });
  await page.goto(`/?incident_id=${incidentId}`);
  let editor = await openObservationEditor(page, source.record_id);
  await selectRepeatedObservation(page);
  const target = editor.getByLabel("Existing Indicator", { exact: true });
  await expect(target.locator("option")).toHaveCount(101);
  await expect(
    target.locator(`option[value="${newTarget.record_id}"]`),
  ).toHaveCount(0);
  const failedBodies: string[] = [];
  let failCandidate = true;
  await page.route(
    `**/views/${indicatorsViewSchemaId}/query`,
    async (route) => {
      const body = route.request().postDataJSON();
      if (!body.cursor_token) return route.continue();
      failedBodies.push(route.request().postData() ?? "");
      if (!failCandidate) return route.continue();
      failCandidate = false;
      await route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "service_unavailable",
            details: {},
            status: 503,
            retryable: true,
            message: "Page unavailable",
            request_id: "candidate-page",
          },
        }),
      });
    },
  );
  await editor
    .getByRole("button", { name: "Load more Indicators", exact: true })
    .click();
  await expect(
    editor.getByRole("button", { name: "Retry Indicators", exact: true }),
  ).toBeVisible();
  await expect(target.locator("option")).toHaveCount(101);
  await editor
    .getByRole("button", { name: "Retry Indicators", exact: true })
    .click();
  await target.selectOption(newTarget.record_id);
  expect(failedBodies[1]).toBe(failedBodies[0]);
  await page
    .locator('[data-inspector-edit-field="timeline.raw_activity_text"]')
    .click();
  const raw = page.getByRole("textbox", { name: "RAW Activity", exact: true });
  const edited = `${observationRawText}Reviewed`;
  await raw.fill(edited);
  await expect(
    editor.getByRole("button", { name: "Use selected text", exact: true }),
  ).toBeDisabled();
  await raw.press("Control+Enter");
  await expect(
    editor.getByRole("textbox", { name: "Saved source text", exact: true }),
  ).toHaveValue(edited.replaceAll("\r\n", "\n").replaceAll("\r", "\n"));
  await expect(
    editor.getByRole("button", { name: "Create observation", exact: true }),
  ).toBeDisabled();
  await expect(
    page.getByTestId(indicatorObservationTestId("preview")),
  ).toContainText("Select text from the current saved source.");
  await selectRepeatedObservation(page);
  await expect(target).toHaveValue(newTarget.record_id);
  await editor
    .getByRole("button", { name: "Create observation", exact: true })
    .click();
  await expect
    .poll(
      async () =>
        (await listSourceObservations(page, source.record_id)).data.observations
          .length,
    )
    .toBe(1);
  const current = (
    await queryViewRows(page, incidentId, timelineViewSchemaId)
  ).find((row) => row.record_id === source.record_id);
  if (!current) throw new Error("saved source missing");
  // The ordinary text editor owns its representation; capture never rewrites it.
  const committedText = String(
    current.cells["timeline.raw_activity_text"]?.value,
  );
  const spanStart = Buffer.byteLength(
    committedText.slice(0, committedText.lastIndexOf("alpha.example")),
  );
  let version = current.row_version;
  for (let index = 0; index < 100; index++) {
    const result = await publicHttpOperation({
      operationID: "createManualIndicatorObservation",
      pathParameters: { source_record_id: source.record_id },
      headers: await csrfHeaders(page),
      request: atJsonOrigin(page.request, apiBase),
      body: {
        client_txn_id: uniqueTxn("observation"),
        base_row_version: version,
        source_field_key: "timeline.raw_activity_text",
        span_start_byte: spanStart,
        span_end_byte: spanStart + 13,
      },
    });
    if (!result.ok) throw new Error(`observation seed: ${result.status}`);
    version =
      result.payload.data.affected_records.find(
        (row) => row.record_id === source.record_id,
      )?.row_version ?? 0;
  }
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  editor = await openObservationEditor(page, source.record_id);
  await expect(editor.getByRole("article")).toHaveCount(100);
  const cursors: string[] = [];
  let failPage = true;
  await page.route(
    `**/records/${source.record_id}/indicator-observations?**`,
    async (route) => {
      const cursor = new URL(route.request().url()).searchParams.get(
        "cursor_token",
      );
      if (!cursor) return route.continue();
      cursors.push(cursor);
      if (!failPage) return route.continue();
      failPage = false;
      await route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "service_unavailable",
            details: {},
            status: 503,
            retryable: true,
            message: "Page unavailable",
            request_id: "observation-page",
          },
        }),
      });
    },
  );
  await editor
    .getByRole("button", { name: "Load more observations", exact: true })
    .click();
  await expect(
    editor.getByRole("button", { name: "Retry observations", exact: true }),
  ).toBeVisible();
  await expect(editor.getByRole("article")).toHaveCount(100);
  await editor
    .getByRole("button", { name: "Retry observations", exact: true })
    .click();
  await expect(editor.getByRole("article")).toHaveCount(101);
  expect(cursors[1]).toBe(cursors[0]);
  expect(
    (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
      (row) => row.record_id === source.record_id,
    )?.cells["timeline.raw_activity_text"]?.value,
  ).toBe(committedText);
  await testInfo.attach("observation-paging-evidence", {
    body: JSON.stringify({
      candidateRequests: failedBodies,
      observationCursors: cursors,
      version,
      committedText,
    }),
    contentType: "application/json",
  });
});
