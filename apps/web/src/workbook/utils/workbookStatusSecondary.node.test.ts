import { describe, expect, it } from "vitest";
import type { SheetRef } from "../../shared/sheetRef";
import {
  selectWorkbookStatusSecondary,
  type WorkbookStatusSecondaryCandidate,
} from "./workbookStatusSecondary";

const activeSheetRef: SheetRef = {
  kind: "view_schema",
  id: "cartulary.view.timeline.v2",
};
function candidate(
  kind: WorkbookStatusSecondaryCandidate["kind"],
  sheetRef: SheetRef = activeSheetRef,
): WorkbookStatusSecondaryCandidate {
  return {
    kind,
    message: kind,
    scope: { kind: "surface", sheetRef },
    count: 1,
    action: null,
  };
}

describe("workbook status secondary selection", () => {
  it("uses the generated exhaustive priority independent of input order", () => {
    const candidates = [
      candidate("queued_or_in_flight"),
      candidate("refresh_paused"),
      candidate("authentication_required"),
      candidate("terminal_replay_failure"),
      candidate("same_field_conflict"),
      candidate("queue_overflow"),
      candidate("client_txn_conflict"),
    ];
    for (let index = 0; index < candidates.length; index += 1) {
      expect(
        selectWorkbookStatusSecondary(
          candidates.slice(0, index + 1),
          activeSheetRef,
        )?.kind,
      ).toBe(candidates[index]?.kind);
      expect(
        selectWorkbookStatusSecondary(
          candidates.slice(0, index + 1).reverse(),
          activeSheetRef,
        )?.kind,
      ).toBe(candidates[index]?.kind);
    }
    const global = {
      ...candidate("client_txn_conflict"),
      scope: { kind: "workbook" as const },
      action: { kind: "transaction_recovery" as const, unitId: "first" },
    };
    expect(
      selectWorkbookStatusSecondary(
        [candidate("same_field_conflict"), global],
        { kind: "saved_view", id: "saved-1" },
      ),
    ).toBe(global);
  });

  it("rejects every inactive-surface candidate", () => {
    const refs: SheetRef[] = [
      activeSheetRef,
      { kind: "saved_view", id: "saved-1" },
      { kind: "saved_view", id: "saved-2" },
    ];
    for (const ref of refs)
      for (const active of refs) {
        const local = candidate("same_field_conflict", ref);
        expect(selectWorkbookStatusSecondary([local], active)).toBe(
          ref === active ? local : null,
        );
      }
    expect(
      selectWorkbookStatusSecondary(
        [
          candidate("client_txn_conflict", {
            kind: "view_schema",
            id: "cartulary.view.entities.v2",
          }),
        ],
        activeSheetRef,
      ),
    ).toBeNull();
  });
});
