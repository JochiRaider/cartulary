import type { IncidentControlsSection } from "../app/landingAdminTypes";
import type { IncidentResource } from "./incidentResource";

export type WorkbookIncidentRole =
  | "viewer"
  | "editor"
  | "reviewer"
  | "admin"
  | "";

export type WorkbookDensityMode = "compact" | "default" | "comfortable";

export type WorkbookIncidentControlsMenuItem = {
  readonly description: string;
  readonly label: string;
  readonly section: IncidentControlsSection;
};

export type WorkbookAccountApplicationMenuProps = {
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly incidentControls: {
    readonly activeSection: IncidentControlsSection;
    readonly items: readonly WorkbookIncidentControlsMenuItem[];
    readonly onSelectSection: (
      section: IncidentControlsSection,
      returnFocusTarget?: HTMLElement | null,
    ) => void;
  };
};

export type WorkbookAccountModel = {
  readonly display_name: string;
  readonly is_deployment_admin: boolean;
  readonly user_id: string;
};

export type WorkbookIncidentControlsRendererProps = {
  readonly onIncidentResourceAccepted?:
    | ((resource: IncidentResource) => void)
    | undefined;
  readonly density?: WorkbookDensityMode | undefined;
  readonly onAuthorizationRecovered?:
    | ((result: {
        readonly role: WorkbookIncidentRole;
        readonly userId: string;
      }) => void)
    | undefined;
  readonly activeSection: IncidentControlsSection;
  readonly apiBase?: string | undefined;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly incidentId: string;
  readonly onIncidentAccessLost?: (() => void) | undefined;
  readonly onSessionRoleChange?: (() => Promise<void> | void) | undefined;
};
