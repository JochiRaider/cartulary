//go:build linux

package harnesssession

import (
	"encoding/binary"
	"regexp"
	"regexp/syntax"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/harnessredact"
)

// Expand the fixed negative-prefix predicate into a regular language. Short
// prefixes are secrets too; only a complete, case-sensitive prefix of the
// manifest-owned marker exempts a value (TH-HARNESS-REQ-617). The terminator
// is part of the syntax, never a read boundary.
func unredactedValue(excluded, terminator string) string {
	const marker = harnessredact.Marker
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
	alternatives := make([]string, 0, len(harnessredact.RetainedScanContexts))
	for _, scanContext := range harnessredact.RetainedScanContexts {
		alternatives = append(alternatives, scanContext.Prefix+unredactedValue(scanContext.Excluded, scanContext.Terminator))
	}
	tree, err := syntax.Parse(strings.Join(alternatives, "|"), syntax.Perl)
	if err != nil {
		panic(err)
	}
	program, err := syntax.Compile(tree.Simplify())
	if err != nil {
		panic(err)
	}
	return program
}()

// End-of-text and end-of-line assertions depend only on the current rune (end
// of input or a newline). When the program has no other assertion, a step
// depends only on the pending program counters and the current rune, so the
// Thompson steps can be memoized as a lazily built DFA whose states are exactly
// those counter sets. Other assertions also read the previous rune.
const memoizableEmptyOps = syntax.EmptyEndText | syntax.EmptyEndLine

var secretDFAEligible = func() bool {
	for _, instruction := range secretProgram.Inst {
		if instruction.Op == syntax.InstEmptyWidth && syntax.EmptyOp(instruction.Arg)&^memoizableEmptyOps != 0 {
			return false
		}
	}
	return true
}()

// The memo is per matcher and bounded; reaching the bound clears it, which only
// costs recomputation because every transition is derived from the program.
const secretDFAStateLimit = 2048

type secretState struct {
	pcs   []uint32
	ascii [utf8.RuneSelf]*secretEdge
	other map[rune]*secretEdge
}

type secretEdge struct {
	to    *secretState
	match bool
}

// An unanchored Thompson matcher retains only program counters and at most one
// incomplete UTF-8 rune. Memory is bounded by the fixed policy, not input length.
// Standard regexp matching operates on complete strings and cannot retain this
// state across Reader calls. Each instruction is visited at most once per rune.
// When the program is eligible, each Thompson step is computed once per
// (counter set, rune) and then replayed from the memo with identical verdicts.
type secretMatcher struct {
	active, next []uint32
	seen         []uint64
	generation   uint64
	previous     rune
	partial      []byte
	states       map[string]*secretState
	state        *secretState
}

func newSecretMatcher() *secretMatcher {
	n := len(secretProgram.Inst)
	m := &secretMatcher{active: make([]uint32, 0, n), next: make([]uint32, 0, n), seen: make([]uint64, n), previous: -1}
	if secretDFAEligible {
		m.states = map[string]*secretState{}
		m.state = m.intern(nil)
	}
	return m
}

func (m *secretMatcher) intern(pcs []uint32) *secretState {
	key := make([]byte, 4*len(pcs))
	for i, pc := range pcs {
		binary.LittleEndian.PutUint32(key[4*i:], pc)
	}
	if state := m.states[string(key)]; state != nil {
		return state
	}
	if len(m.states) >= secretDFAStateLimit {
		m.states = map[string]*secretState{}
	}
	state := &secretState{pcs: pcs}
	m.states[string(key)] = state
	return state
}

// memoize runs one Thompson step from the current counter set and records it.
// Only rune-determined assertions exist, so the context derives from r alone.
func (m *secretMatcher) memoize(r rune) *secretEdge {
	from := m.state
	context := syntax.EmptyOpContext(-1, r) & memoizableEmptyOps
	m.generation++
	m.next = m.next[:0]
	match := m.transition(uint32(secretProgram.Start), r, context)
	for _, pc := range from.pcs {
		if match {
			break
		}
		match = m.transition(pc, r, context)
	}
	edge := &secretEdge{match: match}
	if !match {
		pcs := slices.Clone(m.next)
		slices.Sort(pcs)
		edge.to = m.intern(slices.Compact(pcs))
	}
	if r >= 0 && r < utf8.RuneSelf {
		from.ascii[r] = edge
	} else {
		if from.other == nil {
			from.other = map[rune]*secretEdge{}
		}
		from.other[r] = edge
	}
	return edge
}

func (m *secretMatcher) advance(r rune) bool {
	var edge *secretEdge
	if r >= 0 && r < utf8.RuneSelf {
		edge = m.state.ascii[r]
	} else {
		edge = m.state.other[r]
	}
	if edge == nil {
		edge = m.memoize(r)
	}
	if edge.match {
		return true
	}
	m.state = edge.to
	return false
}

func (m *secretMatcher) consume(r rune) bool {
	if m.states != nil {
		return m.advance(r)
	}
	return m.step(r)
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
		if m.states != nil && data[0] < utf8.RuneSelf {
			if m.advance(rune(data[0])) {
				return true
			}
			data = data[1:]
			continue
		}
		if !final && !utf8.FullRune(data) {
			m.partial = append([]byte{}, data...)
			break
		}
		r, n := utf8.DecodeRune(data)
		if m.consume(r) {
			return true
		}
		data = data[n:]
	}
	return final && m.consume(-1)
}
