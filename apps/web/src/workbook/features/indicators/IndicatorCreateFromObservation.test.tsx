import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, expect, it } from "vitest";
import { canonicalCreateFixture } from "../../../testing/indicatorCreateTestSupport";
import {
  observationOwnerFixture,
  observationSource,
  observationTargetId,
  testObservation,
  testObservationReceipt,
} from "../../../testing/observationTestSupport";
import { IndicatorCreateContext } from "./IndicatorCreateContext";
import { IndicatorInspectorWorkflow } from "./IndicatorInspectorWorkflow";
import { ObservationContext } from "./ObservationContext";

afterEach(cleanup);
function workflow(
  c: ReturnType<typeof canonicalCreateFixture>,
  o: ReturnType<typeof observationOwnerFixture>,
  pivot = false,
) {
  return (
    <IndicatorCreateContext.Provider value={c.owner}>
      <ObservationContext.Provider value={o.owner}>
        {pivot ? (
          <IndicatorInspectorWorkflow
            action="indicator.observations.pivot"
            indicatorRecordId={observationTargetId}
          />
        ) : (
          <IndicatorInspectorWorkflow
            action="indicator.observations.manage"
            sourceRecordId={observationSource.recordId}
            source={o.source}
          />
        )}
      </ObservationContext.Provider>
    </IndicatorCreateContext.Provider>
  );
}
async function propose() {
  fireEvent.click(
    await screen.findByRole("button", { name: "Create canonical Indicator…" }),
  );
  fireEvent.change(screen.getByRole("combobox", { name: "Value kind" }), {
    target: { value: "atomic" },
  });
  fireEvent.change(screen.getByRole("textbox", { name: "Canonical value" }), {
    target: { value: "EXAMPLE[.]COM" },
  });
  fireEvent.click(
    screen.getByRole("button", {
      name: "Create canonical Indicator",
    }),
  );
}
it("Canonical workflow requires validated acceptance and explicit resolution with independent replay", async () => {
  const c = canonicalCreateFixture(),
    o = observationOwnerFixture();
  c.transport.send.mockResolvedValueOnce({ kind: "uncertain" });
  o.transport.send
    .mockResolvedValueOnce({ kind: "uncertain" })
    .mockResolvedValueOnce({
      kind: "acknowledged",
      receipt: {
        ...testObservationReceipt,
        replayed: true,
        observation: {
          ...testObservation,
          row_version: 2,
          resolution_status: "resolved",
          resolved_indicator_record_id: observationTargetId,
          resolved_by_user_id: testObservation.created_by_user_id,
          resolved_at: "2026-09-12T01:00:00Z",
          resolution_method: "manual",
        },
        affected_records: [
          ...testObservationReceipt.affected_records,
          { record_id: observationTargetId, row_version: 2 },
        ],
      },
    });
  render(workflow(c, o));
  await propose();
  const replay = await screen.findByRole("button", {
    name: "Replay original canonical create",
  });
  expect(o.transport.send).not.toHaveBeenCalled();
  expect(
    screen.queryByRole("button", {
      name: "Resolve observation to this Indicator",
    }),
  ).toBeNull();
  fireEvent.click(replay);
  const resolve = await screen.findByRole("button", {
    name: "Resolve observation to this Indicator",
  });
  expect(screen.getByText("Indicator available: example.com")).toBeTruthy();
  expect(screen.getByText("Not linked to this Indicator.")).toBeTruthy();
  expect(o.transport.send).not.toHaveBeenCalled();
  fireEvent.click(resolve);
  await screen.findByText(
    "Link outcome unknown. Recover the observation operation separately.",
  );
  fireEvent.click(
    await screen.findByRole("button", {
      name: "Replay original observation request",
    }),
  );
  await screen.findByText("Observation linked to this Indicator.");
  expect(c.transport.send).toHaveBeenCalledTimes(2);
  expect(o.transport.send).toHaveBeenCalledTimes(2);
  expect(o.transport.send.mock.calls[0]?.[0].body).toBe(
    o.transport.send.mock.calls[1]?.[0].body,
  );
  expect(
    JSON.parse(o.transport.send.mock.calls[0]?.[0].body ?? "{}")
      .base_row_version,
  ).toBe(1);
  expect(c.entry()?.resolutionAttemptIds).toEqual([o.entry()?.attempt.id]);
  expect(testObservation.observed_text).toBe("α.example");
});
it("Canonical workflow retains a late result across closure without resolution and keeps the pivot read only", async () => {
  const c = canonicalCreateFixture(),
    o = observationOwnerFixture();
  let finish!: (value: Awaited<ReturnType<typeof c.transport.send>>) => void;
  c.transport.send.mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        finish = resolve;
      }),
  );
  const view = render(workflow(c, o));
  await propose();
  await waitFor(() => expect(c.transport.send).toHaveBeenCalledTimes(1));
  view.unmount();
  await act(async () => finish({ kind: "accepted", receipt: c.receipt }));
  const pivot = render(workflow(c, o, true));
  await screen.findByText("α.example", { selector: "strong" });
  expect(screen.queryByText("Indicator available: example.com")).toBeNull();
  expect(
    screen.queryByRole("button", { name: /Create canonical Indicator/ }),
  ).toBeNull();
  pivot.unmount();
  render(workflow(c, o));
  await screen.findByText("Indicator available: example.com");
  expect(o.transport.send).not.toHaveBeenCalled();
  expect(c.transport.send).toHaveBeenCalledTimes(1);
});
