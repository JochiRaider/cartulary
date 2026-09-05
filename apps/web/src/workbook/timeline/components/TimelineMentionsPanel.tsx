import {
  mentionCreateEntityButtonTestId,
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  mentionRestoreUnresolvedButtonTestId,
  relationshipItemsTestId,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import { type CSSProperties, type ReactNode, useRef } from "react";
import type { MentionResolutionAction } from "../../collaboration/workbookCollaborationMessages";
import {
  WorkbookRelationshipChip,
  WorkbookRelationshipChipDetails,
} from "../../components/WorkbookRelationshipChip";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorMessageFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import { relationshipChipAccessibleName } from "../../models/workbookRelationshipChip";
import type { TimelineInspectorElementRegistry } from "../focus/timelineInspectorElementRegistry";
import {
  type InspectorMention,
  timelineRelationshipChipPresentation,
} from "../models/workbookMentionChips";
import {
  inputStyle,
  inspectorSectionStyle,
  labelStyle,
  secondaryActionButtonStyle,
  sectionTitleStyle,
} from "./TimelineWorkbookStyles";

export type MentionEntityOption = {
  readonly label: string;
  readonly recordId: string;
};

type TimelineMentionsPanelProps = {
  readonly sourceRecordId: string | null;
  readonly canManageMentions: boolean;
  readonly registerCollectionItem: TimelineInspectorElementRegistry["registerCollectionItem"];
  readonly entityIndex: Record<string, { label: string }>;
  readonly getRelationshipLabel: (
    fieldKey: InspectorMention["fieldKey"],
  ) => string;
  readonly hostEntities: readonly MentionEntityOption[];
  readonly identityEntities: readonly MentionEntityOption[];
  readonly inspectorMentions: readonly InspectorMention[];
  readonly relationshipEditors?: ReactNode;
  readonly registerMention: (
    sourceRecordId: string,
    itemRef: string,
    element: HTMLButtonElement | null,
  ) => void;
  readonly onResolveTargetChange: (value: string) => void;
  readonly onSelectMention: (rowRecordId: string, itemRef: string) => void;
  readonly onSetInspectorMessage: (message: WorkbookInspectorFeedback) => void;
  readonly onCreateEntityFromMention: (mention: InspectorMention) => void;
  readonly onSubmitMentionAction: (
    mention: InspectorMention,
    action: MentionResolutionAction,
    resolvedRecordId?: string,
  ) => void;
  readonly selectedMention: InspectorMention | null;
  readonly selectedResolveTargetId: string;
};

export function TimelineMentionsPanel({
  sourceRecordId,
  canManageMentions,
  registerCollectionItem,
  entityIndex,
  getRelationshipLabel,
  hostEntities,
  identityEntities,
  inspectorMentions,
  relationshipEditors,
  registerMention,
  onResolveTargetChange,
  onSelectMention,
  onSetInspectorMessage,
  onCreateEntityFromMention,
  onSubmitMentionAction,
  selectedMention,
  selectedResolveTargetId,
}: TimelineMentionsPanelProps) {
  return (
    <>
      <MentionGroups
        sourceRecordId={sourceRecordId}
        registerCollectionItem={registerCollectionItem}
        entityIndex={entityIndex}
        inspectorMentions={inspectorMentions}
        relationshipEditors={relationshipEditors}
        registerMention={registerMention}
        onSelectMention={onSelectMention}
        selectedMention={selectedMention}
      />
      {selectedMention ? (
        <SelectedMentionSection
          canManageMentions={canManageMentions}
          entityIndex={entityIndex}
          getRelationshipLabel={getRelationshipLabel}
          hostEntities={hostEntities}
          identityEntities={identityEntities}
          onResolveTargetChange={onResolveTargetChange}
          onSetInspectorMessage={onSetInspectorMessage}
          onCreateEntityFromMention={onCreateEntityFromMention}
          onSubmitMentionAction={onSubmitMentionAction}
          selectedMention={selectedMention}
          selectedResolveTargetId={selectedResolveTargetId}
        />
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
                  <p style={groupLabelStyle}>Dismissed</p>
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

function SelectedMentionSection({
  canManageMentions,
  entityIndex,
  getRelationshipLabel,
  hostEntities,
  identityEntities,
  onResolveTargetChange,
  onSetInspectorMessage,
  onCreateEntityFromMention,
  onSubmitMentionAction,
  selectedMention,
  selectedResolveTargetId,
}: {
  readonly canManageMentions: boolean;
  readonly entityIndex: Record<string, { label: string }>;
  readonly getRelationshipLabel: (
    fieldKey: InspectorMention["fieldKey"],
  ) => string;
  readonly hostEntities: readonly MentionEntityOption[];
  readonly identityEntities: readonly MentionEntityOption[];
  readonly onResolveTargetChange: (value: string) => void;
  readonly onSetInspectorMessage: (message: WorkbookInspectorFeedback) => void;
  readonly onCreateEntityFromMention: (mention: InspectorMention) => void;
  readonly onSubmitMentionAction: (
    mention: InspectorMention,
    action: MentionResolutionAction,
    resolvedRecordId?: string,
  ) => void;
  readonly selectedMention: InspectorMention;
  readonly selectedResolveTargetId: string;
}) {
  return (
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

      {selectedMention.status === "unresolved" && canManageMentions ? (
        <div style={inspectorActionStackStyle}>
          <ResolveTargetSelect
            label="Resolve to existing"
            entities={
              selectedMention.entityType === "host"
                ? hostEntities
                : identityEntities
            }
            onChange={onResolveTargetChange}
            selectedResolveTargetId={selectedResolveTargetId}
          />
          <div style={inlineButtonRowStyle}>
            <button
              data-testid={mentionResolveExistingButtonTestId()}
              style={secondaryActionButtonStyle}
              type="button"
              onClick={() => {
                if (selectedResolveTargetId === "") {
                  onSetInspectorMessage(
                    workbookInspectorMessageFeedback(
                      "Select a target first.",
                      "none",
                    ),
                  );
                  return;
                }
                onSubmitMentionAction(
                  selectedMention,
                  "resolve_item",
                  selectedResolveTargetId,
                );
              }}
            >
              Resolve to existing
            </button>
            <button
              data-testid={mentionCreateEntityButtonTestId(
                selectedMention.entityType,
              )}
              style={secondaryActionButtonStyle}
              type="button"
              onClick={() => {
                onCreateEntityFromMention(selectedMention);
              }}
            >
              {selectedMention.entityType === "host"
                ? "Create host"
                : "Create identity"}
            </button>
          </div>
        </div>
      ) : null}

      {selectedMention.status === "resolved" ? (
        <div style={inspectorActionStackStyle}>
          {canManageMentions ? (
            <ResolveTargetSelect
              label="Correct target"
              entities={
                selectedMention.entityType === "host"
                  ? hostEntities
                  : identityEntities
              }
              onChange={onResolveTargetChange}
              selectedResolveTargetId={selectedResolveTargetId}
            />
          ) : null}
          <div style={inlineButtonRowStyle}>
            {canManageMentions ? (
              <button
                data-testid={mentionResolveExistingButtonTestId()}
                style={secondaryActionButtonStyle}
                type="button"
                onClick={() => {
                  if (selectedResolveTargetId === "") {
                    onSetInspectorMessage(
                      workbookInspectorMessageFeedback(
                        "Select a target first.",
                        "none",
                      ),
                    );
                    return;
                  }
                  onSubmitMentionAction(
                    selectedMention,
                    "resolve_item",
                    selectedResolveTargetId,
                  );
                }}
              >
                Correct target
              </button>
            ) : null}
            {canManageMentions ? (
              <button
                data-testid={mentionDismissButtonTestId()}
                style={secondaryActionButtonStyle}
                type="button"
                onClick={() => {
                  onSubmitMentionAction(selectedMention, "dismiss_item");
                }}
              >
                Dismiss
              </button>
            ) : null}
            {canManageMentions ? (
              <button
                data-testid={mentionRestoreUnresolvedButtonTestId()}
                style={secondaryActionButtonStyle}
                type="button"
                onClick={() => {
                  onSubmitMentionAction(
                    selectedMention,
                    "revert_to_unresolved",
                  );
                }}
              >
                Revert to unresolved
              </button>
            ) : null}
          </div>
        </div>
      ) : null}

      {selectedMention.status === "dismissed" && canManageMentions ? (
        <div style={inlineButtonRowStyle}>
          <button
            data-testid={mentionRestoreUnresolvedButtonTestId()}
            style={secondaryActionButtonStyle}
            type="button"
            onClick={() => {
              onSubmitMentionAction(selectedMention, "revert_to_unresolved");
            }}
          >
            Restore to unresolved
          </button>
        </div>
      ) : null}
    </section>
  );
}

function ResolveTargetSelect({
  entities,
  label,
  onChange,
  selectedResolveTargetId,
}: {
  readonly entities: readonly MentionEntityOption[];
  readonly label: string;
  readonly onChange: (value: string) => void;
  readonly selectedResolveTargetId: string;
}) {
  return (
    <label style={labelStyle}>
      {label}
      <select
        className="cartulary-mention-resolve-select"
        data-testid={mentionResolveTargetSelectTestId()}
        style={selectStyle}
        value={selectedResolveTargetId}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      >
        <option value="">Select target</option>
        {entities.map((entity) => (
          <option key={entity.recordId} value={entity.recordId}>
            {entity.label}
          </option>
        ))}
      </select>
    </label>
  );
}

const selectStyle = {
  ...inputStyle,
  appearance: "auto",
} satisfies CSSProperties;

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

const inlineButtonRowStyle = {
  display: "flex",
  gap: "0.5rem",
  flexWrap: "wrap",
} satisfies CSSProperties;

const inspectorActionStackStyle = {
  display: "grid",
  gap: "0.75rem",
} satisfies CSSProperties;
