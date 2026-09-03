import type { WorkbookResolvedMutation } from "../mutations/workbookConflictResolutionAdapter";
import type { WorkbookConflictEntry } from "./workbookConflictModel";

export type WorkbookSurfaceRefresh = () => Promise<void> | void;
export type WorkbookSurfaceResolvedMutationApply = (
  mutation: WorkbookResolvedMutation,
  conflict: WorkbookConflictEntry,
) => Promise<void> | void;
export type WorkbookSurfaceConflictFocusRestore = (
  conflict: WorkbookConflictEntry,
) => void;
export type WorkbookSurfaceBlockedEditDiscard = (
  unitId: string,
) => Promise<boolean> | boolean;

type WorkbookSurfaceRegistration = {
  readonly applyResolvedMutation: WorkbookSurfaceResolvedMutationApply | null;
  readonly discardBlockedEdit: WorkbookSurfaceBlockedEditDiscard | null;
  readonly refresh: WorkbookSurfaceRefresh;
  readonly restoreConflictFocus: WorkbookSurfaceConflictFocusRestore | null;
};

/** Owns mounted surface callbacks and retained refresh debt. */
export class WorkbookSurfaceRegistry {
  readonly #registrations = new Map<string, WorkbookSurfaceRegistration>();
  readonly #dirtySurfaces = new Set<string>();
  readonly #onDebtChanged: () => void;

  constructor(onDebtChanged: () => void) {
    this.#onDebtChanged = onDebtChanged;
  }

  register(
    viewSchemaId: string,
    refresh: WorkbookSurfaceRefresh,
    applyResolvedMutation?: WorkbookSurfaceResolvedMutationApply,
    restoreConflictFocus?: WorkbookSurfaceConflictFocusRestore,
    discardBlockedEdit?: WorkbookSurfaceBlockedEditDiscard,
  ): () => void {
    this.#registrations.set(viewSchemaId, {
      applyResolvedMutation: applyResolvedMutation ?? null,
      discardBlockedEdit: discardBlockedEdit ?? null,
      refresh,
      restoreConflictFocus: restoreConflictFocus ?? null,
    });
    if (this.#dirtySurfaces.delete(viewSchemaId)) {
      void Promise.resolve(refresh()).catch(() => {
        this.#dirtySurfaces.add(viewSchemaId);
        this.#onDebtChanged();
      });
    }
    return () => {
      if (this.#registrations.get(viewSchemaId)?.refresh === refresh) {
        this.#registrations.delete(viewSchemaId);
      }
    };
  }

  applyResolvedMutation(viewSchemaId: string) {
    return this.#registrations.get(viewSchemaId)?.applyResolvedMutation ?? null;
  }

  discardBlockedEdit(viewSchemaId: string) {
    return this.#registrations.get(viewSchemaId)?.discardBlockedEdit ?? null;
  }

  restoreConflictFocus(viewSchemaId: string) {
    return this.#registrations.get(viewSchemaId)?.restoreConflictFocus ?? null;
  }

  async refresh(viewSchemaId: string): Promise<void> {
    const refresh = this.#registrations.get(viewSchemaId)?.refresh;
    if (refresh === undefined) {
      this.#dirtySurfaces.add(viewSchemaId);
      return;
    }
    try {
      await refresh();
      this.#dirtySurfaces.delete(viewSchemaId);
    } catch {
      this.#dirtySurfaces.add(viewSchemaId);
    }
  }
}
