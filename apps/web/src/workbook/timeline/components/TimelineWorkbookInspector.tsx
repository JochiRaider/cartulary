import { workbookInspectorCloseButtonTestId } from "@cartulary/ui-contracts";
import type { InspectorConfig } from "@cartulary/view-contracts";
import { X } from "lucide-react";
import type { CSSProperties, ReactNode } from "react";
import {
  inspectorNoRowState,
  inspectorPanelIsDeclared,
} from "../../models/workbookInspectorModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { InspectorMention } from "../models/workbookMentionChips";
import type { WorkbookRow } from "../models/workbookTimelineModel";
import type { MentionResolutionAction } from "../services/workbookShellPhase4";
import {
  type MentionEntityOption,
  TimelineMentionsPanel,
} from "./TimelineMentionsPanel";

export function TimelineWorkbookInspector({
  canManageMentions,
  currentHistoryDeleted,
  draftRow,
  entityIndex,
  getRelationshipLabel,
  hostEntities,
  identityEntities,
  inspectorConfig,
  inspectorMessage,
  inspectorMentions,
  onResolveTargetChange,
  onSelectMention,
  onSetInspectorMessage,
  onClose,
  onSubmitMentionAction,
  renderEvidenceAttachSection,
  renderInspectorFieldEditors,
  renderRelationshipEditors,
  renderRowHistorySection,
  rowHistoryRecordId,
  selectedMention,
  selectedResolveTargetId,
  selectedRow,
}: {
  readonly canManageMentions: boolean;
  readonly currentHistoryDeleted: boolean;
  readonly draftRow: WorkbookRow | null;
  readonly entityIndex: Record<string, { label: string }>;
  readonly getRelationshipLabel: (
    fieldKey: InspectorMention["fieldKey"],
  ) => string;
  readonly hostEntities: readonly MentionEntityOption[];
  readonly identityEntities: readonly MentionEntityOption[];
  readonly inspectorConfig: InspectorConfig;
  readonly inspectorMessage: string | null;
  readonly inspectorMentions: readonly InspectorMention[];
  readonly onResolveTargetChange: (value: string) => void;
  readonly onSelectMention: (rowRecordId: string, itemRef: string) => void;
  readonly onSetInspectorMessage: (message: string) => void;
  readonly onClose: () => void;
  readonly onSubmitMentionAction: (
    mention: InspectorMention,
    action: MentionResolutionAction,
    resolvedRecordId?: string,
  ) => void;
  readonly renderEvidenceAttachSection: (row: WorkbookRow) => ReactNode;
  readonly renderInspectorFieldEditors: (row: WorkbookRow) => ReactNode;
  readonly renderRelationshipEditors: (row: WorkbookRow) => ReactNode;
  readonly renderRowHistorySection: () => ReactNode;
  readonly rowHistoryRecordId: string | null;
  readonly selectedMention: InspectorMention | null;
  readonly selectedResolveTargetId: string;
  readonly selectedRow: WorkbookRow | null;
}) {
  const showDetailsPanel = inspectorPanelIsDeclared(inspectorConfig, "details");
  const showEvidencePanel = inspectorPanelIsDeclared(
    inspectorConfig,
    "evidence",
  );
  const showHistoryPanel = inspectorPanelIsDeclared(inspectorConfig, "history");
  const showRelationshipsPanel = inspectorPanelIsDeclared(
    inspectorConfig,
    "relationships",
  );
  return (
    <aside
      aria-label="Timeline inspector"
      data-testid="timeline-inspector"
      style={inspectorShellStyle}
    >
      <div style={inspectorHeaderStyle}>
        <div style={inspectorTopRowStyle}>
          <p style={eyebrowStyle}>Inspector</p>
          <button
            aria-label="Close inspector"
            data-testid={workbookInspectorCloseButtonTestId(
              timelineViewSchemaId,
            )}
            style={closeButtonStyle}
            type="button"
            onClick={onClose}
          >
            <X aria-hidden="true" size={16} />
          </button>
        </div>
        <h2 style={inspectorTitleStyle}>
          {inspectorTitle(selectedRow, draftRow, currentHistoryDeleted)}
        </h2>
      </div>
      {selectedRow?.recordId ? (
        <>
          {showDetailsPanel ? renderInspectorFieldEditors(selectedRow) : null}
          {showEvidencePanel ? renderEvidenceAttachSection(selectedRow) : null}
          {showHistoryPanel ? renderRowHistorySection() : null}
          {showRelationshipsPanel ? (
            <TimelineMentionsPanel
              canManageMentions={canManageMentions}
              entityIndex={entityIndex}
              getRelationshipLabel={getRelationshipLabel}
              hostEntities={hostEntities}
              identityEntities={identityEntities}
              inspectorMentions={inspectorMentions}
              relationshipEditors={renderRelationshipEditors(selectedRow)}
              onResolveTargetChange={onResolveTargetChange}
              onSelectMention={onSelectMention}
              onSetInspectorMessage={onSetInspectorMessage}
              onSubmitMentionAction={onSubmitMentionAction}
              selectedMention={selectedMention}
              selectedResolveTargetId={selectedResolveTargetId}
            />
          ) : null}
          <InspectorMessage message={inspectorMessage} />
        </>
      ) : currentHistoryDeleted && rowHistoryRecordId !== null ? (
        <>
          {showHistoryPanel ? renderRowHistorySection() : null}
          <InspectorMessage message={inspectorMessage} />
        </>
      ) : (
        <>
          {draftRow && showDetailsPanel
            ? renderInspectorFieldEditors(draftRow)
            : null}
          {draftRow && showEvidencePanel
            ? renderEvidenceAttachSection(draftRow)
            : null}
          <p style={bodyStyle}>{inspectorNoRowState(inspectorConfig)}</p>
          <InspectorMessage message={inspectorMessage} />
        </>
      )}
    </aside>
  );
}

function inspectorTitle(
  selectedRow: WorkbookRow | null,
  draftRow: WorkbookRow | null,
  currentHistoryDeleted: boolean,
) {
  if (selectedRow?.recordId) {
    return (
      selectedRow.values.activitySynopsisText.trim() || "Selected timeline row"
    );
  }
  if (currentHistoryDeleted) {
    return "Deleted timeline row";
  }
  if (draftRow) {
    return "Draft timeline row";
  }
  return "no_row_selected";
}

function InspectorMessage({ message }: { readonly message: string | null }) {
  return message ? (
    <p data-testid="timeline-inspector-message" style={bodyStyle}>
      {message}
    </p>
  ) : null;
}

const eyebrowStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase",
  color: "var(--ct-colors-accent)",
} satisfies CSSProperties;

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;

const inspectorShellStyle = {
  boxSizing: "border-box",
  blockSize: "100%",
  maxBlockSize: "100%",
  borderRadius: "var(--ct-rounded-sm) 0 0 var(--ct-rounded-sm)",
  border: "var(--ct-component-inspector-border)",
  borderInlineEnd: 0,
  background: "var(--ct-component-inspector-backgroundColor)",
  color: "var(--ct-component-inspector-textColor)",
  padding: "var(--ct-spacing-panel-padding)",
  overflow: "auto",
} satisfies CSSProperties;

const inspectorHeaderStyle = {
  display: "grid",
  gap: "0.35rem",
  marginBottom: "1rem",
} satisfies CSSProperties;

const inspectorTopRowStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "0.5rem",
} satisfies CSSProperties;

const inspectorTitleStyle = {
  margin: 0,
  fontSize: "1.25rem",
} satisfies CSSProperties;

const closeButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  inlineSize: "1.8rem",
  blockSize: "1.8rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
  cursor: "pointer",
} satisfies CSSProperties;
