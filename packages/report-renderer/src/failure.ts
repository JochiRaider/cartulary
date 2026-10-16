// Only these bounded codes cross the renderer's process boundary. Vendor
// exceptions can contain redacted source text, paths, and URLs; never forward them.
export class RenderFailure extends Error {
  readonly failureCode: string;
  readonly reasonCode: string;

  readonly pageCounts:
    | { expected: number; observed: number; source: "slides" | "pdf" }
    | undefined;

  constructor(
    failureCode: string,
    reasonCode: string,
    pageCounts?: {
      expected: number;
      observed: number;
      source: "slides" | "pdf";
    },
  ) {
    super(
      `${failureCode}: ${reasonCode}${pageCounts ? ` (${pageCounts.source} expected=${pageCounts.expected}, observed=${pageCounts.observed})` : ""}`,
    );
    this.name = "RenderFailure";
    this.failureCode = failureCode;
    this.reasonCode = reasonCode;
    this.pageCounts = pageCounts;
  }
}
