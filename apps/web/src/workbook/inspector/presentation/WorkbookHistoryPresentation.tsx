import { type CSSProperties, type ReactNode, useRef } from "react";
import { workbookTypography } from "../../components/workbookFormStyles";
import { WorkbookInspectorTechnicalDetails } from "./WorkbookInspectorFeedback";
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
  highlighted = false,
  onClose,
}: {
  readonly highlighted?: boolean;
  readonly actions?: ReactNode | undefined;
  readonly event: WorkbookHistoryEventPresentation;
  readonly testId?: string | undefined;
  readonly onClose?: () => void;
}) {
  const summary = useRef<HTMLElement>(null);
  return (
    <li data-testid={testId} style={eventStyle}>
      {highlighted ? (
        <strong data-history-requested-change="true" tabIndex={-1}>
          Requested change
        </strong>
      ) : null}
      <strong>{event.summary}</strong>
      <p style={metadataStyle}>
        {event.operation} · Changed by {event.actorLabel}
      </p>
      <time dateTime={event.committedAt} style={metadataStyle}>
        {formatHistoryTimestamp(event.committedAt)}
      </time>
      <details
        onToggle={(event) => {
          if (event.target !== event.currentTarget || event.currentTarget.open)
            return;
          const containedFocus = event.currentTarget.contains(
            document.activeElement,
          );
          onClose?.();
          if (containedFocus) queueMicrotask(() => summary.current?.focus());
        }}
        onKeyDown={(event) => {
          if (
            event.key !== "Escape" ||
            event.defaultPrevented ||
            !event.currentTarget.open
          )
            return;
          event.preventDefault();
          event.stopPropagation();
          const nearest =
            event.target instanceof Element
              ? event.target.closest("details")
              : event.currentTarget;
          if (nearest) {
            nearest.open = false;
            nearest.querySelector<HTMLElement>(":scope > summary")?.focus();
          }
        }}
      >
        <summary ref={summary} style={disclosureStyle}>
          Event details
        </summary>
        <div style={detailStyle}>
          <p style={metadataStyle}>
            Exact committed timestamp:{" "}
            <time dateTime={event.committedAt}>{event.committedAt}</time>
          </p>
          {event.units.map((unit) => (
            <section key={unit.key} style={unitStyle} aria-label={unit.title}>
              <h4 style={unitTitleStyle}>{unit.title}</h4>
              <dl style={changesStyle}>
                {unit.changes.map((change) => (
                  <div key={change.fieldKey} style={changeStyle}>
                    <dt>
                      <strong>{change.label}</strong>
                      <code style={fieldKeyStyle}>{change.fieldKey}</code>
                    </dt>
                    <dd style={valueStyle}>
                      <span style={metadataStyle}>Before: </span>
                      {change.before}
                    </dd>
                    <dd style={valueStyle}>
                      <span style={metadataStyle}>After: </span>
                      {change.after}
                    </dd>
                  </div>
                ))}
              </dl>
              <p style={metadataStyle}>
                Record references: {unit.recordIds.join(", ")}
              </p>
            </section>
          ))}
          <WorkbookInspectorTechnicalDetails fields={event.technicalFields} />
          {actions}
        </div>
      </details>
    </li>
  );
}

function formatHistoryTimestamp(value: string): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? value
    : `${date.toISOString().slice(0, 19).replace("T", " ")} UTC +00:00`;
}

const listStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  margin: 0,
  padding: 0,
  listStyle: "none",
} satisfies CSSProperties;
const eventStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  paddingBlock: "var(--ct-spacing-sm)",
  borderBottom: "var(--ct-border-hairline)",
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const metadataStyle = {
  ...workbookTypography("compact-metadata"),
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;
const disclosureStyle = {
  cursor: "pointer",
  ...workbookTypography("button"),
} satisfies CSSProperties;
const detailStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  paddingBlock: "var(--ct-spacing-sm)",
} satisfies CSSProperties;
const unitStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const unitTitleStyle = {
  ...workbookTypography("section-heading"),
  margin: 0,
} satisfies CSSProperties;
const changesStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  margin: 0,
} satisfies CSSProperties;
const changeStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const valueStyle = {
  ...workbookTypography("ui"),
  margin: 0,
  whiteSpace: "pre-wrap",
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const fieldKeyStyle = {
  ...workbookTypography("mono"),
  display: "block",
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;
