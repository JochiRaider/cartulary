package crossownertransaction

// Limits is the immutable protocol projection supplied by application
// composition from the admitted Extensions registry. It is not configuration.
type Limits struct {
	MinimumParticipants, Participants, SerializationKeysPerParticipant, AggregateSerializationKeys, ValidationFindings int
	ParticipantInputBytes, AggregateInputBytes, PrepareResultBytes, AggregatePrepareResultBytes, ResultBytes           int64
}

func (l Limits) valid() bool {
	return l.MinimumParticipants > 0 && l.Participants >= l.MinimumParticipants && l.SerializationKeysPerParticipant > 0 &&
		l.AggregateSerializationKeys >= l.SerializationKeysPerParticipant && l.ValidationFindings > 0 &&
		l.ParticipantInputBytes > 0 && l.AggregateInputBytes >= l.ParticipantInputBytes &&
		l.PrepareResultBytes > 0 && l.AggregatePrepareResultBytes >= l.PrepareResultBytes && l.ResultBytes > 0
}

// boundedTotal validates before addition, so even an overflowing sum cannot
// turn an oversized result into an admitted one.
func boundedTotal(total, size, individual, aggregate int64) (int64, bool) {
	if total < 0 || size < 1 || size > individual || total > aggregate || size > aggregate-total {
		return total, false
	}
	return total + size, true
}
