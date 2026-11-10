package sourcecheck

import (
	"strings"
	"testing"
)

func TestAnalyzeSQL(t *testing.T) {
	runSQLHelperReturnTests(t)
	runSQLBranchStateTests(t)
	runSQLAssignmentStateTests(t)
	runSQLLoopStateTests(t)
	t.Run("unknown evaluation state cannot substitute package constants for locals", func(t *testing.T) {
		for _, body := range []string{`label:{const table="users";tx.Query(ctx,"SELECT * FROM "+table)}`, `defer func(){label:{const table="users";tx.Query(ctx,"SELECT * FROM "+table)}}()`} {
			got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import "github.com/jackc/pgx/v5"
const table="artifacts"
func Read(tx pgx.Tx){` + body + `}`})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, d := range got["p.go"] {
				if d.Name == "Read" {
					for _, e := range d.SQL {
						found = found || len(e.Unresolved) > 0 || strings.Contains(e.Text, "users")
					}
				}
			}
			if !found {
				t.Errorf("local constant inherited package value: %#v", got)
			}
		}
	})

	t.Run("closure and escaped collections cannot establish SQL", func(t *testing.T) {
		for _, body := range []string{
			`tables:=[]string{"artifacts"};defer func(){for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}}();tables=[]string{input}`,
			`tables:=[]string{"artifacts"};defer func(){for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}}()`,
			`tables:=[]string{input};f:=func(){tables=[]string{"artifacts"}};for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)};_=f`,
			`tables:=[]string{"artifacts"};other:=tables;other[0]=input;for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`,
			`tables:=[]string{"artifacts"};mutate(tables);for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`,
			`tables:=[]string{"artifacts"};tables[0]=input;for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`,
			`for _,item:=range []struct{table string}{{"artifacts"}}{item.table=input;tx.Exec(ctx,"SELECT * FROM "+item.table)}`,
		} {
			got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import "github.com/jackc/pgx/v5"
func Read(tx pgx.Tx,input string){` + body + `}`})
			if err != nil {
				t.Fatal(err)
			}
			rejected := false
			for _, d := range got["p.go"] {
				for _, e := range d.SQL {
					rejected = rejected || len(e.Unresolved) > 0
				}
			}
			if !rejected {
				t.Errorf("collection supplied stale SQL: %s %#v", body, got)
			}
		}
		for _, body := range []string{
			`defer func(){for _,table:=range []string{"artifacts"}{tx.Exec(ctx,"SELECT * FROM "+table)}}()`,
			`tables:=[]string{"artifacts"};for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`,
			`tables:=[]string{input};defer func(){for _,value:=range tables{tx.Exec(ctx,"SELECT * FROM artifacts WHERE id=$1",value)}}()`,
		} {
			got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import "github.com/jackc/pgx/v5"
func Read(tx pgx.Tx,input string){` + body + `}`})
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range got["p.go"] {
				for _, e := range d.SQL {
					if len(e.Unresolved) > 0 {
						t.Errorf("local or non-SQL collection rejected: %s %#v", body, got)
					}
				}
			}
		}
	})

	t.Run("generic callable results and constraints cannot hide execution", func(t *testing.T) {
		for _, source := range []string{
			`type QueryFunc[T any] func(context.Context,string,...any)(T,error);func Read(input string){query:=QueryFunc[pgx.Rows](opaque());query(ctx,input)}`,
			`func Query[T any](ctx context.Context,q string)(T,error){panic("unused")};func Read(input string){query:=Query[pgx.Rows];query(ctx,input)}`,
			`type QueryFunc[T any] func(context.Context,string,...any)(T,error);func Read(query QueryFunc[pgx.Rows],input string){query(ctx,input)}`,
			`type QueryFunc[T any] func(context.Context,string,...any)(bool,T,error);func Read(query QueryFunc[pgx.Rows],input string){query(ctx,input)}`,
			`func Read[T any](query func(context.Context,string,...any)(T,error),input string){query(ctx,input)}`,
			`func Read[F ~func(context.Context,string,...any)(pgx.Rows,error)](query F,input string){query(ctx,input)}`,
			`type Constraint interface{~func(context.Context,string,...any)(pgx.Rows,error)};func Read[F Constraint](query F,input string){query(ctx,input)}`,
		} {
			got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import("context";"github.com/jackc/pgx/v5")
` + source})
			if err != nil {
				t.Fatal(err)
			}
			rejected := false
			for _, d := range got["p.go"] {
				for _, e := range d.SQL {
					rejected = rejected || e.Disposition == "rejected"
				}
			}
			if !rejected {
				t.Errorf("generic SQL callable disappeared: %s %#v", source, got)
			}
		}
		for _, source := range []string{
			`func Read[T any](callback func(context.Context)(T,error)){callback(ctx)}`,
			`func decode[T any](input string)(T,error){panic("unused")};func Read(){decode[string]("data")}`,
			`func decode[T any](input string)(T,error){panic("unused")};func Read(){decode("data")}`,
			`func Read[F ~func(string)string](callback F){callback("data")}`,
			`func Read[T comparable](value T){consume(value)}`,
			`type Constraint interface{~func(string)string};func Read[F Constraint](callback F){callback("data")}`,
		} {
			got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import "context"
` + source})
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range got["p.go"] {
				for _, e := range d.SQL {
					if e.Disposition == "rejected" {
						t.Errorf("generic data callback rejected: %s %#v", source, got)
					}
				}
			}
		}
	})

	t.Run("SQL admission rejects every dot import independently of usage", func(t *testing.T) {
		for _, pkg := range []string{"github.com/jackc/pgx/v5/pgproto3", "github.com/jackc/pgx/v5", "github.com/JochiRaider/cartulary/internal/platform/sqlshape", "github.com/JochiRaider/cartulary/internal/helper", "strings"} {
			for _, file := range []string{"internal/example/read.go", "internal/gen/example/read.go"} {
				for _, body := range []string{`message:=Query{String:input};_=message`, `buf,err:=(&Query{String:input}).Encode(nil);if err==nil {writer.Write(buf)}`, `factory:=NewFrontend;_=factory`, `_=input`} {
					source := `//go:build unreachable
// Code generated for fixture admission; DO NOT EDIT.
package example
import . "` + pkg + `"
func Read(input string){` + body + `}`
					if _, err := AnalyzeSQL(map[string]string{file: source}); err == nil || !strings.Contains(err.Error(), file+":4:8") || !strings.Contains(err.Error(), pkg) {
						t.Fatalf("missing located dot admission: %s %v", file, err)
					}
					imports, err := ParseImports(file, []byte(source))
					if err != nil || len(imports) != 1 || imports[0] != pkg {
						t.Fatalf("import-only behavior changed: %v %v", imports, err)
					}
				}
			}
		}
		for _, body := range []string{`message:=proto.Query{String:input};_=message`, `buf,err:=(&proto.Query{String:input}).Encode(nil);if err==nil{writer.Write(buf)}`, `factory:=proto.NewFrontend;_=factory`} {
			got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import proto "github.com/jackc/pgx/v5/pgproto3"
func Read(input string){` + body + `}`})
			if err != nil {
				t.Fatal(err)
			}
			rejected := false
			for _, d := range got["p.go"] {
				for _, e := range d.SQL {
					rejected = rejected || e.Disposition == "rejected"
				}
			}
			if !rejected {
				t.Fatal(got)
			}
		}
	})

	t.Run("closure SQL requires stable captures", func(t *testing.T) {
		for _, body := range []string{
			`q:="SELECT * FROM artifacts";defer func(){tx.Exec(ctx,q)}();q=input`,
			`q:=shape.Text(shape.Literal("SELECT * FROM artifacts"));defer func(){tx.Exec(ctx,q)}();q=input`,
			`q:="SELECT * FROM artifacts";f:=func(){tx.Exec(ctx,q)};q=input;f()`,
			`q:="SELECT * FROM artifacts";go func(){tx.Exec(ctx,q)}();q=input`,
			`q:="SELECT * FROM artifacts";q="SELECT * FROM artifacts";func(){tx.Exec(ctx,q)}()`,
			`q:="SELECT * FROM artifacts";f:=func(){tx.Exec(ctx,q)};consume(f);q=input`,
			`q:="SELECT * FROM artifacts";defer func(){tx.Exec(ctx,q)}();if cond {q=input}`,
			`q:="SELECT * FROM artifacts";defer func(){tx.Exec(ctx,q)}();for cond {q=input}`,
			`q:="SELECT * FROM artifacts";p:=&q;defer func(){tx.Exec(ctx,q)}();*p=input`,
			`q:="SELECT * FROM artifacts";defer func(){tx.Exec(ctx,q)}();mutate(&q)`,
			`q:="SELECT * FROM artifacts";f:=func(){q=input};tx.Exec(ctx,q);f()`,
			`q:="SELECT * FROM artifacts";f:=func(){q=input};f();tx.Exec(ctx,q)`,
			`q:="SELECT * FROM artifacts";f:=func(){defer func(){q=input}()};f();tx.Exec(ctx,q)`,
			`defer func(){tx.Exec(ctx,input)}()`,
			`result="SELECT * FROM artifacts";defer func(){tx.Exec(ctx,result)}()`,
			`s.q="SELECT * FROM artifacts";defer func(){tx.Exec(ctx,s.q)}()`,
			`items[0]="SELECT * FROM artifacts";defer func(){tx.Exec(ctx,items[0])}()`,
			`q:="SELECT * FROM artifacts";p:=&q;tx.Exec(ctx,q);_=p`,
			`stmt,_:=db.Prepare("SELECT * FROM artifacts");defer func(){stmt.Exec()}()`,
			`var batch pgx.Batch;batch.Queue("SELECT * FROM artifacts");defer func(){tx.SendBatch(ctx,&batch)}()`,
		} {
			got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import("database/sql";"github.com/jackc/pgx/v5";shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape")
func Read(tx pgx.Tx,db *sql.DB,input string)(result string){` + body + `}`})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, e := range got["p.go"][0].SQL {
				found = found || e.Disposition == "rejected"
			}
			if !found {
				t.Errorf("unstable capture accepted: %s %#v", body, got)
			}
		}
		for _, body := range []string{
			`q:="SELECT * FROM artifacts";defer func(){tx.Exec(ctx,q)}()`,
			`const q="SELECT * FROM artifacts";go func(){tx.Exec(ctx,q)}()`,
			`q:="SELECT * FROM artifacts";f:=func(){tx.Exec(ctx,q)};consume(f)`,
			`q:="SELECT * FROM artifacts";copy:=q;q=input;defer func(){tx.Exec(ctx,copy)}()`,
			`q:="SELECT * FROM artifacts";{q:=input;consume(q)};defer func(){tx.Exec(ctx,q)}()`,
			`q:=input;defer func(){q:="SELECT * FROM artifacts";tx.Exec(ctx,q)}();q=input`,
			`q:="SELECT * FROM artifacts";defer func(){defer func(){tx.Exec(ctx,q)}()}()`,
			`defer func(){tx.Exec(ctx,"SELECT * FROM artifacts WHERE id=$1",input)}();input="other"`,
			`defer func(){stmt,_:=db.Prepare("SELECT * FROM artifacts");stmt.Exec()}()`,
			`defer func(){var batch pgx.Batch;batch.Queue("SELECT * FROM artifacts");tx.SendBatch(ctx,&batch)}()`,
			`q:=shape.Literal("SELECT * FROM artifacts");defer func(){tx.Exec(ctx,shape.Text(q))}()`,
		} {
			got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import("database/sql";"github.com/jackc/pgx/v5";shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape")
func Read(tx pgx.Tx,db *sql.DB,input string){` + body + `}`})
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range got["p.go"][0].SQL {
				if e.Disposition == "rejected" {
					t.Errorf("stable capture rejected: %s %#v", body, got)
				}
			}
		}
	})

	t.Run("global captures and unknown immutable copies remain unresolved", func(t *testing.T) {
		for _, source := range []string{
			`var q="SELECT * FROM artifacts";func Read(tx pgx.Tx){defer func(){tx.Exec(ctx,q)}()}`,
			`func Read(tx pgx.Tx,input string){q:=input;defer func(){tx.Exec(ctx,q)}()}`,
		} {
			got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import "github.com/jackc/pgx/v5"
` + source})
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
				t.Fatalf("unproven capture accepted: %#v", got)
			}
		}
		got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import "github.com/jackc/pgx/v5"
const q="SELECT * FROM artifacts"
func Read(tx pgx.Tx,input string){defer func(){tx.Exec(ctx,q)}();value:="SELECT * FROM artifacts";defer tx.Exec(ctx,value);value=input}`})
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range got["p.go"] {
			for _, e := range d.SQL {
				if len(e.Unresolved) > 0 {
					t.Fatal(got)
				}
			}
		}
	})

	t.Run("SQL callable values have canonical admission", func(t *testing.T) {
		for _, declaration := range []string{
			`type QueryFunc = func(context.Context,string,...any)(pgx.Rows,error)`,
			`type First = func(context.Context,string,...any)(pgx.Rows,error); type QueryFunc = First`,
			`type QueryFunc func(context.Context,string,...any)(pgx.Rows,error)`,
			`type QueryFunc = imported.QueryFunc`,
			`type QueryFunc = Missing`,
			`type QueryFunc = Other; type Other = QueryFunc`,
		} {
			for _, body := range []string{
				`query := store.Query; query(ctx,input)`, `store.Query(ctx,input)`,
				`callback(ctx,input)`, `(callback)(ctx,input)`, `copy := callback; _ = copy`,
				`consume(callback)`, `return callback`, `items[0](ctx,input)`,
				`value := items[0]; _ = value`, `factory()(ctx,input)`,
				`var value QueryFunc; value(ctx,input)`, `(*pointer)(ctx,input)`,
				`_,f:=store.Factory();f(ctx,input)`, `f:=store.Factory;_,query:=f();query(ctx,input)`,
			} {
				sources := map[string]string{
					"internal/example/read.go": `package example
import("context";"github.com/jackc/pgx/v5";imported "github.com/JochiRaider/cartulary/internal/example/callables")
` + declaration + `
type Store struct{ Query QueryFunc }
func (s Store) Factory()(bool,QueryFunc){return false,nil}
func factory() QueryFunc { return nil }
func Read(store Store,callback QueryFunc,items []QueryFunc,pointer *QueryFunc,input string){` + body + `}`,
					"internal/example/callables/types.go": `package callables
import("context";"github.com/jackc/pgx/v5")
type QueryFunc = func(context.Context,string,...any)(pgx.Rows,error)`,
				}
				got, err := AnalyzeSQL(sources)
				if err != nil {
					t.Fatal(err)
				}
				rejected := false
				for _, d := range got["internal/example/read.go"] {
					if d.Name == "Read" {
						for _, e := range d.SQL {
							rejected = rejected || e.Disposition == "rejected"
						}
					}
				}
				if !rejected {
					t.Errorf("SQL callable disappeared: %s / %s: %#v", declaration, body, got)
				}
			}
		}
		for _, source := range []string{
			`type Store struct{Query func(context.Context,string,...any)(pgx.Rows,error)};func Read(s Store){f:=s.Query;f(ctx,input)}`,
			`func Read(f func(context.Context,string,...any)(pgx.Rows,error)){f(ctx,input)}`,
			`func Read(){type Local func(context.Context,string,...any)(pgx.Rows,error);var f Local;f(ctx,input)}`,
		} {
			got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import("context";"github.com/jackc/pgx/v5")
` + source})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, d := range got["p.go"] {
				for _, e := range d.SQL {
					found = found || e.Disposition == "rejected"
				}
			}
			if !found {
				t.Fatalf("inline/local callable escaped: %s %#v", source, got)
			}
		}

		for _, body := range []string{`callback("data")`, `f:=store.Query;f("data")`, `factory()("data")`, `(store.Query)("data")`, `x:=foreign.String();consume(*x)`} {
			got, err := AnalyzeSQL(map[string]string{"p.go": `package p
type Callback func(string)string
type Store struct{Query Callback}
func factory() Callback {return nil}
func Read(callback Callback,store Store){` + body + `}`})
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range got["p.go"] {
				for _, e := range d.SQL {
					if e.Disposition == "rejected" {
						t.Fatalf("ordinary callback rejected: %s %#v", body, got)
					}
				}
			}
		}
	})

	t.Run("raw mutable builders cannot qualify SQL", func(t *testing.T) {
		for _, body := range []string{
			`var b strings.Builder; b.WriteString("SELECT * FROM records"); tx.Query(ctx,b.String())`,
			`var b strings.Builder; b.WriteString("--"); reset:=b.Reset; reset(); b.WriteString("DELETE FROM host_grid_projection"); tx.Exec(ctx,b.String())`,
			`var b strings.Builder; write:=b.WriteString; write("SELECT * FROM records"); tx.Query(ctx,b.String())`,
			`var b strings.Builder; b.WriteString("SELECT * FROM records"); text:=b.String; tx.Query(ctx,text())`,
			`var b strings.Builder; write:=(*strings.Builder).WriteString; write(&b,"SELECT * FROM records"); tx.Query(ctx,b.String())`,
			`var b strings.Builder; other:=&b; other.WriteString("SELECT * FROM records"); tx.Query(ctx,b.String())`,
			`var b strings.Builder; mutate:=func(){b.Reset()}; b.WriteString("SELECT * FROM records"); mutate(); tx.Query(ctx,b.String())`,
			`var b strings.Builder; defer b.Reset(); b.WriteString("SELECT * FROM records"); tx.Query(ctx,b.String())`,
			`var b strings.Builder; go b.Reset(); b.WriteString("SELECT * FROM records"); tx.Query(ctx,b.String())`,
			`var b strings.Builder; fmt.Fprintf(&b,"SELECT * FROM %s","records"); tx.Query(ctx,b.String())`,
		} {
			t.Run(body, func(t *testing.T) {
				got, err := AnalyzeSQL(map[string]string{"p.go": "package p\nimport (\"strings\";\"fmt\";\"github.com/jackc/pgx/v5\")\nfunc Read(tx pgx.Tx){" + body + "}"})
				if err != nil {
					t.Fatal(err)
				}
				rejected := false
				for _, expression := range got["p.go"][0].SQL {
					rejected = rejected || expression.Disposition == "rejected" && len(expression.Unresolved) > 0
				}
				if !rejected {
					t.Fatalf("raw builder SQL accepted: %#v", got)
				}
			})
		}
		for _, body := range []string{
			`var b strings.Builder; reset:=b.Reset; reset(); b.WriteString("ordinary text"); _=b.String()`,
			`var b strings.Builder; b.WriteString(input); tx.Query(ctx,"SELECT * FROM records WHERE id=$1",b.String())`,
		} {
			got, err := AnalyzeSQL(map[string]string{"p.go": "package p\nimport (\"strings\";\"github.com/jackc/pgx/v5\")\nfunc Read(tx pgx.Tx,input string){" + body + "}"})
			if err != nil {
				t.Fatal(err)
			}
			for _, expression := range got["p.go"][0].SQL {
				if len(expression.Unresolved) > 0 {
					t.Fatalf("non-SQL builder rejected: %#v", expression)
				}
			}
		}
	})
	t.Run("empty execution is accounted separately from unresolved input", func(t *testing.T) {
		got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import "github.com/jackc/pgx/v5"
func Read(tx pgx.Tx){tx.Exec(ctx,"");tx.Exec(ctx,opaque())}`})
		if err != nil {
			t.Fatal(err)
		}
		expressions := got["p.go"][0].SQL
		if len(expressions) != 2 || expressions[0].Kind != "empty" || expressions[1].Disposition != "rejected" {
			t.Fatal(got)
		}
		for _, e := range expressions {
			if e.Line < 1 || e.Column < 1 {
				t.Fatal(e)
			}
		}
	})
	t.Run("native SQL argument positions and unsupported paths", func(t *testing.T) {
		for _, tc := range []struct{ typ, body string }{
			{"*pgconn.PgConn", `handle.ExecParams(ctx,opaque(),nil,nil,nil,nil)`},
			{"*sql.DB", `handle.Query(opaque())`}, {"*sql.DB", `handle.QueryRow(opaque())`}, {"*sql.DB", `handle.Exec(opaque())`},
			{"*sql.DB", `handle.Prepare(opaque())`}, {"*sql.DB", `handle.PrepareContext(ctx,opaque())`},
			{"*sql.Conn", `handle.QueryContext(ctx,opaque())`},
			{"*pgconn.Batch", `handle.ExecParams(opaque(),nil,nil,nil,nil)`},
			{"*pgconn.PgConn", `handle.Prepare(ctx,"name",opaque(),nil)`},
			{"*pgconn.PgConn", `handle.ExecPrepared(ctx,"name",nil,nil,nil)`},
			{"*pgconn.PgConn", `handle.CopyFrom(ctx,reader,opaque())`},
			{"*pgconn.PgConn", `handle.StartPipeline(ctx)`},
			{"*pgconn.PgConn", `handle.Frontend()`},
			{"*pgconn.PgConn", `handle.Conn()`},
			{"*sql.Conn", `handle.Raw(callback)`},
			{"*sql.Stmt", `handle.Query(values)`},
			{"*sql.DB", `execute:=handle.Query;execute(opaque())`},
		} {
			got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import("database/sql";"github.com/jackc/pgx/v5/pgconn")
func Read(handle ` + tc.typ + `){` + tc.body + `}`})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, e := range got["p.go"][0].SQL {
				found = found || len(e.Unresolved) > 0
			}
			if !found {
				t.Errorf("unaccounted native call %s %s: %#v", tc.typ, tc.body, got)
			}
		}
	})
	t.Run("discovery is independent of assignment and control syntax", func(t *testing.T) {
		for _, body := range []string{
			`s.q = shape.Literal("SELECT * FROM users")`,
			`items[0] = shape.Literal("SELECT * FROM users")`,
			`ok := tx.QueryRow(ctx, opaque()).Scan(new(int)) == nil; _ = ok`,
			`switch tx.QueryRow(ctx, opaque()).Scan(new(int)) { default: }`,
			`switch mode { case tx.QueryRow(ctx, opaque()).Scan(new(int)): }`,
			`items[index(tx.QueryRow(ctx, opaque()))] = 1`,
			`var callback = func(){tx.Query(ctx, opaque())}; _ = callback`,
			`return; tx.Query(ctx, opaque())`,
			`return; s.q="SELECT * FROM users"`,
			`label: s.q="SELECT * FROM users"`,
			`ch <- tx.QueryRow(ctx, opaque())`,
			`defer tx.Exec(ctx,opaque())`,
			`go tx.Exec(ctx,opaque())`,
		} {
			got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import ("github.com/jackc/pgx/v5"; shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape")
func Read(tx pgx.Tx) {` + body + `}`})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, e := range got["p.go"][0].SQL {
				found = found || strings.Contains(e.Text, "users") || len(e.Unresolved) > 0
			}
			if !found {
				t.Errorf("SQL disappeared: %s %#v", body, got)
			}
		}
	})
	t.Run("driver execution boundaries retain declared return types", func(t *testing.T) {
		for _, body := range []string{
			`acquired,_:=pool.Acquire(ctx); acquired.Query(ctx,opaque())`,
			`borrowed,_:=pool.Begin(ctx); borrowed.Prepare(ctx,"statement",opaque())`,
			`var batch pgx.Batch; batch.Queue(opaque())`,
			`typed:=raw.(pgx.Tx); typed.Exec(ctx,opaque())`,
			`connection,_:=pgx.Connect(ctx,"dsn"); connection.PgConn().Exec(ctx,opaque())`,
			`tx.CopyFrom(ctx,pgx.Identifier{dynamic},columns,rows)`,
		} {
			got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import ("github.com/jackc/pgx/v5"; "github.com/jackc/pgx/v5/pgxpool")
func Read(pool *pgxpool.Pool,raw any,tx pgx.Tx){` + body + `}`})
			if err != nil {
				t.Fatal(err)
			}
			unresolved := false
			for _, e := range got["p.go"][0].SQL {
				unresolved = unresolved || len(e.Unresolved) > 0
			}
			if !unresolved {
				t.Fatalf("unsupported SQL disappeared: %s %#v", body, got)
			}
		}
		got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import "github.com/jackc/pgx/v5"
func Write(tx pgx.Tx){tx.CopyFrom(ctx,pgx.Identifier{"public","records"},columns,rows)}`})
		if err != nil {
			t.Fatal(err)
		}
		if got["p.go"][0].SQL[0].Kind != "copy" || got["p.go"][0].SQL[0].Text != `"public"."records"` {
			t.Fatalf("copy write relation missing: %#v", got)
		}
	})

	t.Run("imported immutable query preserves appended lock", func(t *testing.T) {
		got, err := AnalyzeSQL(map[string]string{
			"internal/source/select.go": "package source\nconst Select = `SELECT record_id FROM timeline_events`",
			"internal/owner/read.go":    "package owner\nimport definition \"github.com/JochiRaider/cartulary/internal/source\"\nfunc Lock(){tx.Query(ctx,definition.Select+` FOR UPDATE`)}",
		})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, d := range got["internal/owner/read.go"] {
			for _, q := range d.SQL {
				if strings.Contains(q.Text, "timeline_events FOR UPDATE") && len(q.Unresolved) == 0 {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("complete imported locking SQL missing: %#v", got)
		}
	})
	for _, tc := range []struct{ name, source, want string }{
		{"leading constant", "q := base + ` UNION SELECT user_id FROM users`; tx.Query(ctx,q)", "FROM users"},
		{"appended lock", "q := base; q += ` FOR UPDATE`; tx.Query(ctx,q)", "FOR UPDATE"},
		{"branch lock", "q := base; if lock { q += ` FOR NO KEY UPDATE` }; tx.Query(ctx,q)", "FOR NO KEY UPDATE"},
		{"switch write", "q := base; switch mode {case 1: q = `DELETE FROM records`; default: q += ` FOR SHARE`}; tx.Exec(ctx,q)", "DELETE FROM records"},
		{"shadowing", "q := base; {q := `SELECT * FROM users`; tx.Query(ctx,q)}; tx.Query(ctx,q)", "FROM users"},
		{"condition calls", "if tx.QueryRow(ctx,base+` FOR UPDATE`).Scan(&value)==nil {}", "FOR UPDATE"},
		{"format", "q := fmt.Sprintf(`%s FOR UPDATE`,base); tx.Query(ctx,q)", "FOR UPDATE"},
		{"literal percent", "q := fmt.Sprintf(`%s WHERE value LIKE '%%ok%%'`,base); tx.Query(ctx,q)", "LIKE '%ok%'"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := AnalyzeSQL(map[string]string{"p/constants.go": "package p\nconst ( base = `SELECT * FROM records`; other = base )", "p/renamed.go": "package p\nimport (\"fmt\"; \"strings\")\nfunc Read(){" + tc.source + "}"})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, d := range got["p/renamed.go"] {
				for _, e := range d.SQL {
					if strings.Contains(e.Text, tc.want) {
						found = true
					}
				}
			}
			if !found {
				t.Fatalf("missing %q: %#v", tc.want, got)
			}
		})
	}
	t.Run("unsupported formatting is visible without an execution sink", func(t *testing.T) {
		for _, expression := range []string{
			`fmt.Sprintf("SELECT * FROM %[2]s", "records", "users")`,
			`fmt.Sprintf("SELECT * FROM %10s", "users")`,
			`fmt.Sprintf("SELECT * FROM %q", "users")`,
			`fmt.Sprintf("SELECT * FROM %s", "records", "users")`,
			`fmt.Sprintf("SELECT * FROM %s")`,
		} {
			got, err := AnalyzeSQL(map[string]string{"p.go": "package p\nimport \"fmt\"\nvar query = " + expression})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, declaration := range got["p.go"] {
				for _, sql := range declaration.SQL {
					found = found || len(sql.Unresolved) > 0
				}
			}
			if !found {
				t.Fatalf("unsupported SQL disappeared: %s %#v", expression, got)
			}
		}
	})
	t.Run("finite helpers and collection alternatives", func(t *testing.T) {
		got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import "github.com/jackc/pgx/v5"
func query(mode int) string { switch mode {case 1:return "SELECT * FROM records";default:return "SELECT * FROM users"} }
func Read(tx pgx.Tx){for _, item:=range []struct{sql string}{{"SELECT * FROM hosts"},{"SELECT * FROM parties"}}{tx.Query(ctx,item.sql)};tx.Query(ctx,query(mode))}
`})
		if err != nil {
			t.Fatal(err)
		}
		for _, table := range []string{"records", "users", "hosts", "parties"} {
			found := false
			for _, expression := range got["p.go"][1].SQL {
				if strings.Contains(expression.Text, "FROM "+table) {
					found = true
				}
				if len(expression.Unresolved) > 0 {
					t.Fatal(expression)
				}
			}
			if !found {
				t.Fatalf("missing %s: %#v", table, got)
			}
		}
	})
	for _, body := range []string{
		`q:=unknown + " UNION SELECT * FROM users";_=q`,
		`q:=unknown; borrowed:=connection; borrowed.Query(ctx,q)`,
		`q:="SELECT * FROM records";switch value.(type){case string:q=opaque()};tx.Query(ctx,q)`,
		`q:="SELECT * FROM records";defer func(){q=opaque()}();tx.Query(ctx,q)`,
		`q:="SELECT * FROM records";switch mode {case 1:q+=" FOR UPDATE";fallthrough;case 2:q+=" NOWAIT"};tx.Query(ctx,q)`,
		`q:="SELECT * FROM records";unknown(&q);tx.Query(ctx,q)`,
		`var q strings.Builder;q.WriteString("SELECT * FROM records");unknown(&q);tx.Query(ctx,q.String())`,
		`var q strings.Builder;q.WriteString("SELECT * FROM records");other:=&q;other.WriteString(" FOR UPDATE");tx.Query(ctx,q.String())`,
		`q:=unreviewed(); borrowed:=connection; borrowed.Query(ctx,q)`,
		`fmt:=shadow;tx.Query(ctx,fmt.Sprintf("SELECT * FROM records"))`,
	} {
		t.Run("unsupported "+body, func(t *testing.T) {
			got, err := AnalyzeSQL(map[string]string{"p.go": "package p\nimport (\"strings\"; driver \"github.com/jackc/pgx/v5\")\nfunc Read(connection driver.Tx,tx driver.Tx){" + body + "}"})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, expression := range got["p.go"][0].SQL {
				if len(expression.Unresolved) > 0 {
					found = true
				}
			}
			if !found {
				t.Fatalf("unsupported SQL disappeared: %#v", got)
			}
		})
	}
	t.Run("database fields retain receiver identity", func(t *testing.T) {
		got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import renamed "github.com/jackc/pgx/v5"
type Reader struct{ connection renamed.Tx }
func(r *Reader) Read(){ other:=r.connection;other.Query(ctx,opaque()) }
`})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, e := range got["p.go"][0].SQL {
			if len(e.Unresolved) > 0 {
				found = true
			}
		}
		if !found {
			t.Fatal(got)
		}
	})

	t.Run("closed constructor qualification", func(t *testing.T) {
		prefix := `package p
import shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape"
import "github.com/jackc/pgx/v5"
`
		if _, err := AnalyzeSQL(map[string]string{"p.go": strings.Replace(prefix, "shape ", ". ", 1) + `var q=Literal(input)`}); err == nil {
			t.Fatal("dot import bypassed constructor checks")
		}
		if _, err := AnalyzeSQL(map[string]string{"p.go": prefix + `var constructor=shape.Literal`}); err == nil {
			t.Fatal("constructor escaped")
		}
		got, err := AnalyzeSQL(map[string]string{"p.go": prefix + `func Read(tx pgx.Tx){tx.Query(ctx,shape.Text(shape.Literal(input)))}`})
		if err != nil {
			t.Fatal(err)
		}
		if len(got["p.go"][0].SQL) != 2 || got["p.go"][0].SQL[0].Kind != "closed" || len(got["p.go"][0].SQL[1].Unresolved) == 0 {
			t.Fatal(got)
		}
	})

	t.Run("unresolved execution is explicit", func(t *testing.T) {
		got, err := AnalyzeSQL(map[string]string{"p.go": "package p\nimport \"github.com/jackc/pgx/v5\"\nfunc Read(tx pgx.Tx){q:=unreviewed(); tx.Query(ctx,q)}"})
		if err != nil {
			t.Fatal(err)
		}
		if len(got["p.go"]) != 1 || len(got["p.go"][0].SQL) != 1 || got["p.go"][0].SQL[0].Text != "" || len(got["p.go"][0].SQL[0].Unresolved) == 0 {
			t.Fatalf("unresolved SQL disappeared: %#v", got)
		}
	})
	t.Run("non SQL is not evidence", func(t *testing.T) {
		got, err := AnalyzeSQL(map[string]string{"p.go": "package p\nfunc Read(){note:=`hello`; _=note}"})
		if err != nil {
			t.Fatal(err)
		}
		if len(got["p.go"][0].SQL) != 0 {
			t.Fatal(got)
		}
	})
}

func TestSQLImportedDatabaseTypes(t *testing.T) {
	t.Run("pinned native API inventory", assertNativeSQLInventory)
	t.Run("native construction and controls", assertNativeSQLConstruction)
	t.Run("native build variants", assertNativeBuildVariants)
	t.Run("canonical signature identities", assertCanonicalSQLIdentities)
	for _, body := range []string{`type Wrapper struct{*pgx.Conn};func Read(w Wrapper){w.Query(ctx,opaque())}`, `type Wrapper struct{*pgx.Conn};func Read(w Wrapper){w.Conn.Query(ctx,opaque())}`} {
		got, err := AnalyzeSQL(map[string]string{"p.go": "package p\nimport \"github.com/jackc/pgx/v5\"\n" + body})
		if err != nil {
			t.Fatal(err)
		}
		unresolved := false
		for _, d := range got["p.go"] {
			for _, e := range d.SQL {
				unresolved = unresolved || len(e.Unresolved) > 0
			}
		}
		if !unresolved {
			t.Fatalf("embedded SQL port escaped: %#v", got)
		}
	}

	got, err := AnalyzeSQL(map[string]string{
		"internal/port/db.go": `package port
 import ("context"; driver "github.com/jackc/pgx/v5")
 type Reader interface{Query(context.Context,string,...any)(driver.Rows,error)}
 type Export struct{ Source Reader }
 type Alias = Export
 `,
		"internal/reader/read.go": `package reader
 import p "github.com/JochiRaider/cartulary/internal/port"
 func Read(input p.Alias){input.Source.Query(ctx,unknown())}
 `,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got["internal/reader/read.go"][0].SQL) != 1 || len(got["internal/reader/read.go"][0].SQL[0].Unresolved) == 0 {
		t.Fatalf("imported SQL sink disappeared: %#v", got)
	}
}
