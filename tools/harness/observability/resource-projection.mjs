import { readExecutionIndex } from "./execution-reader.mjs";
import { readResourceObservations } from "./resource-reader.mjs";
import { createResourceAccumulator } from "./resource-accumulator.mjs";
import { presentResources } from "./resource-presentation.mjs";

export function resourceProjection(runRoot, run, execution = readExecutionIndex(runRoot, run)) {
  const accumulator = createResourceAccumulator();
  const retained = readResourceObservations(runRoot, run, accumulator.consume, execution);
  if (!retained.index) return retained;
  return presentResources(retained.index, retained.mode, accumulator.finish());
}
