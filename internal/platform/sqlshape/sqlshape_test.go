package sqlshape

import "testing"

func TestClosedSQLFragments(t *testing.T) {
	t.Run("templates reject lexical escapes", func(t *testing.T) {
		for _, template := range []string{
			"SELECT 1 FROM artifacts /* %s */", "SELECT 1 -- %s", "SELECT /* outer /* %s */ end */ 1",
			"SELECT '%s'", "SELECT E'%s'", `SELECT "%s"`, "SELECT $$%s$$", "SELECT $tag$%s$tag$",
			"SELECT prefix%s", "SELECT %ssuffix", "SELECT $%s", "SELECT '%s", "SELECT 'unterminated",
		} {
			for _, command := range []bool{false, true} {
				func() {
					defer func() {
						if recover() == nil {
							t.Errorf("accepted %q command=%v", template, command)
						}
					}()
					if command {
						Command(template, Quote("*/ UNION SELECT 1 FROM users --"))
					} else {
						Format(template, Quote("*/ UNION SELECT 1 FROM users --"))
					}
				}()
			}
		}
		for _, template := range []string{"SELECT 1 /* comment */", "SELECT 1 -- comment", "SELECT 'unterminated", "SELECT $$nul\x00$$", "SELECT E'\\\x00'"} {
			func() {
				defer func() {
					if recover() == nil {
						t.Errorf("literal accepted %q", template)
					}
				}()
				Literal(template)
			}()
		}
	})
	t.Run("quoted data and punctuation preserve boundaries", func(t *testing.T) {
		for _, template := range []string{`SELECT '-- /* %s */'`, `SELECT "--"`, `SELECT $$/* */$$`, `SELECT E'\\'`} {
			if Text(Literal(template)) != template {
				t.Fatal(template)
			}
		}
		got := Text(Format("SELECT %s FROM artifacts WHERE id=ANY(%s)", Quote("*/ UNION SELECT 1 FROM users --"), Parameter(1)))
		if got != "SELECT E'*/ UNION SELECT 1 FROM users --' FROM artifacts WHERE id=ANY( $1 )" {
			t.Fatal(got)
		}
	})

	for _, name := range []string{"=;DELETE FROM records", "--", "/*", "*/", "", "word"} {
		if _, err := Operator("pg_catalog", name); err == nil {
			t.Fatalf("accepted operator %q", name)
		}
	}
	operator, err := Operator(`pg"catalog`, "=")
	if err != nil || Text(operator) != `OPERATOR("pg""catalog".=)` {
		t.Fatalf("operator quoting: %q %v", Text(operator), err)
	}
	if got := Text(Join([]Fragment{Identifier("a"), Identifier("b")}, Literal(","))); got != `"a" , "b"` {
		t.Fatal(got)
	}

	if got := CommandText(Command("UPDATE artifacts SET %s = $1", Identifier(`a"b`))); got != `UPDATE artifacts SET "a""b" = $1` {
		t.Fatal(got)
	}
	if got := Text(Format("SELECT %s FROM records WHERE record_id = %s", Literal("row_version"), Parameter(2))); got != "SELECT row_version FROM records WHERE record_id = $2" {
		t.Fatal(got)
	}
	// Independent fragments must not manufacture identifiers or comment tokens.
	if got := Text(Concat(Literal("SEL"), Literal("ECT"), Literal("-"), Literal("-"))); got != "SEL ECT - -" {
		t.Fatal(got)
	}
	if got := Text(Quote("x'\\y")); got != `E'x''\\y'` {
		t.Fatal(got)
	}
	if !((Fragment{}).Empty()) {
		t.Fatal("zero fragment must be empty")
	}
	for _, run := range []func(){func() { Identifier("x\x00y") }, func() { Parameter(0) }, func() { Format("%s %s", Literal("one")) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid construction accepted")
				}
			}()
			run()
		}()
	}
}
