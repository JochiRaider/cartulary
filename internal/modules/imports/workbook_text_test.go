package imports

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestWorkbookCSVStreamedProjectionRetention(t *testing.T) {
	// This isolates the parser/serialization path used by approval. Forced-GC
	// samples measure live retained rows, not total allocation or server capacity.
	var manifest struct {
		Projection struct {
			RowCounts []int  `json:"row_counts"`
			CellBytes int    `json:"cell_bytes"`
			Interval  int    `json:"sampling_interval_rows"`
			MaxHeap   uint64 `json:"max_live_heap_growth_bytes"`
		} `json:"streamed_projection"`
	}
	raw, err := os.ReadFile("../../../fixtures/import-preservation/v1/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	config := manifest.Projection
	if len(config.RowCounts) != 2 || config.Interval < 1 || config.CellBytes < 1 || config.MaxHeap == 0 {
		t.Fatal("invalid streaming fixture contract")
	}
	for _, rows := range config.RowCounts {
		t.Run(fmt.Sprint(rows), func(t *testing.T) {
			text := strings.Repeat("x", config.CellBytes)
			source := "value\n" + strings.Repeat(text+"\n", rows)
			runtime.GC()
			var baseline runtime.MemStats
			runtime.ReadMemStats(&baseline)
			peak, count := uint64(0), 0
			digest := sha256.New()
			preview, err := scanGenericCSV(t.Context(), strings.NewReader(source), func(row map[string]any) error {
				count++
				if row["source_row_ref"] != count+1 {
					t.Fatal("row order changed")
				}
				value, err := json.Marshal(row)
				if err != nil {
					return err
				}
				_, _ = digest.Write(value)
				if count%config.Interval == 0 || count == rows {
					runtime.GC()
					var current runtime.MemStats
					runtime.ReadMemStats(&current)
					if current.HeapAlloc > baseline.HeapAlloc {
						peak = max(peak, current.HeapAlloc-baseline.HeapAlloc)
					}
				}
				return nil
			})
			runtime.KeepAlive(source)
			if err != nil || count != rows || preview.InferredRowCount != rows || len(preview.PreviewRows) != 50 {
				t.Fatalf("streamed projection incomplete: count=%d error=%v", count, err)
			}
			t.Logf("streamed rows=%d sampled live heap growth=%d bytes", rows, peak)
			if peak > config.MaxHeap {
				t.Fatalf("stream retained source-sized row state: %d > %d", peak, config.MaxHeap)
			}
		})
	}
}

func TestWorkbookCSVInterpretationVectors(t *testing.T) {
	root := "../../../fixtures/import-preservation/v1"
	raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Cases []struct {
			ID, File, SHA256, Stage string
			Accepted                bool
			Records                 [][]string
		}
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, vector := range manifest.Cases {
		t.Run(vector.ID, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(root, vector.File))
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprintf("%x", sha256.Sum256(raw)) != vector.SHA256 {
				t.Fatal("fixture digest mismatch")
			}
			// Fragment every byte, including multi-byte scalar boundaries.
			for _, fragment := range []bool{false, true} {
				var records [][]string
				var reader interface{ Read([]byte) (int, error) } = bytes.NewReader(raw)
				if fragment {
					reader = &oneByteWorkbookReader{Reader: bytes.NewReader(raw)}
				}
				preview, err := scanGenericCSV(t.Context(), reader, func(row map[string]any) error {
					fields := []string{}
					for _, cell := range row["cells"].([]map[string]any) {
						fields = append(fields, cell["display_text"].(string))
					}
					records = append(records, fields)
					return nil
				})
				if (err == nil) != vector.Accepted {
					t.Fatalf("fragment=%v accepted=%v err=%v", fragment, vector.Accepted, err)
				}
				if !vector.Accepted {
					continue
				}
				header := []string{}
				for _, col := range preview.Columns {
					value, _ := col["source_header_text"].(string)
					header = append(header, value)
				}
				actual := append([][]string{header}, records...)
				if !reflect.DeepEqual(actual, vector.Records) {
					t.Fatalf("records=%q expected=%q", actual, vector.Records)
				}
			}
		})
	}
}

type oneByteWorkbookReader struct{ *bytes.Reader }

func (r *oneByteWorkbookReader) Read(p []byte) (int, error) {
	if len(p) > 1 {
		p = p[:1]
	}
	return r.Reader.Read(p)
}
