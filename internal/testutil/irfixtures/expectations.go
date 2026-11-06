package irfixtures

// This oracle deliberately imports no product parser, normalizer, matcher,
// graph implementation, or identity/digest implementation. It computes source
// facts only. Expected product outcomes are frozen review inputs, not recordings.
import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

type TextCase struct {
	SourceRecord   int      `json:"source_record"`
	SourceIdentity string   `json:"source_identity"`
	Cells          []string `json:"cells"`
}

type Expectations struct {
	SchemaID            string              `json:"schema_id"`
	Revision            int                 `json:"revision"`
	Records             map[string]int      `json:"records"`
	TimePairs           int                 `json:"time_pairs"`
	EmptyTimes          int                 `json:"empty_times"`
	UTCOffsetMinutes    int                 `json:"utc_offset_minutes"`
	MultilineRaw        int                 `json:"multiline_raw"`
	MultilineSynopsis   int                 `json:"multiline_synopsis"`
	TabbedRaw           int                 `json:"tabbed_raw"`
	TabbedSynopsis      int                 `json:"tabbed_synopsis"`
	AliasAssociations   int                 `json:"alias_associations"`
	DistinctAliases     int                 `json:"distinct_aliases"`
	AmbiguousAliases    map[string][]string `json:"ambiguous_aliases"`
	TimelineCases       []TextCase          `json:"timeline_cases"`
	FlowObservations    int                 `json:"flow_observations"`
	DistinctFlowRecords int                 `json:"distinct_flow_records"`
	FlowBytes           string              `json:"flow_bytes"`
	FlowPackets         string              `json:"flow_packets"`
	ServerFirst         int                 `json:"server_first"`
	MissingOrientation  int                 `json:"missing_orientation"`
}

func (c *Corpus) Records(name string, delimiter rune) ([][]string, error) {
	data, err := c.Read(name)
	if err != nil {
		return nil, err
	}
	// encoding/csv intentionally folds CRLF, including inside quoted fields.
	// Shield quoted CR bytes so this independent reader can assert exact source
	// text against the product lexer. NUL is not present in admitted originals.
	if bytes.IndexByte(data, 0) >= 0 {
		return nil, fmt.Errorf("NUL in fixture CSV")
	}
	protected := bytes.Clone(data)
	quoted := false
	for i, b := range protected {
		if b == '"' {
			quoted = !quoted
		}
		if b == '\r' && quoted {
			protected[i] = 0
		}
	}
	r := csv.NewReader(bytes.NewReader(protected))
	r.Comma = delimiter
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		for i, cell := range row {
			row[i] = strings.ReplaceAll(cell, "\x00", "\r")
		}
	}
	return rows, nil
}

// BuildExpectations is invoked by the cataloged Make test. It can retain a
// candidate in that run's outputs, but never replaces the authored expectations.
func BuildExpectations(c *Corpus) (Expectations, error) {
	e := Expectations{SchemaID: "cartulary.ir_fixture_expectations.v1", Revision: 1, Records: map[string]int{}, UTCOffsetMinutes: -240, AmbiguousAliases: map[string][]string{}, TimelineCases: []TextCase{}}
	all := map[string][][]string{}
	for _, source := range c.Manifest.Sources {
		delim := ','
		if source.Use == "clipboard" {
			delim = '\t'
		}
		rows, err := c.Records("ir-incident/v1/"+source.Path, delim)
		if err != nil {
			return e, err
		}
		if len(rows) != source.Records+1 || !slices.Equal(rows[0], source.Columns) {
			return e, fmt.Errorf("source shape changed: %s", source.Path)
		}
		e.Records[source.Path] = len(rows) - 1
		all[source.Path] = rows[1:]
	}
	if !reflect.DeepEqual(all["timeline.csv"], all["timeline.tsv"]) {
		return e, fmt.Errorf("CSV and TSV differ")
	}
	evidence := regexp.MustCompile(`evidence_ref="([^"]+)"`)
	seenIDs := map[string]bool{}
	selected := map[string]bool{}
	for i, row := range all["timeline.csv"] {
		match := evidence.FindStringSubmatch(row[7])
		if len(match) != 2 || seenIDs[match[1]] {
			return e, fmt.Errorf("missing or duplicate timeline source identity at record %d", i+1)
		}
		seenIDs[match[1]] = true
		kinds := []string{}
		if i == 0 {
			kinds = append(kinds, "first")
		}
		if strings.Contains(row[7], "\n") {
			e.MultilineRaw++
			kinds = append(kinds, "multiline_raw")
		}
		if strings.Contains(row[8], "\n") {
			e.MultilineSynopsis++
			kinds = append(kinds, "multiline_synopsis")
		}
		if strings.Contains(row[7], "\t") {
			e.TabbedRaw++
			kinds = append(kinds, "tabbed_raw")
		}
		if strings.Contains(row[8], "\t") {
			e.TabbedSynopsis++
			kinds = append(kinds, "tabbed_synopsis")
		}
		if strings.Contains(row[7], "\r\n") {
			kinds = append(kinds, "quoted_crlf")
		}
		if strings.Contains(row[7], "\u0301") {
			kinds = append(kinds, "decomposed_unicode")
		}
		if row[5] == "" && row[6] == "" {
			e.EmptyTimes++
			kinds = append(kinds, "empty_time")
		} else {
			utc, err := time.Parse(time.RFC3339, row[5])
			if err != nil {
				return e, err
			}
			local, err := time.Parse("2006-01-02T15:04:05", row[6])
			if err != nil {
				return e, err
			}
			if utc.Sub(local) != 4*time.Hour {
				return e, fmt.Errorf("inconsistent source time pair at record %d", i+1)
			}
			e.TimePairs++
		}
		keep := false
		for _, kind := range kinds {
			if !selected[kind] {
				keep = true
				selected[kind] = true
			}
		}
		if keep {
			e.TimelineCases = append(e.TimelineCases, TextCase{SourceRecord: i + 1, SourceIdentity: match[1], Cells: row})
		}
	}
	rosters := map[string]map[string]bool{"host": {}, "identity": {}}
	for _, row := range all["hosts.csv"] {
		rosters["host"][row[1]] = true
	}
	for _, row := range all["identities.csv"] {
		rosters["identity"][row[1]] = true
	}
	aliases := map[string][]string{}
	for _, row := range all["aliases.csv"] {
		if !rosters[row[0]][row[1]] {
			return e, fmt.Errorf("orphan alias %q", row[3])
		}
		if slices.Contains(aliases[row[3]], row[1]) {
			return e, fmt.Errorf("duplicate association %q", row[3])
		}
		aliases[row[3]] = append(aliases[row[3]], row[1])
		e.AliasAssociations++
	}
	e.DistinctAliases = len(aliases)
	for alias, ids := range aliases {
		if len(ids) > 1 {
			slices.Sort(ids)
			e.AmbiguousAliases[alias] = ids
		}
	}
	unique := map[string]bool{}
	var totalBytes, totalPackets uint64
	for _, name := range c.Manifest.Dependencies[0].Datasets {
		rows, err := c.Records("network-flow-orientation/v1/"+name+".csv", ',')
		if err != nil {
			return e, err
		}
		columns := map[string]int{}
		for i, column := range rows[0] {
			columns[column] = i
		}
		for _, required := range []string{"connection.transferBytes", "connection.transferPackets", "searchSubject.orientation", "peer.orientation"} {
			if _, ok := columns[required]; !ok {
				return e, fmt.Errorf("missing flow column %s", required)
			}
		}
		e.Records[name+".csv"] = len(rows) - 1
		for _, row := range rows[1:] {
			// Length framing keeps embedded separators unambiguous.
			var key strings.Builder
			for _, cell := range row {
				fmt.Fprintf(&key, "%d:%s", len(cell), cell)
			}
			unique[key.String()] = true
			e.FlowObservations++
			b, err := strconv.ParseUint(row[columns["connection.transferBytes"]], 10, 64)
			if err != nil {
				return e, err
			}
			totalBytes += b
			p, err := strconv.ParseUint(row[columns["connection.transferPackets"]], 10, 64)
			if err != nil {
				return e, err
			}
			totalPackets += p
			first, second := row[columns["searchSubject.orientation"]], row[columns["peer.orientation"]]
			switch {
			case first == "server" && second == "client":
				e.ServerFirst++
			case first == "client" && second == "server":
			case first == "" && second == "":
				e.MissingOrientation++
			default:
				return e, fmt.Errorf("invalid orientation in %s", name)
			}
		}
	}
	e.DistinctFlowRecords = len(unique)
	e.FlowBytes = strconv.FormatUint(totalBytes, 10)
	e.FlowPackets = strconv.FormatUint(totalPackets, 10)
	return e, nil
}
