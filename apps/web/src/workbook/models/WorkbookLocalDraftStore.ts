/** Memory-only authoring drafts keyed by a source owner's opaque row and editor identities.
 * Mounted controls, capture algorithms and listeners remain source-owned.
 */
export class WorkbookLocalDraftStore {
  readonly draftValues = new Map<string, string>();
  readonly focusKeysByRow = new Map<string, Set<string>>();
  private readonly revisions = new Map<string, number>();
  private readonly baselines = new Map<string, string>();
  private sequence = 0;
  private snapshot = 0;
  private batchDepth = 0;
  private changed = false;
  batch = (work: () => void) => {
    this.batchDepth++;
    try {
      work();
    } finally {
      this.batchDepth--;
      if (!this.batchDepth && this.changed) {
        this.changed = false;
        this.publish();
      }
    }
  };
  private readonly listeners = new Set<() => void>();
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  getSnapshot = () => this.snapshot;
  private publish() {
    if (this.batchDepth) {
      this.changed = true;
      return;
    }
    this.snapshot++;
    for (const listener of this.listeners) listener();
  }

  write(
    key: string,
    value: string,
    baseline?: string,
    newRevision = false,
  ): void {
    if (!newRevision && this.draftValues.get(key) === value) return;
    if (!this.draftValues.has(key) && baseline !== undefined)
      this.baselines.set(key, baseline);
    this.draftValues.set(key, value);
    this.revisions.set(key, ++this.sequence);
    this.publish();
  }
  revision(key: string): number {
    return this.revisions.get(key) ?? 0;
  }
  baseline(key: string): string | undefined {
    return this.baselines.get(key);
  }
  advanceBaseline(key: string, value: string): void {
    if (this.draftValues.has(key)) this.baselines.set(key, value);
  }
  reviewBaseline(key: string, value: string): void {
    this.advanceBaseline(key, value);
    if (this.draftValues.has(key)) this.revisions.set(key, ++this.sequence);
  }
  remove(key: string): void {
    const existed = this.draftValues.delete(key);
    this.revisions.delete(key);
    this.baselines.delete(key);
    if (existed) this.publish();
  }
  move(from: string, to: string): void {
    const value = this.draftValues.get(from);
    if (value === undefined) return;
    this.draftValues.set(to, value);
    this.revisions.set(to, this.revision(from));
    const baseline = this.baselines.get(from);
    if (baseline !== undefined) this.baselines.set(to, baseline);
    this.remove(from);
  }

  clear(): void {
    this.draftValues.clear();
    this.focusKeysByRow.clear();
    this.revisions.clear();
    this.baselines.clear();
    this.publish();
  }
}
