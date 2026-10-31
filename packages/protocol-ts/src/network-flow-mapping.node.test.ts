import { describe, expect, expectTypeOf, it } from "vitest";
import type {
  ApprovedMapping,
  EmptyValuePolicy,
  MappingCandidate,
} from "./entrypoints/network-flow.js";
import {
  validateCartularyNetworkFlowApprovedMappingV1 as validateApproved,
  validateCartularyNetworkFlowMappingCandidateV1 as validateCandidate,
} from "./generated/network-flow-validators.js";

describe("Network Flow mapping policy contracts", () => {
  it("distinguishes candidate defaults from concrete approved policies", () => {
    type CandidateSource = Extract<
      MappingCandidate["field_mappings"][number],
      { mapping_kind: "source_column" }
    >;
    type ApprovedSource = Extract<
      ApprovedMapping["field_mappings"][number],
      { mapping_kind: "source_column" }
    >;
    expectTypeOf<CandidateSource["empty_value_policy"]>().toEqualTypeOf<
      EmptyValuePolicy | "profile_default" | undefined
    >();
    expectTypeOf<
      ApprovedSource["empty_value_policy"]
    >().toEqualTypeOf<EmptyValuePolicy>();

    const field = {
      mapping_kind: "source_column",
      field_key: "network_flow.src_port",
      source_column_ordinal: 1,
      transform_id: "port_number_v1",
      combinability: "single_source_only",
    } as const;
    const candidate = {
      target_kind: "network_flow_table",
      target_table_schema_id: "cartulary.network_flow_table.v1",
      source_profile_id: "cisco_sna_flow_search_csv_v1",
      parser_profile_id: "rfc4180_headered_csv_v1",
      unknown_column_policy: "preserve_unmapped_raw",
      timestamp_profile: {
        schema_id: "cartulary.network_flow.timestamp_profile.v1",
        mode: "cisco_sna_datetime_v1",
        precision: "milliseconds",
      },
      field_mappings: [field],
    };
    const approved = {
      ...candidate,
      source_columns: [
        {
          source_column_ordinal: 1,
          raw_header_text: "port",
          normalized_header_for_suggestion: "port",
          raw_header_sha256: "a".repeat(64),
          sample_values: [],
          detected_empty_count: 0,
        },
      ],
    };
    const explicitFields = [
      { ...field, empty_value_policy: "empty_string_is_null" },
    ];
    expect(
      validateCandidate({ ...candidate, field_mappings: explicitFields }),
    ).toBe(true);
    expect(
      validateApproved({ ...approved, field_mappings: explicitFields }),
    ).toBe(true);
    expect(validateCandidate(candidate)).toBe(true);
    expect(validateApproved(approved)).toBe(false);
    for (const policy of [
      "profile_default",
      "empty_string_is_null",
      "empty_string_is_invalid",
      "empty_string_preserved",
    ]) {
      const fields = [{ ...field, empty_value_policy: policy }];
      expect(
        validateCandidate({ ...candidate, field_mappings: fields }),
        policy,
      ).toBe(true);
      expect(
        validateApproved({ ...approved, field_mappings: fields }),
        policy,
      ).toBe(policy !== "profile_default");
    }
    for (const policy of [
      null,
      "",
      " ",
      " profile_default ",
      "unknown",
      1,
      false,
      [],
      {},
    ]) {
      const fields = [{ ...field, empty_value_policy: policy }];
      expect(validateCandidate({ ...candidate, field_mappings: fields })).toBe(
        false,
      );
      expect(validateApproved({ ...approved, field_mappings: fields })).toBe(
        false,
      );
    }
    // The policy exception must not make other required candidate members optional.
    for (const member of ["transform_id", "combinability"] as const) {
      const fields: Record<string, unknown>[] = [
        { ...field, empty_value_policy: "empty_string_is_null" },
      ];
      delete fields[0]![member];
      expect(validateCandidate({ ...candidate, field_mappings: fields })).toBe(
        false,
      );
    }
    const variants = [
      {
        mapping_kind: "ignored_source_column",
        source_column_ordinal: 1,
        ignore_reason: "user_ignored",
      },
      {
        mapping_kind: "system_derivation",
        field_key: "network_flow.observation_source_ref",
        derivation_id: "network_flow.observation_source_ref.v1",
        combinability: "single_source_only",
      },
    ];
    for (const variant of variants) {
      expect(
        validateCandidate({ ...candidate, field_mappings: [variant] }),
      ).toBe(true);
      expect(validateApproved({ ...approved, field_mappings: [variant] })).toBe(
        true,
      );
      for (const extra of [
        ...[
          "profile_default",
          "empty_string_is_null",
          "empty_string_is_invalid",
          "empty_string_preserved",
          null,
          "",
          " ",
          "unknown",
          1,
          false,
          [],
          {},
        ].map((empty_value_policy) => ({ empty_value_policy })),
        { unknown_member: true },
      ]) {
        const fields = [{ ...variant, ...extra }];
        expect(
          validateCandidate({ ...candidate, field_mappings: fields }),
        ).toBe(false);
        expect(validateApproved({ ...approved, field_mappings: fields })).toBe(
          false,
        );
      }
    }
    const concrete = {
      ...field,
      empty_value_policy: "empty_string_is_null",
      unknown_member: true,
    };
    expect(
      validateCandidate({ ...candidate, field_mappings: [concrete] }),
    ).toBe(false);
    expect(validateApproved({ ...approved, field_mappings: [concrete] })).toBe(
      false,
    );

    // Object validators enforce exact decoded names. Duplicate decoded names and
    // malformed Unicode are covered before typed decoding at the raw boundary.
    const validCandidate = {
      ...candidate,
      display_name_override: "flows",
      field_mappings: [...explicitFields, ...variants],
    };
    const validApproved = {
      ...approved,
      ...validCandidate,
      source_columns: [
        {
          ...approved.source_columns[0],
          sample_values: [
            { safe_sample: null, raw_value_sha256: "b".repeat(64) },
          ],
        },
      ],
    };
    const cloneJSON = (value: unknown): unknown =>
      JSON.parse(JSON.stringify(value));
    const objectAt = (value: unknown, path: (string | number)[]) => {
      let object = value as Record<string | number, unknown>;
      for (const key of path) {
        object = object[key] as Record<string | number, unknown>;
      }
      return object;
    };
    for (const [baseline, validate, isCandidate] of [
      [validCandidate, validateCandidate, true],
      [validApproved, validateApproved, false],
    ] as const) {
      expect(validate(baseline)).toBe(true);
      expect(
        validate(
          JSON.parse(
            JSON.stringify(baseline).replace(
              '"target_kind"',
              '"\\u0074arget_kind"',
            ),
          ),
        ),
      ).toBe(true);
      const paths: (string | number)[][] = [
        [],
        ["timestamp_profile"],
        ["field_mappings", 0],
        ["field_mappings", 1],
        ["field_mappings", 2],
      ];
      if (!isCandidate) {
        paths.push(
          ["source_columns", 0],
          ["source_columns", 0, "sample_values", 0],
        );
      }
      for (const path of paths) {
        for (const key of Object.keys(objectAt(baseline, path))) {
          const aliases = [key.toUpperCase()];
          if (key.includes("s")) aliases.push(key.replace("s", "ſ"));
          if (key.includes("k")) aliases.push(key.replace("k", "K"));
          for (const alias of aliases) {
            for (const form of ["alone", "first", "last"]) {
              const value = cloneJSON(baseline);
              const object = objectAt(value, path);
              const original = object[key];
              if (form === "alone" || form === "first") delete object[key];
              object[alias] = form === "alone" ? original : null;
              if (form === "first") object[key] = original;
              expect(validate(value), `${path}/${key}/${alias}/${form}`).toBe(
                false,
              );
            }
          }
          const optional =
            (path.length === 0 && key === "display_name_override") ||
            (isCandidate &&
              ((path.length === 0 &&
                ["parser_profile_id", "unknown_column_policy"].includes(key)) ||
                (path.length === 1 && key === "precision") ||
                (path.length === 2 && key === "empty_value_policy")));
          if (!optional) {
            const value = cloneJSON(baseline);
            delete objectAt(value, path)[key];
            expect(validate(value), `${path}/${key}/missing`).toBe(false);
          }
          for (const invalid of [null, {}]) {
            if (
              invalid === null &&
              ["safe_sample", "raw_value_sha256"].includes(key)
            ) {
              continue;
            }
            const value = cloneJSON(baseline);
            objectAt(value, path)[key] = invalid;
            expect(validate(value), `${path}/${key}/type`).toBe(false);
          }
        }
      }
    }
  });
});
