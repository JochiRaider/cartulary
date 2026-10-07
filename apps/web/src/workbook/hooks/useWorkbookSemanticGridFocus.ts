import type {
  GridColumn,
  GridDataRow,
  GridDataState,
  GridHandle,
} from "@cartulary/grid-adapter";
import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type {
  WorkbookGridEntryFocusAcknowledgement,
  WorkbookGridEntryFocusOwner,
} from "../models/workbookGridEntryFocus";
import { useWorkbookBrowsingRegistry } from "../query/WorkbookQueryBrowsingContext";

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
  onNavigateRecord,
  visibleColumns,
  viewSchemaId,
}: {
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly dataState: GridDataState;
  readonly draftFieldKeys?: readonly string[] | undefined;
  readonly focusOwner: WorkbookGridEntryFocusOwner;
  readonly gridHandleRef: WorkbookGridHandleRef;
  readonly onNavigateRecord?: ((recordId: string) => void) | undefined;
  readonly visibleColumns: readonly GridColumn<Row>[];
  readonly viewSchemaId: string;
}) {
  const browsingRegistry = useWorkbookBrowsingRegistry();
  useSyncExternalStore(
    browsingRegistry.subscribe,
    browsingRegistry.getSnapshot,
  );
  const presentationReady = browsingRegistry.presentationReady(viewSchemaId);
  const selectionCommit = useRef<(() => void) | null>(null);
  const [, selectionChanged] = useState(0);
  useLayoutEffect(() => {
    const committed = selectionCommit.current;
    selectionCommit.current = null;
    committed?.();
  });
  const navigateRecord = useRef(onNavigateRecord);
  navigateRecord.current = onNavigateRecord;
  useLayoutEffect(() => {
    const unbind = browsingRegistry.bindGrid(
      viewSchemaId,
      gridHandleRef,
      (recordId, committed) => {
        navigateRecord.current?.(recordId);
        selectionCommit.current = committed;
        selectionChanged((revision) => revision + 1);
      },
    );
    return () => {
      const anchor = gridHandleRef.current?.getActiveCell?.();
      if (anchor?.rowIdentity.kind === "core_record")
        browsingRegistry
          ?.find(viewSchemaId)
          ?.rememberAnchor(anchor.rowIdentity.recordId);
      unbind?.();
    };
  }, [browsingRegistry, viewSchemaId, gridHandleRef]);
  useEffect(() => {
    const anchor = gridHandleRef.current?.getActiveCell?.();
    if (anchor?.rowIdentity.kind === "core_record")
      browsingRegistry
        ?.find(viewSchemaId)
        ?.rememberAnchor(anchor.rowIdentity.recordId);
    gridHandleRef.current?.setAccessibleDescription?.(
      browsingRegistry.find(viewSchemaId)
        ? "Row indices and selection refer to the loaded window. Use the workbook browsing controls to reach additional records."
        : undefined,
    );
  });
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
      browsingRegistry.refreshGridBinding(viewSchemaId);
      setMountedRoot(handle?.getScrollElement() ?? null);
      setRegisteredFocus(() => handle?.requestFocus ?? null);
    },
    [gridHandleRef, browsingRegistry, viewSchemaId],
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
        !presentationReady ||
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
    presentationReady,
    registeredFocus,
    request,
    stableDraftFieldKeys,
    viewSchemaId,
    visibleFieldKeys,
  ]);

  return registerGridHandle;
}
