import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
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
import {
  acceptedQueryMetadata,
  renderWithWorkbookQueryBrowsing as render,
  renderHookWithWorkbookQueryBrowsing as renderHook,
  workbookQueryMeta,
} from "../../testing/workbookQueryTestSupport";
import { createWorkbookViewQueryAdapter } from "../adapters/createWorkbookViewQueryAdapter";
import { WorkbookAssessmentAuthoringOwner } from "../features/assessments/WorkbookAssessmentAuthoringOwner";
import {
  emptyWorkbookQueryState,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import { assessmentsViewSchemaId } from "../models/workbookSurfaceRegistry";
import { useAssessmentSurfaceQuery } from "./useAssessmentSurfaceQuery";

const assessmentsContract = requireViewContract(assessmentsViewSchemaId);
const incidentId = "00000000-0000-4000-8000-000000000001";
const assessmentCurrentId = "00000000-0000-4000-8000-000000000301";
const assessmentObsoleteId = "00000000-0000-4000-8000-000000000302";
const viewQuery = createWorkbookViewQueryAdapter({
  apiBase: undefined,
  incidentId,
});

function assessmentRow(
  recordId: string,
  rowVersion: number,
  rationale: string,
) {
  return fullWorkbookViewRow(assessmentsContract, recordId, rowVersion, {
    "assessment.subject_ref": "host-1",
    "assessment.subject_type": "host",
    "assessment.assessment_state": "suspected",
    "assessment.rationale": rationale,
  });
}

function withoutLocalViewSchema(row: unknown): unknown {
  const { view_schema_id: _viewSchemaId, ...wireRow } = row as Record<
    string,
    unknown
  >;
  return wireRow;
}

function queryResponse(rows: readonly unknown[]) {
  return jsonResponse({
    data: {
      incident_id: incidentId,
      view_schema_id: assessmentsViewSchemaId,
      rows: rows.map(withoutLocalViewSchema),
    },
    meta: workbookQueryMeta(assessmentsViewSchemaId),
  });
}

function AssessmentQueryHarness({
  active = true,
  onAuthorityUncertain,
  queryState = emptyWorkbookQueryState(),
}: {
  readonly active?: boolean;
  readonly onAuthorityUncertain?: (() => void) | undefined;
  readonly queryState?: WorkbookQueryState;
}) {
  const query = useAssessmentSurfaceQuery({
    active,
    onAuthorityUncertain,
    queryState,
    viewQuery,
  });
  return (
    <>
      <button onClick={() => void query.refresh()} type="button">
        refresh
      </button>
      <button
        onClick={() =>
          query.applyRecordChanged({
            record_id: assessmentCurrentId,
            row_version: 2,
            change_set_id: "change-1",
            client_txn_id: "txn-1",
            actor_user_id: "user-2",
            changed_field_keys: ["assessment.rationale"],
            affected_views: [
              {
                view_schema_id: assessmentsViewSchemaId,
                change_kind: "patch",
                patch_cells: {
                  record_id: assessmentCurrentId,
                  row_version: 2,
                  cells: {
                    "assessment.rationale": { value: "Patched rationale" },
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
      <output aria-label="assessment-load-state">{query.loadState.kind}</output>
      <output aria-label="assessment-rows">
        {query.rows
          .map(
            (row) =>
              `${row.record_id}:${String(
                row.cells["assessment.rationale"]?.value ?? "",
              )}`,
          )
          .join(",")}
      </output>
    </>
  );
}

describe("useAssessmentSurfaceQuery", () => {
  it("requires an accepted current query before authorization recovery can resume", async () => {
    const onAuthorityUncertain = vi.fn();
    const query = vi.fn().mockResolvedValue({
      kind: "rejected",
      failure: { kind: "invalid_contract", message: "Malformed query" },
    });
    const view = renderHook(() =>
      useAssessmentSurfaceQuery({
        active: true,
        queryState: emptyWorkbookQueryState(),
        onAuthorityUncertain,
        viewQuery: { query },
      }),
    );
    await act(async () => {
      await expect(
        view.result.current.refresh({ requireAcceptance: true }),
      ).rejects.toMatchObject({
        recovery: { kind: "unavailable", failure: "contract" },
      });
    });
    expect(onAuthorityUncertain).not.toHaveBeenCalled();
    query.mockResolvedValue({ kind: "aborted" });
    await act(async () => {
      await expect(
        view.result.current.refresh({ requireAcceptance: true }),
      ).rejects.toMatchObject({ recovery: { kind: "cancelled" } });
    });
    view.unmount();
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("rejects a stale rapid-filter response and applies a newer live patch", async () => {
    const staleResponse = deferred<Response>();
    let staleSignal: AbortSignal | null | undefined;
    let callCount = 0;
    vi.stubGlobal(
      "fetch",
      vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
        callCount += 1;
        if (callCount === 1) {
          staleSignal = init?.signal;
          return staleResponse.promise;
        }
        return Promise.resolve(
          queryResponse([
            assessmentRow(assessmentCurrentId, 1, "Current rationale"),
          ]),
        );
      }),
    );
    const rendered = render(<AssessmentQueryHarness />);

    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() => expect(callCount).toBe(1));
    rendered.rerender(
      <AssessmentQueryHarness
        queryState={{
          ...emptyWorkbookQueryState(),
          sort: [
            {
              direction: "desc",
              fieldKey: "assessment.assessed_at",
            },
          ],
        }}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() =>
      expect(screen.getByLabelText("assessment-rows").textContent).toBe(
        `${assessmentCurrentId}:Current rationale`,
      ),
    );
    expect(staleSignal?.aborted).toBe(true);

    fireEvent.click(screen.getByRole("button", { name: "patch" }));
    await waitFor(() =>
      expect(screen.getByLabelText("assessment-rows").textContent).toBe(
        `${assessmentCurrentId}:Patched rationale`,
      ),
    );
    staleResponse.resolve(
      queryResponse([
        assessmentRow(assessmentObsoleteId, 1, "Obsolete rationale"),
      ]),
    );
    await Promise.resolve();
    await Promise.resolve();
    expect(screen.getByLabelText("assessment-rows").textContent).toBe(
      `${assessmentCurrentId}:Patched rationale`,
    );
  });

  it("retains accepted rows on a stale error and clears them on access loss", async () => {
    const onAuthorityUncertain = vi.fn();
    let responseKind: "ready" | "error" | "denied" = "ready";
    vi.stubGlobal(
      "fetch",
      vi.fn(() => {
        if (responseKind === "error") {
          return Promise.resolve(errorResponse("query_failed", 500));
        }
        if (responseKind === "denied") {
          return Promise.resolve(errorResponse("authorization_denied", 403));
        }
        return Promise.resolve(
          queryResponse([
            assessmentRow(assessmentCurrentId, 1, "Current rationale"),
          ]),
        );
      }),
    );
    render(
      <AssessmentQueryHarness onAuthorityUncertain={onAuthorityUncertain} />,
    );

    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() =>
      expect(screen.getByLabelText("assessment-load-state").textContent).toBe(
        "ready",
      ),
    );
    responseKind = "error";
    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() =>
      expect(screen.getByLabelText("assessment-load-state").textContent).toBe(
        "stale_error",
      ),
    );
    expect(screen.getByLabelText("assessment-rows").textContent).toContain(
      assessmentCurrentId,
    );

    responseKind = "denied";
    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() =>
      expect(screen.getByLabelText("assessment-load-state").textContent).toBe(
        "permission_denied",
      ),
    );
    expect(onAuthorityUncertain).toHaveBeenCalledOnce();
    expect(screen.getByLabelText("assessment-rows").textContent).toBe("");
  });

  it("does not query while inactive and aborts active work on teardown", async () => {
    const pending = deferred<Response>();
    let signal: AbortSignal | null | undefined;
    const fetchMock = vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
      signal = init?.signal;
      return pending.promise;
    });
    vi.stubGlobal("fetch", fetchMock);
    const rendered = render(<AssessmentQueryHarness active={false} />);

    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    expect(fetchMock).not.toHaveBeenCalled();
    rendered.rerender(<AssessmentQueryHarness />);
    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledOnce());
    rendered.unmount();
    expect(signal?.aborted).toBe(true);
  });
});

describe("Assessment committed query reconciliation", () => {
  it("rejects older full rows after socket or HTTP observations without manufacturing filtered membership", async () => {
    const owner = new WorkbookAssessmentAuthoringOwner(
      incidentId,
      { create: () => "unused" },
      { accepted: () => {}, refresh: async () => {} },
    );
    owner.setAuthority({
      actorId: "actor",
      sessionIdentity: "session",
      incidentId,
      role: "editor",
      closed: false,
    });
    const old = assessmentRow(assessmentCurrentId, 1, "Original");
    const latest = assessmentRow(assessmentCurrentId, 3, "Latest");
    const query = vi.fn().mockResolvedValue({
      kind: "accepted",
      value: {
        incidentId,
        viewSchemaId: assessmentsViewSchemaId,
        rows: [old],
        ...acceptedQueryMetadata(assessmentsViewSchemaId),
      },
    });
    const hook = renderHook(() =>
      useAssessmentSurfaceQuery({
        active: true,
        onAuthorityUncertain: undefined,
        queryState: emptyWorkbookQueryState(),
        viewQuery: { query },
        committedRecords: owner,
      }),
    );
    await act(async () => hook.result.current.refresh());
    act(() => owner.acceptVersion(assessmentCurrentId, 3));
    await act(async () => {
      await expect(
        hook.result.current.refresh({ requireAcceptance: true }),
      ).rejects.toBeTruthy();
    });
    expect(hook.result.current.loadState.kind).toBe("stale_error");
    act(() => owner.acceptRow(latest));
    expect(hook.result.current.rows[0]?.row_version).toBe(3);
    await act(async () => hook.result.current.refresh());
    expect(hook.result.current.rows[0]?.row_version).toBe(3);
    query.mockResolvedValue({
      kind: "accepted",
      value: {
        incidentId,
        viewSchemaId: assessmentsViewSchemaId,
        rows: [],
        ...acceptedQueryMetadata(assessmentsViewSchemaId),
      },
    });
    await act(async () =>
      hook.result.current.refresh({ requireAcceptance: true }),
    );
    expect(hook.result.current.rows).toEqual([]);
    expect(owner.latestRow(assessmentCurrentId)).toEqual(latest);
    hook.unmount();
  });
  it("accepts an authorized startup query independently of mutation authority initialization", async () => {
    const owner = new WorkbookAssessmentAuthoringOwner(
      incidentId,
      { create: () => "unused" },
      { accepted: () => {}, refresh: async () => {} },
    );
    const row = assessmentRow(assessmentCurrentId, 1, "Readable startup row");
    const query = vi.fn().mockResolvedValue({
      kind: "accepted",
      value: {
        incidentId,
        viewSchemaId: assessmentsViewSchemaId,
        rows: [row],
        ...acceptedQueryMetadata(assessmentsViewSchemaId),
      },
    });
    const hook = renderHook(() =>
      useAssessmentSurfaceQuery({
        active: true,
        onAuthorityUncertain: undefined,
        queryState: emptyWorkbookQueryState(),
        viewQuery: { query },
        committedRecords: owner,
      }),
    );
    await act(async () => hook.result.current.refresh());
    expect(hook.result.current.rows).toEqual([row]);
    expect(owner.canSubmit()).toBe(false);
    hook.unmount();
  });
  it("removes deleted rows monotonically and ignores an in-flight query after authority suspension", async () => {
    const owner = new WorkbookAssessmentAuthoringOwner(
      incidentId,
      { create: () => "unused" },
      { accepted: () => {}, refresh: async () => {} },
    );
    const authority = {
      actorId: "actor",
      sessionIdentity: "session",
      incidentId,
      role: "editor",
      closed: false,
    } as const;
    owner.setAuthority(authority);
    const row = assessmentRow(assessmentCurrentId, 1, "Protected");
    const query = vi.fn().mockResolvedValue({
      kind: "accepted",
      value: {
        incidentId,
        viewSchemaId: assessmentsViewSchemaId,
        rows: [row],
        ...acceptedQueryMetadata(assessmentsViewSchemaId),
      },
    });
    const hook = renderHook(() =>
      useAssessmentSurfaceQuery({
        active: true,
        onAuthorityUncertain: undefined,
        queryState: emptyWorkbookQueryState(),
        viewQuery: { query },
        committedRecords: owner,
      }),
    );
    await act(async () => hook.result.current.refresh());
    act(() => owner.acceptVersion(assessmentCurrentId, 2, true));
    expect(hook.result.current.rows).toEqual([]);
    await act(async () => {
      await expect(
        hook.result.current.refresh({ requireAcceptance: true }),
      ).rejects.toBeTruthy();
    });
    let resolve!: (value: unknown) => void;
    query.mockImplementation(
      () =>
        new Promise((done) => {
          resolve = done;
        }),
    );
    let pending!: Promise<void>;
    act(() => {
      pending = hook.result.current.refresh();
    });
    act(() => owner.suspend());
    await act(async () => {
      resolve({
        kind: "accepted",
        value: {
          incidentId,
          viewSchemaId: assessmentsViewSchemaId,
          rows: [assessmentRow(assessmentCurrentId, 3, "Late")],
          ...acceptedQueryMetadata(assessmentsViewSchemaId),
        },
      });
      await pending;
    });
    expect(hook.result.current.rows).toEqual([]);
    expect(owner.latestRow(assessmentCurrentId)).toBeNull();
    act(() => owner.setAuthority(authority));
    expect(hook.result.current.rows).toEqual([]);
    hook.unmount();
  });
});
