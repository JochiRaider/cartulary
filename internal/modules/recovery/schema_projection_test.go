package recovery

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	contractrecovery "github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
	"github.com/dlclark/regexp2"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// AssertCurrentArtifactSchema validates actual capture bytes against their
// packaged schema. It reads no source-tree files or documentary requirements.
func AssertCurrentArtifactSchema(t testing.TB, body []byte) {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(body, &value); err != nil {
		t.Fatal(err)
	}
	id, ok := value["schema_id"].(string)
	if !ok {
		t.Fatal("captured artifact has no schema identity")
	}
	compiler := jsonschema.NewCompiler()
	compiler.UseRegexpEngine(func(pattern string) (jsonschema.Regexp, error) {
		expression, err := regexp2.Compile(pattern, regexp2.ECMAScript)
		if err != nil {
			return nil, err
		}
		expression.MatchTimeout = time.Second
		return recoverySchemaRegexp{expression}, nil
	})
	selected := ""
	for _, artifact := range contractrecovery.Artifacts {
		if !strings.HasSuffix(artifact.Path, ".schema.json") {
			continue
		}
		var schema map[string]any
		if err := json.Unmarshal([]byte(artifact.JSON), &schema); err != nil {
			t.Fatal(err)
		}
		location := "https://recovery.invalid/" + strings.TrimPrefix(artifact.Path, "contracts/recovery/")
		if err := compiler.AddResource(location, schema); err != nil {
			t.Fatal(err)
		}
		if schema["$id"] == id {
			selected = location
		}
	}
	if selected == "" {
		t.Fatalf("captured schema %s is not packaged", id)
	}
	schema, err := compiler.Compile(selected)
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(value); err != nil {
		t.Fatalf("captured %s violates its packaged schema: %v", id, err)
	}
}

type recoverySchemaRegexp struct{ *regexp2.Regexp }

func (expression recoverySchemaRegexp) MatchString(value string) bool {
	matched, err := expression.Regexp.MatchString(value)
	return err == nil && matched
}
