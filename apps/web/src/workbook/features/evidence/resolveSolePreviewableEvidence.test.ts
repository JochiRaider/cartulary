import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type {
  EvidenceCapabilityPort,
  EvidenceHandleOutcome,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookReadScope } from "../../query/WorkbookQueryRow";
import { createTimelineInspectorElementRegistry } from "../../timeline/focus/timelineInspectorElementRegistry";
import { useTimelineLinkedEvidenceReview } from "../../timeline/hooks/useTimelineLinkedEvidenceReview";
import type { WorkbookRow } from "../../timeline/models/timelineRowModel";
import { resolveSolePreviewableEvidence } from "./resolveSolePreviewableEvidence";

const accepted: EvidenceHandleOutcome = {
  kind: "accepted",
  value: {
    href: "/opaque/unused",
    filename: "notes.txt",
    previewKind: "text_inline",
  },
};
function blocked(reason: string): EvidenceHandleOutcome {
  return {
    kind: "rejected",
    failure: {
      kind: "terminal",
      message: "File cannot be previewed.",
      publicCode: "evidence_access_unavailable",
      publicReason: reason as WorkbookOperationFailure["publicReason"],
    },
  };
}
const unavailable: EvidenceHandleOutcome = {
  kind: "rejected",
  failure: { kind: "terminal", message: "Request failed." },
};
function port(
  outcomes: Record<string, EvidenceHandleOutcome>,
): EvidenceCapabilityPort {
  return {
    issueHandle: vi.fn(
      async ({ evidenceRecordId }) => outcomes[evidenceRecordId] ?? unavailable,
    ),
  };
}
const readScope: WorkbookReadScope = {
  actorId: "actor",
  sessionIdentity: "session",
  incidentId: "incident",
  epoch: 1,
};
function timelineRow(ids: readonly string[], version = 1): WorkbookRow {
  const values = {
    dateEnteredText: "",
    analystText: "",
    mitreStageText: "",
    deviceObjectText: "",
    ipAddressText: "",
    activityUTCText: "",
    activityLocalText: "",
    rawActivityText: "",
    activitySynopsisText: "",
    dataSourceText: "",
  };
  return {
    key: "source",
    recordId: "source",
    rowVersion: version,
    viewSchemaId: "cartulary.view.timeline.v2",
    captureState: "rough",
    values,
    committedValues: values,
    collectionValues: { hostRefs: [], identityRefs: [], tags: [] },
    collectionDrafts: { hostRefs: "", identityRefs: "", tags: "" },
    pendingSignature: null,
    rawRow: {
      view_schema_id: "cartulary.view.timeline.v2",
      record_id: "source",
      row_version: version,
      observation: {
        recordId: "source",
        rowVersion: version,
        scope: readScope,
      },
      cells: {
        "timeline.attached_evidence_ids": {
          value: {
            items: ids.map((id, index) => ({
              item_ref: `record_ref:link-${index}`,
              item_kind: "record_ref",
              display_text: `Title ${index}`,
              linked_record_id: id,
            })),
          },
        },
      },
    },
  };
}
function review(row: WorkbookRow, evidence: EvidenceCapabilityPort) {
  const rowsRef: { current: readonly WorkbookRow[] } = { current: [row] };
  const registry = createTimelineInspectorElementRegistry({
    lifecycleKey: "incident",
    reviewGeneration: 1,
    subject: null,
  });
  const props = {
    canRead: true,
    inspectorOpen: true,
    selectedRowId: "source",
    scopeKey: "scope-1",
  };
  const rendered = renderHook(
    (input: typeof props) =>
      useTimelineLinkedEvidenceReview({
        ...input,
        port: evidence,
        rowsRef,
        readScope,
        elementRegistry: registry,
        onAccessFailure: vi.fn(),
        onRestoreGridFocus: vi.fn(),
      }),
    { initialProps: props },
  );
  return { ...rendered, rowsRef, props };
}
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.useRealTimers();
});

describe("Timeline linked Evidence preview resolution", () => {
  it("uses one deadline for the complete sequential discovery", async () => {
    vi.useFakeTimers();
    let finishFirst!: (value: EvidenceHandleOutcome) => void;
    const first = new Promise<EvidenceHandleOutcome>((resolve) => {
      finishFirst = resolve;
    });
    const unresolved = new Promise<EvidenceHandleOutcome>(() => undefined);
    const source: EvidenceCapabilityPort = {
      issueHandle: vi
        .fn()
        .mockReturnValueOnce(first)
        .mockReturnValueOnce(unresolved),
    };
    const resolving = resolveSolePreviewableEvidence(
      ["one", "two", "three"],
      source,
      new AbortController().signal,
      () => true,
    );
    expect(source.issueHandle).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(20_000);
    finishFirst(accepted);
    await vi.advanceTimersByTimeAsync(0);
    expect(source.issueHandle).toHaveBeenCalledTimes(2);
    await vi.advanceTimersByTimeAsync(10_000);
    expect(await resolving).toEqual({ kind: "indeterminate" });
    expect(source.issueHandle).toHaveBeenCalledTimes(2);
  });

  it("distinguishes zero, one, multiple, and mixed blocked and previewable links", async () => {
    const signal = new AbortController().signal;
    const current = () => true;
    expect(
      await resolveSolePreviewableEvidence([], port({}), signal, current),
    ).toEqual({ kind: "none" });
    const single = port({ one: accepted });
    expect(
      await resolveSolePreviewableEvidence(["one"], single, signal, current),
    ).toEqual({ kind: "sole", recordId: "one" });
    expect(single.issueHandle).toHaveBeenCalledWith({
      evidenceRecordId: "one",
      kind: "preview",
      signal: expect.any(AbortSignal),
    });
    const multiple = port({ one: accepted, two: accepted, three: accepted });
    expect(
      await resolveSolePreviewableEvidence(
        ["one", "two", "three"],
        multiple,
        signal,
        current,
      ),
    ).toEqual({ kind: "multiple" });
    expect(multiple.issueHandle).toHaveBeenCalledTimes(2);
    const mixed = port({
      one: blocked("unsupported_preview"),
      two: accepted,
      three: blocked("evidence_quarantined"),
    });
    expect(
      await resolveSolePreviewableEvidence(
        ["one", "two", "three"],
        mixed,
        signal,
        current,
      ),
    ).toEqual({ kind: "sole", recordId: "two" });
    expect(mixed.issueHandle).toHaveBeenCalledTimes(3);
  });

  it("does not infer a sole item from failed or incomplete observation", async () => {
    const signal = new AbortController().signal;
    const current = () => true;
    for (const failure of [
      unavailable,
      blocked("unrecognized_reason"),
      {
        kind: "rejected",
        failure: { kind: "authorization_lost", message: "Private" },
      } as EvidenceHandleOutcome,
    ]) {
      const source = port({ one: accepted, two: failure, three: accepted });
      expect(
        await resolveSolePreviewableEvidence(
          ["one", "two", "three"],
          source,
          signal,
          current,
        ),
      ).toEqual({ kind: "indeterminate" });
      expect(source.issueHandle).toHaveBeenCalledTimes(2);
    }
    const controller = new AbortController();
    const request = vi.fn(async () => {
      controller.abort();
      return accepted;
    });
    expect(
      await resolveSolePreviewableEvidence(
        ["one", "two"],
        { issueHandle: request },
        controller.signal,
        current,
      ),
    ).toEqual({ kind: "indeterminate" });
    expect(request).toHaveBeenCalledTimes(1);
  });
});

describe("Timeline linked Evidence review intent", () => {
  it("ends a stalled multi-link Space check without inferring a sole preview", async () => {
    vi.useFakeTimers();
    let settle!: (value: EvidenceHandleOutcome) => void;
    const pending = new Promise<EvidenceHandleOutcome>((resolve) => {
      settle = resolve;
    });
    const source: EvidenceCapabilityPort = {
      issueHandle: vi
        .fn()
        .mockResolvedValueOnce(accepted)
        .mockReturnValueOnce(pending),
    };
    const row = timelineRow(["one", "two", "three"]);
    const view = review(row, source);
    await act(async () => view.result.current.startSpace(row));
    expect(view.result.current.spaceStatus).toBe("checking");
    await act(async () => vi.advanceTimersByTimeAsync(30_000));
    expect(view.result.current.spaceStatus).toBe("indeterminate");
    expect(view.result.current.access.overlay).toBeNull();
    await act(async () => settle(accepted));
    expect(view.result.current.access.overlay).toBeNull();
    view.unmount();
  });

  it("uses the returned target ID and a fresh handle after bounded Space probes", async () => {
    const row = timelineRow(["blocked", "target", "blocked-again"]);
    const source = port({
      blocked: blocked("unsupported_preview"),
      target: accepted,
      "blocked-again": blocked("evidence_quarantined"),
    });
    const view = review(row, source);
    await act(async () => view.result.current.startSpace(row));
    expect(
      vi
        .mocked(source.issueHandle)
        .mock.calls.map(([input]) => input.evidenceRecordId),
    ).toEqual(["blocked", "target", "blocked-again", "target"]);
    expect(view.result.current.access.overlay).not.toBeNull();
    expect(view.result.current.spaceStatus).toBe("idle");
  });

  it("leaves the list in place when capability observation is indeterminate", async () => {
    const row = timelineRow(["target", "failed"]);
    const source = port({ target: accepted, failed: unavailable });
    const view = review(row, source);
    await act(async () => view.result.current.startSpace(row));
    expect(view.result.current.spaceStatus).toBe("indeterminate");
    expect(view.result.current.access.overlay).toBeNull();
    expect(source.issueHandle).toHaveBeenCalledTimes(2);
    view.rowsRef.current = [timelineRow(["target", "failed"], 2)];
    view.rerender({ ...view.props });
    expect(view.result.current.spaceStatus).toBe("idle");
  });

  it("discards a late probe after a row observation or authority change", async () => {
    for (const change of ["row", "links", "authority"] as const) {
      let settle!: (value: EvidenceHandleOutcome) => void;
      const pending = new Promise<EvidenceHandleOutcome>((resolve) => {
        settle = resolve;
      });
      const source: EvidenceCapabilityPort = {
        issueHandle: vi.fn(() => pending),
      };
      const row = timelineRow(["one", "two"]);
      const view = review(row, source);
      await act(async () => view.result.current.startSpace(row));
      if (change === "row")
        view.rowsRef.current = [timelineRow(["one", "two"], 2)];
      if (change === "links") view.rowsRef.current = [timelineRow(["one"])];
      view.rerender(
        change === "authority"
          ? { ...view.props, canRead: false, scopeKey: "scope-2" }
          : { ...view.props },
      );
      await act(async () => settle(accepted));
      expect(view.result.current.access.overlay).toBeNull();
      expect(source.issueHandle).toHaveBeenCalledTimes(1);
      view.unmount();
    }
  });
});
