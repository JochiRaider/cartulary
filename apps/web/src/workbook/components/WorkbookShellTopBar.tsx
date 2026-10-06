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
import {
  type ReactNode,
  type RefObject,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import { useRegisteredOverlayNavigation } from "../../shared/useRegisteredOverlayNavigation";
import type { WorkbookCollaborationSnapshot } from "../collaboration/WorkbookCollaborationCoordinator";
import type { WorkbookLayoutSnapshot } from "../layout/useWorkbookLayoutFacade";
import {
  activeSystemViewTitleStyle,
  builtInSurfaceFocusStyles,
  currentUserChipStyle,
  currentUserSlotStyle,
  shellIncidentIdentityStyle,
  shellTopBarActionsStyle,
  shellTopBarControlsStyle,
  shellTopBarStyle,
  shellTopBarUnsupportedControlsStyle,
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
import {
  builtInWorkbookSurfacePanelId,
  requiredBuiltInWorkbookSurfaceIds,
} from "../models/workbookSurfaceRegistry";
import { displayInitials } from "../utils/workbookPresence";
import { SystemViewSwitcher } from "./SystemViewSwitcher";
import { WorkbookIncidentIdentityDisclosure } from "./WorkbookIncidentIdentityDisclosure";
import { WorkbookShellSlotRegion } from "./WorkbookShellSlots";
import { WorkbookPresenceSummary } from "./WorkbookStatusStrip";
import { workbookCommandStateStyles } from "./workbookFormStyles";

type WorkbookShellTopBarProps = {
  readonly recovery?: ReactNode;
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
  recovery,
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
  // Presentation focus only: never a second owner of workbook selection.
  const entrySurface =
    !networkAnalysisActive &&
    requiredBuiltInWorkbookSurfaceIds.includes(surface)
      ? surface
      : requiredBuiltInWorkbookSurfaceIds[0];
  const [focusedSurface, setFocusedSurface] = useState<string | null>(null);
  const surfaceControls = useRef(new Map<string, HTMLButtonElement>());
  const selectorFocus = useRef<HTMLElement | null>(null);
  const desktop = layout.chromeMode === "base";
  const previousDesktop = useRef(desktop);
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
  useLayoutEffect(() => {
    if (previousDesktop.current === desktop) return;
    previousDesktop.current = desktop;
    const previousControl = selectorFocus.current;
    selectorFocus.current = null;
    setFocusedSurface(null);
    if (desktop) surfacesMenuNavigation.close({ restoreTriggerFocus: false });
    // Only replace focus removed with this selector, never newer external focus.
    if (
      previousControl &&
      (document.activeElement === previousControl ||
        (document.activeElement === document.body &&
          !previousControl.isConnected))
    ) {
      const replacement = desktop
        ? entrySurface === undefined
          ? null
          : surfaceControls.current.get(entrySurface)
        : surfacesMenuTriggerRef.current;
      replacement?.focus({ preventScroll: true });
    }
  });
  const selectExplicitSurface = (viewSchemaId: string) => {
    // An explicitly selected menu item retires; it no longer owns replacement focus.
    if (!desktop) selectorFocus.current = null;
    onSelectSurface(viewSchemaId, { focusFirstGridTarget: true });
  };
  const incidentKeyLabel = incidentIdentity?.incident_key ?? "Incident";
  const incidentTitleLabel =
    incidentIdentity?.title ?? incidentIdentityError ?? "Loading incident";

  return (
    <WorkbookShellSlotRegion
      slot="top-bar"
      style={shellTopBarStyle}
      viewSchemaId={surface}
    >
      <style>
        {workbookCommandStateStyles}
        {builtInSurfaceFocusStyles}
      </style>
      <div
        style={
          layout.chromeMode === "below_supported_minimum"
            ? shellTopBarUnsupportedControlsStyle
            : shellTopBarControlsStyle
        }
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
          <WorkbookIncidentIdentityDisclosure
            key={incidentKeyLabel}
            incidentKey={incidentKeyLabel}
            title={incidentTitleLabel}
          />
        </div>
        <span
          aria-hidden="true"
          data-testid={workbookResponsiveBandTestId()}
          data-workbook-block-mode={layout.blockMode}
          data-workbook-responsive-band={layout.chromeMode}
          hidden
        />
        {layout.chromeMode === "base" ? (
          <div
            data-grid-editor-external-action="true"
            aria-label="Built-in workbook surfaces"
            role="tablist"
            aria-orientation="horizontal"
            style={tabStripStyle}
            onBlur={(event) => {
              if (
                event.relatedTarget instanceof Node &&
                event.currentTarget.contains(event.relatedTarget)
              )
                return;
              setFocusedSurface(null);
              if (event.relatedTarget !== null) selectorFocus.current = null;
            }}
          >
            {requiredBuiltInWorkbookSurfaceIds.map((viewSchemaId, index) => {
              const contract = requireViewContract(viewSchemaId);
              const selected =
                !networkAnalysisActive && surface === viewSchemaId;
              return (
                <button
                  aria-selected={selected}
                  aria-controls={builtInWorkbookSurfacePanelId(viewSchemaId)}
                  id={surfaceTabTestId(viewSchemaId)}
                  role="tab"
                  data-testid={surfaceTabTestId(viewSchemaId)}
                  data-view-schema-id={viewSchemaId}
                  data-workbook-tab-index={String(index)}
                  key={viewSchemaId}
                  ref={(element) => {
                    if (element)
                      surfaceControls.current.set(viewSchemaId, element);
                    else surfaceControls.current.delete(viewSchemaId);
                  }}
                  tabIndex={
                    (focusedSurface ?? entrySurface) === viewSchemaId ? 0 : -1
                  }
                  onFocus={(event) => {
                    selectorFocus.current = event.currentTarget;
                    setFocusedSurface(viewSchemaId);
                  }}
                  onKeyDown={(event) => {
                    if (
                      event.defaultPrevented ||
                      event.nativeEvent.isComposing ||
                      event.altKey ||
                      event.ctrlKey ||
                      event.metaKey ||
                      event.shiftKey
                    )
                      return;
                    const index =
                      requiredBuiltInWorkbookSurfaceIds.indexOf(viewSchemaId);
                    const count = requiredBuiltInWorkbookSurfaceIds.length;
                    // Manual activation; wrapping and Home/End follow local roving precedent.
                    const next =
                      event.key === "ArrowRight"
                        ? (index + 1) % count
                        : event.key === "ArrowLeft"
                          ? (index + count - 1) % count
                          : event.key === "Home"
                            ? 0
                            : event.key === "End"
                              ? count - 1
                              : null;
                    if (next === null) return;
                    event.preventDefault();
                    event.stopPropagation();
                    const target = requiredBuiltInWorkbookSurfaceIds[next];
                    if (target)
                      surfaceControls.current
                        .get(target)
                        ?.focus({ preventScroll: true });
                  }}
                  onClick={() => selectExplicitSurface(viewSchemaId)}
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
          </div>
        ) : (
          <div
            data-grid-editor-external-action="true"
            style={surfacesMenuFrameStyle}
            onFocusCapture={(event) => {
              selectorFocus.current = event.target as HTMLElement;
            }}
            onBlurCapture={(event) => {
              if (
                event.relatedTarget !== null &&
                !(
                  event.relatedTarget instanceof Node &&
                  event.currentTarget.contains(event.relatedTarget)
                )
              )
                selectorFocus.current = null;
            }}
          >
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
                      data-testid={workbookSurfacesMenuOptionTestId(
                        viewSchemaId,
                      )}
                      data-view-schema-id={viewSchemaId}
                      key={viewSchemaId}
                      onClick={() => {
                        surfacesMenuNavigation.close({
                          restoreTriggerFocus: false,
                        });
                        selectExplicitSurface(viewSchemaId);
                      }}
                      onKeyDown={(event) => {
                        surfacesMenuNavigation.onItemKeyDown(
                          event,
                          viewSchemaId,
                        );
                      }}
                      ref={surfacesMenuNavigation.registerItem(viewSchemaId)}
                      role="menuitemradio"
                      style={{
                        ...surfacesMenuItemStyle,
                        ...(selected ? surfacesMenuItemSelectedStyle : null),
                      }}
                      tabIndex={surfacesMenuNavigation.tabIndexFor(
                        viewSchemaId,
                      )}
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
        <div
          style={{
            ...systemViewSlotStyle,
            ...(layout.chromeMode === "below_supported_minimum"
              ? { minWidth: "min-content" }
              : {}),
          }}
        >
          {recovery}
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
            onSelect={selectExplicitSurface}
          />
          {activeSystemSurfaceTitle ? (
            <span style={activeSystemViewTitleStyle}>
              {activeSystemSurfaceTitle}
            </span>
          ) : null}
        </div>
      </div>
      <div style={shellTopBarActionsStyle}>
        {layout.chromeMode === "base" ||
        layout.chromeMode === "narrow_desktop" ? (
          <WorkbookPresenceSummary records={collaboration.presence.header} />
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
