import { useCallback, useState } from "react";
import {
  baseSurfaceIdentityForViewSchemaId,
  fallbackIdentityAfterSavedViewDelete,
  removeSavedViewList,
  savedViewIdentityForSelection,
  upsertSavedViewList,
} from "./workbookSavedViewRuntime";
import type { SavedViewResource } from "./workbookSavedViews";
import type { WorkbookSheetRef } from "./workbookStartup";

type WorkbookShellMutableRef<T> = {
  current: T;
};

export function useWorkbookShellRuntime({
  initialViewSchemaId,
  surfaceSelectionVersionRef,
}: {
  readonly initialViewSchemaId: string;
  readonly surfaceSelectionVersionRef: WorkbookShellMutableRef<number>;
}) {
  const [surface, setSurface] = useState(initialViewSchemaId);
  const [startupSheetRef, setStartupSheetRef] = useState<WorkbookSheetRef>(
    () => ({ kind: "view_schema", id: initialViewSchemaId }),
  );
  const [sheetReloadToken, setSheetReloadToken] = useState(0);
  const [pendingGridFocusSurface, setPendingGridFocusSurface] = useState<
    string | null
  >(null);
  const [savedViews, setSavedViews] = useState<SavedViewResource[]>([]);

  const applyWorkbookIdentity = useCallback(
    (
      identity: {
        readonly sheetRef: WorkbookSheetRef;
        readonly viewSchemaId: string;
      },
      options: {
        readonly bumpSelectionVersion?: boolean;
        readonly focusFirstGridTarget?: boolean;
        readonly reloadSheet?: boolean;
      } = {},
    ) => {
      if (options.bumpSelectionVersion !== false) {
        surfaceSelectionVersionRef.current += 1;
      }
      setSurface(identity.viewSchemaId);
      setStartupSheetRef({ ...identity.sheetRef });
      if (options.focusFirstGridTarget) {
        setPendingGridFocusSurface(identity.viewSchemaId);
      }
      if (options.reloadSheet) {
        setSheetReloadToken((current) => current + 1);
      }
    },
    [surfaceSelectionVersionRef],
  );

  const selectWorkbookSurface = useCallback(
    (
      viewSchemaId: string,
      options: { readonly focusFirstGridTarget?: boolean } = {},
    ) => {
      applyWorkbookIdentity(baseSurfaceIdentityForViewSchemaId(viewSchemaId), {
        focusFirstGridTarget: options.focusFirstGridTarget === true,
      });
    },
    [applyWorkbookIdentity],
  );

  const selectSavedViewIdentity = useCallback(
    (
      savedView: Pick<SavedViewResource, "saved_view_id" | "view_schema_id">,
    ) => {
      applyWorkbookIdentity(savedViewIdentityForSelection(savedView), {
        reloadSheet: true,
      });
    },
    [applyWorkbookIdentity],
  );

  const applyStartupIdentity = useCallback(
    (identity: {
      readonly sheetRef: WorkbookSheetRef;
      readonly viewSchemaId: string;
    }) => {
      applyWorkbookIdentity(identity, { bumpSelectionVersion: false });
    },
    [applyWorkbookIdentity],
  );

  const upsertSavedView = useCallback((savedView: SavedViewResource) => {
    setSavedViews((current) => upsertSavedViewList(current, savedView));
  }, []);

  const replaceSavedViews = useCallback(
    (nextSavedViews: SavedViewResource[]) => {
      setSavedViews(nextSavedViews);
    },
    [],
  );

  const deleteSavedViewIdentity = useCallback(
    (savedView: SavedViewResource, activeSheetRef: WorkbookSheetRef) => {
      setSavedViews((current) =>
        removeSavedViewList(current, savedView.saved_view_id),
      );
      const fallback = fallbackIdentityAfterSavedViewDelete(
        activeSheetRef,
        savedView,
      );
      if (fallback !== null) {
        applyWorkbookIdentity(fallback, { reloadSheet: true });
      }
    },
    [applyWorkbookIdentity],
  );

  return {
    commands: {
      applyStartupIdentity,
      deleteSavedViewIdentity,
      replaceSavedViews,
      selectSavedViewIdentity,
      selectWorkbookSurface,
      setPendingGridFocusSurface,
      upsertSavedView,
    },
    snapshot: {
      pendingGridFocusSurface,
      savedViews,
      sheetReloadToken,
      startupSheetRef,
      surface,
    },
  };
}
