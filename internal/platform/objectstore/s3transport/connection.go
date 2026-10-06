// Package s3transport constructs the single supported S3 connection policy.
// Object payloads, bucket lifecycle and error interpretation belong to callers.
package s3transport

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Options struct {
	Endpoint            string
	Region              string
	AccessKey           string
	SecretKey           string
	RootCertificatePath string
}

type Connection struct {
	Client    *s3.Client
	transport *http.Transport
}

var ErrConfiguration = errors.New("S3 connection configuration rejected")

func New(options Options) (*Connection, error) {
	endpoint, err := url.Parse("https://" + options.Endpoint)
	if err != nil || endpoint.Hostname() == "" || endpoint.Host != options.Endpoint || endpoint.User != nil || endpoint.Path != "" || endpoint.RawQuery != "" || endpoint.Fragment != "" || strings.TrimSpace(options.Endpoint) != options.Endpoint || options.AccessKey == "" || len(options.SecretKey) < 16 {
		return nil, ErrConfiguration
	}
	region := options.Region
	if region == "" {
		region = "us-east-1"
	}
	if len(region) > 63 || strings.Trim(region, "abcdefghijklmnopqrstuvwxyz0123456789-") != "" {
		return nil, ErrConfiguration
	}
	tlsConfig, err := cryptography.TLSClient(cryptography.TLSClientOptions{ServerName: endpoint.Hostname(), RootCertificatePath: options.RootCertificatePath})
	if err != nil {
		return nil, ErrConfiguration
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig, DialContext: (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second, ResponseHeaderTimeout: 30 * time.Second, IdleConnTimeout: 90 * time.Second,
		MaxIdleConns: 16, MaxIdleConnsPerHost: 8, ForceAttemptHTTP2: true,
	}
	client := &http.Client{Transport: endpointTransport{authority: endpoint.Host, transport: transport}, CheckRedirect: func(*http.Request, []*http.Request) error { return ErrConfiguration }}
	sdk := s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint.String()), Region: region, UsePathStyle: true,
		Credentials: credentials.NewStaticCredentialsProvider(options.AccessKey, options.SecretKey, ""),
		HTTPClient:  client, RetryMaxAttempts: 1,
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		// Owner content validation is independent. Do not negotiate a server-selected
		// checksum algorithm on reads; upload operations explicitly select SHA-256.
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})
	return &Connection{Client: sdk, transport: transport}, nil
}

func (c *Connection) Close() error {
	if c != nil && c.transport != nil {
		c.transport.CloseIdleConnections()
	}
	return nil
}

type endpointTransport struct {
	authority string
	transport *http.Transport
}

func (t endpointTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Scheme != "https" || request.URL.Host != t.authority || request.URL.User != nil {
		return nil, ErrConfiguration
	}
	return t.transport.RoundTrip(request)
}

func BucketExists(ctx context.Context, client *s3.Client, bucket string) (bool, error) {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if ErrorCode(err) == "NotFound" || ErrorCode(err) == "NoSuchBucket" {
		return false, nil
	}
	return err == nil, err
}
