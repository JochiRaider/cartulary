import { mentionItemTestId } from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { type ComponentProps, useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import { taskAuthority } from "../../../testing/taskWorkbookTestSupport";
import {
  mentionReview,
  mentionWorkbookRow,
} from "../../../testing/timelineMentionTestSupport";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import type {
  RecordPatchOutcome,
  RecordPatchTransport,
} from "../../adapters/workbookRecordPatchTransport";
import { WorkbookInspectorDraftStore } from "../../inspector/WorkbookInspectorDraftStore";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { WorkbookExplicitPatchOwner } from "../../runtime/WorkbookExplicitPatchOwner";
import { WorkbookTimelineMentionOperationOwner } from "../actions/WorkbookTimelineMentionOperationOwner";
import { timelineMentionCandidatePolicy } from "../adapters/createTimelineMentionCandidateReader";
import { createTimelineMentionResolutionAdapter } from "../adapters/createTimelineMentionResolutionAdapter";
import { createTimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { createTimelineInspectorElementRegistry } from "../focus/timelineInspectorElementRegistry";
import { useTimelineMentionActions } from "../hooks/useTimelineMentionActions";
import { timelineCollectionBindings } from "../models/timelineFieldRegistry";
import { createDraftRow, rowFromApi } from "../models/timelineRowModel";
import {
  buildInspectorMentions,
  type CollectionItem,
} from "../models/workbookMentionChips";
import { TimelineCollectionCell } from "./TimelineCollectionCell";
import { TimelineMentionsPanel } from "./TimelineMentionsPanel";

afterEach(cleanup);

function inspectionRegistry() {
  return createTimelineInspectorElementRegistry({
    reviewGeneration: 1,
    lifecycleKey: "incident-1:timeline",
    subject: null,
  });
}

function fixture(
  registry = inspectionRegistry(),
): ComponentProps<typeof TimelineCollectionCell> {
  const draft = createDraftRow(1);
  return {
    editorDraftRegistry: createTimelineEditorDraftRegistry(),
    binding: timelineCollectionBindings[2],
    entityIndex: {},
    handleCollectionKeyDown: vi.fn(),
    handleSelectRow: vi.fn(),
    handleInspectCollection: vi.fn(),
    label: "Tags",
    queueCollectionSave: vi.fn(),
    readOnly: false,
    registerInput: vi.fn(),
    registerTrigger: registry.registerCollectionTrigger,
    isInspectionControlTarget: registry.isInspectionControlTarget,
    rememberReturnFocus: vi.fn(),
    registerCollectionItem: vi.fn(),
    row: {
      ...draft,
      key: "record-1",
      recordId: "record-1",
      rowVersion: 3,
      collectionValues: {
        ...draft.collectionValues,
        tags: [
          {
            itemRef: "tag-1",
            itemKind: "tag",
            rawText: "",
            displayText: "first",
          },
          {
            itemRef: "tag-2",
            itemKind: "tag",
            rawText: "",
            displayText: "  hidden Ω 東京  ",
          },
          {
            itemRef: "tag-3",
            itemKind: "tag",
            rawText: "",
            displayText: "long ".repeat(80),
          },
        ],
      },
    },
    surface: "grid",
    updateTimelineSurfaceFocusAnchor: vi.fn(),
  };
}

const tagRecordId = "00000000-0000-4000-8000-000000000919";
const tagContract = requireViewContract(timelineViewSchemaId);
function savedTagRow(
  version: number,
  tags: readonly string[],
  id = tagRecordId,
) {
  return {
    ...fullWorkbookViewRow(tagContract, id, version, {
      "timeline.tags": {
        items: tags.map((tag) => ({
          item_ref: `record_tag:${tag}`,
          item_kind: "tag",
          raw_text: tag,
          display_text: tag,
        })),
      },
    }),
    view_schema_id: timelineViewSchemaId,
  };
}
function tagRemovalFixture(
  tags: readonly string[] = ["alpha", "beta", "gamma"],
) {
  const initial = savedTagRow(1, tags);
  const refresh = vi.fn(async () => {});
  const registerConflict = vi.fn();
  const coordinate = vi.fn(async () => ({
    kind: "settled" as const,
    minimumRowVersion: 0,
  }));
  const patches = new WorkbookExplicitPatchOwner(
    taskAuthority.incidentId,
    { create: () => "timeline-remove-1" },
    {
      coordinate,
      registerConflict,
      accepted: vi.fn(),
      refresh,
    },
  );
  const send = vi.fn<RecordPatchTransport["send"]>(
    async (): Promise<RecordPatchOutcome> => ({
      kind: "acknowledged",
      receipt: {
        changeSetId: "change-1",
        viewSchemaId: timelineViewSchemaId,
        row: savedTagRow(
          2,
          tags.filter((tag) => tag !== "beta"),
        ),
      },
    }),
  );
  patches.configure(
    { send },
    undefined,
    async (_view, id) => patches.latestRow(id) ?? initial,
  );
  patches.setAuthority(taskAuthority);
  patches.observeQuery(initial);
  const drafts = new WorkbookInspectorDraftStore();
  drafts.setAuthority(taskAuthority);
  const base = fixture();
  function Surface({
    readOnly = false,
    visibleRecordId = tagRecordId,
  }: {
    readOnly?: boolean;
    visibleRecordId?: string;
  }) {
    const [row, setRow] = useState(rowFromApi(initial));
    return (
      <TimelineCollectionCell
        {...base}
        row={
          visibleRecordId === tagRecordId
            ? row
            : rowFromApi(savedTagRow(1, ["beta"], visibleRecordId))
        }
        surface="inspector"
        readOnly={readOnly}
        tagRemovalOwner={{
          patches,
          drafts,
          sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
          presentation: "timeline",
          accepted: (_key, mutation) => setRow(rowFromApi(mutation.row)),
        }}
      />
    );
  }
  return {
    patches,
    send,
    refresh,
    registerConflict,
    coordinate,
    Surface,
    base,
    initial,
  };
}

describe("Timeline collection inspection", () => {
  it("removes one exact saved tag after acknowledgement while preserving the unsent input and next focus", async () => {
    const f = tagRemovalFixture();
    f.base.editorDraftRegistry.setDraft(
      { rowKey: tagRecordId, field: "tags", surface: "grid" },
      "grid unsent Ω",
    );
    const held = deferred<RecordPatchOutcome>();
    f.send.mockReturnValueOnce(held.promise);
    render(<f.Surface />);
    const input = screen.getByRole("textbox") as HTMLInputElement;
    fireEvent.input(input, { target: { value: "  unsent Ω  " } });
    const button = screen.getByRole("button", { name: "Remove tag: beta" });
    button.focus();
    fireEvent.click(button);
    fireEvent.keyDown(button, { key: "Enter" });
    fireEvent.click(button);
    fireEvent.keyDown(button, { key: " " });
    fireEvent.click(button);
    await waitFor(() => expect(f.send).toHaveBeenCalledTimes(1));
    const request = JSON.parse(f.send.mock.calls[0]?.[0]?.body ?? "{}");
    expect(f.send.mock.calls[0]?.[0]?.recordId).toBe(tagRecordId);
    expect(request.changes).toEqual([
      {
        field_key: "timeline.tags",
        action_payload: {
          kind: "collection_actions_v1",
          actions: [{ op: "remove_tag", item_ref: "record_tag:beta" }],
        },
      },
    ]);
    expect(screen.getByRole("button", { name: "Remove tag: beta" })).toBe(
      button,
    );
    expect(button.getAttribute("aria-disabled")).toBe("true");
    expect(input.value).toBe("  unsent Ω  ");
    expect(
      f.base.editorDraftRegistry.draftValue({
        rowKey: tagRecordId,
        field: "tags",
        surface: "grid",
      }),
    ).toBe("grid unsent Ω");
    await act(async () =>
      held.resolve({
        kind: "acknowledged",
        receipt: {
          changeSetId: "change-1",
          viewSchemaId: timelineViewSchemaId,
          row: savedTagRow(2, ["alpha", "gamma"]),
        },
      }),
    );
    await waitFor(() =>
      expect(
        screen.queryByRole("button", { name: "Remove tag: beta" }),
      ).toBeNull(),
    );
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Remove tag: gamma" }),
    );
    expect(input.value).toBe("  unsent Ω  ");
    expect(
      f.base.editorDraftRegistry.draftValue({
        rowKey: tagRecordId,
        field: "tags",
        surface: "grid",
      }),
    ).toBe("grid unsent Ω");
    expect(f.base.queueCollectionSave).not.toHaveBeenCalled();
    expect(f.send).toHaveBeenCalledTimes(1);
  });

  it("keeps first and last saved tag keyboard focus in the local collection after removal", async () => {
    for (const target of ["alpha", "gamma"]) {
      const f = tagRemovalFixture();
      f.send.mockResolvedValueOnce({
        kind: "acknowledged",
        receipt: {
          changeSetId: `change-${target}`,
          viewSchemaId: timelineViewSchemaId,
          row: savedTagRow(
            2,
            ["alpha", "beta", "gamma"].filter((tag) => tag !== target),
          ),
        },
      });
      const view = render(<f.Surface />);
      const button = screen.getByRole("button", {
        name: `Remove tag: ${target}`,
      });
      button.focus();
      fireEvent.click(button);
      await waitFor(() =>
        expect(
          screen.queryByRole("button", { name: `Remove tag: ${target}` }),
        ).toBeNull(),
      );
      expect(document.activeElement).toBe(
        screen.getByRole("button", {
          name: target === "alpha" ? "Remove tag: beta" : "Remove tag: beta",
        }),
      );
      view.unmount();
    }
    const alone = tagRemovalFixture(["solo"]);
    alone.send.mockResolvedValueOnce({
      kind: "acknowledged",
      receipt: {
        changeSetId: "change-solo",
        viewSchemaId: timelineViewSchemaId,
        row: savedTagRow(2, []),
      },
    });
    render(<alone.Surface />);
    const last = screen.getByRole("button", { name: "Remove tag: solo" });
    last.focus();
    fireEvent.click(last);
    await waitFor(() =>
      expect(
        screen.queryByRole("button", { name: "Remove tag: solo" }),
      ).toBeNull(),
    );
    expect(screen.getByText("No items")).toBeTruthy();
    expect(document.activeElement).toBe(screen.getByRole("textbox"));
  });

  it("retains an uncertain saved-tag request for exact replay and read-only refresh recovery", async () => {
    const f = tagRemovalFixture();
    f.send.mockResolvedValueOnce({ kind: "uncertain" });
    f.refresh
      .mockResolvedValueOnce()
      .mockRejectedValueOnce(new Error("refresh failed"));
    render(<f.Surface />);
    fireEvent.click(screen.getByRole("button", { name: "Remove tag: beta" }));
    const retry = await screen.findByRole("button", {
      name: "Retry original change",
    });
    const firstRequest = f.send.mock.calls[0]?.[0];
    fireEvent.click(retry);
    await waitFor(() => expect(f.send).toHaveBeenCalledTimes(2));
    expect(f.send.mock.calls[1]?.[0]).toBe(firstRequest);
    await screen.findByRole("button", { name: "Refresh saved change" });
    expect(
      screen.getByText(/Remove tag: beta.*Saved, version 2/u),
    ).toBeTruthy();
    expect(f.send).toHaveBeenCalledTimes(2);
    fireEvent.click(
      screen.getByRole("button", { name: "Refresh saved change" }),
    );
    await waitFor(() =>
      expect(f.patches.getSnapshot().entries[0]?.reconciliation).toBe(
        "complete",
      ),
    );
    expect(f.send).toHaveBeenCalledTimes(2);
  });

  it("keeps rejected tags and denies viewer removal without losing inspection", async () => {
    const f = tagRemovalFixture();
    f.send.mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "validation", message: "Tag removal rejected" },
    });
    const view = render(<f.Surface />);
    const button = screen.getByRole("button", { name: "Remove tag: beta" });
    button.focus();
    fireEvent.click(button);
    await screen.findByText("Tag removal rejected");
    expect(screen.getByRole("button", { name: "Remove tag: beta" })).toBe(
      button,
    );
    expect(document.activeElement).toBe(button);
    act(() => f.patches.setAuthority({ ...taskAuthority, role: "viewer" }));
    expect(
      screen.queryByRole("button", { name: "Remove tag: beta" }),
    ).toBeNull();
    expect(
      screen.getAllByRole("note").map((note) => note.textContent),
    ).toContain("beta");
    view.rerender(<f.Surface readOnly />);
    expect(
      screen.queryByRole("button", { name: "Remove tag: beta" }),
    ).toBeNull();
    expect(
      screen.getAllByRole("note").map((note) => note.textContent),
    ).toContain("beta");
    expect(f.send).toHaveBeenCalledTimes(1);
  });

  it("rejects a stale saved tag after an earlier collection write without dispatching removal", async () => {
    const f = tagRemovalFixture();
    const gate = deferred<{ kind: "settled"; minimumRowVersion: number }>();
    f.coordinate.mockReturnValueOnce(gate.promise);
    render(<f.Surface />);
    fireEvent.click(screen.getByRole("button", { name: "Remove tag: beta" }));
    f.patches.observeQuery(savedTagRow(2, ["alpha", "gamma"]));
    await act(async () =>
      gate.resolve({ kind: "settled", minimumRowVersion: 2 }),
    );
    await screen.findByRole("button", { name: "Dismiss change notice" });
    expect(f.send).not.toHaveBeenCalled();
    expect(
      screen.getByRole("button", { name: "Remove tag: beta" }),
    ).toBeTruthy();
    expect(f.patches.getSnapshot().entries[0]?.phase).toBe(
      "preparation_failed",
    );
  });

  it("keeps a same-field tag conflict in collection review and never switches the captured record", async () => {
    const f = tagRemovalFixture();
    f.send.mockResolvedValueOnce({
      kind: "rejected",
      failure: {
        kind: "same_field_conflict",
        message: "Review changed tags",
        conflict: {
          record_id: tagRecordId,
          field_key: "timeline.tags",
          conflict_token: "tag-conflict",
          conflict_resolution_class: "collection_review",
          base_row_version: 1,
          current_row_version: 2,
          client_value: {
            actions: [{ op: "remove_tag", item_ref: "record_tag:beta" }],
          },
          server_value: { items: [] },
        },
      },
    });
    render(<f.Surface />);
    fireEvent.click(screen.getByRole("button", { name: "Remove tag: beta" }));
    await waitFor(() =>
      expect(f.patches.getSnapshot().entries[0]?.phase).toBe("conflict"),
    );
    expect(f.registerConflict).toHaveBeenCalledWith(
      expect.objectContaining({
        viewSchemaId: timelineViewSchemaId,
        rowLabel: tagRecordId,
        focusOrigin: "inspector",
      }),
    );
    expect(f.patches.getSnapshot().entries[0]?.intent.operationLabel).toBe(
      "Remove tag: beta",
    );
    expect(
      f.patches.getSnapshot().entries[0]?.intent.recoveryDestination,
    ).toEqual({
      kind: "region",
      panel: "relationships",
      regionId: "mentions",
    });
    expect(
      screen.getByRole("button", { name: "Remove tag: beta" }),
    ).toBeTruthy();
  });

  it("cancels saved-tag focus restoration after newer keyboard or scroll intent", async () => {
    const f = tagRemovalFixture();
    const held = deferred<RecordPatchOutcome>();
    f.send.mockReturnValueOnce(held.promise);
    const view = render(<f.Surface />);
    const remove = screen.getByRole("button", { name: "Remove tag: beta" });
    remove.focus();
    fireEvent.click(remove);
    await waitFor(() => expect(f.send).toHaveBeenCalledTimes(1));
    const input = screen.getByRole("textbox");
    fireEvent.keyDown(remove, { key: "Tab" });
    input.focus();
    await act(async () =>
      held.resolve({
        kind: "acknowledged",
        receipt: {
          changeSetId: "later",
          viewSchemaId: timelineViewSchemaId,
          row: savedTagRow(2, ["alpha", "gamma"]),
        },
      }),
    );
    await waitFor(() =>
      expect(
        screen.queryByRole("button", { name: "Remove tag: beta" }),
      ).toBeNull(),
    );
    expect(document.activeElement).toBe(input);
    view.unmount();
    const scrolled = tagRemovalFixture();
    const pending = deferred<RecordPatchOutcome>();
    scrolled.send.mockReturnValueOnce(pending.promise);
    render(<scrolled.Surface />);
    const second = screen.getByRole("button", { name: "Remove tag: beta" });
    second.focus();
    fireEvent.click(second);
    await waitFor(() => expect(scrolled.send).toHaveBeenCalledTimes(1));
    fireEvent.scroll(document);
    await act(async () =>
      pending.resolve({
        kind: "acknowledged",
        receipt: {
          changeSetId: "after-scroll",
          viewSchemaId: timelineViewSchemaId,
          row: savedTagRow(2, ["alpha", "gamma"]),
        },
      }),
    );
    await waitFor(() =>
      expect(
        screen.queryByRole("button", { name: "Remove tag: beta" }),
      ).toBeNull(),
    );
    expect(document.activeElement).not.toBe(
      screen.getByRole("button", { name: "Remove tag: gamma" }),
    );
  });
  it("keeps an admitted removal on its original record after inspector navigation", async () => {
    const f = tagRemovalFixture();
    const held = deferred<RecordPatchOutcome>();
    f.send.mockReturnValueOnce(held.promise);
    const view = render(<f.Surface />);
    const source = screen.getByRole("button", { name: "Remove tag: beta" });
    source.focus();
    fireEvent.click(source);
    await waitFor(() => expect(f.send).toHaveBeenCalledTimes(1));
    const otherRecordId = "00000000-0000-4000-8000-000000000920";
    view.rerender(<f.Surface visibleRecordId={otherRecordId} />);
    const other = screen.getByRole("button", { name: "Remove tag: beta" });
    other.focus();
    await act(async () =>
      held.resolve({
        kind: "acknowledged",
        receipt: {
          changeSetId: "after-navigation",
          viewSchemaId: timelineViewSchemaId,
          row: savedTagRow(2, ["alpha", "gamma"]),
        },
      }),
    );
    expect(f.send.mock.calls[0]?.[0]?.recordId).toBe(tagRecordId);
    expect(screen.getByRole("button", { name: "Remove tag: beta" })).toBe(
      other,
    );
    expect(document.activeElement).toBe(other);
    expect(f.patches.latestRow(tagRecordId)?.row_version).toBe(2);
  });
  it("offers exact saved-tag removal in the complete writable inspector collection", () => {
    const props = fixture();
    render(<TimelineCollectionCell {...props} surface="inspector" />);
    expect(
      screen.getByRole("button", { name: "Remove tag: first" }),
    ).toBeTruthy();
    expect(
      screen.getByRole("button", { name: /Remove tag: hidden Ω 東京/u }),
    ).toBeTruthy();
    expect((screen.getByRole("textbox") as HTMLInputElement).value).toBe("");
    expect(props.queueCollectionSave).not.toHaveBeenCalled();
  });

  it("publishes one revision for an equal-valued native input and fences older settlement", () => {
    const props = fixture();
    const identity = {
      rowKey: props.row.key,
      field: "tags" as const,
      surface: "grid" as const,
    };
    const token = "same native Ω";
    props.editorDraftRegistry.setDraft(identity, token);
    props.editorDraftRegistry.activateCollectionInput("record-1:tags:grid");
    const view = render(<TimelineCollectionCell {...props} />);
    const input = screen.getByRole("textbox") as HTMLInputElement;
    input.focus();
    const captured = props.editorDraftRegistry.captureRow(
      props.row.key,
      "grid",
      new Set(["timeline.tags"]),
    );
    const before =
      props.editorDraftRegistry.revisionForFocusKey("record-1:tags:grid");
    fireEvent.input(input, { target: { value: token } });
    const newer =
      props.editorDraftRegistry.revisionForFocusKey("record-1:tags:grid");
    expect(newer).toBe(before + 1);
    expect(props.queueCollectionSave).not.toHaveBeenCalled();
    view.rerender(
      <TimelineCollectionCell
        {...props}
        row={{ ...props.row, rowVersion: 4 }}
      />,
    );
    props.editorDraftRegistry.materializeRow(props.row);
    fireEvent.compositionEnd(input);
    fireEvent.blur(input, { relatedTarget: input });
    expect(
      props.editorDraftRegistry.revisionForFocusKey("record-1:tags:grid"),
    ).toBe(newer);
    input.focus();
    input.setSelectionRange(2, 6, "backward");
    act(() =>
      props.editorDraftRegistry.settleRevisions(
        props.row.key,
        captured ?? new Map(),
      ),
    );
    expect(props.editorDraftRegistry.draftValue(identity)).toBe(token);
    expect(input.value).toBe(token);
    expect(document.activeElement).toBe(input);
    expect([
      input.selectionStart,
      input.selectionEnd,
      input.selectionDirection,
    ]).toEqual([2, 6, "backward"]);
  });

  it("targets hidden tags without editing or committing pending text", () => {
    const registry = inspectionRegistry();
    const props = fixture(registry);
    const identity = {
      rowKey: props.row.key,
      field: "tags" as const,
      surface: "grid" as const,
    };
    props.editorDraftRegistry.setDraft(identity, "pending raw Ω");
    props.editorDraftRegistry.activateCollectionInput("record-1:tags:grid");
    const onGridKeyDown = vi.fn();
    const { rerender } = render(
      <table onKeyDown={onGridKeyDown}>
        <tbody>
          <tr>
            <td>
              <TimelineCollectionCell {...props} />
            </td>
          </tr>
        </tbody>
      </table>,
    );
    const overflow = screen.getByRole("button", {
      name: "Inspect 2 more tags",
    });
    const input = screen.getByRole("textbox");
    input.focus();
    fireEvent.keyDown(input, { key: "Delete" });
    expect(onGridKeyDown).not.toHaveBeenCalled();
    vi.mocked(props.handleCollectionKeyDown).mockClear();
    fireEvent.pointerDown(overflow);
    fireEvent.blur(input, { relatedTarget: overflow });
    fireEvent.keyDown(overflow, { key: "Enter" });
    fireEvent.keyDown(overflow, { key: " " });
    fireEvent.keyDown(overflow, { key: "F2" });
    expect(props.handleCollectionKeyDown).not.toHaveBeenCalled();
    overflow.focus();
    fireEvent.click(overflow);
    expect(props.handleInspectCollection).toHaveBeenCalledWith(
      "record-1",
      "timeline.tags",
      "tag-2",
    );
    expect(props.rememberReturnFocus).toHaveBeenCalledWith(
      "record-1",
      "timeline.tags",
      null,
    );
    expect(props.queueCollectionSave).not.toHaveBeenCalled();
    expect(props.editorDraftRegistry.draftValue(identity)).toBe(
      "pending raw Ω",
    );
    const other = render(
      <TimelineCollectionCell
        {...fixture(registry)}
        row={{ ...props.row, key: "record-2", recordId: "record-2" }}
      />,
    );
    input.focus();
    fireEvent.blur(input, {
      relatedTarget: within(other.container).getByRole("button", {
        name: "Inspect 2 more tags",
      }),
    });
    expect(props.queueCollectionSave).not.toHaveBeenCalled();
    other.unmount();
    const widthControl = document.createElement("button");
    widthControl.dataset.gridEditorExternalAction = "true";
    document.body.append(widthControl);
    input.focus();
    fireEvent.blur(input, { relatedTarget: widthControl });
    expect(props.queueCollectionSave).not.toHaveBeenCalled();
    expect(props.editorDraftRegistry.draftValue(identity)).toBe(
      "pending raw Ω",
    );
    widthControl.remove();
    rerender(
      <TimelineCollectionCell {...props} surface="inspector" readOnly />,
    );
    expect(
      screen.queryByRole("button", { name: "Inspect 2 more tags" }),
    ).toBeNull();
    const tags = screen.getAllByRole("note");
    expect(tags).toHaveLength(3);
    expect(tags[1]?.textContent).toBe("  hidden Ω 東京  ");
    expect(tags[2]?.textContent).toBe("long ".repeat(80));
    expect(props.registerCollectionItem).toHaveBeenCalledWith(
      "record-1",
      "timeline.tags",
      "tag-2",
      tags[1],
    );
    expect((screen.getByRole("textbox") as HTMLInputElement).value).toBe("");
    expect((screen.getByRole("textbox") as HTMLInputElement).readOnly).toBe(
      true,
    );
  });

  it("mounts only explicit grid authoring, preserves native selection on refresh, and retains independent detached drafts", () => {
    const props = fixture();
    const grid = {
      rowKey: props.row.key,
      field: "tags" as const,
      surface: "grid" as const,
    };
    const inspector = { ...grid, surface: "inspector" as const };
    const view = render(<TimelineCollectionCell {...props} />);
    expect(screen.queryByRole("textbox")).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Add tags token" }));
    const input = screen.getByRole("textbox") as HTMLInputElement;
    expect(document.activeElement).toBe(input);
    fireEvent.input(input, { target: { value: "raw Ω 東京" } });
    input.setSelectionRange(2, 5, "backward");
    view.rerender(
      <TimelineCollectionCell
        {...props}
        row={{ ...props.row, rowVersion: 4 }}
      />,
    );
    expect(screen.getByRole("textbox")).toBe(input);
    expect(document.activeElement).toBe(input);
    expect([
      input.selectionStart,
      input.selectionEnd,
      input.selectionDirection,
    ]).toEqual([2, 5, "backward"]);
    act(() => props.editorDraftRegistry.setDraft(inspector, "Inspector only"));
    const captured = props.editorDraftRegistry.captureRow(
      props.row.key,
      "grid",
      new Set(["timeline.tags"]),
    );
    fireEvent.compositionStart(input);
    act(() =>
      props.editorDraftRegistry.settleRevisions(
        props.row.key,
        captured ?? new Map(),
      ),
    );
    expect(input.value).toBe("raw Ω 東京");
    fireEvent.compositionEnd(input);
    view.unmount();
    expect(props.editorDraftRegistry.draftValue(grid)).toBe("raw Ω 東京");
    render(<TimelineCollectionCell {...props} />);
    expect(screen.queryByRole("textbox")).toBeNull();
    expect(screen.getByLabelText("Tags token draft retained")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Add tags token" }));
    expect((screen.getByRole("textbox") as HTMLInputElement).value).toBe(
      "raw Ω 東京",
    );
    expect(props.editorDraftRegistry.draftValue(inspector)).toBe(
      "Inspector only",
    );
    const restored = screen.getByRole("textbox");
    fireEvent.compositionStart(restored);
    fireEvent.keyDown(restored, { key: "Escape", isComposing: true });
    expect(props.editorDraftRegistry.draftValue(grid)).toBe("raw Ω 東京");
    fireEvent.compositionEnd(restored);
    fireEvent.keyDown(restored, { key: "Escape" });
    expect(screen.queryByRole("textbox")).toBeNull();
    expect(props.editorDraftRegistry.draftValue(grid)).toBe("");
    expect(props.editorDraftRegistry.draftValue(inspector)).toBe(
      "Inspector only",
    );
    expect(props.queueCollectionSave).not.toHaveBeenCalled();
  });

  it("keeps recordless draft text local and registers independent editor surfaces", () => {
    const props = fixture();
    const row = {
      ...createDraftRow(1),
      collectionDrafts: {
        hostRefs: "",
        identityRefs: "",
        tags: "raw draft 東京",
      },
    };
    const { rerender } = render(
      <TimelineCollectionCell {...props} row={row} />,
    );
    const input = screen.getByRole("textbox");
    expect((input as HTMLInputElement).value).toBe("raw draft 東京");
    expect(screen.queryByRole("button")).toBeNull();
    expect(props.handleInspectCollection).not.toHaveBeenCalled();
    expect(props.queueCollectionSave).not.toHaveBeenCalled();
    expect(props.registerInput).toHaveBeenCalledWith(
      row.key,
      "tags",
      "grid",
      input,
    );
    rerender(<TimelineCollectionCell {...props} surface="inspector" />);
    expect(props.registerInput).toHaveBeenCalledWith(
      "record-1",
      "tags",
      "inspector",
      screen.getByRole("textbox"),
    );
  });
  it("keeps complete mention details readable when management capability is lost", () => {
    const item: CollectionItem = {
      entityMentionId: "host-1",
      itemRef: "entity_mention:host-1",
      entityType: "host",
      itemKind: "resolved_ref",
      rawText: "  alias Ω 東京  ",
      displayText: "Canonical host",
      resolvedRecordId: "host-target",
      mentionRowVersion: 2,
      resolutionMethod: "auto_match",
      autoResolved: true,
      provenance: "auto_match",
      confidence: 100,
      matchedAliasText: "alias Ω 東京",
    };
    const mentions = buildInspectorMentions(
      {
        recordId: "record-1",
        collectionValues: {
          hostRefs: [
            item,
            {
              ...item,
              entityMentionId: "host-2",
              itemRef: "entity_mention:host-2",
              displayText: "Other host",
            },
          ],
          identityRefs: [],
        },
      },
      [],
    );
    let props: Omit<ComponentProps<typeof TimelineMentionsPanel>, "actions"> = {
      sourceRecordId: "record-1",
      registerCollectionItem: vi.fn(),
      entityIndex: {},
      getRelationshipLabel: () => "Hosts",
      inspectorMentions: mentions,
      registerMention: vi.fn(),
      onSelectMention: vi.fn(),
      selectedMention: mentions[0] ?? null,
    };
    const review = mentionReview();
    const owner = new WorkbookTimelineMentionOperationOwner(
      review.subject.incidentId,
      { create: () => "test-key" },
      { remember: vi.fn(), settle: vi.fn(), accepted: vi.fn() },
    );
    const send = vi.fn();
    owner.configure({
      ...createTimelineMentionResolutionAdapter({ apiBase: undefined }),
      send,
    });
    const row = {
      ...mentionWorkbookRow(review),
      recordId: "record-1",
      collectionValues: {
        ...mentionWorkbookRow(review).collectionValues,
        hostRefs: [item],
      },
    };
    const candidatePort = {
      policy: timelineMentionCandidatePolicy,
      page: vi.fn(async () => ({
        kind: "accepted" as const,
        value: { candidates: [], hasMore: false, nextCursor: null },
      })),
    };
    function Panel({ viewer = false }: { viewer?: boolean }) {
      owner.setAuthority({
        ...review.authority,
        role: viewer ? "viewer" : "editor",
      });
      const [selectedTargetId, setSelectedTargetId] = useState("");
      const actions = useTimelineMentionActions({
        owner,
        candidatePort,
        rowsRef: { current: [row] },
        earlierSaves: { current: Promise.resolve() },
        selectedMention: props.selectedMention,
        selectedMentionRef: review.subject.itemRef,
        selectedRowId: review.subject.sourceRecordId,
        inspectorReviewGeneration: 0,
        inspectorAttachmentGeneration: 0,
        refreshProjection: async () => {},
        reviewSurfaceKey: "cell-test",
        selectedTargetId,
        setSelectedTargetId,
        presentationKey: "cell-test",
        presentationActive: true,
        waitForCommittedRecordIdle: async () => ({ row, rowVersion: 4 }),
        setInspectorMessage: vi.fn(),
      });
      return <TimelineMentionsPanel {...props} actions={actions} />;
    }
    const { rerender } = render(<Panel />);
    const selected = screen.getByRole("region", {
      name: "Selected Hosts item",
    });
    expect(selected.previousElementSibling?.getAttribute("data-testid")).toBe(
      mentionItemTestId(item.itemRef),
    );
    expect(selected.nextElementSibling?.getAttribute("data-testid")).toBe(
      mentionItemTestId("entity_mention:host-2"),
    );
    const correction = screen.getByText(
      "Correction and resolution",
    ).parentElement;
    const details = screen.getByText("Mention details");
    expect((details.parentElement as HTMLDetailsElement).open).toBe(false);
    fireEvent.click(details);
    expect((details.parentElement as HTMLDetailsElement).open).toBe(true);
    expect(
      screen.getByRole("button", {
        name: "Auto-resolved host: Canonical host; matched alias Ω 東京",
      }),
    ).toBeTruthy();
    expect(
      screen.getByText("Source text").nextElementSibling?.textContent,
    ).toBe(item.rawText);
    expect(
      screen.getByText("Matched alias").nextElementSibling?.textContent,
    ).toBe(item.matchedAliasText);
    expect(screen.getByText("Provenance").nextElementSibling?.textContent).toBe(
      "auto_match",
    );
    rerender(<Panel viewer />);
    expect((screen.getByRole("combobox") as HTMLInputElement).disabled).toBe(
      false,
    );
    expect(
      (
        screen.getByRole("button", {
          name: "Dismiss",
        }) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    fireEvent.click(
      screen.getByRole("button", {
        name: "Auto-resolved host: Canonical host; matched alias Ω 東京",
      }),
    );
    expect(props.onSelectMention).toHaveBeenCalledWith(
      "record-1",
      item.itemRef,
    );
    expect(send).not.toHaveBeenCalled();
    const firstMention = mentions[0];
    if (!firstMention) throw new Error("Missing selected mention fixture");
    const dismissed = {
      ...firstMention,
      status: "dismissed" as const,
      isActiveRelationshipValue: false,
    };
    props = {
      ...props,
      inspectorMentions: [dismissed, ...mentions.slice(1)],
      selectedMention: dismissed,
    };
    rerender(<Panel viewer />);
    expect(screen.getByText("Correction and resolution").parentElement).toBe(
      correction,
    );
    expect(
      screen
        .getByRole("region", { name: "Selected Hosts item" })
        .previousElementSibling?.getAttribute("data-testid"),
    ).toBe(mentionItemTestId(item.itemRef));
  });
});
