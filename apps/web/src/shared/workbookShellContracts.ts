import type { IncidentControlsSection } from "../app/landingAdminTypes";

export type WorkbookIncidentRole =
  | "viewer"
  | "editor"
  | "reviewer"
  | "admin"
  | "";

export type WorkbookDensityMode = "compact" | "default" | "comfortable";

export type WorkbookIncidentSnapshot = {
  readonly closed_at?: string | null;
  readonly current_phase: string | null;
  readonly description: string | null;
  readonly incident_id: string;
  readonly incident_key: string;
  readonly incident_version: number;
  readonly primary_external_case_ref: string | null;
  readonly severity: string | null;
  readonly status?: "active" | "closed";
  readonly title: string;
  readonly tlp: string | null;
};

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
  readonly activeSection: IncidentControlsSection;
  readonly apiBase?: string | undefined;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly incidentId: string;
  readonly onIncidentAccessLost?: (() => void) | undefined;
  readonly onIncidentSnapshot?:
    | ((incident: WorkbookIncidentSnapshot) => void)
    | undefined;
  readonly onSessionRoleChange?: (() => Promise<void> | void) | undefined;
};
