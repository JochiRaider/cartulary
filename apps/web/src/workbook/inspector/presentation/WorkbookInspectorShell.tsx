import {
  workbookInspectorCloseButtonTestId,
  workbookInspectorPanelTestId,
} from "@cartulary/ui-contracts";
import type {
  InspectorConfig,
  InspectorPanel,
} from "@cartulary/view-contracts";
import { X } from "lucide-react";
import { type CSSProperties, type ReactNode, useId } from "react";
import { workbookSurfaceInspectorPanelStyle } from "../../layout/WorkbookSurfaceLayout";
import type { WorkbookInspectorSubject } from "../workbookInspectorSubject";
import {
  WorkbookInspectorCompactMetadata,
  WorkbookInspectorTechnicalDetails,
} from "./WorkbookInspectorFeedback";
import {
  type WorkbookInspectorTechnicalField,
  workbookInspectorNoRowMessage,
} from "./workbookInspectorPresentationModel";

export function WorkbookInspectorShell({
  accessibleLabel,
  children,
  config,
  eyebrow = "Inspector",
  heading,
  mode = "record",
  noRowHeading,
  onClose,
  subject,
  testId,
}: {
  readonly accessibleLabel: string;
  readonly children?: ReactNode;
  readonly config: InspectorConfig;
  readonly eyebrow?: string | undefined;
  readonly heading?: string | undefined;
  readonly mode?: "record" | "creation" | undefined;
  readonly noRowHeading: string;
  readonly onClose: () => void;
  readonly subject: WorkbookInspectorSubject | null;
  readonly testId?: string | undefined;
}) {
  const headingId = useId();
  return (
    <aside
      aria-label={accessibleLabel}
      aria-labelledby={headingId}
      data-inspector-state={
        subject === null
          ? mode === "creation"
            ? "creation"
            : "no_row_selected"
          : "ready"
      }
      data-record-id={subject?.recordId}
      data-row-version={subject?.rowVersion}
      data-testid={testId}
      data-view-schema-id={config.viewSchemaId}
      style={shellStyle}
    >
      <header style={headerStyle}>
        <div style={titleRowStyle}>
          <div style={titleStackStyle}>
            <p style={eyebrowStyle}>{eyebrow}</p>
            <h2 id={headingId} style={titleStyle}>
              {heading ?? subject?.label ?? noRowHeading}
            </h2>
          </div>
          <button
            aria-label="Close inspector"
            data-testid={workbookInspectorCloseButtonTestId(
              config.viewSchemaId,
            )}
            style={closeButtonStyle}
            type="button"
            onClick={onClose}
          >
            <X aria-hidden="true" size={16} />
          </button>
        </div>
        {subject === null ? (
          mode === "record" ? (
            <p style={messageStyle}>{workbookInspectorNoRowMessage}</p>
          ) : null
        ) : (
          <RecordContext subject={subject} />
        )}
      </header>
      {children}
      {subject === null ? null : (
        <section aria-label="Record technical metadata" style={metadataStyle}>
          <WorkbookInspectorTechnicalDetails
            fields={subjectTechnicalFields(subject)}
          />
        </section>
      )}
    </aside>
  );
}

export function WorkbookInspectorPanelSection({
  children,
  panel,
  viewSchemaId,
}: {
  readonly children?: ReactNode;
  readonly panel: InspectorPanel;
  readonly viewSchemaId: string;
}) {
  return (
    <section
      data-testid={workbookInspectorPanelTestId(viewSchemaId, panel.panelId)}
      style={panelSectionStyle}
    >
      <h3 style={panelTitleStyle}>{panel.label}</h3>
      {children}
    </section>
  );
}

function RecordContext({
  subject,
}: {
  readonly subject: WorkbookInspectorSubject;
}) {
  return (
    <div style={recordContextStyle}>
      <WorkbookInspectorCompactMetadata>
        <span>{subject.surfaceLabel}</span>
        {subject.stateLabel ? <span>{subject.stateLabel}</span> : null}
      </WorkbookInspectorCompactMetadata>
    </div>
  );
}

function subjectTechnicalFields(
  subject: WorkbookInspectorSubject,
): WorkbookInspectorTechnicalField[] {
  return [
    { label: "Record ID", value: subject.recordId },
    { label: "Row version", value: String(subject.rowVersion) },
  ];
}

const shellStyle = {
  ...workbookSurfaceInspectorPanelStyle,
  display: "grid",
  alignContent: "start",
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;
const headerStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const titleRowStyle = {
  display: "flex",
  alignItems: "start",
  justifyContent: "space-between",
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;
const titleStackStyle = { minWidth: 0 } satisfies CSSProperties;
const eyebrowStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  fontSize: "var(--ct-typography-metadata-fontSize)",
  letterSpacing: "var(--ct-typography-metadata-letterSpacing)",
  textTransform: "uppercase" as const,
} satisfies CSSProperties;
const titleStyle = {
  margin: 0,
  overflowWrap: "anywhere" as const,
} satisfies CSSProperties;
const closeButtonStyle = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "transparent",
  color: "inherit",
  minInlineSize: "var(--ct-density-default-rowHeight)",
  minBlockSize: "var(--ct-density-default-rowHeight)",
} satisfies CSSProperties;
const recordContextStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const metadataStyle = {
  borderBlockStart: "var(--ct-border-hairline)",
  paddingBlockStart: "var(--ct-spacing-sm)",
} satisfies CSSProperties;
const panelSectionStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  paddingBlock: "var(--ct-spacing-sm)",
  borderBlockStart: "var(--ct-border-hairline)",
} satisfies CSSProperties;
const panelTitleStyle = { margin: 0 } satisfies CSSProperties;
const messageStyle = { margin: 0 } satisfies CSSProperties;
