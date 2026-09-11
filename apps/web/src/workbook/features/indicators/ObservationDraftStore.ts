import { freezeObservation, type ObservationDraft } from "./observationModel";

export class ObservationDraftStore {
  private readonly entries = new Map<string, ObservationDraft>();
  constructor(private readonly changed: () => void = () => {}) {}
  get(key: string) {
    return this.entries.get(key);
  }
  ensure(key: string): ObservationDraft {
    const prior = this.get(key);
    if (prior) return prior;
    const draft = freezeObservation({
      key,
      revision: 0,
      source: null,
      selection: null,
      parsedType: "" as const,
      target: null,
    });
    this.entries.set(key, draft);
    return draft;
  }
  update(
    key: string,
    change: Partial<Omit<ObservationDraft, "key" | "revision">>,
  ) {
    const prior = this.ensure(key);
    const next = freezeObservation({
      ...prior,
      ...change,
      revision: prior.revision + 1,
    });
    this.entries.set(key, next);
    this.changed();
    return next;
  }
  clear() {
    this.entries.clear();
    this.changed();
  }
}
