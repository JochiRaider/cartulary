import {
  genericEditSubmitTestId,
  timelineScalarEditorTestId,
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
import { afterEach, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import { taskAuthority } from "../../../testing/taskWorkbookTestSupport";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import type {
  RecordPatchOutcome,
  RecordPatchTransport,
} from "../../adapters/workbookRecordPatchTransport";
import { WorkbookInspectorDraftStore } from "../../inspector/WorkbookInspectorDraftStore";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { WorkbookExplicitPatchOwner } from "../../runtime/WorkbookExplicitPatchOwner";
import { rowFromApi } from "../models/timelineRowModel";
import { TimelineInspectorDetails } from "./TimelineInspectorDetails";

const field = "timeline.raw_activity_text";
const otherField = "timeline.activity_synopsis_text";
const recordId = "00000000-0000-4000-8000-000000000901";
const contract = requireViewContract(timelineViewSchemaId);
function raw(version = 1, value = "Saved activity", id = recordId) {
  return {
    ...fullWorkbookViewRow(contract, id, version, { [field]: value }),
    view_schema_id: timelineViewSchemaId,
  };
}
function fixture() {
  const drafts = new WorkbookInspectorDraftStore();
  drafts.setAuthority(taskAuthority);
  let sequence = 0;
  const refresh = vi.fn(async () => {});
  const accepted = vi.fn();
  const patches = new WorkbookExplicitPatchOwner(
    taskAuthority.incidentId,
    { create: () => `inspector-${++sequence}` },
    {
      coordinate: async () => ({ kind: "settled", minimumRowVersion: 0 }),
      registerConflict: vi.fn(),
      accepted: vi.fn(),
      refresh,
    },
  );
  const send = vi.fn<RecordPatchTransport["send"]>(async () => ({
    kind: "acknowledged",
    receipt: {
      changeSetId: "change",
      viewSchemaId: timelineViewSchemaId,
      row: raw(2, "Submitted activity"),
    },
  }));
  patches.configure(
    { send },
    undefined,
    async (_view, id) => patches.latestRow(id) ?? raw(),
  );
  patches.setAuthority(taskAuthority);
  patches.observeQuery(raw());
  const owner = {
    patches,
    drafts,
    refresh,
    accepted,
    sheetRef: { kind: "view_schema" as const, id: timelineViewSchemaId },
    presentation: "timeline",
  };
  const surface = (row = raw()) => (
    <TimelineInspectorDetails
      key={row.record_id}
      row={rowFromApi(row)}
      owner={owner}
    />
  );
  return { owner, send, surface };
}
function edit(fieldKey = field) {
  const button = document.querySelector<HTMLButtonElement>(
    `[data-inspector-edit-field="${fieldKey}"]`,
  );
  if (!button) throw new Error(`Missing Edit ${fieldKey}`);
  fireEvent.click(button);
  return screen.getByTestId(
    timelineScalarEditorTestId({ fieldKey, recordId, surface: "inspector" }),
  ) as HTMLTextAreaElement;
}
afterEach(cleanup);

it("Timeline Details keeps a rejected recovery beside its field with one new announcement", async () => {
  const f = fixture();
  f.send.mockResolvedValueOnce({ kind: "uncertain" });
  f.send.mockResolvedValueOnce({
    kind: "rejected",
    failure: { kind: "validation", message: "Review this recovery rejection." },
  });
  render(f.surface());
  const input = edit();
  fireEvent.change(input, { target: { value: "  Exact source text Ω\t\n  " } });
  fireEvent.keyDown(input, { key: "Enter", metaKey: true });
  const retry = await screen.findByRole("button", {
    name: "Retry original change",
  });
  fireEvent.click(retry);
  await waitFor(() => expect(f.send).toHaveBeenCalledTimes(2));
  expect(f.send.mock.calls[1]?.[0]).toBe(f.send.mock.calls[0]?.[0]);
  await waitFor(() => expect(screen.getAllByRole("alert")).toHaveLength(1));
  expect(screen.getByRole("alert").textContent).toBe(
    "Review this recovery rejection.",
  );
  expect(input.value).toBe("  Exact source text Ω\t\n  ");
  // A rejection during replay does not prove the first attempt failed.
  expect(f.owner.patches.getSnapshot().entries[0]?.phase).toBe("uncertain");
  expect(
    screen
      .getByRole("button", { name: "Retry original change" })
      .closest("[data-inspector-editor-field]"),
  ).not.toBeNull();
});

it("Timeline Details copies saved text exactly without including disclosure controls", () => {
  const f = fixture();
  const value = "  Saved Ω\t\nsecond line\n";
  const view = render(f.surface(raw(1, value)));
  const saved = document.querySelector(
    `[data-inspector-saved-field="${field}"] dd > div[id]`,
  );
  if (!saved?.firstChild) throw new Error("Missing saved value");
  const selection = document.getSelection();
  const range = document.createRange();
  range.selectNodeContents(saved);
  selection?.removeAllRanges();
  selection?.addRange(range);
  const setData = vi.fn();
  fireEvent.copy(saved, { clipboardData: { setData } });
  expect(setData).toHaveBeenLastCalledWith("text/plain", value);
  range.setStart(saved.firstChild, 2);
  range.setEnd(saved.firstChild, 9);
  fireEvent.copy(saved, { clipboardData: { setData } });
  expect(setData).toHaveBeenLastCalledWith("text/plain", "Saved Ω");
  range.collapse();
  fireEvent.copy(saved, { clipboardData: { setData } });
  expect(setData).toHaveBeenCalledTimes(2);
  selection?.removeAllRanges();
  expect(f.send).not.toHaveBeenCalled();
  const whitespace = " \t\r\n  ";
  view.rerender(f.surface(raw(1, whitespace)));
  expect(screen.getByText("Whitespace only")).not.toBeNull();
  expect(
    document.querySelector(
      `[data-inspector-saved-field="${field}"] dd > div[id]`,
    )?.textContent,
  ).toBe(whitespace);
});

it("Timeline Details reads saved values and retains authoring until explicit submission", async () => {
  const f = fixture();
  const view = render(f.surface());
  expect(screen.queryByRole("textbox")).toBeNull();
  const input = edit();
  fireEvent.change(input, { target: { value: "Submitted activity" } });
  input.setSelectionRange(3, 7);
  fireEvent.blur(input);
  fireEvent.keyDown(input, { key: "Tab" });
  fireEvent.keyDown(input, { key: "Enter" });
  expect(f.send).not.toHaveBeenCalled();
  expect(screen.getByText("Saved activity")).toBeTruthy();
  view.rerender(f.surface(raw(2)));
  expect(screen.getByRole("textbox")).toBe(input);
  expect(input.selectionStart).toBe(3);
  expect(input.selectionEnd).toBe(7);
  edit(otherField);
  edit();
  fireEvent.keyDown(screen.getByRole("textbox"), { key: "Escape" });
  expect(screen.queryByRole("textbox")).toBeNull();
  edit();
  expect(f.send).not.toHaveBeenCalled();
  fireEvent.keyDown(screen.getByRole("textbox"), {
    key: "Enter",
    ctrlKey: true,
  });
  await waitFor(() =>
    expect(
      f.send,
      JSON.stringify(f.owner.patches.getSnapshot().entries),
    ).toHaveBeenCalledOnce(),
  );
  expect(JSON.parse(f.send.mock.calls[0]?.[0].body ?? "{}").changes).toEqual([
    { field_key: field, value: "Submitted activity" },
  ]);
  act(() => f.owner.drafts.setAuthority({ ...taskAuthority, role: "viewer" }));
  const reason = document.querySelector("[data-inspector-read-only-reason]");
  const fields = document.querySelector("dl");
  expect(reason?.textContent).toBe("Current access permits reading only.");
  expect(fields?.getAttribute("aria-describedby")).toBe(reason?.id);
  if (!reason || !fields) throw new Error("Missing read-only explanation");
  expect(
    reason.compareDocumentPosition(fields) & Node.DOCUMENT_POSITION_FOLLOWING,
  ).toBeTruthy();
  const action = document.querySelector<HTMLButtonElement>(
    `[data-inspector-edit-field="${otherField}"]`,
  );
  expect(action?.disabled).toBe(true);
  expect(action?.getAttribute("aria-describedby")).toBe(reason.id);
});

it("Timeline Details keeps newer authoring after acknowledgement and retries refresh without writes", async () => {
  const f = fixture();
  const pending = deferred<RecordPatchOutcome>();
  f.send.mockReturnValue(pending.promise);
  f.owner.refresh.mockRejectedValue(new Error("Read unavailable"));
  render(f.surface());
  const input = edit();
  fireEvent.change(input, { target: { value: "Submitted activity" } });
  fireEvent.click(
    screen.getByTestId(genericEditSubmitTestId(timelineViewSchemaId)),
  );
  await waitFor(() =>
    expect(
      f.send,
      JSON.stringify(f.owner.patches.getSnapshot().entries),
    ).toHaveBeenCalledOnce(),
  );
  fireEvent.change(input, { target: { value: "Newer authoring" } });
  input.setSelectionRange(2, 6);
  await act(async () =>
    pending.resolve({
      kind: "acknowledged",
      receipt: {
        changeSetId: "change",
        viewSchemaId: timelineViewSchemaId,
        row: raw(2, "Submitted activity"),
      },
    }),
  );
  await waitFor(() =>
    expect(f.owner.patches.getSnapshot().entries[0]?.reconciliation).toBe(
      "required",
    ),
  );
  expect(input.value).toBe("Newer authoring");
  expect(input.selectionStart).toBe(2);
  expect(input.selectionEnd).toBe(6);
  expect(f.owner.accepted).toHaveBeenCalledOnce();
  const entry = f.owner.patches.getSnapshot().entries[0];
  if (!entry) throw new Error("Missing retained receipt");
  await act(async () => {
    await f.owner.patches.refresh(entry.id);
    await f.owner.patches.refresh(entry.id);
  });
  expect(
    f.send,
    JSON.stringify(f.owner.patches.getSnapshot().entries),
  ).toHaveBeenCalledOnce();
  expect(f.owner.refresh).toHaveBeenCalledTimes(3);
  expect(f.owner.patches.getSnapshot().entries[0]?.receipt?.changeSetId).toBe(
    "change",
  );
});

it("Timeline Details sends explicit null empty and unchanged source text as distinct intents", async () => {
  for (const value of [null, "", "  source\r\ntext\t  "] as const) {
    const f = fixture();
    const initial = raw();
    if (value) {
      f.owner.drafts.update(
        {
          viewSchemaId: timelineViewSchemaId,
          recordId,
          fieldKey: field,
          action: "value",
        },
        initial,
        value,
        "previous-editor",
      );
      f.owner.drafts.detach("previous-editor");
    }
    f.owner.patches.observeQuery(initial);
    render(f.surface(initial));
    const input = edit();
    if (value === null)
      fireEvent.click(
        screen.getByRole("button", {
          name: `Clear ${contract.fieldMap[field]?.label}`,
        }),
      );
    else if (value === "") fireEvent.change(input, { target: { value } });
    fireEvent.click(
      screen.getByTestId(genericEditSubmitTestId(timelineViewSchemaId)),
    );
    await waitFor(() => expect(f.send).toHaveBeenCalledOnce());
    expect(JSON.parse(f.send.mock.calls[0]?.[0].body ?? "{}").changes).toEqual([
      { field_key: field, value },
    ]);
    cleanup();
  }
});
