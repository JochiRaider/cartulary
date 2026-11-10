package sourcecheck

import (
	"fmt"
	"strings"
	"testing"
)

// Assertions bind to the actual execution site, not an unrelated declaration
// containing the same SQL text.
func assertSQLStateSink(t *testing.T, body string, unresolved bool, texts ...string) {
	t.Helper()
	assertSQLStateSite(t, body, "tx.Exec(", unresolved, texts...)
}

func assertSQLStateSite(t *testing.T, body, site string, unresolved bool, texts ...string) {
	t.Helper()
	source := `package p
import ("github.com/jackc/pgx/v5"; "database/sql"; shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape")
func pair(input string)(bool,string){return true,input}
func triple(input string)(bool,bool,string){return true,true,input}
func lists(input string)(bool,[]string){return true,[]string{input}}
func Read(tx pgx.Tx, db *sql.DB, input string, flag bool){
` + body + "\n}"
	got, err := AnalyzeSQL(map[string]string{"p.go": source})
	if err != nil {
		t.Fatal(err)
	}
	position := strings.LastIndex(source, site)
	if position < 0 {
		t.Fatal("missing fixture execution site", site)
	}
	line := strings.Count(source[:position], "\n") + 1
	column := position - strings.LastIndex(source[:position], "\n")
	var sinks []SQLExpression
	for _, declaration := range got["p.go"] {
		if declaration.Name != "Read" {
			continue
		}
		for _, expression := range declaration.SQL {
			if expression.Line == line && expression.Column == column && (expression.Kind == "execution" || expression.Kind == "forward" || expression.Kind == "closed" || expression.Kind == "empty") {
				sinks = append(sinks, expression)
			}
		}
	}
	if len(sinks) == 0 {
		t.Fatalf("execution site %d:%d disappeared: %#v", line, column, got)
	}
	foundUnresolved := false
	for _, expression := range sinks {
		foundUnresolved = foundUnresolved || (expression.Disposition == "rejected" || expression.Disposition == "transport") && len(expression.Unresolved) > 0
	}
	if foundUnresolved != unresolved {
		t.Fatalf("unresolved=%v, want %v: %#v", foundUnresolved, unresolved, sinks)
	}
	for _, text := range texts {
		found := false
		for _, expression := range sinks {
			found = found || expression.Text == text
		}
		if !found {
			t.Errorf("missing execution alternative %q: %#v", text, sinks)
		}
	}
}

func runSQLBranchStateTests(t *testing.T) {
	t.Run("ordinary data retains lexical type context", func(t *testing.T) {
		got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import "errors"
type A struct{ID string};type B struct{Count int};type Provider struct{Query bool}
func Variant(input any){switch value:=input.(type){case A:if value.ID==""{return};case B:if value.Count==0{return}}}
func Named(flag bool)(count int,err error){if flag{count=1;err=errors.New("data")};return count,err}
func Runes(value string){for _,r:=range value{if r>='A' && r<='Z'{r+='a'-'A'};consume(r)}}
func Providers(items []Provider){for _,provider:=range items{if !provider.Query{return}}}`})
		if err != nil {
			t.Fatal(err)
		}
		for _, declarations := range got {
			for _, d := range declarations {
				for _, e := range d.SQL {
					if len(e.Unresolved) > 0 {
						t.Fatalf("ordinary data rejected in %s: %#v", d.Name, e)
					}
				}
			}
		}
	})
	t.Run("branch assignments retain declared port identity", func(t *testing.T) {
		got, err := AnalyzeSQL(map[string]string{"p.go": `package p
import ("context";"github.com/jackc/pgx/v5")
type Runner interface{Apply(context.Context,string,...any)(pgx.Rows,error)}
type A struct{};func(A)Apply(context.Context,string,...any)(pgx.Rows,error){return nil,nil}
type B struct{};func(B)Apply(context.Context,string,...any)(pgx.Rows,error){return nil,nil}
func Read(r Runner,input string,flag bool){if flag{r=A{}}else{r=B{}};r.Apply(ctx,input)}`})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, d := range got["p.go"] {
			if d.Name == "Read" {
				for _, e := range d.SQL {
					found = found || len(e.Unresolved) > 0 && e.Line == 6
				}
			}
		}
		if !found {
			t.Fatal("declared interface boundary disappeared", got)
		}
	})
	t.Run("branch collection execution alternatives", func(t *testing.T) {
		for _, branch := range []string{
			`if flag {tables=[]string{"users"}} else {tables=[]string{"artifacts"}}`,
			`if flag {tables=[]string{"artifacts"}} else {tables=[]string{"users"}}`,
			`switch {case flag:tables=[]string{"users"};default:tables=[]string{"artifacts"}}`,
			`switch {case flag:tables=[]string{"artifacts"};default:tables=[]string{"users"}}`,
			`if flag {tables=[]string{"users"}}`,
			`switch {case flag:tables=[]string{"users"}}`,
			`if flag {if flag{tables=[]string{"users","artifacts"}}} else {tables=[]string{}}`,
		} {
			t.Run(branch, func(t *testing.T) {
				assertSQLStateSink(t, `tables:=[]string{"artifacts"};`+branch+`;for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`, false, "SELECT * FROM artifacts", "SELECT * FROM users")
			})
		}
	})
	t.Run("collection construction snapshots and scope", func(t *testing.T) {
		for _, body := range []string{
			`table:="users";tables:=[]string{table};table="artifacts";for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`,
			`tables:=[]string{"users"};{tables:=[]string{"artifacts"};_=0};for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`,
			`tables:=[]string{"users"};if flag {tables=[]string{"artifacts"};return};for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`,
		} {
			assertSQLStateSink(t, body, false, "SELECT * FROM users")
		}
		assertSQLStateSink(t, `var tables=[]string{"artifacts"};for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`, false, "SELECT * FROM artifacts")
		assertSQLStateSink(t, `q:="SELECT * FROM artifacts";switch q:=any(input).(type){case string:_=q};tx.Exec(ctx,q)`, false, "SELECT * FROM artifacts")
		assertSQLStateSink(t, `tables:=[]string{};if flag{tables=[]string{"artifacts"}};for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`, false, "SELECT * FROM artifacts")
		assertSQLStateSink(t, `tables:=[]string{"artifacts"};if flag{tables=[]string{input}};for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`, true)
		assertSQLStateSink(t, `table:=input;tables:=[]string{table};table="artifacts";for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`, true)
		assertSQLStateSink(t, `tables:=[]struct{table string}{{"artifacts"}};if flag{tables=[]struct{table string}{{"users"},{"artifacts"}}};for _,item:=range tables{tx.Exec(ctx,"SELECT * FROM "+item.table)}`, false, "SELECT * FROM users", "SELECT * FROM artifacts")
		for _, count := range []int{64, 65} {
			body := fmt.Sprintf(`tables:=[]string{%s};for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`, strings.Repeat(`"artifacts",`, count))
			assertSQLStateSink(t, body, count > 64)
		}
	})
	t.Run("alternative limits reject instead of truncate", func(t *testing.T) {
		for _, count := range []int{256, 257} {
			body := `table:="t0";`
			for i := 1; i < count; i++ {
				body += fmt.Sprintf(`if flag{table="t%d"};`, i)
			}
			assertSQLStateSink(t, body+`tx.Exec(ctx,"SELECT * FROM "+table)`, count > 256)
		}
	})
	t.Run("branch provenance includes every continuing path", func(t *testing.T) {
		for _, branch := range []string{`if flag{stmt=opaque()}`, `if flag{}else{stmt=opaque()}`, `switch{case flag:stmt=opaque();default:}`, `switch{case flag:;default:stmt=opaque()}`} {
			assertSQLStateSite(t, `stmt,_:=db.Prepare("SELECT * FROM artifacts");`+branch+`;stmt.Exec()`, "stmt.Exec(", true)
		}
		assertSQLStateSite(t, `stmt,_:=db.Prepare("SELECT * FROM artifacts");if flag{stmt=opaque();return};stmt.Exec()`, "stmt.Exec(", false)
	})
	t.Run("collection helper rename and split preserve alternatives", func(t *testing.T) {
		helper := `func query(flag bool)string{tables:=[]string{"artifacts"};if flag{tables=[]string{"users"}};for _,table:=range tables{return "SELECT * FROM "+table};return "SELECT 1"}`
		read := `func Read(tx pgx.Tx,flag bool){tx.Exec(ctx,query(flag))}`
		for _, sources := range []map[string]string{
			{"p/original.go": "package p\nimport \"github.com/jackc/pgx/v5\"\n" + helper + "\n" + read},
			{"p/renamed.go": "package p\nimport \"github.com/jackc/pgx/v5\"\n" + read, "p/helper.go": "package p\n" + helper},
		} {
			got, err := AnalyzeSQL(sources)
			if err != nil {
				t.Fatal(err)
			}
			texts := map[string]bool{}
			for _, declarations := range got {
				for _, d := range declarations {
					if d.Name == "Read" {
						for _, e := range d.SQL {
							if e.Kind == "execution" {
								if e.Disposition != "analyzed" {
									t.Fatal(e)
								}
								texts[e.Text] = true
							}
						}
					}
				}
			}
			if !texts["SELECT * FROM users"] || !texts["SELECT * FROM artifacts"] {
				t.Fatal("lost helper alternative", got)
			}
		}
	})
}

func runSQLAssignmentStateTests(t *testing.T) {
	t.Run("every assignment result replaces previous SQL", func(t *testing.T) {
		for _, body := range []string{
			`q:="SELECT * FROM artifacts";_,q=pair(input);tx.Exec(ctx,q)`,
			`q:="SELECT * FROM artifacts";ok,q:=pair(input);_=ok;tx.Exec(ctx,q)`,
			`_,q:=pair(input);tx.Exec(ctx,q)`,
			`var _,q=pair(input);tx.Exec(ctx,q)`,
			`q:="SELECT * FROM artifacts";_,_,q=triple(input);tx.Exec(ctx,q)`,
			`tables:=[]string{"artifacts"};_,tables=lists(input);for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`,
			`q:="SELECT * FROM artifacts";if flag{_,q=pair(input)};tx.Exec(ctx,q)`,
		} {
			t.Run(body, func(t *testing.T) { assertSQLStateSink(t, body, true) })
		}
	})
	t.Run("parallel assignment uses incoming values", func(t *testing.T) {
		for _, body := range []string{
			`q,other:="SELECT * FROM artifacts","SELECT * FROM users";q,other=other,q;tx.Exec(ctx,q)`,
			`q:="SELECT * FROM artifacts";q,q="SELECT * FROM artifacts","SELECT * FROM users";tx.Exec(ctx,q)`,
			`q:="SELECT * FROM users";var tables,q2=[]string{q},"SELECT * FROM artifacts";q=q2;for _,table:=range tables{tx.Exec(ctx,table)}`,
		} {
			assertSQLStateSink(t, body, false, "SELECT * FROM users")
		}
		assertSQLStateSink(t, `_,value:=pair(input);tx.Exec(ctx,"SELECT * FROM artifacts WHERE id=$1",value)`, false, "SELECT * FROM artifacts WHERE id=$1")
	})
}

func runSQLLoopStateTests(t *testing.T) {
	t.Run("loop exit cannot discard unresolved SQL composition", func(t *testing.T) {
		got, err := AnalyzeSQL(map[string]string{"p.go": `package p
func Build(flag bool){q:="SELECT * FROM artifacts";for flag{q+=" UNION SELECT * FROM users"};_=q}`})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, d := range got["p.go"] {
			for _, e := range d.SQL {
				found = found || e.Disposition == "rejected" && len(e.Unresolved) > 0
			}
		}
		if !found {
			t.Fatal("unresolved composition disappeared at loop exit", got)
		}
	})
	t.Run("loop carried SQL cannot qualify a repeated sink", func(t *testing.T) {
		for _, body := range []string{
			`q:="SELECT * FROM artifacts";for i:=0;i<2;i++{tx.Exec(ctx,q);q=input}`,
			`q:="SELECT * FROM artifacts";for i:=0;i<2;q=input{tx.Exec(ctx,q);i++}`,
			`q:="SELECT * FROM artifacts";for i:=0;i<2;i++{tx.Exec(ctx,q);q=input;continue;q="SELECT * FROM artifacts"}`,
			`q:="SELECT * FROM artifacts";for i:=0;i<2;i++{tx.Exec(ctx,q);if flag{q=input;continue};q="SELECT * FROM artifacts"}`,
			`q:="SELECT * FROM artifacts";for i:=0;i<2;i++{for j:=0;j<2;j++{q=input};tx.Exec(ctx,q)}`,
			`q:="SELECT * FROM artifacts";for _,item:=range []string{"a","b"}{tx.Exec(ctx,q);q=input;_=item}`,
			`q:="SELECT * FROM artifacts";for _,q=range inputs{tx.Exec(ctx,q)}`,
			`q:="SELECT * FROM artifacts";for q=range inputs{tx.Exec(ctx,q)}`,
			`q:="SELECT * FROM artifacts";for flag{q=input;break;q="SELECT * FROM artifacts"};tx.Exec(ctx,q)`,
			`q:="SELECT * FROM artifacts";for flag{q=input};tx.Exec(ctx,q)`,
			`tables:=[]string{"artifacts"};for flag{for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)};tables=[]string{input}}`,
			`q:="SELECT * FROM artifacts";for flag{tx.Exec(ctx,q);_,q=pair(input)}`,
			`q:="SELECT * FROM artifacts";again:tx.Exec(ctx,q);q=input;goto again`,
			`q:="SELECT * FROM artifacts";for i:=0;i<2;tx.Exec(ctx,q){q=input;continue;q="SELECT * FROM artifacts"}`,
			`q:="SELECT * FROM artifacts";for i:=0;tx.Exec(ctx,q)==nil;i++{q=input}`,
			`tables:=[]string{"artifacts"};for flag{tables=[]string{input}};for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`,
		} {
			t.Run(body, func(t *testing.T) { assertSQLStateSink(t, body, true) })
		}
	})
	t.Run("loop invariants and local initialization stay qualified", func(t *testing.T) {
		for _, body := range []string{
			`q:="SELECT * FROM artifacts";for flag{tx.Exec(ctx,q)}`,
			`for flag{q:="SELECT * FROM artifacts";tx.Exec(ctx,q);q=input}`,
			`q:=input;for flag{q="SELECT * FROM artifacts";tx.Exec(ctx,q);q=input}`,
			`q:="SELECT * FROM artifacts";for flag{input="other";tx.Exec(ctx,q,input)}`,
			`for flag{tables:=[]string{"artifacts"};for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}}`,
			`q:="SELECT * FROM artifacts";switch{case flag:q="SELECT * FROM artifacts";break;default:q="SELECT * FROM artifacts"};tx.Exec(ctx,q)`,
			`q:="SELECT * FROM artifacts";for flag{q="SELECT * FROM artifacts";if flag{continue};tx.Exec(ctx,q);q=input}`,
		} {
			assertSQLStateSink(t, body, false, "SELECT * FROM artifacts")
		}
		assertSQLStateSink(t, `fragment:=shape.Literal("SELECT * FROM artifacts");for flag{tx.Exec(ctx,shape.Text(fragment));fragment=shape.Literal("SELECT * FROM artifacts WHERE id=$1")}`, false)
		assertSQLStateSite(t, `stmt,_:=db.Prepare("SELECT * FROM artifacts");for flag{stmt.Exec();stmt=opaque()}`, "stmt.Exec(", true)
	})
}
