import {
  appRouteTestId,
  landingAdminMenuItemTestId,
  landingAdminPanelTestId,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  lazy,
  Suspense,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type {
  WorkbookAccountApplicationMenuProps,
  WorkbookAccountModel,
} from "../shared/workbookShellContracts";
import { WorkbookMutationRuntimeRegistry } from "../workbook/runtime/WorkbookMutationRuntimeRegistry";
import {
  AccountSecurityPanel,
  type AccountSessionEvent,
  DeploymentUsersPanel,
} from "./AccountAdministrationPanels";
import { AccountApplicationMenu } from "./AccountApplicationMenu";
import {
  AccountAppearancePanel,
  AccountProfilePanel,
} from "./AccountSettingsPanels";
import { AuthGateway } from "./AuthGateway";
import type {
  ExtensionProfileResource,
  SessionData,
} from "./api/publicHttpTypes";
import { AppSessionController } from "./appSessionController";
import { AdministrativeAuditPanel } from "./DeploymentAuditPanel";
import { IncidentAdminPanel } from "./IncidentAdminPanel";
import { IncidentImportPanel } from "./IncidentImportPanel";
import { IncidentLanding } from "./IncidentLanding";
import type { IncidentCreationController } from "./incidentCreationModel";
import {
  type IncidentDirectoryController,
  incidentDirectoryStatusText,
} from "./incidentDirectoryModel";
import {
  IncidentDirectoryShell,
  LandingAdminShell,
} from "./LandingAdminLayout";
import type {
  AccountSettingsPanelToken,
  DeploymentAdministrationPanelToken,
  DeploymentPanelDescriptor,
} from "./landingAdminTypes";
import {
  ReferencePackAdminPanel,
  type ReferencePackJobResource,
} from "./ReferencePackAdminPanel";
import {
  type AppRouteState,
  type AppRouteWriteMode,
  readAppRouteState,
} from "./routeState";
import { useAppRouteRuntime } from "./useAppRouteRuntime";
import { useAppSession } from "./useAppSession";
import { useIncidentCreation } from "./useIncidentCreation";
import { useIncidentDirectory } from "./useIncidentDirectory";

const LazyWorkbookShell = lazy(async () => {
  const module = await import("../workbook/WorkbookShell");
  return { default: module.WorkbookShell };
});

type AccountMenuContext =
  | "deployment-administration"
  | "incidents"
  | "workbook";

type AccountMenuOptions = {
  currentIncidentRole?: WorkbookAccountApplicationMenuProps["currentIncidentRole"];
  incidentControls?: WorkbookAccountApplicationMenuProps["incidentControls"];
  onOpenIncidentDirectory?: (() => void) | undefined;
  triggerTestId?: string | undefined;
};

type AppProps = {
  readonly readingProfile?: CartularyReadingProfile | undefined;
  readonly themeId?: string | undefined;
};

export type CartularyReadingProfile = "default" | "hyperlegible";

const defaultAuthPrompt = "Use your deployment account.";
const accessLostLandingNotice =
  "The current incident is no longer visible. Returned to the landing screen.";
const defaultRevokedSessionMessage =
  "The current session ended. Sign in again to continue.";
function extensionClaimed(
  profiles: readonly ExtensionProfileResource[] | null,
  profileId: ExtensionProfileResource["profile_id"],
) {
  return (
    profiles?.some(
      (profile) => profile.profile_id === profileId && profile.claimed,
    ) ?? false
  );
}

export function App({ readingProfile = "default", themeId }: AppProps = {}) {
  const { commitRoute: publishRoute, route, routeRef } = useAppRouteRuntime();
  const creationControllerRef = useRef<IncidentCreationController | null>(null);
  const directoryControllerRef = useRef<IncidentDirectoryController | null>(
    null,
  );
  const sessionControllerRef = useRef<AppSessionController | null>(null);
  const commitRoute = useCallback(
    (next: AppRouteState, mode: AppRouteWriteMode) => {
      creationControllerRef.current?.leaveSurface();
      if (next.incidentId !== "" || next.deploymentAdministration) {
        directoryControllerRef.current?.setActive(false);
      }
      sessionControllerRef.current?.navigationChanged();
      publishRoute(next, mode);
    },
    [publishRoute],
  );
  useEffect(() => {
    const leaveCreation = () => {
      creationControllerRef.current?.leaveSurface();
      const next = readAppRouteState();
      if (next.incidentId !== "" || next.deploymentAdministration) {
        directoryControllerRef.current?.setActive(false);
      }
      sessionControllerRef.current?.navigationChanged();
    };
    window.addEventListener("popstate", leaveCreation);
    return () => window.removeEventListener("popstate", leaveCreation);
  }, []);
  const workbookMutationRuntimeRegistry = useMemo(
    () => new WorkbookMutationRuntimeRegistry(),
    [],
  );
  const [sessionController] = useState(
    () =>
      new AppSessionController({
        retireLifetime: (lifetime) => {
          workbookMutationRuntimeRegistry.sessionUnavailable();
          creationControllerRef.current?.setSession(lifetime);
          directoryControllerRef.current?.setSession(lifetime);
        },
        replaceAccount: () => workbookMutationRuntimeRegistry.replaceAccount(),
      }),
  );
  sessionControllerRef.current = sessionController;
  const sessionSnapshot = useAppSession(sessionController, () =>
    workbookMutationRuntimeRegistry.dispose(),
  );
  const session = sessionSnapshot.session;
  const sessionRef = useRef<SessionData | null>(session);
  sessionRef.current = session;
  const workbookAuthorizationRecovery = useMemo(
    () =>
      sessionController.recoveryPort(
        (incidentId) =>
          routeRef.current.incidentId === incidentId &&
          !routeRef.current.deploymentAdministration,
      ),
    [sessionController, routeRef],
  );
  const accountPreferences =
    sessionSnapshot.preferences.kind === "ready"
      ? sessionSnapshot.preferences.value
      : null;
  const extensionProfiles =
    sessionSnapshot.extensions.kind === "ready"
      ? sessionSnapshot.extensions.value
      : null;
  const [landingNotice, setLandingNotice] = useState<string | null>(null);
  const error = sessionSnapshot.error;
  const [authPrompt, setAuthPrompt] = useState(defaultAuthPrompt);
  const [activeDeploymentPanel, setActiveDeploymentPanel] =
    useState<DeploymentAdministrationPanelToken>("deployment-users");
  const [accountSettingsPanel, setAccountSettingsPanel] =
    useState<AccountSettingsPanelToken | null>(null);
  const accountMenuTriggerRef = useRef<HTMLButtonElement>(null);
  const accountSettingsCloseRef = useRef<HTMLButtonElement>(null);
  const navigationHeadingRef = useRef<HTMLHeadingElement>(null);
  const navigationFocusRequestRef = useRef<{
    readonly destination: "incidents" | "deployment-administration";
    readonly accountId: string;
    readonly originIdentity: string;
  } | null>(null);
  const previousAccountSettingsRef = useRef<AccountSettingsPanelToken | null>(
    null,
  );
  const restoreAccountMenuRef = useRef(false);
  const accountNavigationIdentity = `${session?.user_id ?? ""}:${route.incidentId}:${route.deploymentAdministration}`;
  const previousAccountNavigationIdentityRef = useRef(
    accountNavigationIdentity,
  );

  useLayoutEffect(() => {
    if (
      previousAccountNavigationIdentityRef.current === accountNavigationIdentity
    )
      return;
    previousAccountNavigationIdentityRef.current = accountNavigationIdentity;
    restoreAccountMenuRef.current = false;
    setAccountSettingsPanel(null);
  }, [accountNavigationIdentity]);

  const closeAccountSettings = useCallback(() => {
    restoreAccountMenuRef.current = true;
    setAccountSettingsPanel(null);
  }, []);

  useLayoutEffect(() => {
    const previous = previousAccountSettingsRef.current;
    previousAccountSettingsRef.current = accountSettingsPanel;
    if (previous === null && accountSettingsPanel !== null) {
      restoreAccountMenuRef.current = false;
      accountSettingsCloseRef.current?.focus({ preventScroll: true });
    } else if (
      previous !== null &&
      accountSettingsPanel === null &&
      restoreAccountMenuRef.current
    ) {
      restoreAccountMenuRef.current = false;
      if (accountMenuTriggerRef.current?.isConnected)
        accountMenuTriggerRef.current.focus({ preventScroll: true });
    }
  }, [accountSettingsPanel]);

  useLayoutEffect(() => {
    const request = navigationFocusRequestRef.current;
    if (request === null) return;
    if (
      session === null ||
      session.user_id !== request.accountId ||
      (request.destination === "deployment-administration" &&
        !session.is_deployment_admin)
    ) {
      navigationFocusRequestRef.current = null;
      return;
    }
    const current =
      route.deploymentAdministration && session.is_deployment_admin
        ? "deployment-administration"
        : route.incidentId === ""
          ? "incidents"
          : "workbook";
    if (current !== request.destination) {
      // Route publication may follow the menu-closing render. Cancel only
      // after another context replaces the request's originating context.
      if (accountNavigationIdentity !== request.originIdentity)
        navigationFocusRequestRef.current = null;
      return;
    }
    if (!navigationHeadingRef.current?.isConnected) return;
    navigationFocusRequestRef.current = null;
    navigationHeadingRef.current.focus({ preventScroll: true });
  });
  const [referencePackJob, setReferencePackJob] =
    useState<ReferencePackJobResource | null>(null);
  const refreshCurrentSession = useCallback(async () => {
    await sessionController.refreshSession();
    sessionController.refreshResources();
  }, [sessionController]);
  const handleAccountSessionEvent = useCallback(
    async (event: AccountSessionEvent) => {
      if (sessionController.getSnapshot().lifetime !== sessionSnapshot.lifetime)
        return;
      if (event.kind === "resource_refresh") {
        await refreshCurrentSession();
        return;
      }
      setAuthPrompt(event.message);
      setReferencePackJob(null);
      setLandingNotice(null);
      if (event.kind === "logout_confirmed")
        sessionController.logoutConfirmed();
      else if (event.kind === "session_lost") sessionController.sessionLost();
      else sessionController.credentialsRevoked();
    },
    [refreshCurrentSession, sessionController, sessionSnapshot.lifetime],
  );
  const handleSessionLost = useCallback(() => {
    if (sessionController.getSnapshot().lifetime !== sessionSnapshot.lifetime)
      return;
    sessionController.sessionLost();
    setReferencePackJob(null);
    setLandingNotice(null);
    setAuthPrompt(defaultRevokedSessionMessage);
  }, [sessionController, sessionSnapshot.lifetime]);
  useEffect(() => {
    if (
      session !== null &&
      route.deploymentAdministration &&
      !session.is_deployment_admin
    ) {
      setActiveDeploymentPanel("deployment-users");
      setLandingNotice(
        "Deployment administration requires deployment admin access.",
      );
      commitRoute(
        { incidentId: "", deploymentAdministration: false },
        "replace",
      );
    }
    if (
      session === null ||
      extensionProfiles === null ||
      !extensionClaimed(extensionProfiles, "reference_pack")
    )
      setReferencePackJob(null);
  }, [commitRoute, extensionProfiles, route.deploymentAdministration, session]);
  const directory = useIncidentDirectory({
    sessionIdentity: sessionSnapshot.lifetime,
    active:
      session !== null &&
      route.incidentId === "" &&
      !route.deploymentAdministration,
    sessionLost: handleSessionLost,
  });
  directoryControllerRef.current = directory.controller;

  const currentUserLabel = useMemo(() => {
    if (session === null) {
      return "Anonymous";
    }
    return session.display_name;
  }, [session]);
  const currentWorkbookAccount = useMemo<WorkbookAccountModel | undefined>(
    () =>
      session === null
        ? undefined
        : {
            display_name: session.display_name,
            is_deployment_admin: session.is_deployment_admin,
            user_id: session.user_id,
          },
    [session],
  );
  const landingStatusText =
    landingNotice ??
    (route.deploymentAdministration
      ? sessionSnapshot.observing
        ? "Checking current session."
        : error !== null
          ? "Current session could not be refreshed."
          : "Deployment administration ready."
      : incidentDirectoryStatusText(directory.state));
  const availableDeploymentPanels = useMemo(() => {
    const panels: DeploymentPanelDescriptor[] = [];
    if (session?.is_deployment_admin) {
      panels.push(
        {
          token: "deployment-users",
          label: "Deployment users",
          description: "Local user administration",
        },
        {
          token: "administrative-audit",
          label: "Administrative audit",
          description: "Deployment audit events",
        },
      );
      if (extensionClaimed(extensionProfiles, "reference_pack")) {
        panels.push({
          token: "reference-packs",
          label: "Reference packs",
          description: "Pack operations",
        });
      }
      if (extensionClaimed(extensionProfiles, "incident_portability")) {
        panels.push({
          token: "incident-import",
          label: "Incident import",
          description: "Create incident from bundle",
        });
      }
    }
    return panels;
  }, [extensionProfiles, session?.is_deployment_admin]);
  useEffect(() => {
    if (
      !availableDeploymentPanels.some(
        (panel) => panel.token === activeDeploymentPanel,
      )
    ) {
      setActiveDeploymentPanel("deployment-users");
    }
  }, [activeDeploymentPanel, availableDeploymentPanels]);
  const readingProfileAttribute =
    readingProfile === "hyperlegible" ? readingProfile : undefined;
  const rootPageStyle =
    readingProfile === "hyperlegible"
      ? {
          ...pageStyle,
          fontFamily: "var(--ct-typography-accessible-reading-fontFamily)",
        }
      : pageStyle;
  const workbookRootPageStyle =
    readingProfile === "hyperlegible"
      ? {
          ...workbookRoutePageStyle,
          fontFamily: "var(--ct-typography-accessible-reading-fontFamily)",
        }
      : workbookRoutePageStyle;

  const openIncident = useCallback(
    (incidentId: string) => {
      const nextRoute = {
        incidentId,
        deploymentAdministration: false,
      };
      setLandingNotice(null);
      commitRoute(nextRoute, "push");
    },
    [commitRoute],
  );

  const navigateToIncidentDirectory = useCallback(() => {
    const nextRoute = {
      incidentId: "",
      deploymentAdministration: false,
    };
    setLandingNotice(null);
    commitRoute(nextRoute, "push");
  }, [commitRoute]);

  const navigateToDeploymentAdministration = useCallback(() => {
    if (!sessionRef.current?.is_deployment_admin) {
      const nextRoute = {
        incidentId: "",
        deploymentAdministration: false,
      };
      setLandingNotice(
        "Deployment administration requires deployment admin access.",
      );
      commitRoute(nextRoute, "replace");
      return;
    }
    const nextRoute = {
      incidentId: "",
      deploymentAdministration: true,
    };
    setLandingNotice(null);
    setActiveDeploymentPanel("deployment-users");
    commitRoute(nextRoute, "push");
  }, [commitRoute]);

  const creation = useIncidentCreation({
    sessionIdentity: sessionSnapshot.lifetime,
    sessionLost: handleSessionLost,
    openIncident: async (incident, signal, canNavigate) => {
      const recovery = sessionController.recoveryPort(() => canNavigate());
      const result = await recovery.recover({
        incidentId: incident.incident_id,
        signal,
      });
      if (!canNavigate() || result.kind === "cancelled") return "cancelled";
      if (result.kind === "access_lost") return "access_lost";
      if (result.kind !== "authorized") return "unavailable";
      openIncident(incident.incident_id);
      return "opened";
    },
  });
  creationControllerRef.current = creation.controller;
  const handleIncidentAccessLost = useCallback(() => {
    if (readAppRouteState().incidentId !== route.incidentId) return;
    directoryControllerRef.current?.setActive(false);
    navigationFocusRequestRef.current = {
      destination: "incidents",
      accountId: sessionRef.current?.user_id ?? "",
      originIdentity: accountNavigationIdentity,
    };
    setLandingNotice(accessLostLandingNotice);
    commitRoute({ incidentId: "", deploymentAdministration: false }, "replace");
  }, [accountNavigationIdentity, commitRoute, route.incidentId]);

  const renderAccountMenu = useCallback(
    (currentContext: AccountMenuContext, options: AccountMenuOptions = {}) => (
      <AccountApplicationMenu
        subjectKey={`${session?.user_id ?? ""}:${route.incidentId}:${currentContext}`}
        triggerFocusRef={accountMenuTriggerRef}
        canOpenDeploymentAdministration={session?.is_deployment_admin ?? false}
        currentContext={currentContext}
        currentIncidentRole={options.currentIncidentRole}
        currentUserLabel={currentUserLabel}
        incidentControls={options.incidentControls}
        onOpenAccountSettings={(panel) => {
          creationControllerRef.current?.leaveSurface();
          setAccountSettingsPanel(panel);
        }}
        onOpenDeploymentAdministration={() => {
          navigationFocusRequestRef.current = {
            destination: "deployment-administration",
            accountId: session?.user_id ?? "",
            originIdentity: accountNavigationIdentity,
          };
          navigateToDeploymentAdministration();
        }}
        onOpenIncidentDirectory={() => {
          navigationFocusRequestRef.current = {
            destination: "incidents",
            accountId: session?.user_id ?? "",
            originIdentity: accountNavigationIdentity,
          };
          (options.onOpenIncidentDirectory ?? navigateToIncidentDirectory)();
        }}
        triggerTestId={options.triggerTestId}
      />
    ),
    [
      accountNavigationIdentity,
      currentUserLabel,
      navigateToDeploymentAdministration,
      navigateToIncidentDirectory,
      session?.is_deployment_admin,
      session?.user_id,
      route.incidentId,
    ],
  );

  const renderAccountSettings = useCallback(() => {
    if (accountSettingsPanel === null || session === null) {
      return null;
    }
    const tabs: ReadonlyArray<{
      label: string;
      token: AccountSettingsPanelToken;
    }> = [
      { token: "account-profile", label: "Profile" },
      { token: "account-appearance", label: "Appearance" },
      { token: "account-security", label: "Security" },
    ];
    return (
      <div style={accountSettingsBackdropStyle}>
        <section
          aria-label="Account settings"
          role="dialog"
          style={accountSettingsDialogStyle}
          onKeyDown={(event) => {
            if (event.key !== "Escape" || event.defaultPrevented) return;
            event.preventDefault();
            event.stopPropagation();
            closeAccountSettings();
          }}
        >
          <header style={accountSettingsHeaderStyle}>
            <div>
              <p style={accountSettingsEyebrowStyle}>Account settings</p>
              <h2 style={accountSettingsTitleStyle}>
                {tabs.find((tab) => tab.token === accountSettingsPanel)?.label}
              </h2>
            </div>
            <button
              ref={accountSettingsCloseRef}
              style={accountSettingsCloseButtonStyle}
              type="button"
              onClick={closeAccountSettings}
            >
              Close
            </button>
          </header>
          <div style={accountSettingsTabsStyle} role="tablist">
            {tabs.map((tab) => (
              <button
                key={tab.token}
                aria-selected={accountSettingsPanel === tab.token}
                role="tab"
                style={
                  accountSettingsPanel === tab.token
                    ? accountSettingsTabSelectedStyle
                    : accountSettingsTabStyle
                }
                type="button"
                onClick={() => {
                  setAccountSettingsPanel(tab.token);
                }}
              >
                {tab.label}
              </button>
            ))}
          </div>
          <div style={accountSettingsPanelStyle}>
            {accountSettingsPanel === "account-profile" ? (
              <AccountProfilePanel onRefreshSession={refreshCurrentSession} />
            ) : null}
            {accountSettingsPanel === "account-appearance" ? (
              <AccountAppearancePanel
                preferences={accountPreferences}
                onPreferencesChange={(value) => {
                  sessionController.preferencesChanged(
                    value,
                    sessionSnapshot.lifetime,
                  );
                }}
              />
            ) : null}
            {accountSettingsPanel === "account-security" ? (
              <AccountSecurityPanel
                onSessionEvent={handleAccountSessionEvent}
              />
            ) : null}
          </div>
        </section>
      </div>
    );
  }, [
    accountPreferences,
    accountSettingsPanel,
    closeAccountSettings,
    handleAccountSessionEvent,
    refreshCurrentSession,
    sessionController,
    sessionSnapshot.lifetime,
    session,
  ]);

  const resourceStatus = (
    <>
      {sessionSnapshot.preferences.kind === "failed" ? (
        <p role="status">
          Appearance preferences unavailable; using the default presentation.{" "}
          <button
            type="button"
            onClick={() => {
              void sessionController.refreshPreferences();
            }}
          >
            Retry preferences
          </button>
        </p>
      ) : null}
      {sessionSnapshot.extensions.kind === "failed" ? (
        <p role="status">
          Extension discovery unavailable.{" "}
          <button
            type="button"
            onClick={() => {
              void sessionController.refreshExtensions();
            }}
          >
            Retry extensions
          </button>
        </p>
      ) : null}
      {session !== null && sessionSnapshot.error !== null ? (
        <p role="status">
          Session refresh unavailable.{" "}
          <button
            type="button"
            onClick={() => {
              void sessionController.refreshSession();
            }}
          >
            Retry session
          </button>
        </p>
      ) : null}
    </>
  );

  if (route.incidentId !== "" && session !== null) {
    return (
      <main
        aria-busy={sessionSnapshot.observing}
        className="cartulary-shell"
        data-bootstrap-state="authenticated"
        data-cartulary-theme={themeId}
        data-reading-profile={readingProfileAttribute}
        data-testid={appRouteTestId("app-shell")}
        style={workbookRootPageStyle}
      >
        {resourceStatus}
        <section style={workbookFrameStyle}>
          <Suspense
            fallback={
              <p
                aria-live="polite"
                data-testid={appRouteTestId("workbook-loading")}
                role="status"
                style={routeLoadingStyle}
              >
                Loading workbook…
              </p>
            }
          >
            <LazyWorkbookShell
              key={sessionSnapshot.lifetime}
              onSessionLost={handleSessionLost}
              authorizationRecovery={workbookAuthorizationRecovery}
              account={currentWorkbookAccount}
              accountDensityMode={accountPreferences?.density_mode ?? null}
              accountApplicationMenu={({
                currentIncidentRole,
                incidentControls,
              }) =>
                renderAccountMenu("workbook", {
                  currentIncidentRole,
                  incidentControls,
                  onOpenIncidentDirectory: navigateToIncidentDirectory,
                  triggerTestId: appRouteTestId("workbook-current-user"),
                })
              }
              currentUserLabel={currentUserLabel}
              incidentId={route.incidentId}
              extensionProfiles={extensionProfiles}
              onIncidentAccessLost={handleIncidentAccessLost}
              renderIncidentControls={(props) => (
                <IncidentAdminPanel {...props} />
              )}
              mutationRuntimeRegistry={workbookMutationRuntimeRegistry}
            />
          </Suspense>
        </section>
        {renderAccountSettings()}
      </main>
    );
  }

  if (session === null && sessionSnapshot.state === "unavailable") {
    return (
      <main
        className="cartulary-shell"
        data-testid={appRouteTestId("app-shell")}
        data-bootstrap-state="unavailable"
        style={rootPageStyle}
      >
        <h1>Session unavailable</h1>
        <p role="status">
          The current session could not be checked. Try again.
        </p>
        <button
          type="button"
          disabled={sessionSnapshot.observing}
          onClick={() => {
            void sessionController.refreshSession();
          }}
        >
          Retry session
        </button>
      </main>
    );
  }
  if (session === null) {
    return (
      <AuthGateway
        bootstrapState={
          sessionSnapshot.state === "unresolved" || sessionSnapshot.observing
            ? "loading"
            : sessionSnapshot.ended
              ? "revoked"
              : "anonymous"
        }
        message={authPrompt}
        onAuthenticated={(nextSession) => {
          if (
            sessionController.authenticationCompleted(
              nextSession,
              sessionSnapshot.revision,
            )
          )
            setAuthPrompt(defaultAuthPrompt);
        }}
        onAuthenticationUncertain={() =>
          sessionController.confirmAuthentication(sessionSnapshot.revision)
        }
        publicError={error}
        readingProfile={readingProfile}
      />
    );
  }

  return (
    <main
      aria-busy={sessionSnapshot.observing}
      className="cartulary-shell"
      data-bootstrap-state="authenticated"
      data-cartulary-theme={themeId}
      data-reading-profile={readingProfileAttribute}
      data-testid={appRouteTestId("app-shell")}
      style={rootPageStyle}
    >
      {resourceStatus}
      {route.deploymentAdministration && session.is_deployment_admin ? (
        <LandingAdminShell
          headingRef={navigationHeadingRef}
          accountMenu={renderAccountMenu("deployment-administration")}
          activePanel={activeDeploymentPanel}
          availablePanels={availableDeploymentPanels}
          currentUserLabel={currentUserLabel}
          onActivePanelChange={setActiveDeploymentPanel}
          statusText={landingStatusText}
        >
          <section
            id={landingAdminPanelTestId("deployment-users")}
            aria-labelledby={landingAdminMenuItemTestId("deployment-users")}
            data-testid={landingAdminPanelTestId("deployment-users")}
            hidden={activeDeploymentPanel !== "deployment-users"}
            style={landingAdminPanelRegionStyle}
          >
            <DeploymentUsersPanel
              autoLoadUsers={activeDeploymentPanel === "deployment-users"}
              enterpriseAuthClaimed={extensionClaimed(
                extensionProfiles,
                "enterprise_authentication",
              )}
              onRefreshSession={refreshCurrentSession}
              session={session}
            />
          </section>
          <section
            id={landingAdminPanelTestId("administrative-audit")}
            aria-labelledby={landingAdminMenuItemTestId("administrative-audit")}
            data-testid={landingAdminPanelTestId("administrative-audit")}
            hidden={activeDeploymentPanel !== "administrative-audit"}
            style={landingAdminPanelRegionStyle}
          >
            <AdministrativeAuditPanel />
          </section>
          {extensionClaimed(extensionProfiles, "reference_pack") ? (
            <section
              id={landingAdminPanelTestId("reference-packs")}
              aria-labelledby={landingAdminMenuItemTestId("reference-packs")}
              data-testid={landingAdminPanelTestId("reference-packs")}
              hidden={activeDeploymentPanel !== "reference-packs"}
              style={landingAdminPanelRegionStyle}
            >
              <ReferencePackAdminPanel
                activeJob={referencePackJob}
                session={session}
                onJobChange={setReferencePackJob}
              />
            </section>
          ) : null}
          {extensionClaimed(extensionProfiles, "incident_portability") ? (
            <section
              id={landingAdminPanelTestId("incident-import")}
              aria-labelledby={landingAdminMenuItemTestId("incident-import")}
              data-testid={landingAdminPanelTestId("incident-import")}
              hidden={activeDeploymentPanel !== "incident-import"}
              style={landingAdminPanelRegionStyle}
            >
              <IncidentImportPanel onOpenImportedIncident={openIncident} />
            </section>
          ) : null}
        </LandingAdminShell>
      ) : (
        <IncidentDirectoryShell
          headingRef={navigationHeadingRef}
          accountMenu={renderAccountMenu("incidents")}
          currentUserLabel={currentUserLabel}
          statusText={landingStatusText}
        >
          <section
            id={landingAdminPanelTestId("incidents")}
            aria-labelledby={landingAdminMenuItemTestId("incidents")}
            data-testid={landingAdminPanelTestId("incidents")}
            style={landingAdminPanelRegionStyle}
          >
            <IncidentLanding
              creation={creation}
              directory={directory}
              onOpenIncident={openIncident}
              notice={landingNotice}
            />
          </section>
        </IncidentDirectoryShell>
      )}
      {renderAccountSettings()}
    </main>
  );
}

const landingAdminPanelRegionStyle: CSSProperties = {
  minWidth: 0,
  minHeight: 0,
  padding: "var(--ct-spacing-md)",
};

const accountSettingsBackdropStyle: CSSProperties = {
  position: "fixed",
  inset: 0,
  zIndex: 40,
  display: "grid",
  placeItems: "center",
  padding: "var(--ct-spacing-lg)",
  background: "rgba(12, 16, 24, 0.42)",
};

const accountSettingsDialogStyle: CSSProperties = {
  width: "min(60rem, calc(100vw - 2rem))",
  maxHeight: "min(52rem, calc(100vh - 2rem))",
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

const pageStyle = {
  minHeight: "100vh",
  margin: 0,
  padding: 0,
  background: "var(--ct-colors-canvas)",
  color: "var(--ct-colors-ink)",
  fontFamily: "var(--ct-typography-ui-fontFamily)",
  fontSize: "var(--ct-typography-ui-fontSize)",
  lineHeight: "var(--ct-typography-ui-lineHeight)",
};

const workbookRoutePageStyle = {
  ...pageStyle,
  blockSize: "var(--ct-app-viewport-block-size)",
  minBlockSize: 0,
  minHeight: 0,
  overflow: "hidden",
};

const workbookFrameStyle = {
  display: "grid",
  gridTemplateRows: "minmax(0, 1fr)",
  blockSize: "100%",
  minBlockSize: 0,
  minWidth: 0,
  overflow: "hidden",
};

const routeLoadingStyle = {
  margin: "1rem 0 0",
  padding: "1rem 1.2rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
};
