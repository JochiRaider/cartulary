export type GeneratedValidationError = {
  readonly instancePath?: string;
  readonly keyword?: string;
};

export type GeneratedValidator = ((value: unknown) => boolean) & {
  readonly errors?: readonly GeneratedValidationError[] | null;
};

export type DecodeFailure = {
  readonly boundary: "generated_protocol";
  readonly instancePath: string;
  readonly reasonCategory:
    | "constraint_violation"
    | "invalid_format"
    | "invalid_type"
    | "invalid_value"
    | "required_member"
    | "schema_mismatch"
    | "unknown_member";
  readonly schemaId: string;
};

export type DecodeResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: DecodeFailure };

export type Decoder<T> = {
  readonly schemaId: string;
  readonly decode: (value: unknown) => DecodeResult<T>;
};

function reasonCategory(
  keyword: string | undefined,
): DecodeFailure["reasonCategory"] {
  switch (keyword) {
    case "additionalProperties":
      return "unknown_member";
    case "required":
      return "required_member";
    case "type":
      return "invalid_type";
    case "const":
    case "enum":
      return "invalid_value";
    case "format":
      return "invalid_format";
    case "oneOf":
    case "anyOf":
      return "schema_mismatch";
    default:
      return "constraint_violation";
  }
}

export function createDecoder<T>(
  schemaId: string,
  validate: GeneratedValidator,
): Decoder<T> {
  return Object.freeze({
    schemaId,
    decode(value: unknown): DecodeResult<T> {
      if (validate(value)) {
        return { ok: true, value: value as T };
      }
      const firstError = validate.errors?.[0];
      return {
        ok: false,
        error: {
          boundary: "generated_protocol",
          instancePath: firstError?.instancePath ?? "",
          reasonCategory: reasonCategory(firstError?.keyword),
          schemaId,
        },
      };
    },
  });
}
