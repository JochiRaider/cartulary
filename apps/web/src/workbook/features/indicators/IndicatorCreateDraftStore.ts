import type { IndicatorObservation } from "../../mutations/workbookMutationCommandPorts";
import {
  changeIndicatorCreateValue,
  type IndicatorCreateDraft,
  indicatorCreateSeed,
} from "./indicatorCreateModel";
import { freezeObservation } from "./observationModel";

export class IndicatorCreateDraftStore {
  private readonly drafts = new Map<string, IndicatorCreateDraft>();
  constructor(private readonly changed: () => void = () => {}) {}
  get(id: string) {
    return this.drafts.get(id);
  }
  ensure(observation: IndicatorObservation) {
    const existing = this.get(observation.observation_id);
    if (existing) return existing;
    const draft = freezeObservation({
      observationId: observation.observation_id,
      revision: 0,
      values: indicatorCreateSeed(observation),
    });
    this.drafts.set(draft.observationId, draft);
    return draft;
  }
  update(id: string, key: string, value: string) {
    const draft = this.get(id);
    if (!draft) return;
    this.drafts.set(
      id,
      freezeObservation({
        ...draft,
        revision: draft.revision + 1,
        values: changeIndicatorCreateValue(draft.values, key, value),
      }),
    );
    this.changed();
  }
  clear() {
    this.drafts.clear();
    this.changed();
  }
}
