import {
  mentionCreateEntityButtonTestId,
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  mentionRestoreUnresolvedButtonTestId,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties, ReactNode } from "react";
import type { InspectorMention } from "../models/workbookMentionChips";
import type { MentionResolutionAction } from "../services/workbookCollaborationMessages";
import { RelationshipChip, relationshipItemLabel } from "./TimelineCellEditors";

export type MentionEntityOption = {
  readonly label: string;
  readonly recordId: string;
};

type MentionStatus = InspectorMention["status"];

const mentionStatuses = [
  "unresolved",
  "resolved",
  "dismissed",
] as const satisfies readonly MentionStatus[];

export type TimelineMentionsPanelProps = {
  readonly canManageMentions: boolean;
  readonly entityIndex: Record<string, { label: string }>;
  readonly getRelationshipLabel: (
    fieldKey: InspectorMention["fieldKey"],
  ) => string;
  readonly hostEntities: readonly MentionEntityOption[];
  readonly identityEntities: readonly MentionEntityOption[];
  readonly inspectorMentions: readonly InspectorMention[];
  readonly relationshipEditors?: ReactNode;
  readonly onResolveTargetChange: (value: string) => void;
  readonly onSelectMention: (rowRecordId: string, itemRef: string) => void;
  readonly onSetInspectorMessage: (message: string) => void;
  readonly onSubmitMentionAction: (
    mention: InspectorMention,
    action: MentionResolutionAction,
    resolvedRecordId?: string,
  ) => void;
  readonly selectedMention: InspectorMention | null;
  readonly selectedResolveTargetId: string;
};

export function TimelineMentionsPanel({
  canManageMentions,
  entityIndex,
  getRelationshipLabel,
  hostEntities,
  identityEntities,
  inspectorMentions,
  relationshipEditors,
  onResolveTargetChange,
  onSelectMention,
  onSetInspectorMessage,
  onSubmitMentionAction,
  selectedMention,
  selectedResolveTargetId,
}: TimelineMentionsPanelProps) {
  return (
    <>
      <MentionGroups
        entityIndex={entityIndex}
        inspectorMentions={inspectorMentions}
        relationshipEditors={relationshipEditors}
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
          onSubmitMentionAction={onSubmitMentionAction}
          selectedMention={selectedMention}
          selectedResolveTargetId={selectedResolveTargetId}
        />
      ) : null}
    </>
  );
}

function MentionGroups({
  entityIndex,
  inspectorMentions,
  relationshipEditors,
  onSelectMention,
  selectedMention,
}: {
  readonly entityIndex: Record<string, { label: string }>;
  readonly inspectorMentions: readonly InspectorMention[];
  readonly relationshipEditors?: ReactNode;
  readonly onSelectMention: (rowRecordId: string, itemRef: string) => void;
  readonly selectedMention: InspectorMention | null;
}) {
  return (
    <section
      data-testid={timelineInspectorSectionTestId("relationships")}
      style={inspectorSectionStyle}
    >
      <h3 style={sectionTitleStyle}>Relationships</h3>
      {relationshipEditors}
      <div style={mentionGroupStyle}>
        {mentionStatuses.map((status) => {
          const group = inspectorMentions.filter(
            (item) => item.status === status,
          );
          return (
            <div key={status} style={mentionGroupColumnStyle}>
              <p data-density-role="narrow-metadata" style={groupLabelStyle}>
                {statusLabel(status)}
              </p>
              {group.length > 0 ? (
                group.map((item) => (
                  <button
                    key={item.itemRef}
                    data-testid={mentionItemTestId(item.itemRef)}
                    tabIndex={0}
                    style={{
                      ...mentionListButtonStyle,
                      ...(selectedMention?.itemRef === item.itemRef
                        ? mentionListButtonSelectedStyle
                        : null),
                    }}
                    type="button"
                    onClick={() => {
                      onSelectMention(item.rowRecordId, item.itemRef);
                    }}
                  >
                    <RelationshipChip
                      entityIndex={entityIndex}
                      item={item}
                      selected={selectedMention?.itemRef === item.itemRef}
                    />
                  </button>
                ))
              ) : (
                <span style={emptyRelationshipStyle}>None</span>
              )}
            </div>
          );
        })}
      </div>
    </section>
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
  readonly onSetInspectorMessage: (message: string) => void;
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
      <dl style={detailListStyle}>
        <div>
          <dt style={detailTermStyle}>Raw token</dt>
          <dd style={detailValueStyle}>{selectedMention.rawText}</dd>
        </div>
        <div>
          <dt style={detailTermStyle}>Field</dt>
          <dd style={detailValueStyle}>
            {getRelationshipLabel(selectedMention.fieldKey)}
          </dd>
        </div>
        <div>
          <dt style={detailTermStyle}>Status</dt>
          <dd style={detailValueStyle}>{selectedMention.status}</dd>
        </div>
        <div>
          <dt style={detailTermStyle}>Target</dt>
          <dd style={detailValueStyle}>
            {selectedMention.resolvedRecordId
              ? relationshipItemLabel(selectedMention, entityIndex)
              : "None"}
          </dd>
        </div>
      </dl>

      {selectedMention.status === "unresolved" ? (
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
                  onSetInspectorMessage("Select a target first.");
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
                onSubmitMentionAction(selectedMention, "resolve_item");
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
                    onSetInspectorMessage("Select a target first.");
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

      {selectedMention.status === "dismissed" ? (
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

function statusLabel(status: MentionStatus) {
  return status === "dismissed"
    ? "Dismissed"
    : status === "resolved"
      ? "Resolved"
      : "Unresolved";
}

const inputStyle = {
  boxSizing: "border-box",
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
} satisfies CSSProperties;

const selectStyle = {
  ...inputStyle,
  appearance: "auto",
} satisfies CSSProperties;

const secondaryActionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
} satisfies CSSProperties;

const labelStyle = {
  display: "grid",
  gap: "0.4rem",
  fontSize: "0.95rem",
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;

const inspectorSectionStyle = {
  display: "grid",
  gap: "0.75rem",
  marginBottom: "1rem",
} satisfies CSSProperties;

const sectionTitleStyle = {
  margin: 0,
  fontSize: "1rem",
} satisfies CSSProperties;

const emptyRelationshipStyle = {
  color: "var(--ct-colors-ink-tertiary)",
  fontSize: "0.9rem",
} satisfies CSSProperties;

const mentionGroupStyle = {
  display: "grid",
  gap: "0.75rem",
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

const detailListStyle = {
  display: "grid",
  gap: "0.75rem",
  margin: 0,
} satisfies CSSProperties;

const detailTermStyle = {
  fontSize: "0.75rem",
  letterSpacing: "0.08em",
  textTransform: "uppercase",
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;

const detailValueStyle = {
  margin: "0.2rem 0 0",
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
