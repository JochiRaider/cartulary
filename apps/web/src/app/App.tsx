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
import {
  type APIError,
  extractError,
  publicErrorView,
} from "../services/browserApi";
import type {
  WorkbookAccountApplicationMenuProps,
  WorkbookAccountModel,
} from "../shared/workbookShellContracts";
import { WorkbookMutationRuntimeRegistry } from "../workbook/runtime/WorkbookMutationRuntimeRegistry";
import {
  AccountSecurityPanel,
  DeploymentUsersPanel,
} from "./AccountAdministrationPanels";
import { AccountApplicationMenu } from "./AccountApplicationMenu";
import {
  AccountAppearancePanel,
  AccountProfilePanel,
} from "./AccountSettingsPanels";
import { AuthGateway } from "./AuthGateway";
import {
  type AccountPreferencesResource,
  createAppAuthorizationRecoveryPort,
  type ExtensionProfileResource,
  loadAccountPreferences,
  loadCredentialState,
  loadExtensions,
  loadSession,
  type SessionData,
} from "./api/appShellClient";
import { AdministrativeAuditPanel } from "./DeploymentAuditPanel";
import { IncidentAdminPanel } from "./IncidentAdminPanel";
import { IncidentImportPanel } from "./IncidentImportPanel";
import { IncidentLanding } from "./IncidentLanding";
import type { IncidentCreationController } from "./incidentCreationModel";
import {
  directoryIsLoading,
  type IncidentDirectoryController,
} from "./incidentDirectoryModel";
import {
  IncidentDirectoryShell,
  LandingAdminShell,
} from "./LandingAdminLayout";
import type {
  AccountSettingsPanelToken,
  AppBootstrapState,
  DeploymentAdministrationPanelToken,
  LandingAdminPanelDescriptor,
  LandingRefreshState,
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
import { useIncidentCreation } from "./useIncidentCreation";
import { useIncidentDirectory } from "./useIncidentDirectory";

const LazyWorkbookShell = lazy(async () => {
  const module = await import("../workbook/WorkbookShell");
  return { default: module.WorkbookShell };
});

const LazyDebugHarnessShell = lazy(async () => {
  const module = await import("./debug/DebugHarnessShell");
  return { default: module.DebugHarnessShell };
});

type ShellRefreshOptions = {
  anonymousMessage?: string;
  landingNotice?: string | null;
  routeSnapshot: AppRouteState;
};

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
function isAbortError(error: unknown): boolean {
  if (typeof DOMException !== "undefined" && error instanceof DOMException) {
    return error.name === "AbortError";
  }
  return (
    typeof error === "object" &&
    error !== null &&
    "name" in error &&
    (error as { name?: unknown }).name === "AbortError"
  );
}

function isSessionRequiredError(status: number, error: APIError | null) {
  return (
    status === 401 && (error === null || error.code === "session_required")
  );
}

function extensionClaimed(
  profiles: readonly ExtensionProfileResource[],
  profileId: ExtensionProfileResource["profile_id"],
) {
  return profiles.some(
    (profile) => profile.profile_id === profileId && profile.claimed,
  );
}

function creationSessionIdentity(session: SessionData | null) {
  return session === null
    ? null
    : `${session.user_id}:${session.authenticated_at}:${session.provider_type}`;
}

export function App({ readingProfile = "default", themeId }: AppProps = {}) {
  const { commitRoute: publishRoute, route, routeRef } = useAppRouteRuntime();
  const creationControllerRef = useRef<IncidentCreationController | null>(null);
  const directoryControllerRef = useRef<IncidentDirectoryController | null>(
    null,
  );
  const activeRefreshRef = useRef<{
    controller: AbortController | null;
    requestID: number;
  }>({ controller: null, requestID: 0 });
  const commitRoute = useCallback(
    (next: AppRouteState, mode: AppRouteWriteMode) => {
      creationControllerRef.current?.leaveSurface();
      if (
        next.incidentId !== "" ||
        next.deploymentAdministration ||
        next.debugHarness
      ) {
        directoryControllerRef.current?.setActive(false);
      }
      activeRefreshRef.current.controller?.abort();
      publishRoute(next, mode);
    },
    [publishRoute],
  );
  useEffect(() => {
    const leaveCreation = () => {
      creationControllerRef.current?.leaveSurface();
      const next = readAppRouteState();
      if (
        next.incidentId !== "" ||
        next.deploymentAdministration ||
        next.debugHarness
      ) {
        directoryControllerRef.current?.setActive(false);
      }
      activeRefreshRef.current.controller?.abort();
    };
    window.addEventListener("popstate", leaveCreation);
    return () => window.removeEventListener("popstate", leaveCreation);
  }, []);
  const workbookMutationRuntimeRegistry = useMemo(
    () => new WorkbookMutationRuntimeRegistry(),
    [],
  );
  useEffect(
    () => () => {
      workbookMutationRuntimeRegistry.dispose();
    },
    [workbookMutationRuntimeRegistry],
  );
  const [session, publishSession] = useState<SessionData | null>(null);
  const sessionRef = useRef<SessionData | null>(null);
  const setSession = useCallback((next: SessionData | null) => {
    if (
      creationSessionIdentity(sessionRef.current) !==
      creationSessionIdentity(next)
    ) {
      creationControllerRef.current?.setSession(creationSessionIdentity(next));
      directoryControllerRef.current?.setSession(creationSessionIdentity(next));
    }
    sessionRef.current = next;
    publishSession(next);
  }, []);
  const workbookAuthorizationRecovery = useMemo(
    () =>
      createAppAuthorizationRecoveryPort({
        onSessionRecovered: setSession,
      }),
    [setSession],
  );
  const [accountPreferences, setAccountPreferences] =
    useState<AccountPreferencesResource | null>(null);
  const [credentialError, setCredentialError] = useState<APIError | null>(null);
  const [extensionProfiles, setExtensionProfiles] = useState<
    ExtensionProfileResource[]
  >([]);
  const [appBootstrapState, setAppBootstrapState] =
    useState<AppBootstrapState>("loading");
  const [landingRefreshState, setLandingRefreshState] =
    useState<LandingRefreshState>("idle");
  const [landingNotice, setLandingNotice] = useState<string | null>(null);
  const [error, setError] = useState<APIError | null>(null);
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
  const refreshShell = useCallback(
    async (options: ShellRefreshOptions) => {
      const requestID = activeRefreshRef.current.requestID + 1;
      const previousController = activeRefreshRef.current.controller;
      const controller = new AbortController();
      activeRefreshRef.current = {
        controller,
        requestID,
      };
      previousController?.abort();

      const canCommit = () =>
        activeRefreshRef.current.requestID === requestID &&
        activeRefreshRef.current.controller === controller &&
        !controller.signal.aborted;

      const hadSessionAtRefreshStart =
        sessionRef.current !== null || options.anonymousMessage !== undefined;

      if (!hadSessionAtRefreshStart) {
        setAppBootstrapState("loading");
        setLandingRefreshState("idle");
      } else {
        setAppBootstrapState("authenticated");
        setLandingRefreshState("loading");
      }
      setError(null);
      setLandingNotice(options.landingNotice ?? null);

      try {
        const sessionResult = await loadSession({
          signal: controller.signal,
        });
        if (!canCommit()) {
          return;
        }

        if (!sessionResult.ok) {
          const sessionError = extractError(sessionResult.payload);
          setSession(null);
          setAccountPreferences(null);
          setCredentialError(null);
          setExtensionProfiles([]);
          setReferencePackJob(null);
          setLandingNotice(null);
          setError(
            hadSessionAtRefreshStart ||
              !isSessionRequiredError(sessionResult.status, sessionError)
              ? sessionError
              : null,
          );
          setAuthPrompt(
            options.anonymousMessage ??
              (hadSessionAtRefreshStart
                ? defaultRevokedSessionMessage
                : defaultAuthPrompt),
          );
          setAppBootstrapState(
            isSessionRequiredError(sessionResult.status, sessionError)
              ? hadSessionAtRefreshStart
                ? "revoked"
                : "anonymous"
              : "public_error_envelope",
          );
          setLandingRefreshState("idle");
          return;
        }

        const nextSession = (sessionResult.payload as { data: SessionData })
          .data;
        // Invalidate account-scoped operations as soon as replacement authentication
        // is observed, before any supporting bootstrap requests can settle.
        if (
          sessionRef.current !== null &&
          creationSessionIdentity(sessionRef.current) !==
            creationSessionIdentity(nextSession)
        ) {
          setSession(nextSession);
          setAccountPreferences(null);
          setCredentialError(null);
          setExtensionProfiles([]);
          setReferencePackJob(null);
        }
        const [credentialResult, preferencesResult, extensionsResult] =
          await Promise.all([
            loadCredentialState({ signal: controller.signal }),
            loadAccountPreferences({ signal: controller.signal }),
            loadExtensions({ signal: controller.signal }),
          ]);
        if (!canCommit()) return;
        const nextCredentialError = extractError(credentialResult.payload);
        const preferencesError = extractError(preferencesResult.payload);
        const extensionsError = extractError(extensionsResult.payload);
        const sessionFailure = [
          credentialResult,
          preferencesResult,
          extensionsResult,
        ].find(
          (result) =>
            !result.ok &&
            isSessionRequiredError(result.status, extractError(result.payload)),
        );
        if (sessionFailure) {
          setSession(null);
          setAccountPreferences(null);
          setCredentialError(null);
          setExtensionProfiles([]);
          setReferencePackJob(null);
          setLandingNotice(null);
          setError(extractError(sessionFailure.payload));
          setAuthPrompt(
            options.anonymousMessage ?? defaultRevokedSessionMessage,
          );
          setAppBootstrapState("revoked");
          setLandingRefreshState("idle");
          return;
        }
        setSession(nextSession);
        setAccountPreferences(
          preferencesResult.ok ? preferencesResult.payload.data : null,
        );
        setCredentialError(nextCredentialError);
        setAuthPrompt(defaultAuthPrompt);
        if (!extensionsResult.ok || !credentialResult.ok) {
          setExtensionProfiles([]);
          setReferencePackJob(null);
          setError(extensionsError ?? nextCredentialError ?? preferencesError);
          setAppBootstrapState("public_error_envelope");
          setLandingRefreshState("failed");
          return;
        }
        const nextExtensionProfiles = extensionsResult.payload.data.extensions;
        setExtensionProfiles(nextExtensionProfiles);
        const deploymentRouteDenied =
          options.routeSnapshot.deploymentAdministration &&
          !nextSession.is_deployment_admin;
        if (
          deploymentRouteDenied ||
          !extensionClaimed(nextExtensionProfiles, "reference_pack")
        )
          setReferencePackJob(null);
        setError(null);
        setAppBootstrapState("authenticated");
        setLandingRefreshState("idle");
        if (deploymentRouteDenied) {
          setActiveDeploymentPanel("deployment-users");
          setLandingNotice(
            "Deployment administration requires deployment admin access.",
          );
          commitRoute(
            {
              incidentId: "",
              debugHarness: false,
              deploymentAdministration: false,
            },
            "replace",
          );
        }
      } catch (error) {
        if (isAbortError(error) || !canCommit()) {
          return;
        }
        if (sessionRef.current === null) {
          setSession(null);
          setAccountPreferences(null);
          setCredentialError(null);
          setExtensionProfiles([]);
          setReferencePackJob(null);
          setLandingNotice(null);
          setError(null);
          setAuthPrompt(options.anonymousMessage ?? defaultAuthPrompt);
          setAppBootstrapState("anonymous");
          setLandingRefreshState("idle");
          return;
        }

        setError(null);
        setAuthPrompt(defaultAuthPrompt);
        setAppBootstrapState("public_error_envelope");
        setLandingRefreshState("failed");
      }
    },
    [commitRoute, setSession],
  );

  const getCurrentRoute = useCallback(() => routeRef.current, [routeRef]);

  const refreshCurrentShell = useCallback(
    (options?: { anonymousMessage?: string }) => {
      if (options?.anonymousMessage !== undefined) {
        // Successful sign-out or credential replacement ends this lifetime now.
        setSession(null);
        setAccountPreferences(null);
        setCredentialError(null);
        setExtensionProfiles([]);
        setReferencePackJob(null);
      }
      return refreshShell({
        routeSnapshot: getCurrentRoute(),
        landingNotice: null,
        ...(typeof options?.anonymousMessage === "string"
          ? {
              anonymousMessage: options.anonymousMessage,
            }
          : {}),
      });
    },
    [getCurrentRoute, refreshShell, setSession],
  );

  useEffect(() => {
    return () => {
      activeRefreshRef.current.controller?.abort();
    };
  }, []);

  const previousDebugHarnessRef = useRef(false);
  useEffect(() => {
    const returningFromDebug = previousDebugHarnessRef.current;
    previousDebugHarnessRef.current = route.debugHarness;
    if (route.debugHarness) return;
    if (
      sessionRef.current !== null &&
      !route.deploymentAdministration &&
      !returningFromDebug
    )
      return;
    void refreshShell({ routeSnapshot: route, landingNotice: null });
  }, [route, refreshShell]);

  const handleSessionLost = useCallback(() => {
    activeRefreshRef.current.controller?.abort();
    activeRefreshRef.current.requestID += 1;
    setSession(null);
    setAccountPreferences(null);
    setCredentialError(null);
    setExtensionProfiles([]);
    setReferencePackJob(null);
    setError(null);
    setLandingNotice(null);
    setLandingRefreshState("idle");
    setAppBootstrapState("revoked");
    setAuthPrompt(defaultRevokedSessionMessage);
  }, [setSession]);
  const directory = useIncidentDirectory({
    sessionIdentity: creationSessionIdentity(session),
    active:
      session !== null &&
      appBootstrapState === "authenticated" &&
      route.incidentId === "" &&
      !route.deploymentAdministration &&
      !route.debugHarness,
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
      ? landingRefreshState === "loading"
        ? "Loading deployment administration."
        : landingRefreshState === "failed" || error !== null
          ? "Failed to load deployment administration."
          : "Deployment administration ready."
      : directory.state.failure !== null
        ? directory.state.failure.message
        : error !== null
          ? (publicErrorView(error)?.statusText ??
            "Failed to load visible incidents.")
          : directoryIsLoading(directory.state) ||
              directory.state.phase === "debouncing"
            ? "Searching visible incidents…"
            : directory.state.phase !== "ready"
              ? "Loading visible incidents…"
              : directory.state.incidents.length === 0
                ? "No visible incidents yet."
                : `Loaded ${directory.state.incidents.length} incident${directory.state.incidents.length === 1 ? "" : "s"}${directory.state.paging?.has_more ? "; more available." : "."}`);
  const availableDeploymentPanels = useMemo(() => {
    const panels: LandingAdminPanelDescriptor[] = [];
    if (session?.is_deployment_admin) {
      panels.push(
        {
          token: "deployment-users",
          label: "Deployment users",
          description: "Local user administration",
          group: "deployment",
        },
        {
          token: "administrative-audit",
          label: "Administrative audit",
          description: "Deployment audit events",
          group: "deployment",
        },
      );
      if (extensionClaimed(extensionProfiles, "reference_pack")) {
        panels.push({
          token: "reference-packs",
          label: "Reference packs",
          description: "Pack operations",
          group: "deployment",
        });
      }
      if (extensionClaimed(extensionProfiles, "incident_portability")) {
        panels.push({
          token: "incident-import",
          label: "Incident import",
          description: "Create incident from bundle",
          group: "deployment",
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
        debugHarness: false,
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
      debugHarness: false,
      deploymentAdministration: false,
    };
    setLandingNotice(null);
    commitRoute(nextRoute, "push");
  }, [commitRoute]);

  const navigateToDeploymentAdministration = useCallback(() => {
    if (!sessionRef.current?.is_deployment_admin) {
      const nextRoute = {
        incidentId: "",
        debugHarness: false,
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
      debugHarness: false,
      deploymentAdministration: true,
    };
    setLandingNotice(null);
    setActiveDeploymentPanel("deployment-users");
    commitRoute(nextRoute, "push");
  }, [commitRoute]);

  const creation = useIncidentCreation({
    sessionIdentity: creationSessionIdentity(session),
    sessionLost: handleSessionLost,
    openIncident: async (incident, signal, canNavigate) => {
      if (!canNavigate()) return "cancelled";
      const identity = creationSessionIdentity(sessionRef.current);
      const result = await loadSession({ signal });
      if (!canNavigate()) return "cancelled";
      if (!result.ok) {
        if (isSessionRequiredError(result.status, extractError(result.payload)))
          handleSessionLost();
        return "unavailable";
      }
      const nextSession = (result.payload as { data: SessionData }).data;
      if (creationSessionIdentity(nextSession) !== identity) {
        setSession(nextSession);
        return "cancelled";
      }
      setSession(nextSession);
      if (
        !nextSession.memberships.some(
          (membership) => membership.incident_id === incident.incident_id,
        )
      )
        return "access_lost";
      if (!canNavigate()) return "cancelled";
      activeRefreshRef.current.controller?.abort();
      activeRefreshRef.current.requestID += 1;
      setLandingRefreshState("idle");
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
    commitRoute(
      { incidentId: "", debugHarness: false, deploymentAdministration: false },
      "replace",
    );
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
              <AccountProfilePanel onRefreshShell={refreshCurrentShell} />
            ) : null}
            {accountSettingsPanel === "account-appearance" ? (
              <AccountAppearancePanel
                preferences={accountPreferences}
                onPreferencesChange={setAccountPreferences}
              />
            ) : null}
            {accountSettingsPanel === "account-security" ? (
              <AccountSecurityPanel
                credentialStateError={credentialError}
                onRefreshShell={refreshCurrentShell}
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
    credentialError,
    refreshCurrentShell,
    session,
  ]);

  if (route.incidentId !== "" && session !== null) {
    return (
      <main
        aria-busy={appBootstrapState === "loading"}
        className="cartulary-shell"
        data-bootstrap-state={appBootstrapState}
        data-cartulary-theme={themeId}
        data-reading-profile={readingProfileAttribute}
        data-testid={appRouteTestId("app-shell")}
        style={workbookRootPageStyle}
      >
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

  if (route.debugHarness) {
    return (
      <main
        className="cartulary-shell"
        data-cartulary-theme={themeId}
        data-reading-profile={readingProfileAttribute}
        style={rootPageStyle}
      >
        <section style={utilityPanelStyle}>
          <div style={landingHeroStyle}>
            <p style={landingEyebrowStyle}>Cartulary</p>
            <h1 style={landingHeadlineStyle}>Debug harness shell</h1>
            <p style={landingBodyStyle}>
              Authentication and incident-directory harness controls now live
              behind the explicit `?debug=harness` flag so the default root path
              can behave like the real incident landing.
            </p>
          </div>

          <Suspense
            fallback={
              <p
                aria-live="polite"
                data-testid={appRouteTestId("debug-harness-loading")}
                role="status"
                style={routeLoadingStyle}
              >
                Loading debug harness…
              </p>
            }
          >
            <LazyDebugHarnessShell />
          </Suspense>
        </section>
      </main>
    );
  }

  if (session === null) {
    return (
      <AuthGateway
        bootstrapState={
          appBootstrapState === "revoked" ||
          appBootstrapState === "public_error_envelope"
            ? appBootstrapState
            : appBootstrapState === "loading"
              ? "loading"
              : "anonymous"
        }
        message={authPrompt}
        onAuthenticated={async () => {
          await refreshShell({
            routeSnapshot: getCurrentRoute(),
            landingNotice: null,
          });
        }}
        publicError={error}
        readingProfile={readingProfile}
      />
    );
  }

  return (
    <main
      aria-busy={appBootstrapState === "loading"}
      className="cartulary-shell"
      data-bootstrap-state={appBootstrapState}
      data-cartulary-theme={themeId}
      data-reading-profile={readingProfileAttribute}
      data-testid={appRouteTestId("app-shell")}
      style={rootPageStyle}
    >
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
              onRefreshShell={refreshCurrentShell}
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
              bootstrapState={appBootstrapState}
              error={error}
              directory={directory}
              onOpenIncident={openIncident}
              statusText={landingStatusText}
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

const utilityPanelStyle = {
  width: "min(78rem, 100%)",
  margin: "2rem auto",
  padding: "2rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-panel)",
  border: "var(--ct-border-hairline)",
};

const workbookFrameStyle = {
  display: "grid",
  gridTemplateRows: "minmax(0, 1fr)",
  blockSize: "100%",
  minBlockSize: 0,
  minWidth: 0,
  overflow: "hidden",
};

const landingHeroStyle = {
  marginBottom: "1.5rem",
};

const landingEyebrowStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.18em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-accent)",
};

const landingHeadlineStyle = {
  margin: "0.45rem 0 0",
  fontSize: "clamp(2rem, 4vw, 3rem)",
  lineHeight: 1.05,
};

const landingBodyStyle = {
  margin: "0.9rem 0 0",
  maxWidth: "42rem",
  lineHeight: 1.6,
  color: "var(--ct-colors-ink-muted)",
};

const routeLoadingStyle = {
  margin: "1rem 0 0",
  padding: "1rem 1.2rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
};
