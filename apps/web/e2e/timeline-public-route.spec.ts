import {
  pasteGridMatrix,
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
  sortByHeader,
} from "@cartulary/test-utils/grid";
import {
  conflictMarkerTestId,
  draftCellTestId,
  gridShellTestId,
  pendingQueueNoticeTestId,
  rowCellTestId,
  saveStateActionButtonTestId,
  saveStateTestId,
  timelineMutationSubstrateReadyTestId,
  timelineRowVersionTestId,
  timelineScalarEditorTestId,
  workbookConflictLocalValueTestId,
  workbookConflictSavedValueTestId,
  workbookEditRecoveryDiscardButtonTestId,
  workbookEditRecoveryRetryButtonTestId,
  workbookEditRecoveryTestId,
} from "@cartulary/ui-contracts";
import type { APIResponse, Page, Response, Route } from "@playwright/test";
import { expect, test } from "./fixtures";
import { openIncidentFromLanding } from "./pages/incidentDirectory";
import { gridSavedRows } from "./pages/workbookInspector";
import { csrfHeaders } from "./support/auth/browserSession";
import { revokeAllSessions } from "./support/auth/sessions";
import {
  exerciseRevokedPendingReplay,
  installPatchController,
} from "./support/collaboration/replay";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { safelyRemoveRoute as safeUnroute } from "./support/transport/requestInterception";
import { createEnvironmentTestControlClient } from "./support/transport/testControlEnvironment";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
  type ViewApiRow,
} from "./support/workbook/query";
import { readTimelineMutation } from "./support/workbook/rowMutations";

const timelineViewSchemaId = "cartulary.view.timeline.v2";
const exactScenarioTitle =
  "Verify rough Timeline row creation, inline edit, paste, pending save, refresh, and replay through /api/v1/ route contracts.";
const recoveryScenarioTitle =
  "Recover Timeline client transaction blockers through remount-safe IDs, retry, discard, and same-field resolver handoff.";

const allowedCreateFieldKeys = new Set([
  "client_txn_id",
  "timeline.date_entered_text",
  "timeline.analyst_text",
  "timeline.mitre_stage_text",
  "timeline.device_object_text",
  "timeline.ip_address_text",
  "timeline.activity_utc_text",
  "timeline.activity_local_text",
  "timeline.raw_activity_text",
  "timeline.activity_synopsis_text",
  "timeline.data_source_text",
  "timeline.host_refs",
  "timeline.identity_refs",
  "timeline.tags",
  "timeline.attached_evidence_ids",
]);

type PublicErrorEnvelope = {
  error: {
    code: string;
    conflict?: Record<string, unknown>;
    details?: Record<string, unknown>;
    message?: string;
    request_id?: string;
    retryable?: boolean;
    status: number;
  };
};

function visibleRecordIds(page: Page) {
  return gridSavedRows(page, timelineViewSchemaId).evaluateAll((rows) =>
    rows.map((row) => row.getAttribute("data-grid-record-id") ?? ""),
  );
}

function readPostBody(request: { postData: () => string | null }) {
  return JSON.parse(request.postData() ?? "{}") as Record<string, unknown>;
}

async function readEnvelope(response: APIResponse | Response) {
  return (await response.json()) as Record<string, unknown>;
}

async function expectPublicError(
  response: APIResponse | Response,
  expectedStatus: number,
  expectedCode: string,
  details: Record<string, unknown> = {},
) {
  expect(response.status()).toBe(expectedStatus);
  const body = (await response.json()) as PublicErrorEnvelope;
  expect(body.error.code).toBe(expectedCode);
  expect(body.error.status).toBe(expectedStatus);
  expect(body.error).toHaveProperty("request_id");
  expect(body.error).not.toHaveProperty("stack");
  expect(body.error).not.toHaveProperty("trace");
  expect(body.error).not.toHaveProperty("internal");
  for (const [key, value] of Object.entries(details)) {
    expect(body.error.details?.[key]).toBe(value);
  }
  return body;
}

async function patchTimelinePublic(
  page: Page,
  recordId: string,
  body: Record<string, unknown>,
) {
  return page.request.patch(`${apiBase}/api/v1/records/${recordId}`, {
    headers: await csrfHeaders(page),
    data: body,
  });
}

async function postTimelineCreatePublic(
  page: Page,
  incidentId: string,
  body: Record<string, unknown>,
) {
  return page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
    {
      headers: await csrfHeaders(page),
      data: body,
    },
  );
}

async function postTimelinePastePublic(
  page: Page,
  incidentId: string,
  body: Record<string, unknown>,
) {
  return page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
    {
      headers: await csrfHeaders(page),
      data: body,
    },
  );
}

async function armPublicErrorFault(page: Page, body: Record<string, unknown>) {
  const response = await createEnvironmentTestControlClient(page.request, {
    endpointOrigin: apiBase,
  }).request({
    body: { ...body, consume_once: true },
    method: "POST",
    path: "/api/v1/test/runtime/public-error-faults",
  });
  expect(response.status).toBe(201);
  return response.body as { data: Record<string, unknown> };
}

async function dispatchClipboardText(
  page: Page,
  recordId: string,
  fieldKey: string,
  clipboardText: string,
) {
  const editor = page.getByTestId(
    timelineScalarEditorTestId({ fieldKey, recordId, surface: "grid" }),
  );
  const cell = (await editor.count())
    ? editor
    : page.getByTestId(rowCellTestId(recordId, fieldKey));
  if (!(await editor.count())) {
    await scrollGridCellIntoView({
      cellKey: fieldKey,
      page,
      recordId,
      surface: timelineViewSchemaId,
    });
  }
  await cell.scrollIntoViewIfNeeded();
  await cell.evaluate((element, text) => {
    if (element instanceof HTMLElement) {
      element.focus({ preventScroll: true });
    }
    const data = new DataTransfer();
    data.setData("text/plain", String(text));
    element.dispatchEvent(
      new ClipboardEvent("paste", {
        bubbles: true,
        cancelable: true,
        clipboardData: data,
      }),
    );
  }, clipboardText);
}

async function disableWorkbookSockets(page: Page) {
  await page.addInitScript(() => {
    class ClosedWebSocket {
      static readonly CONNECTING = 0;
      static readonly OPEN = 1;
      static readonly CLOSING = 2;
      static readonly CLOSED = 3;

      readonly url: string;
      readyState = ClosedWebSocket.CONNECTING;
      onclose: ((event: CloseEvent) => void) | null = null;
      onerror: ((event: Event) => void) | null = null;
      onmessage: ((event: MessageEvent) => void) | null = null;
      onopen: ((event: Event) => void) | null = null;

      constructor(url: string | URL) {
        this.url = String(url);
      }

      close() {
        this.readyState = ClosedWebSocket.CLOSED;
        this.onclose?.(new CloseEvent("close"));
      }

      send() {}
    }

    Object.defineProperty(window, "WebSocket", {
      configurable: true,
      value: ClosedWebSocket,
    });
  });
}

function summaryValue(row: ViewApiRow | undefined) {
  return row?.cells["timeline.activity_synopsis_text"]?.value;
}

async function openTimelineIncident(page: Page, incidentId: string) {
  await openIncidentFromLanding(page, incidentId);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
}

async function submitTimelineSummary(
  page: Page,
  recordId: string,
  value: string,
) {
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId,
    surface: timelineViewSchemaId,
  });
  await page
    .getByTestId(rowCellTestId(recordId, "timeline.activity_synopsis_text"))
    .click();
  const editor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId,
      surface: "grid",
    }),
  );
  await editor.fill(value);
  await editor.press("Enter");
}

test(
  exactScenarioTitle,
  async ({ browser, page, sessionTracker, workerAdminRequest }) => {
    test.setTimeout(300_000);

    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("TIMELINEPUBLIC"),
      "end-to-end.mutation-lifecycle.row-01 public route browser coverage",
    );
    const alpha = await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("timeline-public-alpha"),
      "timeline.activity_utc_text": "2026-04-10T10:00:00.000Z",
      "timeline.activity_synopsis_text":
        "end-to-end.mutation-lifecycle.row-01 Alpha",
    });
    const beta = await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("timeline-public-beta"),
      "timeline.activity_utc_text": "2026-04-10T10:05:00.000Z",
      "timeline.activity_synopsis_text":
        "end-to-end.mutation-lifecycle.row-01 Beta",
    });
    const pasteSeed = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("timeline-public-paste-seed"),
        "timeline.activity_synopsis_text":
          "end-to-end.mutation-lifecycle.row-01 Paste seed",
      },
    );
    const tabularSeed = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("timeline-public-tabular-seed"),
        "timeline.activity_synopsis_text":
          "end-to-end.mutation-lifecycle.row-01 Tabular seed",
      },
    );

    await test.step("authenticated browser query and rough create use public routes", async () => {
      const queryResponse = page.waitForResponse(
        (response) =>
          response.request().method() === "POST" &&
          response
            .url()
            .endsWith(
              `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
            ),
      );
      await page.goto(`/?incident_id=${incidentId}`);
      await expect(
        page.getByTestId(timelineMutationSubstrateReadyTestId()),
      ).toBeVisible();
      const query = await queryResponse;
      expect(query.ok()).toBeTruthy();
      const queryBody = await readEnvelope(query);
      expect(
        (queryBody.data as { view_schema_id?: string }).view_schema_id,
      ).toBe(timelineViewSchemaId);
      await scrollGridCellIntoView({
        cellKey: "timeline.activity_synopsis_text",
        page,
        recordId: alpha.record_id,
        surface: timelineViewSchemaId,
      });
      await expect(
        page.getByTestId(
          rowCellTestId(alpha.record_id, "timeline.activity_synopsis_text"),
        ),
      ).toHaveText("end-to-end.mutation-lifecycle.row-01 Alpha");

      const createBodies: Record<string, unknown>[] = [];
      const createRoute = `**/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`;
      const createRouteHandler = async (route: Route) => {
        if (route.request().method().toUpperCase() === "POST") {
          createBodies.push(readPostBody(route.request()));
        }
        await route.fallback();
      };
      await page.route(createRoute, createRouteHandler);
      try {
        const draftSynopsisTestId = draftCellTestId(
          "timeline.activity_synopsis_text",
        );
        await scrollGridTargetIntoView({
          page,
          surface: timelineViewSchemaId,
          targetTestId: draftSynopsisTestId,
        });
        const createResponse = page.waitForResponse(
          (response) =>
            response.request().method() === "POST" &&
            response
              .url()
              .endsWith(
                `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
              ),
        );
        await page
          .getByTestId(draftSynopsisTestId)
          .fill("end-to-end.mutation-lifecycle.row-01 Rough summary");
        await page.getByTestId(draftSynopsisTestId).press("Enter");
        const createEnvelope = await readTimelineMutation(await createResponse);
        const createdRow = createEnvelope.data.row;
        expect(createBodies).toHaveLength(1);
        const createBody = createBodies[0] ?? {};
        expect(typeof createBody.client_txn_id).toBe("string");
        expect(Object.keys(createBody).sort()).toEqual([
          "client_txn_id",
          "timeline.activity_synopsis_text",
        ]);
        for (const fieldKey of Object.keys(createBody)) {
          expect(allowedCreateFieldKeys.has(fieldKey)).toBeTruthy();
        }
        expect(createBody).not.toHaveProperty("record_id");
        expect(createBody).not.toHaveProperty("row_version");
        expect(createBody).not.toHaveProperty("timeline.capture_state");
        expect(createBody).not.toHaveProperty("timeline.edited_at");
        await scrollGridCellIntoView({
          cellKey: "timeline.activity_synopsis_text",
          page,
          recordId: createdRow.record_id,
          surface: timelineViewSchemaId,
        });
        await expect(
          page.getByTestId(
            rowCellTestId(
              createdRow.record_id,
              "timeline.activity_synopsis_text",
            ),
          ),
        ).toHaveText("end-to-end.mutation-lifecycle.row-01 Rough summary");
        await scrollGridCellIntoView({
          cellKey: "timeline.capture_state",
          page,
          recordId: createdRow.record_id,
          surface: timelineViewSchemaId,
        });
        await expect(
          page.getByTestId(
            rowCellTestId(createdRow.record_id, "timeline.capture_state"),
          ),
        ).toHaveText("rough");
      } finally {
        await safeUnroute(page, createRoute, createRouteHandler);
      }

      for (const [fieldKey, value] of [
        ["record_id", beta.record_id],
        ["row_version", 1],
        ["timeline.evidence_count", 4],
        ["timeline.capture_state", "reviewed"],
        ["timeline.edited_at", "2026-04-10T10:15:00Z"],
      ] as const) {
        const response = await postTimelineCreatePublic(page, incidentId, {
          client_txn_id: uniqueTxn(`timeline-public-create-${fieldKey}`),
          "timeline.activity_synopsis_text": `end-to-end.mutation-lifecycle.row-01 forbidden ${fieldKey}`,
          [fieldKey]: value,
        });
        await expectPublicError(response, 400, "invalid_mutation_payload", {
          field: fieldKey,
          reason_code: "unknown_field",
        });
      }
      const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
      expect(
        rows.some((row) =>
          String(summaryValue(row) ?? "").startsWith(
            "end-to-end.mutation-lifecycle.row-01 forbidden ",
          ),
        ),
      ).toBe(false);
    });

    await test.step("inline edit route covers payload boundaries, save-state, refresh, and conflicts", async () => {
      const patchController = await installPatchController(page);
      try {
        const heldPatch = patchController.holdNextPatch({
          recordId: beta.record_id,
        });
        await scrollGridCellIntoView({
          cellKey: "timeline.activity_synopsis_text",
          page,
          recordId: beta.record_id,
          surface: timelineViewSchemaId,
        });
        const betaInlineCell = page.getByTestId(
          rowCellTestId(beta.record_id, "timeline.activity_synopsis_text"),
        );
        await betaInlineCell.click();
        const betaInlineEditor = page.getByTestId(
          timelineScalarEditorTestId({
            fieldKey: "timeline.activity_synopsis_text",
            recordId: beta.record_id,
            surface: "grid",
          }),
        );
        await betaInlineEditor.fill(
          "end-to-end.mutation-lifecycle.row-01 Beta inline edit",
        );
        await betaInlineEditor.press("Enter");
        const heldCall = await heldPatch.waitForHit;
        expect(heldCall.recordId).toBe(beta.record_id);
        expect(heldCall.body).toMatchObject({
          view_schema_id: timelineViewSchemaId,
          base_row_version: beta.row_version,
        });
        expect(typeof heldCall.body.client_txn_id).toBe("string");
        const changes = heldCall.body.changes;
        expect(Array.isArray(changes)).toBe(true);
        expect(changes).toHaveLength(1);
        expect(changes).toEqual([
          {
            field_key: "timeline.activity_synopsis_text",
            value: "end-to-end.mutation-lifecycle.row-01 Beta inline edit",
          },
        ]);
        await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");

        await sortByHeader(
          page,
          timelineViewSchemaId,
          "timeline.activity_synopsis_text",
        );
        await expect
          .poll(async () => visibleRecordIds(page))
          .toContain(beta.record_id);

        const patchResponse = page.waitForResponse(
          (response) =>
            response.request().method() === "PATCH" &&
            response.url().endsWith(`/api/v1/records/${beta.record_id}`),
        );
        heldPatch.release();
        const patchEnvelope = (await (await patchResponse).json()) as {
          data: { change_set_id: string; row: ViewApiRow };
        };
        const completedPatch = await heldPatch.waitForCompletion;
        expect(completedPatch.status).toBe(200);
        await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
        await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(
          0,
        );
        await expect(
          page.getByTestId(
            rowCellTestId(beta.record_id, "timeline.activity_synopsis_text"),
          ),
        ).toHaveText("end-to-end.mutation-lifecycle.row-01 Beta inline edit");
        await expect(
          page.getByTestId(timelineRowVersionTestId(beta.record_id)),
        ).toHaveText(String(patchEnvelope.data.row.row_version));

        const replayResponse = await patchTimelinePublic(
          page,
          beta.record_id,
          heldCall.body,
        );
        expect(replayResponse.ok()).toBeTruthy();
        const replayEnvelope = (await replayResponse.json()) as {
          data: { change_set_id: string; row: ViewApiRow };
        };
        expect(replayEnvelope.data.change_set_id).toBe(
          patchEnvelope.data.change_set_id,
        );
        expect(replayEnvelope.data.row.row_version).toBe(
          patchEnvelope.data.row.row_version,
        );
        expect(summaryValue(replayEnvelope.data.row)).toBe(
          "end-to-end.mutation-lifecycle.row-01 Beta inline edit",
        );

        await expectPublicError(
          await patchTimelinePublic(page, beta.record_id, {
            ...heldCall.body,
            changes: [
              {
                field_key: "timeline.activity_synopsis_text",
                value: "end-to-end.mutation-lifecycle.row-01 divergent replay",
              },
            ],
          }),
          409,
          "client_txn_conflict",
        );
        await expectPublicError(
          await patchTimelinePublic(page, beta.record_id, {
            view_schema_id: timelineViewSchemaId,
            base_row_version: patchEnvelope.data.row.row_version,
            client_txn_id: uniqueTxn("timeline-public-empty-changes"),
            changes: [],
          }),
          400,
          "invalid_mutation_payload",
          { field: "changes", reason_code: "empty_changes" },
        );
        await expectPublicError(
          await patchTimelinePublic(page, beta.record_id, {
            view_schema_id: timelineViewSchemaId,
            base_row_version: patchEnvelope.data.row.row_version,
            client_txn_id: uniqueTxn("timeline-public-duplicate-field"),
            changes: [
              { field_key: "timeline.activity_synopsis_text", value: "first" },
              { field_key: "timeline.activity_synopsis_text", value: "second" },
            ],
          }),
          400,
          "invalid_mutation_payload",
          { field: "changes", reason_code: "duplicate_field_key" },
        );
        await expectPublicError(
          await patchTimelinePublic(page, beta.record_id, {
            view_schema_id: timelineViewSchemaId,
            base_row_version: patchEnvelope.data.row.row_version,
            client_txn_id: uniqueTxn("timeline-public-over-max"),
            changes: Array.from({ length: 33 }, (_, index) => ({
              field_key: "timeline.activity_synopsis_text",
              value: `over max ${index}`,
            })),
          }),
          400,
          "invalid_mutation_payload",
          { field: "changes", reason_code: "change_count_exceeded" },
        );
        await expectPublicError(
          await patchTimelinePublic(page, beta.record_id, {
            view_schema_id: timelineViewSchemaId,
            base_row_version: patchEnvelope.data.row.row_version + 100,
            client_txn_id: uniqueTxn("timeline-public-row-version-conflict"),
            changes: [
              {
                field_key: "timeline.activity_synopsis_text",
                value:
                  "end-to-end.mutation-lifecycle.row-01 future row version",
              },
            ],
          }),
          409,
          "row_version_conflict",
        );

        const conflictRow = await createViewRow(
          page,
          incidentId,
          timelineViewSchemaId,
          {
            client_txn_id: uniqueTxn("timeline-public-conflict-row"),
            "timeline.activity_synopsis_text":
              "end-to-end.mutation-lifecycle.row-01 conflict base",
          },
        );
        const serverConflictPatch = await patchTimelinePublic(
          page,
          conflictRow.record_id,
          {
            view_schema_id: timelineViewSchemaId,
            base_row_version: conflictRow.row_version,
            client_txn_id: uniqueTxn("timeline-public-same-field-server"),
            changes: [
              {
                field_key: "timeline.activity_synopsis_text",
                value: "end-to-end.mutation-lifecycle.row-01 conflict server",
              },
            ],
          },
        );
        expect(serverConflictPatch.ok()).toBeTruthy();
        const sameFieldBody = await expectPublicError(
          await patchTimelinePublic(page, conflictRow.record_id, {
            view_schema_id: timelineViewSchemaId,
            base_row_version: conflictRow.row_version,
            client_txn_id: uniqueTxn("timeline-public-same-field-client"),
            changes: [
              {
                field_key: "timeline.activity_synopsis_text",
                value: "end-to-end.mutation-lifecycle.row-01 conflict client",
              },
            ],
          }),
          409,
          "same_field_conflict",
        );
        expect(sameFieldBody.error.conflict?.field_key).toBe(
          "timeline.activity_synopsis_text",
        );

        await page.reload();
        await scrollGridCellIntoView({
          cellKey: "timeline.activity_synopsis_text",
          page,
          recordId: beta.record_id,
          surface: timelineViewSchemaId,
        });
        await expect(
          page.getByTestId(
            rowCellTestId(beta.record_id, "timeline.activity_synopsis_text"),
          ),
        ).toHaveText("end-to-end.mutation-lifecycle.row-01 Beta inline edit");
        await expect(
          page.getByTestId(timelineRowVersionTestId(beta.record_id)),
        ).toHaveText(String(patchEnvelope.data.row.row_version));
      } finally {
        await patchController.dispose();
      }
    });

    await test.step("paste dispatch, target validation, non-conflicting commit, and grouped conflicts use public routes", async () => {
      const scalarPasteRequests: string[] = [];
      const captureScalarPaste = (request: {
        method: () => string;
        url: () => string;
      }) => {
        if (
          request.method() === "POST" &&
          request.url().includes("/clipboard-paste")
        ) {
          scalarPasteRequests.push(request.url());
        }
      };
      page.on("request", captureScalarPaste);
      try {
        const scalarPatchResponse = page.waitForResponse(
          (response) =>
            response.request().method() === "PATCH" &&
            response.url().endsWith(`/api/v1/records/${pasteSeed.record_id}`),
        );
        await scrollGridCellIntoView({
          cellKey: "timeline.activity_synopsis_text",
          page,
          recordId: pasteSeed.record_id,
          surface: timelineViewSchemaId,
        });
        const scalarPasteCell = page.getByTestId(
          rowCellTestId(pasteSeed.record_id, "timeline.activity_synopsis_text"),
        );
        await scalarPasteCell.click();
        await page
          .getByTestId(
            timelineScalarEditorTestId({
              fieldKey: "timeline.activity_synopsis_text",
              recordId: pasteSeed.record_id,
              surface: "grid",
            }),
          )
          .fill("end-to-end.mutation-lifecycle.row-01 scalar, comma only");
        await dispatchClipboardText(
          page,
          pasteSeed.record_id,
          "timeline.activity_synopsis_text",
          "end-to-end.mutation-lifecycle.row-01 scalar, comma only",
        );
        const scalarPatch = await scalarPatchResponse;
        expect(scalarPatch.ok()).toBeTruthy();
        expect(readPostBody(scalarPatch.request())).toMatchObject({
          view_schema_id: timelineViewSchemaId,
          changes: [
            {
              field_key: "timeline.activity_synopsis_text",
              value: "end-to-end.mutation-lifecycle.row-01 scalar, comma only",
            },
          ],
        });
        expect(scalarPasteRequests).toHaveLength(0);
        await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      } finally {
        page.off("request", captureScalarPaste);
      }

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
        matrix: [
          [
            "end-to-end.mutation-lifecycle.row-01 tabular summary",
            "timeline-public-host.example.test",
          ],
        ],
        page,
        recordId: tabularSeed.record_id,
        surface: timelineViewSchemaId,
      });
      const pasteBody = readPostBody(await pasteRequest);
      expect(pasteBody).toMatchObject({
        view_schema_id: timelineViewSchemaId,
        clipboard_text:
          "end-to-end.mutation-lifecycle.row-01 tabular summary\ttimeline-public-host.example.test",
        format: "tsv",
        start_field_key: "timeline.activity_synopsis_text",
        columns: [
          "timeline.activity_synopsis_text",
          "timeline.data_source_text",
        ],
        targets: [
          {
            kind: "record",
            record_id: tabularSeed.record_id,
          },
        ],
      });
      expect(typeof pasteBody.client_txn_id).toBe("string");
      expect((await pasteResponse).ok()).toBeTruthy();
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(
        page.getByTestId(
          rowCellTestId(
            tabularSeed.record_id,
            "timeline.activity_synopsis_text",
          ),
        ),
      ).toHaveText("end-to-end.mutation-lifecycle.row-01 tabular summary");

      const foreignIncidentId = await createIncident(
        page,
        uniqueIncidentKey("TIMELINEPUBLICFOREIGN"),
        "end-to-end.mutation-lifecycle.row-01 foreign paste target",
      );
      const foreign = await createViewRow(
        page,
        foreignIncidentId,
        timelineViewSchemaId,
        {
          client_txn_id: uniqueTxn("timeline-public-foreign-row"),
          "timeline.activity_synopsis_text":
            "end-to-end.mutation-lifecycle.row-01 foreign untouched",
        },
      );
      await expectPublicError(
        await postTimelinePastePublic(page, incidentId, {
          view_schema_id: timelineViewSchemaId,
          client_txn_id: uniqueTxn("timeline-public-invalid-paste-target"),
          clipboard_text:
            "end-to-end.mutation-lifecycle.row-01 should not create\nend-to-end.mutation-lifecycle.row-01 should not disclose",
          format: "tsv",
          start_field_key: "timeline.activity_synopsis_text",
          columns: ["timeline.activity_synopsis_text"],
          targets: [
            { kind: "create" },
            {
              kind: "record",
              record_id: foreign.record_id,
              base_row_version: foreign.row_version,
            },
          ],
        }),
        404,
        "incident_not_found",
      );
      const localRowsAfterRejectedPaste = await queryViewRows(
        page,
        incidentId,
        timelineViewSchemaId,
      );
      expect(
        localRowsAfterRejectedPaste.some(
          (row) =>
            summaryValue(row) ===
            "end-to-end.mutation-lifecycle.row-01 should not create",
        ),
      ).toBe(false);
      const foreignRowsAfterRejectedPaste = await queryViewRows(
        page,
        foreignIncidentId,
        timelineViewSchemaId,
      );
      expect(summaryValue(foreignRowsAfterRejectedPaste[0])).toBe(
        "end-to-end.mutation-lifecycle.row-01 foreign untouched",
      );

      const staleIncidentId = await createIncident(
        page,
        uniqueIncidentKey("TIMELINEPUBLICSTALE"),
        "end-to-end.mutation-lifecycle.row-01 grouped paste conflict",
      );
      const staleFirst = await createViewRow(
        page,
        staleIncidentId,
        timelineViewSchemaId,
        {
          client_txn_id: uniqueTxn("timeline-public-stale-first"),
          "timeline.activity_synopsis_text":
            "end-to-end.mutation-lifecycle.row-01 stale first base",
        },
      );
      const staleSecond = await createViewRow(
        page,
        staleIncidentId,
        timelineViewSchemaId,
        {
          client_txn_id: uniqueTxn("timeline-public-stale-second"),
          "timeline.activity_synopsis_text":
            "end-to-end.mutation-lifecycle.row-01 stale second base",
        },
      );
      const staleContext = await browser.newContext({
        storageState: await page.context().storageState(),
      });
      const stalePage = await staleContext.newPage();
      try {
        await disableWorkbookSockets(stalePage);
        await stalePage.goto(`/?incident_id=${staleIncidentId}`);
        await expect(
          stalePage.getByTestId(timelineMutationSubstrateReadyTestId()),
        ).toBeVisible();
        const visibleStaleRecordIds = (
          await gridSavedRows(stalePage, timelineViewSchemaId).evaluateAll(
            (rows) =>
              rows.map((row) => row.getAttribute("data-grid-record-id") ?? ""),
          )
        ).filter(
          (recordId) =>
            recordId === staleFirst.record_id ||
            recordId === staleSecond.record_id,
        );
        expect(visibleStaleRecordIds).toHaveLength(2);
        const staleStartRecordId = visibleStaleRecordIds[0];
        const staleNextRecordId = visibleStaleRecordIds[1];
        if (
          staleStartRecordId === undefined ||
          staleNextRecordId === undefined
        ) {
          throw new Error("expected two visible stale Timeline records");
        }
        const stalePasteTextByRecordId = new Map([
          [
            staleFirst.record_id,
            "end-to-end.mutation-lifecycle.row-01 stale first client",
          ],
          [
            staleSecond.record_id,
            "end-to-end.mutation-lifecycle.row-01 stale second client",
          ],
        ]);
        const staleStartText = stalePasteTextByRecordId.get(staleStartRecordId);
        const staleNextText = stalePasteTextByRecordId.get(staleNextRecordId);
        if (staleStartText === undefined || staleNextText === undefined) {
          throw new Error(
            "visible stale Timeline record did not map to paste text",
          );
        }
        await scrollGridCellIntoView({
          cellKey: "timeline.activity_synopsis_text",
          page: stalePage,
          recordId: staleStartRecordId,
          surface: timelineViewSchemaId,
        });
        const staleStartCell = stalePage
          .getByTestId(
            rowCellTestId(
              staleStartRecordId,
              "timeline.activity_synopsis_text",
            ),
          )
          .locator("xpath=ancestor::*[@role='gridcell'][1]");
        await staleStartCell.dispatchEvent("mousedown", { button: 0 });
        await staleStartCell.focus();
        await expect(stalePage.getByTestId("workbook-focus-anchor")).toHaveText(
          `${timelineViewSchemaId}:${staleStartRecordId}:timeline.activity_synopsis_text`,
        );
        await patchRecord(page, staleFirst.record_id, {
          view_schema_id: timelineViewSchemaId,
          base_row_version: staleFirst.row_version,
          client_txn_id: uniqueTxn("timeline-public-stale-first-server"),
          changes: [
            {
              field_key: "timeline.activity_synopsis_text",
              value: "end-to-end.mutation-lifecycle.row-01 stale first server",
            },
          ],
        });
        await patchRecord(page, staleSecond.record_id, {
          view_schema_id: timelineViewSchemaId,
          base_row_version: staleSecond.row_version,
          client_txn_id: uniqueTxn("timeline-public-stale-second-server"),
          changes: [
            {
              field_key: "timeline.activity_synopsis_text",
              value: "end-to-end.mutation-lifecycle.row-01 stale second server",
            },
          ],
        });
        const groupedPasteResponse = stalePage.waitForResponse(
          (response) =>
            response.request().method() === "POST" &&
            response
              .url()
              .includes(
                `/api/v1/incidents/${staleIncidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
              ),
        );
        await dispatchClipboardText(
          stalePage,
          staleStartRecordId,
          "timeline.activity_synopsis_text",
          [
            staleStartText,
            staleNextText,
            "end-to-end.mutation-lifecycle.row-01 stale created",
          ].join("\n"),
        );
        const groupedEnvelope = (await (await groupedPasteResponse).json()) as {
          data: {
            conflicts: Array<Record<string, unknown>>;
            rows: ViewApiRow[];
          };
        };
        expect(groupedEnvelope.data.conflicts).toHaveLength(2);
        expect(groupedEnvelope.data.rows).toHaveLength(1);
        await expect(
          stalePage.getByTestId(gridShellTestId(timelineViewSchemaId)),
        ).toBeVisible();
        await expect(stalePage.getByTestId(saveStateTestId())).toHaveText(
          "Conflict",
        );
        await expect(
          stalePage.getByTestId(
            conflictMarkerTestId(
              staleStartRecordId,
              "timeline.activity_synopsis_text",
            ),
          ),
        ).toBeVisible();
        await expect(
          stalePage.getByTestId(
            conflictMarkerTestId(
              staleNextRecordId,
              "timeline.activity_synopsis_text",
            ),
          ),
        ).toBeVisible();
        await expect(
          stalePage.getByTestId("paste-conflict-navigator"),
        ).toBeVisible();
        await expect(
          stalePage.getByTestId("paste-conflict-position"),
        ).toHaveText("1 of 2");
        await expect(
          stalePage.getByTestId(workbookConflictLocalValueTestId()),
        ).toHaveValue(staleStartText);
        await stalePage.getByTestId("paste-conflict-next").click();
        await expect(
          stalePage.getByTestId("paste-conflict-position"),
        ).toHaveText("2 of 2");
        await expect(
          stalePage.getByTestId(workbookConflictLocalValueTestId()),
        ).toHaveValue(staleNextText);
      } finally {
        await staleContext.close();
      }
    });

    await test.step("public string date edits preserve unparseable authored text", async () => {
      await openTimelineIncident(page, incidentId);
      await scrollGridCellIntoView({
        cellKey: "timeline.activity_utc_text",
        page,
        recordId: beta.record_id,
        surface: timelineViewSchemaId,
      });
      const validationResponse = page.waitForResponse(
        (response) =>
          response.request().method() === "PATCH" &&
          response.url().endsWith(`/api/v1/records/${beta.record_id}`),
      );
      const betaUtcCell = page.getByTestId(
        rowCellTestId(beta.record_id, "timeline.activity_utc_text"),
      );
      await betaUtcCell.click();
      const betaUtcEditor = page.getByTestId(
        timelineScalarEditorTestId({
          fieldKey: "timeline.activity_utc_text",
          recordId: beta.record_id,
          surface: "grid",
        }),
      );
      await betaUtcEditor.fill("not-a-timestamp");
      await betaUtcEditor.press("Enter");
      const validationEnvelope = await readTimelineMutation(
        await validationResponse,
      );
      expect(
        validationEnvelope.data.row.cells["timeline.activity_utc_text"]?.value,
      ).toBe("not-a-timestamp");
      expect(
        validationEnvelope.data.row.cells["timeline.activity_time_pair_state"]
          ?.value,
      ).toBe("disabled");
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
      await expect(
        page.getByTestId(
          rowCellTestId(beta.record_id, "timeline.activity_utc_text"),
        ),
      ).toHaveText("not-a-timestamp");
      await page.reload();
      await openTimelineIncident(page, incidentId);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    });

    await test.step("unknown public error fallback is induced through a harness-owned public boundary", async () => {
      const faultPath = `/api/v1/records/${alpha.record_id}`;
      const armedFault = await armPublicErrorFault(page, {
        method: "PATCH",
        path: faultPath,
        status: 418,
        code: "future_private_public_error",
        message:
          "stack trace at handler (/home/cartulary/internal/private.go:42)",
        retryable: false,
        details: {
          reason_code: "future_private_public_error",
          private_path: "/home/cartulary/internal/private.go",
        },
      });
      expect(armedFault.data).toMatchObject({
        schema_id: "cartulary.test.public_error_fault.v1",
        method: "PATCH",
        path: faultPath,
        status: 418,
        code: "future_private_public_error",
        retryable: false,
        consume_once: true,
      });

      const unknownPatchResponse = page.waitForResponse(
        (response) =>
          response.request().method() === "PATCH" &&
          response.url().endsWith(faultPath),
      );
      const alphaSummaryCell = page.getByTestId(
        rowCellTestId(alpha.record_id, "timeline.activity_synopsis_text"),
      );
      await scrollGridCellIntoView({
        cellKey: "timeline.activity_synopsis_text",
        page,
        recordId: alpha.record_id,
        surface: timelineViewSchemaId,
      });
      await alphaSummaryCell.scrollIntoViewIfNeeded();
      await alphaSummaryCell.click();
      const alphaSummaryEditor = page.getByTestId(
        timelineScalarEditorTestId({
          fieldKey: "timeline.activity_synopsis_text",
          recordId: alpha.record_id,
          surface: "grid",
        }),
      );
      await alphaSummaryEditor.fill(
        "end-to-end.mutation-lifecycle.row-01 unknown fallback local",
      );
      await alphaSummaryEditor.press("Enter");
      const unknownEnvelope = await expectPublicError(
        await unknownPatchResponse,
        418,
        "future_private_public_error",
        { reason_code: "future_private_public_error" },
      );
      expect(unknownEnvelope.error.retryable).toBe(false);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
      const notice = page.getByTestId(pendingQueueNoticeTestId());
      await expect(notice).toBeVisible();
      await expect(notice).toContainText("Request failed.");
      await expect(notice).not.toContainText("/home/cartulary");
      await expect(notice).not.toContainText("stack trace");
      await expect(notice).not.toContainText("private.go");
      await expect(alphaSummaryEditor).toHaveValue(
        "end-to-end.mutation-lifecycle.row-01 unknown fallback local",
      );

      page.once("dialog", async (dialog) => {
        await dialog.accept();
      });
      await page.reload();
      await openTimelineIncident(page, incidentId);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    });

    await test.step("real session revocation replays queued public PATCH work in FIFO order", async () => {
      await exerciseRevokedPendingReplay({
        createdBy: "end-to-end.mutation-lifecycle.row-01",
        incidentKeyPrefix: "TIMELINEPUBLICREVOKE",
        localValues: [
          "end-to-end.mutation-lifecycle.row-01 revoked A local",
          "end-to-end.mutation-lifecycle.row-01 revoked B local",
          "end-to-end.mutation-lifecycle.row-01 revoked C local",
        ],
        page,
        scenario: "revoked",
        sessionTracker,
        triggerRevocation: async ({ member }) => {
          await revokeAllSessions(
            workerAdminRequest,
            member.user_id,
            "end-to-end.mutation-lifecycle.row-01 browser revoke-all",
          );
        },
      });
    });
  },
);

test(recoveryScenarioTitle, async ({ browser, page }) => {
  test.setTimeout(180_000);

  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TIMELINEPUBLICRECOVERY"),
    "end-to-end.mutation-lifecycle.row-01 transaction conflict recovery",
  );
  const remountRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("timeline-public-remount-row"),
      "timeline.activity_synopsis_text":
        "end-to-end.mutation-lifecycle.row-01 remount base",
    },
  );
  const retryRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("timeline-public-retry-row"),
    "timeline.activity_synopsis_text":
      "end-to-end.mutation-lifecycle.row-01 retry base",
  });
  const discardRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("timeline-public-discard-row"),
      "timeline.activity_synopsis_text":
        "end-to-end.mutation-lifecycle.row-01 discard base",
    },
  );
  const sameFieldRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("timeline-public-same-field-row"),
      "timeline.activity_synopsis_text":
        "end-to-end.mutation-lifecycle.row-01 resolver base",
    },
  );

  await openTimelineIncident(page, incidentId);
  const patchController = await installPatchController(page);
  try {
    await test.step("new logical edits remain unique across a workbook remount", async () => {
      await submitTimelineSummary(
        page,
        remountRow.record_id,
        "end-to-end.mutation-lifecycle.row-01 remount first",
      );
      await expect.poll(() => patchController.calls.length).toBe(1);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      const firstClientTxnId = String(
        patchController.calls[0]?.body.client_txn_id,
      );

      await page.reload();
      await openTimelineIncident(page, incidentId);
      await submitTimelineSummary(
        page,
        remountRow.record_id,
        "end-to-end.mutation-lifecycle.row-01 remount second",
      );
      await expect.poll(() => patchController.calls.length).toBe(2);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      const secondClientTxnId = String(
        patchController.calls[1]?.body.client_txn_id,
      );

      expect(firstClientTxnId).toMatch(
        /^timeline-client-[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
      );
      expect(secondClientTxnId).toMatch(
        /^timeline-client-[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
      );
      expect(secondClientTxnId).not.toBe(firstClientTxnId);
    });

    await test.step("retry rekeys only the blocker and rematerializes the current row version", async () => {
      patchController.failNextPatch(409, "client_txn_conflict", {
        recordId: retryRow.record_id,
      });
      await submitTimelineSummary(
        page,
        retryRow.record_id,
        "end-to-end.mutation-lifecycle.row-01 retry local",
      );
      await expect.poll(() => patchController.calls.length).toBe(3);
      const blockedCall = patchController.calls[2];
      expect(blockedCall?.status).toBe(409);
      const blockedClientTxnId = String(blockedCall?.body.client_txn_id);

      const recoveryPanel = page.getByTestId(workbookEditRecoveryTestId());
      await expect(recoveryPanel).toBeVisible();
      expect(await recoveryPanel.getByRole("button").allTextContents()).toEqual(
        ["Retry with a new request ID", "Discard blocked edit"],
      );
      await expect(
        page.getByTestId(pendingQueueNoticeTestId()),
      ).not.toContainText(blockedClientTxnId);

      const serverAdvanceResponse = await patchTimelinePublic(
        page,
        retryRow.record_id,
        {
          view_schema_id: timelineViewSchemaId,
          base_row_version: blockedCall?.body.base_row_version,
          client_txn_id: uniqueTxn("timeline-public-retry-server-advance"),
          changes: [
            {
              field_key: "timeline.data_source_text",
              value: "end-to-end.mutation-lifecycle.row-01 server advance",
            },
          ],
        },
      );
      expect(serverAdvanceResponse.ok()).toBeTruthy();
      const serverAdvance = (await serverAdvanceResponse.json()) as {
        data: { row: ViewApiRow };
      };
      await expect(
        page.getByTestId(timelineRowVersionTestId(retryRow.record_id)),
      ).toHaveText(String(serverAdvance.data.row.row_version));

      await page.getByTestId(saveStateActionButtonTestId()).click();
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeFocused();
      const retryButton = page.getByTestId(
        workbookEditRecoveryRetryButtonTestId(),
      );
      await retryButton.focus();
      await retryButton.press("Enter");
      await expect.poll(() => patchController.calls.length).toBe(4);
      const retriedCall = patchController.calls[3];
      expect(retriedCall?.status).toBe(200);
      expect(retriedCall?.body.client_txn_id).not.toBe(blockedClientTxnId);
      expect(retriedCall?.body.base_row_version).toBe(
        serverAdvance.data.row.row_version,
      );
      expect(retriedCall?.body.changes).toEqual(blockedCall?.body.changes);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(recoveryPanel).toHaveCount(0);
    });

    await test.step("discard restores committed display without a server mutation", async () => {
      patchController.failNextPatch(409, "client_txn_conflict", {
        recordId: discardRow.record_id,
      });
      await submitTimelineSummary(
        page,
        discardRow.record_id,
        "end-to-end.mutation-lifecycle.row-01 discard local",
      );
      await expect.poll(() => patchController.calls.length).toBe(5);
      await expect(
        page.getByTestId(workbookEditRecoveryTestId()),
      ).toBeVisible();
      const discardButton = page.getByTestId(
        workbookEditRecoveryDiscardButtonTestId(),
      );
      await discardButton.focus();
      await discardButton.press("Space");
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(page.getByTestId(workbookEditRecoveryTestId())).toHaveCount(
        0,
      );
      await expect(
        page.getByTestId(
          rowCellTestId(
            discardRow.record_id,
            "timeline.activity_synopsis_text",
          ),
        ),
      ).toHaveText("end-to-end.mutation-lifecycle.row-01 discard base");
      expect(patchController.calls).toHaveLength(5);
    });
  } finally {
    await patchController.dispose();
  }

  await test.step("a genuine same-field conflict after rekey opens the normal resolver", async () => {
    const staleContext = await browser.newContext({
      storageState: await page.context().storageState(),
    });
    const stalePage = await staleContext.newPage();
    try {
      await disableWorkbookSockets(stalePage);
      await openTimelineIncident(stalePage, incidentId);
      const stalePatchController = await installPatchController(stalePage);
      try {
        stalePatchController.failNextPatch(409, "client_txn_conflict", {
          recordId: sameFieldRow.record_id,
        });
        await submitTimelineSummary(
          stalePage,
          sameFieldRow.record_id,
          "end-to-end.mutation-lifecycle.row-01 resolver local",
        );
        await expect.poll(() => stalePatchController.calls.length).toBe(1);
        const blockedCall = stalePatchController.calls[0];

        const serverPatch = await patchTimelinePublic(
          page,
          sameFieldRow.record_id,
          {
            view_schema_id: timelineViewSchemaId,
            base_row_version: sameFieldRow.row_version,
            client_txn_id: uniqueTxn("timeline-public-resolver-server"),
            changes: [
              {
                field_key: "timeline.activity_synopsis_text",
                value: "end-to-end.mutation-lifecycle.row-01 resolver server",
              },
            ],
          },
        );
        expect(serverPatch.ok()).toBeTruthy();

        await stalePage
          .getByTestId(workbookEditRecoveryRetryButtonTestId())
          .click();
        await expect.poll(() => stalePatchController.calls.length).toBe(2);
        const retriedCall = stalePatchController.calls[1];
        expect(retriedCall?.status).toBe(409);
        expect(retriedCall?.body.client_txn_id).not.toBe(
          blockedCall?.body.client_txn_id,
        );
        expect(retriedCall?.body.base_row_version).toBe(
          sameFieldRow.row_version,
        );
        await expect(
          stalePage.getByTestId(workbookConflictLocalValueTestId()),
        ).toHaveValue("end-to-end.mutation-lifecycle.row-01 resolver local");
        await expect(
          stalePage.getByTestId(workbookConflictSavedValueTestId()),
        ).toHaveValue("end-to-end.mutation-lifecycle.row-01 resolver server");
        await expect(
          stalePage.getByTestId(workbookEditRecoveryTestId()),
        ).toHaveCount(0);
      } finally {
        await stalePatchController.dispose();
      }
    } finally {
      await staleContext.close();
    }
  });
});
