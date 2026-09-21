import {
  mentionItemTestId,
  relationshipItemsTestId,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import { type CSSProperties, Fragment, type ReactNode, useRef } from "react";
import {
  WorkbookRelationshipChip,
  WorkbookRelationshipChipDetails,
} from "../../components/WorkbookRelationshipChip";
import { workbookTypography } from "../../components/workbookFormStyles";
import { WorkbookInspectorFeedbackView } from "../../inspector/presentation/WorkbookInspectorFeedback";
import type { WorkbookInspectorFeedback } from "../../inspector/workbookInspectorErrorModel";
import { relationshipChipAccessibleName } from "../../models/workbookRelationshipChip";
import type { TimelineInspectorElementRegistry } from "../focus/timelineInspectorElementRegistry";
import type { CollectionFieldKey } from "../models/timelineFieldRegistry";
import {
  type InspectorMention,
  timelineRelationshipChipPresentation,
} from "../models/workbookMentionChips";
import {
  TimelineMentionActionControls,
  type TimelineMentionActions,
} from "./TimelineMentionActionControls";
import { inspectorSectionStyle } from "./TimelineWorkbookStyles";

type TimelineMentionsPanelProps = {
  readonly sourceRecordId: string | null;
  readonly registerCollectionItem: TimelineInspectorElementRegistry["registerCollectionItem"];
  readonly entityIndex: Record<string, { label: string }>;
  readonly getRelationshipLabel: (
    fieldKey: InspectorMention["fieldKey"],
  ) => string;
  readonly inspectorMentions: readonly InspectorMention[];
  readonly relationshipEditors?: Readonly<
    Record<CollectionFieldKey, ReactNode>
  >;
  readonly feedback?: WorkbookInspectorFeedback | null;
  readonly registerMention: (
    sourceRecordId: string,
    itemRef: string,
    element: HTMLButtonElement | null,
  ) => void;
  readonly onSelectMention: (rowRecordId: string, itemRef: string) => void;
  readonly selectedMention: InspectorMention | null;
  readonly actions: TimelineMentionActions;
};
export function TimelineMentionsPanel({
  sourceRecordId,
  entityIndex,
  inspectorMentions,
  relationshipEditors,
  registerMention,
  registerCollectionItem,
  onSelectMention,
  selectedMention,
  actions,
  getRelationshipLabel,
  feedback,
}: TimelineMentionsPanelProps) {
  const selectedDetails = selectedMention ? (
    <section
      style={inspectorSectionStyle}
      aria-label={`Selected ${getRelationshipLabel(selectedMention.fieldKey)} item`}
    >
      <details>
        <summary>Mention details</summary>
        <WorkbookRelationshipChipDetails
          presentation={timelineRelationshipChipPresentation({
            entityIndex,
            item: selectedMention,
            selected: true,
          })}
        />
      </details>
      <TimelineMentionActionControls
        key={`${selectedMention.rowRecordId}:${selectedMention.fieldKey}:${selectedMention.itemRef}`}
        actions={actions}
      />
    </section>
  ) : null;
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
      <Fragment key={`${item.rowRecordId}:${item.fieldKey}:${item.itemRef}`}>
        {item.status === "dismissed" && items[0]?.itemRef === item.itemRef ? (
          <div>
            <p style={groupLabelStyle}>Dismissed in this session</p>
            <p>Observed here; use History for durable changes.</p>
          </div>
        ) : null}
        <button
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
                buttons.current
                  .get(next.itemRef)
                  ?.focus({ preventScroll: true });
              event.preventDefault();
            }
            if (event.key !== "Tab" && event.key !== "Escape")
              event.stopPropagation();
          }}
        >
          {presentation.rawText !== presentation.label ? (
            <span style={mentionRawStyle}>{presentation.rawText}</span>
          ) : null}
          <span style={mentionSummaryStyle}>
            <WorkbookRelationshipChip decorative presentation={presentation} />
            <span style={workbookTypography("metadata")}>
              {presentation.state === "auto_resolved"
                ? "Automatically resolved"
                : presentation.state === "resolved"
                  ? "Resolved"
                  : presentation.state === "dismissed"
                    ? "Dismissed"
                    : "Unresolved"}
            </span>
          </span>
        </button>
        {selectedMention?.rowRecordId === item.rowRecordId &&
        selectedMention.fieldKey === item.fieldKey &&
        selectedMention.itemRef === item.itemRef
          ? selectedDetails
          : null}
        {feedback?.destination?.kind === "relationship_item" &&
        feedback.destination.fieldKey === item.fieldKey &&
        feedback.destination.itemRef === item.itemRef &&
        feedback.sourceRecordId === item.rowRecordId ? (
          <WorkbookInspectorFeedbackView feedback={feedback} />
        ) : null}
      </Fragment>
    );
  };
  return (
    <div
      data-testid={timelineInspectorSectionTestId("relationships")}
      style={inspectorSectionStyle}
    >
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
                {active.length === 0 ? <span>No items</span> : null}
                {[...active, ...dismissed].map((item) =>
                  renderMention(
                    item,
                    item.status === "dismissed" ? dismissed : active,
                  ),
                )}
              </div>
              {relationshipEditors?.[fieldKey]}
            </section>
          );
        },
      )}
      {relationshipEditors?.["timeline.tags"]}
    </div>
  );
}

const mentionGroupColumnStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;

const groupLabelStyle = {
  ...workbookTypography("section-heading"),
  margin: 0,
  color: "var(--ct-colors-ink-subtle)",
} satisfies CSSProperties;

const mentionListButtonStyle = {
  ...workbookTypography("ui"),
  display: "grid",
  gap: "var(--ct-spacing-xxs)",
  border: "none",
  background: "transparent",
  color: "var(--ct-colors-ink)",
  padding: "var(--ct-spacing-xs)",
  textAlign: "left",
  cursor: "pointer",
} satisfies CSSProperties;

const mentionRawStyle = {
  display: "-webkit-box",
  WebkitBoxOrient: "vertical",
  WebkitLineClamp: 2,
  overflow: "hidden",
  overflowWrap: "anywhere",
  whiteSpace: "pre-wrap",
} satisfies CSSProperties;
const mentionSummaryStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-xs)",
  alignItems: "baseline",
  minWidth: 0,
} satisfies CSSProperties;

const mentionListButtonSelectedStyle = {
  boxShadow: "0 0 0 2px var(--ct-colors-accent)",
  outline: "2px solid transparent",
  outlineOffset: "2px",
} satisfies CSSProperties;
