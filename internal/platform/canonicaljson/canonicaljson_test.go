package canonicaljson_test

import (
	"math"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestRFC8785SerializationSample(t *testing.T) {
	t.Parallel()
	input := []byte(`{"numbers":[333333333.33333329,1E30,4.50,2e-3,0.000000000000000000000000001],"string":"€$\u000f\nA'B\\\"/","literals":[null,true,false]}`)
	want := `{"literals":[null,true,false],"numbers":[333333333.3333333,1e+30,4.5,0.002,1e-27],"string":"€$\u000f\nA'B\\\"/"}`
	got, err := canonicaljson.Canonicalize(input)
	if err != nil {
		t.Fatalf("canonicalize RFC 8785 sample: %v", err)
	}
	if string(got) != want {
		t.Fatalf("canonical JSON = %s, want %s", got, want)
	}
}

func TestRFC8785UsesUTF16PropertyOrdering(t *testing.T) {
	t.Parallel()
	input := []byte(`{"\ue000":1,"😀":2,"a":3}`)
	want := `{"a":3,"😀":2,"":1}`
	got, err := canonicaljson.Canonicalize(input)
	if err != nil {
		t.Fatalf("canonicalize UTF-16 ordering sample: %v", err)
	}
	if string(got) != want {
		t.Fatalf("canonical JSON = %s, want %s", got, want)
	}
}

func TestRFC8785RejectsNonFiniteValues(t *testing.T) {
	t.Parallel()
	if _, err := canonicaljson.Marshal(struct {
		Number float64 `json:"number"`
	}{Number: math.Inf(1)}); err == nil {
		t.Fatal("expected non-finite value rejection")
	}
}
