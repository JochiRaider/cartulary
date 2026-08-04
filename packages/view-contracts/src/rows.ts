import { hasOwn, isRecord, viewRowInvariant } from "./invariants.js";
import type {
  NormalizedViewRowPatchV1,
  NormalizedViewRowV1,
  ViewContract,
  ViewRowCellV1,
} from "./types.js";

function requireRowObject(
  value: unknown,
  source: string,
  label: string,
): Record<string, unknown> {
  if (!isRecord(value)) {
    viewRowInvariant(source, `${label} must be an object`);
  }
  return value;
}

function requireRowRecordId(value: unknown, source: string): string {
  if (typeof value !== "string" || value.trim() === "") {
    viewRowInvariant(source, "record_id must be a non-empty string");
  }
  return value;
}

function requireRowVersion(value: unknown, source: string): number {
  if (typeof value !== "number" || !Number.isSafeInteger(value) || value < 1) {
    viewRowInvariant(source, "row_version must be a positive safe integer");
  }
  return value;
}

function declaredDataFieldKeys(contract: ViewContract): ReadonlySet<string> {
  return new Set(contract.fields.map((field) => field.fieldKey));
}

function technicalFieldKeys(contract: ViewContract): ReadonlySet<string> {
  return new Set(contract.technicalFields);
}

function sanitizeViewRowCell(
  value: unknown,
  source: string,
  fieldKey: string,
): ViewRowCellV1 {
  const cell = requireRowObject(value, source, `cells.${fieldKey}`);
  if (!hasOwn(cell, "value")) {
    viewRowInvariant(source, `cell ${fieldKey} missing value`);
  }
  return Object.freeze({ value: cell.value });
}

function sanitizeGroupValues({
  allowSparse,
  contract,
  source,
  value,
}: {
  readonly allowSparse: boolean;
  readonly contract: ViewContract;
  readonly source: string;
  readonly value: unknown;
}): Readonly<Record<string, unknown>> | undefined {
  const groupingFields = new Set(contract.groupingFields);
  if (value === undefined) {
    if (!allowSparse && groupingFields.size > 0) {
      viewRowInvariant(source, "group_values is required for grouped schema");
    }
    return undefined;
  }
  if (groupingFields.size === 0) {
    viewRowInvariant(
      source,
      "group_values is not allowed for ungrouped schema",
    );
  }
  const raw = requireRowObject(value, source, "group_values");
  for (const fieldKey of Object.keys(raw)) {
    if (!groupingFields.has(fieldKey)) {
      viewRowInvariant(source, `unknown group_values field ${fieldKey}`);
    }
  }
  if (!allowSparse) {
    for (const fieldKey of groupingFields) {
      if (!hasOwn(raw, fieldKey)) {
        viewRowInvariant(source, `missing group_values field ${fieldKey}`);
      }
    }
  }
  return Object.freeze({ ...raw });
}

function validateOptionalRowViewSchemaId(
  raw: Record<string, unknown>,
  contract: ViewContract,
  source: string,
) {
  const value = raw.view_schema_id;
  if (value === undefined) {
    return;
  }
  if (value !== contract.viewSchemaId) {
    viewRowInvariant(source, `view_schema_id must be ${contract.viewSchemaId}`);
  }
}

function normalizeViewRowCells({
  allowSparse,
  contract,
  rawCells,
  source,
}: {
  readonly allowSparse: boolean;
  readonly contract: ViewContract;
  readonly rawCells: unknown;
  readonly source: string;
}): Readonly<Record<string, ViewRowCellV1>> {
  const cells = requireRowObject(rawCells, source, "cells");
  const dataFields = declaredDataFieldKeys(contract);
  const technicalFields = technicalFieldKeys(contract);
  const normalized: Record<string, ViewRowCellV1> = {};

  for (const [fieldKey, cell] of Object.entries(cells)) {
    if (technicalFields.has(fieldKey)) {
      viewRowInvariant(source, `technical cell ${fieldKey} is not allowed`);
    }
    if (!dataFields.has(fieldKey)) {
      viewRowInvariant(source, `unknown cell ${fieldKey}`);
    }
    normalized[fieldKey] = sanitizeViewRowCell(cell, source, fieldKey);
  }

  if (!allowSparse) {
    for (const field of contract.fields) {
      if (!hasOwn(cells, field.fieldKey)) {
        viewRowInvariant(source, `missing cell ${field.fieldKey}`);
      }
    }
  }

  return Object.freeze(normalized);
}

export function normalizeViewRowV1(
  contract: ViewContract,
  row: unknown,
  source = contract.viewSchemaId,
): NormalizedViewRowV1 {
  const raw = requireRowObject(row, source, "view_row_v1");
  validateOptionalRowViewSchemaId(raw, contract, source);
  return Object.freeze({
    recordId: requireRowRecordId(raw.record_id, source),
    rowVersion: requireRowVersion(raw.row_version, source),
    viewSchemaId: contract.viewSchemaId,
    cells: normalizeViewRowCells({
      allowSparse: false,
      contract,
      rawCells: raw.cells,
      source,
    }),
    groupValues: sanitizeGroupValues({
      allowSparse: false,
      contract,
      source,
      value: raw.group_values,
    }),
  }) as NormalizedViewRowV1;
}

export function normalizeViewRowPatchV1(
  contract: ViewContract,
  row: unknown,
  source = contract.viewSchemaId,
): NormalizedViewRowPatchV1 {
  const raw = requireRowObject(row, source, "view_row_patch_v1");
  validateOptionalRowViewSchemaId(raw, contract, source);
  return Object.freeze({
    recordId: requireRowRecordId(raw.record_id, source),
    rowVersion: requireRowVersion(raw.row_version, source),
    viewSchemaId: contract.viewSchemaId,
    cells: normalizeViewRowCells({
      allowSparse: true,
      contract,
      rawCells: raw.cells,
      source,
    }),
    groupValues: sanitizeGroupValues({
      allowSparse: true,
      contract,
      source,
      value: raw.group_values,
    }),
  }) as NormalizedViewRowPatchV1;
}
