package canonicaljson

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf8"
)

// TokenValue is an immutable view into a bounded caller-owned document. Child
// traversal never constructs a JSON tree. Views must not outlive the document.
type TokenValue struct{ raw []byte }

func (v TokenValue) Raw() []byte { return v.raw }
func (v TokenValue) Kind() byte {
	if len(v.raw) == 0 {
		return 0
	}
	return v.raw[0]
}

// ScanTokens validates the complete original document, with bounded nesting,
// before exposing views. Canonical and SafeIntegers are independent findings:
// callers can still enumerate shape defects in syntactically valid input.
type TokenDocument struct {
	Value                   TokenValue
	Canonical, SafeIntegers bool
}

func ScanTokens(ctx context.Context, raw []byte, maximumDepth int) (TokenDocument, error) {
	if err := ctx.Err(); err != nil {
		return TokenDocument{}, err
	}
	// Bound scanner stack growth before encoding/json's syntax pass. Strings
	// are skipped lexically; syntax and Unicode are checked independently below.
	depth, quoted := 0, false
	for i := 0; i < len(raw); i++ {
		if i&4095 == 0 {
			if err := ctx.Err(); err != nil {
				return TokenDocument{}, err
			}
		}
		b := raw[i]
		if quoted {
			if b == '\\' {
				i++
			} else if b == '"' {
				quoted = false
			}
			continue
		}
		switch b {
		case '"':
			quoted = true
		case '{', '[':
			depth++
			if depth > maximumDepth {
				return TokenDocument{}, ErrDepth
			}
		case '}', ']':
			depth--
		}
	}
	if !utf8.Valid(raw) || !validStringEscapes(raw) {
		return TokenDocument{}, ErrInvalidUnicode
	}
	if !json.Valid(raw) {
		return TokenDocument{}, ErrInvalidJSON
	}
	s := tokenScanner{ctx: ctx, raw: raw, maximum: maximumDepth, canonical: true, integers: true}
	start := s.space()
	if err := s.value(0); err != nil {
		return TokenDocument{}, err
	}
	end := s.at
	s.space()
	return TokenDocument{TokenValue{raw[start:end]}, s.canonical, s.integers}, ctx.Err()
}

type tokenScanner struct {
	ctx                 context.Context
	raw                 []byte
	at, maximum         int
	canonical, integers bool
}

func (s *tokenScanner) space() int {
	for s.at < len(s.raw) && strings.ContainsRune(" \t\r\n", rune(s.raw[s.at])) {
		s.at++
		s.canonical = false
	}
	return s.at
}
func (s *tokenScanner) quoted(decode bool) (string, error) {
	start := s.at
	s.at++
	escaped := false
	for s.at < len(s.raw) {
		if s.at&4095 == 0 {
			if err := s.ctx.Err(); err != nil {
				return "", err
			}
		}
		b := s.raw[s.at]
		s.at++
		if b == '\\' {
			escaped = true
			s.at++
			continue
		}
		if b == '"' {
			break
		}
	}
	if !decode && !escaped {
		return "", nil
	}
	var value string
	if json.Unmarshal(s.raw[start:s.at], &value) != nil {
		return "", ErrInvalidJSON
	}
	if escaped {
		var canonical bytes.Buffer
		if appendString(&canonical, value) != nil {
			return "", ErrInvalidUnicode
		}
		if !bytes.Equal(canonical.Bytes(), s.raw[start:s.at]) {
			s.canonical = false
		}
	}
	return value, nil
}
func (s *tokenScanner) value(depth int) error {
	if err := s.ctx.Err(); err != nil {
		return err
	}
	s.space()
	switch s.raw[s.at] {
	case '{', '[':
		if depth >= s.maximum {
			return ErrDepth
		}
		object := s.raw[s.at] == '{'
		end := byte(']')
		if object {
			end = '}'
		}
		s.at++
		s.space()
		first := true
		previous := ""
		for s.raw[s.at] != end {
			if !first {
				s.at++
				s.space()
			}
			if object {
				key, err := s.quoted(true)
				if err != nil {
					return err
				}
				if !first && compareUTF16(previous, key) >= 0 {
					s.canonical = false
				}
				previous = key
				s.space()
				s.at++
			}
			if err := s.value(depth + 1); err != nil {
				return err
			}
			s.space()
			first = false
		}
		s.at++
	case '"':
		_, err := s.quoted(false)
		return err
	case 'n':
		s.at += 4
	case 't':
		s.at += 4
	case 'f':
		s.at += 5
	default:
		start := s.at
		for s.at < len(s.raw) && !strings.ContainsRune(",]} \t\r\n", rune(s.raw[s.at])) {
			s.at++
		}
		text := string(s.raw[start:s.at])
		n, err := strconv.ParseInt(text, 10, 64)
		if err != nil || n < -9007199254740991 || n > 9007199254740991 || text == "-0" {
			s.integers = false
		}
		canonical, err := canonicalNumber(json.Number(text))
		if err != nil || canonical != text {
			s.canonical = false
		}
	}
	return nil
}

// Children visits direct array values or object members without retaining any
// siblings. The callback chooses which bounded values to materialize.
func (v TokenValue) Children(ctx context.Context, visit func(string, int, TokenValue) error) error {
	if v.Kind() != '{' && v.Kind() != '[' {
		return ErrInvalidJSON
	}
	raw := v.raw
	at := 1
	index := 0
	space := func() {
		for at < len(raw) && strings.ContainsRune(" \t\r\n", rune(raw[at])) {
			at++
		}
	}
	space()
	for at < len(raw)-1 {
		if err := ctx.Err(); err != nil {
			return err
		}
		name := ""
		if v.Kind() == '{' {
			start := at
			var err error
			at, err = tokenEnd(ctx, raw, at)
			if err != nil {
				return err
			}
			if json.Unmarshal(raw[start:at], &name) != nil {
				return ErrInvalidJSON
			}
			space()
			at++
			space()
		}
		start := at
		var err error
		at, err = tokenEnd(ctx, raw, at)
		if err != nil {
			return err
		}
		if err := visit(name, index, TokenValue{raw[start:at]}); err != nil {
			return err
		}
		index++
		space()
		if at < len(raw)-1 {
			at++
			space()
		}
	}
	return ctx.Err()
}

// Input has already passed ScanTokens. Find a direct child's extent without
// decoding strings, collecting members, or allocating descendant containers.
func tokenEnd(ctx context.Context, raw []byte, at int) (int, error) {
	depth := 0
	quoted := false
	for at < len(raw) {
		if at&4095 == 0 {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
		}
		b := raw[at]
		if quoted {
			at++
			if b == '\\' {
				at++
			} else if b == '"' {
				quoted = false
				if depth == 0 {
					return at, nil
				}
			}
			continue
		}
		switch b {
		case '"':
			quoted = true
		case '{', '[':
			depth++
		case '}', ']':
			if depth == 0 {
				return at, nil
			}
			depth--
			if depth == 0 {
				return at + 1, nil
			}
		case ',', ' ', '\t', '\n', '\r':
			if depth == 0 {
				return at, nil
			}
		}
		at++
	}
	return at, nil
}
