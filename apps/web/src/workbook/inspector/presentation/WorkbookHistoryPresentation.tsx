import type { CSSProperties, ReactNode } from "react";
import { WorkbookInspectorTechnicalDetails } from "./WorkbookInspectorPresentation";
import type { WorkbookHistoryEventPresentation } from "./workbookInspectorPresentationModel";

export function WorkbookHistoryList({
  children,
}: {
  readonly children: ReactNode;
}) {
  return <ol style={listStyle}>{children}</ol>;
}

export function WorkbookHistoryEvent({
  actions,
  event,
  testId,
}: {
  readonly actions?: ReactNode | undefined;
  readonly event: WorkbookHistoryEventPresentation;
  readonly testId?: string | undefined;
}) {
  return (
    <li data-testid={testId} style={eventStyle}>
      <div style={headerStyle}>
        <strong>{event.summary || event.operation}</strong>
        <time dateTime={event.committedAt}>
          {formatHistoryTimestamp(event.committedAt)}
        </time>
      </div>
      {event.summary && event.operation !== event.summary ? (
        <p style={operationStyle}>{event.operation}</p>
      ) : null}
      {event.actorLabel ? (
        <p style={actorStyle}>Changed by {event.actorLabel}</p>
      ) : null}
      <WorkbookInspectorTechnicalDetails fields={event.technicalFields} />
      {actions}
    </li>
  );
}

function formatHistoryTimestamp(value: string): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toISOString();
}

const listStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  margin: 0,
  paddingInlineStart: "var(--ct-spacing-lg)",
} satisfies CSSProperties;
const eventStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  paddingBlock: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const headerStyle = {
  display: "flex",
  alignItems: "baseline",
  justifyContent: "space-between",
  flexWrap: "wrap" as const,
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const operationStyle = { margin: 0 } satisfies CSSProperties;
const actorStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  fontSize: "var(--ct-typography-metadata-fontSize)",
} satisfies CSSProperties;
