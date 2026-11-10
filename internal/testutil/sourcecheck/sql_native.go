package sourcecheck

import (
	"fmt"
	"go/ast"
	"strings"
)

const pgxPackage = "github.com/jackc/pgx/v5"
const pgconnPackage = pgxPackage + "/pgconn"
const poolPackage = pgxPackage + "/pgxpool"

// Native admission is an exact type/method inventory. Argument indexes exclude
// the receiver. No receiver variable name or method-name-only grant exists.
type nativeSQLMethod struct {
	kind     string // sql, copy, handle, prepared, batch, none, reject
	argument int
	result   string
}

var nativeSQL = nativeSQLRegistry()

func nativeSQLRegistry() map[string]map[string]nativeSQLMethod {
	registry := map[string]map[string]nativeSQLMethod{}
	add := func(typ, names, kind string, arg int, result string) {
		if registry[typ] == nil {
			registry[typ] = map[string]nativeSQLMethod{}
		}
		for _, name := range strings.Fields(names) {
			registry[typ][name] = nativeSQLMethod{kind, arg, result}
		}
	}
	for _, typ := range []string{pgxPackage + ".Conn", pgxPackage + ".Tx", poolPackage + ".Pool", poolPackage + ".Conn", poolPackage + ".Tx"} {
		add(typ, "Exec Query QueryRow", "sql", 1, "")
		add(typ, "CopyFrom", "copy", 1, "")
		add(typ, "SendBatch", "batch", 1, pgxPackage+".BatchResults")
		add(typ, "Begin", "handle", -1, pgxPackage+".Tx")
	}
	for _, typ := range []string{pgxPackage + ".Conn", poolPackage + ".Pool", poolPackage + ".Conn"} {
		add(typ, "BeginTx", "handle", -1, pgxPackage+".Tx")
	}
	for _, typ := range []string{pgxPackage + ".Conn", pgxPackage + ".Tx", poolPackage + ".Tx"} {
		add(typ, "Prepare", "sql", 2, pgconnPackage+".StatementDescription")
	}
	add(pgxPackage+".Conn", "PgConn", "handle", -1, pgconnPackage+".PgConn")
	add(pgxPackage+".Conn", "Close Config Deallocate DeallocateAll IsClosed Ping TypeMap WaitForNotification", "none", -1, "")
	add(pgxPackage+".Conn", "LoadType LoadTypes", "reject", -1, "")
	for _, typ := range []string{pgxPackage + ".Tx", poolPackage + ".Tx"} {
		add(typ, "Commit Rollback", "none", -1, "")
		add(typ, "Conn", "handle", -1, pgxPackage+".Conn")
		add(typ, "LargeObjects", "reject", -1, "")
	}
	add(poolPackage+".Pool", "Acquire", "handle", -1, poolPackage+".Conn")
	add(poolPackage+".Pool", "AcquireFunc", "none", -1, "") // callback parameters are independently inspected
	add(poolPackage+".Pool", "AcquireAllIdle", "reject", -1, "")
	add(poolPackage+".Pool", "Close Config Ping Reset Stat", "none", -1, "")
	add(poolPackage+".Conn", "Conn Hijack", "handle", -1, pgxPackage+".Conn")
	add(poolPackage+".Conn", "Release Ping", "none", -1, "")
	add(pgxPackage+".Batch", "Queue", "sql", 0, pgxPackage+".QueuedQuery")
	add(pgxPackage+".Batch", "Len", "none", -1, "")
	add(pgxPackage+".QueuedQuery", "Query QueryRow Exec", "none", -1, "")
	add(pgxPackage+".BatchResults", "Query QueryRow Exec Close", "none", -1, "")
	add(pgconnPackage+".PgConn", "Exec ExecParams", "sql", 1, "")
	add(pgconnPackage+".PgConn", "Prepare", "sql", 2, pgconnPackage+".StatementDescription")
	add(pgconnPackage+".PgConn", "ExecBatch", "batch", 1, "")
	add(pgconnPackage+".PgConn", "ExecPrepared ExecStatement CopyFrom CopyTo StartPipeline Conn Frontend Hijack ReceiveMessage", "reject", -1, "")
	add(pgconnPackage+".PgConn", "PID TxStatus SecretKey Close CleanupDone IsClosed IsBusy ParameterStatus Deallocate CancelRequest WaitForNotification EscapeString CheckConn Ping SyncConn CustomData", "none", -1, "")
	add(pgconnPackage+".Batch", "ExecParams", "sql", 0, "")
	add(pgconnPackage+".Batch", "ExecPrepared ExecStatement", "reject", -1, "")
	add(pgconnPackage+".Pipeline", "SendPrepare SendDeallocate SendQueryParams SendQueryPrepared SendQueryStatement SendFlushRequest SendPipelineSync Flush Sync GetResults Close", "reject", -1, "")
	for _, typ := range []string{"database/sql.DB", "database/sql.Tx", "database/sql.Conn"} {
		add(typ, "ExecContext QueryContext QueryRowContext", "sql", 1, "")
		add(typ, "PrepareContext", "sql", 1, "database/sql.Stmt")
	}
	for _, typ := range []string{"database/sql.DB", "database/sql.Tx"} {
		add(typ, "Exec Query QueryRow Prepare", "sql", 0, "")
		add(typ, "Prepare", "sql", 0, "database/sql.Stmt")
	}
	add("database/sql.DB", "Begin BeginTx", "handle", -1, "database/sql.Tx")
	add("database/sql.DB", "Conn", "handle", -1, "database/sql.Conn")
	add("database/sql.DB", "Close Ping PingContext SetConnMaxIdleTime SetConnMaxLifetime SetMaxIdleConns SetMaxOpenConns Stats", "none", -1, "")
	add("database/sql.DB", "Driver", "reject", -1, "")
	add("database/sql.Conn", "BeginTx", "handle", -1, "database/sql.Tx")
	add("database/sql.Conn", "PingContext Close", "none", -1, "")
	add("database/sql.Conn", "Raw", "reject", -1, "")
	add("database/sql.Tx", "Commit Rollback", "none", -1, "")
	add("database/sql.Tx", "Stmt StmtContext", "reject", -1, "")
	add("database/sql.Stmt", "Exec ExecContext Query QueryContext QueryRow QueryRowContext", "prepared", -1, "")
	add("database/sql.Stmt", "Close", "none", -1, "")
	return registry
}

func nativeTypeRef(key string) sqlTypeRef {
	i := strings.LastIndex(key, ".")
	if i < 0 {
		return sqlTypeRef{}
	}
	return qualifiedSQLType(key[:i], key[i+1:])
}

// Resolve aliases and embedded declarations without inferring arbitrary calls.
// Conflicting build variants produce a rejection, never a selected capability.
func (types sqlTypes) boundary(ref sqlTypeRef, method string, seen map[string]bool) (nativeSQLMethod, bool) {
	ref = types.resolveAliases(ref, map[string]bool{})
	if declaration := types.localType(ref); declaration != nil {
		key := fmt.Sprintf("local:%d", declaration.Pos())
		if seen[key] {
			return nativeSQLMethod{kind: "reject"}, true
		}
		seen[key] = true
		return types.boundary(types.structural(ref.with(declaration.Type), map[string]bool{}), method, seen)
	}
	switch e := ref.expr.(type) {
	case *ast.BadExpr:
		return nativeSQLMethod{kind: "reject"}, potentialSQLBoundary(method)
	case *ast.StarExpr:
		return types.boundary(ref.with(e.X), method, seen)
	case *ast.ParenExpr:
		return types.boundary(ref.with(e.X), method, seen)
	case *ast.InterfaceType:
		for _, field := range e.Methods.List {
			for _, name := range field.Names {
				if name.Name == method {
					if signature := types.structural(ref.with(field.Type), map[string]bool{}); signature.expr != nil {
						if _, ok := signature.expr.(*ast.FuncType); ok {
							return types.portBoundary(signature, method)
						}
					}
					return nativeSQLMethod{}, false
				}
			}
		}
		var candidates []nativeSQLMethod
		for _, field := range e.Methods.List {
			if len(field.Names) == 0 {
				if b, ok := types.boundary(ref.with(field.Type), method, cloneSQLSeen(seen)); ok {
					candidates = append(candidates, b)
				}
			}
		}
		return mergeSQLBoundaries(candidates)
	case *ast.StructType:
		for _, field := range e.Fields.List {
			for _, name := range field.Names {
				if name.Name == method {
					if signature := types.structural(ref.with(field.Type), map[string]bool{}); signature.expr != nil {
						if _, ok := signature.expr.(*ast.FuncType); ok {
							return types.portBoundary(signature, method)
						}
					}
					return nativeSQLMethod{}, false
				}
			}
		}
		var candidates []nativeSQLMethod
		for _, field := range e.Fields.List {
			if len(field.Names) == 0 {
				if b, ok := types.boundary(ref.with(field.Type), method, cloneSQLSeen(seen)); ok {
					candidates = append(candidates, b)
				}
			}
		}
		return mergeSQLBoundaries(candidates)
	}
	key := ref.key()
	if strings.HasPrefix(key, "database/sql/driver.") || strings.HasPrefix(key, pgxPackage+"/pgproto3.") {
		return nativeSQLMethod{kind: "reject"}, true
	}
	if entries, known := nativeSQL[key]; known {
		if entry, ok := entries[method]; ok {
			return entry, true
		}
		return nativeSQLMethod{kind: "reject"}, true
	}
	if key == "" || seen[key] {
		return nativeSQLMethod{}, false
	}
	seen[key] = true
	if definition := types[key]; definition != nil {
		if declarations, ok := definition.methods[method]; ok {
			candidates := []nativeSQLMethod{}
			nonSQL := false
			for _, declared := range declarations {
				if b, ok := types.portBoundary(declared, method); ok {
					candidates = append(candidates, b)
				} else {
					nonSQL = true
				}
			}
			if nonSQL && len(candidates) > 0 {
				return nativeSQLMethod{kind: "reject"}, true
			}
			return mergeSQLBoundaries(candidates)
		}
		var candidates []nativeSQLMethod
		missing := false
		for _, alternative := range definition.definitions {
			structural := types.structural(alternative, map[string]bool{})
			if b, ok := types.boundary(structural, method, cloneSQLSeen(seen)); ok {
				candidates = append(candidates, b)
			} else {
				missing = true
			}
		}
		if missing && len(candidates) > 0 {
			return nativeSQLMethod{kind: "reject"}, true
		}
		return mergeSQLBoundaries(candidates)
	}
	if foreignSQLFreeType(key) {
		return nativeSQLMethod{kind: "none"}, true
	}

	return nativeSQLMethod{}, false
}

// Potential native execution names with unresolved receivers are failures, not
// permissions inferred from spelling. Declared application/foreign methods can
// establish that a same-named call is outside the admitted database APIs.
func potentialSQLBoundary(name string) bool {
	switch name {
	case "Exec", "Query", "QueryRow", "ExecContext", "QueryContext", "QueryRowContext", "Prepare", "PrepareContext", "Queue", "CopyFrom", "CopyTo", "ExecParams", "ExecPrepared", "ExecStatement", "SendBatch", "ExecBatch", "StartPipeline", "SendPrepare", "SendQueryParams", "SendQueryPrepared", "SendQueryStatement", "Raw":
		return true
	}
	return false
}
func foreignSQLFreeType(key string) bool {
	return key != "" && !repositorySQLType(key) && !strings.HasPrefix(key, pgxPackage+".") && !strings.HasPrefix(key, pgxPackage+"/") && !strings.HasPrefix(key, "database/sql") && strings.Contains(key, ".")
}

func cloneSQLSeen(input map[string]bool) map[string]bool {
	out := map[string]bool{}
	for k, v := range input {
		out[k] = v
	}
	return out
}
func (types sqlTypes) nativeIdentity(ref sqlTypeRef, seen map[string]bool) string {
	ref = types.resolveAliases(ref, seen)
	if star, ok := ref.expr.(*ast.StarExpr); ok {
		return types.nativeIdentity(ref.with(star.X), seen)
	}
	key := ref.key()
	if _, ok := nativeSQL[key]; ok {
		return key
	}
	return ""
}

func (types sqlTypes) isBatch(ref sqlTypeRef) bool {
	key := types.nativeIdentity(ref, map[string]bool{})
	return key == pgxPackage+".Batch" || key == pgconnPackage+".Batch"
}
func mergeSQLBoundaries(candidates []nativeSQLMethod) (nativeSQLMethod, bool) {
	if len(candidates) == 0 {
		return nativeSQLMethod{}, false
	}
	for _, candidate := range candidates[1:] {
		if candidate != candidates[0] {
			return nativeSQLMethod{kind: "reject"}, true
		}
	}
	return candidates[0], true
}

// Repository ports qualify through an actual database result signature, not
// merely a method's spelling or a string parameter. Other application APIs do
// not become SQL boundaries because they happen to be called Query or Prepare.
func (types sqlTypes) portBoundary(ref sqlTypeRef, method string) (nativeSQLMethod, bool) {
	ref = types.resolveAliases(ref, map[string]bool{})
	if _, bad := ref.expr.(*ast.BadExpr); bad {
		return nativeSQLMethod{kind: "reject"}, true
	}
	fn, ok := ref.expr.(*ast.FuncType)
	if !ok {
		return nativeSQLMethod{kind: "reject"}, true
	}
	if fn.Results == nil || len(fn.Results.List) == 0 {
		return nativeSQLMethod{kind: "none"}, true
	}
	hasSQLResult := false
	for _, field := range fn.Results.List {
		switch types.portResult(ref.with(field.Type)) {
		case "reject":
			return nativeSQLMethod{kind: "reject"}, true
		case "sql":
			hasSQLResult = true
		}
	}
	if !hasSQLResult {
		return nativeSQLMethod{kind: "none"}, true
	}
	position, argument := 0, -1
	if fn.Params == nil {
		return nativeSQLMethod{kind: "reject"}, true
	}
	for _, field := range fn.Params.List {
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		parameter, known := types.typeIdentity(ref.with(field.Type), map[string]bool{})
		if !known {
			return nativeSQLMethod{kind: "reject"}, true
		}
		for i := 0; i < count; i++ {
			if parameter == "builtin.string" {
				if argument >= 0 {
					return nativeSQLMethod{kind: "reject"}, true
				}
				argument = position
			}
			position++
		}
	}
	if argument < 0 {
		return nativeSQLMethod{kind: "reject"}, true
	}
	return nativeSQLMethod{kind: "sql", argument: argument}, true
}

// Protocol packages are not SQL-free libraries. Reject direct construction and
// escaped package values as well as receiver methods; otherwise a factory with
// an unqualified result type could hide Send/Flush behind an unknown handle.
func (a *sqlAnalyzer) rejectedNativeValue(selector *ast.SelectorExpr) bool {
	id, ok := selector.X.(*ast.Ident)
	if !ok || id.Obj != nil {
		return false
	}
	imported := a.imports[id.Name]
	return imported == pgxPackage+"/pgproto3" || imported == pgconnPackage && selector.Sel.Name == "Construct"
}
func (types sqlTypes) rawProtocolType(ref sqlTypeRef, seen map[string]bool) bool {
	switch e := ref.expr.(type) {
	case *ast.StarExpr:
		return types.rawProtocolType(ref.with(e.X), seen)
	case *ast.ParenExpr:
		return types.rawProtocolType(ref.with(e.X), seen)
	case *ast.Ident:
		if e.Obj != nil {
			if declaration, ok := e.Obj.Decl.(*ast.TypeSpec); ok {
				key := fmt.Sprintf("raw:%d", declaration.Pos())
				if seen[key] {
					return false
				}
				seen[key] = true
				return types.rawProtocolType(ref.with(declaration.Type), seen)
			}
		}
	}
	key := ref.key()
	if strings.HasPrefix(key, pgxPackage+"/pgproto3.") {
		return true
	}
	if key == "" || seen[key] {
		return false
	}
	seen[key] = true
	if definition := types[key]; definition != nil {
		for _, candidate := range definition.definitions {
			if types.rawProtocolType(candidate, cloneSQLSeen(seen)) {
				return true
			}
		}
	}
	return false
}
