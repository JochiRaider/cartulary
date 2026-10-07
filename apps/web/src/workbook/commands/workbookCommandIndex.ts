import { cartularyDesignPresentation } from "@cartulary/ui-contracts";

const families = cartularyDesignPresentation.workbookWorkbench.command_families;
export type WorkbookCommandFamily = (typeof families)[number];
export type WorkbookCommandTarget =
  | { readonly kind: "shell" }
  | { readonly kind: "surface"; readonly viewSchemaId: string }
  | {
      readonly kind: "record" | "cell";
      readonly viewSchemaId: string;
      readonly recordId: string;
      readonly fieldKey?: string;
    }
  | {
      readonly kind: "selection";
      readonly viewSchemaId: string;
      readonly recordIds: readonly string[];
    };
export type WorkbookCommandDescriptor = {
  readonly id: string;
  readonly family: WorkbookCommandFamily;
  readonly label: string;
  readonly terms: readonly string[];
  readonly targetKind: WorkbookCommandTarget["kind"];
  readonly availability: (target: WorkbookCommandTarget) => string | null;
  readonly invoke: (target: WorkbookCommandTarget) => boolean;
};
const normalize = (text: string) => text.normalize("NFC").toLowerCase();
function compare(left: string, right: string) {
  const a = Array.from(left, (value) => value.codePointAt(0) ?? 0);
  const b = Array.from(right, (value) => value.codePointAt(0) ?? 0);
  for (let i = 0; i < Math.min(a.length, b.length); i++)
    if (a[i] !== b[i]) return (a[i] ?? 0) - (b[i] ?? 0);
  return a.length - b.length;
}
export function searchWorkbookCommands(
  commands: readonly WorkbookCommandDescriptor[],
  input: string,
) {
  const tokens = normalize(input)
    .split(/\p{White_Space}+/u)
    .filter(Boolean);
  const phrase = tokens.join(" ");
  return commands
    .flatMap((command) => {
      const label = normalize(command.label);
      const terms = command.terms.map(normalize);
      if (
        !tokens.every(
          (token) =>
            label.includes(token) || terms.some((term) => term.includes(token)),
        )
      )
        return [];
      const rank =
        !tokens.length || label.startsWith(phrase)
          ? 0
          : tokens.every((token) => label.includes(token))
            ? 1
            : 2;
      return [{ command, label, rank }];
    })
    .sort(
      (a, b) =>
        a.rank - b.rank ||
        families.indexOf(a.command.family) -
          families.indexOf(b.command.family) ||
        compare(a.label, b.label) ||
        compare(a.command.id, b.command.id),
    )
    .map((item) => item.command);
}
