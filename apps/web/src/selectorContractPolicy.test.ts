import { readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

const thisFile = fileURLToPath(import.meta.url);
const repoRoot = path.resolve(path.dirname(thisFile), "../../..");

const scannedRoots = [
  "apps/web/src",
  "apps/web/e2e",
  "packages/test-utils/src",
] as const;

const ignoredFilePatterns = [
  /\.d\.ts$/u,
  /selectorContractPolicy\.test\.ts$/u,
] as const;

const ignoredPathSegments = new Set(["node_modules", "dist", "coverage"]);

type SelectorOwnership = {
  readonly owner: string;
  readonly pattern: RegExp;
  readonly reason: string;
  readonly scope: string;
};

const sharedBuilderOwnedSelectorPatterns = [
  /^row-history-/u,
  /^reference-pack-(?:admin-panel|file|import|job-status|reload|cancel|refresh-all|refresh-selected|row|error)(?:-|$)/u,
  /^save-state$/u,
] as const;

const appLocalSelectorOwnership = [
  {
    owner: "apps/web phase 1 auth/account/admin surfaces",
    pattern: /^(?:auth|account|admin)-/u,
    reason:
      "Phase 1 selectors are retained app-local surface anchors until the FE-P1 selector row promotes them.",
    scope: "apps/web Phase1Surface, Phase1Harness, and phase1 browser coverage",
  },
  {
    owner: "apps/web landing and route shell",
    pattern:
      /^(?:app-shell$|incident-landing$|landing-|workbook-(?:current-user|loading|focus-anchor)$|debug-harness-loading$)/u,
    reason:
      "Landing and route-shell selectors are app-local until FE-P1/FE-P2 promote shared shell builders.",
    scope: "apps/web App landing, route handoff, and keyboard focus coverage",
  },
  {
    owner: "apps/web incident administration",
    pattern: /^incident-(?:admin|summary|pref|patch)-/u,
    reason:
      "Incident admin selectors are app-local to the incident admin panel and phase 2 coverage.",
    scope: "apps/web IncidentAdminPanel and phase2 browser coverage",
  },
  {
    owner: "apps/web phase harnesses",
    pattern:
      /^(?:phase1-|phase2-|session-|create-|probe-|current-incident-|patch-|incident-discovery$|default-workbook-pref$|user-workbook-pref$|membership-|reload-extensions$|extensions-list$|last-)/u,
    reason:
      "Debug phase harness selectors are retained app-local harness controls, not shared product selectors.",
    scope: "apps/web Phase1Harness, Phase2Harness, and phase support specs",
  },
  {
    owner: "apps/web timeline workbook surface",
    pattern:
      /^(?:timeline-(?:blur-surface|refresh-error|inspector|inspector-message|evidence-file-(?:draft|timeline-1))|controlled-input$|entity-load-error$)/u,
    reason:
      "Timeline shell selectors are app-local workbook anchors pending later FE phase selector promotion.",
    scope: "apps/web WorkbookShell and timeline workbook support tests",
  },
  {
    owner: "apps/web row fixture controls",
    pattern:
      /^(?:row-record-1-(?:mark-reviewed|replacement-id|supersede)|phase8-row-record-[12])$/u,
    reason:
      "These fixed row selector literals target a unit-test fixture; variable row selector construction remains builder-owned.",
    scope: "apps/web phase 3 autosave unit fixture",
  },
  {
    owner: "apps/web presence and sync surfaces",
    pattern: /^(?:presence-|pending-queue-)/u,
    reason:
      "Presence and pending-queue selectors are app-local runtime affordances with existing shared builders staged separately.",
    scope: "apps/web collaboration unit and browser coverage",
  },
  {
    owner: "apps/web conflict resolver",
    pattern: /^(?:conflict-|paste-conflict-)/u,
    reason:
      "Conflict resolver selectors are retained app-local controls until the FE-P7 conflict surface is promoted.",
    scope: "apps/web WorkbookShell conflict resolver and phase 6/9 coverage",
  },
  {
    owner: "apps/web entity merge and inspector controls",
    pattern: /^(?:merge-|host-inspector$|identity-inspector$)/u,
    reason:
      "Entity merge selectors are app-local workflow controls for existing phase 4 coverage.",
    scope: "apps/web entity inspector merge controls and browser specs",
  },
  {
    owner: "apps/web assessment controls",
    pattern: /^assessment-/u,
    reason:
      "Assessment selectors are app-local workbook controls pending later system-view selector promotion.",
    scope: "apps/web assessment surface and browser specs",
  },
  {
    owner: "apps/web generic mutation controls",
    pattern: /^generic-/u,
    reason:
      "Generic system-view mutation selectors are app-local controls until reusable builders are introduced per surface.",
    scope: "apps/web generic mutation UI and phase 9 coverage",
  },
  {
    owner: "apps/web coordination controls",
    pattern: /^(?:party-link-|task-lifecycle-|decision-supersede-)/u,
    reason:
      "Coordination workflow selectors are app-local controls for current phase 9 browser coverage.",
    scope: "apps/web coordination generic mutation controls",
  },
  {
    owner: "apps/web evidence fixtures",
    pattern:
      /^evidence-(?:preview-panel$|preview-evidence-1$|download-evidence-1$)/u,
    reason:
      "These retained evidence literals are fixed existing fixtures; variable evidence selectors are builder-owned.",
    scope: "apps/web evidence surface tests and visual fixtures",
  },
  {
    owner: "apps/web test-local mocks",
    pattern: /^mock-/u,
    reason:
      "Mock selectors exist only inside unit-test component doubles and do not cross into runtime code.",
    scope: "apps/web unit tests",
  },
] as const satisfies readonly SelectorOwnership[];

const forbiddenPatterns = [
  {
    message: "dynamic data-testid template",
    pattern: /data-testid=\{`[^`]*\$\{[^`]*`\}/u,
  },
  {
    message: "dynamic getByTestId template",
    pattern: /getByTestId\(\s*`[^`]*\$\{[^`]*`/u,
  },
  {
    message: "raw dynamic data-testid selector",
    pattern: /\[data-testid(?:[*^$|~]?=)"?\$\{/u,
  },
  {
    message: "Timeline workbook shell readiness text",
    pattern: /getByText\(\s*["']Timeline workbook shell["']/u,
  },
  {
    message: "Timeline mutation substrate readiness text",
    pattern: /getByText\(\s*["']Timeline mutation substrate["']/u,
  },
  {
    message: "Current incident role readiness text",
    pattern: /getByText\(\s*["']Current incident role:/u,
  },
  {
    message: "heading-name readiness wait",
    pattern: /getByRole\(\s*["']heading["']\s*,\s*\{\s*name:/u,
  },
] as const;

const exactDataTestIdLiteralPattern =
  /data-testid\s*=\s*(["'])([^"'`{}$]+)\1/gu;
const exactTestIdConsumerPattern =
  /\b(?:getByTestId|findByTestId|queryByTestId)\(\s*(["'])([^"'`{}$]+)\1/gu;
const rawDataTestIdSelectorPattern =
  /\[data-testid([*^$|~]?=)(["'])([^"'`{}$]+)\2\]/gu;

function listSourceFiles(relativeRoot: string): string[] {
  const absoluteRoot = path.join(repoRoot, relativeRoot);
  const entries = readdirSync(absoluteRoot, { withFileTypes: true });
  const files: string[] = [];

  for (const entry of entries) {
    const absolutePath = path.join(absoluteRoot, entry.name);
    const relativePath = path.relative(repoRoot, absolutePath);
    if (entry.isDirectory()) {
      if (
        !ignoredPathSegments.has(entry.name) &&
        !entry.name.endsWith("-snapshots")
      ) {
        files.push(...listSourceFiles(relativePath));
      }
      continue;
    }
    if (!entry.isFile()) {
      continue;
    }
    if (!/\.[cm]?[tj]sx?$/u.test(entry.name)) {
      continue;
    }
    if (shouldIgnoreFile(entry.name)) {
      continue;
    }
    files.push(relativePath);
  }

  return files;
}

function shouldIgnoreFile(fileName: string): boolean {
  return ignoredFilePatterns.some((pattern) => pattern.test(fileName));
}

function lineNumberForOffset(content: string, offset: number): number {
  return content.slice(0, offset).split("\n").length;
}

function sharedBuilderOwns(token: string): boolean {
  return sharedBuilderOwnedSelectorPatterns.some((pattern) =>
    pattern.test(token),
  );
}

function appLocalOwnershipFor(token: string): SelectorOwnership | null {
  return (
    appLocalSelectorOwnership.find((entry) => entry.pattern.test(token)) ?? null
  );
}

function assertOwnershipEntriesAreDocumented() {
  const violations = appLocalSelectorOwnership.flatMap((entry, index) => {
    const missing = [];
    if (entry.owner.trim() === "") {
      missing.push("owner");
    }
    if (entry.scope.trim() === "") {
      missing.push("scope");
    }
    if (entry.reason.trim() === "") {
      missing.push("reason");
    }
    return missing.map((field) => `allowlist[${index}] missing ${field}`);
  });
  expect(violations).toEqual([]);
}

function recordExactTokenViolation(
  violations: string[],
  options: {
    readonly content: string;
    readonly file: string;
    readonly index: number;
    readonly kind: string;
    readonly token: string;
  },
) {
  if (sharedBuilderOwns(options.token)) {
    violations.push(
      `${options.file}:${lineNumberForOffset(
        options.content,
        options.index,
      )} ${options.kind} for shared selector "${options.token}" must use @cartulary/ui-contracts`,
    );
    return;
  }
  if (appLocalOwnershipFor(options.token) === null) {
    violations.push(
      `${options.file}:${lineNumberForOffset(
        options.content,
        options.index,
      )} unowned ${options.kind} "${options.token}"`,
    );
  }
}

describe("selector contract policy", () => {
  it("keeps cross-boundary selector templates behind shared builders", () => {
    assertOwnershipEntriesAreDocumented();
    const violations: string[] = [];

    for (const file of scannedRoots.flatMap(listSourceFiles)) {
      const absolutePath = path.join(repoRoot, file);
      if (!statSync(absolutePath).isFile()) {
        continue;
      }
      const content = readFileSync(absolutePath, "utf8");
      for (const { message, pattern } of forbiddenPatterns) {
        const flags = pattern.flags.includes("g")
          ? pattern.flags
          : `${pattern.flags}g`;
        const globalPattern = new RegExp(pattern.source, flags);
        for (const match of content.matchAll(globalPattern)) {
          if (match.index === undefined) {
            continue;
          }
          violations.push(
            `${file}:${lineNumberForOffset(content, match.index)} ${message}`,
          );
        }
      }
      for (const match of content.matchAll(exactDataTestIdLiteralPattern)) {
        if (match.index === undefined || match[2] === undefined) {
          continue;
        }
        recordExactTokenViolation(violations, {
          content,
          file,
          index: match.index,
          kind: "data-testid literal",
          token: match[2],
        });
      }
      for (const match of content.matchAll(exactTestIdConsumerPattern)) {
        if (match.index === undefined || match[2] === undefined) {
          continue;
        }
        recordExactTokenViolation(violations, {
          content,
          file,
          index: match.index,
          kind: "test-id consumer literal",
          token: match[2],
        });
      }
      for (const match of content.matchAll(rawDataTestIdSelectorPattern)) {
        if (
          match.index === undefined ||
          match[1] === undefined ||
          match[3] === undefined
        ) {
          continue;
        }
        if (match[1] !== "=") {
          violations.push(
            `${file}:${lineNumberForOffset(
              content,
              match.index,
            )} raw data-testid prefix or suffix selector`,
          );
          continue;
        }
        recordExactTokenViolation(violations, {
          content,
          file,
          index: match.index,
          kind: "raw data-testid selector literal",
          token: match[3],
        });
      }
    }

    expect(violations).toEqual([]);
  });
});
