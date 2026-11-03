package extensions

import (
	"bytes"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func pathIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for i, ch := range []byte(value) {
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || ch == '_' || (i > 0 && ch >= '0' && ch <= '9') {
			continue
		}
		return false
	}
	return true
}

func validDiagnosticPath(path string) bool {
	if path == "" || path[0] != '$' {
		return false
	}
	for offset := 1; offset < len(path); {
		switch path[offset] {
		case '.':
			start := offset + 1
			offset = start
			for offset < len(path) && path[offset] != '.' && path[offset] != '[' {
				offset++
			}
			if !pathIdentifier(path[start:offset]) {
				return false
			}
		case '[':
			offset++
			if offset >= len(path) {
				return false
			}
			if path[offset] == '"' {
				start := offset
				offset++
				for offset < len(path) && path[offset] != '"' {
					if path[offset] == '\\' {
						offset++
					}
					offset++
				}
				if offset >= len(path) {
					return false
				}
				offset++
				raw := []byte(path[start:offset])
				value, err := canonicaljson.DecodeStrict(raw)
				name, ok := value.(string)
				if err != nil || !ok || pathIdentifier(name) {
					return false
				}
				canonical, err := canonicalJSON(name, false)
				if err != nil || !bytes.Equal(raw, canonical) {
					return false
				}
			} else {
				start := offset
				for offset < len(path) && path[offset] >= '0' && path[offset] <= '9' {
					offset++
				}
				if offset == start || (offset-start > 1 && path[start] == '0') {
					return false
				}
			}
			if offset >= len(path) || path[offset] != ']' {
				return false
			}
			offset++
		default:
			return false
		}
	}
	return !strings.ContainsRune(path, '\x00')
}
