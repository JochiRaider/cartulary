import { requireViewContract } from "@cartulary/view-contracts";
import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import {
  observationAuthority,
  observationOwnerFixture,
  observationSource,
  observationTargetId,
  observationTargetRow,
  testObservationReceipt,
} from "../../../testing/observationTestSupport";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { acceptedQueryMetadata } from "../../../testing/workbookQueryTestSupport";
import { WorkbookRecordHistoryOwner } from "../../history/WorkbookRecordHistoryOwner";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import { useGenericSurfaceQuery } from "../../query/useGenericSurfaceQuery";
import type { WorkbookViewQueryPort } from "../../query/WorkbookViewQueryPort";
import { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { observationIndicatorView } from "./observationModel";
import { reconcileObservationReceipt } from "./reconcileObservationReceipt";

afterEach(cleanup);
it("Observation reconciliation reads source previous and new Indicator identities beyond visible pages", async () => {
  const t = observationOwnerFixture(),
    history = new WorkbookRecordHistoryOwner(
      observationAuthority.incidentId,
      t.ids,
    ),
    a = t.admit();
  history.setAuthority(observationAuthority);
  const old = "50000000-0000-4000-8000-000000000002",
    sourceRow = {
      record_id: observationSource.recordId,
      row_version: 5,
      cells: {
        [observationSource.fieldKey]: { value: observationSource.text },
      },
    };
  const receipt = {
    ...testObservationReceipt,
    affected_records: [
      ...testObservationReceipt.affected_records,
      { record_id: observationTargetId, row_version: 91 },
      { record_id: old, row_version: 3 },
    ],
  };
  const load = vi.spyOn(history, "load").mockImplementation(async (id) => ({
    kind: "accepted",
    value: {
      record_id: id,
      incident_id: observationAuthority.incidentId,
      row_version:
        receipt.affected_records.find((row) => row.record_id === id)
          ?.row_version ?? 0,
      deleted: false,
      representation_generation: "cartulary.history.1",
      items: [],
      paging: { has_more: false, next_cursor: null, limit: 100 },
    },
  }));
  t.reader.records.mockImplementation(async (view, _query, cursor) => ({
    kind: "accepted",
    value:
      view === observationSource.viewSchemaId
        ? { items: [sourceRow], hasMore: false, nextCursor: null }
        : cursor === null
          ? {
              items: [{ ...observationTargetRow, row_version: 91 }],
              hasMore: true,
              nextCursor: "next",
            }
          : {
              items: [
                { ...observationTargetRow, record_id: old, row_version: 3 },
              ],
              hasMore: false,
              nextCursor: null,
            },
  }));
  let current = true;
  const scope = {
    signal: new AbortController().signal,
    isCurrent: () => current,
  };
  await reconcileObservationReceipt(t.owner, history, a, receipt, scope);
  expect(load.mock.calls.map(([id]) => id)).toEqual(
    receipt.affected_records.map((row) => row.record_id),
  );
  expect(
    t.reader.records.mock.calls.map(([view, query, cursor]) => [
      view,
      query,
      cursor,
    ]),
  ).toEqual([
    [observationSource.viewSchemaId, emptyWorkbookQueryState(), null],
    [observationIndicatorView, emptyWorkbookQueryState(), null],
    [observationIndicatorView, emptyWorkbookQueryState(), "next"],
  ]);
  expect(t.owner.latestRow(old)?.row_version).toBe(3);
  expect(
    t.owner.latestRow(observationSource.recordId)?.cells[
      observationSource.fieldKey
    ]?.value,
  ).toBe(observationSource.text);
  load.mockImplementationOnce(async (id) => {
    current = false;
    return {
      kind: "accepted",
      value: {
        record_id: id,
        incident_id: observationAuthority.incidentId,
        row_version: 50,
        deleted: false,
        representation_generation: "cartulary.history.1",
        items: [],
        paging: { has_more: false, next_cursor: null, limit: 100 },
      },
    };
  });
  await expect(
    reconcileObservationReceipt(t.owner, history, a, receipt, scope),
  ).rejects.toThrow();
  expect(t.owner.latestVersion(observationSource.recordId)).toBe(5);
});
it("Observation query materialization fences older HTTP projections against accepted socket evidence", async () => {
  const t = observationOwnerFixture(),
    contract = requireViewContract(observationIndicatorView);
  const row = (version: number) =>
    fullWorkbookViewRow(contract, observationTargetId, version, {
      "indicator.indicator_type": "domain_name",
      "indicator.display_value": "target.example",
    });
  const query = vi.fn<WorkbookViewQueryPort["query"]>(async () => ({
    kind: "accepted",
    value: {
      incidentId: observationAuthority.incidentId,
      viewSchemaId: observationIndicatorView,
      ...acceptedQueryMetadata(observationIndicatorView),
      rows: [row(8)],
    },
  }));
  const hook = renderHook(() =>
    useGenericSurfaceQuery({
      active: true,
      contract,
      indicatorOwner: t.owner,
      onAuthorityUncertain: undefined,
      queryState: emptyWorkbookQueryState(),
      viewQuery: { query },
      viewSchemaId: observationIndicatorView,
    }),
  );
  await act(async () => hook.result.current.refresh());
  expect(hook.result.current.rows[0]?.row_version).toBe(8);
  act(() => t.owner.acceptVersion(observationTargetId, 10));
  query.mockResolvedValue({
    kind: "accepted",
    value: {
      incidentId: observationAuthority.incidentId,
      viewSchemaId: observationIndicatorView,
      ...acceptedQueryMetadata(observationIndicatorView),
      rows: [row(9)],
    },
  });
  await act(async () => {
    await expect(
      hook.result.current.refresh({ requireAcceptance: true }),
    ).rejects.toThrow();
  });
  expect(hook.result.current.rows[0]?.row_version).toBe(8);
  act(() => t.owner.acceptRow(row(10)));
  expect(hook.result.current.rows[0]?.row_version).toBe(10);
  act(() => t.owner.suspend());
  expect(hook.result.current.rows).toEqual([]);
  const runtime = new WorkbookMutationRuntime(
    { incidentId: observationAuthority.incidentId, clientInstanceId: "tab" },
    t.ids,
    { execute: vi.fn() },
  );
  const initial = renderHook(() =>
    useGenericSurfaceQuery({
      active: true,
      contract,
      indicatorOwner: runtime.indicatorRecords,
      onAuthorityUncertain: undefined,
      queryState: emptyWorkbookQueryState(),
      viewQuery: { query },
      viewSchemaId: observationIndicatorView,
    }),
  );
  await act(async () => initial.result.current.refresh());
  expect(initial.result.current.rows).toHaveLength(1);
  act(() => runtime.indicatorObservations.setAuthority(observationAuthority));
  expect(initial.result.current.rows).toHaveLength(1);
  act(() => runtime.indicatorLifecycle.setAuthority(observationAuthority));
  expect(initial.result.current.rows).toHaveLength(1);
  act(() => runtime.indicatorCreate.setAuthority(observationAuthority));
  expect(initial.result.current.rows).toHaveLength(1);
  act(() => runtime.indicatorObservations.suspend());
  expect(initial.result.current.rows).toHaveLength(0);
  initial.unmount();
  runtime.indicatorObservations.retire();
  runtime.indicatorLifecycle.retire();
  runtime.indicatorCreate.retire();
});
