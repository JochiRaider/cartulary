import type { CSSProperties, ReactNode, RefObject } from "react";
import { AccountDialog } from "./AccountDialog";
import type { AccountSettingsPanelToken } from "./landingAdminTypes";

export function AccountSettingsDialog({
  panel,
  onClose,
  onSelect,
  fallbackFocusRef,
  children,
}: {
  panel: AccountSettingsPanelToken;
  onClose: () => void;
  onSelect: (panel: AccountSettingsPanelToken) => void;
  fallbackFocusRef: RefObject<HTMLElement | null>;
  children: ReactNode;
}) {
  const tabs: ReadonlyArray<{
    label: string;
    token: AccountSettingsPanelToken;
  }> = [
    { token: "account-profile", label: "Profile" },
    { token: "account-appearance", label: "Appearance" },
    { token: "account-security", label: "Security" },
  ];
  return (
    <AccountDialog
      label="Account settings"
      onClose={onClose}
      fallbackFocusRef={fallbackFocusRef}
      style={accountSettingsDialogStyle}
    >
      {(dismiss) => (
        <>
          <header style={accountSettingsHeaderStyle}>
            <div>
              <p style={accountSettingsEyebrowStyle}>Account settings</p>
              <h2 style={accountSettingsTitleStyle}>
                {tabs.find((tab) => tab.token === panel)?.label}
              </h2>
            </div>
            <button
              style={accountSettingsCloseButtonStyle}
              type="button"
              onClick={dismiss}
            >
              Close
            </button>
          </header>
          <div style={accountSettingsTabsStyle} role="tablist">
            {tabs.map((tab) => (
              <button
                key={tab.token}
                aria-selected={panel === tab.token}
                role="tab"
                id={`account-tab-${tab.token}`}
                aria-controls="account-settings-tabpanel"
                tabIndex={panel === tab.token ? 0 : -1}
                onKeyDown={(event) => {
                  const index = tabs.findIndex((item) => item.token === panel);
                  const next =
                    event.key === "ArrowRight"
                      ? (index + 1) % tabs.length
                      : event.key === "ArrowLeft"
                        ? (index + tabs.length - 1) % tabs.length
                        : event.key === "Home"
                          ? 0
                          : event.key === "End"
                            ? tabs.length - 1
                            : null;
                  const target = next === null ? undefined : tabs[next];
                  if (!target) return;
                  event.preventDefault();
                  onSelect(target.token);
                  document
                    .getElementById(`account-tab-${target.token}`)
                    ?.focus();
                }}
                style={
                  panel === tab.token
                    ? accountSettingsTabSelectedStyle
                    : accountSettingsTabStyle
                }
                type="button"
                onClick={() => {
                  onSelect(tab.token);
                }}
              >
                {tab.label}
              </button>
            ))}
          </div>
          <div
            role="tabpanel"
            id="account-settings-tabpanel"
            aria-labelledby={`account-tab-${panel}`}
            style={accountSettingsPanelStyle}
          >
            {children}
          </div>
        </>
      )}
    </AccountDialog>
  );
}

const accountSettingsDialogStyle: CSSProperties = {
  width: "min(60rem, 100%)",
  maxHeight: "min(52rem, calc(100% - 2 * var(--ct-spacing-lg)))",
  maxWidth: "calc(100% - 2 * var(--ct-spacing-lg))",
  margin: "auto",
  padding: 0,
  position: "fixed",
  inset: 0,
  color: "var(--ct-colors-ink)",
  boxSizing: "border-box",
  display: "grid",
  gridTemplateRows: "auto auto minmax(0, 1fr)",
  minWidth: 0,
  overflow: "hidden",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-panel)",
};

const accountSettingsHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  gap: "var(--ct-spacing-md)",
  padding: "var(--ct-spacing-md)",
  borderBottom: "var(--ct-border-hairline)",
};

const accountSettingsEyebrowStyle: CSSProperties = {
  margin: 0,
  fontSize: "0.72rem",
  letterSpacing: "0.14em",
  textTransform: "uppercase",
  color: "var(--ct-colors-accent)",
};

const accountSettingsTitleStyle: CSSProperties = {
  margin: "0.2rem 0 0",
  fontSize: "1.15rem",
};

const accountSettingsCloseButtonStyle: CSSProperties = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink)",
  padding: "0.48rem 0.7rem",
  fontWeight: 700,
  cursor: "pointer",
};

const accountSettingsTabsStyle: CSSProperties = {
  display: "flex",
  gap: "0.35rem",
  padding: "var(--ct-spacing-sm) var(--ct-spacing-md)",
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
};

const accountSettingsTabStyle: CSSProperties = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  padding: "0.45rem 0.65rem",
  fontWeight: 700,
  cursor: "pointer",
};

const accountSettingsTabSelectedStyle: CSSProperties = {
  ...accountSettingsTabStyle,
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink)",
};

const accountSettingsPanelStyle: CSSProperties = {
  minWidth: 0,
  minHeight: 0,
  overflow: "auto",
  padding: "var(--ct-spacing-md)",
};
