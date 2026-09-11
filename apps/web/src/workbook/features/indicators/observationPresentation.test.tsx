import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import {
  observationAuthority,
  observationOwnerFixture,
  observationSource,
  testObservation,
} from "../../../testing/observationTestSupport";
import { IndicatorInspectorWorkflow } from "./IndicatorInspectorWorkflow";
import { ObservationContext } from "./ObservationContext";
import { ObservationDetails } from "./ObservationDetails";
import type { ObservationReadPort } from "./observationOperation";
import { WorkbookObservationRecovery } from "./WorkbookObservationRecovery";

afterEach(cleanup);
it("Observation presentation fences late source reads and hides protected state on access loss", async () => {
  const t = observationOwnerFixture(),
    late = deferred<Awaited<ReturnType<ObservationReadPort["observations"]>>>();
  t.reader.observations.mockReturnValueOnce(late.promise).mockResolvedValue({
    kind: "accepted",
    value: { items: [], hasMore: false, nextCursor: null },
  });
  function Panel({ id }: { id: string }) {
    return (
      <ObservationContext.Provider value={t.owner}>
        <IndicatorInspectorWorkflow
          action="indicator.observations.manage"
          sourceRecordId={id}
          source={{
            ...t.source,
            source: () => ({ ...observationSource, recordId: id, text: id }),
          }}
        />
      </ObservationContext.Provider>
    );
  }
  const view = render(<Panel id={observationSource.recordId} />);
  view.rerender(<Panel id="20000000-0000-4000-8000-000000000002" />);
  await act(async () =>
    late.resolve({
      kind: "accepted",
      value: { items: [testObservation], hasMore: false, nextCursor: null },
    }),
  );
  expect(
    screen.queryByText(testObservation.observed_text, { selector: "strong" }),
  ).toBeNull();
  expect(
    screen.getByRole("textbox", { name: "Saved source text" }),
  ).toHaveProperty("value", "20000000-0000-4000-8000-000000000002");
  act(() => t.owner.suspend());
  expect(screen.queryByRole("textbox")).toBeNull();
  act(() => t.owner.setAuthority(observationAuthority));
  expect(
    screen.getByRole("textbox", { name: "Saved source text" }),
  ).toHaveProperty("value", "20000000-0000-4000-8000-000000000002");
});
it("Observation uncertainty survives close and reopen without completing detached drafts or moving focus", async () => {
  const t = observationOwnerFixture(),
    committed = vi.fn();
  t.transport.send.mockResolvedValueOnce({ kind: "uncertain" });
  const view = render(
    <ObservationContext.Provider value={t.owner}>
      <IndicatorInspectorWorkflow
        action="indicator.observations.manage"
        sourceRecordId={observationSource.recordId}
        source={t.source}
        onMutationCommitted={committed}
      />
    </ObservationContext.Provider>,
  );
  const input = screen.getByRole("textbox", {
    name: "Saved source text",
  }) as HTMLTextAreaElement;
  input.setSelectionRange(2, 11);
  fireEvent.click(screen.getByRole("button", { name: "Use selected text" }));
  fireEvent.click(screen.getByRole("button", { name: "Create observation" }));
  await screen.findByText(/The observation outcome is unknown/);
  const a = t.entry()?.attempt;
  view.unmount();
  render(<WorkbookObservationRecovery owner={t.owner} />);
  const trigger = screen.getByRole("button", {
    name: /Indicator observations/,
  });
  fireEvent.click(trigger);
  const recoveryHeading = screen.getByRole("heading");
  expect(recoveryHeading.textContent).toBe("Indicator observation recovery");
  expect(recoveryHeading).toBe(document.activeElement);
  const replay = screen.getByRole("button", {
    name: "Replay original observation request",
  });
  replay.focus();
  fireEvent.click(replay);
  await waitFor(() => expect(t.entry()?.reconciliation).toBe("complete"));
  expect(t.transport.send.mock.calls.map(([attempt]) => attempt)).toEqual([
    a,
    a,
  ]);
  expect(committed).not.toHaveBeenCalled();
  expect(
    t.owner.drafts.get(
      `capture:${observationSource.recordId}:${observationSource.fieldKey}`,
    )?.selection?.text,
  ).toBe("α.example");
  fireEvent.keyDown(
    screen.getByRole("region", { name: "Indicator observation recovery" }),
    { key: "Escape" },
  );
  expect(document.activeElement).toBe(trigger);
});
it("Observation transition focus stays with its subject and provenance remains opaque text", () => {
  const t = observationOwnerFixture(),
    item = {
      ...testObservation,
      origin_locator: "<img src=x onerror=alert(1)>",
    };
  const props = {
    reader: t.owner,
    generation: 1,
    draft: t.owner.drafts.ensure("one"),
    drafts: t.owner.drafts,
    manage: true,
    disabled: false,
    onSubmit: vi.fn(),
  };
  const view = render(<ObservationDetails {...props} item={item} />);
  const dismiss = screen.getByRole("button", { name: "Dismiss observation" });
  dismiss.focus();
  view.rerender(
    <ObservationDetails
      {...props}
      item={{ ...item, row_version: 2, resolution_status: "dismissed" }}
    />,
  );
  expect(document.activeElement?.textContent).toBe(item.observed_text);
  expect(screen.getByText(item.origin_locator)).toBeTruthy();
  expect(view.container.querySelector("img")).toBeNull();
});
