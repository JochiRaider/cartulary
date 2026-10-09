import { readResourceObservations } from "./resource-reader.mjs";
import { createResourceAccumulator } from "./resource-accumulator.mjs";
import { presentResources } from "./resource-presentation.mjs";

export function resourceProjection(runRoot, run) {
  const accumulator = createResourceAccumulator();
  const retained = readResourceObservations(runRoot, run, accumulator.consume);
  if (!retained.index) return retained;
  return presentResources(retained.index, retained.mode, accumulator.finish());
}
