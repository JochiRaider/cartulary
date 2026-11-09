package dataset

import (
	"context"
	"encoding/json"
	"strings"

	"golang.org/x/text/unicode/norm"
)

func decodeObjectContext(ctx context.Context, raw []byte) (Object, error) {
	var object Object
	p, err := materializeRow(ctx, raw, "object", &object)
	if err != nil {
		return Object{}, err
	}
	object.Data = p.data
	object.LookupValues = p.lookups
	object.inputPostings = p.postings
	return object, nil
}
func decodeRelationshipContext(ctx context.Context, raw []byte) (Relationship, error) {
	var relationship Relationship
	p, err := materializeRow(ctx, raw, "relationship", &relationship)
	if err != nil {
		return Relationship{}, err
	}
	relationship.Data = p.data
	return relationship, nil
}
func decodeTyped(raw []byte, kind string, target any) error {
	m, err := document(raw)
	if err != nil {
		return err
	}
	if err := datasetShapes[kind].Walk(m, "$", firstFailure("content_schema_invalid")); err != nil {
		return err
	}
	if json.Unmarshal(raw, target) != nil {
		return invalid()
	}
	return nil
}

func line(s string) bool {
	return s != "" && norm.NFC.IsNormalString(s) && strings.TrimSpace(s) == s && lookupString(s)
}
func description(s *string) bool {
	if s == nil {
		return true
	}
	if *s == "" || !norm.NFC.IsNormalString(*s) || strings.TrimSpace(*s) != *s {
		return false
	}
	for _, r := range *s {
		if r < 32 && r != '\n' && r != '\t' || r >= 127 && r <= 159 {
			return false
		}
	}
	return true
}
