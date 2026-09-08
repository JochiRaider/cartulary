import { lazy, type ReactNode, type RefObject, Suspense } from "react";
import type { IncidentControlsSection } from "../../app/landingAdminTypes";
import type { WorkbookImportController } from "../../imports/WorkbookImportController";
import type {
  WorkbookIncidentControlsMenuItem,
  WorkbookIncidentControlsRendererProps,
  WorkbookIncidentRole,
} from "../../shared/workbookShellContracts";
import { IncidentControlsDrawer } from "./IncidentControlsDrawer";

const LazyImportAssistantFeature = lazy(async () => {
  const feature = await import("../features/ImportAssistantFeature");
  return { default: feature.ImportAssistantFeature };
});

type WorkbookIncidentControlsPresentationProps = {
  readonly onIncidentResourceAccepted?: WorkbookIncidentControlsRendererProps["onIncidentResourceAccepted"];
  readonly density?: WorkbookIncidentControlsRendererProps["density"];
  readonly onAuthorizationRecovered?: WorkbookIncidentControlsRendererProps["onAuthorizationRecovered"];
  readonly activeMenuItem: WorkbookIncidentControlsMenuItem;
  readonly apiBase: string | undefined;
  readonly importController: WorkbookImportController;
  readonly closeButtonRef: RefObject<HTMLButtonElement | null>;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly importAssistantAvailable: boolean;
  readonly incidentId: string;
  readonly onClose: (options: {
    readonly restoreTriggerFocus: boolean;
  }) => void;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly onNavigateToView: (viewSchemaId: string) => void;
  readonly onSessionRoleChange: () => Promise<void>;
  readonly renderIncidentControls:
    | ((props: WorkbookIncidentControlsRendererProps) => ReactNode)
    | undefined;
  readonly section: IncidentControlsSection | null;
};

/** Owns lazy support-surface selection inside the incident controls drawer. */
export function WorkbookIncidentControlsPresentation({
  density,
  onIncidentResourceAccepted,
  onAuthorizationRecovered,
  activeMenuItem,
  apiBase,
  importController,
  closeButtonRef,
  currentIncidentRole,
  importAssistantAvailable,
  incidentId,
  onClose,
  onIncidentAccessLost,
  onNavigateToView,
  onSessionRoleChange,
  renderIncidentControls,
  section,
}: WorkbookIncidentControlsPresentationProps) {
  if (section === null) return null;
  const content =
    section === "import-assistant" && importAssistantAvailable ? (
      <Suspense fallback={<p role="status">Loading import assistant…</p>}>
        <LazyImportAssistantFeature
          density={density}
          controller={importController}
          onNavigateToView={onNavigateToView}
        />
      </Suspense>
    ) : (
      (renderIncidentControls?.({
        density,
        onIncidentResourceAccepted,
        onAuthorizationRecovered,
        activeSection: section,
        apiBase,
        currentIncidentRole,
        incidentId,
        onIncidentAccessLost,
        onSessionRoleChange,
      }) ?? null)
    );
  return (
    <IncidentControlsDrawer
      activeMenuItem={activeMenuItem}
      closeButtonRef={closeButtonRef}
      onClose={onClose}
    >
      {content}
    </IncidentControlsDrawer>
  );
}
