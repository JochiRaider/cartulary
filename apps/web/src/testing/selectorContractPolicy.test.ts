import { readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import * as ts from "typescript";
import { describe, expect, it } from "vitest";

const thisFile = fileURLToPath(import.meta.url);
const repoRoot = path.resolve(path.dirname(thisFile), "../../../..");

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
  /^(?:auth|account|admin)-/u,
  /^app-shell$/u,
  /^debug-harness-loading$/u,
  /^deployment-user-row-/u,
  /^incident-landing$/u,
  /^landing-admin-/u,
  /^landing-/u,
  /^row-history-/u,
  /^saved-view-/u,
  /^reference-pack-(?:admin-panel|file|import|job-status|reload|cancel|refresh-all|refresh-selected|row|error)(?:-|$)/u,
  /^save-state$/u,
  /^timeline-inspector$/u,
  /^timeline-inspector-message$/u,
  /^workbook-(?:current-user|loading)$/u,
] as const;

const appLocalSelectorOwnership = [
  {
    owner: "apps/web workbook route focus support",
    pattern: /^workbook-focus-anchor$/u,
    reason:
      "The workbook focus anchor is a keyboard support fixture retained app-local until its owning focus surface is promoted.",
    scope: "apps/web route handoff and keyboard focus coverage",
  },
  {
    owner: "apps/web incident administration",
    pattern: /^incident-(?:admin|summary|pref|patch|lifecycle|close|reopen)-/u,
    reason:
      "Incident admin selectors are app-local to the incident admin panel and phase 2 coverage.",
    scope: "apps/web IncidentAdminPanel and phase2 browser coverage",
  },
  {
    owner: "apps/web phase harnesses",
    pattern:
      /^(?:phase1-|phase2-|session-|create-|probe-|current-incident-(?:id|key|title|version)$|patch-|incident-discovery$|default-workbook-pref$|user-workbook-pref$|membership-|reload-extensions$|extensions-list$|last-)/u,
    reason:
      "Debug phase harness selectors are retained app-local harness controls, not shared product selectors.",
    scope: "apps/web Phase1Harness, Phase2Harness, and phase support specs",
  },
  {
    owner: "apps/web timeline workbook surface",
    pattern:
      /^(?:timeline-(?:blur-surface|refresh-error|evidence-file-(?:draft|timeline-1))|controlled-input$|entity-load-error$)/u,
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

const rawDataTestIdSelectorPattern =
  /\[data-testid([*^$|~]?=)(["'])([^"'`{}$]+)\2\]/gu;

const callSelectorSinks = [
  {
    argumentIndex: 0,
    kind: "test-id consumer literal",
    names: ["getByTestId", "findByTestId", "queryByTestId"],
  },
  {
    argumentIndex: 0,
    kind: "Phase1Page.requireText literal",
    names: ["requireText"],
  },
  {
    argumentIndex: 0,
    kind: "Phase1Page.setCheckbox literal",
    names: ["setCheckbox"],
  },
] as const;

const objectSelectorSinks = [
  {
    kind: "P1 accessibility focusTestId literal",
    mode: "single",
    name: "focusTestId",
  },
  {
    kind: "P1 accessibility tabStops literal",
    mode: "array",
    name: "tabStops",
  },
] as const;

const jsxSelectorSinks = [
  {
    kind: "data-testid literal",
    name: "data-testid",
  },
] as const;

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

function selectorSinkName(expression: ts.Expression): string | null {
  if (ts.isPropertyAccessExpression(expression)) {
    return expression.name.text;
  }
  if (ts.isIdentifier(expression)) {
    return expression.text;
  }
  return null;
}

function propertyNameText(name: ts.PropertyName): string | null {
  if (ts.isIdentifier(name) || ts.isStringLiteral(name)) {
    return name.text;
  }
  return null;
}

function jsxAttributeNameText(name: ts.JsxAttributeName): string | null {
  if (ts.isIdentifier(name)) {
    return name.text;
  }
  return null;
}

function stringLiteralToken(
  node: ts.Node,
): ts.StringLiteral | ts.NoSubstitutionTemplateLiteral | null {
  if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) {
    return node;
  }
  return null;
}

function recordSelectorSinkArgument(
  violations: string[],
  options: {
    readonly content: string;
    readonly file: string;
    readonly kind: string;
    readonly node: ts.Node;
    readonly sourceFile: ts.SourceFile;
  },
) {
  const token = stringLiteralToken(options.node);
  if (token !== null) {
    recordExactTokenViolation(violations, {
      content: options.content,
      file: options.file,
      index: token.getStart(options.sourceFile),
      kind: options.kind,
      token: token.text,
    });
    return;
  }
  if (ts.isTemplateExpression(options.node)) {
    violations.push(
      `${options.file}:${lineNumberForOffset(
        options.content,
        options.node.getStart(options.sourceFile),
      )} dynamic ${options.kind}`,
    );
  }
}

function scanAstSelectorSinks(
  violations: string[],
  options: {
    readonly content: string;
    readonly file: string;
  },
) {
  const sourceFile = ts.createSourceFile(
    options.file,
    options.content,
    ts.ScriptTarget.Latest,
    true,
    options.file.endsWith(".tsx") ? ts.ScriptKind.TSX : ts.ScriptKind.TS,
  );

  function visit(node: ts.Node) {
    if (ts.isCallExpression(node)) {
      const name = selectorSinkName(node.expression);
      const sink =
        name === null
          ? undefined
          : callSelectorSinks.find((entry) =>
              entry.names.some((candidate) => candidate === name),
            );
      const argument =
        sink === undefined ? undefined : node.arguments[sink.argumentIndex];
      if (sink !== undefined && argument !== undefined) {
        recordSelectorSinkArgument(violations, {
          content: options.content,
          file: options.file,
          kind: sink.kind,
          node: argument,
          sourceFile,
        });
      }
    }

    if (ts.isJsxAttribute(node)) {
      const name = jsxAttributeNameText(node.name);
      const sink = jsxSelectorSinks.find((entry) => entry.name === name);
      if (sink !== undefined && node.initializer !== undefined) {
        if (ts.isStringLiteral(node.initializer)) {
          recordSelectorSinkArgument(violations, {
            content: options.content,
            file: options.file,
            kind: sink.kind,
            node: node.initializer,
            sourceFile,
          });
        } else if (
          ts.isJsxExpression(node.initializer) &&
          node.initializer.expression !== undefined
        ) {
          recordSelectorSinkArgument(violations, {
            content: options.content,
            file: options.file,
            kind: sink.kind,
            node: node.initializer.expression,
            sourceFile,
          });
        }
      }
    }

    if (ts.isPropertyAssignment(node)) {
      const name = propertyNameText(node.name);
      const sink = objectSelectorSinks.find((entry) => entry.name === name);
      if (sink?.mode === "single") {
        recordSelectorSinkArgument(violations, {
          content: options.content,
          file: options.file,
          kind: sink.kind,
          node: node.initializer,
          sourceFile,
        });
      } else if (
        sink?.mode === "array" &&
        ts.isArrayLiteralExpression(node.initializer)
      ) {
        for (const element of node.initializer.elements) {
          recordSelectorSinkArgument(violations, {
            content: options.content,
            file: options.file,
            kind: sink.kind,
            node: element,
            sourceFile,
          });
        }
      }
    }

    ts.forEachChild(node, visit);
  }

  visit(sourceFile);
}

function collectSelectorPolicyViolations(
  file: string,
  content: string,
): string[] {
  const violations: string[] = [];
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
  scanAstSelectorSinks(violations, { content, file });
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
  return violations;
}

describe("selector contract policy", () => {
  it("keeps phase-named selector helpers out of production modules", () => {
    const violations = listSourceFiles("apps/web/src").flatMap((file) => {
      if (
        /(?:^|\/)debug\//u.test(file) ||
        /(?:^|\/)testing\//u.test(file) ||
        /(?:\.test|TestSupport)\.[cm]?[tj]sx?$/u.test(file)
      ) {
        return [];
      }
      const content = readFileSync(path.join(repoRoot, file), "utf8");
      return [...content.matchAll(/\bphase\d+[A-Z]\w*TestIds?\b/gu)].map(
        (match) =>
          `${file}:${lineNumberForOffset(content, match.index ?? 0)} ${match[0]}`,
      );
    });

    expect(violations).toEqual([]);
  });

  it("classifies authentication cross-boundary selector families as shared-builder owned", () => {
    const sharedSelectors = [
      "account-session-user-id",
      "account-error-public",
      "admin-create-user",
      "admin-error-code",
      "app-shell",
      "auth-login-username",
      "auth-error-details",
      "debug-harness-loading",
      "incident-landing",
      "landing-current-user",
      "landing-error-message",
      "workbook-current-user",
      "workbook-loading",
    ];

    for (const token of sharedSelectors) {
      expect(sharedBuilderOwns(token), token).toBe(true);
      expect(appLocalOwnershipFor(token), token).toBeNull();
    }

    expect(appLocalOwnershipFor("phase1-debug-request")).not.toBeNull();
    expect(appLocalOwnershipFor("incident-patch-button")).not.toBeNull();
    expect(appLocalOwnershipFor("workbook-focus-anchor")).not.toBeNull();
  });

  it("keeps cross-boundary selector templates behind shared builders", () => {
    assertOwnershipEntriesAreDocumented();
    const violations: string[] = [];

    for (const file of scannedRoots.flatMap(listSourceFiles)) {
      const absolutePath = path.join(repoRoot, file);
      if (!statSync(absolutePath).isFile()) {
        continue;
      }
      const content = readFileSync(absolutePath, "utf8");
      violations.push(...collectSelectorPolicyViolations(file, content));
    }

    expect(violations).toEqual([]);
  });

  it("rejects raw FE-P1 selector literals at helper-level selector sinks", () => {
    const violations = collectSelectorPolicyViolations(
      "apps/web/e2e/raw-helper-fixture.ts",
      `
        phase1.requireText("auth-bootstrap-secret-base32");
        phase1.setCheckbox("admin-patch-is-active", false);
        await expectP1SurfaceA11y(page, {
          focusTestId: "auth-login-submit",
          tabStops: ["landing-refresh"],
        });
      `,
    );

    expect(violations).toEqual([
      expect.stringContaining(
        'Phase1Page.requireText literal for shared selector "auth-bootstrap-secret-base32"',
      ),
      expect.stringContaining(
        'Phase1Page.setCheckbox literal for shared selector "admin-patch-is-active"',
      ),
      expect.stringContaining(
        'P1 accessibility focusTestId literal for shared selector "auth-login-submit"',
      ),
      expect.stringContaining(
        'P1 accessibility tabStops literal for shared selector "landing-refresh"',
      ),
    ]);
  });
});
