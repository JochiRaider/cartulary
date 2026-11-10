# SQL ownership enforcement

Core 01 REQ-01-690 governs package, provider and source-owner capabilities. The
Testing Harness NLSpec governs analysis completeness. This guide explains their
implementation; no specification or guide is an executable checker input.

## Analysis and construction contract

`internal/testutil/sourcecheck` parses Go source and reports SQL by declaration.
`tools/harness/static-analysis/sql-relations.mjs` classifies physical relations,
reads, writes, locks and supported session controls. Backend boundary and
projection ownership policies consume the same structured analysis. There is no
lexical Go-string extraction fallback.

Supported raw-string construction is deliberately finite:

- Raw/interpreted literals, parentheses, grouped and cross-file constants,
  imported immutable constants and constant concatenation.
- Scope-aware local assignments, reassignment, `+=`, finite `if`/`switch`
  alternatives and bounded literal collections. Shadowing preserves identity.
  Unknown evaluation state cannot resolve a local binding through a same-named
  package constant; unsupported control-flow contexts fail explicitly.
- Uniquely resolved local helpers with exactly one supported string result and
  a complete return summary. Raw `strings.Builder` output is
  unsupported as SQL, including direct local construction. Use immutable
  expressions or sealed fragments instead. Ordinary text builders and bound
  parameter values remain supported. Unknown mutation and ambiguous definitions fail.
- Constant `fmt.Sprintf` templates using only `%s` and `%%`, with an exact
  argument count and analyzable operands. Indexes, flags, width, `%q` and other
  formatting verbs fail instead of approximating their output.

The analysis bounds expression alternatives at 256 and finite range inputs at
64. These are rejection limits, not truncation limits. Arbitrary application
call graphs, recursion and unrestricted runtime SQL are outside the contract.
Simplify a rejected builder or expose a narrow structural capability; do not add
an unknown-expression exemption or raise limits to hide combinatorial growth.

## Assignment and control-flow state

Helper summaries include every admitted explicit and bare return. A named string
result begins as the empty string; each return reads its reaching value. Nested
functions have independent return contexts. Select, goto, labels, labeled jumps
and switch fallthrough reject raw helper inference instead of retaining only
the observed safe returns. Loops use conservative state and cannot establish
summary completeness through an unproven exit. Raw tuple results remain unknown;
native result type resolution is a separate operation.

Any defer in a raw helper's own body rejects its return summary, including cleanup
defers, conditional defers and defers after an earlier return. For example:

```go
func query(input string, flag bool) (q string) {
    q = input
    defer func() { recover() }()
    if flag {
        panic("stop")
    }
    return "SELECT * FROM artifacts"
}
```

When the panic is recovered, this helper returns `input` without reaching the
literal return. Recovery need not write the result, and a panic need not be
explicit. Moving recovery into a named deferred function does not qualify the
helper. The checker rejects every defer instead of inferring recovery safety or
invocation ordering. Nested functions have their own admission contexts; their
defers do not belong to the enclosing helper. SQL inside rejected helpers and
nested functions is still inspected independently.

Named results with nested writers or address escapes cannot qualify raw SQL,
even with `return "SELECT * FROM artifacts"`: a deferred write can replace that
result before the caller receives it. This restriction also covers immediate,
concurrent and escaping writers without guessing scheduling. Cleanup defers remain
supported around ordinary SQL execution, bound data and independently qualified
sealed values, outside raw helper-result inference. Separate resource management
from a defer-free SQL builder. Replace rejected construction with
straight-line immutable expressions, supported finite branches or sealed values;
there is no legacy acceptance or forwarding exemption.
Builders needing an error or success result should return `(sqlshape.Fragment,
error)` or `(sqlshape.Fragment, bool)`. Preserve the typed fragment through callers
and convert it to text only at the database boundary. Links source export and
Network Flow field expressions use this pattern; their tuple status results do
not require raw-string tuple inference.

Every branch starts with independent scalar, collection and native-provenance
state. Continuing alternatives join together, including an omitted else/default;
returning alternatives do not supply successor values. A finite collection stores
its elements when constructed, so changing an initializer variable afterward does
not change the inferred collection. Range analysis considers every supported
element alternative, independent of branch order. Scope exit restores every state
channel, including synthetic record fields and prepared/batch provenance.
Static Go binding types survive reassignment; value provenance does not. Named
results and type-switch refinements retain their lexical types while boundaries
are inspected, so ordinary data fields cannot become spurious SQL callables.

Assignments and declarations evaluate their inputs before replacing destinations.
Every non-blank destination receives its corresponding result or explicit unknown
state. For example, `_, q = pair(input)` cannot preserve an earlier value of `q`.
Native result identities remain recognized at every position; this does not imply
raw-string tuple-result inference. Unsupported raw result positions must become
independently analyzable construction or qualified sealed values. Ordinary data
results used as bound parameters remain supported.

Loops qualify repeated boundaries using conservative incoming state. A raw binding
written in the body or post statement cannot contribute its pre-loop value as
proof of later iterations. Break, continue and return stop that path; unreachable
writes cannot restore qualification. Changed outer bindings remain unknown after
the loop, including zero-iteration and early-exit paths. There is no iteration-count
proof, unrolling or fixed-point solver. Unsupported jump/fallthrough construction
also loses inferred state before affected boundaries are inspected.

Prefer an immutable statement outside the loop, an independent initialization
before each iteration's execution, or sealed fragments. Keep runtime values in
bound parameters. A loop that executes `q` and then assigns `q = input` rejects;
a loop that initializes `q` to a qualified expression before each execution can
pass. SQL discovery alone is insufficient: execution-site assertions must check
the reaching alternatives or unresolved disposition. An unchanged parameter still
has the existing transport disposition and requires its exact consumer grant.

## Discovery and classification

All declared source roots must exist as real directories. Shared admission checks
reject linked path components, linked entries, nonregular sources and metadata,
listing or read failures before a collector claims completeness. Empty directories
remain subject to consumer coverage requirements. File inventories are sorted and
deduplicated; incidental directory names do not exclude executable sources.
Authored, generated and test-support scopes keep their distinct policy authority.
Fixtures must create required directories rather than relying on missing-root skips.

SQL-producing declarations are inspected even without a local execution call.
Expression discovery is independent of bounded evaluation: fields, indices,
nested calls, control expressions and closures cannot hide candidates. Unvisited
calls receive analysis without guessed state. Results include line/column and
analyzed, rejected or transport dispositions. Consumers reject missing fields,
unknown records and internally inconsistent dispositions; no legacy response
format is accepted. Known empty SQL and terminal closed
SQL conversions have explicit records; unknown SQL never becomes an empty pass.
Execution detection includes typed DB ports and native pgx/database/sql
query/exec, prepare, batch and copy calls, including imported aliases, embedded
interfaces and acquired handles. Bulk-copy relation identifiers must be finite
constants. Unknown SQL inputs fail with file/declaration diagnostics.

The SQL classifier separates statements, CTE bindings and query operands.
`SELECT`, `VALUES` and `TABLE` operands can participate in `UNION`, `INTERSECT`
and `EXCEPT`, including `ALL`/`DISTINCT` and parenthesized operands. INSERT source
queries are classified separately from the write target. `TABLE users` remains
a physical read after WITH or a set operator. CTE names shadow only unqualified
relations in their scope; `public.users` still denotes the physical table.
Relation aliases and locks remain local to their query. Unsupported operand
heads, missing operands and unclassified TABLE positions fail explicitly.

`internal/platform/sqlshape` supports more compositional builders with sealed
`Fragment` and `Statement` values:

- `Literal` and `Format` require analyzable templates. `Format` substitutes `%s`
  fragments; percent signs otherwise remain literal. It is distinct from fmt.
  Runtime constructors and Go inspection share `sqltemplate` lexical admission.
  Closed templates forbid SQL comments outside quoted data and require complete
  tokens around substitutions. Holes inside strings, dollar strings, quoted
  identifiers or partial tokens fail, as do NUL, unterminated tokens and Unicode
  escape tokens. Ordinary strings use standard-conforming PostgreSQL semantics;
  backslash escaping belongs to `E` strings. Parentheses may span fragments.
  `Literal` does not substitute percent text. Raw SQL comments remain supported.
  Invalid authored templates panic during construction and fail static analysis.
- `Concat`, `Join` and `Builder` preserve token boundaries. Read fragments cannot
  introduce write, lock or session-control operations. Relation clauses must be
  complete and literal within a qualified template.
- Values use ordinary database parameters. `Quote`, `Identifier`, `Number` and
  `Parameter` provide bounded value/identifier primitives, not relation authority.
  `Operator` preserves validated catalog comparison operators for Recovery.
- `Command` creates a complete mutation/lock template with literal relations;
  commands cannot be substituted as read fragments. `Text`/`CommandText` are
  terminal conversions. Concatenating arbitrary SQL after conversion is rejected.

The constructor surface is explicitly qualified. Adding an exported constructor
requires updating analyzer admission and both positive and negative fixtures.
Runtime sealing supplements static qualification; it does not authorize arbitrary
string templates or SQL function bodies.

## Captures and callable values

SQL execution uses direct native calls or narrow interface-port methods. SQL
function-valued fields, parameters, locals, indexed and returned values are
unsupported on invocation, extraction or transfer, regardless of inline, alias
or defined-type spelling. Replace `query := store.Query` with a direct port call;
do not add a forwarding grant. Ordinary callbacks may perform checked direct SQL. Generic function-value signatures with a
possible raw SQL argument and an unbound result reject rather than guessing the
type argument. Function constraints receive the same callable admission. Use a
concrete port/signature to resolve an ambiguous boundary; ordinary generic domain
callbacks without a possible SQL argument remain supported. Direct repository
helper calls keep body/declaration analysis; extracting such a helper still
requires function-value admission.

Captured raw SQL must be a constant or a local with one explicit initialization
that reaches the capture, no reassignment, no address escape and no nested writer.
For example, `q := "SELECT * FROM artifacts"; defer func() { db.Exec(ctx, q) }()`
is admissible. Adding `q = input` anywhere in the enclosing function rejects the
capture, including before closure creation. A separate immutable value copy can
retain its own qualified initialization; immutable unknown input stays unknown.
Globals, named results and mutable fields/indices cannot supply inferred captures.
Captured collections do not supply finite SQL alternatives. Define the finite list
inside the callback, capture an immutable scalar copy, or use sealed fragments.
Local list inference accepts direct range use and whole-binding assignments;
possible alias exposure, nested writes and field/index mutation reject inferred
SQL. Ordinary collections used as bound data remain supported.
Nested writers and address escapes also invalidate raw SQL at outer boundaries.
Prepared/batch provenance stays in its qualified local scope and does not cross
closure capture. Ordinary DB handle types and sealed SQL retain their existing
checks. Immediate, deferred, concurrent and escaped callbacks follow one rule;
analysis does not infer invocation timing or model heap aliases.

Links validation uses an immutable copy of its selected statement before passing
the row callback. Recovery rebuild callbacks capture sealed fragments, converting
to SQL text only at the direct query call. These patterns preserve the borrowed
transaction and existing query/error semantics without transport exemptions.

Every dot import is rejected in batches submitted for SQL analysis. Use an
explicit alias such as `proto "github.com/jackc/pgx/v5/pgproto3"`; naming the package
does not make prohibited protocol construction admissible. This rule includes
generated and build-constrained inputs. Import-only inspection still reports dot
imports, without granting SQL-analysis admission.

## Native execution admission

`sourcecheck/sql_native.go` identifies database boundaries by declared type and
signature. It follows repository aliases, embedded interfaces, fields, factories,
acquired handles and type-switch cases. An unresolved potential execution receiver
fails explicitly; a familiar receiver variable name never grants authority.
Repository ports require database result signatures and exact string-argument
positions. Unrelated declared application methods remain outside SQL analysis.
Every result position participates: a leading boolean or other data result cannot
hide a later database result. Native API assurance also checks the result inventory,
including pgconn readers, queued queries and preparation results. Preparation and
batch provenance rules still apply; result recognition grants no capability.

Aliases in parameters, results, fields and returned handles resolve in their
declaring scopes, including cross-file/imported chains. All build alternatives
must agree. Alias cycles, missing repository declarations and conflicting
signatures cannot establish a SQL-free disposition. A new defined type keeps
its identity and declared methods; deriving a result from a database result
does not grant a native port signature. Use the admitted result type or a true
alias. Ordinary container results and non-SQL function fields remain data.
The backend check includes generated Go in SQL analysis and applies the same
SQL access policies to it; generated-source exclusions in other architecture
rules do not exclude SQL. Missing or linked generated inputs fail. The separately
versioned Reference Pack format module remains an explicit foreign-library
boundary despite sharing the repository namespace; its runtime dependency
boundary is independently checked by the backend gate.

| Surface | Disposition |
| --- | --- |
| pgx/pgxpool query and exec; pgconn Exec/ExecParams | Inspect SQL argument 1. |
| database/sql context / non-context query and exec | Inspect SQL argument 1 / 0. |
| pgx/pgconn Prepare | Inspect SQL argument 2; names must be finite identifiers or empty. |
| database/sql PrepareContext / Prepare | Inspect SQL argument 1 / 0 and establish local statement provenance. |
| pgx Batch.Queue / pgconn Batch.ExecParams | Inspect SQL argument 0; batch construction and sending stay local. |
| pgx CopyFrom | Inspect finite relation identifiers at argument 1. |
| Result consumption and queued result callbacks | SQL-free; their original query/preparation is inspected separately. |
| database/sql prepared execution | Only directly initialized local handles with qualified preparation provenance. |
| Named prepared execution, statement descriptions, pipelines, textual low-level COPY, raw protocol access | Rejected until separately qualified. |

Raw-protocol package values, message/client construction and pgconn raw connection
construction are rejected, including escaped factory functions. Passive protocol
references have no production admission; qualify a narrow use explicitly before
introducing one. Bulk-copy identifiers are literal `pgx.Identifier` composites
whose parts resolve to finite constants.

Argument positions exclude the receiver. Escaped execution methods and native
method expressions are rejected. Prepared/batch handle reassignment invalidates
provenance; borrowed or opaque batches and prepared handles cannot establish it.
This deliberately avoids general heap, alias and stateful protocol inference.
Reference Data retains immutable domain entries until it constructs each native
batch locally. Test-support adapters accept sealed queries/statements where raw
string forwarding has no explicit grant.

To qualify a new native surface, review its pinned type and complete signature,
SQL/relation argument position, result handles and state requirements. Update the
exact registry and authored exported-method inventory together. Add independent
positive/negative argument-position, returned-handle, escape and policy fixtures.
The existing source-inspection row compares every admitted native type's exported
methods and signatures against the pinned Go/driver APIs; signature drift fails.
Do not automatically refresh the inventory or turn an unknown method into a
SQL-free default. A stateful API needs its own bounded design before admission.

## Adding a provider or shared read capability

1. Define the source-owner implementation behind a narrow read port, preserving
   source ordering, errors, nil/empty semantics and the caller's transaction.
   Keep mutation, locking, HTTP and aggregate assembly dependencies separate.
2. Register the provider through the existing application-composed catalog and
   canonical machine descriptor projection. Add its implementation entry and
   contributed fact-reader entry packages to `projection-sql-policy.mjs`.
3. Assign dedicated read packages an exact provider default. Use declaration
   bindings in mixed-provider/storage packages. Every SQL-bearing declaration
   needs authority; every active provider needs actual source-side SQL coverage.
4. Use `shared_read` only for narrow shared readers. The closure preserves all
   receiving providers and checks every access against each one independently.
   The union of their permissions never authorizes the helper.
5. Update exact package SQL grants in `backend_module_boundaries.json` if needed.
   Provider table restrictions remain in declaration policy. Peer and descendant
   packages do not inherit grants; harmless file renames and splits do.
6. Qualify extraction, transitive/shared helpers, wrong tables, writes and locks,
   then run the affected owner behavior tests and production architecture gates.

The collector traverses every repository-local production import, including
build-constrained and generated Go. It rejects missing/empty packages, unreadable
inputs, linked paths/declarations and unclassified native source. Standard and
foreign imports have explicit exact admissions; those external implementations
are trust boundaries, not repository source. New imports require review. An
already-imported sibling package is scanned just like a newly imported helper.

Timeline's `sourcerepository` provides unlocked facts; `sourcelocking` retains
owner-only locking and shares immutable selection/scanning semantics. Evidence
Timeline facts and Links facts have separate read packages. Records envelope
reads live in `records/enveloperead`. The DB interface lives in
`postgres/dbport`, so a port import does not pull connection/runtime plumbing
into a provider dependency graph.

## Maintenance and transport capabilities

`internal/platform/maintenancequery` is generated through `make generate` from
the canonical Recovery relation catalog and ownership manifest. Its finite
switches remain ordinary checked executable source. It provides admitted reads,
mutable-row deletion, protected-control locking, authoritative restore insertion
and aggregate truncation/locking of exact complete relation sets. Unknown,
partial, duplicate or ineligible relation sets fail. Derived projection tables
cannot be inserted through the authoritative restore capability.

Only explicitly admitted Recovery/assembly/reset packages may import this
maintenance capability. A provider importing it receives its full dependency
closure and fails on reachable write/lock authority. This is physical maintenance
under existing fencing/transaction contracts, not a source-owner permission or
provider escape hatch. Change the authored catalog and regenerate when the
relation inventory changes; never edit the generated Go file.

Central policy v4 also admits narrowly named transparent database transports and
exact session-control commands. A transport may forward its SQL parameter
unchanged; concatenation or another construction is rejected. Controls retain
their complete normalized SQL shape, including multiple statements where
explicitly granted. Changed roles, extra commands and relocated declarations
fail. Test-support grants are distinct. Stale bindings and grants fail.

## Validation and migration

Use the test-operations skill, `make task-guide` and `make target-plan` to inspect
the current exact selection before running it. The focused surfaces are:

- `harness.command_surface.behavior.owner_boundaries` and the Go source-inspection
  row under `make test-slice OWNER=harness.command_surface`.
- `make lint CHECK=backend-boundary` and the SQL ownership row under
  `make test-slice OWNER=module.projections`.
- Changed owner query, fact-reader, source lifecycle and restore/reset behavior
  rows, including closed-fragment and maintenance-capability tests.
- Catalog/schema source checks; generation and generated-artifact validation
  when authored producers or inputs change. Documentation uses lint-markdown.

The [controlling tracker](../handoffs/test-consolidation-refactor-tracker.md)
records exact selections, retained roots, failures, corrective runs and final
readiness. There is no required automatic full suite, visual or release run.

The migration changes only internal interfaces and machine policy. The former
`postgres.DB`, aggregate projection fact readers, untyped contributed query
strings, filename-based grants at these boundaries and permissive source-owner
roles have no compatibility aliases. Host/Identity hydration binds UUID arrays
instead of constructing variable placeholder lists. Ordering, source semantics,
borrowed transactions and restore atomicity remain covered by targeted tests.
There is no public API, schema, stored-data or deployment migration.

The analysis is an architecture gate, not a PostgreSQL executor or proof of all
SQL semantics. Runtime function bodies remain governed by migration ownership
and behavioral tests. Unsupported SQL forms must be simplified or explicitly
qualified before admission; passing registration counts alone is not coverage.
