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
  readonly #refreshing = new Map<string, Promise<void>>();
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
    if (this.#dirtySurfaces.has(viewSchemaId)) {
      void this.refreshRequired(viewSchemaId).catch(() => {
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

  requiresRefresh(viewSchemaId: string): boolean {
    return (
      this.#refreshing.has(viewSchemaId) ||
      this.#dirtySurfaces.has(viewSchemaId)
    );
  }
  async refreshRequired(viewSchemaId: string): Promise<void> {
    const pending = this.#refreshing.get(viewSchemaId);
    if (pending) return pending;
    const refresh = this.#registrations.get(viewSchemaId)?.refresh;
    if (!refresh) {
      this.#dirtySurfaces.add(viewSchemaId);
      throw new Error("The originating surface needs a refresh.");
    }
    const running = Promise.resolve()
      .then(refresh)
      .then(() => {
        this.#dirtySurfaces.delete(viewSchemaId);
      })
      .catch((error: unknown) => {
        this.#dirtySurfaces.add(viewSchemaId);
        throw error;
      })
      .finally(() => {
        if (this.#refreshing.get(viewSchemaId) === running)
          this.#refreshing.delete(viewSchemaId);
      });
    this.#refreshing.set(viewSchemaId, running);
    return running;
  }
  async refreshIfMounted(viewSchemaId: string): Promise<void> {
    if (this.#registrations.has(viewSchemaId))
      return this.refreshRequired(viewSchemaId);
    this.#dirtySurfaces.add(viewSchemaId);
    this.#onDebtChanged();
  }
  async refresh(viewSchemaId: string): Promise<void> {
    try {
      await this.refreshRequired(viewSchemaId);
    } catch {
      /* Refresh debt is retained for ordinary autosave recovery. */
    }
  }
}
