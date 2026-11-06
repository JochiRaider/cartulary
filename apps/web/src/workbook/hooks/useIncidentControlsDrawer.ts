import {
  type RefObject,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { IncidentControlsSection } from "../../app/landingAdminTypes";
import type {
  WorkbookAccountApplicationMenuProps,
  WorkbookIncidentControlsMenuItem,
} from "../../shared/workbookShellContracts";

const baseIncidentControlsMenuItems = [
  {
    section: "summary",
    label: "Summary and preferences",
    description: "Incident summary and workbook defaults",
  },
  {
    section: "incident-fields",
    label: "Promoted fields",
    description: "TLP, phase, and external case",
  },
  {
    section: "memberships",
    label: "Memberships",
    description: "Incident access and roles",
  },
  {
    section: "membership-audit",
    label: "Membership audit",
    description: "Incident membership changes",
  },
] as const satisfies readonly WorkbookIncidentControlsMenuItem[];

const importAssistantMenuItem = {
  section: "import-assistant",
  label: "Import workbook",
  description: "Discover, map, and apply CSV or XLSX data",
} as const satisfies WorkbookIncidentControlsMenuItem;

const reportingMenuItem = {
  section: "snapshot-reports",
  label: "Snapshots and reports",
  description: "Immutable snapshots and internal report drafts",
} as const satisfies WorkbookIncidentControlsMenuItem;

const defaultIncidentControlsMenuItem = baseIncidentControlsMenuItems[0];

function requireIncidentControlsMenuItem(
  section: IncidentControlsSection,
  items: readonly WorkbookIncidentControlsMenuItem[],
): WorkbookIncidentControlsMenuItem {
  return (
    items.find((item) => item.section === section) ??
    defaultIncidentControlsMenuItem
  );
}

export function useIncidentControlsDrawer(
  importAssistantAvailable = false,
  onSectionChange?: (section: IncidentControlsSection | null) => void,
  reportingAvailable = false,
  sourcesAvailable = false,
  fallbackFocusRef?: RefObject<HTMLElement | null>,
) {
  const sectionChange = useRef(onSectionChange);
  sectionChange.current = onSectionChange;
  const [drawerSection, setDrawerSection] =
    useState<IncidentControlsSection | null>(null);
  const [lastSection, setLastSection] =
    useState<IncidentControlsSection>("summary");
  const returnFocusTargetRef = useRef<HTMLElement | null>(null);
  const closeButtonRef = useRef<HTMLButtonElement | null>(null);
  const restoreFocusOnCloseRef = useRef(false);
  const previousDrawerSectionRef = useRef<IncidentControlsSection | null>(null);

  const closeDrawer = useCallback(
    (options: { readonly restoreTriggerFocus: boolean }) => {
      restoreFocusOnCloseRef.current = options.restoreTriggerFocus;
      sectionChange.current?.(null);
      setDrawerSection(null);
    },
    [],
  );

  const openDrawer = useCallback(
    (
      section: IncidentControlsSection,
      returnFocusTarget?: HTMLElement | null,
    ) => {
      returnFocusTargetRef.current = returnFocusTarget ?? null;
      sectionChange.current?.(section);
      restoreFocusOnCloseRef.current = false;
      setLastSection(section);
      setDrawerSection(section);
    },
    [],
  );

  useLayoutEffect(() => {
    const previousDrawerSection = previousDrawerSectionRef.current;
    previousDrawerSectionRef.current = drawerSection;
    if (drawerSection !== null) {
      closeButtonRef.current?.focus({ preventScroll: true });
      return;
    }
    if (previousDrawerSection === null || !restoreFocusOnCloseRef.current) {
      return;
    }
    restoreFocusOnCloseRef.current = false;
    const returnFocusTarget = returnFocusTargetRef.current;
    if (returnFocusTarget?.isConnected) {
      returnFocusTarget.focus({ preventScroll: true });
    } else if (fallbackFocusRef?.current?.isConnected) {
      fallbackFocusRef.current.focus({ preventScroll: true });
    }
  }, [drawerSection, fallbackFocusRef]);

  useEffect(() => {
    if (
      (!importAssistantAvailable &&
        drawerSection === importAssistantMenuItem.section) ||
      (!reportingAvailable && drawerSection === reportingMenuItem.section) ||
      (!sourcesAvailable && drawerSection === "committed-sources")
    ) {
      setDrawerSection(null);
      setLastSection("summary");
    }
  }, [
    drawerSection,
    importAssistantAvailable,
    reportingAvailable,
    sourcesAvailable,
  ]);

  const incidentControlsMenuItems = useMemo(
    () => [
      baseIncidentControlsMenuItems[0],
      ...(importAssistantAvailable ? [importAssistantMenuItem] : []),
      ...(reportingAvailable ? [reportingMenuItem] : []),
      ...(sourcesAvailable
        ? [
            {
              section: "committed-sources",
              label: "Original import sources",
              description: "Retained files, attribution, and mappings",
            } as const,
          ]
        : []),
      ...baseIncidentControlsMenuItems.slice(1),
    ],
    [importAssistantAvailable, reportingAvailable, sourcesAvailable],
  );

  const accountIncidentControls = useMemo<
    WorkbookAccountApplicationMenuProps["incidentControls"]
  >(
    () => ({
      activeSection: lastSection,
      items: incidentControlsMenuItems,
      onSelectSection: openDrawer,
    }),
    [incidentControlsMenuItems, lastSection, openDrawer],
  );

  const activeMenuItem = requireIncidentControlsMenuItem(
    drawerSection ?? lastSection,
    incidentControlsMenuItems,
  );

  return {
    accountIncidentControls,
    activeMenuItem,
    closeButtonRef,
    closeDrawer,
    drawerSection,
  };
}
