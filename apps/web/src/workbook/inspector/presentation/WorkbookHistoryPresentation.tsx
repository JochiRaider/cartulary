import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  type ReactNode,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import { workbookTypography } from "../../components/workbookFormStyles";
import { WorkbookInspectorTechnicalDetails } from "./WorkbookInspectorFeedback";
import type {
  WorkbookHistoryEventPresentation,
  WorkbookHistoryValue,
} from "./workbookInspectorPresentationModel";

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
              {event.summaryUnitKey === unit.key ? null : (
                <h4 style={unitTitleStyle}>{unit.title}</h4>
              )}
              <dl style={changesStyle}>
                {unit.changes.map((change) => (
                  <HistoryChange key={change.fieldKey} change={change} />
                ))}
              </dl>
            </section>
          ))}
          <WorkbookInspectorTechnicalDetails
            fields={[
              ...event.technicalFields,
              ...event.units.flatMap((unit) => [
                { label: `${unit.title}: unit reference`, value: unit.key },
                {
                  label: `${unit.title}: record references`,
                  value: unit.recordIds.join(", "),
                },
                ...unit.changes.map((change) => ({
                  label: `Field: ${change.label}`,
                  value: change.fieldKey,
                })),
              ]),
            ]}
          />
          {actions}
        </div>
      </details>
    </li>
  );
}

type Change =
  WorkbookHistoryEventPresentation["units"][number]["changes"][number];
function HistoryChange({ change }: { readonly change: Change }) {
  const root = useRef<HTMLDivElement>(null);
  const [paired, setPaired] = useState(false);
  const scalar =
    change.layout === "scalar" &&
    [change.before, change.after].every(
      (value) => value.state !== "present" || !Array.isArray(value.value),
    );
  // biome-ignore lint/correctness/useExhaustiveDependencies: Historical value changes require remeasuring the rendered DOM.
  useLayoutEffect(() => {
    const element = root.current;
    if (!element || !scalar) {
      setPaired(false);
      return;
    }
    const measure = () => {
      const width = element.clientWidth;
      const gap = Number.parseFloat(getComputedStyle(element).columnGap) || 0;
      const available = (width - gap) / 2;
      const values = [
        ...element.querySelectorAll<HTMLDivElement>(":scope > dd"),
      ];
      const fits =
        available > 0 &&
        values.length === 2 &&
        values.every((value) => {
          const probe = value.cloneNode(true) as HTMLElement;
          Object.assign(probe.style, {
            position: "absolute",
            visibility: "hidden",
            pointerEvents: "none",
            inlineSize: `${available}px`,
          });
          probe.setAttribute("aria-hidden", "true");
          element.append(probe);
          const line = Number.parseFloat(getComputedStyle(probe).lineHeight);
          const fit =
            Number.isFinite(line) &&
            probe.scrollHeight > 0 &&
            probe.scrollHeight <=
              line *
                cartularyDesignPresentation.inspector
                  .historyScalarPairMaxLines +
                1;
          probe.remove();
          return fit;
        });
      setPaired(fits);
    };
    measure();
    const resize =
      typeof ResizeObserver === "undefined"
        ? null
        : new ResizeObserver(measure);
    resize?.observe(element);
    const spacing = new MutationObserver(measure);
    let parent: HTMLElement | null = element.parentElement;
    while (parent) {
      spacing.observe(parent, {
        attributes: true,
        attributeFilter: ["style", "class", "open"],
      });
      parent = parent.parentElement;
    }
    spacing.observe(document.head, {
      subtree: true,
      childList: true,
      characterData: true,
    });
    document.fonts?.addEventListener("loadingdone", measure);
    window.addEventListener("resize", measure);
    return () => {
      resize?.disconnect();
      spacing.disconnect();
      document.fonts?.removeEventListener("loadingdone", measure);
      window.removeEventListener("resize", measure);
    };
  }, [change, scalar]);
  return (
    <div
      ref={root}
      data-history-comparison={paired && scalar ? "paired" : "stacked"}
      style={{
        ...changeStyle,
        position: "relative",
        gridTemplateColumns:
          paired && scalar ? "repeat(2, minmax(0, 1fr))" : "minmax(0, 1fr)",
      }}
    >
      <dt style={{ gridColumn: "1 / -1" }}>
        <strong>{change.label}</strong>
      </dt>
      <dd style={valueStyle}>
        <span>Before: </span>
        <HistoryValue value={change.before} />
      </dd>
      <dd style={valueStyle}>
        <span>After: </span>
        <HistoryValue value={change.after} />
      </dd>
    </div>
  );
}

function HistoryValue({ value }: { readonly value: WorkbookHistoryValue }) {
  if (value.state === "absent")
    return <span data-history-value="absent">Not present</span>;
  if (value.state === "null")
    return <span data-history-value="null">No value (null)</span>;
  const content = value.value;
  if (typeof content === "string") return <HistoryText value={content} />;
  if (typeof content === "number" || typeof content === "boolean")
    return (
      <span data-history-value="scalar">
        {typeof content === "boolean"
          ? content
            ? "True"
            : "False"
          : String(content)}
      </span>
    );
  return content.length === 0 ? (
    <span data-history-value="collection">No items</span>
  ) : (
    <ul
      data-history-value="collection"
      style={{ margin: 0, paddingInlineStart: "var(--ct-spacing-lg)" }}
    >
      {content.map((item, index) => (
        // biome-ignore lint/suspicious/noArrayIndexKey: Immutable historical collections have no item IDs and may contain duplicate values.
        <li key={`${index}:${item}`}>
          <HistoryText value={item} />
        </li>
      ))}
    </ul>
  );
}
function HistoryText({ value }: { readonly value: string }) {
  if (value === "")
    return <span data-history-value="empty-text">Empty text</span>;
  return (
    <span data-history-value="text">
      {value.trim() === "" ? <small>Whitespace only: </small> : null}
      <span style={{ whiteSpace: "pre-wrap", overflowWrap: "anywhere" }}>
        {value}
      </span>
    </span>
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
