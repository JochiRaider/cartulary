import {
  isStandardizedWorkbookViewSchemaId,
  timelineViewSchemaId,
} from "./workbookSurfaceRegistry";

export type WorkbookSheetRef = {
  readonly id: string;
  readonly kind: "saved_view" | "view_schema";
};

export type WorkbookStartupClearedSheetRef = {
  readonly id: string;
  readonly kind: string;
};

export type WorkbookStartupSource =
  | "default"
  | "explicit"
  | "home"
  | "timeline";
export type WorkbookStartupPointerSource = "default" | "home";

export type WorkbookStartupClearedPointer = {
  readonly reasonCode: string;
  readonly sheetRef: WorkbookStartupClearedSheetRef;
  readonly source: WorkbookStartupPointerSource;
};

export type WorkbookStartupSelection = {
  readonly clearedPointers: readonly WorkbookStartupClearedPointer[];
  readonly defaultSheetRef: WorkbookSheetRef | null;
  readonly homeSheetRef: WorkbookSheetRef | null;
  readonly selectedSavedView: unknown | null;
  readonly selectedSheetRef: WorkbookSheetRef;
  readonly selectedViewSchemaId: string;
  readonly source: WorkbookStartupSource;
};

export type WorkbookStartupCandidate = {
  readonly invalidReasonCode?: string;
  readonly selectedSavedView?: unknown | null;
  readonly selectedViewSchemaId?: string;
  readonly sheetRef?: WorkbookSheetRef | null;
  readonly valid: boolean;
};

const startupSources = new Set<string>([
  "default",
  "explicit",
  "home",
  "timeline",
]);
const pointerSources = new Set<string>(["default", "home"]);

export function isWorkbookSheetRef(value: unknown): value is WorkbookSheetRef {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return false;
  }
  const record = value as Record<string, unknown>;
  return (
    (record.kind === "view_schema" || record.kind === "saved_view") &&
    typeof record.id === "string" &&
    record.id.trim() !== ""
  );
}

function isClearedSheetRef(
  value: unknown,
): value is WorkbookStartupClearedSheetRef {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return false;
  }
  const record = value as Record<string, unknown>;
  return (
    typeof record.kind === "string" &&
    record.kind.trim() !== "" &&
    typeof record.id === "string" &&
    record.id.trim() !== ""
  );
}

export function workbookStartupQueryFromURLParams(
  params: URLSearchParams,
): string {
  const startupParams = new URLSearchParams();
  for (const key of ["view_schema_id", "sheet_ref_kind", "sheet_ref_id"]) {
    const value = params.get(key);
    if (value !== null) {
      startupParams.set(key, value);
    }
  }
  const query = startupParams.toString();
  return query ? `?${query}` : "";
}

function isStartupSource(value: unknown): value is WorkbookStartupSource {
  return typeof value === "string" && startupSources.has(value);
}

function isPointerSource(
  value: unknown,
): value is WorkbookStartupPointerSource {
  return typeof value === "string" && pointerSources.has(value);
}

function nullableSheetRef(value: unknown): WorkbookSheetRef | null | undefined {
  if (value === null || value === undefined) {
    return null;
  }
  return isWorkbookSheetRef(value) ? { ...value } : undefined;
}

function selectedViewSchemaIdFor(
  source: WorkbookStartupSource,
  candidate: WorkbookStartupCandidate,
): string {
  const sheetRef = candidate.sheetRef ?? null;
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
  homeSheetRef: WorkbookSheetRef | null,
  defaultSheetRef: WorkbookSheetRef | null,
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
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return null;
  }
  const record = value as Record<string, unknown>;
  if (
    !isWorkbookSheetRef(record.selected_sheet_ref) ||
    typeof record.selected_view_schema_id !== "string" ||
    !isStandardizedWorkbookViewSchemaId(record.selected_view_schema_id) ||
    !isStartupSource(record.source)
  ) {
    return null;
  }
  const homeSheetRef = nullableSheetRef(record.home_sheet_ref);
  const defaultSheetRef = nullableSheetRef(record.default_sheet_ref);
  if (homeSheetRef === undefined || defaultSheetRef === undefined) {
    return null;
  }
  if (!Array.isArray(record.cleared_pointers)) {
    return null;
  }
  const clearedPointers: WorkbookStartupClearedPointer[] = [];
  for (const pointer of record.cleared_pointers) {
    if (!pointer || typeof pointer !== "object" || Array.isArray(pointer)) {
      return null;
    }
    const raw = pointer as Record<string, unknown>;
    if (
      !isPointerSource(raw.source) ||
      !isClearedSheetRef(raw.sheet_ref) ||
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
    selectedSavedView: record.selected_saved_view ?? null,
    selectedSheetRef: { ...record.selected_sheet_ref },
    selectedViewSchemaId: record.selected_view_schema_id,
    source: record.source,
  };
}
