package producer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

// Candidate artifacts are producer-only proposals. No runtime registry loads
// this index, and passing these schemas does not establish pack conformance.
type CandidateIndex struct {
	SchemaID    string         `json:"schema_id"`
	CandidateID string         `json:"candidate_id"`
	Status      string         `json:"status"`
	Artifacts   []EvidenceFile `json:"artifacts"`
}

type CandidateValidation struct {
	Status             string   `json:"status"`
	IndexSHA256        string   `json:"candidate_index_sha256"`
	SourceSchemaSHA256 string   `json:"source_schema_sha256"`
	Records            int      `json:"validated_source_records"`
	Issues             []string `json:"issues"`
}

func ReadCandidate(root *os.Root, indexBytes []byte) (map[string][]byte, error) {
	var index CandidateIndex
	if err := DecodeClosed(indexBytes, &index); err != nil {
		return nil, err
	}
	if index.SchemaID != "cartulary.attack_candidate_index.v1" || index.CandidateID != "cartulary.attack.all-domains.native.candidate-1" || index.Status != "proposal_only" || len(index.Artifacts) == 0 {
		return nil, errors.New("candidate identity invalid")
	}
	files := map[string][]byte{}
	last := ""
	for _, item := range index.Artifacts {
		if !strings.HasSuffix(item.Path, ".json") || item.Path <= last {
			return nil, errors.New("candidate inventory must be sorted JSON inputs")
		}
		last = item.Path
		b, err := ReadEvidence(root, item)
		if err != nil {
			return nil, err
		}
		files[item.Path] = b
	}
	for _, name := range []string{"field-policy.candidate.v1.json", "source-profile.candidate.v1.json", "native-source.candidate.v1.schema.json", "object.candidate.v1.schema.json", "relationship.candidate.v1.schema.json"} {
		if len(files[name]) == 0 {
			return nil, errors.New("candidate artifact missing")
		}
	}
	return files, nil
}

func CompileCandidate(b []byte, fragment string) (*jsonschema.Schema, error) {
	v, err := canonicaljson.DecodeStrict(b)
	if err != nil {
		return nil, err
	}
	c := jsonschema.NewCompiler()
	c.UseLoader(jsonschema.SchemeURLLoader{})
	c.DefaultDraft(jsonschema.Draft2020)
	c.AssertFormat()
	const location = "https://cartulary.invalid/proposals/input"
	if err = c.AddResource(location, v); err != nil {
		return nil, err
	}
	return c.Compile(location + fragment)
}

func ValidateNativeCandidate(ctx context.Context, root *os.Root, lockBytes, indexBytes, schemaBytes []byte) (CandidateValidation, error) {
	r := CandidateValidation{Status: "proposal_shape_failed", IndexSHA256: Digest(indexBytes), SourceSchemaSHA256: Digest(schemaBytes), Issues: []string{}}
	a, err := Check(root, lockBytes)
	if err != nil {
		return r, err
	}
	l, err := ReadLock(lockBytes)
	if err != nil {
		return r, err
	}
	compiled := map[string]*jsonschema.Schema{}
	for i, source := range l.Artifacts {
		if source.Domain == "license" {
			continue
		}
		if err := ctx.Err(); err != nil {
			return r, err
		}
		b, err := ReadEvidence(root, a.Sources[i])
		if err != nil {
			return r, err
		}
		v, err := canonicaljson.DecodeStrict(b)
		if err != nil {
			return r, err
		}
		bundle, ok := v.(map[string]any)
		if !ok {
			return r, errors.New("native candidate bundle malformed")
		}
		objects, ok := bundle["objects"].([]any)
		if !ok {
			return r, errors.New("native candidate objects missing")
		}
		for j, item := range objects {
			o, ok := item.(map[string]any)
			if !ok {
				return r, errors.New("native candidate source is not an object")
			}
			typ := text(o, "type")
			schema, ok := compiled[typ]
			if !ok {
				schema, err = CompileCandidate(schemaBytes, "#/$defs/"+pointerToken(typ))
				if err != nil {
					return r, err
				}
				compiled[typ] = schema
			}
			if err = schema.Validate(item); err != nil {
				r.Issues = append(r.Issues, fmt.Sprintf("%s:/objects/%d:%v", source.Name, j, err))
			}
			r.Records++
		}
	}
	if len(r.Issues) > 0 {
		r.Status = "proposal_shape_failed"
		return r, errors.New("native candidate schema validation failed")
	}
	r.Status = "proposal_shape_passed"
	return r, nil
}
