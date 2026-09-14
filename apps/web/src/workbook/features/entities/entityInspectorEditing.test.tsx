import {
  genericEditFieldSelectTestId,
  genericEditRecordSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { useMemo, useState } from "react";
import { afterEach, expect, it, vi } from "vitest";
import { taskAuthority } from "../../../testing/taskWorkbookTestSupport";
import type { RecordPatchTransport } from "../../adapters/workbookRecordPatchTransport";
import { entityRowFromApi } from "../../models/entityWorkbookModel";
import type {
  RecordRouteCommandPort,
  TimelineRelatedRecordPort,
} from "../../mutations/workbookMutationCommandPorts";
import { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { useEntityWorkbookInspectorComposition } from "./useEntityWorkbookInspectorComposition";

const schema = "cartulary.view.hosts.v1";
const contract = requireViewContract(schema);
const firstId = "00000000-0000-4000-8000-000000000601";
const secondId = "00000000-0000-4000-8000-000000000602";
function row(id: string, version = 1) {
  return entityRowFromApi(
    {
      record_id: id,
      row_version: version,
      cells: {
        ...Object.fromEntries(
          contract.fields.map((field) => [field.fieldKey, { value: null }]),
        ),
        "host.display_name": { value: id === firstId ? "Host A" : "Host B" },
        "host.location": { value: "Saved location" },
        "host.host_state": { value: "canonical" },
      },
    },
    "host",
  );
}
function fixture() {
  let sequence = 0;
  const runtime = new WorkbookMutationRuntime(
    {
      incidentId: taskAuthority.incidentId,
      clientInstanceId: "inspector-test",
    },
    { create: () => `inspector-${++sequence}` },
    { execute: vi.fn() },
  );
  runtime.explicitPatches.setAuthority(taskAuthority);
  const patch = vi.fn<RecordPatchTransport["send"]>(async (input) => ({
    kind: "acknowledged",
    receipt: {
      changeSetId: "change",
      viewSchemaId: schema,
      row: { ...row(input.recordId, 2).rawRow, view_schema_id: schema },
    },
  }));
  runtime.explicitPatches.configure({ send: patch });
  const query = {
    query: vi.fn(async () => ({
      kind: "accepted" as const,
      value: {
        incidentId: taskAuthority.incidentId,
        rows: [],
        viewSchemaId: "cartulary.view.timeline.v2",
      },
    })),
  };
  const select = vi.fn();
  const refresh = vi.fn(async () => {});
  runtime.registerSurface(schema, refresh);
  function Surface({
    version = 1,
    selected = firstId,
  }: {
    version?: number;
    selected?: string;
  }) {
    const [error, setError] =
      useState<
        Parameters<
          typeof useEntityWorkbookInspectorComposition
        >[0]["mutationError"]
      >(null);
    const [feedback, setFeedback] =
      useState<
        Parameters<
          typeof useEntityWorkbookInspectorComposition
        >[0]["entityActionFeedback"]
      >(null);
    const rows = useMemo(
      () => [row(firstId, version), row(secondId)],
      [version],
    );
    const inspector = useEntityWorkbookInspectorComposition({
      sheetRef: { kind: "view_schema", id: schema },
      canMerge: false,
      contract,
      currentIncidentRole: "editor",
      currentUserId: taskAuthority.actorId,
      entityActionFeedback: feedback,
      entityIndex: {},
      entityType: "host",
      incidentClosed: false,
      inspectorResetKey: "hosts",
      interactionMode: { kind: "editable" },
      mutationError: error,
      mutationRuntime: runtime,
      onClearSurfaceSelection: vi.fn(),
      onRefreshEntities: refresh,
      onRestoreFocus: vi.fn(),
      recordMutationCommands: {} as RecordRouteCommandPort,
      relatedMutationCommands: {} as TimelineRelatedRecordPort,
      rows,
      selectedEntity: rows.find((item) => item.recordId === selected) ?? null,
      setEntityActionFeedback: setFeedback,
      setMutationError: setError,
      setSelectedRecordId: select,
      viewQuery: query,
    });
    return (
      <>
        <button type="button" onClick={inspector.open}>
          Inspect selected
        </button>
        {inspector.node}
      </>
    );
  }
  return { Surface, patch, runtime, select, refresh };
}
afterEach(cleanup);
function chooseField() {
  fireEvent.click(screen.getByRole("button", { name: "Inspect selected" }));
  const legacy = screen.queryByTestId(genericEditRecordSelectTestId(schema));
  if (legacy) fireEvent.change(legacy, { target: { value: firstId } });
  fireEvent.change(screen.getByTestId(genericEditFieldSelectTestId(schema)), {
    target: { value: "host.location" },
  });
}
it("binds ordinary Entity editing exclusively to the inspected record", async () => {
  const f = fixture();
  render(<f.Surface />);
  chooseField();
  const legacy = screen.queryByTestId(genericEditRecordSelectTestId(schema));
  if (legacy) fireEvent.change(legacy, { target: { value: secondId } });
  fireEvent.change(screen.getByTestId(genericEditValueTestId(schema)), {
    target: { value: "Authored" },
  });
  fireEvent.click(screen.getByTestId(genericEditSubmitTestId(schema)));
  await waitFor(() =>
    expect(f.patch).toHaveBeenCalledWith(
      expect.objectContaining({ recordId: firstId }),
      expect.any(AbortSignal),
    ),
  );
});
it("preserves unfinished Entity text through same-record query replacement", async () => {
  const f = fixture();
  const view = render(<f.Surface />);
  chooseField();
  fireEvent.change(screen.getByTestId(genericEditValueTestId(schema)), {
    target: { value: "Unfinished location" },
  });
  view.rerender(<f.Surface version={2} />);
  expect(
    (screen.getByTestId(genericEditValueTestId(schema)) as HTMLInputElement)
      .value,
  ).toBe("Unfinished location");
});
it("retains Entity field authoring for explicit return to its original field", async () => {
  const f = fixture();
  render(<f.Surface />);
  chooseField();
  fireEvent.change(screen.getByTestId(genericEditValueTestId(schema)), {
    target: { value: "Unfinished location" },
  });
  fireEvent.change(screen.getByTestId(genericEditFieldSelectTestId(schema)), {
    target: { value: "host.display_name" },
  });
  fireEvent.change(screen.getByTestId(genericEditFieldSelectTestId(schema)), {
    target: { value: "host.location" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Resume draft" }));
  expect(
    (screen.getByTestId(genericEditValueTestId(schema)) as HTMLInputElement)
      .value,
  ).toBe("Unfinished location");
});
it("does not restore old Entity selection after a detached completion", async () => {
  const f = fixture();
  let complete: () => void = () => {};
  f.refresh.mockImplementationOnce(
    () =>
      new Promise<void>((resolve) => {
        complete = resolve;
      }),
  );
  const view = render(<f.Surface />);
  chooseField();
  fireEvent.change(screen.getByTestId(genericEditValueTestId(schema)), {
    target: { value: "Authored" },
  });
  fireEvent.click(screen.getByTestId(genericEditSubmitTestId(schema)));
  await waitFor(() => expect(f.refresh).toHaveBeenCalled());
  view.rerender(<f.Surface selected={secondId} />);
  await act(async () => complete());
  await waitFor(() => expect(f.select).not.toHaveBeenCalled());
});
it("invalidates Entity field errors when the canonical subject changes", async () => {
  const f = fixture();
  f.patch.mockResolvedValueOnce({
    kind: "rejected",
    failure: { kind: "stale_target", message: "Original Host needs review" },
  });
  const view = render(<f.Surface />);
  chooseField();
  fireEvent.change(screen.getByTestId(genericEditValueTestId(schema)), {
    target: { value: "Authored" },
  });
  fireEvent.click(screen.getByTestId(genericEditSubmitTestId(schema)));
  await screen.findByText("Original Host needs review");
  view.rerender(<f.Surface selected={secondId} />);
  expect(screen.queryByText("Original Host needs review")).toBeNull();
});
