import { describe, expect, it } from "vitest";
import {
  selectWorkbookStatusSecondary,
  type WorkbookStatusSecondaryCandidate,
} from "./workbookStatusSecondary";

const activeSurfaceId = "cartulary.view.timeline.v2";

function candidate(
  kind: WorkbookStatusSecondaryCandidate["kind"],
  surfaceId = activeSurfaceId,
): WorkbookStatusSecondaryCandidate {
  return { kind, message: kind, surfaceId };
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
          activeSurfaceId,
        )?.kind,
      ).toBe(candidates[index]?.kind);
    }
  });

  it("rejects every inactive-surface candidate", () => {
    expect(
      selectWorkbookStatusSecondary(
        [candidate("client_txn_conflict", "cartulary.view.entities.v2")],
        activeSurfaceId,
      ),
    ).toBeNull();
  });
});
