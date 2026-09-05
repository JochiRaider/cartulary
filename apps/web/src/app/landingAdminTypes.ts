import type { ReactNode, RefObject } from "react";
import type { APIError } from "../services/browserApi";
import type { ListAdministrativeAuditEventsResponse } from "./api/publicHttpTypes";
import type { IncidentCreationBinding } from "./incidentCreationModel";
import type {
  IncidentDirectoryController,
  IncidentDirectoryState,
} from "./incidentDirectoryModel";

export type IncidentStatusFilter = "active" | "all" | "closed";

export type AppBootstrapState =
  | "loading"
  | "anonymous"
  | "authenticated"
  | "forbidden"
  | "revoked"
  | "public_error_envelope";

export type LandingRefreshState = "idle" | "loading" | "failed";

export type LandingAdminPanelId =
  | "account-appearance"
  | "account-profile"
  | "account-security"
  | "administrative-audit"
  | "deployment-users"
  | "incident-import"
  | "incidents"
  | "reference-packs";

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

export type LandingAdminPanelDescriptor = {
  description: string;
  group: "account" | "deployment" | "primary";
  label: string;
  token: LandingAdminPanelId;
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
  availablePanels: ReadonlyArray<LandingAdminPanelDescriptor>;
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
  bootstrapState: AppBootstrapState;
  creation: IncidentCreationBinding;
  error: APIError | null;
  directory: {
    controller: IncidentDirectoryController;
    state: IncidentDirectoryState;
  };
  onOpenIncident: (incidentId: string) => void;
  statusText: string;
};

export type AdministrativeAuditEvent =
  ListAdministrativeAuditEventsResponse["data"]["audit_events"][number];
