import {
  incidentControlsCloseButtonTestId,
  incidentControlsPanelTestId,
} from "@cartulary/ui-contracts";
import { X } from "lucide-react";
import type { ReactNode, RefObject } from "react";
import type { WorkbookIncidentControlsMenuItem } from "../../shared/workbookShellContracts";

type IncidentControlsDrawerProps = {
  readonly activeMenuItem: WorkbookIncidentControlsMenuItem;
  readonly children: ReactNode;
  readonly closeButtonRef: RefObject<HTMLButtonElement | null>;
  readonly onClose: (options: {
    readonly restoreTriggerFocus: boolean;
  }) => void;
};

export function IncidentControlsDrawer({
  activeMenuItem,
  children,
  closeButtonRef,
  onClose,
}: IncidentControlsDrawerProps) {
  return (
    <section
      aria-labelledby="incident-controls-panel-title"
      data-testid={incidentControlsPanelTestId()}
      data-workbook-shell-region="support"
      id={incidentControlsPanelTestId()}
      role="dialog"
      style={supportRegionStyle}
      onKeyDown={(event) => {
        if (event.key === "Escape") {
          event.preventDefault();
          onClose({ restoreTriggerFocus: true });
        }
      }}
    >
      <header style={supportRegionHeaderStyle}>
        <div>
          <p style={supportRegionEyebrowStyle}>Controls</p>
          <h2
            id="incident-controls-panel-title"
            style={supportRegionTitleStyle}
          >
            {activeMenuItem.label}
          </h2>
        </div>
        <button
          ref={closeButtonRef}
          aria-label="Close incident controls"
          data-testid={incidentControlsCloseButtonTestId()}
          style={supportRegionCloseButtonStyle}
          type="button"
          onClick={() => {
            onClose({ restoreTriggerFocus: true });
          }}
        >
          <X aria-hidden="true" size={16} />
        </button>
      </header>
      <div style={supportRegionBodyStyle}>{children}</div>
    </section>
  );
}

const supportRegionStyle = {
  position: "absolute" as const,
  zIndex: 12,
  top: "var(--ct-spacing-md)",
  right: "var(--ct-spacing-md)",
  bottom: "var(--ct-spacing-md)",
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr)",
  inlineSize: "min(52rem, calc(100% - var(--ct-spacing-xl)))",
  maxInlineSize: "calc(100% - var(--ct-spacing-xl))",
  minBlockSize: 0,
  overflow: "hidden",
  boxSizing: "border-box" as const,
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-drawer)",
};

const supportRegionHeaderStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "var(--ct-spacing-md)",
  minWidth: 0,
  padding: "0.7rem 0.85rem",
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
};

const supportRegionEyebrowStyle = {
  margin: 0,
  fontSize: "0.69rem",
  lineHeight: 1.2,
  letterSpacing: 0,
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-ink-subtle)",
  fontWeight: 700,
};

const supportRegionTitleStyle = {
  margin: "0.15rem 0 0",
  fontSize: "1rem",
  lineHeight: 1.2,
};

const supportRegionCloseButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  flex: "0 0 auto",
  width: "1.75rem",
  height: "1.75rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  cursor: "pointer",
};

const supportRegionBodyStyle = {
  minBlockSize: 0,
  minWidth: 0,
  overflow: "auto",
  padding: "var(--ct-spacing-md)",
};
