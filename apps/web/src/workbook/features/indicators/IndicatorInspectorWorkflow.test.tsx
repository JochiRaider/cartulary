import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  observationOwnerFixture,
  observationSource,
  observationTargetId,
  testObservation,
  testObservationReceipt,
} from "../../../testing/observationTestSupport";
import { IndicatorInspectorWorkflow } from "./IndicatorInspectorWorkflow";
import { ObservationContext } from "./ObservationContext";

afterEach(cleanup);
describe("IndicatorInspectorWorkflow", () => {
  it("presents loading, retry, and empty observation states accessibly", async () => {
    const t = observationOwnerFixture();
    t.reader.observations
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: { kind: "retryable", message: "Try observations again." },
      })
      .mockResolvedValueOnce({
        kind: "accepted",
        value: { items: [], hasMore: false, nextCursor: null },
      });
    render(
      <ObservationContext.Provider value={t.owner}>
        <IndicatorInspectorWorkflow
          action="indicator.observations.pivot"
          indicatorRecordId={observationTargetId}
        />
      </ObservationContext.Provider>,
    );
    expect(screen.getByRole("status").textContent).toContain("Loading…");
    expect(await screen.findByText("Try observations again.")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Retry observations" }));
    expect(await screen.findByText("No observations.")).toBeTruthy();
    expect(t.reader.observations).toHaveBeenCalledTimes(2);
  });
  it("preserves prior observations while loading the next cursor page", async () => {
    const t = observationOwnerFixture(),
      second = {
        ...testObservation,
        observation_id: "30000000-0000-4000-8000-000000000000",
        observed_text: "example.test",
      };
    t.reader.observations
      .mockResolvedValueOnce({
        kind: "accepted",
        value: {
          items: [testObservation],
          hasMore: true,
          nextCursor: "next-page",
        },
      })
      .mockResolvedValueOnce({
        kind: "accepted",
        value: { items: [second], hasMore: false, nextCursor: null },
      });
    render(
      <ObservationContext.Provider value={t.owner}>
        <IndicatorInspectorWorkflow
          action="indicator.observations.pivot"
          indicatorRecordId={observationTargetId}
        />
      </ObservationContext.Provider>,
    );
    expect(
      await screen.findByText("α.example", { selector: "strong" }),
    ).toBeTruthy();
    fireEvent.click(
      screen.getByRole("button", { name: "Load more observations" }),
    );
    expect(
      await screen.findByText("example.test", { selector: "strong" }),
    ).toBeTruthy();
    expect(screen.getByText("α.example", { selector: "strong" })).toBeTruthy();
    expect(t.reader.observations).toHaveBeenLastCalledWith(
      { kind: "indicator", recordId: observationTargetId },
      "next-page",
      expect.any(AbortSignal),
    );
  });
  it("forwards accepted mutation metadata to stable-ID refresh coordination", async () => {
    const t = observationOwnerFixture(),
      onMutationCommitted = vi.fn(async () => {});
    render(
      <ObservationContext.Provider value={t.owner}>
        <IndicatorInspectorWorkflow
          action="indicator.observations.manage"
          sourceRecordId={observationSource.recordId}
          source={t.source}
          onMutationCommitted={onMutationCommitted}
        />
      </ObservationContext.Provider>,
    );
    const textarea = screen.getByRole("textbox", {
      name: "Saved source text",
    }) as HTMLTextAreaElement;
    textarea.setSelectionRange(2, 11);
    fireEvent.click(screen.getByRole("button", { name: "Use selected text" }));
    fireEvent.click(screen.getByRole("button", { name: "Create observation" }));
    await waitFor(() => expect(onMutationCommitted).toHaveBeenCalledTimes(1));
    expect(t.accepted).toHaveBeenCalledWith(
      testObservationReceipt,
      "secure-observation-1",
    );
    expect(t.projections).toHaveBeenCalledWith(
      expect.objectContaining({ id: "secure-observation-1" }),
      testObservationReceipt,
      expect.anything(),
    );
    expect(t.entry()?.reconciliation).toBe("complete");
  });
});
