import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { httpOperationBindings } from "@cartulary/protocol-ts/http";
import * as ts from "typescript";
import { describe, expect, it } from "vitest";
import { jobResultSummaryCodes } from "../services/commonJobContract";
import { clientErrorCodes } from "../shared/publicError";

const webSourceDirectory = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);
const repositoryRoot = path.resolve(webSourceDirectory, "../../..");

/**
 * Network Flow keeps its own error registry and routes outside the generated
 * HTTP bindings; collaboration branches on WebSocket reasons. Both are owned
 * by other contracts and stay outside this HTTP public-error policy.
 */
function excluded(relative: string): boolean {
  return (
    relative.startsWith("networkFlow/") ||
    relative.startsWith("collaboration/") ||
    relative.startsWith("measurement/") ||
    path.basename(relative).includes("NetworkFlow")
  );
}

function runtimeSourceFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const target = path.join(directory, entry.name);
    if (entry.isDirectory()) return runtimeSourceFiles(target);
    return /\.(?:ts|tsx)$/u.test(entry.name) &&
      !entry.name.includes(".test.") &&
      !entry.name.endsWith("TestSupport.ts") &&
      !entry.name.endsWith("TestSupport.tsx")
      ? [target]
      : [];
  });
}

/** Codes are read from `code` or `publicCode`, including job summary codes. */
function publicCodeReference(node: ts.Node): boolean {
  let expression = node;
  while (
    ts.isParenthesizedExpression(expression) ||
    ts.isNonNullExpression(expression)
  )
    expression = expression.expression;
  if (ts.isBinaryExpression(expression)) {
    // `error?.code ?? ""` keeps the code on its left.
    return (
      expression.operatorToken.kind === ts.SyntaxKind.QuestionQuestionToken &&
      publicCodeReference(expression.left)
    );
  }
  const name = ts.isPropertyAccessExpression(expression)
    ? expression.name.text
    : ts.isIdentifier(expression)
      ? expression.text
      : null;
  return name === "code" || name === "publicCode";
}

type CodeLiteral = { readonly literal: string; readonly line: number };
type CodeComparisons = {
  /** String literals compared or matched against a code. */
  readonly literals: CodeLiteral[];
  /**
   * Lines matching a code against an array literal with `.includes`, which
   * the type checker cannot narrow; `isFailureCode` must be used instead.
   */
  readonly arrayIncludes: number[];
};

/** Every string literal a code is compared or matched against. */
function comparedPublicCodeLiterals(
  source: string,
  tsx = false,
): CodeComparisons {
  // `.ts` sources parse as TS: in TSX, generic arrows such as `<T>(x) =>`
  // would be read as JSX and hide every later comparison.
  const file = ts.createSourceFile(
    tsx ? "subject.tsx" : "subject.ts",
    source,
    ts.ScriptTarget.Latest,
    true,
    tsx ? ts.ScriptKind.TSX : ts.ScriptKind.TS,
  );
  const found: CodeLiteral[] = [];
  const arrayIncludes: number[] = [];
  const record = (literal: ts.Node) => {
    if (
      ts.isStringLiteral(literal) ||
      ts.isNoSubstitutionTemplateLiteral(literal)
    )
      found.push({
        literal: literal.text,
        line: file.getLineAndCharacterOfPosition(literal.getStart()).line + 1,
      });
  };
  const visit = (node: ts.Node) => {
    if (
      ts.isBinaryExpression(node) &&
      [
        ts.SyntaxKind.EqualsEqualsEqualsToken,
        ts.SyntaxKind.ExclamationEqualsEqualsToken,
        ts.SyntaxKind.EqualsEqualsToken,
        ts.SyntaxKind.ExclamationEqualsToken,
      ].includes(node.operatorToken.kind)
    ) {
      if (publicCodeReference(node.left)) record(node.right);
      if (publicCodeReference(node.right)) record(node.left);
    }
    if (
      ts.isCaseClause(node) &&
      ts.isSwitchStatement(node.parent.parent) &&
      publicCodeReference(node.parent.parent.expression)
    )
      record(node.expression);
    if (ts.isCallExpression(node) && node.arguments.length >= 1) {
      const callee = node.expression;
      const [first, second] = node.arguments;
      // `isFailureCode(error.code, ["…"])`
      if (
        ts.isIdentifier(callee) &&
        callee.text === "isFailureCode" &&
        second !== undefined &&
        ts.isArrayLiteralExpression(second)
      )
        second.elements.forEach(record);
      // `["…"].includes(error.code)` is banned: a typed code is matched
      // with `isFailureCode`, whose candidates the type checker verifies.
      if (
        ts.isPropertyAccessExpression(callee) &&
        callee.name.text === "includes" &&
        ts.isArrayLiteralExpression(callee.expression) &&
        first !== undefined &&
        publicCodeReference(first)
      )
        arrayIncludes.push(
          file.getLineAndCharacterOfPosition(node.getStart()).line + 1,
        );
    }
    ts.forEachChild(node, visit);
  };
  visit(file);
  return { literals: found, arrayIncludes };
}

/** Every terminal success code the Job-kind contracts declare. */
function contractJobSuccessCodes(): string[] {
  const codes = new Set<string>();
  const collect = (value: unknown): void => {
    if (Array.isArray(value)) {
      value.forEach(collect);
      return;
    }
    if (value === null || typeof value !== "object") return;
    for (const [key, member] of Object.entries(value)) {
      if (key === "terminal_success_codes" && Array.isArray(member))
        for (const code of member) codes.add(String(code));
      else collect(member);
    }
  };
  const fragments = path.join(repositoryRoot, "contracts/extensions/fragments");
  for (const file of [
    path.join(repositoryRoot, "contracts/jobs/base.v2.json"),
    ...readdirSync(fragments)
      .filter((name) => name.endsWith(".json"))
      .map((name) => path.join(fragments, name)),
  ])
    collect(JSON.parse(readFileSync(file, "utf8")));
  return [...codes].sort();
}

describe("public error code branch policy", () => {
  it("compares public error codes only against declared operation codes or client-local codes", () => {
    // Every operation declares 500 internal_error and 503 service_unavailable
    // (Core 01 REQ-01-702), so the bindings alone seed the server codes.
    // Job summary codes are typed from the Job-kind contracts (below).
    const declared = new Set<string>([
      ...clientErrorCodes,
      ...jobResultSummaryCodes,
    ]);
    for (const binding of Object.values(httpOperationBindings))
      for (const codes of Object.values(binding.error_codes_by_status))
        for (const code of codes as readonly string[]) declared.add(code);
    const violations: string[] = [];
    let compared = 0;
    for (const file of runtimeSourceFiles(webSourceDirectory)) {
      const relative = path
        .relative(webSourceDirectory, file)
        .split(path.sep)
        .join("/");
      if (excluded(relative)) continue;
      const { literals, arrayIncludes } = comparedPublicCodeLiterals(
        readFileSync(file, "utf8"),
        file.endsWith(".tsx"),
      );
      for (const { literal, line } of literals) {
        compared += 1;
        if (literal !== "" && !declared.has(literal))
          violations.push(`${relative}:${line} ${literal}`);
      }
      for (const line of arrayIncludes)
        violations.push(`${relative}:${line} [...].includes(code)`);
    }
    expect(compared).toBeGreaterThan(50);
    expect(violations).toEqual([]);
    const sample = comparedPublicCodeLiterals(
      [
        'if (error.code === "session_revoked") stop();',
        'if ("auth_required" !== failure?.publicCode) stop();',
        'switch (problem.code) { case "service_unavailable": break; }',
        'if (isFailureCode(error.code, ["csrf_failed"])) stop();',
        'if (["invalid_path_parameter"].includes(error?.code ?? "")) stop();',
        'if (job.result_summary?.code === "job_canceled") stop();',
        'if (failure.kind === "transport") stop();',
        'const failed = <T>(failure: F<T>) => failure.code !== "csrf_failed";',
      ].join("\n"),
    );
    expect(sample.literals.map(({ literal }) => literal)).toEqual([
      "session_revoked",
      "auth_required",
      "service_unavailable",
      "csrf_failed",
      "job_canceled",
      "csrf_failed",
    ]);
    expect(sample.arrayIncludes).toEqual([5]);
  });

  it("recognizes exactly the job result codes the Job-kind contracts declare", () => {
    // Core 01 §3.3.9.1: result_summary.code is registry-backed; job_canceled
    // is the common cancellation code and never a declared success code.
    const declared = contractJobSuccessCodes();
    expect(declared).not.toContain("job_canceled");
    expect([...jobResultSummaryCodes].sort()).toEqual(
      [...declared, "job_canceled"].sort(),
    );
  });
});
