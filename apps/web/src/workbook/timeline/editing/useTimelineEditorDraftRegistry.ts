import { useMemo } from "react";
import { WorkbookLocalDraftStore } from "../../models/WorkbookLocalDraftStore";
import {
  type FocusFieldKey,
  inputFocusKey,
  type RowValues,
  type TimelineScalarEditorSurface,
  timelineCollectionBindings,
  timelineScalarBindings,
  timelineScalarEditorSurfaces,
} from "../models/timelineFieldRegistry";
import type { WorkbookRow } from "../models/timelineRowModel";

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

function isUsableEditorElement(
  element: TimelineEditorElement | undefined,
): element is TimelineEditorElement {
  if (
    element === undefined ||
    !element.isConnected ||
    element.disabled ||
    element.hidden ||
    element.closest("[hidden], [aria-hidden='true']") !== null
  ) {
    return false;
  }
  const style = window.getComputedStyle(element);
  return style.display !== "none" && style.visibility !== "hidden";
}

export type TimelineEditorDraftRegistry = ReturnType<
  typeof createTimelineEditorDraftRegistry
>;

export function createTimelineEditorDraftRegistry(
  store = new WorkbookLocalDraftStore(),
) {
  const { draftValues, focusKeysByRow } = store;
  const acceptedDraftRows = new Map<string, string>();
  const captureRows = new Set<string>();
  const inputElements = new Map<string, TimelineEditorElement>();
  const rowListeners = new Map<string, Set<() => void>>();
  const publishRow = (rowKey: string) => {
    for (const listener of rowListeners.get(rowKey) ?? []) listener();
  };

  const rememberFocusKey = (rowKey: string, focusKey: string) => {
    const rowFocusKeys = focusKeysByRow.get(rowKey) ?? new Set<string>();
    rowFocusKeys.add(focusKey);
    focusKeysByRow.set(rowKey, rowFocusKeys);
  };

  const forgetFocusKeyIfUnused = (rowKey: string, focusKey: string) => {
    if (draftValues.has(focusKey) || inputElements.has(focusKey)) {
      return;
    }
    const rowFocusKeys = focusKeysByRow.get(rowKey);
    rowFocusKeys?.delete(focusKey);
    if (rowFocusKeys?.size === 0) focusKeysByRow.delete(rowKey);
  };

  const clearRow = (rowKey: string) => {
    for (const focusKey of focusKeysByRow.get(rowKey) ?? []) {
      store.remove(focusKey);
      inputElements.delete(focusKey);
    }
    focusKeysByRow.delete(rowKey);
    publishRow(rowKey);
  };

  const draftValueForFocusKey = (focusKey: string) => draftValues.get(focusKey);

  return {
    subscribe: store.subscribe,
    getSnapshot: store.getSnapshot,
    retainedGridDrafts() {
      return [...focusKeysByRow.keys()]
        .filter((key) => !key.startsWith("draft-"))
        .flatMap((rowKey) =>
          timelineScalarBindings.flatMap((binding) => {
            const key = inputFocusKey(rowKey, binding.key, "grid"),
              value = draftValues.get(key);
            return value === undefined
              ? []
              : [
                  {
                    key,
                    rowKey,
                    fieldKey: binding.fieldKey,
                    value,
                    discard: () => {
                      store.remove(key);
                      forgetFocusKeyIfUnused(rowKey, key);
                      publishRow(rowKey);
                    },
                  },
                ];
          }),
        );
    },
    resolveRowKey(rowKey: string) {
      return acceptedDraftRows.get(rowKey) ?? rowKey;
    },
    beginCapture(rowKey: string) {
      const firstInput = !captureRows.has(rowKey);
      captureRows.add(rowKey);
      return firstInput;
    },
    acceptCapture(rowKey: string, committed: WorkbookRow) {
      if (!captureRows.delete(rowKey)) return null;
      acceptedDraftRows.set(rowKey, committed.key);
      let active: {
        fieldKey: string;
        value: string;
        selectionRange: { start: number; end: number };
      } | null = null;
      for (const binding of [
        ...timelineScalarBindings,
        ...timelineCollectionBindings,
      ]) {
        const field = "key" in binding ? binding.key : binding.draftKey;
        for (const surface of timelineScalarEditorSurfaces) {
          const oldKey = inputFocusKey(rowKey, field, surface);
          const nextKey = inputFocusKey(committed.key, field, surface);
          const value = draftValues.get(oldKey);
          if (value !== undefined) {
            store.move(oldKey, nextKey);
            if ("key" in binding)
              store.advanceBaseline(
                nextKey,
                committed.committedValues[binding.key],
              );
            rememberFocusKey(committed.key, nextKey);
            store.remove(oldKey);
          }
          const element = inputElements.get(oldKey);
          if (
            surface === "grid" &&
            element !== undefined &&
            element === document.activeElement &&
            "key" in binding
          ) {
            active = {
              fieldKey: binding.fieldKey,
              value: element.value,
              selectionRange: {
                start: element.selectionStart ?? element.value.length,
                end: element.selectionEnd ?? element.value.length,
              },
            };
          }
        }
      }
      publishRow(committed.key);
      return active;
    },
    subscribeRow(rowKey: string, listener: () => void) {
      const listeners = rowListeners.get(rowKey) ?? new Set<() => void>();
      listeners.add(listener);
      rowListeners.set(rowKey, listeners);
      return () => {
        listeners.delete(listener);
        if (!listeners.size) rowListeners.delete(rowKey);
      };
    },
    rowDraftSnapshot(rowKey: string) {
      return JSON.stringify(
        [...(focusKeysByRow.get(rowKey) ?? [])]
          .filter((key) => draftValues.has(key))
          .sort()
          .map((key) => [key, draftValues.get(key)]),
      );
    },
    clearAll() {
      store.clear();
      acceptedDraftRows.clear();
      captureRows.clear();
      inputElements.clear();
      for (const rowKey of rowListeners.keys()) publishRow(rowKey);
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
            store.remove(focusKey);
            forgetFocusKeyIfUnused(rowKey, focusKey);
          }
        }
      }
      publishRow(rowKey);
    },
    clearCapturedScalarField(
      rowKey: string,
      field: keyof RowValues,
      revisions?: ReadonlyMap<string, number>,
    ) {
      let cleared = false;
      for (const surface of timelineScalarEditorSurfaces) {
        const focusKey = inputFocusKey(rowKey, field, surface);
        if (!revisions || revisions.get(focusKey) !== store.revision(focusKey))
          continue;
        store.remove(focusKey);
        cleared = true;
        forgetFocusKeyIfUnused(rowKey, focusKey);
      }
      publishRow(rowKey);
      return cleared;
    },
    clearSubmittedRow(
      rowKey: string,
      submittedValues: RowValues,
      submittedCollections?: Partial<WorkbookRow["collectionDrafts"]>,
      revisions?: ReadonlyMap<string, number>,
    ) {
      const originalRowKey = rowKey;
      rowKey = acceptedDraftRows.get(rowKey) ?? rowKey;
      const owns = (
        field: FocusFieldKey,
        surface: TimelineScalarEditorSurface,
      ) =>
        revisions === undefined ||
        revisions.get(inputFocusKey(originalRowKey, field, surface)) ===
          store.revision(inputFocusKey(rowKey, field, surface));
      for (const binding of timelineCollectionBindings) {
        const focusKey = inputFocusKey(rowKey, binding.draftKey, "grid");
        if (
          submittedCollections !== undefined &&
          owns(binding.draftKey, "grid") &&
          draftValues.get(focusKey) === submittedCollections[binding.draftKey]
        ) {
          store.remove(focusKey);
          // A socket may have mounted this version before HTTP settlement.
          // Clear only this accepted submission in each mounted presentation.
          for (const surface of timelineScalarEditorSurfaces) {
            const element = inputElements.get(
              inputFocusKey(rowKey, binding.draftKey, surface),
            );
            if (
              element &&
              element.value === submittedCollections[binding.draftKey]
            )
              element.value = "";
          }
          forgetFocusKeyIfUnused(rowKey, focusKey);
        }
      }
      for (const binding of timelineScalarBindings) {
        for (const surface of timelineScalarEditorSurfaces) {
          const focusKey = inputFocusKey(rowKey, binding.key, surface);
          if (
            owns(binding.key, surface) &&
            draftValues.get(focusKey) === submittedValues[binding.key]
          ) {
            store.remove(focusKey);
            forgetFocusKeyIfUnused(rowKey, focusKey);
          }
        }
      }
      publishRow(rowKey);
    },
    deleteDraft(identity: TimelineScalarEditorIdentity) {
      const focusKey = inputFocusKey(
        identity.rowKey,
        identity.field,
        identity.surface,
      );
      store.remove(focusKey);
      forgetFocusKeyIfUnused(identity.rowKey, focusKey);
      publishRow(identity.rowKey);
    },
    deleteDraftForFocusKey(focusKey: string) {
      store.remove(focusKey);
      for (const [rowKey, focusKeys] of focusKeysByRow) {
        if (focusKeys.has(focusKey)) {
          forgetFocusKeyIfUnused(rowKey, focusKey);
          publishRow(rowKey);
          break;
        }
      }
    },
    draftValue(identity: TimelineInputIdentity) {
      return draftValueForFocusKey(
        inputFocusKey(identity.rowKey, identity.field, identity.surface),
      );
    },
    draftValueForFocusKey,
    captureRow(rowKey: string, surface: TimelineScalarEditorSurface) {
      return new Map(
        [
          ...timelineScalarBindings.map((binding) =>
            inputFocusKey(rowKey, binding.key, surface),
          ),
          ...timelineCollectionBindings.map((binding) =>
            inputFocusKey(rowKey, binding.draftKey, "grid"),
          ),
        ]
          .filter((key) => draftValues.has(key))
          .map((key) => [key, store.revision(key)]),
      );
    },
    inputElementForFocusKey(focusKey: string) {
      const separator = focusKey.indexOf(":");
      const acceptedRowKey = acceptedDraftRows.get(
        focusKey.slice(0, separator),
      );
      if (separator >= 0 && acceptedRowKey !== undefined)
        focusKey = acceptedRowKey + focusKey.slice(separator);
      const element = inputElements.get(focusKey);
      if (isUsableEditorElement(element)) return element;
      if (element !== undefined) inputElements.delete(focusKey);
      return null;
    },
    materializeRow(
      row: WorkbookRow,
      preferred?: {
        readonly field: keyof RowValues;
        readonly value: string | undefined;
        readonly surface?: TimelineScalarEditorSurface;
      },
    ): WorkbookRow {
      let nextValues: RowValues | null =
        preferred?.surface === "inspector" && row.recordId !== null
          ? { ...row.committedValues }
          : null;
      for (const binding of timelineScalarBindings) {
        let draftValue =
          preferred?.field === binding.key ? preferred.value : undefined;
        if (draftValue === undefined) {
          draftValue = draftValueForFocusKey(
            inputFocusKey(row.key, binding.key, preferred?.surface ?? "grid"),
          );
        }
        if (
          draftValue === undefined ||
          draftValue === (nextValues ?? row.values)[binding.key]
        ) {
          continue;
        }
        nextValues ??= { ...row.values };
        nextValues[binding.key] = draftValue;
      }
      let collections = row.collectionDrafts;
      for (const binding of timelineCollectionBindings) {
        const value = draftValueForFocusKey(
          inputFocusKey(row.key, binding.draftKey, "grid"),
        );
        if (value !== undefined && value !== collections[binding.draftKey])
          collections = { ...collections, [binding.draftKey]: value };
      }
      return nextValues === null && collections === row.collectionDrafts
        ? row
        : {
            ...row,
            values: nextValues ?? row.values,
            collectionDrafts: collections,
          };
    },
    authoringRow(
      row: WorkbookRow,
      surface: TimelineScalarEditorSurface,
    ): WorkbookRow {
      const committedValues = { ...row.committedValues };
      for (const binding of timelineScalarBindings) {
        const baseline = store.baseline(
          inputFocusKey(row.key, binding.key, surface),
        );
        if (baseline !== undefined) committedValues[binding.key] = baseline;
      }
      return { ...row, committedValues };
    },
    needsReview(identity: TimelineScalarEditorIdentity, row: WorkbookRow) {
      const baseline = store.baseline(
        inputFocusKey(identity.rowKey, identity.field, identity.surface),
      );
      return (
        baseline !== undefined &&
        baseline !== row.committedValues[identity.field]
      );
    },
    review(identity: TimelineScalarEditorIdentity, row: WorkbookRow) {
      store.reviewBaseline(
        inputFocusKey(identity.rowKey, identity.field, identity.surface),
        row.committedValues[identity.field],
      );
      publishRow(identity.rowKey);
    },
    acceptPredecessor(
      row: WorkbookRow,
      fields: readonly string[],
      previousValues: RowValues,
    ) {
      for (const binding of timelineScalarBindings) {
        if (!fields.includes(binding.fieldKey)) continue;
        for (const surface of timelineScalarEditorSurfaces) {
          const key = inputFocusKey(row.key, binding.key, surface);
          if (store.baseline(key) === previousValues[binding.key])
            store.advanceBaseline(key, row.committedValues[binding.key]);
        }
      }
      publishRow(row.key);
    },
    registerInput(
      identity: TimelineInputIdentity,
      element: TimelineEditorElement | null,
    ) {
      const focusKey = inputFocusKey(
        identity.rowKey,
        identity.field,
        identity.surface,
      );
      if (element === null) {
        inputElements.delete(focusKey);
        forgetFocusKeyIfUnused(identity.rowKey, focusKey);
        return;
      }
      rememberFocusKey(identity.rowKey, focusKey);
      inputElements.set(focusKey, element);
    },
    retainRows(rowKeys: ReadonlySet<string>) {
      for (const rowKey of focusKeysByRow.keys()) {
        if (!rowKeys.has(rowKey)) {
          for (const key of focusKeysByRow.get(rowKey) ?? [])
            inputElements.delete(key);
        }
      }
    },
    setDraft(
      identity: TimelineInputIdentity,
      value: string,
      baseline?: WorkbookRow,
    ) {
      const focusKey = inputFocusKey(
        identity.rowKey,
        identity.field,
        identity.surface,
      );
      rememberFocusKey(identity.rowKey, focusKey);
      store.write(
        focusKey,
        value,
        baseline && identity.field in baseline.committedValues
          ? baseline.committedValues[identity.field as keyof RowValues]
          : undefined,
      );
      publishRow(identity.rowKey);
    },
  };
}

/** Rebinds mounted input refs without replacing the runtime-owned local draft values. */
export function useTimelineEditorDraftRegistry(
  store: WorkbookLocalDraftStore,
): TimelineEditorDraftRegistry {
  return useMemo(() => createTimelineEditorDraftRegistry(store), [store]);
}
