import { requireViewContract } from "@cartulary/view-contracts";
import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import {
  lifecycleAuthority,
  lifecycleIndicator,
  lifecycleReceipt,
  lifecycleRow,
} from "../../../testing/indicatorLifecycleTestSupport";
import { acceptedQueryMetadata } from "../../../testing/workbookQueryTestSupport";
import { createIndicatorLifecycleAdapter } from "../../adapters/createIndicatorLifecycleAdapter";
import { WorkbookRecordHistoryOwner } from "../../history/WorkbookRecordHistoryOwner";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import { useGenericSurfaceQuery } from "../../query/useGenericSurfaceQuery";
import type { WorkbookViewQueryPort } from "../../query/WorkbookViewQueryPort";
import { indicatorLifecycleViewId } from "./indicatorLifecycleModel";
import type { IndicatorLifecycleTransportPort } from "./indicatorLifecycleOperation";
import { reconcileIndicatorLifecycleReceipt } from "./reconcileIndicatorLifecycleReceipt";
import { WorkbookIndicatorLifecycleOwner } from "./WorkbookIndicatorLifecycleOwner";

function owner() {
  const o = new WorkbookIndicatorLifecycleOwner(
    lifecycleAuthority.incidentId,
    { create: () => "original" },
    {
      canReserve: () => true,
      coordinate: async () => true,
      accepted: () => {},
    },
  );
  o.setAuthority(lifecycleAuthority);
  return o;
}
afterEach(cleanup);
it("Lifecycle reconciliation refreshes every affected identity beyond the first query page and fences stale history", async () => {
  const o = owner(),
    records = vi.fn<IndicatorLifecycleTransportPort["records"]>();
  o.configure({
    ...createIndicatorLifecycleAdapter({
      apiBase: undefined,
      incidentId: lifecycleAuthority.incidentId,
    }),
    records,
  });
  const history = new WorkbookRecordHistoryOwner(
    lifecycleAuthority.incidentId,
    { create: () => "history" },
  );
  history.setAuthority(lifecycleAuthority);
  const receipt = lifecycleReceipt(),
    second = "00000000-0000-4000-8000-000000000099";
  receipt.affected_records.push({ record_id: second, row_version: 3 });
  const load = vi.spyOn(history, "load").mockImplementation(async (id) => ({
    kind: "accepted",
    value: {
      record_id: id,
      incident_id: lifecycleAuthority.incidentId,
      row_version: id === second ? 3 : 2,
      deleted: false,
      representation_generation: "cartulary.history.1",
      items: [],
      paging: { has_more: false, next_cursor: null, limit: 100 },
    },
  }));
  const refresh = vi.fn(async () => {});
  history.registerRecordPresentation(lifecycleIndicator, refresh);
  history.registerRecordPresentation(second, refresh);
  records
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { items: [lifecycleRow(2)], hasMore: true, nextCursor: "opaque" },
    })
    .mockResolvedValueOnce({
      kind: "accepted",
      value: {
        items: [{ ...lifecycleRow(3), record_id: second }],
        hasMore: false,
        nextCursor: null,
      },
    });
  let current = true;
  const scope = {
    signal: new AbortController().signal,
    isCurrent: () => current,
  };
  await reconcileIndicatorLifecycleReceipt(o, history, receipt, scope);
  expect(load.mock.calls.map(([id]) => id)).toEqual([
    lifecycleIndicator,
    second,
  ]);
  expect(
    records.mock.calls.map(([, query, cursor]) => ({ query, cursor })),
  ).toEqual([
    { query: emptyWorkbookQueryState(), cursor: null },
    { query: emptyWorkbookQueryState(), cursor: "opaque" },
  ]);
  expect(refresh).toHaveBeenCalledTimes(2);
  load.mockImplementationOnce(async (id) => {
    current = false;
    return {
      kind: "accepted",
      value: {
        record_id: id,
        incident_id: lifecycleAuthority.incidentId,
        row_version: 20,
        deleted: false,
        representation_generation: "cartulary.history.1",
        items: [],
        paging: { has_more: false, next_cursor: null, limit: 100 },
      },
    };
  });
  await expect(
    reconcileIndicatorLifecycleReceipt(o, history, receipt, scope),
  ).rejects.toThrow();
  expect(o.latestVersion(lifecycleIndicator)).toBe(2);
  expect(refresh).toHaveBeenCalledTimes(2);
});
it("Lifecycle query materialization never regresses HTTP or socket version evidence", async () => {
  const o = owner(),
    query = vi.fn<WorkbookViewQueryPort["query"]>(async () => ({
      kind: "accepted",
      value: {
        incidentId: lifecycleAuthority.incidentId,
        viewSchemaId: indicatorLifecycleViewId,
        ...acceptedQueryMetadata(indicatorLifecycleViewId),
        rows: [lifecycleRow(8)],
      },
    }));
  const hook = renderHook(() =>
    useGenericSurfaceQuery({
      active: true,
      contract: requireViewContract(indicatorLifecycleViewId),
      onAuthorityUncertain: undefined,
      queryState: emptyWorkbookQueryState(),
      viewQuery: { query },
      viewSchemaId: indicatorLifecycleViewId,
      indicatorOwner: o,
    }),
  );
  await act(async () => hook.result.current.refresh());
  expect(hook.result.current.rows[0]?.row_version).toBe(8);
  act(() => o.acceptVersion(lifecycleIndicator, 10));
  query.mockResolvedValue({
    kind: "accepted",
    value: {
      incidentId: lifecycleAuthority.incidentId,
      viewSchemaId: indicatorLifecycleViewId,
      ...acceptedQueryMetadata(indicatorLifecycleViewId),
      rows: [lifecycleRow(9)],
    },
  });
  await act(async () => {
    await expect(
      hook.result.current.refresh({ requireAcceptance: true }),
    ).rejects.toThrow();
  });
  expect(hook.result.current.rows[0]?.row_version).toBe(8);
  expect(hook.result.current.loadState.kind).toBe("stale_error");
  act(() => o.acceptRow(lifecycleRow(10)));
  expect(hook.result.current.rows[0]?.row_version).toBe(10);
  act(() => o.suspend());
  expect(hook.result.current.rows).toEqual([]);
});
