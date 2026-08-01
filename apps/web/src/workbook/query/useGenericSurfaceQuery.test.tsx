import { requireViewContract } from "@cartulary/view-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  deferred,
  errorResponse,
  jsonResponse,
} from "../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { createWorkbookViewQueryAdapter } from "../adapters/createWorkbookViewQueryAdapter";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import {
  findingsViewSchemaId,
  notesViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { useGenericSurfaceQuery } from "./useGenericSurfaceQuery";
import type { WorkbookQueryRow } from "./WorkbookQueryRow";

const findingsContract = requireViewContract(findingsViewSchemaId);
const notesContract = requireViewContract(notesViewSchemaId);
const incidentId = "00000000-0000-4000-8000-000000000001";
const noteCurrentId = "00000000-0000-4000-8000-000000000101";
const noteObsoleteId = "00000000-0000-4000-8000-000000000102";
const findingCurrentId = "00000000-0000-4000-8000-000000000103";
const viewQuery = createWorkbookViewQueryAdapter({
  apiBase: undefined,
  incidentId,
});

function noteRow(recordId: string, rowVersion: number, body: string) {
  return fullWorkbookViewRow(notesContract, recordId, rowVersion, {
    "note.body": body,
    "note.title": `${recordId} title`,
  });
}

function findingRow(recordId: string, rowVersion: number) {
  return fullWorkbookViewRow(findingsContract, recordId, rowVersion, {
    "finding.title": `${recordId} title`,
  });
}

function withoutLocalViewSchema(row: unknown): unknown {
  const { view_schema_id: _viewSchemaId, ...wireRow } = row as Record<
    string,
    unknown
  >;
  return wireRow;
}

function queryResponse(viewSchemaId: string, rows: readonly unknown[]) {
  return jsonResponse({
    data: {
      incident_id: incidentId,
      view_schema_id: viewSchemaId,
      rows: rows.map(withoutLocalViewSchema),
    },
    meta: {
      query: { filters: [], sort: [] },
      request_id: "req-query",
    },
  });
}

function GenericQueryHarness({
  active = true,
  onIncidentAccessLost,
  viewSchemaId = notesViewSchemaId,
}: {
  readonly active?: boolean;
  readonly onIncidentAccessLost?: (() => void) | undefined;
  readonly viewSchemaId?: string;
}) {
  const query = useGenericSurfaceQuery({
    active,
    contract: requireViewContract(viewSchemaId),
    onIncidentAccessLost,
    queryState: emptyWorkbookQueryState(),
    viewQuery,
    viewSchemaId,
  });
  return (
    <>
      <button onClick={() => void query.refresh()} type="button">
        refresh
      </button>
      <button
        onClick={() =>
          query.applyRecordChanged({
            record_id: noteCurrentId,
            row_version: 2,
            change_set_id: "change-1",
            client_txn_id: "txn-1",
            actor_user_id: "user-2",
            changed_field_keys: ["note.body"],
            affected_views: [
              {
                view_schema_id: notesViewSchemaId,
                change_kind: "patch",
                patch_cells: {
                  record_id: noteCurrentId,
                  row_version: 2,
                  cells: {
                    "note.body": { value: "Patched note" },
                  },
                },
              },
            ],
          })
        }
        type="button"
      >
        patch
      </button>
      <output aria-label="generic-load-state">{query.loadState.kind}</output>
      <output aria-label="generic-rows">
        {query.rows
          .map(
            (row) =>
              `${row.record_id}:${String(row.cells["note.body"]?.value ?? "")}`,
          )
          .join(",")}
      </output>
      <output aria-label="generic-row-schema">
        {query.rows
          .map(
            (row) =>
              (
                row as WorkbookQueryRow & {
                  readonly view_schema_id?: string;
                }
              ).view_schema_id ?? "",
          )
          .join(",")}
      </output>
    </>
  );
}

describe("useGenericSurfaceQuery", () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("normalizes Notes rows and applies a newer sparse live patch", async () => {
    const { view_schema_id: _ignored, ...rawNote } = noteRow(
      noteCurrentId,
      1,
      "Current note",
    );
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.resolve(queryResponse(notesViewSchemaId, [rawNote]))),
    );
    render(<GenericQueryHarness />);

    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() =>
      expect(screen.getByLabelText("generic-load-state").textContent).toBe(
        "ready",
      ),
    );
    expect(screen.getByLabelText("generic-row-schema").textContent).toBe(
      notesViewSchemaId,
    );
    fireEvent.click(screen.getByRole("button", { name: "patch" }));
    await waitFor(() =>
      expect(screen.getByLabelText("generic-rows").textContent).toBe(
        `${noteCurrentId}:Patched note`,
      ),
    );
  });

  it("rejects obsolete schema work and retains current rows after a mismatched envelope", async () => {
    const staleNotes = deferred<Response>();
    let staleSignal: AbortSignal | null | undefined;
    let callCount = 0;
    vi.stubGlobal(
      "fetch",
      vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
        callCount += 1;
        if (callCount === 1) {
          staleSignal = init?.signal;
          return staleNotes.promise;
        }
        if (callCount === 3) {
          return Promise.resolve(queryResponse(notesViewSchemaId, []));
        }
        const viewSchemaId = String(input).includes(findingsViewSchemaId)
          ? findingsViewSchemaId
          : notesViewSchemaId;
        return Promise.resolve(
          queryResponse(viewSchemaId, [findingRow(findingCurrentId, 2)]),
        );
      }),
    );
    const rendered = render(<GenericQueryHarness />);

    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() => expect(callCount).toBe(1));
    rendered.rerender(
      <GenericQueryHarness viewSchemaId={findingsViewSchemaId} />,
    );
    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() =>
      expect(screen.getByLabelText("generic-rows").textContent).toContain(
        findingCurrentId,
      ),
    );
    expect(staleSignal?.aborted).toBe(true);
    staleNotes.resolve(
      queryResponse(notesViewSchemaId, [
        noteRow(noteObsoleteId, 1, "Obsolete note"),
      ]),
    );
    await Promise.resolve();
    await Promise.resolve();
    expect(screen.getByLabelText("generic-rows").textContent).toContain(
      findingCurrentId,
    );

    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() =>
      expect(screen.getByLabelText("generic-load-state").textContent).toBe(
        "stale_error",
      ),
    );
    expect(screen.getByLabelText("generic-rows").textContent).toContain(
      findingCurrentId,
    );
  });

  it("clears access-protected rows, stays idle while inactive, and aborts on teardown", async () => {
    const onIncidentAccessLost = vi.fn();
    const pending = deferred<Response>();
    let responseKind: "ready" | "denied" | "pending" = "ready";
    let pendingSignal: AbortSignal | null | undefined;
    const fetchMock = vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
      if (responseKind === "denied") {
        return Promise.resolve(errorResponse("authorization_denied", 403));
      }
      if (responseKind === "pending") {
        pendingSignal = init?.signal;
        return pending.promise;
      }
      return Promise.resolve(
        queryResponse(notesViewSchemaId, [
          noteRow(noteCurrentId, 1, "Current note"),
        ]),
      );
    });
    vi.stubGlobal("fetch", fetchMock);
    const rendered = render(
      <GenericQueryHarness onIncidentAccessLost={onIncidentAccessLost} />,
    );

    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() =>
      expect(screen.getByLabelText("generic-load-state").textContent).toBe(
        "ready",
      ),
    );
    responseKind = "denied";
    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() =>
      expect(screen.getByLabelText("generic-load-state").textContent).toBe(
        "permission_denied",
      ),
    );
    expect(onIncidentAccessLost).toHaveBeenCalledOnce();
    expect(screen.getByLabelText("generic-rows").textContent).toBe("");

    rendered.rerender(
      <GenericQueryHarness
        active={false}
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    expect(fetchMock).toHaveBeenCalledTimes(2);

    responseKind = "pending";
    rendered.rerender(
      <GenericQueryHarness onIncidentAccessLost={onIncidentAccessLost} />,
    );
    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(3));
    rendered.unmount();
    expect(pendingSignal?.aborted).toBe(true);
  });
});
