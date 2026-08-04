// These imports are intentionally unresolved. If any legacy or private package
// path becomes public again, TypeScript reports the corresponding directive as
// unused and the canonical frontend typecheck fails.

// @ts-expect-error the aggregate root package export is unsupported
import "@cartulary/protocol-ts";

// @ts-expect-error the legacy Core HTTP package export is unsupported
import "@cartulary/protocol-ts/core-http";

// @ts-expect-error private facade paths are unsupported
import "@cartulary/protocol-ts/facade/runtimeValidation";

// @ts-expect-error generated paths are unsupported
import "@cartulary/protocol-ts/generated/core-http-types";

// @ts-expect-error internal paths are unsupported
import "@cartulary/protocol-ts/internal/decoder";
