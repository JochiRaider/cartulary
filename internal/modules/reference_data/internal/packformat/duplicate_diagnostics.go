package packformat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"
)

// duplicateManifestFindings runs only after strict admission has established
// valid complete JSON and Unicode but rejected duplicate members. It retains
// no repaired object. The manifest byte limit bounds token and key storage.
func duplicateManifestFindings(ctx context.Context, data []byte, emit FindingSink) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return duplicateValue(ctx, decoder, manifestShape, "$", emit)
}

func diagnosticProperty(s *shape, name string) *shape {
	if s == nil {
		return nil
	}
	if child := s.properties[name]; child != nil {
		return child
	}
	for _, alternatives := range [][]*shape{s.anyOf, s.oneOf} {
		for _, alternative := range alternatives {
			if child := diagnosticProperty(alternative, name); child != nil {
				return child
			}
		}
	}
	return nil
}

func duplicateValue(ctx context.Context, decoder *json.Decoder, s *shape, path string, emit FindingSink) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, composite := token.(json.Delim)
	if !composite {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			token, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := token.(string)
			if !ok {
				return errors.New("reference pack: duplicate scan lost admitted syntax")
			}
			child, childPath := diagnosticProperty(s, name), path
			if child != nil {
				childPath = diagnosticMember(path, name)
			} else if s != nil {
				childPath += `["<unknown>"]`
			}
			// An undeclared subtree collapses to its first unknown segment.
			// Hostile nesting cannot grow diagnostic paths or expose its keys.
			if seen[name] {
				if err := emit(Finding{Path: childPath}); err != nil {
					return err
				}
			}
			seen[name] = true
			if err := duplicateValue(ctx, decoder, child, childPath, emit); err != nil {
				return err
			}
		}
	case '[':
		for i := 0; decoder.More(); i++ {
			child, childPath := (*shape)(nil), path
			if s != nil && s.items != nil {
				child, childPath = s.items, path+"["+strconv.Itoa(i)+"]"
			}
			if err := duplicateValue(ctx, decoder, child, childPath, emit); err != nil {
				return err
			}
		}
	default:
		return errors.New("reference pack: duplicate scan lost admitted delimiter")
	}
	_, err = decoder.Token()
	return err
}
