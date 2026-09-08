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
  useSyncExternalStore,
} from "react";
import type {
  WorkbookAccountApplicationMenuProps,
  WorkbookAccountModel,
} from "../shared/workbookShellContracts";
import { WorkbookMutationRuntimeRegistry } from "../workbook/runtime/WorkbookMutationRuntimeRegistry";
import { AccountApplicationMenu } from "./AccountApplicationMenu";
import { AccountSecurityPanel } from "./AccountSecurityPanel";
import { AccountSettingsDialog } from "./AccountSettingsDialog";
import {
  AccountAppearancePanel,
  AccountProfilePanel,
} from "./AccountSettingsPanels";
import { AuthGateway } from "./AuthGateway";
import { AccountSecurityController } from "./accountSecurityModel";
import { AccountSettingsController } from "./accountSettingsModel";
import type { AdministrativeAuditController } from "./administrativeAuditController";
import type {
  ExtensionProfileResource,
  SessionData,
} from "./api/publicHttpTypes";
import { AppSessionController } from "./appSessionController";
import { AuthenticationController } from "./authenticationModel";
import { AdministrativeAuditPanel } from "./DeploymentAuditPanel";
import { DeploymentUsersPanel } from "./DeploymentUsersPanel";
import { DeploymentUsersController } from "./deploymentUsersModel";
import { IncidentAdminPanel } from "./IncidentAdminPanel";
import { IncidentImportPanel } from "./IncidentImportPanel";
import { IncidentLanding } from "./IncidentLanding";
import { IncidentMembershipAuditFeature } from "./IncidentMembershipAuditPanel";
import {
  IncidentMembershipDepartureDialog,
  IncidentMembershipManagementFeature,
} from "./IncidentMembershipManagementPanel";
import type { IncidentCreationController } from "./incidentCreationModel";
import {
  type IncidentDirectoryController,
  incidentDirectoryStatusText,
} from "./incidentDirectoryModel";
import type { IncidentImportController } from "./incidentImportModel";
import type { IncidentMembershipAuditController } from "./incidentMembershipAuditController";
import type { IncidentMembershipManagementController } from "./incidentMembershipManagementController";
import {
  IncidentDirectoryShell,
  LandingAdminShell,
} from "./LandingAdminLayout";
import type {
  AccountSettingsPanelToken,
  DeploymentAdministrationPanelToken,
  DeploymentPanelDescriptor,
} from "./landingAdminTypes";
import { ReferencePackAdminPanel } from "./ReferencePackAdminPanel";
import type { ReferencePackAdminController } from "./referencePackAdminController";
import { readAppRouteState } from "./routeState";
import { useAdministrativeAudit } from "./useAdministrativeAudit";
import { useAppRouteRuntime } from "./useAppRouteRuntime";
import { useAppSession } from "./useAppSession";
import { useIncidentCreation } from "./useIncidentCreation";
import { useIncidentDirectory } from "./useIncidentDirectory";
import { useIncidentImport } from "./useIncidentImport";
import { useIncidentMembershipAudit } from "./useIncidentMembershipAudit";
import { useIncidentMembershipManagement } from "./useIncidentMembershipManagement";
import { useReferencePackAdmin } from "./useReferencePackAdmin";

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
  readonly authNavigation?:
    | import("./authenticationModel").EnterpriseAuthNavigation
    | undefined;
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

export function App({
  readingProfile = "default",
  themeId,
  authNavigation,
}: AppProps = {}) {
  const membershipManagementRef =
    useRef<IncidentMembershipManagementController | null>(null);
  const deploymentUsersRef = useRef<DeploymentUsersController | null>(null);
  const membershipAuditRef = useRef<IncidentMembershipAuditController | null>(
    null,
  );
  const auditControllerRef = useRef<AdministrativeAuditController | null>(null);
  const accountEditingRef = useRef<AccountSettingsController | null>(null);
  const creationControllerRef = useRef<IncidentCreationController | null>(null);
  const importControllerRef = useRef<IncidentImportController | null>(null);
  const referencePackControllerRef =
    useRef<ReferencePackAdminController | null>(null);
  const directoryControllerRef = useRef<IncidentDirectoryController | null>(
    null,
  );
  const sessionControllerRef = useRef<AppSessionController | null>(null);
  const { commitRoute, route, routeRef } = useAppRouteRuntime({
    hasPendingEdits: () =>
      (membershipManagementRef.current?.hasDepartureWork() ?? false) ||
      (deploymentUsersRef.current?.hasDirtyDraft() ?? false),
    requestLeave: () =>
      membershipManagementRef.current?.hasDepartureWork()
        ? membershipManagementRef.current.requestLeave()
        : (deploymentUsersRef.current?.requestLeave() ?? Promise.resolve(true)),
    beforeCommit: (next) => {
      if (next.incidentId !== routeRef.current.incidentId) {
        membershipAuditRef.current?.retire();
        membershipManagementRef.current?.retire();
      }
      auditControllerRef.current?.setActive(false);
      importControllerRef.current?.setActive(false);
      referencePackControllerRef.current?.setActive(false);
      creationControllerRef.current?.leaveSurface();
      if (next.incidentId !== "" || next.deploymentAdministration)
        directoryControllerRef.current?.setActive(false);
      sessionControllerRef.current?.navigationChanged();
    },
  });
  const workbookMutationRuntimeRegistry = useMemo(
    () => new WorkbookMutationRuntimeRegistry(),
    [],
  );
  const authenticationRef = useRef<AuthenticationController | null>(null);
  const securityRef = useRef<AccountSecurityController | null>(null);
  const [sessionController] = useState(
    () =>
      new AppSessionController({
        retireLifetime: (lifetime) => {
          membershipAuditRef.current?.retire();
          membershipManagementRef.current?.retire();
          auditControllerRef.current?.retire();
          accountEditingRef.current?.retireLifetime();
          authenticationRef.current?.retire();
          securityRef.current?.retire();
          deploymentUsersRef.current?.retire();
          importControllerRef.current?.retire();
          referencePackControllerRef.current?.retire();
          workbookMutationRuntimeRegistry.sessionUnavailable();
          creationControllerRef.current?.setSession(lifetime);
          directoryControllerRef.current?.setSession(lifetime);
        },
        replaceAccount: () => workbookMutationRuntimeRegistry.replaceAccount(),
        capabilitiesReduced: () => {
          auditControllerRef.current?.retire();
          deploymentUsersRef.current?.retire();
          importControllerRef.current?.retire();
          referencePackControllerRef.current?.retire();
        },
      }),
  );
  sessionControllerRef.current = sessionController;
  const [authentication] = useState(
    () =>
      new AuthenticationController(
        () => {
          const revision = sessionController.getSnapshot().revision;
          const current = () =>
            sessionController.getSnapshot().revision === revision &&
            sessionController.getSnapshot().session === null;
          return {
            current,
            admitTransport: sessionController.reserveAuthenticationTransport,
            canAuthenticate: () =>
              !sessionController.getSnapshot().authenticationTransportPending,
            authenticated: (next, signal, flowCurrent) => {
              if (
                current() &&
                !signal.aborted &&
                flowCurrent() &&
                sessionController.authenticationCompleted(next, revision)
              )
                setAuthPrompt(defaultAuthPrompt);
            },
            inspectSession: (signal, flowCurrent) =>
              sessionController.confirmAuthentication(
                revision,
                signal,
                () => current() && flowCurrent(),
              ),
          };
        },
        authNavigation ?? {
          assign: (url) => window.location.assign(url),
          returnTo: () =>
            `${window.location.pathname}${window.location.search}` || "/",
        },
      ),
  );
  authenticationRef.current = authentication;
  const [security] = useState(
    () =>
      new AccountSecurityController(() => {
        const lifetime = sessionController.getSnapshot().lifetime;
        const current = () =>
          lifetime !== null &&
          sessionController.getSnapshot().lifetime === lifetime;
        return {
          actor: sessionController.getSnapshot().session?.user_id ?? "",
          current,
          admitLogout: sessionController.reserveAuthenticationTransport,
          event: async (event, signal, operationCurrent) => {
            if (lifetime === null || !current() || !operationCurrent()) return;
            if (event.kind === "resource_refresh") {
              const result = await sessionController.observeOperationSession(
                lifetime,
                signal,
                () => current() && operationCurrent(),
              );
              if (lifetime === null || !current() || !operationCurrent())
                return;
              if (result.kind !== "accepted")
                throw new Error("Account session refresh unavailable");
              sessionController.refreshResources();
              return;
            }
            setAuthPrompt(event.message);
            setLandingNotice(null);
            if (event.kind === "logout_confirmed")
              sessionController.logoutConfirmed();
            else if (event.kind === "session_lost")
              sessionController.sessionLost();
            else sessionController.credentialsRevoked();
          },
        };
      }),
  );
  securityRef.current = security;
  const [deploymentUsers] = useState(
    () =>
      new DeploymentUsersController({
        identity: () => {
          const current = sessionController.getSnapshot();
          return current.session && current.lifetime
            ? {
                actor: current.session.user_id,
                lifetime: current.lifetime,
                admin: current.session.is_deployment_admin,
              }
            : null;
        },
        subscribe: sessionController.subscribe,
        refresh: async (identity, signal, current) => {
          const result = await sessionController.observeOperationSession(
            identity.lifetime,
            signal,
            current,
          );
          if (!current()) return;
          if (result.kind !== "accepted")
            throw new Error("Session refresh unavailable");
          sessionController.refreshResources();
        },
        revoked: (identity, message) => {
          if (sessionController.getSnapshot().lifetime !== identity.lifetime)
            return;
          sessionController.credentialsRevoked();
          setAuthPrompt(message);
        },
        lost: (identity) => {
          if (sessionController.getSnapshot().lifetime === identity.lifetime)
            sessionController.sessionLost();
        },
      }),
  );
  deploymentUsersRef.current = deploymentUsers;
  useEffect(() => {
    const preventUnload = (event: BeforeUnloadEvent) => {
      if (
        !deploymentUsers.hasDirtyDraft() &&
        !membershipManagementRef.current?.hasDepartureWork()
      )
        return;
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", preventUnload);
    return () => window.removeEventListener("beforeunload", preventUnload);
  }, [deploymentUsers]);
  const changeDeploymentPanel = (next: DeploymentAdministrationPanelToken) => {
    if (next === activeDeploymentPanel) return;
    if (!deploymentUsers.hasDirtyDraft()) {
      auditControllerRef.current?.setActive(false);
      importControllerRef.current?.setActive(false);
      referencePackControllerRef.current?.setActive(false);
      setActiveDeploymentPanel(next);
      return;
    }
    const lifetime = sessionController.getSnapshot().lifetime;
    void deploymentUsers.requestLeave().then((accepted) => {
      if (accepted && lifetime === sessionController.getSnapshot().lifetime) {
        auditControllerRef.current?.setActive(false);
        importControllerRef.current?.setActive(false);
        referencePackControllerRef.current?.setActive(false);
        setActiveDeploymentPanel(next);
      }
    });
  };

  useEffect(() => {
    deploymentUsers.start();
    return deploymentUsers.stop;
  }, [deploymentUsers]);
  const [accountEditing] = useState(
    () => new AccountSettingsController(sessionController),
  );
  accountEditingRef.current = accountEditing;
  const sessionSnapshot = useAppSession(sessionController, () => {
    membershipAuditRef.current?.dispose();
    membershipManagementRef.current?.dispose();
    workbookMutationRuntimeRegistry.dispose();
    accountEditing.dispose();
    authentication.dispose();
    security.dispose();
    deploymentUsers.dispose();
    importControllerRef.current?.dispose();
    referencePackControllerRef.current?.dispose();
  });
  useEffect(() => {
    accountEditing.start();
    return accountEditing.stop;
  }, [accountEditing]);
  const accountEditingSnapshot = useSyncExternalStore(
    accountEditing.subscribe,
    accountEditing.getSnapshot,
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
  const navigationHeadingRef = useRef<HTMLHeadingElement>(null);
  const navigationFocusRequestRef = useRef<{
    readonly destination: "incidents" | "deployment-administration";
    readonly accountId: string;
    readonly originIdentity: string;
  } | null>(null);
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
    setAccountSettingsPanel(null);
  }, [accountNavigationIdentity]);

  const closeAccountSettings = useCallback(() => {
    setAccountSettingsPanel(null);
  }, []);

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
  const handleSessionLost = useCallback(() => {
    if (sessionController.getSnapshot().lifetime !== sessionSnapshot.lifetime)
      return;
    sessionController.sessionLost();
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
  }, [commitRoute, route.deploymentAdministration, session]);
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
  const audit = useAdministrativeAudit({
    authority:
      session?.is_deployment_admin && sessionSnapshot.lifetime
        ? { actorId: session.user_id, lifetime: sessionSnapshot.lifetime }
        : null,
    active:
      session?.is_deployment_admin === true &&
      route.deploymentAdministration &&
      activeDeploymentPanel === "administrative-audit",
    isCurrent: (authority) => {
      const current = sessionController.getSnapshot();
      return (
        current.lifetime === authority.lifetime &&
        current.session?.user_id === authority.actorId &&
        current.session.is_deployment_admin
      );
    },
    confirmAccess: async (authority, signal, current) => {
      const result = await sessionController.observeOperationSession(
        authority.lifetime,
        signal,
        current,
      );
      if (result.kind !== "accepted") return { kind: result.kind };
      if (!current()) return { kind: "cancelled" };
      return {
        kind:
          result.session.user_id === authority.actorId &&
          result.session.is_deployment_admin
            ? "authorized"
            : "access_lost",
      };
    },
    authorizationFailed: (status, authority) => {
      if (sessionController.getSnapshot().lifetime !== authority.lifetime)
        return;
      if (status === 401) {
        handleSessionLost();
        return;
      }
      setLandingNotice(
        "Deployment administration requires deployment admin access.",
      );
      commitRoute(
        { incidentId: "", deploymentAdministration: false },
        "replace",
      );
      void sessionController.refreshSession();
    },
  });
  auditControllerRef.current = audit;
  const referencePackAllowed =
    session?.is_deployment_admin === true &&
    extensionClaimed(extensionProfiles, "reference_pack");
  const referencePacks = useReferencePackAdmin({
    authority:
      referencePackAllowed && sessionSnapshot.lifetime && session
        ? { lifetime: sessionSnapshot.lifetime, actorId: session.user_id }
        : null,
    active:
      referencePackAllowed &&
      route.deploymentAdministration &&
      activeDeploymentPanel === "reference-packs",
    isCurrent: (authority) => {
      const current = sessionController.getSnapshot();
      return (
        current.lifetime === authority.lifetime &&
        current.session?.user_id === authority.actorId &&
        current.session.is_deployment_admin &&
        current.extensions.kind === "ready" &&
        extensionClaimed(current.extensions.value, "reference_pack")
      );
    },
    authorizationFailed: (status) => {
      if (status === 401) handleSessionLost();
      else void sessionController.refreshSession();
    },
    confirmAccess: async (authority, signal, current) => {
      const result = await sessionController.observeOperationSession(
        authority.lifetime,
        signal,
        current,
      );
      if (result.kind !== "accepted") return { kind: result.kind };
      if (
        !current() ||
        sessionController.getSnapshot().lifetime !== authority.lifetime
      )
        return { kind: "cancelled" };
      if (
        result.session.user_id !== authority.actorId ||
        !result.session.is_deployment_admin
      )
        return { kind: "access_lost" };
      const discovery = await sessionController.observeOperationExtensions(
        authority.lifetime,
        signal,
        current,
      );
      if (discovery.kind !== "accepted") return discovery;
      if (!current()) return { kind: "cancelled" };
      return {
        kind: extensionClaimed(discovery.profiles, "reference_pack")
          ? "authorized"
          : "access_lost",
      };
    },
  });
  referencePackControllerRef.current = referencePacks;
  const importAllowed =
    session?.is_deployment_admin === true &&
    extensionClaimed(extensionProfiles, "incident_portability");
  const incidentImport = useIncidentImport({
    authority:
      importAllowed && sessionSnapshot.lifetime && session
        ? { lifetime: sessionSnapshot.lifetime, actorId: session.user_id }
        : null,
    active:
      importAllowed &&
      route.deploymentAdministration &&
      activeDeploymentPanel === "incident-import",
    isCurrent: (authority) => {
      const current = sessionController.getSnapshot();
      return (
        current.lifetime === authority.lifetime &&
        current.session?.user_id === authority.actorId &&
        current.session.is_deployment_admin &&
        current.extensions.kind === "ready" &&
        extensionClaimed(current.extensions.value, "incident_portability")
      );
    },
    authorizationFailed: (status) => {
      if (status === 401) handleSessionLost();
    },
    confirmAccess: async (authority, signal, current) => {
      const result = await sessionController.observeOperationSession(
        authority.lifetime,
        signal,
        current,
      );
      if (result.kind !== "accepted") return { kind: result.kind };
      if (
        !current() ||
        sessionController.getSnapshot().lifetime !== authority.lifetime
      )
        return { kind: "cancelled" };
      if (
        result.session.user_id !== authority.actorId ||
        !result.session.is_deployment_admin
      )
        return { kind: "access_lost" };
      const discovery = await sessionController.observeOperationExtensions(
        authority.lifetime,
        signal,
        current,
      );
      if (discovery.kind !== "accepted") return discovery;
      if (!current()) return { kind: "cancelled" };
      return extensionClaimed(discovery.profiles, "incident_portability")
        ? { kind: "authorized", authority }
        : { kind: "access_lost" };
    },
    openIncident: async (incidentId, signal, canNavigate) => {
      const result = await sessionController
        .recoveryPort(() => canNavigate())
        .recover({ incidentId, signal });
      if (!canNavigate() || result.kind === "cancelled") return "cancelled";
      if (result.kind === "access_lost") return "access_lost";
      if (result.kind !== "authorized") return "unavailable";
      openIncident(incidentId);
      return "opened";
    },
  });
  importControllerRef.current = incidentImport.controller;
  const handleIncidentAccessLost = useCallback(() => {
    if (readAppRouteState().incidentId !== route.incidentId) return;
    membershipManagementRef.current?.retire();
    directoryControllerRef.current?.setActive(false);
    navigationFocusRequestRef.current = {
      destination: "incidents",
      accountId: sessionRef.current?.user_id ?? "",
      originIdentity: accountNavigationIdentity,
    };
    setLandingNotice(accessLostLandingNotice);
    commitRoute({ incidentId: "", deploymentAdministration: false }, "replace");
  }, [accountNavigationIdentity, commitRoute, route.incidentId]);

  const membershipAudit = useIncidentMembershipAudit({
    sessionController,
    recovery: workbookAuthorizationRecovery,
    currentIncidentId: () => routeRef.current.incidentId,
    onIncidentAccessLost: handleIncidentAccessLost,
    onSessionLost: handleSessionLost,
  });
  membershipAuditRef.current = membershipAudit.controller;
  const membershipManagement = useIncidentMembershipManagement({
    sessionController,
    recovery: workbookAuthorizationRecovery,
    currentIncidentId: () => routeRef.current.incidentId,
    onIncidentAccessLost: handleIncidentAccessLost,
    onSessionLost: handleSessionLost,
  });
  membershipManagementRef.current = membershipManagement.controller;

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
          const lifetime = sessionController.getSnapshot().lifetime;
          const open = () => {
            if (sessionController.getSnapshot().lifetime !== lifetime) return;
            creationControllerRef.current?.leaveSurface();
            setAccountSettingsPanel(panel);
          };
          if (deploymentUsersRef.current?.hasDirtyDraft()) {
            void deploymentUsersRef.current.requestLeave().then((accepted) => {
              if (accepted) open();
            });
          } else open();
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
      sessionController,
      session?.is_deployment_admin,
      session?.user_id,
      route.incidentId,
    ],
  );

  const renderAccountSettings = useCallback(() => {
    if (accountSettingsPanel === null || session === null) {
      return null;
    }
    return (
      <AccountSettingsDialog
        panel={accountSettingsPanel}
        onClose={closeAccountSettings}
        onSelect={setAccountSettingsPanel}
        fallbackFocusRef={accountMenuTriggerRef}
      >
        {accountSettingsPanel === "account-profile" ? (
          <AccountProfilePanel
            controller={accountEditing}
            lifetime={sessionSnapshot.lifetime}
            state={accountEditingSnapshot.profile}
          />
        ) : null}
        {accountSettingsPanel === "account-appearance" ? (
          <AccountAppearancePanel
            controller={accountEditing}
            lifetime={sessionSnapshot.lifetime}
            state={accountEditingSnapshot.appearance}
          />
        ) : null}
        {accountSettingsPanel === "account-security" ? (
          <AccountSecurityPanel controller={security} />
        ) : null}
      </AccountSettingsDialog>
    );
  }, [
    accountEditing,
    accountEditingSnapshot,
    sessionSnapshot.lifetime,
    accountSettingsPanel,
    closeAccountSettings,
    security,
    session,
  ]);

  const resourceStatus = (
    <>
      {accountSettingsPanel !== "account-appearance" &&
      (sessionSnapshot.preferences.kind === "failed" ||
        (sessionSnapshot.preferences.kind === "ready" &&
          sessionSnapshot.preferences.refreshError !== undefined)) ? (
        <p role="status">
          {sessionSnapshot.preferences.kind === "ready"
            ? "Appearance preferences could not refresh; keeping the last saved presentation."
            : "Appearance preferences unavailable; using the default presentation."}{" "}
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
              onIncidentControlsSectionChange={(section) => {
                if (section !== "memberships")
                  membershipManagement.controller.setActive(false);
                if (section !== "membership-audit")
                  membershipAudit.controller.setActive(false);
              }}
              renderIncidentControls={(props) =>
                props.activeSection === "membership-audit" ? (
                  <IncidentMembershipAuditFeature
                    {...props}
                    controller={membershipAudit.controller}
                    bindSurface={membershipAudit.bindSurface}
                  />
                ) : props.activeSection === "memberships" ? (
                  <IncidentMembershipManagementFeature
                    {...props}
                    controller={membershipManagement.controller}
                    bindSurface={membershipManagement.bindSurface}
                  />
                ) : (
                  <IncidentAdminPanel {...props} />
                )
              }
              mutationRuntimeRegistry={workbookMutationRuntimeRegistry}
            />
          </Suspense>
        </section>
        {renderAccountSettings()}
        <IncidentMembershipDepartureDialog
          controller={membershipManagement.controller}
          kind="route"
          fallbackFocusRef={accountMenuTriggerRef}
        />
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
        controller={authentication}
        bootstrapState={
          sessionSnapshot.state === "unresolved" || sessionSnapshot.observing
            ? "loading"
            : sessionSnapshot.ended
              ? "revoked"
              : "anonymous"
        }
        message={authPrompt}
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
          onActivePanelChange={changeDeploymentPanel}
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
              fallbackFocusRef={accountMenuTriggerRef}
              enterpriseAuthClaimed={extensionClaimed(
                extensionProfiles,
                "enterprise_authentication",
              )}
              controller={deploymentUsers}
              active={activeDeploymentPanel === "deployment-users"}
            />
          </section>
          <section
            id={landingAdminPanelTestId("administrative-audit")}
            aria-labelledby={landingAdminMenuItemTestId("administrative-audit")}
            data-testid={landingAdminPanelTestId("administrative-audit")}
            hidden={activeDeploymentPanel !== "administrative-audit"}
            style={landingAdminPanelRegionStyle}
          >
            <AdministrativeAuditPanel
              controller={audit}
              active={activeDeploymentPanel === "administrative-audit"}
            />
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
                controller={referencePacks}
                active={activeDeploymentPanel === "reference-packs"}
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
              <IncidentImportPanel binding={incidentImport} />
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
