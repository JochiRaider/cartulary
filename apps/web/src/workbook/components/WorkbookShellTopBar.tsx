import {
  networkAnalysisTestId,
  surfaceTabTestId,
  workbookIncidentIdentityTestId,
  workbookResponsiveBandTestId,
  workbookSurfacesMenuOptionTestId,
  workbookSurfacesMenuTestId,
  workbookSurfacesMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import { type ReactNode, type RefObject, useRef, useState } from "react";
import type { WorkbookCollaborationSnapshot } from "../collaboration/WorkbookCollaborationCoordinator";
import { useRegisteredOverlayNavigation } from "../focus/useRegisteredOverlayNavigation";
import type { WorkbookLayoutSnapshot } from "../layout/useWorkbookLayoutFacade";
import {
  activeSystemViewTitleStyle,
  currentUserChipStyle,
  currentUserSlotStyle,
  shellIncidentIdentityStyle,
  shellIncidentTitleStyle,
  shellTopBarActionsStyle,
  shellTopBarStyle,
  shellTopBarUnsupportedStyle,
  shellTopBarValueStyle,
  surfaceMenuTriggerStyle,
  surfacesMenuFrameStyle,
  surfacesMenuItemSelectedStyle,
  surfacesMenuItemStyle,
  surfacesMenuStyle,
  surfaceTabActiveStyle,
  surfaceTabStyle,
  systemViewSlotStyle,
  tabStripStyle,
} from "../layout/workbookShellStyles";
import type { WorkbookIncidentIdentity } from "../models/workbookIncidentIdentity";
import { requiredBuiltInWorkbookSurfaceIds } from "../models/workbookSurfaceRegistry";
import { displayInitials } from "../utils/workbookPresence";
import { SystemViewSwitcher } from "./SystemViewSwitcher";
import { WorkbookShellSlotRegion } from "./WorkbookShellSlots";
import { WorkbookPresenceSummary } from "./WorkbookStatusStrip";

type WorkbookShellTopBarProps = {
  readonly account: {
    readonly applicationMenu: ReactNode;
    readonly displayName: string;
    readonly title: string;
  };
  readonly activeSurfaceFocusRef: RefObject<HTMLElement | null>;
  readonly activeSystemSurfaceTitle: string | null;
  readonly collaboration: WorkbookCollaborationSnapshot;
  readonly incidentIdentity: WorkbookIncidentIdentity | null;
  readonly incidentIdentityError: string | null;
  readonly layout: WorkbookLayoutSnapshot;
  readonly networkAnalysisActive: boolean;
  readonly networkAnalysisAvailable: boolean;
  readonly onSelectNetworkAnalysis: () => void;
  readonly onSelectSurface: (
    viewSchemaId: string,
    options?: { readonly focusFirstGridTarget?: boolean },
  ) => void;
  readonly surface: string;
};

/** Owns Workbook route navigation and responsive top-bar presentation. */
export function WorkbookShellTopBar({
  account,
  activeSurfaceFocusRef,
  activeSystemSurfaceTitle,
  collaboration,
  incidentIdentity,
  incidentIdentityError,
  layout,
  networkAnalysisActive,
  networkAnalysisAvailable,
  onSelectNetworkAnalysis,
  onSelectSurface,
  surface,
}: WorkbookShellTopBarProps) {
  const [surfacesMenuOpen, setSurfacesMenuOpen] = useState(false);
  const surfacesMenuTriggerRef = useRef<HTMLButtonElement>(null);
  const surfacesMenuNavigation = useRegisteredOverlayNavigation({
    fallbackFocusRef: activeSurfaceFocusRef,
    initialItemKey: requiredBuiltInWorkbookSurfaceIds.includes(surface)
      ? surface
      : (requiredBuiltInWorkbookSurfaceIds[0] ?? null),
    isOpen: surfacesMenuOpen,
    itemKeys: requiredBuiltInWorkbookSurfaceIds,
    onRequestClose: () => setSurfacesMenuOpen(false),
    subjectKey: surface,
    triggerRef: surfacesMenuTriggerRef,
  });
  const incidentKeyLabel = incidentIdentity?.incident_key ?? "Incident";
  const incidentTitleLabel = incidentIdentity?.title ?? "Loading incident";

  return (
    <WorkbookShellSlotRegion
      slot="top-bar"
      style={{
        ...shellTopBarStyle,
        ...(layout.chromeMode === "below_supported_minimum"
          ? shellTopBarUnsupportedStyle
          : null),
      }}
      viewSchemaId={surface}
    >
      <div
        data-testid={workbookIncidentIdentityTestId()}
        style={shellIncidentIdentityStyle}
        title={
          incidentIdentity === null
            ? (incidentIdentityError ?? "Loading incident")
            : `${incidentIdentity.incident_key} ${incidentIdentity.title}`
        }
      >
        <strong style={shellTopBarValueStyle}>{incidentKeyLabel}</strong>
        <span style={shellIncidentTitleStyle}>{incidentTitleLabel}</span>
      </div>
      <span
        aria-hidden="true"
        data-testid={workbookResponsiveBandTestId()}
        data-workbook-block-mode={layout.blockMode}
        data-workbook-responsive-band={layout.chromeMode}
        hidden
      />
      {layout.chromeMode === "base" ? (
        <nav aria-label="Built-in workbook surfaces" style={tabStripStyle}>
          {requiredBuiltInWorkbookSurfaceIds.map((viewSchemaId, index) => {
            const contract = requireViewContract(viewSchemaId);
            const selected = !networkAnalysisActive && surface === viewSchemaId;
            return (
              <button
                aria-current={selected ? "page" : undefined}
                data-testid={surfaceTabTestId(viewSchemaId)}
                data-view-schema-id={viewSchemaId}
                data-workbook-tab-index={String(index)}
                key={viewSchemaId}
                onClick={() => onSelectSurface(viewSchemaId)}
                style={{
                  ...surfaceTabStyle,
                  ...(selected ? surfaceTabActiveStyle : null),
                }}
                type="button"
              >
                {contract.title}
              </button>
            );
          })}
        </nav>
      ) : (
        <div style={surfacesMenuFrameStyle}>
          <button
            aria-controls={
              surfacesMenuOpen ? workbookSurfacesMenuTestId() : undefined
            }
            aria-expanded={surfacesMenuOpen}
            aria-haspopup="menu"
            data-testid={workbookSurfacesMenuTriggerTestId()}
            ref={surfacesMenuTriggerRef}
            style={surfaceMenuTriggerStyle}
            type="button"
            onClick={() => {
              if (surfacesMenuOpen) {
                surfacesMenuNavigation.close({ restoreTriggerFocus: false });
                return;
              }
              surfacesMenuNavigation.prepareOpen(
                requiredBuiltInWorkbookSurfaceIds.includes(surface)
                  ? surface
                  : requiredBuiltInWorkbookSurfaceIds[0],
              );
              setSurfacesMenuOpen(true);
            }}
            onKeyDown={(event) => {
              if (event.key !== "ArrowDown") return;
              event.preventDefault();
              event.stopPropagation();
              surfacesMenuNavigation.prepareOpen(
                requiredBuiltInWorkbookSurfaceIds.includes(surface)
                  ? surface
                  : requiredBuiltInWorkbookSurfaceIds[0],
              );
              setSurfacesMenuOpen(true);
            }}
          >
            Surfaces
          </button>
          {surfacesMenuOpen ? (
            <div
              data-testid={workbookSurfacesMenuTestId()}
              id={workbookSurfacesMenuTestId()}
              role="menu"
              style={surfacesMenuStyle}
              tabIndex={-1}
              onBlur={surfacesMenuNavigation.onOverlayBlur}
              onKeyDown={(event) => {
                if (
                  event.defaultPrevented ||
                  surfacesMenuNavigation.activeKey === null
                ) {
                  return;
                }
                surfacesMenuNavigation.onItemKeyDown(
                  event,
                  surfacesMenuNavigation.activeKey,
                );
              }}
            >
              {requiredBuiltInWorkbookSurfaceIds.map((viewSchemaId) => {
                const contract = requireViewContract(viewSchemaId);
                const selected =
                  !networkAnalysisActive && surface === viewSchemaId;
                return (
                  <button
                    aria-checked={selected}
                    data-testid={workbookSurfacesMenuOptionTestId(viewSchemaId)}
                    data-view-schema-id={viewSchemaId}
                    key={viewSchemaId}
                    onClick={() => {
                      surfacesMenuNavigation.close({
                        restoreTriggerFocus: false,
                      });
                      onSelectSurface(viewSchemaId);
                    }}
                    onKeyDown={(event) => {
                      surfacesMenuNavigation.onItemKeyDown(event, viewSchemaId);
                    }}
                    ref={surfacesMenuNavigation.registerItem(viewSchemaId)}
                    role="menuitemradio"
                    style={{
                      ...surfacesMenuItemStyle,
                      ...(selected ? surfacesMenuItemSelectedStyle : null),
                    }}
                    tabIndex={surfacesMenuNavigation.tabIndexFor(viewSchemaId)}
                    type="button"
                  >
                    {contract.title}
                  </button>
                );
              })}
            </div>
          ) : null}
        </div>
      )}
      <div style={systemViewSlotStyle}>
        {networkAnalysisAvailable ? (
          <button
            aria-current={networkAnalysisActive ? "page" : undefined}
            data-testid={networkAnalysisTestId("tab")}
            onClick={onSelectNetworkAnalysis}
            style={{
              ...surfaceTabStyle,
              ...(networkAnalysisActive ? surfaceTabActiveStyle : null),
            }}
            type="button"
          >
            Network Analysis
          </button>
        ) : null}
        <SystemViewSwitcher
          activeViewSchemaId={surface}
          onSelect={(viewSchemaId) => {
            onSelectSurface(viewSchemaId, { focusFirstGridTarget: true });
          }}
        />
        {activeSystemSurfaceTitle ? (
          <span style={activeSystemViewTitleStyle}>
            {activeSystemSurfaceTitle}
          </span>
        ) : null}
      </div>
      <div style={shellTopBarActionsStyle}>
        {layout.chromeMode === "base" ||
        layout.chromeMode === "narrow_desktop" ? (
          <WorkbookPresenceSummary
            records={collaboration.activeSheetPresenceRecords}
          />
        ) : null}
        <div style={currentUserSlotStyle}>
          {account.applicationMenu ?? (
            <span style={currentUserChipStyle} title={account.title}>
              {displayInitials(account.displayName)}
            </span>
          )}
        </div>
      </div>
    </WorkbookShellSlotRegion>
  );
}
