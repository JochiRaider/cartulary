import {
  accountTestId,
  currentIncidentRoleTestId,
  incidentControlsMenuItemTestId,
  incidentControlsMenuTestId,
  incidentControlsTriggerTestId,
  incidentLandingTestId,
} from "@cartulary/ui-contracts";
import { ChevronDown, UserRound } from "lucide-react";
import { type KeyboardEvent, useEffect, useId, useRef, useState } from "react";
import { useRegisteredOverlayNavigation } from "../shared/useRegisteredOverlayNavigation";
import {
  accountMenuAnchorStyle,
  accountMenuItemStyle,
  accountMenuSeparatorStyle,
  accountMenuStatusItemStyle,
  accountMenuStyle,
  accountMenuTriggerStyle,
  accountMenuTriggerTextStyle,
  accountSubmenuItemDescriptionStyle,
  accountSubmenuItemLabelStyle,
  accountSubmenuItemSelectedStyle,
  accountSubmenuItemStyle,
  accountSubmenuStyle,
  visuallyHiddenStyle,
} from "./landingAdminStyles";
import type { AccountApplicationMenuProps } from "./landingAdminTypes";
import { useAccountMenuLayout } from "./useAccountMenuLayout";

type MenuState = "closed" | "root" | "controls";
type RootAction =
  | "incidents"
  | "deployment-administration"
  | "controls"
  | "account-settings";
const actionLabels: Record<RootAction, string> = {
  incidents: "Incidents",
  "deployment-administration": "Deployment administration",
  controls: "Controls",
  "account-settings": "Account settings",
};

function consumeActivation(event: KeyboardEvent<HTMLElement>) {
  if (event.altKey || event.ctrlKey || event.metaKey) return;
  if (event.key !== "Enter" && event.key !== " ") return;
  event.stopPropagation();
  // Native click remains the only activation path, including Space keyup.
  if (event.repeat) event.preventDefault();
}

export function AccountApplicationMenu({
  canOpenDeploymentAdministration,
  currentContext,
  currentUserLabel,
  currentIncidentRole,
  incidentControls,
  onOpenAccountSettings,
  onOpenDeploymentAdministration,
  onOpenIncidentDirectory,
  subjectKey = currentContext,
  triggerFocusRef,
  triggerTestId,
}: AccountApplicationMenuProps) {
  const [state, setState] = useState<MenuState>("closed");
  const containerRef = useRef<HTMLFieldSetElement>(null);
  const localTriggerRef = useRef<HTMLButtonElement>(null);
  const triggerRef = triggerFocusRef ?? localTriggerRef;
  const controlsTriggerRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);
  const tabDepartureRef = useRef(false);
  const activationAvailableRef = useRef(false);
  activationAvailableRef.current = state !== "closed";
  const id = useId();
  const menuId = `${id}-menu`;
  const controlsId = `${id}-controls`;
  const statusId = `${id}-status`;
  const nestedDescriptionId = `${id}-nested-description`;
  const controlItems =
    currentContext === "workbook" ? (incidentControls?.items ?? []) : [];
  const rootKeys: RootAction[] = ["incidents"];
  if (canOpenDeploymentAdministration)
    rootKeys.push("deployment-administration");
  if (controlItems.length > 0) rootKeys.push("controls");
  rootKeys.push("account-settings");
  const selectedRoot = rootKeys.find((key) => key === currentContext) ?? null;
  const isOpen = state !== "closed";
  const panelBounds = useAccountMenuLayout(isOpen, triggerRef, panelRef);
  const controlsOpen = state === "controls" && controlItems.length > 0;
  const navigationSubject = `${subjectKey}:${currentContext}`;
  const rootNavigation = useRegisteredOverlayNavigation<RootAction>({
    initialItemKey: selectedRoot,
    isOpen,
    itemKeys: rootKeys,
    onRequestClose: () => setState("closed"),
    reconcileItems: true,
    restoreFocusOnSubjectChange: false,
    restoreFocusOnUnmount: false,
    subjectKey: navigationSubject,
    triggerRef,
  });
  const controlsNavigation = useRegisteredOverlayNavigation({
    initialItemKey: incidentControls?.activeSection ?? null,
    isOpen: controlsOpen,
    itemKeys: controlItems.map((item) => item.section),
    onRequestClose: () =>
      setState((current) => (current === "controls" ? "root" : current)),
    reconcileItems: true,
    restoreFocusOnSubjectChange: false,
    restoreFocusOnUnmount: false,
    subjectKey: navigationSubject,
    triggerRef: controlsTriggerRef,
  });

  function dismiss() {
    activationAvailableRef.current = false;
    controlsNavigation.close({ restoreTriggerFocus: false });
    rootNavigation.close({ restoreTriggerFocus: false });
  }

  useEffect(() => {
    if (!isOpen) return;
    const outsidePointer = (event: PointerEvent) => {
      if (
        event.target instanceof Node &&
        containerRef.current?.contains(event.target)
      )
        return;
      activationAvailableRef.current = false;
      setState("closed");
    };
    window.addEventListener("pointerdown", outsidePointer, true);
    return () =>
      window.removeEventListener("pointerdown", outsidePointer, true);
  }, [isOpen]);

  function openRoot(preferred: RootAction | null = selectedRoot) {
    tabDepartureRef.current = false;
    rootNavigation.prepareOpen(preferred);
    setState("root");
  }

  function openControls() {
    if (controlItems.length === 0) return;
    controlsNavigation.prepareOpen(incidentControls?.activeSection);
    setState("controls");
  }

  function dispatch(action: RootAction) {
    if (!activationAvailableRef.current || !rootKeys.includes(action)) return;
    if (action === "controls") {
      if (controlsOpen) controlsNavigation.close({ restoreTriggerFocus: true });
      else openControls();
      return;
    }
    dismiss();
    switch (action) {
      case "incidents":
        onOpenIncidentDirectory();
        return;
      case "deployment-administration":
        onOpenDeploymentAdministration();
        return;
      case "account-settings":
        onOpenAccountSettings("account-profile");
        return;
    }
  }

  function onRootKeyDown(event: KeyboardEvent<HTMLElement>, key: RootAction) {
    if (event.altKey || event.ctrlKey || event.metaKey) return;
    consumeActivation(event);
    if (event.key === "Escape" && controlsOpen) {
      event.preventDefault();
      event.stopPropagation();
      controlsNavigation.close({ restoreTriggerFocus: true });
    } else if (event.key === "ArrowRight" && key === "controls") {
      event.preventDefault();
      event.stopPropagation();
      openControls();
    } else rootNavigation.onItemKeyDown(event, key);
  }

  return (
    <fieldset
      ref={containerRef}
      aria-label="Account and application menu"
      style={accountMenuAnchorStyle}
      onKeyDown={(event) => {
        if (event.key === "Tab") tabDepartureRef.current = true;
      }}
      onBlur={(event) => {
        const next = event.relatedTarget;
        if (
          next instanceof Node &&
          containerRef.current?.contains(next) &&
          !(tabDepartureRef.current && next === triggerRef.current)
        )
          return;
        tabDepartureRef.current = false;
        dismiss();
      }}
    >
      <button
        ref={triggerRef}
        aria-controls={isOpen ? menuId : undefined}
        aria-expanded={isOpen}
        aria-haspopup="menu"
        aria-label="Account and application navigation"
        data-testid={triggerTestId}
        style={accountMenuTriggerStyle}
        title={currentUserLabel}
        type="button"
        onClick={() => (isOpen ? dismiss() : openRoot())}
        onKeyDown={(event) => {
          if (event.altKey || event.ctrlKey || event.metaKey) return;
          consumeActivation(event);
          if (event.key === "ArrowDown" || event.key === "ArrowUp") {
            event.preventDefault();
            event.stopPropagation();
            openRoot(
              event.key === "ArrowUp"
                ? (rootKeys[rootKeys.length - 1] ?? null)
                : (rootKeys[0] ?? null),
            );
          } else if (event.key === "Escape" && isOpen)
            onRootKeyDown(event, "incidents");
        }}
      >
        <UserRound aria-hidden="true" size={16} />
        <span style={accountMenuTriggerTextStyle}>{currentUserLabel}</span>
        <ChevronDown aria-hidden="true" size={15} />
      </button>
      {isOpen ? (
        <div
          ref={panelRef}
          data-testid={accountTestId("application-menu")}
          style={{ ...accountMenuStyle, ...panelBounds }}
        >
          <div id={statusId} style={accountMenuStatusItemStyle}>
            <div>{currentUserLabel}</div>
            {currentContext === "workbook" ? (
              <div data-testid={currentIncidentRoleTestId()}>
                Current incident role: {currentIncidentRole || "viewer"}
              </div>
            ) : null}
            {canOpenDeploymentAdministration ? (
              <div>Deployment administrator</div>
            ) : null}
          </div>
          <span id={nestedDescriptionId} style={visuallyHiddenStyle}>
            Opens incident Controls menu
          </span>
          <div
            id={menuId}
            role="menu"
            aria-label="Account and application navigation"
            aria-describedby={statusId}
          >
            {rootKeys.map((key) => (
              <div key={key} role="none">
                {key === "account-settings" ? (
                  <hr style={accountMenuSeparatorStyle} />
                ) : null}
                <button
                  ref={(element) => {
                    rootNavigation.registerItem(key)(element);
                    if (key === "controls")
                      controlsTriggerRef.current = element;
                  }}
                  type="button"
                  role="menuitem"
                  aria-current={selectedRoot === key ? "page" : undefined}
                  aria-controls={
                    key === "controls" && controlsOpen ? controlsId : undefined
                  }
                  aria-describedby={
                    key === "controls" ? nestedDescriptionId : undefined
                  }
                  aria-expanded={key === "controls" ? controlsOpen : undefined}
                  aria-haspopup={key === "controls" ? "menu" : undefined}
                  data-testid={
                    key === "controls"
                      ? incidentControlsTriggerTestId()
                      : key === "incidents" && currentContext === "workbook"
                        ? incidentLandingTestId("return")
                        : undefined
                  }
                  style={
                    selectedRoot === key
                      ? {
                          ...accountMenuItemStyle,
                          background: "var(--ct-colors-surface-3)",
                          color: "var(--ct-colors-ink)",
                        }
                      : accountMenuItemStyle
                  }
                  tabIndex={controlsOpen ? -1 : rootNavigation.tabIndexFor(key)}
                  onFocus={() => {
                    rootNavigation.onItemFocus(key);
                    if (key !== "controls")
                      setState((current) =>
                        current === "controls" ? "root" : current,
                      );
                  }}
                  onClick={() => dispatch(key)}
                  onKeyDown={(event) => onRootKeyDown(event, key)}
                >
                  {actionLabels[key]}
                </button>
                {key === "controls" && controlsOpen ? (
                  <div
                    id={controlsId}
                    data-testid={incidentControlsMenuTestId()}
                    role="menu"
                    aria-label="Incident Controls"
                    style={accountSubmenuStyle}
                  >
                    {controlItems.map((item) => (
                      <button
                        key={item.section}
                        ref={controlsNavigation.registerItem(item.section)}
                        data-testid={incidentControlsMenuItemTestId(
                          item.section,
                        )}
                        role="menuitem"
                        aria-current={
                          item.section === incidentControls?.activeSection
                            ? "true"
                            : undefined
                        }
                        aria-description="Opens incident Controls drawer"
                        style={
                          item.section === incidentControls?.activeSection
                            ? accountSubmenuItemSelectedStyle
                            : accountSubmenuItemStyle
                        }
                        type="button"
                        tabIndex={controlsNavigation.tabIndexFor(item.section)}
                        onFocus={() =>
                          controlsNavigation.onItemFocus(item.section)
                        }
                        onKeyDown={(event) => {
                          if (event.altKey || event.ctrlKey || event.metaKey)
                            return;
                          consumeActivation(event);
                          if (event.key === "ArrowLeft") {
                            event.preventDefault();
                            event.stopPropagation();
                            controlsNavigation.close({
                              restoreTriggerFocus: true,
                            });
                          } else
                            controlsNavigation.onItemKeyDown(
                              event,
                              item.section,
                            );
                        }}
                        onClick={() => {
                          if (
                            !activationAvailableRef.current ||
                            !controlItems.some(
                              (current) => current.section === item.section,
                            )
                          )
                            return;
                          dismiss();
                          incidentControls?.onSelectSection(
                            item.section,
                            triggerRef.current,
                          );
                        }}
                      >
                        <span style={accountSubmenuItemLabelStyle}>
                          {item.label}
                        </span>
                        <span style={accountSubmenuItemDescriptionStyle}>
                          {item.description}
                        </span>
                      </button>
                    ))}
                  </div>
                ) : null}
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </fieldset>
  );
}
