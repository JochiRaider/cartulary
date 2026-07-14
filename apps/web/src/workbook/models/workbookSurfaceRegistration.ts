import {
  listViewContracts,
  type ViewContract,
} from "@cartulary/view-contracts";
import { artifactSurfacePolicies } from "../policies/artifactSurfacePolicies";
import { assessmentSurfacePolicies } from "../policies/assessmentSurfacePolicies";
import { captureTimelineSurfacePolicies } from "../policies/captureTimelineSurfacePolicies";
import { coordinationSurfacePolicies } from "../policies/coordinationSurfacePolicies";
import { entitiesObservationsSurfacePolicies } from "../policies/entitiesObservationsSurfacePolicies";
import { evidenceSurfacePolicies } from "../policies/evidenceSurfacePolicies";
import type {
  WorkbookSurfacePolicyDefinition,
  WorkbookSurfaceRegistration,
} from "../policies/workbookSurfacePolicy";

export type {
  ReferenceRequirement,
  WorkbookSurfaceOwnerId,
  WorkbookSurfacePolicy,
  WorkbookSurfacePolicyDefinition,
  WorkbookSurfaceRegistration,
  WorkbookSurfaceRenderer,
} from "../policies/workbookSurfacePolicy";

const definitions: readonly WorkbookSurfacePolicyDefinition[] = [
  ...captureTimelineSurfacePolicies,
  ...entitiesObservationsSurfacePolicies,
  ...assessmentSurfacePolicies,
  ...evidenceSurfacePolicies,
  ...artifactSurfacePolicies,
  ...coordinationSurfacePolicies,
];

export function buildWorkbookSurfaceRegistrations(
  contracts: readonly ViewContract[] = listViewContracts(),
  registrationDefinitions: readonly WorkbookSurfacePolicyDefinition[] = definitions,
): readonly WorkbookSurfaceRegistration[] {
  const contractsById = new Map(
    contracts.map((contract) => [contract.viewSchemaId, contract]),
  );
  const seen = new Set<string>();
  const registrations = registrationDefinitions.map((definition) => {
    if (seen.has(definition.viewSchemaId)) {
      throw new Error(
        `Duplicate workbook surface registration: ${definition.viewSchemaId}`,
      );
    }
    seen.add(definition.viewSchemaId);
    const contract = contractsById.get(definition.viewSchemaId);
    if (!contract) {
      throw new Error(
        `Workbook surface registration has no contract: ${definition.viewSchemaId}`,
      );
    }
    return Object.freeze({ ...definition, contract });
  });
  const missing = [...contractsById.keys()].filter((id) => !seen.has(id));
  if (missing.length > 0) {
    throw new Error(
      `Missing workbook surface registration: ${missing.sort().join(", ")}`,
    );
  }
  return Object.freeze(registrations);
}

const registrations = buildWorkbookSurfaceRegistrations();
const registrationsById = new Map(
  registrations.map((registration) => [
    registration.viewSchemaId,
    registration,
  ]),
);

export function listWorkbookSurfaceRegistrations() {
  return registrations;
}

export function requireWorkbookSurfaceRegistration(
  viewSchemaId: string,
): WorkbookSurfaceRegistration {
  const registration = registrationsById.get(viewSchemaId);
  if (!registration) {
    throw new Error(`Unknown workbook surface registration: ${viewSchemaId}`);
  }
  return registration;
}
