import assert from "node:assert/strict";
import test from "node:test";
import { inspectGoSQL } from "../go-import-inspection.mjs";
import { sqlStatements } from "../sql-relations.mjs";
import { symlinkSync } from "node:fs";
import path from "node:path";
import { createFixture } from "./fixtures.mjs";
import { collectProjectionSources, evaluateProjectionSQL } from "../projection-sql-check.mjs";
import { collectGeneratedSQLSources } from "../backend-module-boundary-check.mjs";
import { inspect, repositoryPolicy } from "./backend-boundary-fixtures.mjs";

export function registerBoundariesTests() {
  const recoveredFiles=(body,expression='query(ctx,tx,input,flag)',result='(q string)')=>{
    const pkg='internal/modules/example/provider';
    return [{relative:`${pkg}/renamed.go`,content:`package p
import("context";"github.com/jackc/pgx/v5")
const coverage="SELECT * FROM artifacts"
func cleanup(){}
func consume(string){}
func swallow(){recover()}
func explode(){panic("stop")}
func query(ctx context.Context,tx pgx.Tx,input string,flag bool)${result}{${body}}
func wrapper(ctx context.Context,tx pgx.Tx,input string,flag bool)string{return query(ctx,tx,input,flag)}`},
    {relative:`${pkg}/read.go`,content:`package p
import("context";"fmt";"github.com/jackc/pgx/v5";shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape")
func Read(ctx context.Context,tx pgx.Tx,input string,flag bool){tx.Exec(ctx,${expression})}`},
    {relative:'internal/platform/sqlshape/shape.go',content:'package sqlshape\ntype Fragment struct{}'}];
  };
  const recoveredProjection=parsed=>evaluateProjectionSQL(parsed,[{provider_id:'artifact',source_authority_modules:['artifacts'],projection_table_ids:[]}],
    new Map([['artifacts',['artifacts']],['users',['auth']],['host_grid_projection',['projections']]]),{'internal/modules/example/provider':'artifact'},{});
  const recoveredBackend=files=>inspect(files,{sqlTableAccess:repositoryPolicy().sqlTableAccess});

  test("deferred raw helper exits reject at execution and composition in both consumers",()=>{
    const bodies=[
      'q=input;defer func(){recover()}();if flag{panic("stop")};return "SELECT * FROM artifacts"',
      'q=input;defer func(){recover()}();var items []int;if flag{_ = items[0]};return "SELECT * FROM artifacts"',
      'q=input;defer swallow();if flag{explode()};return "SELECT * FROM artifacts"',
      'q=input;f:=swallow;defer f();if flag{panic("stop")};return "SELECT * FROM artifacts"',
      'q="SELECT * FROM users";defer swallow();if flag{panic("stop")};return "SELECT * FROM artifacts"',
      'q="DELETE FROM host_grid_projection";defer swallow();if flag{panic("stop")};return "SELECT * FROM artifacts"',
      'if flag{defer cleanup()};return "SELECT * FROM artifacts"',
      'switch flag{case true:defer cleanup()};return "SELECT * FROM artifacts"',
      'for flag{defer cleanup();break};return "SELECT * FROM artifacts"',
      'defer cleanup();defer swallow();return "SELECT * FROM artifacts"',
      'return "SELECT * FROM artifacts";defer cleanup();return "SELECT * FROM artifacts"',
      'defer func(){consume(input)}();q="SELECT * FROM artifacts";return',
      ...['return "SELECT * FROM artifacts"','return q','return'].map(exit=>`q="SELECT * FROM artifacts";defer cleanup();${exit}`),
    ];
    const cases=bodies.map(body=>[body,'(q string)']);
    cases.push(['defer cleanup();return "SELECT * FROM artifacts"','string']);
    for(const [body,result] of cases) for(const expression of [
      'query(ctx,tx,input,flag)','wrapper(ctx,tx,input,flag)','""+query(ctx,tx,input,flag)',
      'fmt.Sprintf("%s",query(ctx,tx,input,flag))','shape.Text(shape.Literal(query(ctx,tx,input,flag)))',
    ]) {
      const files=recoveredFiles(body,expression,result),parsed=inspectGoSQL([files]);
      const kind=expression.includes('shape.Literal')?'fragment':'execution';
      const sites=parsed.get(files[1].relative).find(d=>d.name==='Read').sql.filter(e=>e.kind===kind);
      assert.ok(sites.length>0,`${body}: missing ${kind}`);
      for(const site of sites) {
        assert.equal(site.disposition,'rejected',`${body}: ${expression}`);
        assert.equal(site.dynamic,true);
        assert.equal(site.line,3);
        assert.equal(site.column,files[1].content.split('\n')[2].indexOf(kind==='fragment'?'shape.Literal':'tx.Exec')+1);
        assert.ok(site.unresolved.some(reason=>reason.includes(`helper query: deferred control flow is unsupported for raw helper inference at ${files[0].relative}:`)),JSON.stringify(site));
      }
      assert.ok(recoveredProjection(parsed).some(v=>v.includes('unresolved SQL construction')),body);
      assert.ok(recoveredBackend(files).some(v=>v.code==='sql_construction_unresolved'),body);
    }
  });

  test("defer admission preserves nested scopes ordinary cleanup bound data and sealed values",()=>{
    for(const body of [
      'if flag{panic("stop")};return "SELECT * FROM artifacts"',
      'f:=func(){defer cleanup()};f();return "SELECT * FROM artifacts"',
      'f:=func()string{defer swallow();panic("stop")};_ = f;return "SELECT * FROM artifacts"',
    ]) {
      const files=recoveredFiles(body);
      assert.deepEqual(recoveredProjection(inspectGoSQL([files])),[],body);
      assert.deepEqual(recoveredBackend(files),[],body);
    }
    const bound=recoveredFiles('defer swallow();if flag{panic("stop")};return input','"SELECT * FROM artifacts WHERE id=$1",query(ctx,tx,input,flag)');
    const cleanup=recoveredFiles('return "SELECT * FROM artifacts"').map(file=>({...file,content:file.content.replace('{tx.Exec(ctx,','{defer cleanup();tx.Exec(ctx,')}));
    const sealed=recoveredFiles('defer cleanup();return shape.Literal("SELECT * FROM artifacts")','shape.Text(query(ctx,tx,input,flag))','shape.Fragment').map(file=>({...file,
      content:file.content.replace('import("context";"github.com/jackc/pgx/v5")','import("context";"github.com/jackc/pgx/v5";shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape")')
        .replace('func wrapper(ctx context.Context,tx pgx.Tx,input string,flag bool)string','func wrapper(ctx context.Context,tx pgx.Tx,input string,flag bool)shape.Fragment')}));
    for(const files of [bound,cleanup,sealed]) {
      assert.deepEqual(recoveredProjection(inspectGoSQL([files])),[]);
      assert.deepEqual(recoveredBackend(files),[]);
    }
  });

  test("rejected deferred helpers retain independent SQL discovery",()=>{
    const files=recoveredFiles('tx.Exec(ctx,"DELETE FROM host_grid_projection");defer func(){tx.Exec(ctx,"DELETE FROM host_grid_projection")}();return "SELECT * FROM artifacts"');
    const parsed=inspectGoSQL([files]);
    const executions=parsed.get(files[0].relative).find(d=>d.name==='query').sql.filter(e=>e.kind==='execution' && e.text==='DELETE FROM host_grid_projection');
    assert.equal(executions.length,2);
    assert.notEqual(executions[0].column,executions[1].column);
    const projection=recoveredProjection(parsed),backend=recoveredBackend(files);
    assert.ok(projection.some(v=>v.includes('unresolved SQL construction')));
    assert.ok(projection.some(v=>v.includes('host_grid_projection')));
    assert.ok(backend.some(v=>v.code==='sql_construction_unresolved'));
    assert.ok(backend.some(v=>v.code==='sql_table_write_access'));
  });

  test("sealed tuple helpers preserve construction checks without raw tuple inference",()=>{
    const pkg='internal/modules/example/provider';
    const files=[{relative:`${pkg}/read.go`,content:`package p
import("github.com/jackc/pgx/v5";shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape")
func query()(shape.Fragment,error){return shape.Literal("SELECT * FROM artifacts"),nil}
func Read(tx pgx.Tx){q,err:=query();if err!=nil{return};tx.Exec(ctx,shape.Text(q))}`},
      {relative:'internal/platform/sqlshape/shape.go',content:'package sqlshape\ntype Fragment struct{}'},
    ];
    assert.deepEqual(evaluateProjectionSQL(inspectGoSQL([files]),[{provider_id:'artifact',source_authority_modules:['artifacts'],projection_table_ids:[]}],new Map([['artifacts',['artifacts']]]),{[pkg]:'artifact'},{}),[]);
    assert.deepEqual(inspect(files,{sqlTableAccess:repositoryPolicy().sqlTableAccess}),[]);
    const forbidden=files.map(f=>({...f,content:f.content.replace('SELECT * FROM artifacts','SELECT * FROM host_grid_projection')}));
    assert.ok(inspect(forbidden,{sqlTableAccess:repositoryPolicy().sqlTableAccess}).some(v=>v.code==='sql_table_read_access'));
    assert.ok(evaluateProjectionSQL(inspectGoSQL([forbidden]),[{provider_id:'artifact',source_authority_modules:['artifacts'],projection_table_ids:[]}],new Map([['host_grid_projection',['projections']]]),{[pkg]:'artifact'},{}).some(v=>v.includes('undeclared projection table')));
  });
  test("helper return completeness and result effects reach both SQL consumers",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}];
    const owners=new Map([["artifacts",["artifacts"]],["users",["auth"]]]);
    const bodies=[
      'q=input;if flag{return "SELECT * FROM artifacts"};return',
      'q="SELECT * FROM artifacts";if flag{q=input;return};return q',
      'switch flag{case true:return "SELECT * FROM artifacts";default:q=input;return}',
      'q=input;switch flag{case true:return "SELECT * FROM artifacts"};return',
      'select{case <-ch:return input;default:};return "SELECT * FROM artifacts"',
      'if flag{goto unsafe};return "SELECT * FROM artifacts";unsafe:return input',
      'label:{if flag{return input}};return "SELECT * FROM artifacts"',
      'switch{case flag:q=input;fallthrough;default:return "SELECT * FROM artifacts"}',
      'for flag{q=input;return};return "SELECT * FROM artifacts"',
    ];
    for(const effect of ['defer func(){q=input}()','if flag{defer func(){q=input}()}','defer func(){q="SELECT * FROM artifacts"}();defer func(){q=input}()',
      'func(){q=input}()','go func(){q=input}()','writer:=func(){q=input};escape(writer)','mutate(&q)','pointer:=&q;defer mutate(pointer)']) {
      for(const exit of ['return "SELECT * FROM artifacts"','return q','return']) bodies.push(`q="SELECT * FROM artifacts";${effect};${exit}`);
    }
    const files=(body,sink='tx.Exec(ctx,query(input,flag,ch))')=>[{relative:`${pkg}/read.go`,content:`package p
import("github.com/jackc/pgx/v5";shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape")
const coverage="SELECT * FROM artifacts"
func query(input string,flag bool,ch <-chan bool)(q string){${body}}
func Read(tx pgx.Tx,input string,flag bool,ch <-chan bool){${sink}}`}];
    for(const body of bodies) for(const sink of ['tx.Exec(ctx,query(input,flag,ch))','tx.Exec(ctx,shape.Text(shape.Literal(query(input,flag,ch))))']) {
      const input=files(body,sink), parsed=inspectGoSQL([input]);
      const execution=parsed.get(input[0].relative).find(d=>d.name==='Read').sql;
      assert.ok(execution.some(e=>e.unresolved.length>0),body);
      assert.ok(evaluateProjectionSQL(parsed,providers,owners,{[pkg]:"artifact"},{}).length>0,body);
      assert.ok(inspect(input,{sqlTableAccess:repositoryPolicy().sqlTableAccess}).some(v=>v.code==='sql_construction_unresolved'),body);
    }
    for(const body of ['if flag{panic("stop")};return "SELECT * FROM artifacts"',
      'if flag{return "SELECT * FROM artifacts"};q="SELECT * FROM artifacts";return']) {
      const input=files(body);
      assert.deepEqual(evaluateProjectionSQL(inspectGoSQL([input]),providers,owners,{[pkg]:"artifact"},{}),[],body);
      assert.deepEqual(inspect(input,{sqlTableAccess:repositoryPolicy().sqlTableAccess}),[],body);
    }
    for(const alternative of ['SELECT * FROM users','DELETE FROM host_grid_projection']) {
      const input=files(`q=${JSON.stringify(alternative)};if flag{return "SELECT * FROM artifacts"};return`);
      const parsed=inspectGoSQL([input]);
      assert.ok(parsed.get(input[0].relative).find(d=>d.name==='Read').sql.some(e=>e.kind==='execution' && e.text===alternative));
      assert.ok(evaluateProjectionSQL(parsed,providers,owners,{[pkg]:"artifact"},{}).length>0);
      if(alternative.startsWith('DELETE')) assert.ok(inspect(input,{sqlTableAccess:repositoryPolicy().sqlTableAccess}).some(v=>v.code==='sql_table_write_access'));
    }
  });
  test("ordinary branch and loop data remains SQL free in both consumers",()=>{
    const pkg="internal/modules/example/provider";
    const files=[{relative:`${pkg}/read.go`,content:`package p
import "errors"
const coverage="SELECT * FROM artifacts"
type A struct{ID string};type B struct{Count int};type Provider struct{Query bool}
func Variant(input any){switch value:=input.(type){case A:if value.ID==""{return};case B:if value.Count==0{return}}}
func Named(flag bool)(count int,err error){if flag{count=1;err=errors.New("data")};return count,err}
func Runes(value string){for _,r:=range value{if r>='A' && r<='Z'{r+='a'-'A'};consume(r)}}
func Providers(items []Provider){for _,provider:=range items{if !provider.Query{return}}}`}];
    assert.deepEqual(evaluateProjectionSQL(inspectGoSQL([files]),[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}],new Map([["artifacts",["artifacts"]]]),{[pkg]:"artifact"},{}),[]);
    assert.deepEqual(inspect(files,{sqlTableAccess:repositoryPolicy().sqlTableAccess}),[]);
  });
  test("state joins retain declared ports and unresolved composition",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}];
    for(const source of [
      `package p
import("context";"github.com/jackc/pgx/v5")
type Runner interface{Apply(context.Context,string,...any)(pgx.Rows,error)}
type A struct{};func(A)Apply(context.Context,string,...any)(pgx.Rows,error){return nil,nil}
type B struct{};func(B)Apply(context.Context,string,...any)(pgx.Rows,error){return nil,nil}
const coverage="SELECT * FROM artifacts"
func Read(r Runner,input string,flag bool){if flag{r=A{}}else{r=B{}};r.Apply(ctx,input)}`,
      `package p
func Build(flag bool){q:="SELECT * FROM artifacts";for flag{q+=" UNION SELECT * FROM users"};_=q}`,
    ]) {
      const files=[{relative:`${pkg}/read.go`,content:source}];
      assert.ok(evaluateProjectionSQL(inspectGoSQL([files]),providers,new Map([["artifacts",["artifacts"]],["users",["auth"]]]),{[pkg]:"artifact"},{}).some(v=>v.includes('unresolved SQL construction')));
      assert.ok(inspect(files,{sqlTableAccess:repositoryPolicy().sqlTableAccess}).some(v=>v.code==='sql_construction_unresolved'));
    }
  });
  test("loop carried values fail both policies at repeated execution",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}];
    const owners=new Map([["artifacts",["artifacts"]]]);
    const files=body=>[{relative:`${pkg}/read.go`,content:`package p
import "github.com/jackc/pgx/v5"
const coverage="SELECT * FROM artifacts"
func Read(tx pgx.Tx,input string,flag bool){${body}}`}];
    for(const body of [
      'q:="SELECT * FROM artifacts";for i:=0;i<2;i++{tx.Exec(ctx,q);q=input}',
      'q:="SELECT * FROM artifacts";for i:=0;i<2;q=input{tx.Exec(ctx,q);i++}',
      'q:="SELECT * FROM artifacts";for i:=0;i<2;tx.Exec(ctx,q){q=input;continue;q="SELECT * FROM artifacts"}',
      'q:="SELECT * FROM artifacts";for flag{tx.Exec(ctx,q);q=input;continue;q="SELECT * FROM artifacts"}',
      'q:="SELECT * FROM artifacts";for flag{tx.Exec(ctx,q);if flag{q=input;continue};q="SELECT * FROM artifacts"}',
      'q:="SELECT * FROM artifacts";for _,item:=range []string{"a","b"}{tx.Exec(ctx,q);q=input;_=item}',
      'q:="SELECT * FROM artifacts";for flag{q=input;break;q="SELECT * FROM artifacts"};tx.Exec(ctx,q)',
      'tables:=[]string{"artifacts"};for flag{for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)};tables=[]string{input}}',
    ]) {
      const source=files(body);
      assert.ok(evaluateProjectionSQL(inspectGoSQL([source]),providers,owners,{[pkg]:"artifact"},{}).some(v=>v.includes('unresolved SQL construction')),body);
      assert.ok(inspect(source,{sqlTableAccess:repositoryPolicy().sqlTableAccess}).some(v=>v.code==='sql_construction_unresolved'),body);
    }
    for(const body of ['q:="SELECT * FROM artifacts";for flag{tx.Exec(ctx,q)}','for flag{q:="SELECT * FROM artifacts";tx.Exec(ctx,q);q=input}','for flag{tx.Exec(ctx,"SELECT * FROM artifacts WHERE id=$1",input);input="other"}']) {
      const source=files(body);
      assert.deepEqual(evaluateProjectionSQL(inspectGoSQL([source]),providers,owners,{[pkg]:"artifact"},{}),[],body);
      assert.deepEqual(inspect(source,{sqlTableAccess:repositoryPolicy().sqlTableAccess}),[],body);
    }
  });
  test("all assignment destinations invalidate stale SQL in both policies",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}];
    const files=body=>[{relative:`${pkg}/read.go`,content:`package p
import "github.com/jackc/pgx/v5"
const coverage="SELECT * FROM artifacts"
func pair(input string)(bool,string){return true,input}
func triple(input string)(bool,bool,string){return true,true,input}
func lists(input string)(bool,[]string){return true,[]string{input}}
func Read(tx pgx.Tx,input string,flag bool){${body}}`}];
    for(const body of [
      'q:="SELECT * FROM artifacts";_,q=pair(input);tx.Exec(ctx,q)',
      'q:="SELECT * FROM artifacts";ok,q:=pair(input);_=ok;tx.Exec(ctx,q)',
      '_,q:=pair(input);tx.Exec(ctx,q)',
      'var _,q=pair(input);tx.Exec(ctx,q)',
      'q:="SELECT * FROM artifacts";_,_,q=triple(input);tx.Exec(ctx,q)',
      'tables:=[]string{"artifacts"};_,tables=lists(input);for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}',
    ]) {
      const source=files(body);
      assert.ok(evaluateProjectionSQL(inspectGoSQL([source]),providers,new Map([["artifacts",["artifacts"]]]),{[pkg]:"artifact"},{}).some(v=>v.includes('unresolved SQL construction')),body);
      assert.ok(inspect(source,{sqlTableAccess:repositoryPolicy().sqlTableAccess}).some(v=>v.code==='sql_construction_unresolved'),body);
    }
    const safe=files('_,value:=pair(input);tx.Exec(ctx,"SELECT * FROM artifacts WHERE id=$1",value)');
    assert.deepEqual(evaluateProjectionSQL(inspectGoSQL([safe]),providers,new Map([["artifacts",["artifacts"]]]),{[pkg]:"artifact"},{}),[]);
    assert.deepEqual(inspect(safe,{sqlTableAccess:repositoryPolicy().sqlTableAccess}),[]);
  });
  test("branch collection alternatives retain both policy decisions",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}];
    const owners=new Map([["artifacts",["artifacts"]],["users",["auth"]],["host_grid_projection",["projections"]]]);
    const files=body=>[{relative:`${pkg}/read.go`,content:`package p
import "github.com/jackc/pgx/v5"
const coverage="SELECT * FROM artifacts"
func Read(tx pgx.Tx,input string,flag bool){${body}}`}];
    for(const branch of [
      'if flag{tables=[]string{"forbidden"}}else{tables=[]string{"artifacts"}}',
      'if flag{tables=[]string{"artifacts"}}else{tables=[]string{"forbidden"}}',
      'switch{case flag:tables=[]string{"forbidden"};default:tables=[]string{"artifacts"}}',
      'switch{case flag:tables=[]string{"artifacts"};default:tables=[]string{"forbidden"}}',
      'if flag{tables=[]string{"forbidden"}}',
      'switch{case flag:tables=[]string{"forbidden"}}',
    ]) {
      const body=`tables:=[]string{"artifacts"};${branch};for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}`;
      const projection=files(body.replaceAll('forbidden','users'));
      assert.ok(evaluateProjectionSQL(inspectGoSQL([projection]),providers,owners,{[pkg]:"artifact"},{}).some(v=>v.includes('undeclared source authority auth')),branch);
      const backend=files(body.replaceAll('forbidden','host_grid_projection').replaceAll('SELECT * FROM','DELETE FROM'));
      assert.ok(inspect(backend,{sqlTableAccess:repositoryPolicy().sqlTableAccess}).some(v=>v.code==='sql_table_write_access'),branch);
      const allowed=files(body.replaceAll('forbidden','artifacts'));
      assert.deepEqual(evaluateProjectionSQL(inspectGoSQL([allowed]),providers,owners,{[pkg]:"artifact"},{}),[],branch);
      assert.deepEqual(inspect(allowed,{sqlTableAccess:repositoryPolicy().sqlTableAccess}),[],branch);
    }
  });
  test("unknown-state local constants never inherit package SQL authority",()=>{
    const pkg="internal/modules/example/provider";
    const files=[{relative:`${pkg}/read.go`,content:`package p
import "github.com/jackc/pgx/v5"
const table="artifacts"
const coverage="SELECT * FROM artifacts"
func Read(tx pgx.Tx){label:{const table="users";tx.Query(ctx,"SELECT * FROM "+table)}}`}];
    assert.ok(evaluateProjectionSQL(inspectGoSQL([files]),[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}],new Map([["artifacts",["artifacts"]],["users",["auth"]]]),{[pkg]:"artifact"},{}).length>0);
    assert.ok(inspect(files,{sqlTableAccess:repositoryPolicy().sqlTableAccess}).some(v=>v.code==='sql_construction_unresolved'));
  });

  test("captured and escaped collection SQL fails both policies",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}];
    for(const body of [
      'tables:=[]string{"artifacts"};defer func(){for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}}();tables=[]string{input}',
      'tables:=[]string{"artifacts"};other:=tables;other[0]=input;for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)}',
      'tables:=[]string{input};f:=func(){tables=[]string{"artifacts"}};for _,table:=range tables{tx.Exec(ctx,"SELECT * FROM "+table)};_=f',
    ]) {
      const files=[{relative:`${pkg}/read.go`,content:`package p
import "github.com/jackc/pgx/v5"
const coverage="SELECT * FROM artifacts"
func Read(tx pgx.Tx,input string){${body}}`}];
      assert.ok(evaluateProjectionSQL(inspectGoSQL([files]),providers,new Map([["artifacts",["artifacts"]]]),{[pkg]:"artifact"},{}).some(v=>v.includes('unresolved SQL construction')),body);
      assert.ok(inspect(files,{sqlTableAccess:repositoryPolicy().sqlTableAccess}).some(v=>v.code==='sql_construction_unresolved'),body);
    }
  });

  test("generic SQL function signatures reject in both policies",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}];
    for(const source of [
      'type QueryFunc[T any] func(context.Context,string,...any)(T,error);func Read(query QueryFunc[pgx.Rows],input string){query(ctx,input)}',
      'func Read[F ~func(context.Context,string,...any)(pgx.Rows,error)](query F,input string){query(ctx,input)}',
    ]) {
      const files=[{relative:`${pkg}/read.go`,content:`package p
import("context";"github.com/jackc/pgx/v5")
const coverage="SELECT * FROM artifacts"
${source}`}];
      assert.ok(evaluateProjectionSQL(inspectGoSQL([files]),providers,new Map([["artifacts",["artifacts"]]]),{[pkg]:"artifact"},{}).some(v=>v.includes('unresolved SQL construction')));
      assert.ok(inspect(files,{sqlTableAccess:repositoryPolicy().sqlTableAccess}).some(v=>v.code==='sql_construction_unresolved'));
    }
  });

  test("dot imports fail admission in both SQL consumers",()=>{
    for(const pkg of ['github.com/jackc/pgx/v5/pgproto3','github.com/jackc/pgx/v5','github.com/JochiRaider/cartulary/internal/platform/sqlshape','github.com/JochiRaider/cartulary/internal/helper','strings']) {
      for(const relative of ['internal/modules/example/read.go','internal/gen/example/read.go']) {
        const files=[{relative,content:`//go:build other
package example
import . "${pkg}"
func Send(input string){buf,err:=(&Query{String:input}).Encode(nil);if err==nil {writer.Write(buf)}}`}];
        assert.throws(()=>inspectGoSQL([files]),/dot import .* is unsupported in SQL analysis/);
        assert.throws(()=>inspect(files,{sqlTableAccess:repositoryPolicy().sqlTableAccess}),/dot import .* is unsupported in SQL analysis/);
      }
    }
  });

  test("closure capture stability is enforced by both consumers",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}];
    const owners=new Map([["artifacts",["artifacts"]]]);
    const files=body=>[{relative:`${pkg}/read.go`,content:`package p
import("database/sql";"github.com/jackc/pgx/v5")
const coverage="SELECT * FROM artifacts"
func Read(tx pgx.Tx,db *sql.DB,input string){${body}}`}];
    const projection=source=>evaluateProjectionSQL(inspectGoSQL([source]),providers,owners,{[pkg]:"artifact"},{});
    const backend=source=>inspect(source,{sqlTableAccess:repositoryPolicy().sqlTableAccess});
    for(const body of [
      'q:="SELECT * FROM artifacts";defer func(){tx.Exec(ctx,q)}();q=input',
      'q:="SELECT * FROM artifacts";f:=func(){tx.Exec(ctx,q)};q=input;f()',
      'q:="SELECT * FROM artifacts";go func(){tx.Exec(ctx,q)}();q=input',
      'q:="SELECT * FROM artifacts";f:=func(){q=input};tx.Exec(ctx,q);f()',
      'q:="SELECT * FROM artifacts";p:=&q;tx.Exec(ctx,q);_=p',
      'stmt,_:=db.Prepare("SELECT * FROM artifacts");defer func(){stmt.Exec()}()',
      'var batch pgx.Batch;batch.Queue("SELECT * FROM artifacts");defer func(){tx.SendBatch(ctx,&batch)}()',
    ]) {assert.ok(projection(files(body)).some(v=>v.includes('unresolved SQL construction')),body);assert.ok(backend(files(body)).some(v=>v.code==='sql_construction_unresolved'),body);}
    for(const body of ['q:="SELECT * FROM artifacts";defer func(){tx.Exec(ctx,q)}()','q:="SELECT * FROM artifacts";copy:=q;q=input;defer func(){tx.Exec(ctx,copy)}()','defer func(){tx.Exec(ctx,"SELECT * FROM artifacts WHERE id=$1",input)}();input="other"']) {
      assert.deepEqual(projection(files(body)),[],body);assert.deepEqual(backend(files(body)),[],body);
    }
  });

  test("SQL callable values reject in both policies regardless of alias spelling",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}];
    const owners=new Map([["artifacts",["artifacts"]]]);
    for(const declaration of ['type QueryFunc=func(context.Context,string,...any)(pgx.Rows,error)','type QueryFunc func(context.Context,string,...any)(pgx.Rows,error)']) {
      for(const body of ['query:=store.Query;query(ctx,input)','callback(ctx,input)','store.Query(ctx,input)','items[0](ctx,input)','factory()(ctx,input)','consume(callback)']) {
        const files=[{relative:`${pkg}/read.go`,content:`package p
import("context";"github.com/jackc/pgx/v5")
${declaration}
type Store struct{Query QueryFunc}
const coverage="SELECT * FROM artifacts"
func factory()QueryFunc{return nil}
func Read(store Store,callback QueryFunc,items []QueryFunc,input string){${body}}`}];
        assert.ok(evaluateProjectionSQL(inspectGoSQL([files]),providers,owners,{[pkg]:"artifact"},{}).some(v=>v.includes('unresolved SQL construction')),body);
        assert.ok(inspect(files,{sqlTableAccess:repositoryPolicy().sqlTableAccess}).some(v=>v.code==='sql_construction_unresolved'),body);
      }
    }
  });

  test("query operands retain access decisions in both policies",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}];
    const owners=new Map([["artifacts",["artifacts"]],["users",["auth"]],["host_grid_projection",["projections"]]]);
    const files=sql=>[{relative:`${pkg}/read.go`,content:`package p
import "github.com/jackc/pgx/v5"
const coverage="SELECT * FROM artifacts"
func Read(tx pgx.Tx){tx.Query(ctx,${JSON.stringify(sql)})}`}];
    const projection=sql=>evaluateProjectionSQL(inspectGoSQL([files(sql)]),providers,owners,{[pkg]:"artifact"},{});
    for(const sql of ['WITH unused AS (SELECT 1) TABLE users','SELECT * FROM artifacts UNION ALL TABLE users','TABLE artifacts INTERSECT TABLE users']) {
      assert.ok(projection(sql).some(v=>v.includes('undeclared source authority auth')),sql);
      assert.ok(inspect(files(sql.replaceAll('users','host_grid_projection')),{sqlTableAccess:repositoryPolicy().sqlTableAccess}).some(v=>v.code==='sql_table_read_access'),sql);
    }
    for(const sql of ['WITH unused AS (SELECT 1) TABLE artifacts','WITH users AS (TABLE artifacts) TABLE users','SELECT 1 UNION ALL VALUES(2)']) {
      assert.deepEqual(projection(sql),[],sql);
      assert.deepEqual(inspect(files(sql),{sqlTableAccess:repositoryPolicy().sqlTableAccess}),[],sql);
    }
  });
  test("backend SQL: generated types participate and generated SQL retains policy",t=>{
    const f=createFixture(t);
    assert.throws(()=>collectGeneratedSQLSources(f.root));
    f.write('internal/gen/ports/types.go','package ports\nimport "github.com/jackc/pgx/v5"\ntype Rows=pgx.Rows');
    const generated=collectGeneratedSQLSources(f.root);
    const source=[{relative:'internal/modules/example/read.go',content:`package example
import("context";"github.com/JochiRaider/cartulary/internal/gen/ports")
type Reader interface{Query(context.Context,string,...any)(ports.Rows,error)}
func Read(tx Reader){tx.Query(ctx,input)}`}];
    const policy={sqlTableAccess:repositoryPolicy().sqlTableAccess};
    assert.ok(inspect(source,policy,{sqlAdditionalFiles:generated}).some(v=>v.code==='sql_construction_unresolved'));
    const safe=source.map(file=>({...file,content:file.content.replace('tx.Query(ctx,input)','tx.Query(ctx,"SELECT 1")')}));
    assert.deepEqual(inspect(safe,policy,{sqlAdditionalFiles:generated}),[]);
    f.write('internal/gen/ports/write.go','//go:build other\npackage ports\nconst query="DELETE FROM host_grid_projection"');
    assert.ok(inspect([],policy,{sqlAdditionalFiles:collectGeneratedSQLSources(f.root)}).some(v=>v.code==='sql_table_write_access'));
    symlinkSync(path.join(f.root,'internal/gen/ports/types.go'),path.join(f.root,'internal/gen/ports/linked.go'));
    assert.throws(()=>collectGeneratedSQLSources(f.root),/generated SQL source: linked source/);
  });
  test("native port result aliases preserve execution in both policies",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}];
    const owners=new Map([["artifacts",["artifacts"]],["users",["auth"]],["host_grid_projection",["projections"]]]);
    for(const declaration of ['type Rows=pgx.Rows','type First=pgx.Rows;type Rows=First','type Rows=Missing']) for(const results of ['Rows,error','bool,Rows,error']) {
      const files=sql=>[{relative:`${pkg}/read.go`,content:`package p
import("context";"github.com/jackc/pgx/v5")
${declaration};type Text=string;type Reader interface{Query(context.Context,Text,...any)(${results})}
const coverage="SELECT * FROM artifacts"
func Read(tx Reader,input string){tx.Query(ctx,${sql})}`}];
      const projection=source=>evaluateProjectionSQL(inspectGoSQL([source]),providers,owners,{[pkg]:"artifact"},{});
      const backend=source=>inspect(source,{sqlTableAccess:repositoryPolicy().sqlTableAccess});
      assert.ok(projection(files('input')).some(v=>v.includes('unresolved SQL construction')));
      assert.ok(backend(files('input')).some(v=>v.code==='sql_construction_unresolved'));
      if(!declaration.includes('Missing')) {
        assert.ok(projection(files('"SELECT * FROM users"')).some(v=>v.includes('undeclared source authority')));
        assert.ok(backend(files('"DELETE FROM host_grid_projection"')).some(v=>v.code==='sql_table_write_access'));
        assert.deepEqual(projection(files('"SELECT * FROM artifacts"')),[]);
        assert.deepEqual(backend(files('"SELECT * FROM artifacts"')),[]);
      }
    }
  });
  test("raw mutable SQL builders fail both policies without rejecting bound data",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}];
    const owners=new Map([["artifacts",["artifacts"]],["host_grid_projection",["projections"]]]);
    const files=body=>[{relative:`${pkg}/read.go`,content:`package p
import("strings";"github.com/jackc/pgx/v5";shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape")
const coverage="SELECT * FROM artifacts"
func Read(tx pgx.Tx,input string){${body}}`}];
    const projection=source=>evaluateProjectionSQL(inspectGoSQL([source]),providers,owners,{[pkg]:"artifact"},{});
    const backend=source=>inspect(source,{sqlTableAccess:repositoryPolicy().sqlTableAccess});
    for(const body of [
      'var b strings.Builder;b.WriteString("--");reset:=b.Reset;reset();b.WriteString("DELETE FROM host_grid_projection");tx.Exec(ctx,b.String())',
      'var b strings.Builder;b.WriteString("SELECT * FROM artifacts");tx.Query(ctx,b.String())',
      'var b strings.Builder;write:=b.WriteString;write(input);tx.Query(ctx,b.String())',
      'var b strings.Builder;b.WriteString(input);tx.Query(ctx,shape.Text(shape.Literal(b.String())))',
    ]) {
      assert.ok(projection(files(body)).some(v=>v.includes("unresolved SQL construction")),body);
      assert.ok(backend(files(body)).some(v=>v.code==="sql_construction_unresolved"),body);
    }
    for(const body of [
      'var b strings.Builder;reset:=b.Reset;reset();b.WriteString(input);tx.Query(ctx,"SELECT * FROM artifacts WHERE id=$1",b.String())',
      'var b shape.Builder;b.Write(shape.Literal("SELECT * FROM artifacts"));tx.Query(ctx,shape.Text(b.Fragment()))',
      'q:="SELECT * FROM artifacts";tx.Query(ctx,q)',
    ]) { assert.deepEqual(projection(files(body)),[]); assert.deepEqual(backend(files(body)),[]); }
  });
  test("native SQL: raw driver arguments fail both enforcement policies",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}];
    const owners=new Map([["artifacts",["artifacts"]],["users",["auth"]]]);
    for(const [type,call] of [['*pgconn.PgConn','handle.ExecParams(ctx,opaque(),nil,nil,nil,nil)'],['*sql.DB','handle.Query(opaque())'],['*pgconn.PgConn','handle.ExecPrepared(ctx,"name",nil,nil,nil)'],['any','client:=proto.NewFrontend(reader,writer);client.Send(message)'],['any','message:=proto.Query{String:input}']]) {
      const files=[{relative:`${pkg}/read.go`,content:`package p
import("database/sql";"github.com/jackc/pgx/v5/pgconn";proto "github.com/jackc/pgx/v5/pgproto3")
func Read(handle ${type}){${call}}`}];
      assert.ok(evaluateProjectionSQL(inspectGoSQL([files]),providers,owners,{[pkg]:"artifact"},{}).some(v=>v.includes('unresolved SQL construction')));
      assert.ok(inspect(files,{sqlTableAccess:repositoryPolicy().sqlTableAccess}).some(v=>v.code==='sql_construction_unresolved'));
    }
  });
  test("SQL quoted tokens distinguish ordinary and escape-string backslashes",()=>{
    const ordinary="SELECT '\\' FROM artifacts UNION SELECT 1 FROM users";
    assert.deepEqual(sqlStatements(ordinary).flatMap(s=>s.accesses),[{table:"artifacts",operation:"read"},{table:"users",operation:"read"}]);
    const escaped="SELECT E'\\\\' FROM artifacts";
    assert.deepEqual(sqlStatements(escaped).flatMap(s=>s.accesses),[{table:"artifacts",operation:"read"}]);
  });
  test("closed templates: comment and quoted placeholders fail before SQL classification",()=>{
    const prefix='package p\nimport shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape"\nimport "github.com/jackc/pgx/v5"\n';
    for (const constructor of ['Format','Command']) for(const template of ['SELECT 1 FROM artifacts /* %s */','SELECT 1 -- %s','SELECT "%s" FROM artifacts',"SELECT E'%s' FROM artifacts",'SELECT $$%s$$','SELECT prefix%s','SELECT %ssuffix']) {
      assert.throws(()=>sqlStatements(prefix+`func Read(tx pgx.Tx){q:=shape.${constructor}(${JSON.stringify(template)},shape.Quote(input));_=q}`),/unresolved SQL construction/);
    }
    assert.doesNotThrow(()=>sqlStatements(prefix+`func Read(tx pgx.Tx){q:=shape.Literal("SELECT '%s /* */' FROM artifacts");_=q}`));
  });
  test("SQL discovery: field/index and nested execution cannot escape either policy",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:[]}];
    const owners=new Map([["users",["auth"]],["artifacts",["artifacts"]]]);
    for(const body of ['s.q=shape.Literal("SELECT * FROM users")','items[0]=shape.Literal("SELECT * FROM users")','ok:=tx.QueryRow(ctx,unknown()).Scan(&n)==nil','switch tx.QueryRow(ctx,unknown()).Scan(&n){default:}']) {
      const files=[{relative:`${pkg}/read.go`,content:`package p
import ("github.com/jackc/pgx/v5"; shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape")
func Read(tx pgx.Tx){${body}}`}];
      assert.ok(evaluateProjectionSQL(inspectGoSQL([files]),providers,owners,{[pkg]:"artifact"},{}).length>0,body);
    }
    const rules=repositoryPolicy().sqlTableAccess.filter(rule=>rule.id.endsWith('-projection-storage-access'));
    const files=[{relative:`${pkg}/write.go`,content:'package p\nimport shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape"\nfunc Write(){s.q=shape.Command("DELETE FROM host_grid_projection")}'}];
    assert.ok(inspect(files,{sqlTableAccess:rules}).some(v=>v.code==='sql_table_write_access'));
  });
  test("projection SQL: Timeline owner locks stay outside the read capability",()=>{
    const rule=repositoryPolicy().sqlTableAccess.find(r=>r.id==='timeline-source-owner-access');
    const inspectPackage=(pkg,sql)=>inspect([{relative:`internal/modules/timeline/${pkg}/read.go`,content:`package p\nconst query=${JSON.stringify(sql)}`}],{sqlTableAccess:[rule]});
    assert.deepEqual(inspectPackage('sourcerepository','SELECT * FROM timeline_events'),[]);
    assert.deepEqual(inspectPackage('sourcelocking','SELECT * FROM timeline_events FOR UPDATE'),[]);
    assert.ok(inspectPackage('sourcerepository','SELECT * FROM timeline_events FOR UPDATE').some(v=>v.code==='sql_table_lock_access'));
    assert.ok(inspectPackage('sourcelocking/child','SELECT * FROM timeline_events FOR UPDATE').some(v=>v.code==='sql_table_lock_access'));
    assert.ok(inspectPackage('sourcerepository','UPDATE timeline_events SET row_version=1').some(v=>v.code==='sql_table_write_access'));
  });
  test("projection SQL: package rename and split parity across architecture checkers",()=>{
    const pkg="internal/modules/projections/internal/storage";
    const providers=[{provider_id:"host",source_authority_modules:["entities"],projection_table_ids:["host_grid_projection"]}];
    const owners=new Map([["host_grid_projection",["projections"]],["identity_grid_projection",["projections"]]]);
    const rules=repositoryPolicy().sqlTableAccess.filter(rule=>rule.id.endsWith('-projection-storage-access'));
    const bindings={[pkg]:{Write:"host"}};
    assert.ok(inspect([], {sqlTableAccess:rules}, {completeSourceSet:true}).some(v=>v.code==='sql_package_grant_stale'));
    for (const name of ['renamed.go','split.go']) {
      const files=[{relative:`${pkg}/${name}`,content:'package storage\nimport "github.com/jackc/pgx/v5"\nfunc Write(tx pgx.Tx){tx.Exec(ctx,"UPDATE host_grid_projection SET row_version=1")}'}];
      assert.deepEqual(inspect(files,{sqlTableAccess:rules}),[]);
      assert.deepEqual(evaluateProjectionSQL(inspectGoSQL([files]),providers,owners,{},bindings),[]);
    }
    for (const target of [`${pkg}/child/read.go`,'internal/modules/peer/read.go']) {
      const files=[{relative:target,content:'package p\nimport "github.com/jackc/pgx/v5"\nfunc Write(tx pgx.Tx){tx.Exec(ctx,"UPDATE host_grid_projection SET row_version=1")}'}];
      assert.ok(inspect(files,{sqlTableAccess:rules}).some(v=>v.code==='sql_table_write_access'));
      assert.ok(evaluateProjectionSQL(inspectGoSQL([files]),providers,owners,{},bindings).length>0);
    }
    const wrongProvider=[{relative:`${pkg}/host.go`,content:'package storage\nimport "github.com/jackc/pgx/v5"\nfunc Write(tx pgx.Tx){tx.Exec(ctx,"UPDATE identity_grid_projection SET row_version=1")}'}];
    assert.deepEqual(inspect(wrongProvider,{sqlTableAccess:rules}),[]);
    assert.ok(evaluateProjectionSQL(inspectGoSQL([wrongProvider]),providers,owners,{},bindings).some(v=>v.includes('undeclared projection table')));
    for (const sql of ['UPDATE host_grid_projection SET row_version=1','SELECT * FROM host_grid_projection FOR UPDATE']) {
      assert.ok(inspect([{relative:'internal/modules/projections/internal/queryengine/new.go',content:`package p\nconst query=${JSON.stringify(sql)}`}],{sqlTableAccess:rules}).length>0);
    }
  });
  test("projection SQL: dependency closure includes sibling, transitive and constrained sources",(t)=>{
    const fixture=createFixture(t);
    fixture.write("go.mod","module example.test/project\n");
    fixture.write("owner/a.go",'package a\nimport alias "example.test/project/sibling"\nvar _=alias.Read');
    fixture.write("owner/ignored_test.go","invalid");
    fixture.write("sibling/read.go",'package sibling\nimport _ "example.test/project/generated/helper"\nfunc Read(){}');
    fixture.write("generated/helper/build.go",'//go:build never_enabled\n\npackage helper\nconst query="SELECT * FROM users"');
    fixture.write("unreachable/poison.go","invalid");
    const entries=[{package:"owner",provider:"artifact"}];
    const closure=collectProjectionSources(fixture.root,entries);
    assert.deepEqual(closure.files.map(f=>f.relative),["generated/helper/build.go","owner/a.go","sibling/read.go"]);
    assert.deepEqual([...closure.consumers.get('generated/helper')],['artifact']);
    const providers=[{provider_id:'artifact',source_authority_modules:['artifacts'],projection_table_ids:[]}];
    const owners=new Map([['users',['auth']]]);
    assert.ok(evaluateProjectionSQL(inspectGoSQL([closure.files]),providers,owners,{}, {},false,closure.consumers).some(v=>v.includes('unassigned SQL declaration')));
    assert.throws(()=>collectProjectionSources(fixture.root,[{package:"missing"}]),/ENOENT/);
    assert.throws(()=>collectProjectionSources(fixture.root,[]),/empty/);
    fixture.write('sibling/read.go','package sibling\nimport _ "example.test/project/missing"');
    assert.throws(()=>collectProjectionSources(fixture.root,entries),/ENOENT/);
    fixture.write('sibling/read.go','package sibling\nimport _ "foreign.test/sql"');
    assert.throws(()=>collectProjectionSources(fixture.root,entries),/unadmitted foreign/);
    assert.deepEqual(collectProjectionSources(fixture.root,entries,new Set(['foreign.test/sql'])).foreign,['foreign.test/sql']);
    fixture.write('sibling/read.go','package sibling\n//go:linkname Read hidden.Read\nfunc Read()');
    assert.throws(()=>collectProjectionSources(fixture.root,entries),/linked Go declaration/);
    fixture.write('sibling/read.go','package sibling');
    fixture.write('sibling/hidden.s','unclassified executable source');
    assert.throws(()=>collectProjectionSources(fixture.root,entries),/unclassified native source/);
    symlinkSync(path.join(fixture.root,"generated/helper/build.go"),path.join(fixture.root,"owner/linked.go"));
    assert.throws(()=>collectProjectionSources(fixture.root,entries),/linked/);
    symlinkSync(path.join(fixture.root,"owner"),path.join(fixture.root,"linked"));
    assert.throws(()=>collectProjectionSources(fixture.root,[{package:'linked'}]),/linked/);
  });

  test("projection SQL: shared dependencies retain every consumer and require source coverage",()=>{
    const providers=[
      {provider_id:'artifact',source_authority_modules:['artifacts','links'],projection_table_ids:['artifact_grid_projection']},
      {provider_id:'assessment',source_authority_modules:['assessments','links'],projection_table_ids:['assessment_grid_projection']}
    ];
    const owners=new Map([['artifacts',['artifacts']],['record_links',['links']],['artifact_grid_projection',['projections']]]);
    const consumers=new Map([['shared',new Set(['artifact','assessment'])]]);
    const check=(sql,defaults={shared:'shared_read'})=>evaluateProjectionSQL(inspectGoSQL([[{relative:'shared/read.go',content:`package shared\nimport "github.com/jackc/pgx/v5"\nfunc Read(tx pgx.Tx){tx.Query(ctx,${JSON.stringify(sql)})}`}]]),providers,owners,defaults,{},false,consumers);
    assert.deepEqual(check('SELECT * FROM record_links'),[]);
    assert.ok(check('SELECT * FROM artifacts').some(v=>v.includes('provider assessment: artifacts outside consumer authority')));
    assert.ok(check('SELECT * FROM record_links FOR UPDATE').some(v=>v.includes('reachable lock capability')));
    assert.ok(check('UPDATE record_links SET deleted_at=NULL').some(v=>v.includes('reachable write capability')));
    assert.ok(check('SELECT * FROM record_links',{shared:'source_owner:links'}).some(v=>v.includes('unknown role')));
    assert.ok(check('SELECT * FROM artifact_grid_projection').some(v=>v.includes('no qualified source SQL coverage')));
  });

  test("projection SQL: shared constants and nested expressions retain physical relations",()=>{
    assert.throws(()=>sqlStatements('package p\nconst expr = "(SELECT * FROM \'unterminated)"'));
    const source='package p\nfunc Read() { _ = `SELECT * FROM (` + base + `) source` }';
    const statements=sqlStatements(source,{goConstants:'const base = `SELECT * FROM public."hosts"`'});
    assert.deepEqual(statements.flatMap(s=>s.accesses),[{table:"hosts",operation:"read"}]);
    for (const expression of ['(SELECT id FROM hosts)', 'COALESCE((SELECT id FROM hosts), 0)']) {
      assert.deepEqual(sqlStatements(`package p\nconst expr = ${JSON.stringify(expression)}`).flatMap(s=>s.accesses),[{table:"hosts",operation:"read"}]);
    }
  });

  test("projection SQL: full expressions retain suffix authority and locking",()=>{
    const pkg="internal/modules/example/provider";
    const providers=[{provider_id:"artifact",source_authority_modules:["artifacts"],projection_table_ids:["artifact_grid_projection"]}];
    const owners=new Map([["artifacts",["artifacts"]],["users",["auth"]]]);
    const inspect=(body)=>inspectGoSQL([[{relative:`${pkg}/constants.go`,content:'package p\nconst base = "SELECT record_id FROM artifacts"'},{relative:`${pkg}/split.go`,content:`package p\nimport "github.com/jackc/pgx/v5"\nfunc Read(tx pgx.Tx){${body}}` }]]);
    for (const [body,reason] of [
      ['q := (base) + " UNION SELECT user_id FROM users";tx.Query(ctx,q)','undeclared source authority auth'],
      ['q := base; q += " FOR UPDATE";tx.Query(ctx,q)','lock of authoritative source'],
      ['q := base; switch choice {case 1:q += " FOR SHARE";default:q += " FOR KEY SHARE"};tx.Query(ctx,q)','lock of authoritative source'],
    ]) assert.ok(evaluateProjectionSQL(inspect(body),providers,owners,{[pkg]:"artifact"},{}).some(message=>message.includes(reason)),body);
    for (const body of ['q:=opaque();tx.Query(ctx,q)','q:=base+opaque();tx.Query(ctx,q)','q:=opaque()+" UNION SELECT user_id FROM users";_=q']) assert.ok(evaluateProjectionSQL(inspect(body),providers,owners,{[pkg]:"artifact"},{}).some(message=>message.includes('unresolved SQL construction')),body);
  });
  test("projection SQL: closed fragments reject dynamic constructors and relation holes",()=>{
    const prefix='package p\nimport shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape"\nimport "github.com/jackc/pgx/v5"\n';
    for (const body of ['shape.Literal(untrusted)','shape.Format("SELECT * FROM %s",shape.Quote(table))','shape.Format("SELECT \'%s\' FROM artifacts",shape.Quote(value))','shape.Text(shape.Literal("SELECT * FROM artifacts"))+" FOR UPDATE"']) assert.throws(()=>sqlStatements(prefix+`func Read(tx pgx.Tx){q:=${body};tx.Query(ctx,q)}`),body);
    assert.deepEqual(sqlStatements(prefix+'func Read(tx pgx.Tx){q:=shape.Literal("SELECT * FROM artifacts");tx.Query(ctx,shape.Text(q))}').flatMap(statement=>statement.accesses),[{table:"artifacts",operation:"read"}]);
    for (const expression of ['shape.Concat(shape.Literal("SELECT * FROM artifacts FOR"),shape.Literal("SHARE"))','shape.Concat(shape.Literal("SELECT *"),shape.Literal("INTO users"),shape.Literal("FROM artifacts"))','shape.Literal("COPY users TO STDOUT")','shape.Literal("CALL hidden_write()")','shape.Concat(shape.Literal("TABLE"),shape.Literal("users"))']) assert.throws(()=>sqlStatements(prefix+`func Read(tx pgx.Tx){tx.Query(ctx,shape.Text(${expression}))}`),expression);
    assert.deepEqual(sqlStatements(prefix+'func Read(tx pgx.Tx){tx.Query(ctx,shape.Text(shape.Literal("TABLE users")))}').flatMap(s=>s.accesses),[{table:'users',operation:'read'}]);
  });

  test("SQL command templates retain writes and reject unresolved relations",()=>{
    const copy='package p\nimport "github.com/jackc/pgx/v5"\nfunc Write(tx pgx.Tx){tx.CopyFrom(ctx,pgx.Identifier{"public","users"},columns,rows)}';
    assert.deepEqual(sqlStatements(copy).flatMap(s=>s.accesses),[{table:'users',operation:'write'}]);
    assert.throws(()=>sqlStatements(copy.replace('"users"','unknown')));

    const prefix='package p\nimport shape "github.com/JochiRaider/cartulary/internal/platform/sqlshape"\nimport "github.com/jackc/pgx/v5"\n';
    assert.deepEqual(sqlStatements(prefix+'func Write(tx pgx.Tx){q:=shape.Command("UPDATE artifacts SET %s=$1",shape.Identifier(column));tx.Exec(ctx,shape.CommandText(q))}').flatMap(s=>s.accesses),[{table:'artifacts',operation:'write'}]);
    for (const body of ['q:=shape.Command("UPDATE %s SET x=1",shape.Identifier(table))','q:=shape.Command(template)','q:=shape.CommandText(shape.Command("DELETE FROM artifacts"))+";DELETE FROM users";tx.Exec(ctx,q)']) assert.throws(()=>sqlStatements(prefix+`func Write(tx pgx.Tx){${body}}`),body);
  });

  test("projection SQL: source authority, physical access and declaration closure", () => {
    const providers = [{provider_id:"host",source_authority_modules:["entities"],projection_table_ids:["host_grid_projection"]}];
    const owners = new Map([["hosts",["entities"]],["records",["records"]],["host_grid_projection",["projections"]],["other_projection",["projections"]]]);
    const pkg = "internal/modules/projections/internal/storage";
    const bindings = {[pkg]:{Read:"host",Write:"host",Compiler:"compiled_projection"}};
    const run = (sql, name="Read", file=`${pkg}/renamed.go`) => evaluateProjectionSQL(inspectGoSQL([[{relative:file,content:`package fixture\nfunc ${name}() { _ = ${JSON.stringify(sql)} }`}]]),providers,owners,{},bindings);
    assert.ok(evaluateProjectionSQL(inspectGoSQL([[{relative:`${pkg}/control.go`,content:'package fixture\nimport "github.com/jackc/pgx/v5"\nfunc Read(tx pgx.Tx){tx.Exec(ctx,"SET LOCAL row_security = off")}'}]]),providers,owners,{},bindings).some(message=>message.includes('session authority')));

    for (const sql of ['SELECT * FROM public."hosts"', 'WITH local AS (SELECT * FROM hosts) SELECT * FROM local', 'UPDATE host_grid_projection SET row_version=1', 'SELECT * FROM host_grid_projection FOR UPDATE']) assert.deepEqual(run(sql),[]);
    for (const [sql,reason] of [['SELECT * FROM records','undeclared source authority'],['UPDATE hosts SET x=1','authoritative source'],['SELECT * FROM hosts FOR UPDATE','authoritative source'],['SELECT * FROM other_projection','undeclared projection table'],['SELECT * FROM missing','schema owner'],['SELECT * FROM %s','unresolved provider relation']]) assert.ok(run(sql).some(x=>x.includes(reason)),sql);
    assert.ok(run('SELECT * FROM hosts','Extracted').some(x=>x.includes('unassigned SQL declaration')));
    assert.deepEqual(run('SELECT * FROM hosts','Read',`${pkg}/split.go`),[]);
    assert.ok(run('SELECT * FROM hosts','Read',`${pkg}/child/bypass.go`).some(x=>x.includes('unassigned SQL declaration')));
    assert.ok(run('SELECT * FROM').length>0);
    const defaults = {[pkg]:"host"};
    assert.deepEqual(evaluateProjectionSQL(inspectGoSQL([[{relative:`${pkg}/new.go`,content:'package fixture\nconst Added = "SELECT * FROM hosts"'}]]),providers,owners,defaults,{}),[]);
    assert.ok(evaluateProjectionSQL(new Map(),providers,owners,{},bindings,true).some(x=>x.includes('stale SQL binding')));
    assert.ok(evaluateProjectionSQL(new Map(),providers,owners).some(x=>x.includes('no SQL declaration coverage')));
  });
}
