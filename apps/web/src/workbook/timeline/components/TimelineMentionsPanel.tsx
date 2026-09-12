import {
  mentionItemTestId,
  relationshipItemsTestId,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import { type CSSProperties, type ReactNode, useRef } from "react";
import {
  WorkbookRelationshipChip,
  WorkbookRelationshipChipDetails,
} from "../../components/WorkbookRelationshipChip";
import { relationshipChipAccessibleName } from "../../models/workbookRelationshipChip";
import type { TimelineInspectorElementRegistry } from "../focus/timelineInspectorElementRegistry";
import {
  type InspectorMention,
  timelineRelationshipChipPresentation,
} from "../models/workbookMentionChips";
import {
  TimelineMentionActionControls,
  type TimelineMentionActions,
} from "./TimelineMentionActionControls";
import {
  inspectorSectionStyle,
  sectionTitleStyle,
} from "./TimelineWorkbookStyles";

type TimelineMentionsPanelProps = {
  readonly sourceRecordId: string | null;
  readonly registerCollectionItem: TimelineInspectorElementRegistry["registerCollectionItem"];
  readonly entityIndex: Record<string, { label: string }>;
  readonly getRelationshipLabel: (
    fieldKey: InspectorMention["fieldKey"],
  ) => string;
  readonly inspectorMentions: readonly InspectorMention[];
  readonly relationshipEditors?: ReactNode;
  readonly registerMention: (
    sourceRecordId: string,
    itemRef: string,
    element: HTMLButtonElement | null,
  ) => void;
  readonly onSelectMention: (rowRecordId: string, itemRef: string) => void;
  readonly selectedMention: InspectorMention | null;
  readonly actions: TimelineMentionActions;
};
export function TimelineMentionsPanel(props: TimelineMentionsPanelProps) {
  const { selectedMention, entityIndex, getRelationshipLabel, actions } = props;
  return (
    <>
      <MentionGroups {...props} />
      {selectedMention ? (
        <section style={inspectorSectionStyle}>
          <h3 style={sectionTitleStyle}>Selected mention</h3>
          <p style={selectedMentionTextStyle}>
            {getRelationshipLabel(selectedMention.fieldKey)}
          </p>
          <WorkbookRelationshipChipDetails
            presentation={timelineRelationshipChipPresentation({
              entityIndex,
              item: selectedMention,
              selected: true,
            })}
          />
          <TimelineMentionActionControls
            key={selectedMention.entityMentionId ?? selectedMention.itemRef}
            actions={actions}
          />
        </section>
      ) : null}
    </>
  );
}

function MentionGroups({
  sourceRecordId,
  entityIndex,
  inspectorMentions,
  relationshipEditors,
  registerMention,
  registerCollectionItem,
  onSelectMention,
  selectedMention,
}: {
  readonly sourceRecordId: string | null;
  readonly entityIndex: Record<string, { label: string }>;
  readonly inspectorMentions: readonly InspectorMention[];
  readonly relationshipEditors?: ReactNode;
  readonly registerMention: TimelineMentionsPanelProps["registerMention"];
  readonly registerCollectionItem: TimelineInspectorElementRegistry["registerCollectionItem"];
  readonly onSelectMention: (rowRecordId: string, itemRef: string) => void;
  readonly selectedMention: InspectorMention | null;
}) {
  const buttons = useRef(new Map<string, HTMLButtonElement>());
  const renderMention = (
    item: InspectorMention,
    items: readonly InspectorMention[],
  ) => {
    const presentation = timelineRelationshipChipPresentation({
      entityIndex,
      item,
      selected: selectedMention?.itemRef === item.itemRef,
    });
    return (
      <button
        key={item.itemRef}
        type="button"
        data-testid={mentionItemTestId(item.itemRef)}
        aria-label={relationshipChipAccessibleName(presentation)}
        aria-pressed={presentation.selected}
        ref={(element) => {
          registerMention(item.rowRecordId, item.itemRef, element);
          registerCollectionItem(
            item.rowRecordId,
            item.fieldKey,
            item.itemRef,
            element,
          );
          if (element === null) buttons.current.delete(item.itemRef);
          else buttons.current.set(item.itemRef, element);
        }}
        style={{
          ...mentionListButtonStyle,
          ...(presentation.selected ? mentionListButtonSelectedStyle : null),
        }}
        onClick={() => onSelectMention(item.rowRecordId, item.itemRef)}
        onKeyDown={(event) => {
          if (event.key === "ArrowLeft" || event.key === "ArrowRight") {
            const index = items.findIndex(
              (candidate) => candidate.itemRef === item.itemRef,
            );
            const next = items[index + (event.key === "ArrowLeft" ? -1 : 1)];
            if (next)
              buttons.current.get(next.itemRef)?.focus({ preventScroll: true });
            event.preventDefault();
          }
          if (event.key !== "Tab" && event.key !== "Escape")
            event.stopPropagation();
        }}
      >
        <WorkbookRelationshipChip
          expanded
          decorative
          presentation={presentation}
        />
      </button>
    );
  };
  return (
    <div
      data-testid={timelineInspectorSectionTestId("relationships")}
      style={inspectorSectionStyle}
    >
      {relationshipEditors}
      {(["timeline.host_refs", "timeline.identity_refs"] as const).map(
        (fieldKey) => {
          const items = inspectorMentions.filter(
            (item) => item.fieldKey === fieldKey,
          );
          const active = items.filter((item) => item.isActiveRelationshipValue);
          const dismissed = items.filter((item) => item.status === "dismissed");
          const recordId = sourceRecordId;
          return (
            <section
              key={fieldKey}
              aria-label={
                fieldKey === "timeline.host_refs"
                  ? "Host mentions"
                  : "Identity mentions"
              }
              style={mentionGroupColumnStyle}
            >
              <p style={groupLabelStyle}>
                {fieldKey === "timeline.host_refs" ? "Hosts" : "Identities"}
              </p>
              <div
                data-testid={
                  recordId === null
                    ? undefined
                    : relationshipItemsTestId(recordId, fieldKey)
                }
                style={mentionGroupColumnStyle}
              >
                {active.length === 0 ? (
                  <span>No items</span>
                ) : (
                  active.map((item) => renderMention(item, active))
                )}
              </div>
              {dismissed.length > 0 ? (
                <div style={mentionGroupColumnStyle}>
                  <p style={groupLabelStyle}>Dismissed in this session</p>
                  <p>Observed here; use History for durable changes.</p>
                  {dismissed.map((item) => renderMention(item, dismissed))}
                </div>
              ) : null}
            </section>
          );
        },
      )}
    </div>
  );
}

const mentionGroupColumnStyle = {
  display: "grid",
  gap: "0.5rem",
} satisfies CSSProperties;

const groupLabelStyle = {
  margin: 0,
  fontSize: "0.8rem",
  letterSpacing: "0.08em",
  textTransform: "uppercase",
  color: "var(--ct-colors-ink-subtle)",
} satisfies CSSProperties;

const mentionListButtonStyle = {
  border: "none",
  background: "transparent",
  color: "var(--ct-colors-ink)",
  padding: 0,
  textAlign: "left",
  cursor: "pointer",
} satisfies CSSProperties;

const mentionListButtonSelectedStyle = {
  boxShadow: "0 0 0 2px var(--ct-colors-accent)",
  outline: "2px solid transparent",
  outlineOffset: "2px",
} satisfies CSSProperties;

const selectedMentionTextStyle = {
  margin: 0,
  overflowWrap: "anywhere",
} satisfies CSSProperties;
