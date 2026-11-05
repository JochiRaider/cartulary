package networkflow

import (
	"bytes"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// edgeSemantics is the owner key shared by composition, selectors and SQL.
// Directed First/Second are from/to. Unknown First/Second are canonical A/B.
// A directed first port is absent because client ephemeral ports do not group.
type edgeSemantics struct {
	DirectionKind    string
	CounterSemantics string
	First, Second    flowEndpoint
	Protocol         int32
}

func edgeSemanticsForRow(row flowRow, table tableRecord) (edgeSemantics, bool) {
	k := edgeSemantics{DirectionKind: row.Direction.Kind, CounterSemantics: counterSemanticsForProfile(table.SourceProfileID), Protocol: row.IPProtocol}
	if k.CounterSemantics == "" || !row.Direction.valid() {
		return k, false
	}
	if from, to, known := row.directedEndpoints(); known {
		k.First, k.Second = from, to
		k.First.Port = nil
	} else {
		k.First, k.Second = row.endpoints()
	}
	return k, k.valid()
}

func (k edgeSemantics) valid() bool {
	if k.CounterSemantics != "conversation_total" && k.CounterSemantics != "source_reported_flow_total" {
		return false
	}
	for _, e := range []flowEndpoint{k.First, k.Second} {
		ip, err := parseIPLiteral(e.IP)
		if err != nil || ip != e.IP || e.Port != nil && (*e.Port < 0 || *e.Port > 65535) {
			return false
		}
	}
	if k.Protocol < 0 || k.Protocol > 255 {
		return false
	}
	switch k.DirectionKind {
	case directionUnknown:
		return compareFlowEndpoints(k.First, k.Second) <= 0
	case directionClientServer, directionSourceDestination:
		return k.First.Port == nil
	default:
		return false
	}
}

func (k edgeSemantics) identity(incident uuid.UUID, start, end *time.Time) string {
	var b bytes.Buffer
	writeLengthFramedPart(&b, "cartulary.network_flow_flow_edge_id.v2")
	writeLengthFramedPart(&b, incident.String())
	writeLengthFramedPart(&b, k.DirectionKind)
	writeLengthFramedPart(&b, k.CounterSemantics)
	for _, e := range []flowEndpoint{k.First, k.Second} {
		writeLengthFramedPart(&b, e.IP)
		if e.Port == nil {
			writeLengthFramedPart(&b, "n")
		} else {
			writeLengthFramedPart(&b, "p")
			writeLengthFramedPart(&b, strconv.FormatInt(int64(*e.Port), 10))
		}
	}
	writeLengthFramedPart(&b, strconv.FormatInt(int64(k.Protocol), 10))
	prefix := "nff_"
	if start != nil && end != nil {
		prefix = "nfbe_"
		writeLengthFramedPart(&b, formatTimestamp(*start))
		writeLengthFramedPart(&b, formatTimestamp(*end))
	}
	return prefix + sha256Hex(b.Bytes())
}

func (k edgeSemantics) projectionDirection() string {
	if k.DirectionKind == directionUnknown {
		return "undirected"
	}
	return "forward"
}

func (s graphSelector) semantics() edgeSemantics {
	return edgeSemantics{DirectionKind: s.DirectionKind, CounterSemantics: s.CounterSemantics, First: flowEndpoint{s.SourceEndpointValue, s.SourcePort}, Second: flowEndpoint{s.DestinationEndpointValue, s.DestinationPort}, Protocol: s.Protocol}
}
