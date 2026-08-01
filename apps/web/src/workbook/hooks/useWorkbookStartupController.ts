import { useCallback, useEffect, useMemo, useState } from "react";
import type { SheetRef } from "../../shared/sheetRef";
import { isSheetRef } from "../../shared/sheetRef";
import { baseSurfaceIdentityForViewSchemaId } from "../models/workbookSavedViewRuntime";
import {
  knownWorkbookViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import type { WorkbookPreferencePort } from "../ports/WorkbookPreferencePort";

type WorkbookStartupMutableRef<T> = { current: T };

export type WorkbookIdentity = {
  readonly sheetRef: SheetRef;
  readonly viewSchemaId: string | null;
};

type ApplyWorkbookIdentityOptions = {
  readonly bumpSelectionVersion?: boolean;
  readonly focusFirstGridTarget?: boolean;
  readonly reloadSheet?: boolean;
};

export function useWorkbookStartupController({
  incidentId,
  preferencePort,
  surfaceSelectionVersionRef,
}: {
  readonly incidentId: string;
  readonly preferencePort: WorkbookPreferencePort;
  readonly surfaceSelectionVersionRef: WorkbookStartupMutableRef<number>;
}) {
  const params = useMemo(() => new URLSearchParams(window.location.search), []);
  const initialViewSchemaId = useMemo(() => {
    const explicit = params.get("view_schema_id");
    return explicit
      ? knownWorkbookViewSchemaId(explicit)
      : timelineViewSchemaId;
  }, [params]);
  const [surface, setSurface] = useState(initialViewSchemaId);
  const [startupSheetRef, setStartupSheetRef] = useState<SheetRef>(() => {
    const explicitExtensionRef = {
      kind: params.get("sheet_ref_kind"),
      extension_profile_id: params.get("extension_profile_id"),
      workspace_key: params.get("sheet_ref_id"),
    };
    return isSheetRef(explicitExtensionRef)
      ? explicitExtensionRef
      : { kind: "view_schema", id: initialViewSchemaId };
  });
  const [sheetReloadToken, setSheetReloadToken] = useState(0);
  const [pendingGridFocusSurface, setPendingGridFocusSurface] = useState<
    string | null
  >(null);

  const applyWorkbookIdentity = useCallback(
    (
      identity: WorkbookIdentity,
      options: ApplyWorkbookIdentityOptions = {},
    ) => {
      if (options.bumpSelectionVersion !== false) {
        surfaceSelectionVersionRef.current += 1;
      }
      if (identity.viewSchemaId !== null) {
        setSurface(identity.viewSchemaId);
      }
      setStartupSheetRef({ ...identity.sheetRef });
      if (options.focusFirstGridTarget && identity.viewSchemaId !== null) {
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

  const selectExtensionWorkspace = useCallback(
    (sheetRef: Extract<SheetRef, { kind: "extension_workspace" }>) => {
      applyWorkbookIdentity({ sheetRef, viewSchemaId: null });
    },
    [applyWorkbookIdentity],
  );

  const applyStartupIdentity = useCallback(
    (identity: WorkbookIdentity) => {
      applyWorkbookIdentity(identity, { bumpSelectionVersion: false });
    },
    [applyWorkbookIdentity],
  );

  const setWorkbookHomeSheetRef = useCallback(async () => {
    const result = await preferencePort.setHomeSheet({
      sheetRef: startupSheetRef,
      signal: new AbortController().signal,
    });
    if (result.kind !== "accepted") {
      throw new Error(
        result.kind === "aborted"
          ? "Workbook home preference update was aborted."
          : result.failure.message,
      );
    }
  }, [preferencePort, startupSheetRef]);

  const setWorkbookDefaultSheetRef = useCallback(async () => {
    const result = await preferencePort.setDefaultSheet({
      sheetRef: startupSheetRef,
      signal: new AbortController().signal,
    });
    if (result.kind !== "accepted") {
      throw new Error(
        result.kind === "aborted"
          ? "Workbook default preference update was aborted."
          : result.failure.message,
      );
    }
  }, [preferencePort, startupSheetRef]);

  useEffect(() => {
    const next = new URLSearchParams(window.location.search);
    next.set("incident_id", incidentId);
    if (startupSheetRef.kind === "saved_view") {
      next.delete("view_schema_id");
      next.set("sheet_ref_kind", startupSheetRef.kind);
      next.set("sheet_ref_id", startupSheetRef.id);
      next.delete("extension_profile_id");
    } else if (startupSheetRef.kind === "extension_workspace") {
      next.delete("view_schema_id");
      next.set("sheet_ref_kind", startupSheetRef.kind);
      next.set("sheet_ref_id", startupSheetRef.workspace_key);
      next.set("extension_profile_id", startupSheetRef.extension_profile_id);
    } else {
      next.set("view_schema_id", surface);
      next.delete("sheet_ref_kind");
      next.delete("sheet_ref_id");
      next.delete("extension_profile_id");
    }
    next.delete("workspace_key");
    next.delete("surface");
    window.history.replaceState({}, "", `/?${next.toString()}`);
  }, [incidentId, startupSheetRef, surface]);

  return {
    commands: {
      applyStartupIdentity,
      applyWorkbookIdentity,
      selectExtensionWorkspace,
      selectWorkbookSurface,
      setPendingGridFocusSurface,
      setWorkbookDefaultSheetRef,
      setWorkbookHomeSheetRef,
    },
    refs: { params },
    snapshot: {
      pendingGridFocusSurface,
      sheetReloadToken,
      startupSheetRef,
      surface,
    },
  };
}
