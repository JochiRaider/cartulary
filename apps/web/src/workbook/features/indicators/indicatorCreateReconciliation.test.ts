import { expect, it, vi } from "vitest";
import { canonicalCreateFixture } from "../../../testing/indicatorCreateTestSupport";
import {
  observationAuthority,
  observationTargetId,
} from "../../../testing/observationTestSupport";
import { WorkbookRecordHistoryOwner } from "../../history/WorkbookRecordHistoryOwner";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import { reconcileIndicatorCreateReceipt } from "./reconcileIndicatorCreateReceipt";

it("Canonical refresh reads unfiltered pages and fences historical receipts against newer history", async () => {
  const t = canonicalCreateFixture(),
    history = new WorkbookRecordHistoryOwner(
      observationAuthority.incidentId,
      t.ids,
    );
  history.setAuthority(observationAuthority);
  const load = vi.spyOn(history, "load").mockResolvedValue({
    kind: "accepted",
    value: {
      record_id: observationTargetId,
      incident_id: observationAuthority.incidentId,
      row_version: 9,
      deleted: false,
      items: [],
      paging: { has_more: false, next_cursor: null, limit: 100 },
    },
  });
  t.reader.records
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { items: [], hasMore: true, nextCursor: "outside-visible-page" },
    })
    .mockResolvedValueOnce({
      kind: "accepted",
      value: {
        items: [{ ...t.receipt.row, row_version: 9 }],
        hasMore: false,
        nextCursor: null,
      },
    });
  const scope = { signal: new AbortController().signal, isCurrent: () => true };
  await reconcileIndicatorCreateReceipt(
    t.owner,
    t.reader,
    history,
    observationAuthority.incidentId,
    t.receipt,
    scope,
    vi.fn(),
  );
  expect(load).toHaveBeenCalledWith(observationTargetId, scope.signal);
  expect(
    t.reader.records.mock.calls.map(([, query, cursor]) => [query, cursor]),
  ).toEqual([
    [emptyWorkbookQueryState(), null],
    [emptyWorkbookQueryState(), "outside-visible-page"],
  ]);
  expect(t.owner.latestRow(observationTargetId)?.row_version).toBe(9);
  t.reader.records.mockResolvedValueOnce({
    kind: "accepted",
    value: { items: [t.receipt.row], hasMore: false, nextCursor: null },
  });
  await expect(
    reconcileIndicatorCreateReceipt(
      t.owner,
      t.reader,
      history,
      observationAuthority.incidentId,
      t.receipt,
      scope,
      vi.fn(),
    ),
  ).rejects.toThrow();
  expect(t.owner.latestRow(observationTargetId)?.row_version).toBe(9);
  expect(t.transport.send).not.toHaveBeenCalled();
});
it("Canonical refresh retains receipt on unavailable target access loss and detached history", async () => {
  const t = canonicalCreateFixture(),
    history = new WorkbookRecordHistoryOwner(
      observationAuthority.incidentId,
      t.ids,
    ),
    loseAccess = vi.fn();
  history.setAuthority(observationAuthority);
  const load = vi.spyOn(history, "load").mockResolvedValue({
      kind: "rejected",
      failure: { kind: "authorization_lost", message: "Membership changed" },
    }),
    scope = { signal: new AbortController().signal, isCurrent: () => true };
  t.owner.registerReconciliation(async (_attempt, receipt, scope) =>
    reconcileIndicatorCreateReceipt(
      t.owner,
      t.reader,
      history,
      observationAuthority.incidentId,
      receipt,
      scope,
      loseAccess,
    ),
  );
  await t.owner.execute(t.admit());
  expect(loseAccess).toHaveBeenCalledOnce();
  expect(t.entry()?.receipt).toEqual(t.receipt);
  expect(t.entry()?.refresh).toBe("required");
  load.mockImplementationOnce(async () => {
    t.owner.retire();
    return {
      kind: "rejected",
      failure: { kind: "authorization_lost", message: "Old account" },
    };
  });
  await expect(
    reconcileIndicatorCreateReceipt(
      t.owner,
      t.reader,
      history,
      observationAuthority.incidentId,
      t.receipt,
      scope,
      loseAccess,
    ),
  ).rejects.toThrow();
  expect(loseAccess).toHaveBeenCalledOnce();
  expect(t.transport.send).toHaveBeenCalledOnce();
});
