import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type {
  IndicatorMutationAccepted,
  IndicatorObservation,
  IndicatorWorkflowPort,
} from "../../mutations/workbookMutationCommandPorts";
import { IndicatorInspectorWorkflow } from "./IndicatorInspectorWorkflow";

const observation = {
  observation_id: "30000000-0000-4000-8000-000000000001",
  incident_id: "10000000-0000-4000-8000-000000000001",
  source_record_id: "20000000-0000-4000-8000-000000000001",
  source_field_key: "timeline.raw_activity_text",
  origin_kind: "manual_entry",
  origin_locator: "bytes:0-7",
  observed_text: "1.2.3.4",
  parsed_indicator_type: "ipv4_addr",
  normalized_candidate: "1.2.3.4",
  resolution_status: "unresolved",
  resolved_indicator_record_id: null,
  row_version: 1,
  created_by_user_id: "40000000-0000-4000-8000-000000000001",
  created_at: "2026-08-03T12:00:00Z",
  resolved_by_user_id: null,
  resolved_at: null,
  resolution_method: null,
} as const satisfies IndicatorObservation;

function workflowPort() {
  return {
    appendStateInterval: vi.fn<IndicatorWorkflowPort["appendStateInterval"]>(),
    createManualObservation:
      vi.fn<IndicatorWorkflowPort["createManualObservation"]>(),
    listObservations: vi.fn<IndicatorWorkflowPort["listObservations"]>(),
    listSourceObservations:
      vi.fn<IndicatorWorkflowPort["listSourceObservations"]>(),
    listStateIntervals: vi.fn<IndicatorWorkflowPort["listStateIntervals"]>(),
    transitionObservation:
      vi.fn<IndicatorWorkflowPort["transitionObservation"]>(),
  } satisfies IndicatorWorkflowPort;
}

describe("IndicatorInspectorWorkflow", () => {
  it("presents loading, retry, and empty observation states accessibly", async () => {
    const port = workflowPort();
    port.listObservations
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: { kind: "retryable", message: "Try observations again." },
      })
      .mockResolvedValueOnce({
        kind: "accepted",
        value: { items: [], paging: null },
      });

    render(
      <IndicatorInspectorWorkflow
        beginMutation={() => vi.fn()}
        action="indicator.observations.pivot"
        indicatorRecordId="50000000-0000-4000-8000-000000000001"
        port={port}
        rowVersion={1}
      />,
    );

    expect(screen.getByRole("status").textContent).toContain("Loading…");
    expect((await screen.findByRole("alert")).textContent).toContain(
      "Try observations again.",
    );
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(screen.getByRole("status").textContent).toContain("Loading…");
    expect(
      await screen.findByText(
        "No active observations resolve to this Indicator.",
      ),
    ).toBeTruthy();
    expect(port.listObservations).toHaveBeenCalledTimes(2);
  });

  it("preserves prior observations while loading the next cursor page", async () => {
    const secondObservation = {
      ...observation,
      observation_id: "30000000-0000-4000-8000-000000000002",
      observed_text: "example.test",
    } satisfies IndicatorObservation;
    const port = workflowPort();
    port.listObservations
      .mockResolvedValueOnce({
        kind: "accepted",
        value: {
          items: [observation],
          paging: { has_more: true, limit: 1, next_cursor: "next-page" },
        },
      })
      .mockResolvedValueOnce({
        kind: "accepted",
        value: { items: [secondObservation], paging: null },
      });

    render(
      <IndicatorInspectorWorkflow
        beginMutation={() => vi.fn()}
        action="indicator.observations.pivot"
        indicatorRecordId="50000000-0000-4000-8000-000000000001"
        port={port}
        rowVersion={1}
      />,
    );

    expect(await screen.findByText("1.2.3.4")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Load more" }));
    expect(await screen.findByText("example.test")).toBeTruthy();
    expect(screen.getByText("1.2.3.4")).toBeTruthy();
    expect(port.listObservations).toHaveBeenLastCalledWith({
      cursorToken: "next-page",
      indicatorRecordId: "50000000-0000-4000-8000-000000000001",
    });
  });

  it("forwards accepted mutation metadata to stable-ID refresh coordination", async () => {
    const accepted = {
      affectedRecords: [
        { record_id: observation.source_record_id, row_version: 2 },
      ],
      changeSetId: "60000000-0000-4000-8000-000000000001",
      replayed: false,
      resource: observation,
    } satisfies IndicatorMutationAccepted<IndicatorObservation>;
    const port = workflowPort();
    port.listSourceObservations.mockResolvedValue({
      kind: "accepted",
      value: { items: [], paging: null },
    });
    port.createManualObservation.mockResolvedValue({
      kind: "accepted",
      value: accepted,
    });
    const onMutationCommitted = vi.fn(async () => undefined);

    render(
      <IndicatorInspectorWorkflow
        beginMutation={() => vi.fn()}
        action="indicator.observations.manage"
        onMutationCommitted={onMutationCommitted}
        port={port}
        rowVersion={1}
        sourceFields={[
          {
            fieldKey: "timeline.raw_activity_text",
            label: "Raw activity",
          },
        ]}
        sourceRecordId={observation.source_record_id}
      />,
    );

    await waitFor(() => expect(port.listSourceObservations).toHaveBeenCalled());
    fireEvent.click(screen.getByRole("button", { name: "Create observation" }));
    await waitFor(() =>
      expect(onMutationCommitted).toHaveBeenCalledWith(accepted),
    );
    expect(screen.getByRole("status").textContent).toContain(
      `Observation ${observation.observation_id} created.`,
    );
  });
});
