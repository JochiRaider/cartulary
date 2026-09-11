import {
  emptyLifecycleValues,
  freezeLifecycle,
  type LifecycleDraft,
  type LifecycleDraftValues,
} from "./indicatorLifecycleModel";

/** Borrowed by one account/incident owner; presentation has no storage lifetime. */
export class IndicatorLifecycleDraftStore {
  private readonly drafts = new Map<string, LifecycleDraft>();
  constructor(private readonly changed: () => void) {}
  get(recordId: string) {
    return this.drafts.get(recordId) ?? null;
  }
  open(recordId: string, label: string, version: number) {
    const existing = this.get(recordId);
    if (existing) return existing;
    const draft = freezeLifecycle({
      recordId,
      label,
      baseRowVersion: version,
      revision: 0,
      values: structuredClone(emptyLifecycleValues),
    });
    this.drafts.set(recordId, draft);
    this.changed();
    return draft;
  }
  update(recordId: string, values: LifecycleDraftValues) {
    const draft = this.get(recordId);
    if (!draft) return;
    this.drafts.set(
      recordId,
      freezeLifecycle({
        ...draft,
        revision: draft.revision + 1,
        values: structuredClone(values),
      }),
    );
    this.changed();
  }
  review(recordId: string, version: number) {
    const draft = this.get(recordId);
    if (
      !draft ||
      !Number.isSafeInteger(version) ||
      version < draft.baseRowVersion
    )
      return;
    this.drafts.set(
      recordId,
      freezeLifecycle({
        ...draft,
        revision: draft.revision + 1,
        baseRowVersion: version,
      }),
    );
    this.changed();
  }
  discard(recordId: string) {
    if (this.drafts.delete(recordId)) this.changed();
  }
  clear() {
    this.drafts.clear();
    this.changed();
  }
}
