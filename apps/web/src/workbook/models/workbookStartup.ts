import type { SheetRef } from "../../shared/sheetRef";
import { isSheetRef } from "../../shared/sheetRef";
import {
  isStandardizedWorkbookViewSchemaId,
  timelineViewSchemaId,
} from "./workbookSurfaceRegistry";

export type WorkbookStartupQuery = {
  readonly extensionProfileId?: string;
  readonly sheetRefId?: string;
  readonly sheetRefKind?: string;
};

export type WorkbookStartupRejectedReference = {
  readonly kind: string;
  readonly id?: string;
  readonly extension_profile_id?: string;
  readonly workspace_key?: string;
};

export type WorkbookStartupSource =
  | "default"
  | "explicit"
  | "home"
  | "timeline";
export type WorkbookStartupPointerSource = "default" | "home";
export type WorkbookStartupSavedViewResource = Readonly<
  Record<string, unknown>
>;

export type WorkbookStartupClearedPointer = {
  readonly reasonCode: string;
  readonly sheetRef: WorkbookStartupRejectedReference;
  readonly source: WorkbookStartupPointerSource;
};

export type WorkbookStartupSelection = {
  readonly clearedPointers: readonly WorkbookStartupClearedPointer[];
  readonly defaultSheetRef: SheetRef | null;
  readonly homeSheetRef: SheetRef | null;
  readonly selectedSavedView: WorkbookStartupSavedViewResource | null;
  readonly selectedSheetRef: SheetRef;
  readonly selectedViewSchemaId: string | null;
  readonly source: WorkbookStartupSource;
};

export type WorkbookStartupCandidate = {
  readonly invalidReasonCode?: string;
  readonly selectedSavedView?: WorkbookStartupSavedViewResource | null;
  readonly selectedViewSchemaId?: string | null;
  readonly sheetRef?: SheetRef | null;
  readonly valid: boolean;
};

const startupSources = new Set<string>([
  "default",
  "explicit",
  "home",
  "timeline",
]);
const pointerSources = new Set<string>(["default", "home"]);

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function isRejectedReference(
  value: unknown,
): value is WorkbookStartupRejectedReference {
  if (!isRecord(value)) {
    return false;
  }
  const record = value;
  if (isSheetRef(record)) {
    return true;
  }
  return (
    typeof record.kind === "string" &&
    record.kind.trim() !== "" &&
    typeof record.id === "string" &&
    record.id.trim() !== ""
  );
}

export function workbookStartupQueryFromURLParams(
  params: URLSearchParams,
): WorkbookStartupQuery {
  const viewSchemaId = params.get("view_schema_id");
  const sheetRefKind = params.get("sheet_ref_kind");
  const sheetRefId = params.get("sheet_ref_id");
  const extensionProfileId = params.get("extension_profile_id");
  if (
    sheetRefKind === null &&
    sheetRefId === null &&
    extensionProfileId === null
  ) {
    return viewSchemaId === null
      ? {}
      : { sheetRefId: viewSchemaId, sheetRefKind: "view_schema" };
  }
  return {
    ...(sheetRefKind === null ? {} : { sheetRefKind }),
    ...(sheetRefId === null ? {} : { sheetRefId }),
    ...(extensionProfileId === null ? {} : { extensionProfileId }),
  };
}

function isStartupSource(value: unknown): value is WorkbookStartupSource {
  return typeof value === "string" && startupSources.has(value);
}

function isPointerSource(
  value: unknown,
): value is WorkbookStartupPointerSource {
  return typeof value === "string" && pointerSources.has(value);
}

function nullableSheetRef(value: unknown): SheetRef | null | undefined {
  if (value === null || value === undefined) {
    return null;
  }
  return isSheetRef(value) ? { ...value } : undefined;
}

function nullableStartupSavedView(
  value: unknown,
): WorkbookStartupSavedViewResource | null | undefined {
  if (value === null || value === undefined) {
    return null;
  }
  if (!isRecord(value)) {
    return undefined;
  }
  return { ...value };
}

function selectedViewSchemaIdFor(
  source: WorkbookStartupSource,
  candidate: WorkbookStartupCandidate,
): string | null {
  const sheetRef = candidate.sheetRef ?? null;
  if (sheetRef?.kind === "extension_workspace") {
    if (
      candidate.selectedViewSchemaId !== null &&
      candidate.selectedViewSchemaId !== undefined
    ) {
      throw new Error(
        `Startup candidate ${source} assigned a view schema to an extension workspace`,
      );
    }
    return null;
  }
  const viewSchemaId =
    candidate.selectedViewSchemaId ??
    (sheetRef?.kind === "view_schema" ? sheetRef.id : "");
  if (!isStandardizedWorkbookViewSchemaId(viewSchemaId)) {
    throw new Error(
      `Startup candidate ${source} selected unsupported view_schema_id ${viewSchemaId}`,
    );
  }
  return viewSchemaId;
}

function fallbackSelection(
  clearedPointers: readonly WorkbookStartupClearedPointer[],
  homeSheetRef: SheetRef | null,
  defaultSheetRef: SheetRef | null,
): WorkbookStartupSelection {
  return {
    clearedPointers,
    defaultSheetRef,
    homeSheetRef,
    selectedSavedView: null,
    selectedSheetRef: { kind: "view_schema", id: timelineViewSchemaId },
    selectedViewSchemaId: timelineViewSchemaId,
    source: "timeline",
  };
}

export function resolveWorkbookStartupFallback(input: {
  readonly default?: WorkbookStartupCandidate | null;
  readonly explicit?: WorkbookStartupCandidate | null;
  readonly home?: WorkbookStartupCandidate | null;
}): WorkbookStartupSelection {
  let homeSheetRef = input.home?.sheetRef ?? null;
  let defaultSheetRef = input.default?.sheetRef ?? null;
  const clearedPointers: WorkbookStartupClearedPointer[] = [];

  for (const [source, candidate] of [
    ["explicit", input.explicit],
    ["home", input.home],
    ["default", input.default],
  ] as const) {
    if (!candidate?.sheetRef) {
      continue;
    }
    const sheetRef = candidate.sheetRef;
    if (candidate.valid) {
      return {
        clearedPointers,
        defaultSheetRef,
        homeSheetRef,
        selectedSavedView: candidate.selectedSavedView ?? null,
        selectedSheetRef: { ...sheetRef },
        selectedViewSchemaId: selectedViewSchemaIdFor(source, candidate),
        source,
      };
    }
    if (source === "home" || source === "default") {
      clearedPointers.push({
        reasonCode: candidate.invalidReasonCode ?? "invalid_sheet_ref",
        sheetRef: { ...sheetRef },
        source,
      });
      if (source === "home") {
        homeSheetRef = null;
      } else {
        defaultSheetRef = null;
      }
    }
  }

  return fallbackSelection(clearedPointers, homeSheetRef, defaultSheetRef);
}

export function normalizeWorkbookStartupSelection(
  value: unknown,
): WorkbookStartupSelection | null {
  if (!isRecord(value)) {
    return null;
  }
  const record = value;
  const selectedSheetRef = record.selected_sheet_ref;
  if (!isSheetRef(selectedSheetRef) || !isStartupSource(record.source)) {
    return null;
  }
  const selectedViewSchemaId = record.selected_view_schema_id;
  if (selectedSheetRef.kind === "extension_workspace") {
    if (selectedViewSchemaId !== null || record.selected_saved_view !== null) {
      return null;
    }
  } else if (
    typeof selectedViewSchemaId !== "string" ||
    !isStandardizedWorkbookViewSchemaId(selectedViewSchemaId)
  ) {
    return null;
  }
  if (
    selectedSheetRef.kind === "saved_view" &&
    record.selected_saved_view === null
  ) {
    return null;
  }
  const homeSheetRef = nullableSheetRef(record.home_sheet_ref);
  const defaultSheetRef = nullableSheetRef(record.default_sheet_ref);
  const selectedSavedView = nullableStartupSavedView(
    record.selected_saved_view,
  );
  if (
    homeSheetRef === undefined ||
    defaultSheetRef === undefined ||
    selectedSavedView === undefined
  ) {
    return null;
  }
  if (!Array.isArray(record.cleared_pointers)) {
    return null;
  }
  const clearedPointers: WorkbookStartupClearedPointer[] = [];
  for (const pointer of record.cleared_pointers) {
    if (!isRecord(pointer)) {
      return null;
    }
    const raw = pointer;
    if (
      !isPointerSource(raw.source) ||
      !isRejectedReference(raw.sheet_ref) ||
      typeof raw.reason_code !== "string" ||
      raw.reason_code.trim() === ""
    ) {
      return null;
    }
    clearedPointers.push({
      reasonCode: raw.reason_code,
      sheetRef: { ...raw.sheet_ref },
      source: raw.source,
    });
  }

  return {
    clearedPointers,
    defaultSheetRef,
    homeSheetRef,
    selectedSavedView,
    selectedSheetRef: { ...selectedSheetRef },
    selectedViewSchemaId,
    source: record.source,
  };
}
