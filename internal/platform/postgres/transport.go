package postgres

import (
	"crypto/tls"
	"net"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
)

type certificateDSN struct {
	endpoint *url.URL
	tls      cryptography.TLSClientOptions
}

// parseCertificateDSN is deliberately independent of driver parsing. The
// driver accepts filesystem and environment expansion that this application
// does not support. No unadmitted string or filename reaches that parser.
func parseCertificateDSN(dsn string) (certificateDSN, error) {
	reject := func() (certificateDSN, error) {
		return certificateDSN{}, configurationError(ReasonSelectedCredentialInvalid)
	}
	if dsn == "" || len(dsn) > int(filesystemRootDSNMaximumBytes) || !utf8.ValidString(dsn) || strings.ContainsAny(dsn, "\x00\r\n\t #") {
		return reject()
	}
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Opaque != "" || u.Fragment != "" || u.User == nil {
		return reject()
	}
	if _, password := u.User.Password(); password || !validConnectionName(u.User.Username()) {
		return reject()
	}
	host := u.Hostname()
	if host == "" || strings.ContainsAny(host, ",/\\@ \t\r\n\x00") || (strings.Contains(host, ":") && net.ParseIP(host) == nil) {
		return reject()
	}
	if port := u.Port(); port != "" {
		value, err := strconv.ParseUint(port, 10, 16)
		if err != nil || value == 0 || strconv.FormatUint(value, 10) != port {
			return reject()
		}
	} else if strings.HasSuffix(u.Host, ":") {
		return reject()
	}
	if !strings.HasPrefix(u.Path, "/") || !validConnectionName(strings.TrimPrefix(u.Path, "/")) {
		return reject()
	}
	query, err := url.ParseQuery(u.RawQuery)
	if err != nil || len(query) != 5 {
		return reject()
	}
	for name, values := range query {
		if len(values) != 1 || values[0] == "" {
			return reject()
		}
		switch name {
		case "sslmode":
			if values[0] != "verify-full" {
				return reject()
			}
		case "require_auth":
			if values[0] != "none" {
				return reject()
			}
		case "sslrootcert", "sslcert", "sslkey":
			value := values[0]
			if !strings.HasPrefix(value, "/") || value == "/" || path.Clean(value) != value || strings.ContainsAny(value, "\\\x00\r\n") {
				return reject()
			}
		default:
			return reject()
		}
	}
	return certificateDSN{endpoint: u, tls: cryptography.TLSClientOptions{
		ServerName: host, RootCertificatePath: query.Get("sslrootcert"),
		ClientCertificatePath: query.Get("sslcert"), ClientPrivateKeyPath: query.Get("sslkey"),
	}}, nil
}

func validConnectionName(name string) bool {
	return name != "" && utf8.ValidString(name) && !strings.ContainsAny(name, "/\\\x00\r\n\t")
}

// inheritedPostgresSettings inspects names only. A supplied environment is the
// complete resolver view; construction separately checks the real process
// environment because that is the view consulted internally by pgx.
func inheritedPostgresSettings(env map[string]string) bool {
	if env != nil {
		for name := range env {
			if strings.HasPrefix(name, "PG") {
				return true
			}
		}
		return false
	}
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(name, "PG") {
			return true
		}
	}
	return false
}

func connectionInputs(dsn string) (string, *tls.Config, error) {
	admitted, err := parseCertificateDSN(dsn)
	if err != nil || inheritedPostgresSettings(nil) {
		return "", nil, configurationError(ReasonSelectedCredentialInvalid)
	}
	tlsConfig, err := cryptography.TLSClient(admitted.tls)
	if err != nil {
		return "", nil, configurationError(ReasonSelectedCredentialInvalid)
	}
	// Empty file values override pgx's home-directory defaults. The driver
	// cannot read a passfile at an empty pathname. Certificate bytes were read
	// exactly once above and are installed before any connection can open.
	// The temporary parsed configuration also requires verified TLS; no
	// plaintext configuration or placeholder password is ever constructed.
	controlled := *admitted.endpoint
	controlled.RawQuery = url.Values{
		"sslmode": {"verify-full"}, "require_auth": {"none"},
		"sslrootcert": {""}, "sslcert": {""}, "sslkey": {""}, "passfile": {""},
	}.Encode()
	return controlled.String(), tlsConfig, nil
}
