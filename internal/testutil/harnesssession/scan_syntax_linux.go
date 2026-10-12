//go:build linux

package harnesssession

import (
	"regexp"
	"regexp/syntax"
	"strings"
	"unicode/utf8"
)

// Expand the fixed negative-prefix predicate into a regular language. Short
// prefixes are secrets too; only a complete, case-sensitive <redacted> prefix
// exempts a value. The terminator is part of the syntax, never a read boundary.
func unredactedValue(excluded, terminator string) string {
	const marker = "<redacted>"
	parts := []string{}
	for i := range len(marker) {
		prefix := regexp.QuoteMeta(marker[:i])
		parts = append(parts, prefix+"[^"+excluded+regexp.QuoteMeta(marker[i:i+1])+"][^"+excluded+"]*"+terminator)
		if i > 0 {
			parts = append(parts, prefix+terminator)
		}
	}
	return "(?:" + strings.Join(parts, "|") + ")"
}

var secretProgram = func() *syntax.Prog {
	key := `(?i:"(?:[^"]*(?:password|secret|cookie|dsn|access_key|private_key)[^"]*|[^"]*token(?:[^s"][^"]*)?)"\s*:\s*")`
	json := key + unredactedValue(`"`, `"`)
	assignment := `(?:PASSWORD|SECRET|TOKEN|COOKIE|DSN|ACCESS_KEY|PRIVATE_KEY)=` + unredactedValue(`\s`, `(?:\s|$)`)
	postgres := `(?i:(?:postgres|postgresql)://[^\s/:]+:)` + unredactedValue(`\s/@`, `@`)
	tree, err := syntax.Parse(json+"|"+assignment+"|"+postgres, syntax.Perl)
	if err != nil {
		panic(err)
	}
	program, err := syntax.Compile(tree.Simplify())
	if err != nil {
		panic(err)
	}
	return program
}()

// An unanchored Thompson matcher retains only program counters and at most one
// incomplete UTF-8 rune. Memory is bounded by the fixed policy, not input length.
// Standard regexp matching operates on complete strings and cannot retain this
// state across Reader calls. Each instruction is visited at most once per rune.
type secretMatcher struct {
	active, next []uint32
	seen         []uint64
	generation   uint64
	previous     rune
	partial      []byte
}

func newSecretMatcher() *secretMatcher {
	n := len(secretProgram.Inst)
	return &secretMatcher{active: make([]uint32, 0, n), next: make([]uint32, 0, n), seen: make([]uint64, n), previous: -1}
}

func (m *secretMatcher) transition(pc uint32, r rune, context syntax.EmptyOp) bool {
	if m.seen[pc] == m.generation {
		return false
	}
	m.seen[pc] = m.generation
	i := &secretProgram.Inst[pc]
	switch i.Op {
	case syntax.InstMatch:
		return true
	case syntax.InstAlt, syntax.InstAltMatch:
		return m.transition(i.Out, r, context) || m.transition(i.Arg, r, context)
	case syntax.InstCapture, syntax.InstNop:
		return m.transition(i.Out, r, context)
	case syntax.InstEmptyWidth:
		if syntax.EmptyOp(i.Arg)&context == syntax.EmptyOp(i.Arg) {
			return m.transition(i.Out, r, context)
		}
	case syntax.InstRune, syntax.InstRune1, syntax.InstRuneAny, syntax.InstRuneAnyNotNL:
		if r >= 0 && i.MatchRune(r) {
			m.next = append(m.next, i.Out)
		}
	}
	return false
}

func (m *secretMatcher) step(r rune) bool {
	m.generation++
	m.next = m.next[:0]
	context := syntax.EmptyOpContext(m.previous, r)
	if m.transition(uint32(secretProgram.Start), r, context) {
		return true
	}
	for _, pc := range m.active {
		if m.transition(pc, r, context) {
			return true
		}
	}
	m.active, m.next = m.next, m.active
	m.previous = r
	return false
}

func (m *secretMatcher) feed(data []byte, final bool) bool {
	if len(m.partial) > 0 {
		data = append(m.partial, data...)
		m.partial = nil
	}
	for len(data) > 0 {
		if !final && !utf8.FullRune(data) {
			m.partial = append([]byte{}, data...)
			break
		}
		r, n := utf8.DecodeRune(data)
		if m.step(r) {
			return true
		}
		data = data[n:]
	}
	return final && m.step(-1)
}
