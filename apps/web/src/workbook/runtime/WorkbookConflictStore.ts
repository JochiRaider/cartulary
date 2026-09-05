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
  readonly sheetRef?: SheetRef | undefined;
  readonly conflict: Parameters<typeof workbookConflictEntry>[0]["conflict"];
  readonly focusKey?: string | null | undefined;
  readonly refresh?: WorkbookConflictRefresh | undefined;
  readonly rowLabel: string;
  readonly surfaceLabel: string;
  readonly viewSchemaId: string;
};

/** Owns conflict drafts, refresh callbacks, and panel activation state. */
class WorkbookConflictState {
  readonly #entries = new Map<string, WorkbookConflictEntry>();
  readonly #refreshByKey = new Map<string, WorkbookConflictRefresh>();
  #panelDismissed = false;

  get size(): number {
    return this.#entries.size;
  }

  get panelOpen(): boolean {
    return this.#entries.size > 0 && !this.#panelDismissed;
  }

  entries(): readonly WorkbookConflictEntry[] {
    return Array.from(this.#entries.values());
  }

  get(key: string): WorkbookConflictEntry | undefined {
    return this.#entries.get(key);
  }

  register(registration: WorkbookConflictRegistration): WorkbookConflictEntry {
    const entry = workbookConflictEntry(registration);
    const current = this.#entries.get(entry.key);
    if (current === undefined) this.#panelDismissed = false;
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
    this.#entries.set(key, { ...conflict, mergedDraft });
    return true;
  }

  clear(key: string): WorkbookConflictEntry | undefined {
    const conflict = this.#entries.get(key);
    this.#entries.delete(key);
    this.#refreshByKey.delete(key);
    this.#panelDismissed = false;
    return conflict;
  }

  activate(): void {
    this.#panelDismissed = false;
  }

  dismiss(key: string): WorkbookConflictEntry | undefined {
    const conflict = this.#entries.get(key);
    if (conflict === undefined) return undefined;
    this.#panelDismissed = true;
    return conflict;
  }
}

export function createWorkbookConflictStore() {
  const state = new WorkbookConflictState();
  return {
    get size() {
      return state.size;
    },
    get panelOpen() {
      return state.panelOpen;
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
    activate: () => state.activate(),
    dismiss: (key: string) => state.dismiss(key),
  };
}

export type WorkbookConflictStore = ReturnType<
  typeof createWorkbookConflictStore
>;
