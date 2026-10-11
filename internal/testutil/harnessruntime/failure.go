package harnessruntime

import (
	"cmp"
	"math"
	"slices"
	"strings"
)

// PublicFailure exposes only the closed public classification, never private
// paths, credentials or arbitrary helper diagnostics from a boundary error.
func PublicFailure(err error) (string, string) {
	f := failureOf(err)
	return f.Class, f.Reason
}

func normalizedFailure(f *failure) *failure {
	if f == nil {
		return nil
	}
	n := *f
	n.Class = strings.ToLower(strings.TrimSpace(n.Class))
	n.Reason = strings.ToLower(strings.TrimSpace(n.Reason))
	if !slices.Contains(failureReasons, n.Reason) {
		n.Reason = ""
	}
	if n.Class == "" {
		n.Class = failureReasonClasses[n.Reason]
	}
	if !slices.Contains(failureClasses, n.Class) {
		n.Class = "harness"
	}
	if n.Reason == "" {
		n.Reason = failureClassReasons[n.Class]
	}
	return &n
}
func cleanupFailureOf(err error) *failure {
	f := failureOf(err)
	if f.Reason != "cleanup_error" && f.Reason != "artifact_error" && f.Reason != "timeout_failure" {
		f = failed("harness", "cleanup_error")
	}
	f.Class = failureReasonClasses[f.Reason]
	f.Lifecycle = "cleanup_finalizers"
	return f
}
func failureExitCode(f *failure, signal string) int {
	f = normalizedFailure(f)
	if f == nil {
		return 0
	}
	if f.Reason == "cancelled_or_interrupted" {
		if signal == "SIGINT" {
			return 130
		}
		if signal == "SIGTERM" {
			return 143
		}
		return 15
	}
	if code, ok := failureExits[f.Reason]; ok {
		return code
	}
	return 1
}
func rankOf(values []string, value string) int {
	i := slices.Index(values, value)
	if i < 0 {
		return math.MaxInt
	}
	return i
}
func numericOrder(v int64) int64 {
	if v == 0 {
		return math.MaxInt64
	}
	return v
}
func primaryFailure(values []*failure) *failure {
	all := []*failure{}
	nonCleanup := []*failure{}
	for _, v := range values {
		if n := normalizedFailure(v); n != nil {
			all = append(all, n)
			if n.Reason != "cleanup_error" && n.Lifecycle != "cleanup_finalizers" {
				nonCleanup = append(nonCleanup, n)
			}
		}
	}
	if len(all) == 0 {
		return nil
	}
	if len(nonCleanup) == 0 {
		slices.SortStableFunc(all, func(a, b *failure) int {
			return cmp.Or(cmp.Compare(numericOrder(a.Sequence), numericOrder(b.Sequence)), cmp.Compare(numericOrder(a.ChildOrder), numericOrder(b.ChildOrder)))
		})
		return all[0]
	}
	slices.SortStableFunc(nonCleanup, func(a, b *failure) int {
		return cmp.Or(cmp.Compare(rankOf(failureClasses, a.Class), rankOf(failureClasses, b.Class)), cmp.Compare(rankOf(failureLifecycle, a.Lifecycle), rankOf(failureLifecycle, b.Lifecycle)), cmp.Compare(numericOrder(a.Sequence), numericOrder(b.Sequence)), cmp.Compare(numericOrder(a.ChildOrder), numericOrder(b.ChildOrder)), strings.Compare(a.Artifact, b.Artifact), strings.Compare(a.Reason, b.Reason))
	})
	return nonCleanup[0]
}
