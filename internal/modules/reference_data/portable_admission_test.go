package reference_data

import (
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

func testPortableVerificationOrder(t *testing.T) {
	versions := make([]importedReferenceVersion, 5)
	for i, key := range []string{"a", "b", "c", "d", "e"} {
		versions[i].reference = IncidentBundleVersionReference{PackSetMember: PackSetMember{Key: key, Version: "1", PayloadSHA256: "digest"}}
	}
	for _, test := range []struct {
		name  string
		edges map[int][]int
		want  []int
	}{
		{"stable roots", nil, []int{0, 1, 2, 3, 4}},
		{"reverse chain", map[int][]int{0: {1}, 1: {2}, 2: {3}, 3: {4}}, []int{4, 3, 2, 1, 0}},
		{"stable newly ready", map[int][]int{0: {2}, 1: {3}}, []int{2, 0, 3, 1, 4}},
		{"residual cycle and its dependants", map[int][]int{0: {1}, 1: {0}, 2: {0}}, []int{3, 4, 0, 1, 2}},
		{"self dependency", map[int][]int{1: {1}}, []int{0, 2, 3, 4, 1}},
		{"duplicate lexical declarations", map[int][]int{0: {2, 2}}, []int{1, 2, 0, 3, 4}},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, insert := range [][]int{{0, 1, 2, 3, 4}, {4, 3, 2, 1, 0}, {2, 4, 0, 3, 1}} {
				inputs := map[int]portableContainerInput{}
				for _, i := range insert {
					input := portableContainerInput{}
					for _, j := range test.edges[i] {
						m := versions[j].reference
						input.identity.Dependencies = append(input.identity.Dependencies, packformat.Dependency{Key: m.Key, Version: m.Version, SHA256: m.PayloadSHA256})
					}
					// A different payload cannot create a scheduling dependency.
					input.identity.Dependencies = append(input.identity.Dependencies, packformat.Dependency{Key: "a", Version: "1", SHA256: "other"})
					inputs[i] = input
				}
				if got := portableVerificationOrder(versions, inputs); !slices.Equal(got, test.want) {
					t.Fatal("schedule", got, test.want)
				}
			}
		})
	}
}
