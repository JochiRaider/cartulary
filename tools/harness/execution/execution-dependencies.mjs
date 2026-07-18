import {
  compareExecutionDependencyIDs,
  executionDependencyMetadata as loadExecutionDependencyMetadata,
  targetForExecutionDependencyID,
} from "../generated-artifacts/execution-topology.mjs";

export function executionDependencyInfo(id) {
  return loadExecutionDependencyMetadata().get(id) ?? null;
}

export function compareExecutionDependencies(left, right) {
  return compareExecutionDependencyIDs(left, right);
}

export function targetForExecutionDependency(id, label = "execution_dependency") {
  return targetForExecutionDependencyID(id, label);
}
