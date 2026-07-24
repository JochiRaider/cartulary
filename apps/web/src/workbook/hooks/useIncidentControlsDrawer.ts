import type { IncidentControlsSection } from "@cartulary/ui-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
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

export function useIncidentControlsDrawer(importAssistantAvailable = false) {
  const [drawerSection, setDrawerSection] =
    useState<IncidentControlsSection | null>(null);
  const [lastSection, setLastSection] =
    useState<IncidentControlsSection>("summary");
  const returnFocusTargetRef = useRef<HTMLElement | null>(null);
  const closeButtonRef = useRef<HTMLButtonElement | null>(null);

  const deferFocus = useCallback((resolveTarget: () => HTMLElement | null) => {
    window.setTimeout(() => {
      resolveTarget()?.focus({ preventScroll: true });
    }, 0);
  }, []);

  const closeDrawer = useCallback(
    (options: { readonly restoreTriggerFocus: boolean }) => {
      setDrawerSection(null);
      if (options.restoreTriggerFocus) {
        deferFocus(() => returnFocusTargetRef.current);
      }
    },
    [deferFocus],
  );

  const openDrawer = useCallback(
    (
      section: IncidentControlsSection,
      returnFocusTarget?: HTMLElement | null,
    ) => {
      returnFocusTargetRef.current = returnFocusTarget ?? null;
      setLastSection(section);
      setDrawerSection(section);
    },
    [],
  );

  useEffect(() => {
    if (drawerSection === null) {
      return;
    }
    deferFocus(() => closeButtonRef.current);
  }, [drawerSection, deferFocus]);

  useEffect(() => {
    if (
      !importAssistantAvailable &&
      drawerSection === importAssistantMenuItem.section
    ) {
      setDrawerSection(null);
      setLastSection("summary");
    }
  }, [drawerSection, importAssistantAvailable]);

  const incidentControlsMenuItems = useMemo(
    () =>
      importAssistantAvailable
        ? [
            baseIncidentControlsMenuItems[0],
            importAssistantMenuItem,
            ...baseIncidentControlsMenuItems.slice(1),
          ]
        : baseIncidentControlsMenuItems,
    [importAssistantAvailable],
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
