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
import { type CSSProperties, type ReactNode, useId } from "react";
import { workbookTypography } from "../../components/workbookFormStyles";
import { workbookSurfaceInspectorPanelStyle } from "../../layout/WorkbookSurfaceLayout";
import { useWorkbookInspectorNavigation } from "../../layout/workbookInspectorNavigation";
import type { WorkbookInspectorSubject } from "../workbookInspectorSubject";
import { WorkbookInspectorActionButton } from "./WorkbookInspectorActions";
import {
  WorkbookInspectorCompactMetadata,
  WorkbookInspectorTechnicalDetails,
} from "./WorkbookInspectorFeedback";
import {
  type WorkbookInspectorTechnicalField,
  workbookInspectorNoRowMessage,
} from "./workbookInspectorPresentationModel";

export type WorkbookInspectorSection = {
  readonly panel: InspectorPanel;
  readonly content: ReactNode;
  readonly focusDestination: (section: HTMLElement) => HTMLElement;
  readonly elementRef?: ((element: HTMLElement | null) => void) | undefined;
};
const noSections: readonly WorkbookInspectorSection[] = [];

export function WorkbookInspectorShell({
  accessibleLabel,
  children,
  config,
  elementRef,
  eyebrow = "Inspector",
  heading,
  mode = "record",
  noRowHeading,
  onClose,
  subject,
  testId,
  sections = noSections,
}: {
  readonly accessibleLabel: string;
  readonly children?: ReactNode;
  readonly config: InspectorConfig;
  readonly elementRef?: ((element: HTMLElement | null) => void) | undefined;
  readonly eyebrow?: string | undefined;
  readonly heading?: string | undefined;
  readonly mode?: "record" | "creation" | undefined;
  readonly noRowHeading: string;
  readonly onClose: () => void;
  readonly subject: WorkbookInspectorSubject | null;
  readonly testId?: string | undefined;
  readonly sections?: readonly WorkbookInspectorSection[] | undefined;
}) {
  const headingId = useId();
  const navigationId = useId();
  const scope = JSON.stringify([
    config.viewSchemaId,
    subject?.recordId ?? null,
    subject?.kind ?? mode,
  ]);
  const {
    active,
    menuOpen,
    bodyRef,
    navigationRef,
    triggerRef,
    closeRef,
    choose,
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
      `}</style>
      <header style={headerStyle}>
        <div style={titleRowStyle}>
          <div style={titleStackStyle}>
            <p style={eyebrowStyle}>{subject ? "Inspector" : eyebrow}</p>
            <h2 id={headingId} style={titleStyle}>
              {subject?.label ?? heading ?? noRowHeading}
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
          </button>
        </div>
        {subject === null ? (
          mode === "record" ? (
            <p style={messageStyle}>{workbookInspectorNoRowMessage}</p>
          ) : null
        ) : (
          <RecordContext subject={subject} />
        )}
        {active ? (
          <fieldset
            ref={navigationRef}
            aria-label="Section navigation"
            style={navigationStyle}
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
          </fieldset>
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
            <p style={fullLabelStyle}>{subject.label}</p>
          </details>
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
        {children}
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

export function WorkbookInspectorPanelSection({
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
  padding: "var(--ct-spacing-panel-padding)",
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
const eyebrowStyle = {
  ...workbookTypography("metadata"),
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  textTransform: "uppercase" as const,
} satisfies CSSProperties;
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
  flexShrink: 0,
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
