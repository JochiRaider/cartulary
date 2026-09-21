import {
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
import { acceptedQueryMetadata } from "../../../testing/workbookQueryTestSupport";
import type { RecordPatchTransport } from "../../adapters/workbookRecordPatchTransport";
import { WorkbookHistoryContext } from "../../history/WorkbookHistoryContext";
import { entityRowFromApi } from "../../models/entityWorkbookModel";
import type { TimelineRelatedRecordPort } from "../../mutations/workbookMutationCommandPorts";
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
  runtime.explicitPatches.configure(
    { send: patch },
    undefined,
    async (_view, id) =>
      runtime.explicitPatches.latestRow(id) ?? row(id, 1).rawRow,
  );
  const query = {
    query: vi.fn(async () => ({
      kind: "accepted" as const,
      value: {
        incidentId: taskAuthority.incidentId,
        rows: [],
        viewSchemaId: "cartulary.view.timeline.v2",
        ...acceptedQueryMetadata("cartulary.view.timeline.v2"),
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
      relatedMutationCommands: {} as TimelineRelatedRecordPort,
      rows,
      selectedEntity: rows.find((item) => item.recordId === selected) ?? null,
      setEntityActionFeedback: setFeedback,
      setMutationError: setError,
      setSelectedRecordId: select,
      viewQuery: query,
    });
    return (
      <WorkbookHistoryContext value={runtime}>
        <button type="button" onClick={inspector.open}>
          Inspect selected
        </button>
        {inspector.node}
      </WorkbookHistoryContext>
    );
  }
  return { Surface, patch, runtime, select, refresh };
}
afterEach(cleanup);
function attachField(field: string) {
  const action = document.querySelector<HTMLButtonElement>(
    `[data-inspector-edit-field="${field}"]`,
  );
  if (!action) throw new Error(`Missing Edit action for ${field}`);
  fireEvent.click(action);
}
it("opens with saved values and submits only explicitly while Escape retains unfinished text", async () => {
  const f = fixture();
  render(<f.Surface />);
  fireEvent.click(screen.getByRole("button", { name: "Inspect selected" }));
  expect(screen.queryByTestId(genericEditValueTestId(schema))).toBeNull();
  attachField("host.location");
  const input = screen.getByTestId(genericEditValueTestId(schema));
  fireEvent.change(input, { target: { value: "Unfinished" } });
  fireEvent.blur(input);
  fireEvent.keyDown(input, { key: "Tab" });
  expect(f.patch).not.toHaveBeenCalled();
  fireEvent.keyDown(input, { key: "Escape" });
  expect(screen.queryByTestId(genericEditValueTestId(schema))).toBeNull();
  expect(
    screen.getByText("Unfinished work retained for Location."),
  ).not.toBeNull();
  fireEvent.click(
    screen.getByRole("button", { name: "Resume draft for Location" }),
  );
  const resumed = screen.getByTestId(genericEditValueTestId(schema));
  expect((resumed as HTMLInputElement).value).toBe("Unfinished");
  const update = screen.getByTestId(genericEditSubmitTestId(schema));
  const close = screen.getByRole("button", { name: "Close editor" });
  expect(update.parentElement).toBe(close.parentElement);
  fireEvent.click(close);
  expect(f.patch).not.toHaveBeenCalled();
  expect(
    screen.getByRole("button", { name: "Discard draft for Location" }),
  ).not.toBeNull();
  fireEvent.click(
    screen.getByRole("button", { name: "Resume draft for Location" }),
  );
  fireEvent.keyDown(screen.getByTestId(genericEditValueTestId(schema)), {
    key: "Enter",
    ctrlKey: true,
  });
  await waitFor(() => expect(f.patch).toHaveBeenCalledOnce());
});
function chooseField() {
  fireEvent.click(screen.getByRole("button", { name: "Inspect selected" }));
  const legacy = screen.queryByTestId(genericEditRecordSelectTestId(schema));
  if (legacy) fireEvent.change(legacy, { target: { value: firstId } });
  attachField("host.location");
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
  attachField("host.display_name");
  attachField("host.location");
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
