import { describe, expect, it } from "vitest";
import { decideWorkbookRecordFreshness } from "./timelineRowsModel";

describe("timelineRowsModel", () => {
  it("classifies stale row-version updates only when durable identity is comparable", () => {
    expect(
      decideWorkbookRecordFreshness({ recordId: "row-1", rowVersion: 2 }, 3),
    ).toEqual({ comparable: true, stale: true });
    expect(
      decideWorkbookRecordFreshness({ recordId: "row-1", rowVersion: 3 }, 3),
    ).toEqual({ comparable: true, stale: false });
    expect(
      decideWorkbookRecordFreshness({ recordId: null, rowVersion: 1 }, 3),
    ).toEqual({ comparable: false, stale: false });
    expect(
      decideWorkbookRecordFreshness({ recordId: "row-1", rowVersion: null }, 3),
    ).toEqual({ comparable: false, stale: false });
    expect(
      decideWorkbookRecordFreshness(
        { recordId: "row-1", rowVersion: 1 },
        undefined,
      ),
    ).toEqual({ comparable: false, stale: false });
  });
});
