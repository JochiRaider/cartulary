package telemetry

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
)

type exporterTransport struct {
	tls    *tls.Config
	client *http.Client
	close  func()
}

func newExporterTransport(endpoint string, timeout time.Duration) (exporterTransport, error) {
	parsed, err := parseEndpoint(endpoint)
	if err != nil {
		return exporterTransport{}, err
	}
	config, err := cryptography.TLSClient(cryptography.TLSClientOptions{ServerName: parsed.Hostname()})
	if err != nil {
		return exporterTransport{}, err
	}
	transport := &http.Transport{TLSClientConfig: config, DialContext: (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext, TLSHandshakeTimeout: 5 * time.Second, IdleConnTimeout: 30 * time.Second, MaxIdleConns: 6, MaxIdleConnsPerHost: 6, ForceAttemptHTTP2: true}
	client := &http.Client{Transport: exporterEndpointTransport{transport: transport, authority: parsed.Host}, Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("telemetry redirect rejected") }}
	return exporterTransport{tls: config, client: client, close: transport.CloseIdleConnections}, nil
}

type exporterEndpointTransport struct {
	transport *http.Transport
	authority string
}

func (t exporterEndpointTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Scheme != "https" || request.URL.Host != t.authority || request.URL.User != nil {
		return nil, errors.New("telemetry endpoint rejected")
	}
	return t.transport.RoundTrip(request)
}
