package incidents

import (
	"errors"
	"reflect"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/workbook/startup/bootstrapport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// Application is the Incidents application boundary. It owns transaction and
// policy coordination while the private repository owns persistence details.
type Application struct {
	pool                postgres.DB
	authStore           *authn.Store
	repository          *repository
	admission           *admission.Checker
	preferenceBootstrap bootstrapport.Writer
	now                 func() time.Time
}

type ApplicationDependencies struct {
	Postgres            postgres.DB
	PreferenceBootstrap bootstrapport.Writer
	Now                 func() time.Time
}

// These values are persisted with public idempotency payloads and therefore
// remain stable even though HTTP response selection belongs to httpapi.
const (
	persistedSuccessStatus = 200
	persistedCreatedStatus = 201
)

func NewApplication(dependencies ApplicationDependencies) (*Application, error) {
	if isNilApplicationDependency(dependencies.Postgres) {
		return nil, errors.New("incidents: Postgres dependency is required")
	}
	if isNilApplicationDependency(dependencies.PreferenceBootstrap) {
		return nil, errors.New("incidents: workbook preference bootstrap dependency is required")
	}
	if dependencies.Now == nil {
		return nil, errors.New("incidents: mutation clock dependency is required")
	}
	return &Application{
		pool:                dependencies.Postgres,
		authStore:           authn.NewStore(dependencies.Postgres),
		repository:          newRepository(dependencies.Postgres),
		admission:           admission.NewChecker(dependencies.Postgres),
		preferenceBootstrap: dependencies.PreferenceBootstrap,
		now:                 dependencies.Now,
	}, nil
}

func isNilApplicationDependency(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (a *Application) recordedMutationTime() (time.Time, error) {
	if a == nil || a.now == nil {
		return time.Time{}, errInvalidMutationTime
	}
	recorded := a.now().UTC()
	if recorded.IsZero() {
		return time.Time{}, errInvalidMutationTime
	}
	return recorded, nil
}
