import { useMemo } from "react";
import {
  type FocusFieldKey,
  inputFocusKey,
  type RowValues,
  type TimelineScalarEditorSurface,
  timelineScalarBindings,
  timelineScalarEditorSurfaces,
  type WorkbookRow,
} from "../models/workbookTimelineModel";

type TimelineScalarEditorIdentity = {
  readonly field: keyof RowValues;
  readonly rowKey: string;
  readonly surface: TimelineScalarEditorSurface;
};

type TimelineInputIdentity = {
  readonly field: FocusFieldKey;
  readonly rowKey: string;
  readonly surface: TimelineScalarEditorSurface;
};

type TimelineEditorElement = HTMLInputElement | HTMLTextAreaElement;

export type TimelineEditorDraftRegistry = ReturnType<
  typeof createTimelineEditorDraftRegistry
>;

export function createTimelineEditorDraftRegistry() {
  const draftValues = new Map<string, string>();
  const inputElements = new Map<string, TimelineEditorElement>();
  const inputTestIds = new Map<string, string>();
  const focusKeysByRow = new Map<string, Set<string>>();

  const rememberFocusKey = (rowKey: string, focusKey: string) => {
    const rowFocusKeys = focusKeysByRow.get(rowKey) ?? new Set<string>();
    rowFocusKeys.add(focusKey);
    focusKeysByRow.set(rowKey, rowFocusKeys);
  };

  const forgetFocusKeyIfUnused = (rowKey: string, focusKey: string) => {
    if (
      draftValues.has(focusKey) ||
      inputElements.has(focusKey) ||
      inputTestIds.has(focusKey)
    ) {
      return;
    }
    const rowFocusKeys = focusKeysByRow.get(rowKey);
    rowFocusKeys?.delete(focusKey);
    if (rowFocusKeys?.size === 0) focusKeysByRow.delete(rowKey);
  };

  const clearRow = (rowKey: string) => {
    for (const focusKey of focusKeysByRow.get(rowKey) ?? []) {
      draftValues.delete(focusKey);
      inputElements.delete(focusKey);
      inputTestIds.delete(focusKey);
    }
    focusKeysByRow.delete(rowKey);
  };

  const draftValueForFocusKey = (focusKey: string) => draftValues.get(focusKey);

  return {
    clearAll() {
      draftValues.clear();
      inputElements.clear();
      inputTestIds.clear();
      focusKeysByRow.clear();
    },
    clearRow,
    clearScalarDraftsForRow(
      rowKey: string,
      preserveFocusKeys: ReadonlySet<string> = new Set(),
    ) {
      for (const binding of timelineScalarBindings) {
        for (const surface of timelineScalarEditorSurfaces) {
          const focusKey = inputFocusKey(rowKey, binding.key, surface);
          if (!preserveFocusKeys.has(focusKey)) {
            draftValues.delete(focusKey);
            forgetFocusKeyIfUnused(rowKey, focusKey);
          }
        }
      }
    },
    clearScalarDraftsForField(rowKey: string, field: keyof RowValues) {
      for (const surface of timelineScalarEditorSurfaces) {
        const focusKey = inputFocusKey(rowKey, field, surface);
        draftValues.delete(focusKey);
        forgetFocusKeyIfUnused(rowKey, focusKey);
      }
    },
    clearSubmittedRow(rowKey: string, submittedValues: RowValues) {
      for (const binding of timelineScalarBindings) {
        for (const surface of timelineScalarEditorSurfaces) {
          const focusKey = inputFocusKey(rowKey, binding.key, surface);
          if (draftValues.get(focusKey) === submittedValues[binding.key]) {
            draftValues.delete(focusKey);
            forgetFocusKeyIfUnused(rowKey, focusKey);
          }
        }
      }
    },
    deleteDraft(identity: TimelineScalarEditorIdentity) {
      const focusKey = inputFocusKey(
        identity.rowKey,
        identity.field,
        identity.surface,
      );
      draftValues.delete(focusKey);
      forgetFocusKeyIfUnused(identity.rowKey, focusKey);
    },
    deleteDraftForFocusKey(focusKey: string) {
      draftValues.delete(focusKey);
      for (const [rowKey, focusKeys] of focusKeysByRow) {
        if (focusKeys.has(focusKey)) {
          forgetFocusKeyIfUnused(rowKey, focusKey);
          break;
        }
      }
    },
    draftValue(identity: TimelineScalarEditorIdentity) {
      return draftValueForFocusKey(
        inputFocusKey(identity.rowKey, identity.field, identity.surface),
      );
    },
    draftValueForFocusKey,
    inputElementForFocusKey(focusKey: string) {
      return inputElements.get(focusKey) ?? null;
    },
    inputTestIdForFocusKey(focusKey: string) {
      return inputTestIds.get(focusKey) ?? null;
    },
    materializeRow(
      row: WorkbookRow,
      preferred?: {
        readonly field: keyof RowValues;
        readonly value: string | undefined;
      },
    ): WorkbookRow {
      let nextValues: RowValues | null = null;
      for (const binding of timelineScalarBindings) {
        let draftValue =
          preferred?.field === binding.key ? preferred.value : undefined;
        if (draftValue === undefined) {
          for (const surface of timelineScalarEditorSurfaces) {
            draftValue = draftValueForFocusKey(
              inputFocusKey(row.key, binding.key, surface),
            );
            if (draftValue !== undefined) break;
          }
        }
        if (
          draftValue === undefined ||
          draftValue === row.values[binding.key]
        ) {
          continue;
        }
        nextValues ??= { ...row.values };
        nextValues[binding.key] = draftValue;
      }
      return nextValues === null ? row : { ...row, values: nextValues };
    },
    registerInput(
      identity: TimelineInputIdentity,
      dataTestId: string,
      element: TimelineEditorElement | null,
    ) {
      const focusKey = inputFocusKey(
        identity.rowKey,
        identity.field,
        identity.surface,
      );
      if (element === null) {
        inputElements.delete(focusKey);
        inputTestIds.delete(focusKey);
        forgetFocusKeyIfUnused(identity.rowKey, focusKey);
        return;
      }
      rememberFocusKey(identity.rowKey, focusKey);
      inputElements.set(focusKey, element);
      inputTestIds.set(focusKey, dataTestId);
    },
    retainRows(rowKeys: ReadonlySet<string>) {
      for (const rowKey of focusKeysByRow.keys()) {
        if (!rowKeys.has(rowKey)) clearRow(rowKey);
      }
    },
    setDraft(identity: TimelineScalarEditorIdentity, value: string) {
      const focusKey = inputFocusKey(
        identity.rowKey,
        identity.field,
        identity.surface,
      );
      rememberFocusKey(identity.rowKey, focusKey);
      draftValues.set(focusKey, value);
    },
  };
}

/** Owns scalar invalid-draft and semantic input-ref lifetime for one schema. */
export function useTimelineEditorDraftRegistry(
  schemaKey: string,
): TimelineEditorDraftRegistry {
  return useMemo(() => {
    void schemaKey;
    return createTimelineEditorDraftRegistry();
  }, [schemaKey]);
}
