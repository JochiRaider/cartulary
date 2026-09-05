import {
  cleanup,
  fireEvent,
  render,
  screen,
  within,
} from "@testing-library/react";
import type { ComponentProps } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createTimelineInspectorElementRegistry } from "../focus/timelineInspectorElementRegistry";
import { timelineCollectionBindings } from "../models/timelineFieldRegistry";
import { createDraftRow } from "../models/timelineRowModel";
import {
  buildInspectorMentions,
  type CollectionItem,
} from "../models/workbookMentionChips";
import { TimelineCollectionCell } from "./TimelineCollectionCell";
import { TimelineMentionsPanel } from "./TimelineMentionsPanel";

afterEach(cleanup);

function inspectionRegistry() {
  return createTimelineInspectorElementRegistry({
    invalidationGeneration: 1,
    lifecycleKey: "incident-1:timeline",
    subject: null,
  });
}

function fixture(
  registry = inspectionRegistry(),
): ComponentProps<typeof TimelineCollectionCell> {
  const draft = createDraftRow(1);
  return {
    activateCollectionInput: vi.fn(),
    activeCollectionInputKey: null,
    binding: timelineCollectionBindings[2],
    deactivateCollectionInput: vi.fn(),
    entityIndex: {},
    handleCollectionInputChange: vi.fn(),
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
    retainedDraft: undefined,
    retainDraft: vi.fn(),
    updateTimelineSurfaceFocusAnchor: vi.fn(),
  };
}

describe("Timeline collection inspection", () => {
  it("targets hidden tags without editing or committing pending text", () => {
    const registry = inspectionRegistry();
    const props = fixture(registry);
    const onGridKeyDown = vi.fn();
    const { rerender } = render(
      <table onKeyDown={onGridKeyDown}>
        <tbody>
          <tr>
            <td>
              <TimelineCollectionCell
                {...props}
                retainedDraft="pending raw Ω"
              />
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
    expect(props.retainDraft).toHaveBeenCalledWith("pending raw Ω");
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
    rerender(
      <TimelineCollectionCell
        {...props}
        surface="inspector"
        readOnly
        retainedDraft="pending raw Ω"
      />,
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
    expect((screen.getByRole("textbox") as HTMLInputElement).value).toBe(
      "pending raw Ω",
    );
    expect((screen.getByRole("textbox") as HTMLInputElement).readOnly).toBe(
      true,
    );
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
        collectionValues: { hostRefs: [item], identityRefs: [] },
      },
      [],
    );
    const props: ComponentProps<typeof TimelineMentionsPanel> = {
      sourceRecordId: "record-1",
      canManageMentions: true,
      registerCollectionItem: vi.fn(),
      entityIndex: {},
      getRelationshipLabel: () => "Hosts",
      hostEntities: [],
      identityEntities: [],
      inspectorMentions: mentions,
      registerMention: vi.fn(),
      onResolveTargetChange: vi.fn(),
      onSelectMention: vi.fn(),
      onSetInspectorMessage: vi.fn(),
      onCreateEntityFromMention: vi.fn(),
      onSubmitMentionAction: vi.fn(),
      selectedMention: mentions[0] ?? null,
      selectedResolveTargetId: "",
    };
    const { rerender } = render(<TimelineMentionsPanel {...props} />);
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
    rerender(<TimelineMentionsPanel {...props} canManageMentions={false} />);
    expect(screen.getAllByRole("button")).toHaveLength(1);
    expect(screen.queryByRole("combobox")).toBeNull();
    fireEvent.click(screen.getByRole("button"));
    expect(props.onSelectMention).toHaveBeenCalledWith(
      "record-1",
      item.itemRef,
    );
    expect(props.onSubmitMentionAction).not.toHaveBeenCalled();
  });
});
