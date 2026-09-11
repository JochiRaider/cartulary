import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import {
  lifecycleAuthority,
  lifecycleIndicator,
  lifecycleInterval,
} from "../../../testing/indicatorLifecycleTestSupport";
import { IndicatorLifecycleContext } from "./IndicatorLifecycleContext";
import { IndicatorLifecycleDraftStore } from "./IndicatorLifecycleDraftStore";
import { IndicatorLifecycleWorkflow } from "./IndicatorLifecycleWorkflow";
import { emptyLifecycleValues } from "./indicatorLifecycleModel";
import type {
  IndicatorLifecycleOwnerPort,
  LifecycleSnapshot,
} from "./indicatorLifecycleOperation";

afterEach(cleanup);
function presentationOwner() {
  let snapshot: LifecycleSnapshot = {
    authority: lifecycleAuthority,
    generation: 1,
    revision: 1,
    entries: [],
  };
  const listeners = new Set<() => void>();
  const owner: IndicatorLifecycleOwnerPort = {
    drafts: new IndicatorLifecycleDraftStore(() => {
      snapshot = { ...snapshot, revision: snapshot.revision + 1 };
      for (const listener of listeners) listener();
    }),
    getSnapshot: () => snapshot,
    subscribe: (listener) => {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
    canSubmit: () => true,
    canReplay: () => true,
    blocksRecord: () => false,
    latestVersion: () => 1,
    latestRow: () => null,
    acceptRow: () => null,
    intervals: vi.fn<IndicatorLifecycleOwnerPort["intervals"]>(async () => ({
      kind: "accepted",
      value: { items: [lifecycleInterval()], hasMore: false, nextCursor: null },
    })),
    records: vi.fn<IndicatorLifecycleOwnerPort["records"]>(async () => ({
      kind: "accepted",
      value: { items: [], hasMore: false, nextCursor: null },
    })),
    admit: vi.fn(() => null),
    execute: async () => {},
    replay: async () => {},
    refresh: async () => {},
    review: async () => true,
    dismiss: () => {},
  };
  return owner;
}
const subject = {
  kind: "live" as const,
  recordId: lifecycleIndicator,
  rowVersion: 1,
  viewSchemaId: "cartulary.view.indicators.v1",
  label: "example.test",
  surfaceLabel: "Indicators",
};
it("Indicator lifecycle shows complete interval details and field-local UTC authoring errors", async () => {
  const owner = presentationOwner();
  render(
    <IndicatorLifecycleContext.Provider value={owner}>
      <IndicatorLifecycleWorkflow
        action="indicator.lifecycle.manage"
        subject={subject}
      />
    </IndicatorLifecycleContext.Provider>,
  );
  expect(await screen.findByText("Observed context")).toBeTruthy();
  expect(screen.getByText("Analyst supplied name")).toBeTruthy();
  expect(screen.getByText("Recorded by (server)")).toBeTruthy();
  expect(screen.getByText("No end specified")).toBeTruthy();
  expect(screen.getByText("0")).toBeTruthy();
  fireEvent.click(
    screen.getByRole("button", { name: "Append lifecycle interval" }),
  );
  expect(screen.getByRole("alert").textContent).toContain(
    "valid effective start in UTC",
  );
  expect(owner.admit).not.toHaveBeenCalled();
  fireEvent.change(screen.getByLabelText("Effective from (UTC)"), {
    target: { value: "2026-09-11T12:00" },
  });
  fireEvent.change(screen.getByLabelText("Confidence (optional)"), {
    target: { value: "0" },
  });
  fireEvent.click(
    screen.getByRole("button", { name: "Append lifecycle interval" }),
  );
  expect(owner.admit).toHaveBeenCalledTimes(1);
});

it("Indicator lifecycle support selection reaches a second query page and retains subject drafts", async () => {
  const owner = presentationOwner();
  const secondId = "00000000-0000-4000-8000-000000000012";
  vi.mocked(owner.records)
    .mockResolvedValueOnce({
      kind: "accepted",
      value: {
        items: [
          {
            record_id: "00000000-0000-4000-8000-000000000011",
            row_version: 1,
            cells: {
              "timeline.activity_synopsis_text": {
                value: "First supporting event",
              },
            },
          },
        ],
        hasMore: true,
        nextCursor: "next-record-page",
      },
    })
    .mockResolvedValueOnce({
      kind: "accepted",
      value: {
        items: [
          {
            record_id: secondId,
            row_version: 1,
            cells: {
              "timeline.activity_synopsis_text": {
                value: "Outside the grid page",
              },
            },
          },
        ],
        hasMore: false,
        nextCursor: null,
      },
    });
  const view = render(
    <IndicatorLifecycleContext.Provider value={owner}>
      <IndicatorLifecycleWorkflow
        action="indicator.lifecycle.manage"
        subject={subject}
      />
    </IndicatorLifecycleContext.Provider>,
  );
  fireEvent.click(
    await screen.findByRole("button", { name: "Load more supporting records" }),
  );
  const checkbox = await screen.findByRole("checkbox", {
    name: /Outside the grid page/u,
  });
  fireEvent.click(checkbox);
  expect(
    owner.drafts
      .get(lifecycleIndicator)
      ?.values.support.map((item) => item.recordId),
  ).toEqual([secondId]);
  fireEvent.change(screen.getByLabelText("Rationale (optional)"), {
    target: { value: "retained words" },
  });
  view.rerender(
    <IndicatorLifecycleContext.Provider value={owner}>
      <IndicatorLifecycleWorkflow
        action="indicator.lifecycle.manage"
        subject={{ ...subject, recordId: secondId }}
      />
    </IndicatorLifecycleContext.Provider>,
  );
  await waitFor(() =>
    expect(owner.drafts.get(secondId)?.values).toEqual(emptyLifecycleValues),
  );
  view.rerender(
    <IndicatorLifecycleContext.Provider value={owner}>
      <IndicatorLifecycleWorkflow
        action="indicator.lifecycle.manage"
        subject={subject}
      />
    </IndicatorLifecycleContext.Provider>,
  );
  expect(
    (screen.getByLabelText("Rationale (optional)") as HTMLTextAreaElement)
      .value,
  ).toBe("retained words");
});
