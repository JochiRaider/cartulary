package telemetry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
	"go.opentelemetry.io/otel/attribute"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	collectlogs "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectmetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collecttrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/proto"
)

type telemetrySink struct {
	t     *testing.T
	calls atomic.Int32
}

func (s *telemetrySink) accept(signal, userAgent string) {
	s.t.Helper()
	version := map[string]string{"traces": "v1.45.0", "metrics": "v1.45.0", "logs": "v0.21.0"}[signal]
	expected := "Cartulary/1.2.3 OTel-OTLP-Exporter-go/" + version
	if userAgent != expected && !strings.HasPrefix(userAgent, expected+" grpc-go/") {
		s.t.Errorf("wrong %s exporter identity: %q", signal, userAgent)
	}
	s.calls.Add(1)
}
func (s *telemetrySink) grpcAccept(ctx context.Context, signal string) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		s.t.Fatal("collector has no authenticated peer")
	}
	info, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok || info.State.Version != tls.VersionTLS13 {
		s.t.Error("collector accepted unqualified transport")
	}
	md, _ := metadata.FromIncomingContext(ctx)
	s.accept(signal, strings.Join(md.Get("user-agent"), " "))
}

type traceCollector struct {
	collecttrace.UnimplementedTraceServiceServer
	sink *telemetrySink
}

func (c traceCollector) Export(ctx context.Context, request *collecttrace.ExportTraceServiceRequest) (*collecttrace.ExportTraceServiceResponse, error) {
	c.sink.grpcAccept(ctx, "traces")
	if len(request.ResourceSpans) == 0 {
		c.sink.t.Error("empty trace export")
	}
	return &collecttrace.ExportTraceServiceResponse{}, nil
}

type metricCollector struct {
	collectmetrics.UnimplementedMetricsServiceServer
	sink *telemetrySink
}

func (c metricCollector) Export(ctx context.Context, request *collectmetrics.ExportMetricsServiceRequest) (*collectmetrics.ExportMetricsServiceResponse, error) {
	c.sink.grpcAccept(ctx, "metrics")
	if len(request.ResourceMetrics) == 0 {
		c.sink.t.Error("empty metric export")
	}
	return &collectmetrics.ExportMetricsServiceResponse{}, nil
}

type logCollector struct {
	collectlogs.UnimplementedLogsServiceServer
	sink *telemetrySink
}

func (c logCollector) Export(ctx context.Context, request *collectlogs.ExportLogsServiceRequest) (*collectlogs.ExportLogsServiceResponse, error) {
	c.sink.grpcAccept(ctx, "logs")
	if len(request.ResourceLogs) == 0 {
		c.sink.t.Error("empty log export")
	}
	return &collectlogs.ExportLogsServiceResponse{}, nil
}

func collectorTLS(t *testing.T, mode string) *tls.Config {
	t.Helper()
	ca, err := tlstest.NewAuthority()
	if err != nil {
		t.Fatal(err)
	}
	name := "127.0.0.1"
	if mode == "wrong-name" {
		name = "other.example.test"
	}
	identity, err := ca.Issue("collector", []string{name}, x509.ExtKeyUsageServerAuth)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	write := func(name string, data []byte) string {
		p, err := tlstest.WriteFile(directory, name, data)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	cert, key := write("server.pem", identity.CertificatePEM), write("server.key", identity.PrivateKeyPEM)
	config, err := cryptography.TLSServer(name, cert, key)
	if err != nil {
		t.Fatal(err)
	}
	if mode == "expired" {
		block, _ := pem.Decode(identity.CertificatePEM)
		leaf, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			t.Fatal(err)
		}
		leaf.NotBefore = time.Now().Add(-2 * time.Hour)
		leaf.NotAfter = time.Now().Add(-time.Hour)
		raw, err := ca.Sign(leaf, leaf.PublicKey)
		if err != nil {
			t.Fatal(err)
		}
		pair, err := tls.X509KeyPair(append(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: raw}), ca.CertificatePEM...), identity.PrivateKeyPEM)
		if err != nil {
			t.Fatal(err)
		}
		config = &tls.Config{Certificates: []tls.Certificate{pair}, MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS13, CurvePreferences: []tls.CurveID{tls.CurveP256}}
	}
	if mode == "wrong-root" {
		ca, err = tlstest.NewAuthority()
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("SSL_CERT_FILE", write("root.pem", ca.CertificatePEM))
	if mode == "tls12" {
		config = &tls.Config{Certificates: config.Certificates, MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12}
	}
	return config
}

func startTelemetryCollector(t *testing.T, kind, mode string) (string, *telemetrySink) {
	t.Helper()
	config := collectorTLS(t, mode)
	sink := &telemetrySink{t: t}
	if kind == "otlp_http" {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.TLS == nil || r.TLS.Version != tls.VersionTLS13 {
				t.Error("HTTP collector accepted unqualified transport")
			}
			if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/x-protobuf" {
				t.Error("unexpected OTLP HTTP request")
			}
			var message proto.Message
			signal := strings.TrimPrefix(r.URL.Path, "/v1/")
			switch signal {
			case "traces":
				message = &collecttrace.ExportTraceServiceRequest{}
			case "metrics":
				message = &collectmetrics.ExportMetricsServiceRequest{}
			case "logs":
				message = &collectlogs.ExportLogsServiceRequest{}
			default:
				t.Error("unexpected signal")
			}
			payload, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			if err != nil || len(payload) == 0 || proto.Unmarshal(payload, message) != nil {
				t.Error("invalid OTLP protobuf")
			}
			sink.accept(signal, r.Header.Get("User-Agent"))
			w.Header().Set("Content-Type", "application/x-protobuf")
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewUnstartedServer(handler)
		server.TLS = config
		server.StartTLS()
		t.Cleanup(server.Close)
		return server.URL, sink
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer(grpc.Creds(credentials.NewTLS(config)))
	collecttrace.RegisterTraceServiceServer(server, traceCollector{sink: sink})
	collectmetrics.RegisterMetricsServiceServer(server, metricCollector{sink: sink})
	collectlogs.RegisterLogsServiceServer(server, logCollector{sink: sink})
	stopped := make(chan struct{})
	go func() { defer close(stopped); _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); <-stopped; _ = listener.Close() })
	return "https://" + listener.Addr().String(), sink
}

func TestQualifiedTelemetryTransports(t *testing.T) {
	selected := os.Getenv("CARTULARY_TELEMETRY_TRANSPORT_CASE")
	if selected == "" {
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cases := []string{"redirect"}
		for _, kind := range []string{"otlp_http", "otlp_grpc"} {
			for _, mode := range []string{"valid", "wrong-root", "wrong-name", "expired", "tls12"} {
				cases = append(cases, kind+"/"+mode)
			}
		}
		for _, selected := range cases {
			t.Run(selected, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				command := exec.CommandContext(ctx, executable, "-test.run=^TestQualifiedTelemetryTransports$", "-test.v")
				for _, entry := range os.Environ() {
					if !strings.HasPrefix(entry, "SSL_CERT_FILE=") && !strings.HasPrefix(entry, "SSL_CERT_DIR=") && !strings.HasPrefix(entry, "CARTULARY_TELEMETRY_TRANSPORT_CASE=") {
						command.Env = append(command.Env, entry)
					}
				}
				command.Env = append(command.Env, "CARTULARY_TELEMETRY_TRANSPORT_CASE="+selected)
				if output, err := command.CombinedOutput(); err != nil {
					t.Fatalf("isolated collector process: %v\n%s", err, output)
				}
			})
		}
		return
	}
	// System trust is intentionally captured once per application process.
	// Every collector case starts in a fresh process, as a real trust restart does.
	if selected == "redirect" {

		t.Run("plaintext and redirects rejected", func(t *testing.T) {
			if _, err := newExporterTransport("http://127.0.0.1:4318", time.Second); err == nil {
				t.Fatal("plaintext admitted")
			}
			config := collectorTLS(t, "valid")
			var reached atomic.Int32
			destination := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached.Add(1); w.WriteHeader(http.StatusOK) }))
			destination.TLS = config.Clone()
			destination.StartTLS()
			defer destination.Close()
			redirect := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, destination.URL, http.StatusTemporaryRedirect)
			}))
			redirect.TLS = config
			redirect.StartTLS()
			defer redirect.Close()
			transport, err := newExporterTransport(redirect.URL, time.Second)
			if err != nil {
				t.Fatal(err)
			}
			defer transport.close()
			request, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, redirect.URL, strings.NewReader("private-export"))
			response, err := transport.client.Do(request)
			if response != nil {
				_ = response.Body.Close()
			}
			if err == nil || reached.Load() != 0 {
				t.Fatal("redirect followed")
			}
		})

		return
	}
	for _, kind := range []string{"otlp_http", "otlp_grpc"} {
		t.Run(kind, func(t *testing.T) {
			for _, mode := range []string{"valid", "wrong-root", "wrong-name", "expired", "tls12"} {
				if selected != kind+"/"+mode {
					continue
				}
				t.Run(mode, func(t *testing.T) {
					endpoint, sink := startTelemetryCollector(t, kind, mode)
					cfg := validTelemetryBootstrapConfig(t).Telemetry
					cfg.Exporter.Kind = kind
					cfg.Exporter.Endpoint = endpoint
					cfg.Exporter.Retry.Enabled = false
					cfg.Exporter.Retry.MaxElapsedMS = 0
					cfg.Resource.ServiceVersion = "1.2.3"
					cfg.Processor.ExportTimeoutMS = 750
					transport, err := newExporterTransport(endpoint, 750*time.Millisecond)
					if err != nil {
						t.Fatal(err)
					}
					defer transport.close()
					plan, err := BuildExporterRequestHeaders(nil, "Cartulary/1.2.3 OTel-OTLP-Exporter-go/v1.0.0")
					if err != nil {
						t.Fatal(err)
					}
					for _, signal := range []string{"traces", "metrics", "logs"} {
						t.Run(signal, func(t *testing.T) {
							ctx, cancel := context.WithTimeout(t.Context(), time.Second)
							defer cancel()
							err := withContainedOTelEnvironment(func() error {
								switch signal {
								case "traces":
									exporter, err := newTraceExporter(ctx, cfg, plan, transport)
									if err != nil {
										return err
									}
									defer exporter.Shutdown(context.Background())
									span := tracetest.SpanStub{Name: "qualification", SpanContext: trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{1}, TraceFlags: trace.FlagsSampled}), StartTime: time.Now(), EndTime: time.Now()}.Snapshot()
									return exporter.ExportSpans(ctx, []sdktrace.ReadOnlySpan{span})
								case "metrics":
									exporter, err := newMetricExporter(ctx, cfg, plan, transport)
									if err != nil {
										return err
									}
									defer exporter.Shutdown(context.Background())
									return exporter.Export(ctx, &metricdata.ResourceMetrics{ScopeMetrics: []metricdata.ScopeMetrics{{Metrics: []metricdata.Metrics{{Name: "qualification", Data: metricdata.Gauge[int64]{DataPoints: []metricdata.DataPoint[int64]{{Value: 1}}}}}}}})
								default:
									exporter, err := newLogExporter(ctx, cfg, plan, transport)
									if err != nil {
										return err
									}
									defer exporter.Shutdown(context.Background())
									var record sdklog.Record
									record.SetBody(attribute.StringValue("qualification"))
									record.SetTimestamp(time.Now())
									return exporter.Export(ctx, []sdklog.Record{record})
								}
							})
							if mode == "valid" && err != nil {
								t.Fatal(err)
							}
							if mode != "valid" && err == nil {
								t.Fatal("unqualified collector accepted")
							}
						})
					}
					want := int32(0)
					if mode == "valid" {
						want = 3
					}
					if sink.calls.Load() != want {
						t.Fatalf("collector requests=%d want=%d", sink.calls.Load(), want)
					}
				})
			}
		})
	}
}
