package sourcecheck

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// This inventory is an authored admission checkpoint for the pinned APIs. New
// methods or changed signatures require reviewing nativeSQLRegistry, not an
// automatic refresh. No source/specification prose is read by this assurance.
var qualifiedNativeInventory = map[string]string{
	"database/sql.Conn": `BeginTx:func(context.Context, *sql.TxOptions) (*sql.Tx, error)
Close:func() error
ExecContext:func(context.Context, string, ...interface {}) (sql.Result, error)
PingContext:func(context.Context) error
PrepareContext:func(context.Context, string) (*sql.Stmt, error)
QueryContext:func(context.Context, string, ...interface {}) (*sql.Rows, error)
QueryRowContext:func(context.Context, string, ...interface {}) *sql.Row
Raw:func(func(interface {}) error) error`,
	"database/sql.DB": `Begin:func() (*sql.Tx, error)
BeginTx:func(context.Context, *sql.TxOptions) (*sql.Tx, error)
Close:func() error
Conn:func(context.Context) (*sql.Conn, error)
Driver:func() driver.Driver
Exec:func(string, ...interface {}) (sql.Result, error)
ExecContext:func(context.Context, string, ...interface {}) (sql.Result, error)
Ping:func() error
PingContext:func(context.Context) error
Prepare:func(string) (*sql.Stmt, error)
PrepareContext:func(context.Context, string) (*sql.Stmt, error)
Query:func(string, ...interface {}) (*sql.Rows, error)
QueryContext:func(context.Context, string, ...interface {}) (*sql.Rows, error)
QueryRow:func(string, ...interface {}) *sql.Row
QueryRowContext:func(context.Context, string, ...interface {}) *sql.Row
SetConnMaxIdleTime:func(time.Duration)
SetConnMaxLifetime:func(time.Duration)
SetMaxIdleConns:func(int)
SetMaxOpenConns:func(int)
Stats:func() sql.DBStats`,
	"database/sql.Stmt": `Close:func() error
Exec:func(...interface {}) (sql.Result, error)
ExecContext:func(context.Context, ...interface {}) (sql.Result, error)
Query:func(...interface {}) (*sql.Rows, error)
QueryContext:func(context.Context, ...interface {}) (*sql.Rows, error)
QueryRow:func(...interface {}) *sql.Row
QueryRowContext:func(context.Context, ...interface {}) *sql.Row`,
	"database/sql.Tx": `Commit:func() error
Exec:func(string, ...interface {}) (sql.Result, error)
ExecContext:func(context.Context, string, ...interface {}) (sql.Result, error)
Prepare:func(string) (*sql.Stmt, error)
PrepareContext:func(context.Context, string) (*sql.Stmt, error)
Query:func(string, ...interface {}) (*sql.Rows, error)
QueryContext:func(context.Context, string, ...interface {}) (*sql.Rows, error)
QueryRow:func(string, ...interface {}) *sql.Row
QueryRowContext:func(context.Context, string, ...interface {}) *sql.Row
Rollback:func() error
Stmt:func(*sql.Stmt) *sql.Stmt
StmtContext:func(context.Context, *sql.Stmt) *sql.Stmt`,
	"github.com/jackc/pgx/v5.Batch": `Len:func() int
Queue:func(string, ...interface {}) *pgx.QueuedQuery`,
	"github.com/jackc/pgx/v5.BatchResults": `Close:func() error
Exec:func() (pgconn.CommandTag, error)
Query:func() (pgx.Rows, error)
QueryRow:func() pgx.Row`,
	"github.com/jackc/pgx/v5.Conn": `Begin:func(context.Context) (pgx.Tx, error)
BeginTx:func(context.Context, pgx.TxOptions) (pgx.Tx, error)
Close:func(context.Context) error
Config:func() *pgx.ConnConfig
CopyFrom:func(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error)
Deallocate:func(context.Context, string) error
DeallocateAll:func(context.Context) error
Exec:func(context.Context, string, ...interface {}) (pgconn.CommandTag, error)
IsClosed:func() bool
LoadType:func(context.Context, string) (*pgtype.Type, error)
LoadTypes:func(context.Context, []string) ([]*pgtype.Type, error)
PgConn:func() *pgconn.PgConn
Ping:func(context.Context) error
Prepare:func(context.Context, string, string) (*pgconn.StatementDescription, error)
Query:func(context.Context, string, ...interface {}) (pgx.Rows, error)
QueryRow:func(context.Context, string, ...interface {}) pgx.Row
SendBatch:func(context.Context, *pgx.Batch) pgx.BatchResults
TypeMap:func() *pgtype.Map
WaitForNotification:func(context.Context) (*pgconn.Notification, error)`,
	"github.com/jackc/pgx/v5.QueuedQuery": `Exec:func(func(pgconn.CommandTag) error)
Query:func(func(pgx.Rows) error)
QueryRow:func(func(pgx.Row) error)`,
	"github.com/jackc/pgx/v5.Tx": `Begin:func(context.Context) (pgx.Tx, error)
Commit:func(context.Context) error
Conn:func() *pgx.Conn
CopyFrom:func(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error)
Exec:func(context.Context, string, ...interface {}) (pgconn.CommandTag, error)
LargeObjects:func() pgx.LargeObjects
Prepare:func(context.Context, string, string) (*pgconn.StatementDescription, error)
Query:func(context.Context, string, ...interface {}) (pgx.Rows, error)
QueryRow:func(context.Context, string, ...interface {}) pgx.Row
Rollback:func(context.Context) error
SendBatch:func(context.Context, *pgx.Batch) pgx.BatchResults`,
	"github.com/jackc/pgx/v5/pgconn.Batch": `ExecParams:func(string, [][]uint8, []uint32, []int16, []int16)
ExecPrepared:func(string, [][]uint8, []int16, []int16)
ExecStatement:func(*pgconn.StatementDescription, [][]uint8, []int16, []int16)`,
	"github.com/jackc/pgx/v5/pgconn.PgConn": `CancelRequest:func(context.Context) error
CheckConn:func() error
CleanupDone:func() chan struct {}
Close:func(context.Context) error
Conn:func() net.Conn
CopyFrom:func(context.Context, io.Reader, string) (pgconn.CommandTag, error)
CopyTo:func(context.Context, io.Writer, string) (pgconn.CommandTag, error)
CustomData:func() map[string]interface {}
Deallocate:func(context.Context, string) error
EscapeString:func(string) (string, error)
Exec:func(context.Context, string) *pgconn.MultiResultReader
ExecBatch:func(context.Context, *pgconn.Batch) *pgconn.MultiResultReader
ExecParams:func(context.Context, string, [][]uint8, []uint32, []int16, []int16) *pgconn.ResultReader
ExecPrepared:func(context.Context, string, [][]uint8, []int16, []int16) *pgconn.ResultReader
ExecStatement:func(context.Context, *pgconn.StatementDescription, [][]uint8, []int16, []int16) *pgconn.ResultReader
Frontend:func() *pgproto3.Frontend
Hijack:func() (*pgconn.HijackedConn, error)
IsBusy:func() bool
IsClosed:func() bool
PID:func() uint32
ParameterStatus:func(string) string
Ping:func(context.Context) error
Prepare:func(context.Context, string, string, []uint32) (*pgconn.StatementDescription, error)
ReceiveMessage:func(context.Context) (pgproto3.BackendMessage, error)
SecretKey:func() []uint8
StartPipeline:func(context.Context) *pgconn.Pipeline
SyncConn:func(context.Context) error
TxStatus:func() uint8
WaitForNotification:func(context.Context) error`,
	"github.com/jackc/pgx/v5/pgconn.Pipeline": `Close:func() error
Flush:func() error
GetResults:func() (interface {}, error)
SendDeallocate:func(string)
SendFlushRequest:func()
SendPipelineSync:func()
SendPrepare:func(string, string, []uint32)
SendQueryParams:func(string, [][]uint8, []uint32, []int16, []int16)
SendQueryPrepared:func(string, [][]uint8, []int16, []int16)
SendQueryStatement:func(*pgconn.StatementDescription, [][]uint8, []int16, []int16)
Sync:func() error`,
	"github.com/jackc/pgx/v5/pgxpool.Conn": `Begin:func(context.Context) (pgx.Tx, error)
BeginTx:func(context.Context, pgx.TxOptions) (pgx.Tx, error)
Conn:func() *pgx.Conn
CopyFrom:func(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error)
Exec:func(context.Context, string, ...interface {}) (pgconn.CommandTag, error)
Hijack:func() *pgx.Conn
Ping:func(context.Context) error
Query:func(context.Context, string, ...interface {}) (pgx.Rows, error)
QueryRow:func(context.Context, string, ...interface {}) pgx.Row
Release:func()
SendBatch:func(context.Context, *pgx.Batch) pgx.BatchResults`,
	"github.com/jackc/pgx/v5/pgxpool.Pool": `Acquire:func(context.Context) (*pgxpool.Conn, error)
AcquireAllIdle:func(context.Context) []*pgxpool.Conn
AcquireFunc:func(context.Context, func(*pgxpool.Conn) error) error
Begin:func(context.Context) (pgx.Tx, error)
BeginTx:func(context.Context, pgx.TxOptions) (pgx.Tx, error)
Close:func()
Config:func() *pgxpool.Config
CopyFrom:func(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error)
Exec:func(context.Context, string, ...interface {}) (pgconn.CommandTag, error)
Ping:func(context.Context) error
Query:func(context.Context, string, ...interface {}) (pgx.Rows, error)
QueryRow:func(context.Context, string, ...interface {}) pgx.Row
Reset:func()
SendBatch:func(context.Context, *pgx.Batch) pgx.BatchResults
Stat:func() *pgxpool.Stat`,
	"github.com/jackc/pgx/v5/pgxpool.Tx": `Begin:func(context.Context) (pgx.Tx, error)
Commit:func(context.Context) error
Conn:func() *pgx.Conn
CopyFrom:func(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error)
Exec:func(context.Context, string, ...interface {}) (pgconn.CommandTag, error)
LargeObjects:func() pgx.LargeObjects
Prepare:func(context.Context, string, string) (*pgconn.StatementDescription, error)
Query:func(context.Context, string, ...interface {}) (pgx.Rows, error)
QueryRow:func(context.Context, string, ...interface {}) pgx.Row
Rollback:func(context.Context) error
SendBatch:func(context.Context, *pgx.Batch) pgx.BatchResults`,
}

func assertNativeSQLInventory(t *testing.T) {
	t.Helper()
	types := map[string]reflect.Type{
		pgxPackage + ".Conn": reflect.TypeOf((*pgx.Conn)(nil)), pgxPackage + ".Tx": reflect.TypeOf((*pgx.Tx)(nil)).Elem(),
		poolPackage + ".Pool": reflect.TypeOf((*pgxpool.Pool)(nil)), poolPackage + ".Conn": reflect.TypeOf((*pgxpool.Conn)(nil)), poolPackage + ".Tx": reflect.TypeOf((*pgxpool.Tx)(nil)),
		pgxPackage + ".Batch": reflect.TypeOf((*pgx.Batch)(nil)), pgxPackage + ".QueuedQuery": reflect.TypeOf((*pgx.QueuedQuery)(nil)), pgxPackage + ".BatchResults": reflect.TypeOf((*pgx.BatchResults)(nil)).Elem(),
		pgconnPackage + ".PgConn": reflect.TypeOf((*pgconn.PgConn)(nil)), pgconnPackage + ".Batch": reflect.TypeOf((*pgconn.Batch)(nil)), pgconnPackage + ".Pipeline": reflect.TypeOf((*pgconn.Pipeline)(nil)),
		"database/sql.DB": reflect.TypeOf((*sql.DB)(nil)), "database/sql.Conn": reflect.TypeOf((*sql.Conn)(nil)), "database/sql.Tx": reflect.TypeOf((*sql.Tx)(nil)), "database/sql.Stmt": reflect.TypeOf((*sql.Stmt)(nil)),
	}
	actual := map[string]string{}
	for key, typ := range types {
		methods := []string{}
		for i := 0; i < typ.NumMethod(); i++ {
			method := typ.Method(i)
			fn := method.Type
			offset := 0
			if typ.Kind() != reflect.Interface {
				offset = 1
			}
			in, out := []reflect.Type{}, []reflect.Type{}
			for j := offset; j < fn.NumIn(); j++ {
				in = append(in, fn.In(j))
			}
			for j := 0; j < fn.NumOut(); j++ {
				out = append(out, fn.Out(j))
			}
			signature := reflect.FuncOf(in, out, fn.IsVariadic()).String()
			methods = append(methods, method.Name+":"+signature)
			boundary, exists := nativeSQL[key][method.Name]
			if !exists {
				t.Errorf("unclassified native API %s.%s %s", key, method.Name, signature)
				continue
			}
			if boundary.kind == "sql" && (boundary.argument >= len(in) || in[boundary.argument].Kind() != reflect.String) {
				t.Errorf("incorrect SQL argument: %s.%s", key, method.Name)
			}
			if boundary.kind == "sql" {
				for _, result := range out {
					if result == reflect.TypeFor[error]() {
						continue
					}
					key := ""
					for result.Kind() == reflect.Pointer {
						key += "*"
						result = result.Elem()
					}
					key += result.PkgPath() + "." + result.Name()
					if !nativeSQLResult(key) {
						t.Errorf("unclassified native SQL result %s.%s: %s", key, method.Name, signature)
					}
				}
			}
			if boundary.kind == "copy" && in[boundary.argument] != reflect.TypeOf(pgx.Identifier{}) {
				t.Errorf("incorrect relation argument: %s.%s", key, method.Name)
			}
			if boundary.result != "" {
				result := out[0]
				if result.Kind() == reflect.Pointer {
					result = result.Elem()
				}
				if result.PkgPath()+"."+result.Name() != boundary.result {
					t.Errorf("incorrect handle result: %s.%s", key, method.Name)
				}
			}
		}
		for method := range nativeSQL[key] {
			if _, ok := typ.MethodByName(method); !ok {
				t.Errorf("stale native admission %s.%s", key, method)
			}
		}
		sort.Strings(methods)
		actual[key] = strings.Join(methods, "\n")
	}
	if !reflect.DeepEqual(actual, qualifiedNativeInventory) {
		encoded, _ := json.Marshal(actual)
		t.Errorf("native API inventory requires review: %s", encoded)
	}
}

func assertConnectionRetirement(t *testing.T) {
	const prefix = `package p
import("database/sql"; d "database/sql/driver")
type Connection = sql.Conn
type Value = any
type Failure = error
`
	for _, callback := range []string{
		`func(any) error { return d.ErrBadConn }`,
		`func(_ any) error { return (d.ErrBadConn) }`,
		`func(interface{}) error { return d.ErrBadConn }`,
		`func(_ Value) Failure { return d.ErrBadConn }`,
	} {
		source := prefix + `func Read(conn *Connection){conn.Raw(` + callback + `)}`
		got, err := AnalyzeSQL(map[string]string{"p.go": source})
		if err != nil {
			t.Fatal(err)
		}
		assertQualifiedSQLSites(t, source, got, qualifiedSQLSite{"conn.Raw(", "closed", ""})
		for _, declaration := range got["p.go"] {
			for _, expression := range declaration.SQL {
				if len(expression.Unresolved) != 0 {
					t.Fatalf("retirement rejected: %s: %#v", callback, expression)
				}
			}
		}
	}
	for _, body := range []string{
		`conn.Raw(func(any) error { return nil })`,
		`conn.Raw(func(value any) error { return d.ErrBadConn })`,
		`conn.Raw(func(any) (err error) { return d.ErrBadConn })`,
		`conn.Raw(func(any) error { defer work(); return d.ErrBadConn })`,
		`conn.Raw(func(any) error { work(); return d.ErrBadConn })`,
		`conn.Raw(func(any) error { return wrapped(d.ErrBadConn) })`,
		`conn.Raw(func(any) error { return d.ErrSkip })`,
		`conn.Raw(func(any) error { return })`,
		`conn.Raw(func(any) {})`,
		`conn.Raw(func() error { return d.ErrBadConn })`,
		`conn.Raw(func(a, b any) error { return d.ErrBadConn })`,
		`conn.Raw(func(string) error { return d.ErrBadConn })`,
		`conn.Raw(func(any) string { return d.ErrBadConn })`,
		`callback:=func(any) error { return d.ErrBadConn }; conn.Raw(callback)`,
		`raw:=conn.Raw; raw(func(any) error { return d.ErrBadConn })`,
		`(*sql.Conn).Raw(conn,func(any) error { return d.ErrBadConn })`,
		`d:=struct{ErrBadConn error}{}; conn.Raw(func(any) error { return d.ErrBadConn })`,
		`type any string; conn.Raw(func(any) error { return d.ErrBadConn })`,
		`type error string; conn.Raw(func(any) error { return d.ErrBadConn })`,
	} {
		got, err := AnalyzeSQL(map[string]string{"p.go": prefix + `func Read(conn *sql.Conn){` + body + `}`})
		if err != nil {
			t.Fatal(err)
		}
		rejected := false
		for _, declaration := range got["p.go"] {
			for _, expression := range declaration.SQL {
				rejected = rejected || len(expression.Unresolved) > 0
			}
		}
		if !rejected {
			t.Fatalf("raw access admitted: %s", body)
		}
	}
	source := prefix + `func Read(conn *sql.Conn){conn.Raw(func(any) error { conn.ExecContext(ctx,"DELETE FROM users"); return d.ErrBadConn })}`
	got, err := AnalyzeSQL(map[string]string{"p.go": source})
	if err != nil {
		t.Fatal(err)
	}
	assertQualifiedSQLSites(t, source, got, qualifiedSQLSite{"conn.ExecContext(", "execution", "DELETE FROM users"})
}

func assertNativeSQLConstruction(t *testing.T) {
	// Independent driver-surface matrix: a wrong argument index must fail the
	// concrete-query case as well as reject the unresolved-query case.
	cases := []struct{ typ, call string }{}
	add := func(types, calls []string) {
		for _, typ := range types {
			for _, call := range calls {
				cases = append(cases, struct{ typ, call string }{typ, call})
			}
		}
	}
	add([]string{"*pgx.Conn", "pgx.Tx", "*pgxpool.Pool", "*pgxpool.Conn", "*pgxpool.Tx"}, []string{`handle.Exec(ctx,QUERY)`, `handle.Query(ctx,QUERY)`, `handle.QueryRow(ctx,QUERY)`})
	add([]string{"*pgx.Conn", "pgx.Tx", "*pgxpool.Tx"}, []string{`handle.Prepare(ctx,"name",QUERY)`})
	add([]string{"*pgconn.PgConn"}, []string{`handle.Exec(ctx,QUERY)`, `handle.ExecParams(ctx,QUERY,nil,nil,nil,nil)`, `handle.Prepare(ctx,"name",QUERY,nil)`})
	add([]string{"*pgx.Batch"}, []string{`handle.Queue(QUERY)`})
	add([]string{"*pgconn.Batch"}, []string{`handle.ExecParams(QUERY,nil,nil,nil,nil)`})
	add([]string{"*sql.DB", "*sql.Tx"}, []string{`handle.Exec(QUERY)`, `handle.Query(QUERY)`, `handle.QueryRow(QUERY)`, `handle.Prepare(QUERY)`})
	add([]string{"*sql.DB", "*sql.Tx", "*sql.Conn"}, []string{`handle.ExecContext(ctx,QUERY)`, `handle.QueryContext(ctx,QUERY)`, `handle.QueryRowContext(ctx,QUERY)`, `handle.PrepareContext(ctx,QUERY)`})
	for _, tc := range cases {
		for _, query := range []string{`"SELECT * FROM users"`, `opaque()`} {
			body := strings.ReplaceAll(tc.call, "QUERY", query)
			got, err := AnalyzeSQL(map[string]string{"p.go": nativeFixtureImports + `func Read(handle ` + tc.typ + `){` + body + `}`})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, e := range got["p.go"][0].SQL {
				if e.Kind == "execution" && (query == `opaque()` && len(e.Unresolved) > 0 || query != `opaque()` && e.Text == "SELECT * FROM users" && len(e.Unresolved) == 0) {
					found = true
				}
			}
			if !found {
				t.Errorf("incorrect native boundary %s: %s %#v", tc.typ, body, got)
			}
		}
	}
	for _, typ := range []string{"*pgx.Conn", "pgx.Tx", "*pgxpool.Pool", "*pgxpool.Conn", "*pgxpool.Tx"} {
		for _, relation := range []string{`"public","users"`, `dynamic`} {
			got, err := AnalyzeSQL(map[string]string{"p.go": nativeFixtureImports + `func Read(handle ` + typ + `){handle.CopyFrom(ctx,pgx.Identifier{` + relation + `},columns,rows)}`})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, e := range got["p.go"][0].SQL {
				if relation == `dynamic` {
					found = found || len(e.Unresolved) > 0
				} else {
					found = found || e.Kind == "copy" && e.Text == `"public"."users"` && len(e.Unresolved) == 0
				}
			}
			if !found {
				t.Errorf("bulk relation boundary %s %s: %#v", typ, relation, got)
			}
		}
	}
	for _, call := range []string{`stmt.Exec(data)`, `stmt.Query(data)`, `stmt.QueryRow(data)`, `stmt.ExecContext(ctx,data)`, `stmt.QueryContext(ctx,data)`, `stmt.QueryRowContext(ctx,data)`} {
		got, err := AnalyzeSQL(map[string]string{"p.go": nativeFixtureImports + `func Read(db *sql.DB){stmt,_:=db.Prepare("SELECT 1");` + call + `}`})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, e := range got["p.go"][0].SQL {
			if len(e.Unresolved) > 0 {
				t.Errorf("bound value became SQL: %#v", e)
			}
			found = found || e.Kind == "closed"
		}
		if !found {
			t.Error("prepared execution missing", call)
		}
	}
	for _, group := range []struct{ typ, methods string }{
		{"*pgconn.PgConn", "ExecPrepared ExecStatement CopyFrom CopyTo StartPipeline Conn Frontend Hijack ReceiveMessage"},
		{"*pgconn.Batch", "ExecPrepared ExecStatement"},
		{"*pgconn.Pipeline", "SendPrepare SendDeallocate SendQueryParams SendQueryPrepared SendQueryStatement SendFlushRequest SendPipelineSync Flush Sync GetResults Close"},
		{"*pgx.Conn", "LoadType LoadTypes"}, {"pgx.Tx", "LargeObjects"}, {"*pgxpool.Tx", "LargeObjects"}, {"*pgxpool.Pool", "AcquireAllIdle"},
		{"*sql.DB", "Driver"}, {"*sql.Conn", "Raw"}, {"*sql.Tx", "Stmt StmtContext"},
	} {
		for _, method := range strings.Fields(group.methods) {
			got, err := AnalyzeSQL(map[string]string{"p.go": nativeFixtureImports + `func Read(handle ` + group.typ + `){handle.` + method + `(opaque)}`})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, e := range got["p.go"][0].SQL {
				found = found || strings.Contains(strings.Join(e.Unresolved, " "), "unsupported native SQL API")
			}
			if !found {
				t.Errorf("unsupported native API accepted: %s.%s", group.typ, method)
			}
		}
	}

	for _, tc := range []struct {
		body  string
		sites []qualifiedSQLSite
	}{
		{`stmt,_:=handle.Prepare("SELECT * FROM users");stmt.Query("data")`, []qualifiedSQLSite{{"handle.Prepare(", "execution", "SELECT * FROM users"}, {"stmt.Query(", "closed", ""}}},
		{`var batch pgx.Batch;batch.Queue("SELECT * FROM users");pool.SendBatch(ctx,&batch)`, []qualifiedSQLSite{{"batch.Queue(", "execution", "SELECT * FROM users"}, {"pool.SendBatch(", "closed", ""}}},
		{`batch:=&pgx.Batch{};batch.Queue("SELECT * FROM users");pool.SendBatch(ctx,batch)`, []qualifiedSQLSite{{"batch.Queue(", "execution", "SELECT * FROM users"}, {"pool.SendBatch(", "closed", ""}}},
		{`var batch pgconn.Batch;batch.ExecParams("SELECT * FROM users",nil,nil,nil,nil);conn.ExecBatch(ctx,&batch)`, []qualifiedSQLSite{{"batch.ExecParams(", "execution", "SELECT * FROM users"}, {"conn.ExecBatch(", "closed", ""}}},
	} {
		source := nativeFixtureImports + `func Read(handle *sql.DB,pool *pgxpool.Pool,conn *pgconn.PgConn){` + tc.body + `}`
		got, err := AnalyzeSQL(map[string]string{"p.go": source})
		if err != nil {
			t.Fatal(err)
		}
		assertQualifiedSQLSites(t, source, got, tc.sites...)
		for _, e := range got["p.go"][0].SQL {
			if len(e.Unresolved) > 0 {
				t.Errorf("qualified handle rejected %s: %#v", tc.body, e)
			}
		}
	}
	for _, body := range []string{
		`batch.QueuedQueries=opaque();pool.SendBatch(ctx,batch)`,
		`pool.SendBatch(ctx,batch)`,
	} {
		got, err := AnalyzeSQL(map[string]string{"p.go": nativeFixtureImports + `func Read(batch *pgx.Batch,pool *pgxpool.Pool){` + body + `}`})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, e := range got["p.go"][0].SQL {
			found = found || len(e.Unresolved) > 0
		}
		if !found {
			t.Errorf("opaque batch accepted: %s", body)
		}
	}
	for _, body := range []string{
		`type Alias = sql.DB;func Read(handle *Alias){handle.Query(opaque())}`,
		`func Read(){type Local=sql.DB;var handle *Local;handle.Query(opaque())}`,
		`var handle *sql.DB;func Read(){handle.Query(opaque())}`,
		`type Box[T any] struct{Value T};func Read(box Box[*sql.DB]){box.Value.Query(opaque())}`,
		`var handle,_=sql.Open("driver","dsn");func Read(){handle.Query(opaque())}`,
		`func acquire()(error,*sql.DB){return nil,nil};func Read(){_,handle:=acquire();handle.Query(opaque())}`,
		`func Read(handle *sql.DB){method:=(*sql.DB).Query;method(handle,opaque())}`,
		`func Read(handle *sql.DB){(*sql.DB).Query(handle,opaque())}`,
		`func Read(handle *pgx.Conn){handle.Prepare(ctx,"SELECT * FROM artifacts","SELECT * FROM users")}`,
		`func Read(){var batch pgx.Batch;opaque(&batch)}`,
		`func Read(db *sql.DB,borrowed *sql.Stmt){stmt,_:=db.Prepare("SELECT 1");if cond {stmt=borrowed}else{stmt,_=db.Prepare("SELECT 2")};stmt.Exec()}`,
		`func Read(db *sql.DB){stmt,_:=db.Prepare("SELECT 1");stmt=opaque();stmt.Exec()}`,
		`func Read(borrowed *pgx.Batch,tx pgx.Tx){var batch *pgx.Batch=borrowed;tx.SendBatch(ctx,batch)}`,
		`func Read(handle *sql.DB){value:=items[handle.Exec]}`,
		`type Box[T any] struct{Value T};func Read(box Box[*sql.DB]){method:=box.Value.Exec}`,
		`type Wrapper struct{Run func(string)(pgconn.CommandTag,error)};func Read(w Wrapper){method:=w.Run}`,

		`func Read(){handle,_:=sql.Open("driver","dsn");handle.Query(opaque())}`,
		`func Read(handle *sql.DB){borrowed,_:=handle.Conn(ctx);borrowed.QueryContext(ctx,opaque())}`,
	} {
		got, err := AnalyzeSQL(map[string]string{"p.go": nativeFixtureImports + body})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, d := range got["p.go"] {
			for _, e := range d.SQL {
				found = found || len(e.Unresolved) > 0
			}
		}
		if !found {
			t.Errorf("native identity lost: %s", body)
		}
	}
	for _, source := range []string{
		`package p
import "net/http"
func Read(request *http.Request){request.URL.Query()}`,
		`package p
import "net/url"
func Read(){parsed,_:=url.Parse(input);parsed.Query()}`,
		`package p
import "golang.org/x/sys/unix"
func Run(){unix.Exec(path,args,env)}`,
		`package p
type Participant interface{Prepare()error}
func Read(participants []Participant){for _,participant:=range participants {participant.Prepare()}}`,
	} {
		got, err := AnalyzeSQL(map[string]string{"p.go": source})
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range got["p.go"] {
			if len(d.SQL) > 0 {
				t.Errorf("non-SQL boundary: %s %#v", source, d)
			}
		}
	}
	for _, body := range []string{
		`client:=proto.NewFrontend(reader,writer);client.Send(message)`,
		`factory:=proto.NewFrontend;client:=factory(reader,writer);client.Send(message)`,
		`message:=proto.Query{String:input}`,
		`type Message=proto.Query;message:=Message{String:input}`,
		`conn:=pgconn.Construct(state)`,
		`factory:=pgconn.Construct;conn:=factory(state)`,
	} {
		got, err := AnalyzeSQL(map[string]string{"p.go": nativeFixtureImports + `import proto "github.com/jackc/pgx/v5/pgproto3";func Read(){` + body + `}`})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, d := range got["p.go"] {
			for _, e := range d.SQL {
				found = found || len(e.Unresolved) > 0
			}
		}
		if !found {
			t.Error("raw protocol construction escaped", body)
		}
	}
	gotTypes, err := AnalyzeSQL(map[string]string{
		"contracts/query.go": `package contracts;type Query struct{};type Exec struct{}`,
		"p.go":               `package p;import c "github.com/JochiRaider/cartulary/contracts";func Read() c.Query {var x c.Exec;return c.Query{}}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, declarations := range gotTypes {
		for _, d := range declarations {
			if len(d.SQL) > 0 {
				t.Fatal("type reference became SQL", d)
			}
		}
	}
	gotSwitch, err := AnalyzeSQL(map[string]string{"p.go": nativeFixtureImports + `func Read(input any){switch handle:=input.(type){case *sql.DB:handle.Query(opaque());case pgx.Tx:handle.Query(ctx,opaque())}}`})
	if err != nil {
		t.Fatal(err)
	}
	if len(gotSwitch["p.go"][0].SQL) != 2 {
		t.Fatal("type switch native boundaries", gotSwitch)
	}
	for _, e := range gotSwitch["p.go"][0].SQL {
		if strings.Contains(strings.Join(e.Unresolved, " "), "receiver") {
			t.Fatal("type switch lost native identity", e)
		}
	}
	got, err := AnalyzeSQL(map[string]string{"p.go": `package p
 type App struct{}
 func (App) Query(value string) string {return value}
 func (App) Exec(value string) error {return nil}
 func (App) Prepare(value string) error {return nil}
 func Read(db App){db.Query(opaque());db.Exec(opaque());db.Prepare(opaque())}`})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range got["p.go"] {
		if len(d.SQL) > 0 {
			t.Errorf("application methods became SQL boundaries: %#v", d)
		}
	}
}

const nativeFixtureImports = `package p
import("database/sql";"github.com/jackc/pgx/v5";"github.com/jackc/pgx/v5/pgconn";"github.com/jackc/pgx/v5/pgxpool")
`

func assertNativeBuildVariants(t *testing.T) {
	got, err := AnalyzeSQL(map[string]string{
		"p/linux.go": `package p
import("sync";"github.com/jackc/pgx/v5")
type Wrapper struct{mu sync.RWMutex;tx pgx.Tx}
func(w Wrapper) Read(){w.mu.RLock();w.tx.Query(ctx,opaque())}`,
		"p/other.go": `package p
type Wrapper struct{}
func(w Wrapper) Close()error{return nil}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	expressions := got["p/linux.go"][0].SQL
	if len(expressions) != 1 || len(expressions[0].Unresolved) == 0 {
		t.Fatal(got)
	}
	got, err = AnalyzeSQL(map[string]string{
		"p/a.go": nativeFixtureImports + `type Wrapper struct{db *sql.DB};func(w Wrapper) Read(){w.db.Query(opaque())}`,
		"p/b.go": nativeFixtureImports + `type Wrapper struct{db *pgx.Conn}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got["p/a.go"][0].SQL) == 0 || len(got["p/a.go"][0].SQL[0].Unresolved) == 0 {
		t.Fatal("conflicting native field selected", got)
	}
}

func assertCanonicalSQLIdentities(t *testing.T) {
	for _, decl := range []string{
		`type Box[T any] struct{Value T};type App interface{Query(string)Box[string];Exec(string)Box[int];Prepare(string)error}`,
		`type Result=library.ArchiveLimits;type App interface{Query(string)Result;Exec(string)error;Prepare(string)error}`,
	} {
		got, err := AnalyzeSQL(map[string]string{"p.go": nativeFixtureImports + `import library "github.com/JochiRaider/cartulary/referencepacks/format/envelope";` + decl + `;func Read(tx App){tx.Query(input);tx.Exec(input);tx.Prepare(input)}`})
		if err != nil {
			t.Fatal(err)
		}
		if len(got["p.go"][0].SQL) > 0 {
			t.Fatal("qualified data boundary misclassified", got)
		}
	}
	{
		got, err := AnalyzeSQL(map[string]string{
			"p/a.go": nativeFixtureImports + `type Rows=pgx.Rows;type Reader interface{Query(string)(Rows,error)};func Read(tx Reader){tx.Query("SELECT 1")}`,
			"p/b.go": nativeFixtureImports + `type Rows=pgx.Rows`,
		})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, e := range got["p/a.go"][0].SQL {
			if e.Kind == "execution" && e.Text == "SELECT 1" && len(e.Unresolved) == 0 {
				found = true
			}
		}
		if !found {
			t.Fatal("equivalent build aliases rejected", got)
		}
	}
	for _, decl := range []string{
		`type Rows pgx.Rows`, `type Rows interface{pgx.Rows}`, `type Tag pgconn.CommandTag;type Rows=Tag`,
	} {
		got, err := AnalyzeSQL(map[string]string{"p.go": nativeFixtureImports + decl + `;type Reader interface{Query(string)(Rows,error)};func Read(tx Reader){tx.Query(input)}`})
		if err != nil {
			t.Fatal(err)
		}
		if q := got["p.go"][0].SQL; len(q) == 0 || len(q[0].Unresolved) == 0 {
			t.Fatal("defined database result laundered", got)
		}
	}
	for _, decl := range []string{
		`type App interface{Query(string)[]unknown;Exec(string)[32]byte;Prepare(string)map[string]unknown}`,
		`type Handler func(string)string;type App struct{Query Handler;Exec Handler;Prepare Handler}`,
	} {
		got, err := AnalyzeSQL(map[string]string{"p.go": nativeFixtureImports + decl + `;func Read(tx App){tx.Query(input);tx.Exec(input);tx.Prepare(input)}`})
		if err != nil {
			t.Fatal(err)
		}
		if len(got["p.go"][0].SQL) > 0 {
			t.Fatal("data/function field misclassified", got)
		}
	}
	executions := func(expressions []SQLExpression) []SQLExpression {
		out := []SQLExpression{}
		for _, e := range expressions {
			if e.Kind != "statement" {
				out = append(out, e)
			}
		}
		return out
	}
	for _, declaration := range []string{
		`type Rows=pgx.Rows;type Text=string;type Reader interface{Query(context.Context,Text,...any)(Rows,error)}`,
		`type Rows=*pgconn.MultiResultReader;type Reader interface{Query(context.Context,string,...any)Rows}`,
		`type Rows=*pgconn.ResultReader;type Reader interface{Query(context.Context,string,...any)Rows}`,
		`type Rows=*pgx.QueuedQuery;type Reader interface{Query(context.Context,string,...any)Rows}`,
		`type Rows=pgx.Rows;type Reader interface{Query(context.Context,string,...any)(bool,Rows,error)}`,
		`type Rows=pgx.Rows;type Reader interface{Query(context.Context,string,...any)(err error, ok bool, rows Rows)}`,
		`type First=pgx.Rows;type Rows=First;type Reader interface{Query(context.Context,string,...any)(Rows,error)}`,
		`type Rows=sql.Rows;type Reader interface{Query(context.Context,string,...any)(*Rows,error)}`,
		`type Result=pgconn.CommandTag;type Reader interface{Query(context.Context,string,...any)(Result,error)}`,
		`type Reader interface{Query(context.Context,string,...any)(pgx.Rows,error)}`,
	} {
		for _, query := range []string{`"SELECT * FROM records"`, `input`} {
			source := nativeFixtureImports + `import "context";` + declaration + `
func Read(tx Reader,input string){tx.Query(ctx,` + query + `)}`
			got, err := AnalyzeSQL(map[string]string{"p.go": source})
			if err != nil {
				t.Fatal(err)
			}
			expressions := executions(got["p.go"][0].SQL)
			if len(expressions) != 1 {
				t.Fatalf("signature lost execution: %s %#v", declaration, got)
			}
			if query == `input` {
				if len(expressions[0].Unresolved) == 0 {
					t.Fatal("unresolved alias execution passed", got)
				}
			} else if expressions[0].Text != "SELECT * FROM records" || len(expressions[0].Unresolved) > 0 {
				t.Fatal("alias changed query", got)
			}
		}
	}
	for _, body := range []string{
		`type Outer interface{Reader};func Read(tx Outer){tx.Query(ctx,input)}`,
		`type Outer struct{DB Reader};func Read(tx Outer){tx.DB.Query(ctx,input)}`,
		`type Handle=pgx.Conn;func acquire()*Handle{return nil};func Read(){acquire().Query(ctx,input)}`,
		`func Read(){type Rows=pgx.Rows;type Local interface{Query(context.Context,string,...any)(Rows,error)};var tx Local;tx.Query(ctx,input)}`,
	} {
		got, err := AnalyzeSQL(map[string]string{"p.go": nativeFixtureImports + `import "context";type Rows=pgx.Rows;type Reader interface{Query(context.Context,string,...any)(Rows,error)};` + body})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, d := range got["p.go"] {
			if d.Name == "Read" {
				for _, e := range d.SQL {
					found = found || len(e.Unresolved) > 0
				}
			}
		}
		if !found {
			t.Fatal("alias handle escaped", body, got)
		}
	}
	base := `package p;import("context";"github.com/jackc/pgx/v5");type Reader interface{Query(context.Context,string,...any)(Rows,error)};func Read(tx Reader){tx.Query(ctx,input)}`
	for _, aliases := range []map[string]string{
		{"p/a.go": `package p;import "github.com/jackc/pgx/v5";type Rows=pgx.Rows`},
		{"p/a.go": `package p;import "github.com/jackc/pgx/v5";type Rows=pgx.Rows`, "p/b.go": `package p;type Rows=string`},
		{"p/a.go": `package p;type Rows=Other;type Other=Rows`},
		{"p/a.go": `package p;import "github.com/JochiRaider/cartulary/internal/missing";type Rows=missing.Rows`},
		{},
	} {
		aliases["p/read.go"] = base
		got, err := AnalyzeSQL(aliases)
		if err != nil {
			t.Fatal(err)
		}
		expressions := got["p/read.go"][0].SQL
		if len(expressions) == 0 || len(expressions[0].Unresolved) == 0 {
			t.Fatal("missing/conflicting alias accepted", got)
		}
	}
	got, err := AnalyzeSQL(map[string]string{
		"internal/port/types.go": `package port;import("context";"github.com/jackc/pgx/v5");type Rows=pgx.Rows;type Text=string;type Reader interface{Query(context.Context,Text,...any)(Rows,error)}`,
		"internal/use/read.go":   `package use;import p "github.com/JochiRaider/cartulary/internal/port";func Read(tx p.Reader){tx.Query(ctx,"SELECT * FROM records")}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if q := executions(got["internal/use/read.go"][0].SQL); len(q) != 1 || q[0].Text != "SELECT * FROM records" || len(q[0].Unresolved) > 0 {
		t.Fatal("imported alias lost signature", got)
	}
	for _, declaration := range []string{
		`type Rows=string;type App interface{Query(string)Rows;Exec(string)error;Prepare(string)error}`,
		`type App pgx.Conn;func(App)Query(string)string{return ""};func(App)Exec(string)error{return nil};func(App)Prepare(string)error{return nil}`,
	} {
		got, err := AnalyzeSQL(map[string]string{"p.go": nativeFixtureImports + declaration + `;func Read(tx App){tx.Query(input);tx.Exec(input);tx.Prepare(input)}`})
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range got["p.go"] {
			if len(d.SQL) > 0 {
				t.Fatal("non-SQL defined method became native", got)
			}
		}
	}
}
