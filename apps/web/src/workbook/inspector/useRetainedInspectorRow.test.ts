import { renderHook } from "@testing-library/react";
import { expect, it } from "vitest";
import { timelineRow } from "../../testing/timelineWorkbookTestSupport";
import { entityRowFromApi } from "../models/entityWorkbookModel";
import { acceptWorkbookRowObservation } from "../query/acceptWorkbookRowObservation";
import type {
  WorkbookQueryRow,
  WorkbookReadScope,
} from "../query/WorkbookQueryRow";
import {
  normalizeTimelineFullRow,
  rowFromApi,
} from "../timeline/models/timelineRowModel";
import { useRetainedInspectorRow } from "./useRetainedInspectorRow";

const originalScope: WorkbookReadScope = {
  actorId: "account-one",
  sessionIdentity: "session-one",
  incidentId: "incident",
  epoch: 0,
};
const accepted = (recordId: string, version: number, scope = originalScope) =>
  acceptWorkbookRowObservation(
    { record_id: recordId, row_version: version, cells: {} },
    scope,
  );

it("retains one inspector source across window eviction without treating absence as deletion", () => {
  const row = accepted("original", 4);
  const hook = renderHook(
    (props: {
      rows: readonly WorkbookQueryRow[];
      recordId: string | null;
      scope: WorkbookReadScope | null;
      readable: boolean;
    }) => useRetainedInspectorRow({ ...props, sourceRow: (row) => row }),
    {
      initialProps: {
        rows: [row],
        recordId: "original",
        scope: originalScope,
        readable: true,
      } as {
        rows: readonly WorkbookQueryRow[];
        recordId: string | null;
        scope: WorkbookReadScope | null;
        readable: boolean;
      },
    },
  );
  const update = (
    values: Partial<{
      rows: readonly WorkbookQueryRow[];
      recordId: string | null;
      scope: WorkbookReadScope | null;
      readable: boolean;
    }>,
  ) =>
    hook.rerender({
      rows: [],
      recordId: "original",
      scope: originalScope,
      readable: true,
      ...values,
    });
  update({});
  expect(hook.result.current).toBe(row);
  update({ rows: [accepted("original", 2), accepted("wrong-record", 999)] });
  expect(hook.result.current).toBe(row);
  update({ recordId: "different", rows: [row] });
  expect(hook.result.current).toBeNull();
  update({ rows: [row] });
  // Interaction permissions and presentation resets do not change source authority.
  update({ scope: { ...originalScope } });
  expect(hook.result.current).toBe(row);
  const replacement = {
    ...originalScope,
    actorId: "account-two",
    sessionIdentity: "session-two",
  };
  update({ scope: replacement, rows: [{ ...row }, structuredClone(row)] });
  expect(hook.result.current).toBeNull();
  const fresh = accepted("original", 5, replacement);
  update({ scope: replacement, rows: [accepted("original", 99), fresh] });
  expect(hook.result.current).toBe(fresh);
  const revoked = { ...replacement, epoch: replacement.epoch + 1 };
  update({ scope: revoked, readable: false, rows: [fresh] });
  expect(hook.result.current).toBeNull();
  update({ scope: revoked, rows: [structuredClone(fresh)] });
  expect(hook.result.current).toBeNull();
  // Same-account reauthentication keeps independent drafts/receipts, but requires new read evidence.
  const resumed = {
    ...replacement,
    sessionIdentity: "session-three",
    epoch: 2,
  };
  update({ scope: resumed, rows: [fresh] });
  expect(hook.result.current).toBeNull();
  const resumedRow = accepted("original", 6, resumed);
  update({ scope: resumed, rows: [resumedRow, { ...row, row_version: 999 }] });
  expect(hook.result.current).toBe(resumedRow);
  update({
    scope: { ...resumed, incidentId: "other-incident" },
    rows: [resumedRow],
  });
  expect(hook.result.current).toBeNull();
  update({ scope: null, readable: false });
  expect(hook.result.current).toBeNull();
  hook.unmount();
});

it("preserves accepted provenance through cloning and source conversions without relabeling authority", () => {
  const raw = acceptWorkbookRowObservation(
    timelineRow({
      recordId: "record",
      rowVersion: 4,
      captureState: "rough",
      summary: "Accepted",
    }),
    originalScope,
  );
  const normalized = normalizeTimelineFullRow(
    structuredClone(raw),
    "test source conversion",
  );
  expect(rowFromApi(normalized).rawRow?.observation).toEqual(raw.observation);
  const host = accepted("host", 7);
  expect(
    entityRowFromApi(structuredClone(host), "host").rawRow.observation,
  ).toEqual(host.observation);
  const reconstructed = normalizeTimelineFullRow(
    { ...raw, row_version: 99 },
    "changed source version",
  );
  expect(reconstructed.observation).toBeUndefined();
});
