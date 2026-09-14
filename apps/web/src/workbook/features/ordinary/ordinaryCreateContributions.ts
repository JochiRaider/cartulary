import type { EntityRecordWriteBoundary } from "../../mutations/entityRecordWriteBoundary";
import { artifactOrdinaryCreate } from "../artifacts/artifactOrdinaryCreate";
import { coordinationOrdinaryCreate } from "../coordination/coordinationOrdinaryCreate";
import {
  createEntityOrdinaryCreate,
  entityOrdinaryCreate,
} from "../entities/entityOrdinaryCreate";
import { evidenceOrdinaryCreate } from "../evidence/evidenceOrdinaryCreate";
import { indicatorOrdinaryCreate } from "../indicators/indicatorOrdinaryCreate";
import { partyOrdinaryCreate } from "../parties/partyOrdinaryCreate";

/** Closed composition of adopted ordinary entries, with no schema lifecycle branches. */
export const ordinaryCreateContributions = [
  entityOrdinaryCreate,
  coordinationOrdinaryCreate,
  partyOrdinaryCreate,
  artifactOrdinaryCreate,
  evidenceOrdinaryCreate,
  indicatorOrdinaryCreate,
];

export function createOrdinaryCreateContributions(
  entityWrites: EntityRecordWriteBoundary,
) {
  return [
    createEntityOrdinaryCreate(entityWrites),
    ...ordinaryCreateContributions.filter(
      (item) => item !== entityOrdinaryCreate,
    ),
  ];
}
