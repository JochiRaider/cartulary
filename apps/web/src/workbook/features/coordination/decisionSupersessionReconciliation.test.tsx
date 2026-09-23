import { requireViewContract } from "@cartulary/view-contracts";
import { act, cleanup } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import {
  decisionAuthority,
  decisionReceipt,
  decisionRecordChanged,
  decisionReplacementId,
  decisionRow,
  decisionTargetId,
} from "../../../testing/decisionSupersessionTestSupport";
import {
  acceptedQueryMetadata,
  renderHookWithWorkbookQueryBrowsing as renderHook,
} from "../../../testing/workbookQueryTestSupport";
import { createWorkbookDecisionSupersessionAdapter } from "../../adapters/createWorkbookDecisionSupersessionAdapter";
import { WorkbookRecordHistoryOwner } from "../../history/WorkbookRecordHistoryOwner";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import { useGenericSurfaceQuery } from "../../query/useGenericSurfaceQuery";
import type { WorkbookViewQueryPort } from "../../query/WorkbookViewQueryPort";
import { decisionViewId } from "./decisionSupersessionModel";
import type { DecisionSupersessionTransportPort } from "./decisionSupersessionOperation";
import { reconcileDecisionReceipt } from "./reconcileDecisionReceipt";
import { WorkbookDecisionSupersessionOwner } from "./WorkbookDecisionSupersessionOwner";

function owner() {
  const value = new WorkbookDecisionSupersessionOwner(
    decisionAuthority.incidentId,
    { create: () => "txn" },
    { canReserve: () => true, coordinate: async () => true },
  );
  value.setAuthority(decisionAuthority);
  return value;
}
afterEach(cleanup);
it("Decision reconciliation reads both pages and histories while missing or stale rows retain debt", async () => {
  const value = owner();
  const page = vi.fn<DecisionSupersessionTransportPort["page"]>();
  value.configure({
    ...createWorkbookDecisionSupersessionAdapter({
      apiBase: undefined,
      incidentId: decisionAuthority.incidentId,
    }),
    page,
  });
  const history = new WorkbookRecordHistoryOwner(decisionAuthority.incidentId, {
    create: () => "txn",
  });
  history.setAuthority(decisionAuthority);
  const load = vi.spyOn(history, "load").mockImplementation(async (id) => ({
    kind: "accepted",
    value: {
      incident_id: decisionAuthority.incidentId,
      record_id: id,
      row_version: id === decisionTargetId ? 5 : 7,
      deleted: false,
      representation_generation: "cartulary.history.1",
      items: [],
      paging: { has_more: false, next_cursor: null, limit: 100 },
    },
  }));
  const refresh = vi.fn(async () => {});
  history.registerRecordPresentation(decisionTargetId, refresh);
  const scope = { signal: new AbortController().signal, isCurrent: () => true };
  page
    .mockResolvedValueOnce({
      kind: "accepted",
      value: {
        rows: [decisionRow(decisionTargetId, "superseded", 5)],
        hasMore: true,
        nextCursor: "second",
      },
    })
    .mockResolvedValueOnce({
      kind: "accepted",
      value: {
        rows: [decisionRow(decisionReplacementId, "approved", 7)],
        hasMore: false,
        nextCursor: null,
      },
    });
  await reconcileDecisionReceipt(value, history, decisionReceipt(), scope);
  expect(page.mock.calls.map(([cursor]) => cursor)).toEqual([null, "second"]);
  expect(load.mock.calls.map(([id]) => id)).toEqual([
    decisionTargetId,
    decisionReplacementId,
  ]);
  expect(refresh).toHaveBeenCalledTimes(1);
  value.acceptVersion(decisionReplacementId, 9);
  page.mockResolvedValue({
    kind: "accepted",
    value: {
      rows: [
        decisionRow(decisionTargetId, "superseded", 5),
        decisionRow(decisionReplacementId, "approved", 7),
      ],
      hasMore: false,
      nextCursor: null,
    },
  });
  await expect(
    reconcileDecisionReceipt(value, history, decisionReceipt(), scope),
  ).rejects.toThrow("Both affected Decisions");
  page.mockResolvedValue({
    kind: "accepted",
    value: { rows: [], hasMore: true, nextCursor: "loop" },
  });
  await expect(
    reconcileDecisionReceipt(value, history, decisionReceipt(), scope),
  ).rejects.toThrow("Both affected Decisions");
});
it("Decision query and live patches cannot regress accepted high-water versions", async () => {
  const value = owner(),
    row = decisionRow(decisionTargetId, "executed", 8);
  const query = vi.fn<WorkbookViewQueryPort["query"]>(async () => ({
    kind: "accepted",
    value: {
      incidentId: decisionAuthority.incidentId,
      viewSchemaId: decisionViewId,
      ...acceptedQueryMetadata(decisionViewId),
      rows: [row],
    },
  }));
  const hook = renderHook(() =>
    useGenericSurfaceQuery({
      active: true,
      contract: requireViewContract(decisionViewId),
      onAuthorityUncertain: undefined,
      queryState: emptyWorkbookQueryState(),
      viewQuery: { query },
      viewSchemaId: decisionViewId,
      decisionOwner: value,
    }),
  );
  await act(async () => hook.result.current.refresh());
  expect(hook.result.current.rows[0]?.row_version).toBe(8);
  value.acceptVersion(decisionTargetId, 10);
  query.mockResolvedValue({
    kind: "accepted",
    value: {
      incidentId: decisionAuthority.incidentId,
      viewSchemaId: decisionViewId,
      ...acceptedQueryMetadata(decisionViewId),
      rows: [decisionRow(decisionTargetId, "proposed", 5)],
    },
  });
  await act(async () => {
    await expect(
      hook.result.current.refresh({ requireAcceptance: true }),
    ).rejects.toThrow();
  });
  expect(hook.result.current.rows[0]?.row_version).toBe(8);
  expect(hook.result.current.loadState.kind).toBe("stale_error");
  const payload = decisionRecordChanged();
  act(() => {
    expect(hook.result.current.applyRecordChanged(payload)).toEqual({
      kind: "stale",
    });
  });
  expect(hook.result.current.rows[0]?.cells["decision.status"]?.value).toBe(
    "executed",
  );
  query.mockResolvedValue({
    kind: "accepted",
    value: {
      incidentId: decisionAuthority.incidentId,
      viewSchemaId: decisionViewId,
      ...acceptedQueryMetadata(decisionViewId),
      rows: [decisionRow(decisionTargetId, "executed", 10)],
    },
  });
  await act(async () =>
    hook.result.current.refresh({ requireAcceptance: true }),
  );
  expect(hook.result.current.rows[0]?.row_version).toBe(10);
});
