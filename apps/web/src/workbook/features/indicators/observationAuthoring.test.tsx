import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { useReducer } from "react";
import { afterEach, expect, it, vi } from "vitest";
import {
  observationReaderFixture,
  observationSource,
  observationTargetRow,
  testObservation,
} from "../../../testing/observationTestSupport";
import { ObservationCaptureEditor } from "./ObservationCaptureEditor";
import { ObservationDetails } from "./ObservationDetails";
import { ObservationDraftStore } from "./ObservationDraftStore";
import { ObservationTargetPicker } from "./ObservationTargetPicker";

afterEach(cleanup);

it("Observation authoring previews the exact selection and omits unchosen type and target", async () => {
  const submit = vi.fn(),
    reader = observationReaderFixture();
  let store: ObservationDraftStore;
  function Editor({
    ready = true,
    version = 4,
  }: {
    ready?: boolean;
    version?: number;
  }) {
    const [, redraw] = useReducer((value) => value + 1, 0);
    store ??= new ObservationDraftStore(redraw);
    return (
      <ObservationCaptureEditor
        source={{ ...observationSource, rowVersion: version }}
        ready={ready}
        draft={store.ensure("source")}
        drafts={store}
        reader={reader}
        generation={0}
        disabled={false}
        onSubmit={submit}
      />
    );
  }
  const view = render(<Editor />);
  const text = screen.getByRole("textbox", {
    name: "Saved source text",
  }) as HTMLTextAreaElement;
  text.setSelectionRange(12, 21);
  fireEvent.click(screen.getByRole("button", { name: "Use selected text" }));
  expect(screen.getByLabelText("Selected text preview").textContent).toContain(
    "α.example",
  );
  fireEvent.click(screen.getByRole("button", { name: "Create observation" }));
  expect(submit.mock.calls[0]?.[0]).toEqual({
    action: "create",
    source: observationSource,
    selection: { startByte: 14, endByte: 24, text: "α.example" },
  });
  view.rerender(<Editor ready={false} />);
  expect(
    (
      screen.getByRole("button", {
        name: "Create observation",
      }) as HTMLButtonElement
    ).disabled,
  ).toBe(true);
  view.rerender(<Editor version={5} />);
  expect(
    (
      screen.getByRole("button", {
        name: "Create observation",
      }) as HTMLButtonElement
    ).disabled,
  ).toBe(true);
  expect(screen.getByLabelText("Selected text preview").textContent).toContain(
    "Select text from the current saved source.",
  );
});

it("Observation targets page independently and retain the selected identity across filters", async () => {
  const reader = observationReaderFixture();
  reader.records
    .mockResolvedValueOnce({
      kind: "accepted",
      value: {
        items: [
          {
            ...observationTargetRow,
            record_id: "50000000-0000-4000-8000-000000000002",
            cells: {
              "indicator.display_value": { value: "earlier.example" },
              "indicator.indicator_type": { value: "domain_name" },
            },
          },
        ],
        nextCursor: "page-two",
        hasMore: true,
      },
    })
    .mockResolvedValueOnce({
      kind: "accepted",
      value: {
        items: [observationTargetRow],
        nextCursor: null,
        hasMore: false,
      },
    });
  const change = vi.fn();
  render(
    <ObservationTargetPicker
      reader={reader}
      generation={0}
      selected={null}
      onChange={change}
    />,
  );
  fireEvent.click(
    await screen.findByRole("button", { name: "Load more Indicators" }),
  );
  await screen.findByRole("option", { name: "target.example · domain_name" });
  fireEvent.change(screen.getByLabelText("Existing Indicator"), {
    target: { value: observationTargetRow.record_id },
  });
  expect(change).toHaveBeenCalledWith({
    recordId: observationTargetRow.record_id,
    label: "target.example",
    type: "domain_name",
  });
  expect(reader.records.mock.calls[1]?.[2]).toBe("page-two");
  fireEvent.change(screen.getByLabelText("Indicator type"), {
    target: { value: "domain_name" },
  });
  await waitFor(() =>
    expect(reader.records.mock.calls.at(-1)?.[1].filters).toEqual([
      {
        fieldKey: "indicator.indicator_type",
        op: "eq",
        arg: { value: "domain_name" },
      },
    ]),
  );
});

it("Observation transition editors isolate targets and ordinary restore has no prior target", async () => {
  const reader = observationReaderFixture(),
    submit = vi.fn();
  let store: ObservationDraftStore;
  function Details() {
    const [, redraw] = useReducer((value) => value + 1, 0);
    store ??= new ObservationDraftStore(redraw);
    return (
      <>
        <ObservationDetails
          item={testObservation}
          reader={reader}
          generation={0}
          draft={store.ensure("one")}
          drafts={store}
          manage
          disabled={false}
          onSubmit={submit}
        />
        <ObservationDetails
          item={{
            ...testObservation,
            observation_id: "30000000-0000-4000-8000-000000000002",
            resolution_status: "dismissed",
          }}
          reader={reader}
          generation={0}
          draft={store.ensure("two")}
          drafts={store}
          manage
          disabled={false}
          onSubmit={submit}
        />
      </>
    );
  }
  render(<Details />);
  fireEvent.click(screen.getByRole("button", { name: "Resolve" }));
  await screen.findByRole("option", { name: "target.example · domain_name" });
  fireEvent.change(screen.getByLabelText("Existing Indicator"), {
    target: { value: observationTargetRow.record_id },
  });
  fireEvent.click(screen.getByRole("button", { name: "Resolve observation" }));
  expect(submit.mock.calls[0]?.[0]).toEqual({
    action: "resolve",
    observation: testObservation,
    targetId: observationTargetRow.record_id,
  });
  fireEvent.click(screen.getByRole("button", { name: "Restore observation" }));
  expect(submit.mock.calls[1]?.[0]).toMatchObject({
    action: "restore",
    observation: { observation_id: "30000000-0000-4000-8000-000000000002" },
  });
  expect(submit.mock.calls[1]?.[0]).not.toHaveProperty("targetId");
  expect(submit.mock.calls[1]?.[1].target).toBeNull();
});
