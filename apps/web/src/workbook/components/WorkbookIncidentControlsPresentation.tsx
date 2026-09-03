import { lazy, type ReactNode, type RefObject, Suspense } from "react";
import type { IncidentControlsSection } from "../../app/landingAdminTypes";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import type {
  WorkbookIncidentControlsMenuItem,
  WorkbookIncidentControlsRendererProps,
  WorkbookIncidentRole,
  WorkbookIncidentSnapshot,
} from "../../shared/workbookShellContracts";
import { IncidentControlsDrawer } from "./IncidentControlsDrawer";

const LazyImportAssistantFeature = lazy(async () => {
  const feature = await import("../features/ImportAssistantFeature");
  return { default: feature.ImportAssistantFeature };
});

type WorkbookIncidentControlsPresentationProps = {
  readonly activeMenuItem: WorkbookIncidentControlsMenuItem;
  readonly apiBase: string | undefined;
  readonly availability: ExtensionAvailabilityController;
  readonly closeButtonRef: RefObject<HTMLButtonElement | null>;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly importAssistantAvailable: boolean;
  readonly incidentId: string;
  readonly onClose: (options: {
    readonly restoreTriggerFocus: boolean;
  }) => void;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly onIncidentSnapshot:
    | ((incident: WorkbookIncidentSnapshot) => void)
    | undefined;
  readonly onNavigateToView: (viewSchemaId: string) => void;
  readonly onSessionRoleChange: () => Promise<void>;
  readonly renderIncidentControls:
    | ((props: WorkbookIncidentControlsRendererProps) => ReactNode)
    | undefined;
  readonly section: IncidentControlsSection | null;
};

/** Owns lazy support-surface selection inside the incident controls drawer. */
export function WorkbookIncidentControlsPresentation({
  activeMenuItem,
  apiBase,
  availability,
  closeButtonRef,
  currentIncidentRole,
  importAssistantAvailable,
  incidentId,
  onClose,
  onIncidentAccessLost,
  onIncidentSnapshot,
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
          apiBase={apiBase}
          availability={availability}
          currentIncidentRole={currentIncidentRole}
          incidentId={incidentId}
          onNavigateToView={onNavigateToView}
        />
      </Suspense>
    ) : (
      (renderIncidentControls?.({
        activeSection: section,
        apiBase,
        currentIncidentRole,
        incidentId,
        onIncidentAccessLost,
        onIncidentSnapshot,
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
