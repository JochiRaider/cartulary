import type {
  GridColumn,
  GridDataRow,
  GridDataState,
  GridHandle,
} from "@cartulary/grid-adapter";
import { useCallback, useEffect, useRef, useState } from "react";
import type {
  WorkbookGridEntryFocusAcknowledgement,
  WorkbookGridEntryFocusOwner,
} from "../models/workbookGridEntryFocus";

type WorkbookGridHandleRef = {
  current: GridHandle | null;
};

const noDraftFieldKeys: readonly string[] = [];

function useStableFieldKeys(fieldKeys: readonly string[]): readonly string[] {
  const stableFieldKeysRef = useRef(fieldKeys);
  if (
    stableFieldKeysRef.current.length !== fieldKeys.length ||
    fieldKeys.some(
      (fieldKey, index) => stableFieldKeysRef.current[index] !== fieldKey,
    )
  ) {
    stableFieldKeysRef.current = fieldKeys;
  }
  return stableFieldKeysRef.current;
}

function gridDataStateIsBusy(kind: GridDataState["kind"]): boolean {
  return kind === "initial_loading" || kind === "refreshing";
}

export function useWorkbookSemanticGridFocus<Row>({
  dataRows,
  dataState,
  draftFieldKeys = noDraftFieldKeys,
  focusOwner,
  gridHandleRef,
  visibleColumns,
  viewSchemaId,
}: {
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly dataState: GridDataState;
  readonly draftFieldKeys?: readonly string[] | undefined;
  readonly focusOwner: WorkbookGridEntryFocusOwner;
  readonly gridHandleRef: WorkbookGridHandleRef;
  readonly visibleColumns: readonly GridColumn<Row>[];
  readonly viewSchemaId: string;
}) {
  const { acknowledge, request } = focusOwner;
  const observedHandleRef = useRef(false);
  const [firstMountedHandle, setFirstMountedHandle] =
    useState<GridHandle | null>(null);
  const registerGridHandle = useCallback(
    (handle: GridHandle | null) => {
      gridHandleRef.current = handle;
      if (handle !== null && !observedHandleRef.current) {
        observedHandleRef.current = true;
        setFirstMountedHandle(handle);
      }
    },
    [gridHandleRef],
  );
  const stableDraftFieldKeys = useStableFieldKeys(draftFieldKeys);
  const visibleFieldKeys = useStableFieldKeys(
    visibleColumns.map((column) => column.fieldKey),
  );

  useEffect(() => {
    const mountedHandle =
      firstMountedHandle === null ? null : gridHandleRef.current;
    if (
      request.kind !== "pending" ||
      request.viewSchemaId !== viewSchemaId ||
      mountedHandle === null ||
      gridDataStateIsBusy(dataState.kind)
    ) {
      return;
    }

    const acknowledgement: WorkbookGridEntryFocusAcknowledgement = {
      generation: request.generation,
      viewSchemaId: request.viewSchemaId,
    };
    for (const fieldKey of stableDraftFieldKeys) {
      if (mountedHandle.focusDraftCell(fieldKey)) {
        acknowledge(acknowledgement);
        return;
      }
    }
    for (const row of dataRows) {
      for (const fieldKey of visibleFieldKeys) {
        if (
          mountedHandle.focusAnchor({
            fieldKey,
            rowIdentity: row.rowIdentity,
            surface: { kind: "view_schema", viewSchemaId },
          })
        ) {
          acknowledge(acknowledgement);
          return;
        }
      }
    }
    if (mountedHandle.focusRoot()) {
      acknowledge(acknowledgement);
    }
  }, [
    dataRows,
    dataState.kind,
    acknowledge,
    firstMountedHandle,
    gridHandleRef,
    request,
    stableDraftFieldKeys,
    viewSchemaId,
    visibleFieldKeys,
  ]);

  return registerGridHandle;
}
