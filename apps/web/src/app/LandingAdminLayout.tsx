import {
  incidentLandingTestId,
  landingAdminMenuItemTestId,
  landingAdminPanelTestId,
  landingAdminShellTestId,
} from "@cartulary/ui-contracts";
import { FileClock, Package, Upload, UsersRound } from "lucide-react";
import { type KeyboardEvent, type MutableRefObject, useRef } from "react";
import {
  brandBlockStyle,
  incidentDirectoryShellStyle,
  landingAccountNavStyle,
  landingAdminContentStyle,
  landingAdminHeaderMetaStyle,
  landingAdminHeaderStyle,
  landingAdminMenuItemDescriptionStyle,
  landingAdminMenuItemLabelStyle,
  landingAdminMenuItemSelectedStyle,
  landingAdminMenuItemStyle,
  landingAdminMenuItemsStyle,
  landingAdminMenuItemTextStyle,
  landingAdminMenuStyle,
  landingAdminMetaValueStyle,
  landingAdminShellStyle,
  landingAdminTitleStyle,
  landingAdminWorkspaceStyle,
  landingEyebrowStyle,
  landingToolbarLabelStyle,
  menuGroupItemsStyle,
  menuGroupStyle,
  menuGroupTitleStyle,
  visuallyHiddenStyle,
} from "./landingAdminStyles";
import type {
  DeploymentAdministrationPanelToken,
  DeploymentPanelDescriptor,
  IncidentDirectoryShellProps,
  LandingAdminShellProps,
} from "./landingAdminTypes";

const panelIcons: Record<DeploymentAdministrationPanelToken, typeof FileClock> =
  {
    "deployment-users": UsersRound,
    "administrative-audit": FileClock,
    "reference-packs": Package,
    "incident-import": Upload,
  };

export function LandingAdminShell({
  headingRef,
  accountMenu,
  activePanel,
  availablePanels,
  children,
  currentUserLabel,
  onActivePanelChange,
  statusText,
}: LandingAdminShellProps) {
  const menuItemRefs = useRef(
    new Map<DeploymentAdministrationPanelToken, HTMLButtonElement>(),
  );

  function focusPanelMenuItem(panel: DeploymentAdministrationPanelToken) {
    const focus = () => {
      menuItemRefs.current.get(panel)?.focus();
    };
    if (typeof window.requestAnimationFrame === "function") {
      window.requestAnimationFrame(focus);
      return;
    }
    window.setTimeout(focus, 0);
  }

  function selectPanel(
    panel: DeploymentAdministrationPanelToken,
    focus = false,
  ) {
    onActivePanelChange(panel);
    if (focus) {
      focusPanelMenuItem(panel);
    }
  }

  function handleMenuKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const currentIndex = availablePanels.findIndex(
      (panel) => panel.token === activePanel,
    );
    const lastIndex = availablePanels.length - 1;
    const selectByIndex = (index: number) => {
      event.preventDefault();
      selectPanel(availablePanels[index]?.token ?? "deployment-users", true);
    };

    switch (event.key) {
      case "ArrowUp":
      case "ArrowLeft":
        selectByIndex(currentIndex <= 0 ? lastIndex : currentIndex - 1);
        return;
      case "ArrowDown":
      case "ArrowRight":
        selectByIndex(currentIndex >= lastIndex ? 0 : currentIndex + 1);
        return;
      case "Home":
        selectByIndex(0);
        return;
      case "End":
        selectByIndex(lastIndex);
        return;
      default:
        return;
    }
  }

  return (
    <section
      data-testid={landingAdminShellTestId("shell")}
      style={landingAdminShellStyle}
    >
      <header style={landingAdminHeaderStyle}>
        <div style={brandBlockStyle}>
          <p style={landingEyebrowStyle}>Cartulary</p>
          <h1
            data-testid={landingAdminShellTestId("heading")}
            ref={headingRef}
            tabIndex={-1}
            style={landingAdminTitleStyle}
          >
            Deployment administration
          </h1>
        </div>
        <dl style={landingAdminHeaderMetaStyle}>
          <div>
            <dt style={landingToolbarLabelStyle}>Session</dt>
            <dd
              data-testid={incidentLandingTestId("current-user")}
              style={landingAdminMetaValueStyle}
            >
              {currentUserLabel}
            </dd>
          </div>
        </dl>
        <div style={landingAccountNavStyle}>{accountMenu}</div>
      </header>

      <div style={landingAdminWorkspaceStyle}>
        <nav
          data-testid={landingAdminShellTestId("menu")}
          style={landingAdminMenuStyle}
          aria-label="Deployment administration"
          onKeyDown={handleMenuKeyDown}
        >
          <div style={landingAdminMenuItemsStyle}>
            {availablePanels.length > 0 ? (
              <MenuGroup
                title="Administration"
                panels={availablePanels}
                activePanel={activePanel}
                menuItemRefs={menuItemRefs}
                onSelect={selectPanel}
              />
            ) : null}
          </div>
        </nav>
        <div style={landingAdminContentStyle}>{children}</div>
      </div>

      <p aria-live="polite" role="status" style={visuallyHiddenStyle}>
        {statusText}
      </p>
    </section>
  );
}

export function IncidentDirectoryShell({
  headingRef,
  accountMenu,
  children,
  currentUserLabel,
  statusText,
}: IncidentDirectoryShellProps) {
  return (
    <section
      data-testid={landingAdminShellTestId("shell")}
      style={incidentDirectoryShellStyle}
    >
      <header style={landingAdminHeaderStyle}>
        <div style={brandBlockStyle}>
          <p style={landingEyebrowStyle}>Cartulary</p>
          <h1
            data-testid={landingAdminShellTestId("heading")}
            ref={headingRef}
            tabIndex={-1}
            style={landingAdminTitleStyle}
          >
            Incident directory
          </h1>
        </div>
        <dl style={landingAdminHeaderMetaStyle}>
          <div>
            <dt style={landingToolbarLabelStyle}>Session</dt>
            <dd
              data-testid={incidentLandingTestId("current-user")}
              style={landingAdminMetaValueStyle}
            >
              {currentUserLabel}
            </dd>
          </div>
        </dl>
        <div style={landingAccountNavStyle}>{accountMenu}</div>
      </header>
      <div style={landingAdminContentStyle}>{children}</div>
      <p aria-live="polite" role="status" style={visuallyHiddenStyle}>
        {statusText}
      </p>
    </section>
  );
}

function MenuGroup({
  activePanel,
  menuItemRefs,
  onSelect,
  panels,
  title,
}: {
  activePanel: DeploymentAdministrationPanelToken;
  menuItemRefs: MutableRefObject<
    Map<DeploymentAdministrationPanelToken, HTMLButtonElement>
  >;
  onSelect: (
    panel: DeploymentAdministrationPanelToken,
    focus?: boolean,
  ) => void;
  panels: ReadonlyArray<DeploymentPanelDescriptor>;
  title: string;
}) {
  if (panels.length === 0) {
    return null;
  }
  return (
    <div style={menuGroupStyle}>
      <p style={menuGroupTitleStyle}>{title}</p>
      <div style={menuGroupItemsStyle}>
        {panels.map((panel) => {
          const token = panel.token;
          const selected = token === activePanel;
          return (
            <PanelButton
              key={panel.token}
              panel={panel}
              selected={selected}
              refCallback={(element) => {
                if (element === null) {
                  menuItemRefs.current.delete(panel.token);
                  return;
                }
                menuItemRefs.current.set(panel.token, element);
              }}
              onClick={() => {
                onSelect(token);
              }}
            />
          );
        })}
      </div>
    </div>
  );
}

function PanelButton({
  onClick,
  panel,
  refCallback,
  selected,
}: {
  onClick: () => void;
  panel: DeploymentPanelDescriptor;
  refCallback?: (element: HTMLButtonElement | null) => void;
  selected: boolean;
}) {
  const Icon = panelIcons[panel.token];
  const style = selected
    ? landingAdminMenuItemSelectedStyle
    : landingAdminMenuItemStyle;
  return (
    <button
      id={landingAdminMenuItemTestId(panel.token)}
      ref={refCallback}
      aria-controls={landingAdminPanelTestId(panel.token)}
      aria-pressed={selected}
      data-testid={landingAdminMenuItemTestId(panel.token)}
      style={style}
      type="button"
      onClick={onClick}
    >
      <Icon size={17} strokeWidth={2.2} />
      <span style={landingAdminMenuItemTextStyle}>
        <span style={landingAdminMenuItemLabelStyle}>{panel.label}</span>
        <span style={landingAdminMenuItemDescriptionStyle}>
          {panel.description}
        </span>
      </span>
    </button>
  );
}
