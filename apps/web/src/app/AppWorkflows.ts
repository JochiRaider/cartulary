import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import type { IncidentResource } from "../shared/incidentResource";
import type { AppSessionController } from "./appSessionController";
import { createIncidentLifecycleWorkflow } from "./createIncidentLifecycleWorkflow";
import { createIncidentMembershipAuditWorkflow } from "./createIncidentMembershipAuditWorkflow";
import { createIncidentMembershipManagementWorkflow } from "./createIncidentMembershipManagementWorkflow";
import { createIncidentMetadataWorkflow } from "./createIncidentMetadataWorkflow";
import { createNetworkFlowImportWorkflow } from "./createNetworkFlowImportWorkflow";
import { createWorkbookImportWorkflow } from "./createWorkbookImportWorkflow";
import { createWorkbookPreferencesWorkflow } from "./createWorkbookPreferencesWorkflow";
import { createWorkbookSavedViewsWorkflow } from "./createWorkbookSavedViewsWorkflow";
import type { IncidentResourceAuthority } from "./incidentResourceController";

export type AppWorkflowOptions = {
  readonly sessionController: AppSessionController;
  readonly recovery: AuthorizationRecoveryPort;
  readonly currentIncidentId: () => string;
  readonly onIncidentAccessLost: () => void;
  readonly onSessionLost: () => void;
  readonly onResourceAccepted: (
    resource: IncidentResource,
    authority: IncidentResourceAuthority,
  ) => void;
};
type WorkflowName =
  | "metadata"
  | "lifecycle"
  | "membershipManagement"
  | "membershipAudit"
  | "savedViews"
  | "preferences"
  | "workbookImport"
  | "networkFlowImport";
type Contribution = {
  synchronize: () => void;
  retire: () => void;
  dispose: () => void;
};

/** Fixed application membership. Feature policy remains in the named adapters. */
export function createAppWorkflows(options: () => AppWorkflowOptions) {
  let disposed = false;
  let retiring = false;
  let retired = false;
  let accessLoss: {
    source: AppSessionController;
    snapshot: ReturnType<AppSessionController["getSnapshot"]>;
    incident: string;
  } | null = null;
  const environment = () => ({
    ...options(),
    accepting: () => !disposed && !retiring,
    onIncidentAccessLost: () => {
      if (disposed || retiring) return;
      const source = options().sessionController;
      const snapshot = source.getSnapshot();
      const incident = options().currentIncidentId();
      if (
        accessLoss?.source === source &&
        accessLoss.snapshot === snapshot &&
        accessLoss.incident === incident
      )
        return;
      accessLoss = { source, snapshot, incident };
      options().onIncidentAccessLost();
    },
  });
  const members = {
    metadata: createIncidentMetadataWorkflow(environment),
    lifecycle: createIncidentLifecycleWorkflow(environment),
    membershipManagement:
      createIncidentMembershipManagementWorkflow(environment),
    membershipAudit: createIncidentMembershipAuditWorkflow(environment),
    savedViews: createWorkbookSavedViewsWorkflow(environment),
    preferences: createWorkbookPreferencesWorkflow(environment),
    workbookImport: createWorkbookImportWorkflow(environment),
    networkFlowImport: createNetworkFlowImportWorkflow(environment),
  } satisfies Record<WorkflowName, Contribution>;
  const contributions = Object.values(members);
  let synchronizing = false;
  let again = false;
  let generation = 0;
  let subscription = 0;
  let disconnect = () => {};
  const synchronize = () => {
    if (disposed || retiring) return;
    if (synchronizing) {
      again = true;
      return;
    }
    synchronizing = true;
    retired = false;
    try {
      do {
        again = false;
        const epoch = generation;
        const source = options().sessionController;
        const snapshot = source.getSnapshot();
        const incident = options().currentIncidentId();
        for (const member of contributions) {
          if (disposed || generation !== epoch) break;
          if (
            options().sessionController !== source ||
            source.getSnapshot() !== snapshot ||
            options().currentIncidentId() !== incident
          ) {
            again = true;
            break;
          }
          member.synchronize();
        }
      } while (again && !disposed);
    } finally {
      synchronizing = false;
    }
  };
  return {
    ...members,
    synchronize,
    connect: (source: AppSessionController) => {
      if (disposed) return () => {};
      disconnect();
      const token = ++subscription;
      const unsubscribe = source.subscribe(() => {
        if (
          !disposed &&
          token === subscription &&
          source === options().sessionController
        )
          synchronize();
      });
      synchronize();
      let released = false;
      const release = () => {
        if (released) return;
        released = true;
        if (token === subscription) ++subscription;
        unsubscribe();
      };
      disconnect = release;
      return release;
    },
    retire: () => {
      if (disposed || retiring || retired) return;
      retired = true;
      ++generation;
      retiring = true;
      try {
        for (const member of contributions) member.retire();
      } finally {
        retiring = false;
      }
    },
    dispose: () => {
      if (disposed) return;
      disposed = true;
      ++generation;
      ++subscription;
      disconnect();
      for (const member of contributions) member.dispose();
    },
  };
}
