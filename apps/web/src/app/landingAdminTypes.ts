import type { ReactNode, RefObject } from "react";
import type { ListAdministrativeAuditEventsResponse } from "./api/publicHttpTypes";
import type { IncidentCreationBinding } from "./incidentCreationModel";
import type {
  IncidentDirectoryController,
  IncidentDirectoryState,
} from "./incidentDirectoryModel";

export type IncidentStatusFilter = "active" | "all" | "closed";

export type IncidentControlsSection =
  | "import-assistant"
  | "incident-fields"
  | "membership-audit"
  | "memberships"
  | "summary";

export type IncidentControlsLoadState =
  | "loading"
  | "partial"
  | "synced"
  | "unavailable";

export type DeploymentPanelDescriptor = {
  description: string;
  label: string;
  token: DeploymentAdministrationPanelToken;
};

export type AccountSettingsPanelToken =
  | "account-appearance"
  | "account-profile"
  | "account-security";

export type DeploymentAdministrationPanelToken =
  | "administrative-audit"
  | "deployment-users"
  | "incident-import"
  | "reference-packs";

export type LandingAdminShellProps = {
  headingRef?: RefObject<HTMLHeadingElement | null>;
  accountMenu: ReactNode;
  activePanel: DeploymentAdministrationPanelToken;
  availablePanels: ReadonlyArray<DeploymentPanelDescriptor>;
  children: ReactNode;
  currentUserLabel: string;
  onActivePanelChange: (panel: DeploymentAdministrationPanelToken) => void;
  statusText: string;
};

export type IncidentDirectoryShellProps = {
  headingRef?: RefObject<HTMLHeadingElement | null>;
  accountMenu: ReactNode;
  children: ReactNode;
  currentUserLabel: string;
  statusText: string;
};

export type AccountApplicationMenuProps = {
  subjectKey?: string;
  triggerFocusRef?: RefObject<HTMLButtonElement | null>;
  canOpenDeploymentAdministration: boolean;
  currentContext: "deployment-administration" | "incidents" | "workbook";
  currentUserLabel: string;
  currentIncidentRole?: string | null | undefined;
  incidentControls?:
    | {
        readonly activeSection: IncidentControlsSection;
        readonly items: ReadonlyArray<{
          readonly description: string;
          readonly label: string;
          readonly section: IncidentControlsSection;
        }>;
        readonly onSelectSection: (
          section: IncidentControlsSection,
          returnFocusTarget?: HTMLElement | null,
        ) => void;
      }
    | undefined;
  onOpenAccountSettings: (panel: AccountSettingsPanelToken) => void;
  onOpenDeploymentAdministration: () => void;
  onOpenIncidentDirectory: () => void;
  triggerTestId?: string | undefined;
};

export type IncidentLandingProps = {
  creation: IncidentCreationBinding;
  directory: {
    controller: IncidentDirectoryController;
    state: IncidentDirectoryState;
  };
  onOpenIncident: (incidentId: string) => void;
  notice: string | null;
};

export type AdministrativeAuditEvent =
  ListAdministrativeAuditEventsResponse["data"]["audit_events"][number];
