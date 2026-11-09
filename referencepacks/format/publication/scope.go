package publication

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type AuthorityClass string

const ReleaseAuthority AuthorityClass = "release"
const FixtureAuthority AuthorityClass = "fixture"

// Scope is created explicitly by each producer's fixed release or fixture path.
// Its zero value is invalid; a fixture scope cannot authorize release output.
type Scope struct {
	repository, key string
	class           AuthorityClass
}

var repositoryPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}(\.[a-z][a-z0-9_]{0,31}){1,7}$`)

func ReleaseScope(repository, key string) (Scope, error) {
	return scope(repository, key, ReleaseAuthority)
}
func FixtureScope(repository, key string) (Scope, error) {
	return scope(repository, key, FixtureAuthority)
}
func scope(repository, key string, class AuthorityClass) (Scope, error) {
	s := Scope{repository, key, class}
	if !s.valid() {
		return Scope{}, errors.New("invalid publisher scope")
	}
	return s, nil
}
func (s Scope) valid() bool {
	return repositoryPattern.MatchString(s.repository) && repositoryPattern.MatchString(s.key) && len(s.repository) <= 128 && len(s.key) <= 128 && ((s.class == FixtureAuthority && strings.HasPrefix(s.repository, "fixture.")) || (s.class == ReleaseAuthority && !strings.HasPrefix(s.repository, "fixture.")))
}
func (s Scope) Repository() string    { return s.repository }
func (s Scope) PackKey() string       { return s.key }
func (s Scope) Class() AuthorityClass { return s.class }

type Clock interface{ Now() time.Time }
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }
