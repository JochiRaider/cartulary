import type { ViewFieldContract } from "@cartulary/view-contracts";
import type { EntityApiRow } from "../timeline/models/workbookTimelineModel";
import { stringifyGridValue } from "../utils/workbookValueFormat";

export type EntityRow = {
  entityType: "host" | "identity";
  recordId: string;
  rowVersion: number;
  label: string;
  secondaryText: string;
  state: string;
  aliasTexts: string[];
  aliases: EntityAlias[];
  linkedEventCount: number;
  rawRow: EntityApiRow;
  identifiers: Array<{
    key: string;
    label: string;
    identifierClass: string;
    value: string;
  }>;
  reusableIdentifiers: ReusableIdentifier[];
};

export type EntityAlias = {
  itemRef: string;
  itemKind: "alias";
  displayText: string;
  aliasText: string;
};

export type MergePlanLine = {
  label: string;
  outcome: string;
};

export type ReusableIdentifier = {
  itemRef: string;
  itemKind: "reusable_identifier" | string;
  identifierClass: string;
  label: string;
  rawValue: string;
  normalizedValue: string;
  displayText: string;
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
  Array<{ key: string; label: string; identifierClass: string }>
> = {
  host: [
    {
      key: "host.aad_device_id",
      label: "AAD Device ID",
      identifierClass: "aad_device_id",
    },
    { key: "host.fqdn", label: "FQDN", identifierClass: "fqdn" },
    {
      key: "host.hostname",
      label: "Hostname",
      identifierClass: "hostname",
    },
  ],
  identity: [
    {
      key: "identity.aad_object_id",
      label: "AAD Object ID",
      identifierClass: "aad_object_id",
    },
    { key: "identity.sid", label: "SID", identifierClass: "sid" },
    { key: "identity.upn", label: "UPN", identifierClass: "upn" },
    { key: "identity.email", label: "Email", identifierClass: "email" },
    {
      key: "identity.sam_account_name",
      label: "SAM Account Name",
      identifierClass: "sam_account_name",
    },
  ],
};

const reusableIdentifierFields: Record<EntityRow["entityType"], string> = {
  host: "host.reusable_identifiers",
  identity: "identity.reusable_identifiers",
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

function readEntityStringCell(row: EntityApiRow, fieldKey: string): string {
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
        identifierClass: field.identifierClass,
        value,
      };
    })
    .filter(
      (
        value,
      ): value is {
        key: string;
        label: string;
        identifierClass: string;
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
      return [] as EntityAlias[];
    }
    return raw.items
      .map((item) => {
        if (!item || typeof item !== "object") {
          return null;
        }
        const object = item as Record<string, unknown>;
        if (
          object.item_kind !== "alias" ||
          typeof object.item_ref !== "string" ||
          !/^entity_alias:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/u.test(
            object.item_ref,
          ) ||
          typeof object.display_text !== "string" ||
          typeof object.alias_text !== "string" ||
          object.display_text !== object.alias_text
        ) {
          return null;
        }
        return {
          itemRef: object.item_ref,
          itemKind: "alias" as const,
          displayText: object.display_text,
          aliasText: object.alias_text,
        };
      })
      .filter((value): value is EntityAlias => value !== null);
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
    aliasTexts: aliasItems.map((item) => item.aliasText),
    aliases: aliasItems,
    linkedEventCount: readNumberCell(row, linkedEventField),
    rawRow: row,
    identifiers,
    reusableIdentifiers: readReusableIdentifiers(row, entityType),
  };
}

function compareValue(value: string) {
  return value.trim().toLowerCase();
}

function readReusableIdentifiers(
  row: EntityApiRow,
  entityType: EntityRow["entityType"],
): ReusableIdentifier[] {
  const raw = row.cells[reusableIdentifierFields[entityType]]?.value;
  if (
    !raw ||
    typeof raw !== "object" ||
    Array.isArray(raw) ||
    !("items" in raw) ||
    !Array.isArray(raw.items)
  ) {
    return [];
  }
  return raw.items
    .map((item) => {
      if (!item || typeof item !== "object" || Array.isArray(item)) {
        return null;
      }
      const object = item as Record<string, unknown>;
      const itemRef = readNonEmptyString(object.item_ref);
      const identifierClass = readNonEmptyString(object.identifier_class);
      const rawValue = readNonEmptyString(object.raw_value);
      if (itemRef === "" || identifierClass === "" || rawValue === "") {
        return null;
      }
      const displayText =
        readNonEmptyString(object.display_text) ||
        readNonEmptyString(object.normalized_value) ||
        rawValue;
      const normalizedValue =
        readNonEmptyString(object.normalized_value) || compareValue(rawValue);
      return {
        itemRef,
        itemKind: readNonEmptyString(object.item_kind) || "reusable_identifier",
        identifierClass,
        label: identifierClassLabel(entityType, identifierClass),
        rawValue,
        normalizedValue,
        displayText,
      };
    })
    .filter((value): value is ReusableIdentifier => value !== null);
}

function readNonEmptyString(value: unknown): string {
  return typeof value === "string" && value.trim() !== "" ? value : "";
}

function identifierClassLabel(
  entityType: EntityRow["entityType"],
  identifierClass: string,
): string {
  return (
    mergeIdentifierFields[entityType].find(
      (field) => field.identifierClass === identifierClass,
    )?.label ?? identifierClass.replaceAll("_", " ")
  );
}

function identifierSignature(identifierClass: string, value: string): string {
  return `${identifierClass}:${compareValue(value)}`;
}

function reusableIdentifierSignature(identifier: ReusableIdentifier): string {
  return identifierSignature(
    identifier.identifierClass,
    identifier.normalizedValue || identifier.rawValue,
  );
}

function entityIdentifierSignatures(row: EntityRow): Set<string> {
  const signatures = new Set<string>();
  for (const identifier of row.identifiers) {
    signatures.add(
      identifierSignature(identifier.identifierClass, identifier.value),
    );
  }
  for (const identifier of row.reusableIdentifiers) {
    signatures.add(reusableIdentifierSignature(identifier));
  }
  return signatures;
}

export function buildMergePlan(
  survivor: EntityRow,
  loser: EntityRow,
): EntityMergePlan {
  const survivorIdentifierSignatures = entityIdentifierSignatures(survivor);
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
    const loserSignature = identifierSignature(
      field.identifierClass,
      loserValue,
    );
    if (
      survivorValue !== "" &&
      loserValue !== "" &&
      survivorIdentifierSignatures.has(loserSignature)
    ) {
      return [{ label: field.label, outcome: `Duplicate no-op ${loserValue}` }];
    }
    if (survivorValue !== "" && loserValue !== "") {
      return [
        {
          label: field.label,
          outcome: `Carry as reusable ${loserValue}`,
        },
      ];
    }
    return [];
  });
  const loserReusableIdentifierLines = loser.reusableIdentifiers.map(
    (identifier) => {
      const outcome = survivorIdentifierSignatures.has(
        reusableIdentifierSignature(identifier),
      )
        ? `Duplicate no-op ${identifier.displayText}`
        : `Carry as reusable ${identifier.displayText}`;
      return {
        label: identifier.label,
        outcome,
      };
    },
  );

  const survivorAliases = new Set(survivor.aliasTexts.map(compareValue));
  const aliasesToCopy = loser.aliasTexts.filter(
    (value) => !survivorAliases.has(compareValue(value)),
  );
  const duplicateAliases = loser.aliasTexts.filter((value) =>
    survivorAliases.has(compareValue(value)),
  );

  return {
    identifierLines: [...identifierLines, ...loserReusableIdentifierLines],
    aliasesToCopy,
    duplicateAliases,
    provenanceOnlySummary:
      "Merge lineage and source provenance are retained server-side; no editable cell value is copied for them.",
    dependencySummary:
      survivor.linkedEventCount > 0 || loser.linkedEventCount > 0
        ? `Linked events visible on surface: survivor=${survivor.linkedEventCount}, loser=${loser.linkedEventCount}.`
        : "Dependency counts are not exposed on this surface.",
  };
}
