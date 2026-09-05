import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { SheetRef } from "../../shared/sheetRef";
import { isSheetRef } from "../../shared/sheetRef";
import type {
  WorkbookGridEntryFocusAcknowledgement,
  WorkbookGridEntryFocusRequest,
} from "../models/workbookGridEntryFocus";
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
  const gridEntryFocusGenerationRef = useRef(0);
  const [gridEntryFocusRequest, setGridEntryFocusRequest] =
    useState<WorkbookGridEntryFocusRequest>({ kind: "idle" });

  const cancelGridEntryFocus = useCallback(() => {
    setGridEntryFocusRequest((current) =>
      current.kind === "idle" ? current : { kind: "idle" },
    );
  }, []);

  const acknowledgeGridEntryFocus = useCallback(
    (acknowledgement: WorkbookGridEntryFocusAcknowledgement) => {
      setGridEntryFocusRequest((current) =>
        current.kind === "pending" &&
        current.viewSchemaId === acknowledgement.viewSchemaId &&
        current.generation === acknowledgement.generation
          ? { kind: "idle" }
          : current,
      );
    },
    [],
  );

  const applyWorkbookIdentity = useCallback(
    (
      identity: WorkbookIdentity,
      options: ApplyWorkbookIdentityOptions = {},
    ) => {
      cancelGridEntryFocus();
      if (options.bumpSelectionVersion !== false) {
        surfaceSelectionVersionRef.current += 1;
      }
      if (identity.viewSchemaId !== null) {
        setSurface(identity.viewSchemaId);
      }
      setStartupSheetRef({ ...identity.sheetRef });
      if (options.focusFirstGridTarget && identity.viewSchemaId !== null) {
        gridEntryFocusGenerationRef.current += 1;
        setGridEntryFocusRequest({
          generation: gridEntryFocusGenerationRef.current,
          kind: "pending",
          viewSchemaId: identity.viewSchemaId,
        });
      }
      if (options.reloadSheet) {
        setSheetReloadToken((current) => current + 1);
      }
    },
    [cancelGridEntryFocus, surfaceSelectionVersionRef],
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
    // App owns incident navigation. A retiring workbook may still run an effect
    // before its transition unmounts it; it owns sheet selection only in its route.
    if (
      window.location.pathname !== "/" ||
      next.get("incident_id") !== incidentId
    )
      return;
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
      acknowledgeGridEntryFocus,
      applyStartupIdentity,
      applyWorkbookIdentity,
      cancelGridEntryFocus,
      selectExtensionWorkspace,
      selectWorkbookSurface,
      setWorkbookDefaultSheetRef,
      setWorkbookHomeSheetRef,
    },
    refs: { params },
    snapshot: {
      gridEntryFocusRequest,
      sheetReloadToken,
      startupSheetRef,
      surface,
    },
  };
}
