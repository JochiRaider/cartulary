package jobs

import (
	"fmt"
	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
)

const HumanRouteIdentitySchema = jobcontract.HumanRouteIdentitySchema
const AttributedRouteIdentitySchema = jobcontract.AttributedRouteIdentitySchema

func decodeRouteIdentity(raw []byte, schema string) (jobcontract.RouteIdentity, error) {
	identity, err := jobcontract.DecodeIdentity(raw, schema)
	if err != nil {
		return identity, fmt.Errorf("%w: invalid extension idempotency identity", ErrInvalidJobDefinition)
	}
	return identity, nil
}

func validIdentitySchema(schema string) bool { return jobcontract.HasIdentitySchema(schema) }
