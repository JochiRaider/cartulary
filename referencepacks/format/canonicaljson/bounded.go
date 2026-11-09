package canonicaljson

import "errors"

var ErrDepth = errors.New("canonicaljson: nesting limit exceeded")

// DecodeStrictBounded checks nesting before constructing any JSON subtree.
// The outermost object or array has depth one; scalars do not add depth.
// Syntax, Unicode and duplicate-member admission still belong to DecodeStrict.
func DecodeStrictBounded(document []byte, maximum int) (any, error) {
	if maximum < 1 {
		return nil, ErrDepth
	}
	depth, quoted, escaped := 0, false, false
	for _, b := range document {
		if quoted {
			if escaped {
				escaped = false
			} else if b == '\\' {
				escaped = true
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
			if depth > maximum {
				return nil, ErrDepth
			}
		case '}', ']':
			depth--
		}
	}
	return DecodeStrict(document)
}
