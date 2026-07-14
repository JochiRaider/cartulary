import { useCallback, useEffect, useMemo, useState } from "react";
import { apiPath } from "../../services/browserApi";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
} from "../../services/workbookApi";
import { baseSurfaceIdentityForViewSchemaId } from "../models/workbookSavedViewRuntime";
import {
  isWorkbookSheetRef,
  type WorkbookSheetRef,
} from "../models/workbookStartup";
import {
  knownWorkbookViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";

type WorkbookStartupMutableRef<T> = { current: T };

export type WorkbookIdentity = {
  readonly sheetRef: WorkbookSheetRef;
  readonly viewSchemaId: string | null;
};

export type ApplyWorkbookIdentityOptions = {
  readonly bumpSelectionVersion?: boolean;
  readonly focusFirstGridTarget?: boolean;
  readonly reloadSheet?: boolean;
};

export function useWorkbookStartupController({
  apiBase,
  incidentId,
  surfaceSelectionVersionRef,
}: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
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
  const [startupSheetRef, setStartupSheetRef] = useState<WorkbookSheetRef>(
    () => {
      const explicitExtensionRef = {
        kind: params.get("sheet_ref_kind"),
        extension_profile_id: params.get("extension_profile_id"),
        workspace_key: params.get("sheet_ref_id"),
      };
      return isWorkbookSheetRef(explicitExtensionRef)
        ? explicitExtensionRef
        : { kind: "view_schema", id: initialViewSchemaId };
    },
  );
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
    (sheetRef: Extract<WorkbookSheetRef, { kind: "extension_workspace" }>) => {
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
    const result = await fetchWorkbookJSON<Record<string, unknown>>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/workbook-preferences/me`,
      ),
      {
        method: "PUT",
        body: JSON.stringify({ home_sheet_ref: startupSheetRef }),
      },
    );
    if (!result.ok) {
      throw new Error(parseErrorMessage(result.payload));
    }
  }, [apiBase, incidentId, startupSheetRef]);

  const setWorkbookDefaultSheetRef = useCallback(async () => {
    const result = await fetchWorkbookJSON<Record<string, unknown>>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/workbook-preferences/default`,
      ),
      {
        method: "PUT",
        body: JSON.stringify({ default_sheet_ref: startupSheetRef }),
      },
    );
    if (!result.ok) {
      throw new Error(parseErrorMessage(result.payload));
    }
  }, [apiBase, incidentId, startupSheetRef]);

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
