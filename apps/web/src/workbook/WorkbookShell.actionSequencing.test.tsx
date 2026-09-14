import {
  gridRowTestId,
  gridRowVersionAttribute,
  rowCellTestId,
  rowInspectButtonTestId,
  rowInspectorFieldTestId,
  saveStateTestId,
  timelineCaptureActionTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowSupersedeButtonTestId,
} from "@cartulary/ui-contracts";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../testing/fetchMockTestSupport";
import { TimelineWorkbookRuntimeFixture } from "../testing/TimelineWorkbookRuntimeFixture";
import {
  changeInputValue,
  cleanupTimelineWorkbookTestGlobals,
  extractTimelineRecordActionBody,
  extractTimelineRecordPatchBody,
  flushWorkbookAsync,
  installTimelineWorkbookTestGlobals,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRecordActionCalls,
  timelineRow,
  waitForTimelineRecordActionCalls,
  waitForTimelineRecordPatchCalls,
} from "../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);
const incidentId = "10000000-0000-4000-8000-000000000001";
const recordId = "20000000-0000-4000-8000-000000000101";
const replacementId = "20000000-0000-4000-8000-000000000102";
const reason = "Duplicate observation; use the independently verified source";
describe("Timeline workbook action sequencing", () => {
  let fetchMock: TimelineWorkbookFetchMock;
  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });
  afterEach(cleanupTimelineWorkbookTestGlobals);
  async function context() {
    fireEvent.contextMenu(
      await screen.findByTestId(
        rowCellTestId(recordId, "timeline.activity_synopsis_text"),
      ),
      { clientX: 32, clientY: 48 },
    );
  }
  async function inspect() {
    await context();
    fireEvent.click(
      await screen.findByTestId(rowInspectButtonTestId(recordId)),
    );
  }
  async function mark() {
    await context();
    fireEvent.click(
      await screen.findByTestId(timelineRowMarkReviewedButtonTestId(recordId)),
    );
  }
  async function authorSupersession() {
    await context();
    fireEvent.click(
      await screen.findByTestId(timelineRowSupersedeButtonTestId(recordId)),
    );
    const editor = await screen.findByTestId(
      timelineCaptureActionTestId("editor", recordId),
    );
    expect(editor.textContent).toContain("No replacement");
    fireEvent.change(
      await screen.findByTestId(
        timelineCaptureActionTestId("reason", recordId),
      ),
      { target: { value: reason } },
    );
    await screen.findByRole("option", {
      name: new RegExp(`Replacement.*${replacementId}`),
    });
    fireEvent.change(
      screen.getByTestId(timelineCaptureActionTestId("replacement", recordId)),
      { target: { value: replacementId } },
    );
    fireEvent.click(
      screen.getByTestId(timelineCaptureActionTestId("review", recordId)),
    );
  }
  async function confirm() {
    fireEvent.click(
      await screen.findByTestId(
        timelineCaptureActionTestId("confirm", recordId),
      ),
    );
  }
  async function version(expected: number) {
    await waitFor(() =>
      expect(
        screen
          .getByTestId(gridRowTestId(timelineViewSchemaId, recordId))
          .getAttribute(gridRowVersionAttribute),
      ).toBe(String(expected)),
    );
  }
  function service(captureState = "rough", rowVersion = 1) {
    let state = timelineRow({
      recordId,
      rowVersion,
      summary: "Alpha",
      details: "Original details",
      captureState,
    });
    let replacement: string | null = null;
    const initial = state;
    let patchGate: Promise<unknown> | null = null;
    let staleQueries = 0;
    fetchMock.mockImplementation(async (input, init) => {
      const url = String(input),
        method = init?.method ?? "GET";
      if (url.endsWith("/query")) {
        const row = staleQueries-- > 0 ? initial : state;
        return successEnvelope({
          incident_id: incidentId,
          view_schema_id: timelineViewSchemaId,
          rows: [
            {
              ...row,
              cells: {
                ...row.cells,
                "timeline.replacement_record_id": { value: replacement },
              },
            },
            timelineRow({
              recordId: replacementId,
              rowVersion: 1,
              summary: "Replacement",
              captureState: "rough",
            }),
          ],
        });
      }
      if (url.endsWith("/history"))
        return new Response(
          JSON.stringify({
            data: {
              record_id: recordId,
              incident_id: incidentId,
              row_version: state.row_version,
              deleted: false,
              items: [],
            },
            meta: {
              request_id: "history",
              paging: { has_more: false, limit: 50, next_cursor: null },
            },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      if (method === "PATCH") {
        if (patchGate !== null) await patchGate;
        const body = JSON.parse(String(init?.body));
        expect(body.base_row_version).toBe(state.row_version);
        state = timelineRow({
          recordId,
          rowVersion: state.row_version + 1,
          summary: "Alpha",
          details: body.changes[0].value,
          captureState: "enriched",
        });
        return successEnvelope({
          view_schema_id: timelineViewSchemaId,
          change_set_id: "30000000-0000-4000-8000-000000000102",
          row: state,
        });
      }
      if (url.endsWith("/mark-reviewed") || url.endsWith("/supersede")) {
        const body = JSON.parse(String(init?.body));
        expect(body.base_row_version).toBe(state.row_version);
        const nextState = url.endsWith("/mark-reviewed")
          ? "reviewed"
          : "superseded";
        state = {
          ...state,
          row_version: state.row_version + 1,
          cells: {
            ...state.cells,
            "timeline.capture_state": { value: nextState },
          },
        };
        replacement = body.replacement_record_id ?? null;
        return successEnvelope({
          record_id: recordId,
          incident_id: incidentId,
          row_version: state.row_version,
          capture_state: nextState,
          change_set_id: "30000000-0000-4000-8000-000000000103",
          reason: body.reason ?? null,
          replacement_record_id: replacement,
        });
      }
      return new Response("{}", { status: 404 });
    });
    return {
      gatePatch: (promise: Promise<unknown>) => {
        patchGate = promise;
      },
      staleNextQueries: (count: number) => {
        staleQueries = count;
      },
    };
  }
  function renderWorkbook() {
    render(
      <TimelineWorkbookRuntimeFixture
        incidentId={incidentId}
        currentIncidentRole="reviewer"
      />,
    );
  }
  async function editDetails() {
    await inspect();
    const input = (await screen.findByTestId(
      rowInspectorFieldTestId(recordId, "timeline.raw_activity_text"),
    )) as HTMLTextAreaElement;
    await changeInputValue(input, "Material edit after review");
    fireEvent.blur(input);
    await waitForTimelineRecordPatchCalls(fetchMock, 1);
    expect(extractTimelineRecordPatchBody(fetchMock, 0)).toMatchObject({
      base_row_version: 2,
      changes: [
        {
          field_key: "timeline.raw_activity_text",
          value: "Material edit after review",
        },
      ],
    });
  }
  it("sends Timeline actions with the current row version after earlier workbook mutations", async () => {
    service();
    renderWorkbook();
    await screen.findByTestId(saveStateTestId());
    await mark();
    await waitForTimelineRecordActionCalls(fetchMock, "mark-reviewed", 1);
    expect(
      extractTimelineRecordActionBody(fetchMock, "mark-reviewed"),
    ).toMatchObject({ base_row_version: 1 });
    expect(
      extractTimelineRecordActionBody(fetchMock, "mark-reviewed"),
    ).not.toHaveProperty("reason");
    expect(
      screen.queryByTestId(timelineCaptureActionTestId("confirm", recordId)),
    ).toBeNull();
    await version(2);
    await editDetails();
    await version(3);
    await authorSupersession();
    expect(timelineRecordActionCalls(fetchMock, "supersede")).toHaveLength(0);
    await confirm();
    await waitForTimelineRecordActionCalls(fetchMock, "supersede", 1);
    expect(
      extractTimelineRecordActionBody(fetchMock, "supersede"),
    ).toMatchObject({
      base_row_version: 3,
      reason,
      replacement_record_id: replacementId,
    });
    await version(4);
  });
  it("waits for a pending material autosave before supersede dispatch", async () => {
    const pending = deferred<void>(),
      state = service("reviewed", 2);
    state.gatePatch(pending.promise);
    renderWorkbook();
    await editDetails();
    await authorSupersession();
    await flushWorkbookAsync();
    expect(timelineRecordActionCalls(fetchMock, "supersede")).toHaveLength(0);
    pending.resolve();
    await version(3);
    await waitFor(() =>
      expect(
        (
          screen.getByTestId(
            timelineCaptureActionTestId("review", recordId),
          ) as HTMLButtonElement
        ).disabled,
      ).toBe(false),
    );
    expect(
      screen.queryByTestId(timelineCaptureActionTestId("confirm", recordId)),
    ).toBeNull();
    expect(timelineRecordActionCalls(fetchMock, "supersede")).toHaveLength(0);
    fireEvent.click(
      screen.getByTestId(timelineCaptureActionTestId("review", recordId)),
    );
    await confirm();
    await waitForTimelineRecordActionCalls(fetchMock, "supersede", 1);
    expect(
      extractTimelineRecordActionBody(fetchMock, "supersede"),
    ).toMatchObject({
      base_row_version: 3,
      reason,
      replacement_record_id: replacementId,
    });
    await version(4);
    expect(
      screen.getByTestId(rowCellTestId(recordId, "timeline.capture_state"))
        .textContent,
    ).toBe("superseded");
  });
  it("keeps action result row versions ahead of stale projection reloads", async () => {
    const state = service();
    renderWorkbook();
    await screen.findByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    );
    state.staleNextQueries(1);
    await mark();
    await waitForTimelineRecordActionCalls(fetchMock, "mark-reviewed", 1);
    expect(
      extractTimelineRecordActionBody(fetchMock, "mark-reviewed"),
    ).toMatchObject({ base_row_version: 1 });
    await version(2);
    expect(
      screen.getByTestId(rowCellTestId(recordId, "timeline.capture_state"))
        .textContent,
    ).toBe("reviewed");
    await authorSupersession();
    await confirm();
    await waitForTimelineRecordActionCalls(fetchMock, "supersede", 1);
    expect(
      extractTimelineRecordActionBody(fetchMock, "supersede"),
    ).toMatchObject({
      base_row_version: 2,
      replacement_record_id: replacementId,
    });
    await version(3);
  });
  it("requires a second Mark reviewed click after a queued material edit commits", async () => {
    const pending = deferred<void>(),
      state = service("enriched", 2);
    state.gatePatch(pending.promise);
    renderWorkbook();
    await editDetails();
    await mark();
    await flushWorkbookAsync();
    expect(timelineRecordActionCalls(fetchMock, "mark-reviewed")).toHaveLength(
      0,
    );
    pending.resolve();
    await version(3);
    await waitFor(() =>
      expect(
        screen.getByText(
          "Timeline action needs another review. No action was sent.",
        ),
      ).toBeTruthy(),
    );
    expect(
      screen.queryByTestId(timelineCaptureActionTestId("confirm", recordId)),
    ).toBeNull();
    expect(timelineRecordActionCalls(fetchMock, "mark-reviewed")).toHaveLength(
      0,
    );
    await mark();
    await waitForTimelineRecordActionCalls(fetchMock, "mark-reviewed", 1);
    expect(
      extractTimelineRecordActionBody(fetchMock, "mark-reviewed"),
    ).toMatchObject({ base_row_version: 3 });
    expect(
      extractTimelineRecordActionBody(fetchMock, "mark-reviewed"),
    ).not.toHaveProperty("reason");
    await version(4);
  });
  it("fences a changed inspector selection while Timeline action preparation waits", async () => {
    const pending = deferred<void>(),
      state = service("enriched", 2);
    state.gatePatch(pending.promise);
    renderWorkbook();
    await editDetails();
    await mark();
    fireEvent.contextMenu(
      await screen.findByTestId(
        rowCellTestId(replacementId, "timeline.activity_synopsis_text"),
      ),
      { clientX: 32, clientY: 48 },
    );
    fireEvent.click(
      await screen.findByTestId(rowInspectButtonTestId(replacementId)),
    );
    expect(
      await screen.findByTestId(
        rowInspectorFieldTestId(replacementId, "timeline.raw_activity_text"),
      ),
    ).toBeTruthy();
    pending.resolve();
    await version(3);
    await flushWorkbookAsync();
    expect(timelineRecordActionCalls(fetchMock, "mark-reviewed")).toHaveLength(
      0,
    );
    expect(
      screen.queryByTestId(
        rowInspectorFieldTestId(recordId, "timeline.raw_activity_text"),
      ),
    ).toBeNull();
    expect(
      await screen.findByTestId(
        rowInspectorFieldTestId(replacementId, "timeline.raw_activity_text"),
      ),
    ).toBeTruthy();
  });
});
