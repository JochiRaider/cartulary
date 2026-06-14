import type { CSSProperties, ReactNode } from "react";
import {
  type MentionEntityOption,
  TimelineMentionsPanel,
} from "./TimelineMentionsPanel";
import type { InspectorMention } from "./workbookMentionChips";
import type { MentionResolutionAction } from "./workbookShellPhase4";
import type { WorkbookRow } from "./workbookTimelineModel";

export function TimelineWorkbookInspector({
  canManageMentions,
  currentHistoryDeleted,
  draftRow,
  entityIndex,
  getRelationshipLabel,
  hostEntities,
  identityEntities,
  inspectorMessage,
  inspectorMentions,
  onResolveTargetChange,
  onSelectMention,
  onSetInspectorMessage,
  onSubmitMentionAction,
  renderEvidenceAttachSection,
  renderInspectorFieldEditors,
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
  readonly inspectorMessage: string | null;
  readonly inspectorMentions: readonly InspectorMention[];
  readonly onResolveTargetChange: (value: string) => void;
  readonly onSelectMention: (rowRecordId: string, itemRef: string) => void;
  readonly onSetInspectorMessage: (message: string) => void;
  readonly onSubmitMentionAction: (
    mention: InspectorMention,
    action: MentionResolutionAction,
    resolvedRecordId?: string,
  ) => void;
  readonly renderEvidenceAttachSection: (row: WorkbookRow) => ReactNode;
  readonly renderInspectorFieldEditors: (row: WorkbookRow) => ReactNode;
  readonly renderRowHistorySection: () => ReactNode;
  readonly rowHistoryRecordId: string | null;
  readonly selectedMention: InspectorMention | null;
  readonly selectedResolveTargetId: string;
  readonly selectedRow: WorkbookRow | null;
}) {
  return (
    <aside
      aria-label="Timeline inspector"
      data-testid="timeline-inspector"
      style={inspectorShellStyle}
    >
      <div style={inspectorHeaderStyle}>
        <p style={eyebrowStyle}>Inspector</p>
        <h2 style={inspectorTitleStyle}>
          {selectedRow?.recordId
            ? `Timeline row ${selectedRow.recordId}`
            : currentHistoryDeleted && rowHistoryRecordId
              ? `Deleted row ${rowHistoryRecordId}`
              : draftRow
                ? "Draft timeline row"
                : "Select a saved row"}
        </h2>
      </div>
      {selectedRow?.recordId ? (
        <>
          {renderInspectorFieldEditors(selectedRow)}
          {renderEvidenceAttachSection(selectedRow)}
          {renderRowHistorySection()}
          <TimelineMentionsPanel
            canManageMentions={canManageMentions}
            entityIndex={entityIndex}
            getRelationshipLabel={getRelationshipLabel}
            hostEntities={hostEntities}
            identityEntities={identityEntities}
            inspectorMentions={inspectorMentions}
            onResolveTargetChange={onResolveTargetChange}
            onSelectMention={onSelectMention}
            onSetInspectorMessage={onSetInspectorMessage}
            onSubmitMentionAction={onSubmitMentionAction}
            selectedMention={selectedMention}
            selectedResolveTargetId={selectedResolveTargetId}
          />
          <InspectorMessage message={inspectorMessage} />
        </>
      ) : currentHistoryDeleted && rowHistoryRecordId !== null ? (
        <>
          {renderRowHistorySection()}
          <InspectorMessage message={inspectorMessage} />
        </>
      ) : (
        <>
          {draftRow ? renderInspectorFieldEditors(draftRow) : null}
          {draftRow ? renderEvidenceAttachSection(draftRow) : null}
          <p style={bodyStyle}>
            Pick a saved row to inspect unresolved, resolved, and dismissed
            mentions.
          </p>
          <InspectorMessage message={inspectorMessage} />
        </>
      )}
    </aside>
  );
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
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-component-inspector-border)",
  background: "var(--ct-component-inspector-backgroundColor)",
  color: "var(--ct-component-inspector-textColor)",
  padding: "var(--ct-spacing-panel-padding)",
  position: "sticky",
  top: "calc(var(--ct-layout-topBarHeight) + var(--ct-spacing-sm))",
  maxHeight:
    "calc(100vh - var(--ct-layout-topBarHeight) - var(--ct-layout-statusStripHeight) - 16px)",
  overflow: "auto",
} satisfies CSSProperties;

const inspectorHeaderStyle = {
  display: "grid",
  gap: "0.35rem",
  marginBottom: "1rem",
} satisfies CSSProperties;

const inspectorTitleStyle = {
  margin: 0,
  fontSize: "1.25rem",
} satisfies CSSProperties;
