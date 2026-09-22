import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import path from "node:path";

import {
  parseStrictJSON,
  validateSchemaSync,
} from "../../contract/index.mjs";
import { replaceFileAtomically } from "../design-tokens/design-tokens.mjs";

const schemaID = "cartulary.design_presentation.v2";
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
  validateInspectorFieldLayouts(projection.inspector.field_layout_overrides, path.resolve(path.dirname(filePath), "../view-schemas"));
  return {
    inputSha256: createHash("sha256").update(inputBytes).digest("hex"),
    projection,
  };
}

// Authored semantic references, never a Markdown table or generated registry.
export function validateInspectorFieldLayouts(overrides, schemasDirectory) {
  const index = parseStrictJSON(readFileSync(path.join(schemasDirectory, "index.json"), "utf8"));
  const schemas = new Map(index.view_schemas.map((entry) => [entry.view_schema_id, entry]));
  const seen = new Set();
  for (const override of overrides) {
    const key = JSON.stringify([override.view_schema_id, override.field_key]);
    if (seen.has(key)) throw new DesignPresentationValidationError(`Duplicate inspector field layout: ${key}`);
    seen.add(key);
    if (!["property", "narrative"].includes(override.layout))
      throw new DesignPresentationValidationError(`Unknown inspector field layout: ${override.layout}`);
    const entry = schemas.get(override.view_schema_id);
    if (!entry) throw new DesignPresentationValidationError(`Unknown inspector view schema: ${override.view_schema_id}`);
    const source = path.join(schemasDirectory, path.basename(entry.artifact_path));
    const schema = parseStrictJSON(readFileSync(source, "utf8"));
    if (!schema.fields.some((field) => field.field_key === override.field_key))
      throw new DesignPresentationValidationError(`Unknown inspector field: ${key}`);
  }
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
    workbookChrome: {
      queryChipCapacities: document.projection.workbook_chrome.query_chip_capacities,
      savedViewStartupInitiallyOpen: document.projection.workbook_chrome.saved_view_startup_initially_open,
    },
    inspector: {
      referenceViewport: document.projection.inspector.reference_viewport,
      persistentRegionMaxHeightPx: document.projection.inspector.persistent_region_max_height_px,
      fieldActionMinSizePx: document.projection.inspector.field_action_min_size_px,
      propertyStackBelowPx: document.projection.inspector.property_stack_below_px,
      overlayMinViewportWidthPx: document.projection.inspector.overlay_min_viewport_width_px,
      fieldLayoutOverrides: document.projection.inspector.field_layout_overrides.map((entry) => ({
        viewSchemaId: entry.view_schema_id, fieldKey: entry.field_key, layout: entry.layout,
      })),
      dataStates: document.projection.inspector.data_states,
      accessStates: document.projection.inspector.access_states,
      contentStates: document.projection.inspector.content_states,
      headerTitleLines: document.projection.inspector.header_title_lines,
      narrativePreviewLines: document.projection.inspector.narrative_preview_lines,
      actionOutcomes: document.projection.inspector.action_outcomes,
      unavailableCauses: document.projection.inspector.unavailable_causes,
      announcements: document.projection.inspector.announcements,
    },
    gridCellRangeSelection: {
      stationaryTolerancePx: document.projection.grid_cell_range_selection.stationary_tolerance_px,
      edgeBandPx: document.projection.grid_cell_range_selection.edge_band_px,
      maximumScrollPxPerSecond: document.projection.grid_cell_range_selection.maximum_scroll_px_per_second,
      maximumFrameMs: document.projection.grid_cell_range_selection.maximum_frame_ms,
    },
    workbookFrozenColumns: {
      minimumScrollableWidthPx: document.projection.workbook_frozen_columns.minimum_scrollable_width_px,
    },
    workbookColumnSizing: {
      minimumWidthPx: document.projection.workbook_column_sizing.minimum_width_px,
      maximumWidthPx: document.projection.workbook_column_sizing.maximum_width_px,
      keyboardStepPx: document.projection.workbook_column_sizing.keyboard_step_px,
    },
    workbookFind: {
      scopeLabel: document.projection.workbook_find.scope_label,
      zeroMatchesMessage: document.projection.workbook_find.zero_matches_message,
      scopeHelp: document.projection.workbook_find.scope_help,
      live: document.projection.workbook_find.live,
      successfulNavigation: document.projection.workbook_find.successful_navigation,
    },
    presence: document.projection.presence,
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
