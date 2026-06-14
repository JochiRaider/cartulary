import type { ViewFieldContract } from "@cartulary/view-contracts";
import type { EntityApiRow } from "./workbookTimelineModel";
import { stringifyGridValue } from "./workbookValueFormat";

export type EntityRow = {
  entityType: "host" | "identity";
  recordId: string;
  rowVersion: number;
  label: string;
  secondaryText: string;
  state: string;
  aliasTexts: string[];
  linkedEventCount: number;
  rawRow: EntityApiRow;
  identifiers: Array<{
    key: string;
    label: string;
    value: string;
  }>;
};

export type MergePlanLine = {
  label: string;
  outcome: string;
};

export type EntityMergePlan = {
  identifierLines: MergePlanLine[];
  aliasesToCopy: string[];
  duplicateAliases: string[];
  provenanceOnlySummary: string;
  dependencySummary: string;
};

const mergeIdentifierFields: Record<
  EntityRow["entityType"],
  Array<{ key: string; label: string }>
> = {
  host: [
    { key: "host.aad_device_id", label: "AAD Device ID" },
    { key: "host.fqdn", label: "FQDN" },
    { key: "host.hostname", label: "Hostname" },
  ],
  identity: [
    { key: "identity.aad_object_id", label: "AAD Object ID" },
    { key: "identity.sid", label: "SID" },
    { key: "identity.upn", label: "UPN" },
    { key: "identity.email", label: "Email" },
    { key: "identity.sam_account_name", label: "SAM Account Name" },
  ],
};

export function entityContractColumnWidth(field: ViewFieldContract): number {
  switch (field.fieldKey) {
    case "host.display_name":
    case "identity.display_name":
      return 240;
    case "host.hostname":
    case "identity.upn":
    case "identity.email":
    case "identity.sam_account_name":
      return 260;
    case "host.aliases":
    case "identity.aliases":
      return 320;
    case "host.linked_event_count":
    case "identity.linked_event_count":
    case "host.evidence_count":
    case "identity.evidence_count":
      return 140;
    case "host.edited_at":
    case "identity.edited_at":
      return 180;
    case "host.host_state":
    case "identity.identity_state":
    case "identity.mfa_state":
    case "identity.reset_status":
      return 150;
    default:
      return 200;
  }
}

export function readEntityStringCell(
  row: EntityApiRow,
  fieldKey: string,
): string {
  const raw = row.cells[fieldKey]?.value;
  return typeof raw === "string" ? raw : "";
}

function readNumberCell(row: EntityApiRow, fieldKey: string): number {
  const raw = row.cells[fieldKey]?.value;
  return typeof raw === "number" ? raw : 0;
}

function readEntityCellValue(row: EntityApiRow | null, fieldKey: string) {
  return row?.cells[fieldKey]?.value ?? null;
}

export function entityGroupLabel(row: EntityRow, fieldKey: string): string {
  const value = stringifyGridValue(
    readEntityCellValue(row.rawRow, fieldKey),
  ).trim();
  return value === "" ? "Unassigned" : value;
}

export function entityRowFromApi(
  row: EntityApiRow,
  entityType: EntityRow["entityType"],
): EntityRow {
  const labelField =
    entityType === "host" ? "host.display_name" : "identity.display_name";
  const secondaryCandidates =
    entityType === "host"
      ? ["host.hostname", "host.fqdn"]
      : ["identity.email", "identity.upn", "identity.sam_account_name"];
  const stateField =
    entityType === "host" ? "host.host_state" : "identity.identity_state";
  const aliasesField =
    entityType === "host" ? "host.aliases" : "identity.aliases";
  const linkedEventField =
    entityType === "host"
      ? "host.linked_event_count"
      : "identity.linked_event_count";
  const identifiers = mergeIdentifierFields[entityType]
    .map((field) => {
      const value = readEntityStringCell(row, field.key);
      if (value === "") {
        return null;
      }
      return {
        key: field.key,
        label: field.label,
        value,
      };
    })
    .filter(
      (
        value,
      ): value is {
        key: string;
        label: string;
        value: string;
      } => value !== null,
    );
  const aliasItems = (() => {
    const raw = row.cells[aliasesField]?.value;
    if (
      !raw ||
      typeof raw !== "object" ||
      Array.isArray(raw) ||
      !("items" in raw) ||
      !Array.isArray(raw.items)
    ) {
      return [] as string[];
    }
    return raw.items
      .map((item) => {
        if (!item || typeof item !== "object") {
          return null;
        }
        const object = item as Record<string, unknown>;
        if (typeof object.raw_text === "string") {
          return object.raw_text;
        }
        if (typeof object.alias_text === "string") {
          return object.alias_text;
        }
        return typeof object.display_text === "string"
          ? object.display_text
          : null;
      })
      .filter((value): value is string => value !== null);
  })();
  const secondaryText =
    secondaryCandidates
      .map((field) => readEntityStringCell(row, field))
      .find((value) => value !== "") ?? "";
  const label =
    readEntityStringCell(row, labelField) || secondaryText || row.record_id;

  return {
    entityType,
    recordId: row.record_id,
    rowVersion: row.row_version,
    label,
    secondaryText,
    state: readEntityStringCell(row, stateField),
    aliasTexts: aliasItems,
    linkedEventCount: readNumberCell(row, linkedEventField),
    rawRow: row,
    identifiers,
  };
}

function compareValue(value: string) {
  return value.trim().toLowerCase();
}

export function buildMergePlan(
  survivor: EntityRow,
  loser: EntityRow,
): EntityMergePlan {
  const identifierLines: MergePlanLine[] = mergeIdentifierFields[
    survivor.entityType
  ].flatMap((field) => {
    const survivorValue =
      survivor.identifiers.find((identifier) => identifier.key === field.key)
        ?.value ?? "";
    const loserValue =
      loser.identifiers.find((identifier) => identifier.key === field.key)
        ?.value ?? "";
    if (survivorValue === "" && loserValue === "") {
      return [];
    }
    if (survivorValue === "" && loserValue !== "") {
      return [{ label: field.label, outcome: `Promote ${loserValue}` }];
    }
    if (
      survivorValue !== "" &&
      loserValue !== "" &&
      compareValue(survivorValue) === compareValue(loserValue)
    ) {
      return [{ label: field.label, outcome: `Duplicate no-op ${loserValue}` }];
    }
    if (survivorValue !== "" && loserValue !== "") {
      return [
        {
          label: field.label,
          outcome: `Conflict survivor=${survivorValue} loser=${loserValue}`,
        },
      ];
    }
    return [{ label: field.label, outcome: `Survivor keeps ${survivorValue}` }];
  });

  const survivorAliases = new Set(survivor.aliasTexts.map(compareValue));
  const aliasesToCopy = loser.aliasTexts.filter(
    (value) => !survivorAliases.has(compareValue(value)),
  );
  const duplicateAliases = loser.aliasTexts.filter((value) =>
    survivorAliases.has(compareValue(value)),
  );

  return {
    identifierLines,
    aliasesToCopy,
    duplicateAliases,
    provenanceOnlySummary: "Not exposed on this surface.",
    dependencySummary:
      survivor.linkedEventCount > 0 || loser.linkedEventCount > 0
        ? `Linked events visible on surface: survivor=${survivor.linkedEventCount}, loser=${loser.linkedEventCount}.`
        : "Dependency counts are not exposed on this surface.",
  };
}
