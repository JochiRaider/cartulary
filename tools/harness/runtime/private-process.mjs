import process from "node:process";

// Call only in process roots that create private runtime or retained evidence
// and never generate repository source. Every descendant then inherits
// owner-only creation modes without coupling source generators to this policy.
export function enforcePrivateProcessUmask() {
  process.umask(0o077);
}
