package pgtest

import (
	"crypto/x509"
	"errors"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	testcontainers "github.com/testcontainers/testcontainers-go"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
)

const postgresTLSContainerRoot = "/etc/cartulary-postgres-tls"

func preparePostgresTLS(req testcontainers.ContainerRequest) (_ testcontainers.ContainerRequest, _ string, retErr error) {
	parent, _, err := suiteservices.ResolveSuiteRuntimeDir(nil)
	if err != nil {
		return req, "", err
	}
	directory, err := os.MkdirTemp(parent, "postgres-tls-")
	if err != nil {
		return req, "", err
	}
	ok := false
	defer func() {
		if !ok {
			retErr = errors.Join(retErr, os.RemoveAll(directory))
		}
	}()
	ca, err := tlstest.NewAuthority()
	if err != nil {
		return req, "", err
	}
	root, err := tlstest.WriteFile(directory, "root.pem", ca.CertificatePEM)
	if err != nil {
		return req, "", err
	}
	req.Files = append(req.Files, testcontainers.ContainerFile{HostFilePath: root, ContainerFilePath: postgresTLSContainerRoot + "/root.pem", FileMode: 0o644})
	server, err := ca.Issue("postgres", []string{"localhost", "127.0.0.1", "::1", "postgres", "host.docker.internal", "gateway.docker.internal"}, x509.ExtKeyUsageServerAuth)
	if err != nil {
		return req, "", err
	}
	for _, item := range []struct {
		name string
		data []byte
		mode int64
	}{
		{"server.pem", server.CertificatePEM, 0o644}, {"server.key", server.PrivateKeyPEM, 0o600},
		{"pg_hba.conf", []byte("local all all trust\nhostssl all all 0.0.0.0/0 cert\nhostssl all all ::/0 cert\nhostnossl all all 0.0.0.0/0 reject\nhostnossl all all ::/0 reject\n"), 0o644},
	} {
		path, err := tlstest.WriteFile(directory, item.name, item.data)
		if err != nil {
			return req, "", err
		}
		req.Files = append(req.Files, testcontainers.ContainerFile{HostFilePath: path, ContainerFilePath: postgresTLSContainerRoot + "/" + item.name, FileMode: item.mode})
	}
	for _, login := range []string{"cartulary", "cartulary_runtime_login", "cartulary_migration_login", "cartulary_recovery_login"} {
		// A second independently keyed identity allows actual renewal tests
		// without exporting or retaining the fixture CA's private key.
		for _, suffix := range []string{"", ".renewed"} {
			identity, err := ca.Issue(login, nil, x509.ExtKeyUsageClientAuth)
			if err != nil {
				return req, "", err
			}
			if _, err = tlstest.WriteFile(directory, login+suffix+".pem", identity.CertificatePEM); err != nil {
				return req, "", err
			}
			if _, err = tlstest.WriteFile(directory, login+suffix+".key", identity.PrivateKeyPEM); err != nil {
				return req, "", err
			}
		}
	}
	initPath, err := tlstest.WriteFile(directory, "disable-admin-password.sql", []byte("ALTER ROLE cartulary PASSWORD NULL;\n"))
	if err != nil {
		return req, "", err
	}
	req.Files = append(req.Files, testcontainers.ContainerFile{HostFilePath: initPath, ContainerFilePath: "/docker-entrypoint-initdb.d/10-disable-password.sql", FileMode: 0o644})
	req.Entrypoint = []string{"/bin/sh", "-ec", "chown postgres:postgres /etc/cartulary-postgres-tls/server.key; exec /usr/local/bin/docker-entrypoint.sh \"$@\"", "--"}
	req.Cmd = []string{"postgres", "-c", "ssl=on", "-c", "ssl_min_protocol_version=TLSv1.3", "-c", "ssl_max_protocol_version=TLSv1.3", "-c", "ssl_ecdh_curve=prime256v1", "-c", "ssl_ca_file=" + postgresTLSContainerRoot + "/root.pem", "-c", "ssl_cert_file=" + postgresTLSContainerRoot + "/server.pem", "-c", "ssl_key_file=" + postgresTLSContainerRoot + "/server.key", "-c", "hba_file=" + postgresTLSContainerRoot + "/pg_hba.conf"}
	ok = true
	return req, directory, nil
}

func certificateDSN(host, port, database, login, directory string) string {
	u := url.URL{Scheme: "postgres", User: url.User(login), Host: net.JoinHostPort(host, port), Path: "/" + database}
	u.RawQuery = url.Values{
		"sslmode": {"verify-full"}, "require_auth": {"none"},
		"sslrootcert": {filepath.Join(directory, "root.pem")},
		"sslcert":     {filepath.Join(directory, login+".pem")},
		"sslkey":      {filepath.Join(directory, login+".key")},
	}.Encode()
	return u.String()
}

func purposeCertificateDSN(dsn, login string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.User == nil {
		return "", errors.New("test postgres binding must be a certificate URI")
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil || q.Get("sslmode") != "verify-full" || q.Get("require_auth") != "none" || q.Get("sslrootcert") == "" {
		return "", errors.New("test postgres certificate binding missing")
	}
	if _, password := u.User.Password(); password {
		return "", errors.New("test postgres password binding rejected")
	}
	directory := filepath.Dir(q.Get("sslrootcert"))
	u.User = url.User(login)
	q.Set("sslcert", filepath.Join(directory, login+".pem"))
	q.Set("sslkey", filepath.Join(directory, login+".key"))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (h *Harness) DSNTemplate() string {
	if h.dsnTemplate != "" {
		return h.dsnTemplate
	}
	// URL.String escapes braces in a path; the harness template requires its
	// literal placeholder, which is replaced with a generated database name.
	return strings.ReplaceAll(h.dsnFor("{database}"), "%7Bdatabase%7D", "{database}")
}
