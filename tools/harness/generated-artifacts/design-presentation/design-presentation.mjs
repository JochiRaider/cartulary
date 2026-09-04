import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import path from "node:path";

import {
  parseStrictJSON,
  validateSchemaSync,
} from "../../contract/index.mjs";
import { replaceFileAtomically } from "../design-tokens/design-tokens.mjs";

const schemaID = "cartulary.design_presentation.v1";
const generatorID = "cartulary.design_presentation_generation.v1";
const expectedFamilies = Object.freeze([
  "local_validation",
  "same_field_conflict",
  "client_txn_conflict",
  "queue_overflow",
  "stale_refresh",
  "initial_load_failure",
  "authentication_required",
  "permission_or_incident_access_loss",
  "extension_unavailable",
  "evidence_preview_blocked",
  "unknown_future_error",
]);
const expectedGridDataStates = Object.freeze([
  "ready",
  "initial_loading",
  "refreshing",
  "empty",
  "filtered_empty",
  "stale_error",
  "unavailable",
  "permission_denied",
]);
const expectedGridInteractionModes = Object.freeze(["editable", "read_only"]);

export class DesignPresentationValidationError extends Error {
  constructor(message) {
    super(message);
    this.name = "DesignPresentationValidationError";
  }
}

function assertMachineProjectionPath(filePath) {
  const resolved = path.resolve(filePath);
  if (resolved.split(path.sep).includes("docs") || path.extname(resolved) !== ".json") {
    throw new DesignPresentationValidationError(
      `${resolved}: design presentation input must be JSON outside docs/`,
    );
  }
}

export function loadDesignPresentationDocument(filePath) {
  assertMachineProjectionPath(filePath);
  const inputBytes = readFileSync(filePath);
  let projection;
  try {
    projection = parseStrictJSON(
      new TextDecoder("utf-8", { fatal: true }).decode(inputBytes),
      filePath,
    );
    validateSchemaSync(schemaID, projection);
  } catch (error) {
    throw new DesignPresentationValidationError(
      `${filePath}: ${error instanceof Error ? error.message : String(error)}`,
    );
  }
  const families = projection.error_presentations.map((entry) => entry.family);
  if (
    new Set(families).size !== expectedFamilies.length ||
    expectedFamilies.some((family) => !families.includes(family))
  ) {
    throw new DesignPresentationValidationError(
      `${filePath}: error_presentations must contain every current family exactly once`,
    );
  }
  assertExactRows(
    filePath,
    "grid_data_state_presentations",
    projection.grid_data_state_presentations.map((entry) => entry.state),
    expectedGridDataStates,
  );
  assertExactRows(
    filePath,
    "grid_interaction_mode_presentations",
    projection.grid_interaction_mode_presentations.map((entry) => entry.mode),
    expectedGridInteractionModes,
  );
  return {
    inputSha256: createHash("sha256").update(inputBytes).digest("hex"),
    projection,
  };
}

function assertExactRows(filePath, field, actual, expected) {
  if (
    new Set(actual).size !== expected.length ||
    expected.some((identity) => !actual.includes(identity))
  ) {
    throw new DesignPresentationValidationError(
      `${filePath}: ${field} must contain every current identity exactly once`,
    );
  }
}

export function renderDesignPresentationTypeScript(document) {
  const source = {
    errorPresentations: document.projection.error_presentations.map((entry) => ({
      actions: entry.actions,
      family: entry.family,
      focusEffect: entry.focus_effect,
      live: entry.live,
      locus: entry.locus,
      retention: entry.retention,
    })),
    gridDataStatePresentations:
      document.projection.grid_data_state_presentations.map((entry) => ({
        actionRule: entry.action_rule,
        blocking: entry.blocking,
        draftRetention: entry.draft_retention,
        focusEffect: entry.focus_effect,
        live: entry.live,
        message: entry.message,
        messageStrategy: entry.message_strategy,
        placement: entry.placement,
        posture: entry.posture,
        role: entry.role,
        rowRetention: entry.row_retention,
        state: entry.state,
      })),
    gridInteractionModePresentations:
      document.projection.grid_interaction_mode_presentations.map((entry) => ({
        focusEffect: entry.focus_effect,
        live: entry.live,
        messageStrategy: entry.message_strategy,
        mode: entry.mode,
        posture: entry.posture,
        role: entry.role,
        visible: entry.visible,
      })),
    gridStateComposition: {
      coDisplayInteractionMode:
        document.projection.grid_state_composition.co_display_interaction_mode,
      liveRegionRule:
        document.projection.grid_state_composition.live_region_rule,
      primary: document.projection.grid_state_composition.primary,
      suppressInteractionForDataStates:
        document.projection.grid_state_composition
          .suppress_interaction_for_data_states,
    },
    initialLoading: {
      announceOncePerGeneration:
        document.projection.initial_loading.announce_once_per_generation,
      delayMs: document.projection.initial_loading.delay_ms,
      live: document.projection.initial_loading.live,
      message: document.projection.initial_loading.message,
      retryOnDelay: document.projection.initial_loading.retry_on_delay,
    },
    statusSecondaryPriority: document.projection.status_secondary_priority,
    transientConfirmation: {
      live: document.projection.transient_confirmation.live,
      pauseConditions: document.projection.transient_confirmation.pause_conditions,
      resumeResetsElapsed:
        document.projection.transient_confirmation.resume_resets_elapsed,
      stillValidActionPreventsDismissal:
        document.projection.transient_confirmation
          .still_valid_action_prevents_dismissal,
      visibleUnpausedMs:
        document.projection.transient_confirmation.visible_unpaused_ms,
    },
  };
  return [
    `// Code generated by ${generatorID}; DO NOT EDIT.`,
    `// Input SHA-256: ${document.inputSha256}`,
    "",
    `export const cartularyDesignPresentation = ${JSON.stringify(source, null, 2)} as const;`,
    "",
    "export type CartularyErrorPresentation = (typeof cartularyDesignPresentation.errorPresentations)[number];",
    "export type CartularyErrorFamily = CartularyErrorPresentation[\"family\"];",
    "export type CartularyGridDataStatePresentation = (typeof cartularyDesignPresentation.gridDataStatePresentations)[number];",
    "export type CartularyGridDataState = CartularyGridDataStatePresentation[\"state\"];",
    "export type CartularyGridInteractionModePresentation = (typeof cartularyDesignPresentation.gridInteractionModePresentations)[number];",
    "export type CartularyGridInteractionMode = CartularyGridInteractionModePresentation[\"mode\"];",
    "export type CartularyStatusSecondaryKind = (typeof cartularyDesignPresentation.statusSecondaryPriority)[number];",
    "",
  ].join("\n");
}

export { replaceFileAtomically };
