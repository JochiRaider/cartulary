import { pasteGridMatrix, sortByHeader } from "@cartulary/test-utils";
import {
  conflictMarkerTestId,
  draftCellTestId,
  gridShellTestId,
  pendingQueueNoticeTestId,
  rowCellTestId,
  saveStateTestId,
  timelineMutationSubstrateReadyTestId,
  timelineRowVersionTestId,
} from "@cartulary/ui-contracts";
import type { APIResponse, Page, Response, Route } from "@playwright/test";

import { revokeAllSessions } from "./authRuntime";
import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  createViewRow,
  csrfHeaders,
  gridSavedRows,
  patchTimelineRecord,
  queryViewRows,
  safeUnroute,
  uniqueIncidentKey,
  uniqueTxn,
  type ViewApiRow,
} from "./helpers";
import { readTimelineMutation } from "./phase4Helpers";
import {
  exerciseRevokedPendingReplay,
  installPatchController,
} from "./phase6Harness";

const timelineViewSchemaId = "cartulary.view.timeline.v1";
const exactScenarioTitle =
  "FE-E-P4-01 Verify rough Timeline row creation, inline edit, paste, pending save, refresh, and replay through /api/v1/ route contracts.";

const allowedCreateFieldKeys = new Set([
  "client_txn_id",
  "timeline.occurred_at",
  "timeline.summary",
  "timeline.details",
  "timeline.source_text",
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

async function dispatchClipboardText(
  page: Page,
  recordId: string,
  fieldKey: string,
  clipboardText: string,
) {
  const cell = page.getByTestId(rowCellTestId(recordId, fieldKey));
  await cell.click();
  await cell.evaluate((element, text) => {
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
  return row?.cells["timeline.summary"]?.value;
}

test(
  exactScenarioTitle,
  async ({ browser, page, sessionTracker, workerAdminRequest }) => {
    test.setTimeout(300_000);

    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEEP401"),
      "FE-E-P4-01 public route browser coverage",
    );
    const alpha = await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("feep401-alpha"),
      "timeline.occurred_at": "2026-04-10T10:00:00.000Z",
      "timeline.summary": "FE-E-P4-01 Alpha",
    });
    const beta = await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("feep401-beta"),
      "timeline.occurred_at": "2026-04-10T10:05:00.000Z",
      "timeline.summary": "FE-E-P4-01 Beta",
    });
    const pasteSeed = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("feep401-paste-seed"),
        "timeline.summary": "FE-E-P4-01 Paste seed",
      },
    );
    const tabularSeed = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("feep401-tabular-seed"),
        "timeline.summary": "FE-E-P4-01 Tabular seed",
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
      await expect(
        page.getByTestId(rowCellTestId(alpha.record_id, "timeline.summary")),
      ).toHaveValue("FE-E-P4-01 Alpha");

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
          .getByTestId(draftCellTestId("timeline.summary"))
          .fill("FE-E-P4-01 Rough summary");
        await page
          .getByTestId(draftCellTestId("timeline.summary"))
          .press("Enter");
        const createEnvelope = await readTimelineMutation(await createResponse);
        const createdRow = createEnvelope.data.row;
        expect(createBodies).toHaveLength(1);
        const createBody = createBodies[0] ?? {};
        expect(typeof createBody.client_txn_id).toBe("string");
        expect(Object.keys(createBody).sort()).toEqual([
          "client_txn_id",
          "timeline.summary",
        ]);
        for (const fieldKey of Object.keys(createBody)) {
          expect(allowedCreateFieldKeys.has(fieldKey)).toBeTruthy();
        }
        expect(createBody).not.toHaveProperty("record_id");
        expect(createBody).not.toHaveProperty("row_version");
        expect(createBody).not.toHaveProperty("timeline.capture_state");
        expect(createBody).not.toHaveProperty("timeline.edited_at");
        await expect(
          page.getByTestId(
            rowCellTestId(createdRow.record_id, "timeline.summary"),
          ),
        ).toHaveValue("FE-E-P4-01 Rough summary");
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
          client_txn_id: uniqueTxn(`feep401-create-${fieldKey}`),
          "timeline.summary": `FE-E-P4-01 forbidden ${fieldKey}`,
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
          String(summaryValue(row) ?? "").startsWith("FE-E-P4-01 forbidden "),
        ),
      ).toBe(false);
    });

    await test.step("inline edit route covers payload boundaries, save-state, refresh, and conflicts", async () => {
      const patchController = await installPatchController(page);
      try {
        const heldPatch = patchController.holdNextPatch({
          recordId: beta.record_id,
        });
        await page
          .getByTestId(rowCellTestId(beta.record_id, "timeline.summary"))
          .fill("FE-E-P4-01 Beta inline edit");
        await page
          .getByTestId(rowCellTestId(beta.record_id, "timeline.summary"))
          .press("Enter");
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
            field_key: "timeline.summary",
            value: "FE-E-P4-01 Beta inline edit",
          },
        ]);
        await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");

        await sortByHeader(page, timelineViewSchemaId, "timeline.summary");
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
          page.getByTestId(rowCellTestId(beta.record_id, "timeline.summary")),
        ).toHaveValue("FE-E-P4-01 Beta inline edit");
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
          "FE-E-P4-01 Beta inline edit",
        );

        await expectPublicError(
          await patchTimelinePublic(page, beta.record_id, {
            ...heldCall.body,
            changes: [
              {
                field_key: "timeline.summary",
                value: "FE-E-P4-01 divergent replay",
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
            client_txn_id: uniqueTxn("feep401-empty-changes"),
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
            client_txn_id: uniqueTxn("feep401-duplicate-field"),
            changes: [
              { field_key: "timeline.summary", value: "first" },
              { field_key: "timeline.summary", value: "second" },
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
            client_txn_id: uniqueTxn("feep401-over-max"),
            changes: Array.from({ length: 33 }, (_, index) => ({
              field_key: "timeline.summary",
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
            client_txn_id: uniqueTxn("feep401-row-version-conflict"),
            changes: [
              {
                field_key: "timeline.summary",
                value: "FE-E-P4-01 future row version",
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
            client_txn_id: uniqueTxn("feep401-conflict-row"),
            "timeline.summary": "FE-E-P4-01 conflict base",
          },
        );
        const serverConflictPatch = await patchTimelinePublic(
          page,
          conflictRow.record_id,
          {
            view_schema_id: timelineViewSchemaId,
            base_row_version: conflictRow.row_version,
            client_txn_id: uniqueTxn("feep401-same-field-server"),
            changes: [
              {
                field_key: "timeline.summary",
                value: "FE-E-P4-01 conflict server",
              },
            ],
          },
        );
        expect(serverConflictPatch.ok()).toBeTruthy();
        const sameFieldBody = await expectPublicError(
          await patchTimelinePublic(page, conflictRow.record_id, {
            view_schema_id: timelineViewSchemaId,
            base_row_version: conflictRow.row_version,
            client_txn_id: uniqueTxn("feep401-same-field-client"),
            changes: [
              {
                field_key: "timeline.summary",
                value: "FE-E-P4-01 conflict client",
              },
            ],
          }),
          409,
          "same_field_conflict",
        );
        expect(sameFieldBody.error.conflict?.field_key).toBe(
          "timeline.summary",
        );

        await page.reload();
        await expect(
          page.getByTestId(rowCellTestId(beta.record_id, "timeline.summary")),
        ).toHaveValue("FE-E-P4-01 Beta inline edit");
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
        await page
          .getByTestId(rowCellTestId(pasteSeed.record_id, "timeline.summary"))
          .fill("FE-E-P4-01 scalar, comma only");
        await dispatchClipboardText(
          page,
          pasteSeed.record_id,
          "timeline.summary",
          "FE-E-P4-01 scalar, comma only",
        );
        const scalarPatch = await scalarPatchResponse;
        expect(scalarPatch.ok()).toBeTruthy();
        expect(readPostBody(scalarPatch.request())).toMatchObject({
          view_schema_id: timelineViewSchemaId,
          changes: [
            {
              field_key: "timeline.summary",
              value: "FE-E-P4-01 scalar, comma only",
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
        fieldKey: "timeline.summary",
        matrix: [["FE-E-P4-01 tabular summary", "feep401-host.example.test"]],
        page,
        recordId: tabularSeed.record_id,
        surface: timelineViewSchemaId,
      });
      const pasteBody = readPostBody(await pasteRequest);
      expect(pasteBody).toMatchObject({
        view_schema_id: timelineViewSchemaId,
        clipboard_text: "FE-E-P4-01 tabular summary\tfeep401-host.example.test",
        format: "tsv",
        start_field_key: "timeline.summary",
        columns: ["timeline.summary", "timeline.host_refs"],
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
          rowCellTestId(tabularSeed.record_id, "timeline.summary"),
        ),
      ).toHaveValue("FE-E-P4-01 tabular summary");

      const foreignIncidentId = await createIncident(
        page,
        uniqueIncidentKey("FEEP401FOREIGN"),
        "FE-E-P4-01 foreign paste target",
      );
      const foreign = await createViewRow(
        page,
        foreignIncidentId,
        timelineViewSchemaId,
        {
          client_txn_id: uniqueTxn("feep401-foreign-row"),
          "timeline.summary": "FE-E-P4-01 foreign untouched",
        },
      );
      await expectPublicError(
        await postTimelinePastePublic(page, incidentId, {
          view_schema_id: timelineViewSchemaId,
          client_txn_id: uniqueTxn("feep401-invalid-paste-target"),
          clipboard_text:
            "FE-E-P4-01 should not create\nFE-E-P4-01 should not disclose",
          format: "tsv",
          start_field_key: "timeline.summary",
          columns: ["timeline.summary"],
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
          (row) => summaryValue(row) === "FE-E-P4-01 should not create",
        ),
      ).toBe(false);
      const foreignRowsAfterRejectedPaste = await queryViewRows(
        page,
        foreignIncidentId,
        timelineViewSchemaId,
      );
      expect(summaryValue(foreignRowsAfterRejectedPaste[0])).toBe(
        "FE-E-P4-01 foreign untouched",
      );

      const staleIncidentId = await createIncident(
        page,
        uniqueIncidentKey("FEEP401STALE"),
        "FE-E-P4-01 grouped paste conflict",
      );
      const staleFirst = await createViewRow(
        page,
        staleIncidentId,
        timelineViewSchemaId,
        {
          client_txn_id: uniqueTxn("feep401-stale-first"),
          "timeline.summary": "FE-E-P4-01 stale first base",
        },
      );
      const staleSecond = await createViewRow(
        page,
        staleIncidentId,
        timelineViewSchemaId,
        {
          client_txn_id: uniqueTxn("feep401-stale-second"),
          "timeline.summary": "FE-E-P4-01 stale second base",
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
        await stalePage
          .getByTestId(rowCellTestId(staleFirst.record_id, "timeline.summary"))
          .focus();
        await expect(stalePage.getByTestId("workbook-focus-anchor")).toHaveText(
          `${timelineViewSchemaId}:${staleFirst.record_id}:timeline.summary`,
        );
        await patchTimelineRecord(page, staleFirst.record_id, {
          view_schema_id: timelineViewSchemaId,
          base_row_version: staleFirst.row_version,
          client_txn_id: uniqueTxn("feep401-stale-first-server"),
          changes: [
            {
              field_key: "timeline.summary",
              value: "FE-E-P4-01 stale first server",
            },
          ],
        });
        await patchTimelineRecord(page, staleSecond.record_id, {
          view_schema_id: timelineViewSchemaId,
          base_row_version: staleSecond.row_version,
          client_txn_id: uniqueTxn("feep401-stale-second-server"),
          changes: [
            {
              field_key: "timeline.summary",
              value: "FE-E-P4-01 stale second server",
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
          staleFirst.record_id,
          "timeline.summary",
          [
            "FE-E-P4-01 stale first client",
            "FE-E-P4-01 stale second client",
            "FE-E-P4-01 stale created",
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
            conflictMarkerTestId(staleFirst.record_id, "timeline.summary"),
          ),
        ).toBeVisible();
        await expect(
          stalePage.getByTestId(
            conflictMarkerTestId(staleSecond.record_id, "timeline.summary"),
          ),
        ).toBeVisible();
        await expect(
          stalePage.getByTestId("paste-conflict-navigator"),
        ).toBeVisible();
        await expect(
          stalePage.getByTestId("paste-conflict-position"),
        ).toHaveText("1 of 2");
        await expect(stalePage.getByTestId("conflict-local-value")).toHaveValue(
          "FE-E-P4-01 stale first client",
        );
        await stalePage.getByTestId("paste-conflict-next").click();
        await expect(
          stalePage.getByTestId("paste-conflict-position"),
        ).toHaveText("2 of 2");
        await expect(stalePage.getByTestId("conflict-local-value")).toHaveValue(
          "FE-E-P4-01 stale second client",
        );
      } finally {
        await staleContext.close();
      }
    });

    await test.step("public validation errors preserve local drafts without private details", async () => {
      const validationResponse = page.waitForResponse(
        (response) =>
          response.request().method() === "PATCH" &&
          response.url().endsWith(`/api/v1/records/${beta.record_id}`),
      );
      await page
        .getByTestId(rowCellTestId(beta.record_id, "timeline.occurred_at"))
        .fill("not-a-timestamp");
      await page
        .getByTestId(rowCellTestId(beta.record_id, "timeline.occurred_at"))
        .press("Enter");
      await expectPublicError(
        await validationResponse,
        400,
        "invalid_mutation_payload",
        { field: "timeline.occurred_at", reason_code: "invalid_value" },
      );
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();
      await expect(
        page.getByTestId(rowCellTestId(beta.record_id, "timeline.occurred_at")),
      ).toHaveValue("not-a-timestamp");
      page.once("dialog", async (dialog) => {
        await dialog.accept();
      });
      await page.reload();
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    });

    await test.step("real session revocation replays queued public PATCH work in FIFO order", async () => {
      await exerciseRevokedPendingReplay({
        createdBy: "FE-E-P4-01",
        incidentKeyPrefix: "FEEP401REVOKE",
        localValues: [
          "FE-E-P4-01 revoked A local",
          "FE-E-P4-01 revoked B local",
          "FE-E-P4-01 revoked C local",
        ],
        page,
        scenario: "revoked",
        sessionTracker,
        triggerRevocation: async ({ member }) => {
          await revokeAllSessions(
            workerAdminRequest,
            member.user_id,
            "FE-E-P4-01 browser revoke-all",
          );
        },
      });
    });
  },
);
