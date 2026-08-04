// These imports are intentionally unresolved. If a removed alias, private
// module, or low-level export becomes public, TypeScript reports the matching
// directive as unused and the canonical frontend typecheck fails.

// @ts-expect-error the aggregate root package export is unsupported
import "@cartulary/test-utils";

// @ts-expect-error the removed accessibility alias is unsupported
import "@cartulary/test-utils/accessibility";

// @ts-expect-error the removed visual alias is unsupported
import "@cartulary/test-utils/visual";

// @ts-expect-error private implementation paths are unsupported
import "@cartulary/test-utils/browser";

// @ts-expect-error low-level browser mechanics are not facade exports
import { delay } from "@cartulary/test-utils/grid";

void delay;
