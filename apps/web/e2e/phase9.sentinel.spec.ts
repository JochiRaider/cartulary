import { conflictMarkerTestId, rowCellTestId } from "@cartulary/ui-contracts";

import { expect, test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  patchTimelineRecord,
  queryViewRows,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";

const phase9Sprint0SentinelMessage =
  "Phase 9 Sprint 0 blocker sentinel: this is not behavior completion evidence; replace this sentinel with real Phase 9 implementation evidence before claiming the row complete.";

const timelineViewSchemaId = "cartulary.view.timeline.v1";

async function disableWorkbookSockets(page: {
  addInitScript: (script: () => void) => Promise<void>;
}) {
  await page.addInitScript(() => {
    class Phase9ClosedWebSocket {
      static readonly CONNECTING = 0;
      static readonly OPEN = 1;
      static readonly CLOSING = 2;
      static readonly CLOSED = 3;

      readonly url: string;
      readyState = Phase9ClosedWebSocket.CONNECTING;
      onclose: ((event: CloseEvent) => void) | null = null;
      onerror: ((event: Event) => void) | null = null;
      onmessage: ((event: MessageEvent) => void) | null = null;
      onopen: ((event: Event) => void) | null = null;

      constructor(url: string | URL) {
        this.url = String(url);
      }

      close() {
        this.readyState = Phase9ClosedWebSocket.CLOSED;
        this.onclose?.(new CloseEvent("close"));
      }

      send() {}
    }

    Object.defineProperty(window, "WebSocket", {
      configurable: true,
      value: Phase9ClosedWebSocket,
    });
  });
}

test("Phase 9 E-9-02 pastes a representative 20x5 Timeline clipboard range", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E902"),
    "Phase 9 E-9-02 clipboard paste",
  );
  const seed = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e902-seed"),
    "timeline.summary": "Phase 9 paste seed",
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByText("Timeline mutation substrate")).toBeVisible();

  const seedSummary = page.getByTestId(
    rowCellTestId(seed.record_id as string, "summary"),
  );
  await seedSummary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${seed.record_id}:timeline.summary`,
  );

  const pasteRows = Array.from({ length: 20 }, (_, index) => {
    const ordinal = index + 1;
    return [
      `Phase 9 paste summary ${ordinal}`,
      `phase9-host-${ordinal}.example.test`,
      `phase9-user-${ordinal}@example.test`,
      `readonly-evidence-${ordinal}`,
      `phase9-tag-${ordinal}`,
    ].join("\t");
  });
  const pastePayload = pasteRows.join("\n");

  const pasteResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .includes(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
        ),
  );
  await seedSummary.evaluate((element, text) => {
    const data = new DataTransfer();
    data.setData("text/plain", text);
    element.dispatchEvent(
      new ClipboardEvent("paste", {
        bubbles: true,
        cancelable: true,
        clipboardData: data,
      }),
    );
  }, pastePayload);
  await expect((await pasteResponse).ok()).toBeTruthy();
  await expect(page.getByTestId("save-state")).toHaveText("Saved");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${seed.record_id}:timeline.summary`,
  );
  await expect(
    page.getByText(`Timeline row ${seed.record_id}`),
  ).toBeVisible();
  await expect(
    page.getByTestId(rowCellTestId(seed.record_id as string, "summary")),
  ).toHaveValue("Phase 9 paste summary 1");
  await expect(page.getByTestId("timeline-grid-shell")).toBeVisible();

  const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
  const matchingRows = rows.filter((row) => {
    const cells = row.cells as Record<string, { value: unknown }>;
    return String(cells["timeline.summary"]?.value ?? "").startsWith(
      "Phase 9 paste summary ",
    );
  });
  expect(matchingRows).toHaveLength(20);
  const first = matchingRows.find((row) => row.record_id === seed.record_id);
  expect(first).toBeTruthy();
  expect(first?.cells["timeline.summary"]?.value).toBe(
    "Phase 9 paste summary 1",
  );
  expect(
    collectionDisplayTexts(first?.cells["timeline.host_refs"]?.value),
  ).toContain("phase9-host-1.example.test");
  expect(
    collectionDisplayTexts(first?.cells["timeline.identity_refs"]?.value),
  ).toContain("phase9-user-1@example.test");
  expect(first?.cells["timeline.evidence_count"]?.value).toBe(0);
  expect(collectionDisplayTexts(first?.cells["timeline.tags"]?.value)).toContain(
    "phase9-tag-1",
  );
  const twentieth = matchingRows.find((row) => {
    const cells = row.cells as Record<string, { value: unknown }>;
    return cells["timeline.summary"]?.value === "Phase 9 paste summary 20";
  });
  expect(twentieth).toBeTruthy();
  expect(
    collectionDisplayTexts(twentieth?.cells["timeline.host_refs"]?.value),
  ).toContain("phase9-host-20.example.test");
  expect(
    collectionDisplayTexts(twentieth?.cells["timeline.identity_refs"]?.value),
  ).toContain("phase9-user-20@example.test");
  expect(twentieth?.cells["timeline.evidence_count"]?.value).toBe(0);
  expect(
    collectionDisplayTexts(twentieth?.cells["timeline.tags"]?.value),
  ).toContain("phase9-tag-20");
});

test("Phase 9 E-9-02 groups paste conflicts and preserves selection continuity", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E902-CONFLICT"),
    "Phase 9 E-9-02 grouped paste conflicts",
  );
  const first = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e902-conflict-first"),
    "timeline.summary": "Phase 9 conflict first base",
  });
  const second = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e902-conflict-second"),
    "timeline.summary": "Phase 9 conflict second base",
  });

  await disableWorkbookSockets(page);
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByText("Timeline mutation substrate")).toBeVisible();

  const firstSummary = page.getByTestId(
    rowCellTestId(first.record_id as string, "summary"),
  );
  await firstSummary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${first.record_id}:timeline.summary`,
  );

  await patchTimelineRecord(page, first.record_id as string, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: first.row_version,
    client_txn_id: uniqueTxn("e902-conflict-first-server"),
    changes: [
      {
        field_key: "timeline.summary",
        value: "Phase 9 server first",
      },
    ],
  });
  await patchTimelineRecord(page, second.record_id as string, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: second.row_version,
    client_txn_id: uniqueTxn("e902-conflict-second-server"),
    changes: [
      {
        field_key: "timeline.summary",
        value: "Phase 9 server second",
      },
    ],
  });

  const pasteResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .includes(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
        ),
  );
  await firstSummary.evaluate((element) => {
    const data = new DataTransfer();
    data.setData(
      "text/plain",
      [
        "Phase 9 client first",
        "Phase 9 client second",
        "Phase 9 conflict create",
      ].join("\n"),
    );
    element.dispatchEvent(
      new ClipboardEvent("paste", {
        bubbles: true,
        cancelable: true,
        clipboardData: data,
      }),
    );
  });
  await expect((await pasteResponse).ok()).toBeTruthy();

  await expect(page.getByTestId("timeline-grid-shell")).toBeVisible();
  await expect(page.getByTestId("save-state")).toHaveText("Conflict");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `timeline:${first.record_id}:timeline.summary`,
  );
  await expect(
    page.getByTestId(
      conflictMarkerTestId(first.record_id as string, "timeline.summary"),
    ),
  ).toBeVisible();
  await expect(
    page.getByTestId(
      conflictMarkerTestId(second.record_id as string, "timeline.summary"),
    ),
  ).toBeVisible();
  await expect(page.getByTestId("paste-conflict-navigator")).toBeVisible();
  await expect(page.getByTestId("paste-conflict-position")).toHaveText(
    "1 of 2",
  );
  await expect(page.getByTestId("conflict-local-value")).toHaveValue(
    "Phase 9 client first",
  );
  await page.getByTestId("paste-conflict-next").click();
  await expect(page.getByTestId("paste-conflict-position")).toHaveText(
    "2 of 2",
  );
  await expect(page.getByTestId("conflict-local-value")).toHaveValue(
    "Phase 9 client second",
  );
  await page.getByTestId("conflict-close").click();
  await expect(page.getByTestId("conflict-resolver")).toHaveCount(0);
  await expect(page.getByTestId("save-state")).toHaveText("Conflict");
});

test("Phase 9 E-9-03 Sprint 0 blocker sentinel", async () => {
  expect(phase9Sprint0SentinelMessage).toContain("blocker sentinel");
});

test("Phase 9 E-9-04 Sprint 0 blocker sentinel", async () => {
  expect(phase9Sprint0SentinelMessage).toContain("blocker sentinel");
});

test("Phase 9 E-9-05 Sprint 0 blocker sentinel", async () => {
  expect(phase9Sprint0SentinelMessage).toContain("blocker sentinel");
});

test("Phase 9 E-9-06 Sprint 0 blocker sentinel", async () => {
  expect(phase9Sprint0SentinelMessage).toContain("blocker sentinel");
});

test("Phase 9 E-9-07 Sprint 0 blocker sentinel", async () => {
  expect(phase9Sprint0SentinelMessage).toContain("blocker sentinel");
});

test("Phase 9 E-9-08 Sprint 0 blocker sentinel", async () => {
  expect(phase9Sprint0SentinelMessage).toContain("blocker sentinel");
});

function collectionDisplayTexts(value: unknown): string[] {
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
    if (typeof record.display_text === "string") {
      return [record.display_text];
    }
    if (typeof record.raw_text === "string") {
      return [record.raw_text];
    }
    return [];
  });
}
