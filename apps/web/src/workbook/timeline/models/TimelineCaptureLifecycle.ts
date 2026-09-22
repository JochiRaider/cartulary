import type { WorkbookLocalDraftStore } from "../../models/WorkbookLocalDraftStore";
import {
  inputFocusKey,
  timelineCollectionBindings,
  timelineScalarBindings,
  timelineScalarEditorSurfaces,
} from "./timelineFieldRegistry";
import type { WorkbookRow } from "./timelineRowModel";

/** Runtime-owned identity and promotion. Values and revisions stay in the draft store. */
export class TimelineCaptureLifecycle {
  private nextIndex = 1;
  private readonly aliases = new Map<string, string>();
  private readonly captures = new Set<string>();
  constructor(private readonly store: WorkbookLocalDraftStore) {}
  allocateDraftIndex = () => this.nextIndex++;
  resolveRowKey = (key: string) => this.aliases.get(key) ?? key;
  beginCapture = (key: string) => {
    const first = !this.captures.has(key);
    this.captures.add(key);
    return first;
  };
  hasCapture = (key: string) => this.captures.has(key);
  promote(key: string, committed: WorkbookRow) {
    if (key === committed.key || this.aliases.has(key)) return;
    this.aliases.set(key, committed.key);
    this.captures.delete(key);
    for (const binding of [
      ...timelineScalarBindings,
      ...timelineCollectionBindings,
    ]) {
      const field = "key" in binding ? binding.key : binding.draftKey;
      for (const surface of timelineScalarEditorSurfaces) {
        const from = inputFocusKey(key, field, surface);
        const to = inputFocusKey(committed.key, field, surface);
        if (!this.store.draftValues.has(from)) continue;
        // Never replace an independently newer editor context at the destination.
        if (this.store.revision(to) < this.store.revision(from)) {
          this.store.move(from, to);
          if ("key" in binding)
            this.store.advanceBaseline(
              to,
              committed.committedValues[binding.key],
            );
          const keys =
            this.store.focusKeysByRow.get(committed.key) ?? new Set<string>();
          keys.add(to);
          this.store.focusKeysByRow.set(committed.key, keys);
        } else this.store.remove(from);
      }
    }
  }
  retireUnreferenced(referenced: ReadonlySet<string>) {
    const retired: string[] = [];
    for (const [key, accepted] of this.aliases) {
      const hasDraft = [
        ...(this.store.focusKeysByRow.get(accepted) ?? []),
      ].some((focusKey) => this.store.draftValues.has(focusKey));
      if (!referenced.has(key) && !hasDraft) {
        retired.push(key);
        this.aliases.delete(key);
        this.store.focusKeysByRow.delete(key);
      }
    }
    return retired;
  }
  clear() {
    this.aliases.clear();
    this.captures.clear();
  }
}
