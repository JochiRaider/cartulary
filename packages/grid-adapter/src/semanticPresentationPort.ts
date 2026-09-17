import {
  type GridPresentationSnapshot,
  gridRowIdentitiesEqual,
  gridSurfaceIdentitiesEqual,
} from "./core";

/** Observes the existing presentation model without retaining row payloads. */
export function createGridPresentationPort() {
  let snapshot: GridPresentationSnapshot | null = null;
  let revision = 0;
  const listeners = new Set<() => void>();
  const emit = () => {
    for (const listener of listeners) listener();
  };
  return {
    getSnapshot: () => snapshot,
    subscribe(listener: () => void) {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
    publish(model: Omit<GridPresentationSnapshot, "revision">) {
      if (
        snapshot &&
        gridSurfaceIdentitiesEqual(snapshot.surface, model.surface) &&
        snapshot.fieldKeys.length === model.fieldKeys.length &&
        snapshot.fieldKeys.every(
          (key, index) => key === model.fieldKeys[index],
        ) &&
        snapshot.rowIdentities.length === model.rowIdentities.length &&
        snapshot.rowIdentities.every((row, index) => {
          const next = model.rowIdentities[index];
          return next !== undefined && gridRowIdentitiesEqual(row, next);
        })
      )
        return;
      snapshot = {
        surface: model.surface,
        fieldKeys: [...model.fieldKeys],
        rowIdentities: [...model.rowIdentities],
        revision: ++revision,
      };
      emit();
    },
    retire() {
      if (snapshot === null) return;
      snapshot = null;
      emit();
    },
  };
}
