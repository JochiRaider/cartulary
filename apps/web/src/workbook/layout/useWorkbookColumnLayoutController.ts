import type { ViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useState } from "react";
import type { WorkbookLayoutState } from "../models/workbookQuery";
import type { WorkbookSheetRef } from "../models/workbookStartup";
import { workbookContractForViewSchemaId } from "../models/workbookSurfaceQueryRuntime";
import {
  defaultWorkbookLayoutState,
  moveWorkbookColumn,
  reorderWorkbookColumns,
  resolveWorkbookLayoutState,
  setWorkbookColumnHidden,
  setWorkbookColumnWidth,
  type WorkbookResolvedLayoutState,
} from "./workbookColumnLayout";

export function useWorkbookColumnLayoutController({
  activeContract,
  startupSheetRef,
}: {
  readonly activeContract: ViewContract;
  readonly startupSheetRef: WorkbookSheetRef;
}) {
  const [entries, setEntries] = useState<
    Readonly<Record<string, WorkbookResolvedLayoutState>>
  >({});

  const currentLayoutStateForSurface = useCallback(
    (viewSchemaId: string): WorkbookResolvedLayoutState =>
      entries[viewSchemaId] ??
      defaultWorkbookLayoutState(workbookContractForViewSchemaId(viewSchemaId)),
    [entries],
  );

  const applyLayoutStateForSurface = useCallback(
    (viewSchemaId: string, state: WorkbookLayoutState) => {
      const contract = workbookContractForViewSchemaId(viewSchemaId);
      setEntries((current) => ({
        ...current,
        [viewSchemaId]: resolveWorkbookLayoutState(contract, state),
      }));
    },
    [],
  );

  const updateActive = useCallback(
    (
      update: (
        current: WorkbookResolvedLayoutState,
      ) => WorkbookResolvedLayoutState,
    ) => {
      setEntries((current) => {
        const previous =
          current[activeContract.viewSchemaId] ??
          defaultWorkbookLayoutState(activeContract);
        return {
          ...current,
          [activeContract.viewSchemaId]: update(previous),
        };
      });
    },
    [activeContract],
  );

  useEffect(() => {
    if (
      startupSheetRef.kind !== "view_schema" ||
      startupSheetRef.id !== activeContract.viewSchemaId
    ) {
      return;
    }
    applyLayoutStateForSurface(
      activeContract.viewSchemaId,
      defaultWorkbookLayoutState(activeContract),
    );
  }, [activeContract, applyLayoutStateForSurface, startupSheetRef]);

  const activeLayoutState = currentLayoutStateForSurface(
    activeContract.viewSchemaId,
  );
  const activeLayoutControls = useMemo(
    () => ({
      layoutState: activeLayoutState,
      onColumnHiddenChange: (fieldKey: string, hidden: boolean) => {
        updateActive((current) =>
          setWorkbookColumnHidden(activeContract, current, fieldKey, hidden),
        );
      },
      onColumnMove: (fieldKey: string, direction: "earlier" | "later") => {
        updateActive((current) =>
          moveWorkbookColumn(activeContract, current, fieldKey, direction),
        );
      },
      onColumnReorder: (sourceFieldKey: string, targetFieldKey: string) => {
        updateActive((current) =>
          reorderWorkbookColumns(
            activeContract,
            current,
            sourceFieldKey,
            targetFieldKey,
          ),
        );
      },
      onResetColumns: () => {
        updateActive(() => defaultWorkbookLayoutState(activeContract));
      },
      onColumnWidthChange: (fieldKey: string, width: number) => {
        updateActive((current) =>
          setWorkbookColumnWidth(activeContract, current, fieldKey, width),
        );
      },
    }),
    [activeContract, activeLayoutState, updateActive],
  );

  return {
    commands: {
      applyLayoutStateForSurface,
      currentLayoutStateForSurface,
    },
    snapshot: { activeLayoutControls, activeLayoutState },
  };
}
