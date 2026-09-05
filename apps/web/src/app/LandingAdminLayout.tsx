import {
  incidentLandingTestId,
  landingAdminMenuItemTestId,
  landingAdminPanelTestId,
  landingAdminShellTestId,
} from "@cartulary/ui-contracts";
import {
  FileClock,
  FolderOpen,
  LockKeyhole,
  Package,
  Palette,
  Upload,
  UserRound,
  UsersRound,
} from "lucide-react";
import { type KeyboardEvent, type MutableRefObject, useRef } from "react";
import {
  brandBlockStyle,
  incidentDirectoryShellStyle,
  landingAccountNavButtonSelectedStyle,
  landingAccountNavButtonStyle,
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
  IncidentDirectoryShellProps,
  LandingAdminPanelDescriptor,
  LandingAdminPanelId,
  LandingAdminShellProps,
} from "./landingAdminTypes";

const panelIcons: Record<LandingAdminPanelId, typeof FolderOpen> = {
  incidents: FolderOpen,
  "deployment-users": UsersRound,
  "administrative-audit": FileClock,
  "reference-packs": Package,
  "incident-import": Upload,
  "account-profile": UserRound,
  "account-appearance": Palette,
  "account-security": LockKeyhole,
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
    new Map<LandingAdminPanelId, HTMLButtonElement>(),
  );
  const deploymentPanels = availablePanels.filter(
    (panel) => panel.group === "deployment",
  );
  const navigationPanels = deploymentPanels;

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
    const currentIndex = navigationPanels.findIndex(
      (panel) => panel.token === activePanel,
    );
    const lastIndex = navigationPanels.length - 1;
    const selectByIndex = (index: number) => {
      event.preventDefault();
      selectPanel(
        (navigationPanels[index]?.token ??
          "deployment-users") as DeploymentAdministrationPanelToken,
        true,
      );
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
          <h1 ref={headingRef} tabIndex={-1} style={landingAdminTitleStyle}>
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
            {deploymentPanels.length > 0 ? (
              <MenuGroup
                title="Administration"
                panels={deploymentPanels}
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
          <h1 ref={headingRef} tabIndex={-1} style={landingAdminTitleStyle}>
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
  menuItemRefs: MutableRefObject<Map<LandingAdminPanelId, HTMLButtonElement>>;
  onSelect: (
    panel: DeploymentAdministrationPanelToken,
    focus?: boolean,
  ) => void;
  panels: ReadonlyArray<LandingAdminPanelDescriptor>;
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
          const token = panel.token as DeploymentAdministrationPanelToken;
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
  compact = false,
  onClick,
  panel,
  refCallback,
  selected,
}: {
  compact?: boolean;
  onClick: () => void;
  panel: LandingAdminPanelDescriptor;
  refCallback?: (element: HTMLButtonElement | null) => void;
  selected: boolean;
}) {
  const Icon = panelIcons[panel.token];
  const style = compact
    ? selected
      ? landingAccountNavButtonSelectedStyle
      : landingAccountNavButtonStyle
    : selected
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
      <Icon size={compact ? 15 : 17} strokeWidth={2.2} />
      <span style={landingAdminMenuItemTextStyle}>
        <span style={landingAdminMenuItemLabelStyle}>{panel.label}</span>
        {compact ? null : (
          <span style={landingAdminMenuItemDescriptionStyle}>
            {panel.description}
          </span>
        )}
      </span>
    </button>
  );
}
