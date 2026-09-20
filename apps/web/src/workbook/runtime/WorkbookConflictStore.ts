import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";
import {
  type WorkbookConflictEntry,
  workbookConflictEntry,
} from "./workbookConflictModel";

export type WorkbookConflictRefresh = () => Promise<
  WorkbookOperationOutcome<unknown>
>;

export type WorkbookConflictRegistration = {
  readonly draftRevisions?: ReadonlyMap<string, number> | undefined;
  readonly batchOperationId?: string | undefined;
  readonly compoundOperationId?: string | undefined;
  readonly focusOrigin?: "grid" | "inspector" | undefined;
  readonly sheetRef?: SheetRef | undefined;
  readonly conflict: Parameters<typeof workbookConflictEntry>[0]["conflict"];
  readonly focusKey?: string | null | undefined;
  readonly refresh?: WorkbookConflictRefresh | undefined;
  readonly rowLabel: string;
  readonly surfaceLabel: string;
  readonly viewSchemaId: string;
};

/** Owns conflict drafts and refresh callbacks. */
class WorkbookConflictState {
  readonly #entries = new Map<string, WorkbookConflictEntry>();
  readonly #refreshByKey = new Map<string, WorkbookConflictRefresh>();
  #snapshot: readonly WorkbookConflictEntry[] | null = null;

  get size(): number {
    return this.#entries.size;
  }

  entries(): readonly WorkbookConflictEntry[] {
    if (this.#snapshot === null)
      this.#snapshot = Object.freeze(Array.from(this.#entries.values()));
    return this.#snapshot;
  }

  get(key: string): WorkbookConflictEntry | undefined {
    return this.#entries.get(key);
  }

  register(registration: WorkbookConflictRegistration): WorkbookConflictEntry {
    this.#snapshot = null;
    const entry = workbookConflictEntry(registration);
    const current = this.#entries.get(entry.key);
    this.#entries.set(
      entry.key,
      current === undefined
        ? entry
        : {
            ...entry,
            mergedDraft:
              current.resolutionClass === entry.resolutionClass
                ? current.mergedDraft
                : entry.mergedDraft,
          },
    );
    if (registration.refresh !== undefined) {
      this.#refreshByKey.set(entry.key, registration.refresh);
    }
    return entry;
  }

  replace(entry: WorkbookConflictEntry): void {
    if (this.#entries.get(entry.key) === entry) return;
    this.#snapshot = null;
    this.#entries.set(entry.key, entry);
  }

  setRefresh(key: string, refresh: WorkbookConflictRefresh): void {
    this.#refreshByKey.set(key, refresh);
  }

  refresh(key: string): WorkbookConflictRefresh | undefined {
    return this.#refreshByKey.get(key);
  }

  updateDraft(key: string, mergedDraft: string): boolean {
    const conflict = this.#entries.get(key);
    if (conflict === undefined) return false;
    // An existing draft update is still an owner event; only its observation
    // may remain unchanged. The runtime must continue publishing/waking work.
    if (conflict.mergedDraft === mergedDraft) return true;
    this.#snapshot = null;
    this.#entries.set(key, { ...conflict, mergedDraft });
    return true;
  }

  clear(key: string): WorkbookConflictEntry | undefined {
    const conflict = this.#entries.get(key);
    if (conflict !== undefined) this.#snapshot = null;
    this.#entries.delete(key);
    this.#refreshByKey.delete(key);
    return conflict;
  }
}

export function createWorkbookConflictStore() {
  const state = new WorkbookConflictState();
  return {
    get size() {
      return state.size;
    },
    entries: () => state.entries(),
    get: (key: string) => state.get(key),
    register: (registration: WorkbookConflictRegistration) =>
      state.register(registration),
    replace: (entry: WorkbookConflictEntry) => state.replace(entry),
    setRefresh: (key: string, refresh: WorkbookConflictRefresh) =>
      state.setRefresh(key, refresh),
    refresh: (key: string) => state.refresh(key),
    updateDraft: (key: string, mergedDraft: string) =>
      state.updateDraft(key, mergedDraft),
    clear: (key: string) => state.clear(key),
  };
}

export type WorkbookConflictStore = ReturnType<
  typeof createWorkbookConflictStore
>;
