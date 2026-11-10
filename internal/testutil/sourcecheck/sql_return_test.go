package sourcecheck

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func assertSQLHelperReturn(t *testing.T, helper, body string, unresolved bool, texts ...string) {
	t.Helper()
	source := `package p
import("github.com/jackc/pgx/v5"; shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape")
` + helper + `
func Read(tx pgx.Tx,input string,flag bool,ch <-chan bool){` + body + `}`
	got, err := AnalyzeSQL(map[string]string{"p/read.go": source})
	if err != nil {
		t.Fatal(err)
	}
	var sinks []SQLExpression
	for _, declaration := range got["p/read.go"] {
		if declaration.Name != "Read" {
			continue
		}
		for _, expression := range declaration.SQL {
			if expression.Kind == "execution" || expression.Kind == "forward" || expression.Kind == "closed" || expression.Kind == "empty" {
				sinks = append(sinks, expression)
			}
		}
	}
	if len(sinks) == 0 {
		t.Fatalf("helper execution disappeared: %#v", got)
	}
	foundUnknown := false
	actual := []string{}
	for _, sink := range sinks {
		foundUnknown = foundUnknown || len(sink.Unresolved) > 0
		actual = append(actual, sink.Text)
		if sink.Line != strings.Count(source, "\n")+1 {
			t.Fatalf("wrong execution location: %#v", sink)
		}
	}
	if foundUnknown != unresolved {
		t.Fatalf("unresolved=%v, want %v: %#v", foundUnknown, unresolved, sinks)
	}
	if texts != nil {
		sort.Strings(actual)
		sort.Strings(texts)
		if !reflect.DeepEqual(actual, texts) {
			t.Fatalf("return alternatives=%q, want %q", actual, texts)
		}
	}
}

func runSQLHelperReturnTests(t *testing.T) {
	call := `tx.Exec(ctx,query(input,flag,ch))`
	t.Run("helper summaries survive file split and nested calls", func(t *testing.T) {
		helper := `func query(input string,flag bool,ch <-chan bool)(q string){q=input;if flag{return "SELECT * FROM artifacts"};return}`
		wrapper := `func wrapper(input string,flag bool,ch <-chan bool)string{return query(input,flag,ch)}`
		assertSQLHelperReturn(t, helper+"\n"+wrapper, `tx.Exec(ctx,wrapper(input,flag,ch))`, true)
		reader := `package p
import "github.com/jackc/pgx/v5"
func Read(tx pgx.Tx,input string,flag bool,ch <-chan bool){tx.Exec(ctx,query(input,flag,ch))}`
		for _, sources := range []map[string]string{
			{"p/read.go": reader + "\n" + helper},
			{"p/renamed.go": reader, "p/helper.go": "package p\n" + helper},
		} {
			got, err := AnalyzeSQL(sources)
			if err != nil {
				t.Fatal(err)
			}
			rejected := false
			for _, declarations := range got {
				for _, declaration := range declarations {
					if declaration.Name != "Read" {
						continue
					}
					for _, expression := range declaration.SQL {
						rejected = rejected || len(expression.Unresolved) > 0
					}
				}
			}
			if !rejected {
				t.Fatal("split helper lost unknown return", got)
			}
		}
	})
	t.Run("helper returns cannot omit unsafe paths", func(t *testing.T) {
		for _, body := range []string{
			`q=input;if flag{return "SELECT * FROM artifacts"};return`,
			`q="SELECT * FROM artifacts";if flag{q=input;return};return q`,
			`if flag {if flag{return "SELECT * FROM artifacts"};q=input;return};return input`,
			`switch flag{case true:return "SELECT * FROM artifacts";default:q=input;return}`,
			`q=input;switch flag{case true:return "SELECT * FROM artifacts"};return`,
			`select{case <-ch:return input;default:};return "SELECT * FROM artifacts"`,
			`if flag{goto unsafe};return "SELECT * FROM artifacts";unsafe:return input`,
			`label:{if flag{return input}};return "SELECT * FROM artifacts"`,
			`switch {case flag:q=input;fallthrough;default:return "SELECT * FROM artifacts"}`,
			`outer:for flag{break outer};return "SELECT * FROM artifacts"`,
			`for flag{q=input;return};return "SELECT * FROM artifacts"`,
		} {
			t.Run(body, func(t *testing.T) {
				assertSQLHelperReturn(t, `func query(input string,flag bool,ch <-chan bool)(q string){`+body+`}`, call, true)
			})
		}
	})
	t.Run("named result effects apply to every return form", func(t *testing.T) {
		for _, effect := range []string{
			`defer func(){q=input}()`,
			`if flag{defer func(){q=input}()}`,
			`defer func(){q="SELECT * FROM artifacts"}();defer func(){q=input}()`,
			`func(){q=input}()`, `go func(){q=input}()`,
			`writer:=func(){q=input};escape(writer)`,
			`mutate(&q)`, `pointer:=&q;defer mutate(pointer)`,
		} {
			for _, exit := range []string{`return "SELECT * FROM artifacts"`, `return q`, `return`} {
				t.Run(effect+exit, func(t *testing.T) {
					assertSQLHelperReturn(t, `func query(input string,flag bool,ch <-chan bool)(q string){q="SELECT * FROM artifacts";`+effect+`;`+exit+`}`, call, true)
				})
			}
		}
	})
	t.Run("complete named and explicit helper alternatives", func(t *testing.T) {
		for _, body := range []string{
			`if flag{return "SELECT * FROM users"};q="SELECT * FROM artifacts";return`,
			`q="SELECT * FROM users";if flag{q="SELECT * FROM artifacts"};return`,
			`switch flag{case true:q="SELECT * FROM artifacts";default:q="SELECT * FROM users"};return`,
			`for _,table:=range []string{"users","artifacts"}{return "SELECT * FROM "+table};return "SELECT * FROM artifacts"`,
		} {
			assertSQLHelperReturn(t, `func query(input string,flag bool,ch <-chan bool)(q string){`+body+`}`, call, false, "SELECT * FROM artifacts", "SELECT * FROM users")
		}
		assertSQLHelperReturn(t, `func query(input string,flag bool,ch <-chan bool)(q string){if flag{q="SELECT * FROM artifacts"};return}`, call, false, "", "SELECT * FROM artifacts")
		for _, body := range []string{
			`f:=func()string{select{case <-ch:return input;default:return "SELECT * FROM users"}};consume(f);return "SELECT * FROM artifacts"`,
			`if flag{panic("stop")};return "SELECT * FROM artifacts"`,
			`f:=func(){defer cleanup()};consume(f);return "SELECT * FROM artifacts"`,
		} {
			assertSQLHelperReturn(t, `func query(input string,flag bool,ch <-chan bool)(q string){`+body+`}`, call, false, "SELECT * FROM artifacts")
		}
	})
	t.Run("unsupported tuple and unproven loop exit", func(t *testing.T) {
		assertSQLHelperReturn(t, `func query(input string,flag bool,ch <-chan bool)(string,error){return "SELECT * FROM artifacts",nil}`, `q,_:=query(input,flag,ch);tx.Exec(ctx,q)`, true)
		assertSQLHelperReturn(t, `func query(input string,flag bool,ch <-chan bool)string{for {return "SELECT * FROM artifacts"}}`, call, true)
	})
	t.Run("deferred exits reject complete raw helper summaries", func(t *testing.T) {
		bodies := []string{
			`q=input;defer func(){recover()}();if flag{panic("stop")};return "SELECT * FROM artifacts"`,
			`q=input;defer func(){recover()}();var items []int;if flag{_ = items[0]};return "SELECT * FROM artifacts"`,
			`q=input;defer swallow();if flag{explode()};return "SELECT * FROM artifacts"`,
			`q=input;f:=swallow;defer f();if flag{panic("stop")};return "SELECT * FROM artifacts"`,
			`q="SELECT * FROM users";defer swallow();if flag{panic("stop")};return "SELECT * FROM artifacts"`,
			`q="DELETE FROM host_grid_projection";defer swallow();if flag{panic("stop")};return "SELECT * FROM artifacts"`,
			`if flag{defer cleanup()};return "SELECT * FROM artifacts"`,
			`switch flag{case true:defer cleanup()};return "SELECT * FROM artifacts"`,
			`for flag{defer cleanup();break};return "SELECT * FROM artifacts"`,
			`defer cleanup();defer swallow();return "SELECT * FROM artifacts"`,
			`return "SELECT * FROM artifacts";defer cleanup();return "SELECT * FROM artifacts"`,
			`defer func(){consume(input)}();q="SELECT * FROM artifacts";return`,
		}
		for _, exit := range []string{`return "SELECT * FROM artifacts"`, `return q`, `return`} {
			bodies = append(bodies, `q="SELECT * FROM artifacts";defer cleanup();`+exit)
		}
		for _, body := range bodies {
			t.Run(body, func(t *testing.T) {
				assertSQLDeferredReturn(t, "(q string)", body)
			})
		}
		assertSQLDeferredReturn(t, "string", `defer cleanup();return "SELECT * FROM artifacts"`)
	})
}

func assertSQLDeferredReturn(t *testing.T, result, body string) {
	t.Helper()
	const helpers = `package p
func cleanup(){}
func consume(string){}
func swallow(){recover()}
func explode(){panic("stop")}
`
	for _, expression := range []string{
		`query(input,flag)`, `wrapper(input,flag)`, `""+query(input,flag)`,
		`fmt.Sprintf("%s",query(input,flag))`, `shape.Text(shape.Literal(query(input,flag)))`,
	} {
		helper := helpers + `func query(input string,flag bool)` + result + `{` + body + `}
func wrapper(input string,flag bool)string{return query(input,flag)}`
		reader := `package p
import("context";"fmt";"github.com/jackc/pgx/v5";shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape")
func Read(ctx context.Context,tx pgx.Tx,input string,flag bool){tx.Exec(ctx,` + expression + `)}`
		// Split definitions exercise local helper identity independently of filenames.
		got, err := AnalyzeSQL(map[string]string{"p/renamed.go": helper, "p/read.go": reader})
		if err != nil {
			t.Fatal(err)
		}
		kind := "execution"
		if strings.Contains(expression, "shape.Literal") {
			kind = "fragment"
		}
		found := false
		for _, declaration := range got["p/read.go"] {
			if declaration.Name != "Read" {
				continue
			}
			for _, site := range declaration.SQL {
				if site.Kind != kind {
					continue
				}
				found = true
				reasons := strings.Join(site.Unresolved, "\n")
				if site.Disposition != "rejected" || !site.Dynamic || site.Line != 3 || site.Column < 1 ||
					!strings.Contains(reasons, "helper query: deferred control flow is unsupported for raw helper inference at p/renamed.go:") {
					t.Fatalf("deferred summary qualified at %s: %#v", expression, site)
				}
			}
		}
		if !found {
			t.Fatalf("missing %s boundary for %s: %#v", kind, expression, got)
		}
	}
}
