import {
  cartularyDesignPresentation,
  workbookInspectorCloseButtonTestId,
  workbookInspectorPanelTestId,
} from "@cartulary/ui-contracts";
import type {
  InspectorConfig,
  InspectorPanel,
} from "@cartulary/view-contracts";
import { X } from "lucide-react";
import { type CSSProperties, type ReactNode, useId, useRef } from "react";
import { workbookTypography } from "../../components/workbookFormStyles";
import { workbookSurfaceInspectorPanelStyle } from "../../layout/WorkbookSurfaceLayout";
import { useWorkbookInspectorNavigation } from "../../layout/workbookInspectorNavigation";
import type { WorkbookRecordSubject } from "../../ports/WorkbookRecordSubject";
import { WorkbookInspectorActionButton } from "./WorkbookInspectorActions";
import {
  WorkbookInspectorCompactMetadata,
  WorkbookInspectorTechnicalDetails,
} from "./WorkbookInspectorFeedback";
import {
  type WorkbookInspectorAttention,
  type WorkbookInspectorTechnicalField,
  workbookInspectorNoRowMessage,
} from "./workbookInspectorPresentationModel";

export type WorkbookInspectorSection = {
  readonly panel: InspectorPanel;
  readonly content: ReactNode;
  readonly attention?: readonly WorkbookInspectorAttention[];
  readonly focusDestination: (section: HTMLElement) => HTMLElement;
  readonly elementRef?: ((element: HTMLElement | null) => void) | undefined;
};
const noSections: readonly WorkbookInspectorSection[] = [];

type ShellCommon = {
  readonly accessibleLabel: string;
  readonly config: InspectorConfig;
  readonly elementRef?: ((element: HTMLElement | null) => void) | undefined;
  readonly onClose: () => void;
  readonly testId?: string | undefined;
};

type WorkbookInspectorShellProps = ShellCommon &
  (
    | {
        readonly mode: "saved";
        readonly subject: WorkbookRecordSubject;
        readonly sections: readonly WorkbookInspectorSection[];
        readonly feedback?: ReactNode;
        readonly heading?: never;
        readonly context?: never;
      }
    | {
        readonly mode: "empty";
        readonly heading: string;
        readonly subject?: never;
        readonly sections?: never;
        readonly context?: never;
        readonly feedback?: never;
      }
    | {
        readonly mode: "creation";
        readonly context: {
          readonly id: string;
          readonly viewSchemaId: string;
          readonly heading: string;
        };
        readonly sections: readonly WorkbookInspectorSection[];
        readonly subject?: never;
        readonly heading?: never;
        readonly feedback?: never;
      }
  );

export function WorkbookInspectorShell(props: WorkbookInspectorShellProps) {
  const { accessibleLabel, config, elementRef, onClose, testId, mode } = props;
  const subject = mode === "saved" ? props.subject : null;
  const sections = mode === "empty" ? noSections : props.sections;
  const heading =
    mode === "saved"
      ? props.subject.label
      : mode === "creation"
        ? props.context.heading
        : props.heading;
  if (
    mode === "creation" &&
    (!props.context.id || props.context.viewSchemaId !== config.viewSchemaId)
  )
    throw new Error(
      "Inspector creation context must belong to its declared source",
    );
  const headingId = useId();
  const navigationId = useId();
  const attentionRef = useRef<HTMLElement>(null);
  const seenWork = new Set<string>();
  const attention = sections.flatMap((section) =>
    [...(section.attention ?? [])]
      .sort((a, b) => a.order - b.order || a.workId.localeCompare(b.workId))
      .filter((entry) => {
        if (
          !subject ||
          entry.viewSchemaId !== subject.viewSchemaId ||
          entry.recordId !== subject.recordId ||
          !entry.isCurrent() ||
          seenWork.has(entry.workId)
        )
          return false;
        seenWork.add(entry.workId);
        return true;
      })
      .map((entry) => ({ section, entry })),
  );
  const currentAttention = useRef(attention);
  currentAttention.current = attention;
  const admitsAttention = (entry: WorkbookInspectorAttention) =>
    currentAttention.current.some((item) => item.entry === entry) &&
    entry.isCurrent();
  const scope = JSON.stringify([
    config.viewSchemaId,
    subject?.recordId ?? null,
    subject?.kind ?? mode,
    mode === "creation" ? props.context.id : null,
  ]);
  const {
    active,
    menuOpen,
    direct,
    measurementRef,
    onNavigationFocus,
    onNavigationBlur,
    bodyRef,
    navigationRef,
    triggerRef,
    closeRef,
    choose,
    reveal,
    observeScroll,
    toggleMenu,
    dismissMenu,
    remember,
    registerSection,
  } = useWorkbookInspectorNavigation(scope, sections);
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
      ref={elementRef}
      style={shellStyle}
      onKeyDown={(event) => {
        if (
          event.key === "Escape" &&
          !event.defaultPrevented &&
          !event.nativeEvent.isComposing
        ) {
          event.preventDefault();
          event.stopPropagation();
          onClose();
        }
      }}
    >
      <style>{`
        [data-inspector-state] :is(button,input,select,textarea,summary,[tabindex]):focus-visible {
          outline: var(--ct-component-focus-ring-border);
          outline-offset: var(--ct-component-focus-ring-offset);
        }
        [data-inspector-state] :is(button,input,select,textarea):disabled {
          color: var(--ct-colors-ink-subtle) !important;
          background: var(--ct-colors-surface-3) !important;
          cursor: not-allowed;
        }
        [data-inspector-state] button:hover:not(:disabled):not([aria-busy="true"]) {
          border-color: var(--ct-colors-ink-muted);
        }
        [data-inspector-state] button[aria-current="location"] {
          background: var(--ct-colors-surface-3);
          border-color: var(--ct-colors-ink-muted);
          text-decoration: underline;
        }
      `}</style>
      <header style={headerStyle}>
        <div style={titleRowStyle}>
          <div style={titleStackStyle}>
            <h2 id={headingId} style={titleStyle}>
              {heading}
            </h2>
          </div>
          <button
            aria-label="Close inspector"
            ref={closeRef}
            data-testid={workbookInspectorCloseButtonTestId(
              config.viewSchemaId,
            )}
            style={closeButtonStyle}
            type="button"
            onClick={onClose}
          >
            <X aria-hidden="true" size={16} />
            <span>Close</span>
          </button>
        </div>
        {subject === null ? (
          mode === "empty" ? (
            <p style={messageStyle}>{workbookInspectorNoRowMessage}</p>
          ) : null
        ) : null}
        {active ? (
          <fieldset
            ref={navigationRef}
            aria-label="Section navigation"
            style={navigationStyle}
            onFocusCapture={(event) => onNavigationFocus(event.target)}
            onBlurCapture={(event) => onNavigationBlur(event.relatedTarget)}
            onKeyDown={(event) => {
              if (
                event.key === "Escape" &&
                menuOpen &&
                !event.nativeEvent.isComposing &&
                !event.defaultPrevented
              ) {
                event.preventDefault();
                event.stopPropagation();
                dismissMenu();
              }
            }}
          >
            <div
              aria-hidden="true"
              inert
              style={{
                position: "absolute",
                insetInline: 0,
                blockSize: 0,
                overflow: "hidden",
                visibility: "hidden",
                pointerEvents: "none",
              }}
            >
              <div
                data-inspector-navigation-measure
                ref={measurementRef}
                aria-hidden="true"
                inert
                style={measurementStyle}
              >
                {sections.map((section) => (
                  <span
                    key={section.panel.panelId}
                    style={navigationButtonStyle}
                  >
                    {section.panel.label}
                  </span>
                ))}
              </div>
            </div>
            {direct ? (
              <nav
                aria-label="Inspector sections"
                style={directNavigationStyle}
              >
                {sections.map((section) => (
                  <button
                    type="button"
                    key={section.panel.panelId}
                    style={navigationButtonStyle}
                    data-inspector-navigation-panel={section.panel.panelId}
                    aria-current={
                      section.panel.panelId === active.panel.panelId
                        ? "location"
                        : undefined
                    }
                    onClick={() => choose(section)}
                  >
                    {section.panel.label}
                  </button>
                ))}
              </nav>
            ) : (
              <>
                <WorkbookInspectorActionButton
                  ref={triggerRef}
                  aria-controls={navigationId}
                  aria-expanded={menuOpen}
                  onClick={toggleMenu}
                >
                  Sections
                </WorkbookInspectorActionButton>
                <span style={currentSectionStyle}>
                  Current section: {active.panel.label}
                </span>
                {menuOpen ? (
                  <nav
                    id={navigationId}
                    aria-label="Inspector sections"
                    style={navigationMenuStyle}
                  >
                    {sections.map((section) => (
                      <WorkbookInspectorActionButton
                        key={section.panel.panelId}
                        aria-current={
                          section.panel.panelId === active.panel.panelId
                            ? "location"
                            : undefined
                        }
                        onClick={() => choose(section)}
                      >
                        {section.panel.label}
                      </WorkbookInspectorActionButton>
                    ))}
                  </nav>
                ) : null}
              </>
            )}
          </fieldset>
        ) : null}
        {attention.length ? (
          <button
            type="button"
            style={attentionButtonStyle}
            onClick={() => {
              if (attentionRef.current) reveal(attentionRef.current);
              attentionRef.current?.focus({ preventScroll: true });
            }}
          >
            Unfinished work ({attention.length})
          </button>
        ) : null}
      </header>
      <div
        data-inspector-scroll-body
        style={bodyStyle}
        ref={bodyRef}
        onScroll={observeScroll}
      >
        {subject ? (
          <details>
            <summary>Record context</summary>
            <RecordContext subject={subject} />
            <p style={fullLabelStyle}>{subject.label}</p>
          </details>
        ) : null}
        {attention.length ? (
          <section
            ref={attentionRef}
            tabIndex={-1}
            aria-label="Unfinished work"
            style={panelSectionStyle}
          >
            <h3 style={panelTitleStyle}>Unfinished work</h3>
            {attention.map(({ section, entry }) => (
              <div key={entry.workId} data-inspector-attention={entry.category}>
                <button
                  type="button"
                  style={attentionButtonStyle}
                  onClick={() => {
                    if (admitsAttention(entry))
                      choose(section, entry.destination);
                  }}
                >
                  {entry.label}
                </button>
                {entry.actions?.map((action) => (
                  <WorkbookInspectorActionButton
                    key={action.label}
                    onClick={() => {
                      if (admitsAttention(entry)) action.invoke();
                    }}
                  >
                    {action.label}
                  </WorkbookInspectorActionButton>
                ))}
              </div>
            ))}
          </section>
        ) : null}
        {sections.map((section) => (
          <WorkbookInspectorPanelSection
            key={section.panel.panelId}
            panel={section.panel}
            viewSchemaId={config.viewSchemaId}
            onFocus={() => remember(section.panel.panelId)}
            elementRef={(element) => {
              registerSection(section.panel.panelId, element);
              section.elementRef?.(element);
            }}
          >
            {section.content}
          </WorkbookInspectorPanelSection>
        ))}
        {mode === "saved" ? props.feedback : null}
        {subject === null ? null : (
          <section aria-label="Record technical metadata" style={metadataStyle}>
            <WorkbookInspectorTechnicalDetails
              fields={subjectTechnicalFields(subject)}
            />
          </section>
        )}
      </div>
    </aside>
  );
}

function WorkbookInspectorPanelSection({
  children,
  elementRef,
  panel,
  viewSchemaId,
  onFocus,
}: {
  readonly children?: ReactNode;
  readonly elementRef?: ((element: HTMLElement | null) => void) | undefined;
  readonly panel: InspectorPanel;
  readonly viewSchemaId: string;
  readonly onFocus?: (() => void) | undefined;
}) {
  return (
    <section
      data-testid={workbookInspectorPanelTestId(viewSchemaId, panel.panelId)}
      ref={elementRef}
      style={panelSectionStyle}
      tabIndex={-1}
      onFocusCapture={onFocus}
    >
      <h3 style={panelTitleStyle}>{panel.label}</h3>
      {children}
    </section>
  );
}

function RecordContext({
  subject,
}: {
  readonly subject: WorkbookRecordSubject;
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
  subject: WorkbookRecordSubject,
): WorkbookInspectorTechnicalField[] {
  return [
    { label: "Record ID", value: subject.recordId },
    { label: "Row version", value: String(subject.rowVersion) },
  ];
}

const shellStyle = {
  ...workbookSurfaceInspectorPanelStyle,
  ...workbookTypography("ui"),
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr)",
  overflow: "hidden",
  padding: 0,
} satisfies CSSProperties;
const bodyStyle = {
  minHeight: 0,
  minWidth: 0,
  overflow: "auto",
  overflowAnchor: "none",
  display: "grid",
  alignContent: "start",
  gap: "var(--ct-spacing-sm)",
  padding: "var(--ct-spacing-panel-padding)",
  scrollPaddingBlock: "var(--ct-spacing-sm)",
} satisfies CSSProperties;
const fullLabelStyle = {
  marginBlockEnd: 0,
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const headerStyle = {
  padding: "var(--ct-spacing-xs) var(--ct-spacing-panel-padding)",
  borderBlockEnd: "var(--ct-border-hairline)",
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
const titleStyle = {
  ...workbookTypography("surface-title"),
  display: "-webkit-box",
  WebkitBoxOrient: "vertical",
  WebkitLineClamp: cartularyDesignPresentation.inspector.headerTitleLines,
  overflow: "hidden",
  margin: 0,
  overflowWrap: "anywhere" as const,
} satisfies CSSProperties;
const closeButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "var(--ct-spacing-xs)",
  flexShrink: 0,
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "transparent",
  color: "inherit",
  minInlineSize: "var(--ct-density-default-rowHeight)",
  minBlockSize: "var(--ct-density-default-rowHeight)",
} satisfies CSSProperties;
const directNavigationStyle = {
  display: "flex",
  gap: "var(--ct-spacing-xs)",
  whiteSpace: "nowrap",
  padding: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const navigationButtonStyle = {
  ...workbookTypography("button"),
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  padding: "var(--ct-spacing-xs)",
  minBlockSize: cartularyDesignPresentation.inspector.fieldActionMinSizePx,
  boxSizing: "border-box",
  flexShrink: 0,
  color: "var(--ct-colors-ink)",
  background: "transparent",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  cursor: "pointer",
} satisfies CSSProperties;
const measurementStyle = {
  ...directNavigationStyle,
  position: "absolute",
  visibility: "hidden",
  inlineSize: "max-content",
  pointerEvents: "none",
} satisfies CSSProperties;
const attentionButtonStyle = {
  ...workbookTypography("metadata"),
  minBlockSize: cartularyDesignPresentation.inspector.fieldActionMinSizePx,
  color: "var(--ct-colors-ink)",
  background: "transparent",
  border: 0,
  padding: 0,
  textAlign: "start",
  cursor: "pointer",
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
const panelTitleStyle = {
  ...workbookTypography("section-heading"),
  margin: 0,
} satisfies CSSProperties;
const messageStyle = { margin: 0 } satisfies CSSProperties;
const navigationStyle = {
  margin: 0,
  padding: 0,
  border: 0,
  position: "relative",
  display: "flex",
  alignItems: "center",
  gap: "var(--ct-spacing-sm)",
  minWidth: 0,
} satisfies CSSProperties;
const currentSectionStyle = {
  ...workbookTypography("metadata"),
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const navigationMenuStyle = {
  maxBlockSize: "50vh",
  overflow: "auto",
  position: "absolute",
  top: "100%",
  insetInline: 0,
  zIndex: 1,
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  padding: "var(--ct-spacing-sm)",
  background: "var(--ct-colors-surface-2)",
  border: "var(--ct-border-hairline)",
  boxShadow: "var(--ct-elevation-popover)",
} satisfies CSSProperties;
