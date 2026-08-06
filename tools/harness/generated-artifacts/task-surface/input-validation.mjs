const validInputBindings = new Set(["make_variable"]);
const validInputSources = new Set([
  "make_command_line",
  "environment",
  "makefile_default",
  "internal_default",
  "manifest",
]);
const validInputTypes = new Set([
  "enum",
  "exact_1_bool",
  "family_id",
  "owner_id",
  "path",
  "positive_decimal",
  "positive_integer",
  "result_selector",
  "row_ids",
  "run_id",
  "semantic_text",
  "task_surface_report_args",
  "target_name",
  "url",
]);
const validEmptyStringRules = new Set(["invalid", "omitted", "false"]);
const validInputNormalizations = new Set([
  "none",
  "trim",
  "trim_lowercase",
  "path_token",
]);
const validInputInvalidReasons = new Set(["usage_error", "configuration_error"]);
const validInputSummaryEmission = new Set([
  "none",
  "value",
  "redacted_value",
  "source_and_value",
]);
const validInputChildForwarding = new Set([
  "none",
  "argv",
  "runtime_env",
  "argv_and_runtime_env",
]);
const makeVariablePattern = /^[A-Z][A-Z0-9_]*$/;
const validGlobalInputTypes = new Set([
  "machine_cache_root",
  "machine_cache_path",
]);
const expectedGlobalInputs = Object.freeze([
  ["CARTULARY_MACHINE_CACHE_DIR", "machine_cache_root", "xdg_cache_home_then_home"],
  ["GO_CACHE_DIR", "machine_cache_path", "CARTULARY_MACHINE_CACHE_DIR/go/build"],
  ["GO_MOD_CACHE_DIR", "machine_cache_path", "CARTULARY_MACHINE_CACHE_DIR/go/mod"],
  ["GO_TMP_DIR", "machine_cache_path", "CARTULARY_MACHINE_CACHE_DIR/go/tmp"],
]);

export function validateGlobalInputContract(errors, manifest) {
  const inputs = manifest.global_inputs;
  if (!Array.isArray(inputs)) {
    errors.push("global_inputs must be an array");
    return;
  }
  const actualNames = inputs.map((input) => input?.name);
  const expectedNames = expectedGlobalInputs.map(([name]) => name);
  if (JSON.stringify(actualNames) !== JSON.stringify(expectedNames)) {
    errors.push(`global_inputs must be the exact ordered set ${expectedNames.join(", ")}`);
  }
  const allowedKeys = new Set([
    "name",
    "type",
    "allowed_sources",
    "default",
    "empty_string",
    "normalization",
    "invalid_reason",
    "summary_emission",
  ]);
  for (const [index, input] of inputs.entries()) {
    const label = `global_inputs[${index + 1}]`;
    if (!input || typeof input !== "object" || Array.isArray(input)) {
      errors.push(`${label} must be an object`);
      continue;
    }
    for (const key of Object.keys(input)) {
      if (!allowedKeys.has(key)) errors.push(`${label} has unknown key ${key}`);
    }
    const expected = expectedGlobalInputs[index];
    if (!validGlobalInputTypes.has(input.type)) {
      errors.push(`${label}.type has invalid value ${JSON.stringify(input.type)}`);
    }
    if (expected && (input.type !== expected[1] || input.default !== expected[2])) {
      errors.push(`${label} must declare type=${expected[1]} default=${expected[2]}`);
    }
    const sources = input.allowed_sources;
    if (
      !Array.isArray(sources) ||
      JSON.stringify(sources) !==
        JSON.stringify(["make_command_line", "environment", "makefile_default"])
    ) {
      errors.push(`${label}.allowed_sources must declare command line, environment, and Make default precedence`);
    }
    if (input.empty_string !== "invalid") errors.push(`${label}.empty_string must be invalid`);
    if (input.normalization !== "absolute_external_path") {
      errors.push(`${label}.normalization must be absolute_external_path`);
    }
    if (input.invalid_reason !== "configuration_error") {
      errors.push(`${label}.invalid_reason must be configuration_error`);
    }
    if (input.summary_emission !== "source_and_value") {
      errors.push(`${label}.summary_emission must be source_and_value`);
    }
  }
}

export function validatePublicInputContract(errors, entry) {
  const contract = entry.input_contract;
  if (!contract || typeof contract !== "object" || Array.isArray(contract)) {
    errors.push(`${entry.name}.input_contract must be declared for public targets`);
    return;
  }
  const allowedKeys = new Set([
    "undeclared_make_command_line",
    "undeclared_inherited_env",
    "inputs",
  ]);
  for (const key of Object.keys(contract)) {
    if (!allowedKeys.has(key)) {
      errors.push(`${entry.name}.input_contract has unknown key ${key}`);
    }
  }
  if (contract.undeclared_make_command_line !== "usage_error") {
    errors.push(`${entry.name}.input_contract.undeclared_make_command_line must be usage_error`);
  }
  if (contract.undeclared_inherited_env !== "ignore") {
    errors.push(`${entry.name}.input_contract.undeclared_inherited_env must be ignore`);
  }
  if (!Array.isArray(contract.inputs)) {
    errors.push(`${entry.name}.input_contract.inputs must be an array`);
    return;
  }
  const seen = new Set();
  for (const [index, input] of contract.inputs.entries()) {
    const label = `${entry.name}.input_contract.inputs[${index + 1}]`;
    if (!input || typeof input !== "object" || Array.isArray(input)) {
      errors.push(`${label} must be an object`);
      continue;
    }
    const allowedInputKeys = new Set([
      "name",
      "binding",
      "allowed_sources",
      "required",
      "type",
      "values",
      "min",
      "max",
      "default",
      "empty_string",
      "normalization",
      "invalid_reason",
      "summary_emission",
      "child_forwarding",
    ]);
    for (const key of Object.keys(input)) {
      if (!allowedInputKeys.has(key)) {
        errors.push(`${label} has unknown key ${key}`);
      }
    }
    if (typeof input.name !== "string" || !makeVariablePattern.test(input.name)) {
      errors.push(`${label}.name must be a safe Make variable name`);
    } else if (seen.has(input.name)) {
      errors.push(`${entry.name}.input_contract.inputs contains duplicate ${input.name}`);
    } else {
      seen.add(input.name);
    }
    if (!validInputBindings.has(input.binding)) {
      errors.push(`${label}.binding must be make_variable`);
    }
    if (!Array.isArray(input.allowed_sources) || input.allowed_sources.length === 0) {
      errors.push(`${label}.allowed_sources must be a non-empty array`);
    } else {
      for (const source of input.allowed_sources) {
        if (!validInputSources.has(source)) {
          errors.push(`${label}.allowed_sources contains invalid source ${JSON.stringify(source)}`);
        }
      }
    }
    if (typeof input.required !== "boolean") {
      errors.push(`${label}.required must be a boolean`);
    }
    if (!validInputTypes.has(input.type)) {
      errors.push(`${label}.type has invalid value ${JSON.stringify(input.type)}`);
    }
    if (input.type === "enum") {
      if (!Array.isArray(input.values) || input.values.length === 0) {
        errors.push(`${label}.values must be a non-empty array for enum inputs`);
      } else {
        for (const value of input.values) {
          if (typeof value !== "string" || value.trim() === "") {
            errors.push(`${label}.values contains an invalid enum token`);
          }
        }
      }
    }
    if (
      Object.hasOwn(input, "min") &&
      (!Number.isFinite(input.min) || input.min < 0)
    ) {
      errors.push(`${label}.min must be a non-negative finite number`);
    }
    if (
      Object.hasOwn(input, "max") &&
      (!Number.isFinite(input.max) || input.max < 0)
    ) {
      errors.push(`${label}.max must be a non-negative finite number`);
    }
    if (!Object.hasOwn(input, "default")) {
      errors.push(`${label}.default must be declared, using null when omitted`);
    }
    if (!validEmptyStringRules.has(input.empty_string)) {
      errors.push(`${label}.empty_string has invalid value ${JSON.stringify(input.empty_string)}`);
    }
    if (!validInputNormalizations.has(input.normalization)) {
      errors.push(`${label}.normalization has invalid value ${JSON.stringify(input.normalization)}`);
    }
    if (!validInputInvalidReasons.has(input.invalid_reason)) {
      errors.push(`${label}.invalid_reason must be usage_error or configuration_error`);
    }
    if (!validInputSummaryEmission.has(input.summary_emission)) {
      errors.push(`${label}.summary_emission has invalid value ${JSON.stringify(input.summary_emission)}`);
    }
    if (!validInputChildForwarding.has(input.child_forwarding)) {
      errors.push(`${label}.child_forwarding has invalid value ${JSON.stringify(input.child_forwarding)}`);
    }
  }
}
