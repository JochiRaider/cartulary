import {
  compareExecutionDependencyIDs,
  executionDependencyMetadata as loadExecutionDependencyMetadata,
  serviceBackedGoExecutionDependencies as loadServiceBackedGoExecutionDependencies,
  serviceBackedSupportTargets as loadServiceBackedSupportTargets,
  targetForExecutionDependencyID,
  validExecutionDependencyIDs,
  validSupportTargetIDs,
} from "../generated-artifacts/index.mjs";

export const executionDependencyMetadata = loadExecutionDependencyMetadata();

export const validExecutionDependencies = validExecutionDependencyIDs();

export const validSupportTargets = validSupportTargetIDs();

export const serviceBackedGoExecutionDependencies = loadServiceBackedGoExecutionDependencies();

export const serviceBackedSupportTargets = loadServiceBackedSupportTargets();

export function executionDependencyInfo(id) {
  return executionDependencyMetadata.get(id) ?? null;
}

export function compareExecutionDependencies(left, right) {
  return compareExecutionDependencyIDs(left, right);
}

export function targetForExecutionDependency(id, label = "execution_dependency") {
  return targetForExecutionDependencyID(id, label);
}
