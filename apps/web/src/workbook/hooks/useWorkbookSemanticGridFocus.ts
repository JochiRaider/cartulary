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
  const { acknowledge, cancel: cancelRequest, request } = focusOwner;
  const latestRequest = useRef(request);
  latestRequest.current = request;
  useEffect(
    () => () => {
      if (latestRequest.current.kind === "pending")
        cancelRequest(latestRequest.current);
    },
    [cancelRequest],
  );
  const [mountedRoot, setMountedRoot] = useState<HTMLElement | null>(null);
  const [registeredFocus, setRegisteredFocus] = useState<
    GridHandle["requestFocus"] | null
  >(null);
  const registerGridHandle = useCallback(
    (handle: GridHandle | null) => {
      gridHandleRef.current = handle;
      setMountedRoot(handle?.getScrollElement() ?? null);
      setRegisteredFocus(() => handle?.requestFocus ?? null);
    },
    [gridHandleRef],
  );
  const stableDraftFieldKeys = useStableFieldKeys(draftFieldKeys);
  const visibleFieldKeys = useStableFieldKeys(
    visibleColumns.map((column) => column.fieldKey),
  );

  useEffect(() => {
    if (request.kind !== "pending" || request.viewSchemaId !== viewSchemaId)
      return;
    const abort = new AbortController();
    const cancel = () => {
      abort.abort();
      cancelRequest(request);
    };
    if (dataState.kind === "permission_denied") {
      cancel();
      return;
    }
    document.addEventListener("pointerdown", cancel, true);
    document.addEventListener("keydown", cancel, true);
    const acknowledgement: WorkbookGridEntryFocusAcknowledgement = {
      generation: request.generation,
      viewSchemaId: request.viewSchemaId,
    };
    const targets = [
      ...stableDraftFieldKeys.map((fieldKey) => ({
        kind: "draft" as const,
        fieldKey,
      })),
      ...dataRows.flatMap((row) =>
        visibleFieldKeys.map((fieldKey) => ({
          kind: "cell" as const,
          anchor: {
            fieldKey,
            rowIdentity: row.rowIdentity,
            surface: { kind: "view_schema" as const, viewSchemaId },
          },
        })),
      ),
      { kind: "root" as const },
    ];
    void (async () => {
      if (
        registeredFocus === null ||
        mountedRoot === null ||
        gridDataStateIsBusy(dataState.kind)
      )
        return;
      for (const target of targets) {
        if (abort.signal.aborted) return;
        const result = await registeredFocus(target, {
          signal: abort.signal,
        });
        if (abort.signal.aborted) return;
        if (result === "cancelled") {
          cancelRequest(request);
          return;
        }
        if (result === "focused") {
          acknowledge(acknowledgement);
          return;
        }
      }
    })();
    return () => {
      abort.abort();
      document.removeEventListener("pointerdown", cancel, true);
      document.removeEventListener("keydown", cancel, true);
    };
  }, [
    dataRows,
    dataState.kind,
    acknowledge,
    cancelRequest,
    mountedRoot,
    registeredFocus,
    request,
    stableDraftFieldKeys,
    viewSchemaId,
    visibleFieldKeys,
  ]);

  return registerGridHandle;
}
