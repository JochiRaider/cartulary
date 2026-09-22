import type { WorkbookViewApiRow } from "../models/workbookContractRows";
import type { WorkbookResolvedMutation } from "../mutations/workbookConflictResolutionAdapter";
import type { WorkbookConflictEntry } from "./workbookConflictModel";

export type WorkbookSurfaceRefresh = () => Promise<void> | void;
export type WorkbookSurfaceResolvedMutationApply = (
  mutation: WorkbookResolvedMutation,
  conflict: WorkbookConflictEntry,
) => Promise<void> | void;
export type WorkbookSurfaceBlockedEditDiscard = (
  unitId: string,
) => Promise<boolean> | boolean;

export type WorkbookSurfaceBatchApply = (
  rows: readonly WorkbookViewApiRow[],
) => void;

type WorkbookSurfaceRegistration = {
  readonly applyBatch: WorkbookSurfaceBatchApply | null;
  readonly applyResolvedMutation: WorkbookSurfaceResolvedMutationApply | null;
  readonly discardBlockedEdit: WorkbookSurfaceBlockedEditDiscard | null;
  readonly refresh: WorkbookSurfaceRefresh;
};

/** Owns mounted surface callbacks and retained refresh debt. */
export class WorkbookSurfaceRegistry {
  #authorityGeneration = 0;
  readonly #registrations = new Map<string, WorkbookSurfaceRegistration>();
  readonly #dirtySurfaces = new Set<string>();
  readonly #debtGenerations = new Map<string, number>();
  readonly #refreshing = new Map<string, Promise<void>>();
  readonly #onDebtChanged: (viewSchemaId: string) => void;
  #debtSnapshot: readonly string[] = Object.freeze([]);
  #disposed = false;

  constructor(onDebtChanged: (viewSchemaId: string) => void) {
    this.#onDebtChanged = onDebtChanged;
  }

  register(
    viewSchemaId: string,
    refresh: WorkbookSurfaceRefresh,
    applyResolvedMutation?: WorkbookSurfaceResolvedMutationApply,
    discardBlockedEdit?: WorkbookSurfaceBlockedEditDiscard,
    applyBatch?: WorkbookSurfaceBatchApply,
  ): () => void {
    if (this.#disposed) return () => {};
    const registration = {
      applyBatch: applyBatch ?? null,
      applyResolvedMutation: applyResolvedMutation ?? null,
      discardBlockedEdit: discardBlockedEdit ?? null,
      refresh,
    };
    this.#registrations.set(viewSchemaId, registration);
    if (this.#dirtySurfaces.has(viewSchemaId)) {
      void this.refreshRequired(viewSchemaId).catch(() => {
        if (!this.#disposed) this.#onDebtChanged(viewSchemaId);
      });
    }
    return () => {
      if (this.#registrations.get(viewSchemaId) === registration) {
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

  refreshDebts(): readonly string[] {
    const next = [
      ...new Set([...this.#dirtySurfaces, ...this.#refreshing.keys()]),
    ];
    if (
      next.length !== this.#debtSnapshot.length ||
      next.some((id, index) => id !== this.#debtSnapshot[index])
    )
      this.#debtSnapshot = Object.freeze(next);
    return this.#debtSnapshot;
  }

  requiresRefresh(viewSchemaId: string): boolean {
    return (
      this.#refreshing.has(viewSchemaId) ||
      this.#dirtySurfaces.has(viewSchemaId)
    );
  }
  applyBatch(viewSchemaId: string, rows: readonly WorkbookViewApiRow[]) {
    this.#registrations.get(viewSchemaId)?.applyBatch?.(rows);
  }
  invalidateAuthority(): void {
    this.#authorityGeneration++;
  }
  invalidate(viewSchemaId: string): void {
    if (this.#disposed) return;
    this.#debtGenerations.set(
      viewSchemaId,
      (this.#debtGenerations.get(viewSchemaId) ?? 0) + 1,
    );
    this.#dirtySurfaces.add(viewSchemaId);
  }
  async refreshRequired(viewSchemaId: string): Promise<void> {
    this.invalidate(viewSchemaId);
    return this.reconcile(viewSchemaId);
  }
  private async reconcile(viewSchemaId: string): Promise<void> {
    if (this.#disposed) throw new Error("The surface registry is retired.");
    const pending = this.#refreshing.get(viewSchemaId);
    if (pending) {
      await pending.catch(() => {
        /* This caller retries the retained read debt. */
      });
      if (this.#dirtySurfaces.has(viewSchemaId))
        return this.reconcile(viewSchemaId);
      return;
    }
    const registration = this.#registrations.get(viewSchemaId);
    if (!registration)
      throw new Error("The originating surface needs a refresh.");
    const generation = this.#debtGenerations.get(viewSchemaId);
    const authorityGeneration = this.#authorityGeneration;
    const running = Promise.resolve()
      .then(() => {
        if (
          this.#disposed ||
          authorityGeneration !== this.#authorityGeneration ||
          this.#registrations.get(viewSchemaId) !== registration
        )
          throw new Error("The surface refresh binding is obsolete.");
        return registration.refresh();
      })
      .then(() => {
        if (authorityGeneration !== this.#authorityGeneration)
          throw new Error("Authorization changed during refresh.");
        if (this.#registrations.get(viewSchemaId) !== registration)
          throw new Error("The surface changed during refresh.");
        if (this.#debtGenerations.get(viewSchemaId) === generation)
          this.#dirtySurfaces.delete(viewSchemaId);
      })
      .finally(() => {
        if (this.#refreshing.get(viewSchemaId) === running)
          this.#refreshing.delete(viewSchemaId);
        if (!this.#disposed) this.#onDebtChanged(viewSchemaId);
      });
    this.#refreshing.set(viewSchemaId, running);
    await running;
    if (this.#dirtySurfaces.has(viewSchemaId))
      return this.reconcile(viewSchemaId);
  }
  async refreshIfMounted(viewSchemaId: string): Promise<void> {
    if (this.#disposed) return;
    if (this.#registrations.has(viewSchemaId))
      return this.refreshRequired(viewSchemaId);
    this.#dirtySurfaces.add(viewSchemaId);
    this.#onDebtChanged(viewSchemaId);
  }
  async refresh(viewSchemaId: string): Promise<void> {
    try {
      await this.refreshRequired(viewSchemaId);
    } catch {
      /* Refresh debt is retained for ordinary autosave recovery. */
    }
  }

  dispose(): void {
    if (this.#disposed) return;
    this.#disposed = true;
    this.#authorityGeneration++;
    this.#registrations.clear();
    this.#dirtySurfaces.clear();
    this.#debtGenerations.clear();
    this.#refreshing.clear();
    this.#debtSnapshot = Object.freeze([]);
  }
}
