import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  captureTimelineFileSource,
  resolveTimelineFileTarget,
} from "./timelineEvidenceAttachmentPlan";
import {
  createDraftRow,
  normalizeTimelineFullRow,
  rowFromApi,
} from "./timelineRowModel";

const timeline = requireViewContract(timelineViewSchemaId);
const saved = (id: string) =>
  rowFromApi(
    normalizeTimelineFullRow(
      fullWorkbookViewRow(timeline, id, 3, {}),
      "file gesture target",
    ),
  );
const unchanged = (key: string) => key;

describe("Timeline file gesture targets", () => {
  it("captures only the explicit row and current version without selection inputs", () => {
    const a = saved("record-a"),
      b = saved("record-b");
    const result = resolveTimelineFileTarget(
      { kind: "record", recordId: "record-b" },
      [a, b],
      unchanged,
    );
    expect(result).toEqual({
      kind: "resolved",
      source: captureTimelineFileSource(b),
    });
    expect(
      resolveTimelineFileTarget({ kind: "row", key: b.key }, [a, b], unchanged),
    ).toEqual(result);
    expect(result.kind === "resolved" && result.source).not.toBe(b);
  });
  it("retains an exact draft and follows only its accepted lifecycle alias", () => {
    const original = createDraftRow(1),
      replacement = createDraftRow(2),
      accepted = saved("accepted-original");
    expect(
      resolveTimelineFileTarget(
        { kind: "row", key: original.key },
        [original, replacement],
        unchanged,
      ),
    ).toEqual({
      kind: "resolved",
      source: captureTimelineFileSource(original),
    });
    expect(
      resolveTimelineFileTarget(
        { kind: "row", key: original.key },
        [replacement],
        unchanged,
      ),
    ).toEqual({ kind: "unavailable" });
    expect(
      resolveTimelineFileTarget(
        { kind: "row", key: original.key },
        [accepted, replacement],
        (key) => (key === original.key ? accepted.key : key),
      ),
    ).toEqual({
      kind: "resolved",
      source: captureTimelineFileSource(accepted),
    });
  });
  it("fails closed for missing ambiguous and unavailable targets without substituting a draft", () => {
    const row = saved("record-a"),
      draft = createDraftRow(1);
    expect(resolveTimelineFileTarget(null, [row, draft], unchanged)).toEqual({
      kind: "missing",
    });
    expect(
      resolveTimelineFileTarget(
        { kind: "record", recordId: "gone" },
        [row, draft],
        unchanged,
      ),
    ).toEqual({ kind: "unavailable" });
    expect(
      resolveTimelineFileTarget(
        { kind: "record", recordId: "record-a" },
        [row, row, draft],
        unchanged,
      ),
    ).toEqual({ kind: "ambiguous" });
    for (const invalid of [
      { ...row, captureState: "superseded" },
      { ...row, rowVersion: null },
      { ...row, viewSchemaId: "another-view" },
    ])
      expect(
        resolveTimelineFileTarget(
          { kind: "record", recordId: "record-a" },
          [invalid, draft],
          unchanged,
        ),
      ).toEqual({ kind: "unavailable" });
  });
  it("leaves pending creation and source write coordination to the existing owners", () => {
    const draft = { ...createDraftRow(1), pendingSignature: "ordinary-create" };
    expect(
      resolveTimelineFileTarget(
        { kind: "row", key: draft.key },
        [draft],
        unchanged,
      ),
    ).toEqual({ kind: "resolved", source: captureTimelineFileSource(draft) });
  });
});
