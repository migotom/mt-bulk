package clients

import (
	"context"
	"crypto/tls"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/migotom/routeros"
)

// MikrotikAPIDefaultPort is default Mikrotik SSL API server port.
const MikrotikAPIDefaultPort = "8729"

// NewMikrotikAPIClient returns new Mikrotik API client.
func NewMikrotikAPIClient(config Config) Client {
	return &MikrotikAPI{
		Config: config,
	}
}

// MikrotikAPI defines Mikrotik secure API client.
type MikrotikAPI struct {
	mtClient *routeros.Client

	Config
}

// GetConfig returns client's configuration.
func (mikrotikAPI *MikrotikAPI) GetConfig() Config {
	return mikrotikAPI.Config
}

// Connect to routerOS by Mikrotik secure API.
func (mikrotikAPI *MikrotikAPI) Connect(ctx context.Context, IP, Port, User, Password string) (err error) {
	clientCrt := filepath.Join(mikrotikAPI.Config.KeyStore, "client.crt")
	clientKey := filepath.Join(mikrotikAPI.Config.KeyStore, "client.key")
	certificate, err := tls.LoadX509KeyPair(clientCrt, clientKey)
	if err != nil {
		return err
	}

	tlsCfg := tls.Config{
		Certificates:       []tls.Certificate{certificate},
		InsecureSkipVerify: true,
	}

	tlsCfg.CipherSuites = append(tlsCfg.CipherSuites, tls.TLS_RSA_WITH_AES_128_CBC_SHA)
	tlsCfg.CipherSuites = append(tlsCfg.CipherSuites, tls.TLS_RSA_WITH_AES_128_CBC_SHA256)
	tlsCfg.CipherSuites = append(tlsCfg.CipherSuites, tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA)

	tlsCfg.CipherSuites = append(tlsCfg.CipherSuites, tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA)
	tlsCfg.CipherSuites = append(tlsCfg.CipherSuites, tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA)

	mikrotikAPI.mtClient, err = routeros.DialTLS(fmt.Sprintf("%s:%s", IP, Port), User, Password, &tlsCfg, 30*time.Second)
	if err != nil && strings.HasPrefix(err.Error(), "from RouterOS device: invalid user name or password") {
		return ErrorWrongPassword{err}
	}
	if err != nil {
		return ErrorRetryable{fmt.Errorf("Mikrotik API handle sequence error %v", err)}
	}

	mikrotikAPI.mtClient.Async()
	return nil
}

// Close Mikrotik API client session.
func (mikrotikAPI *MikrotikAPI) Close() {
	mikrotikAPI.mtClient.Close()
	return
}

// RunCmd execues given command on remote device, optionally can compare execution result with provided expect regexp.
func (mikrotikAPI MikrotikAPI) RunCmd(body string, expect *regexp.Regexp) (result string, err error) {
	reply, err := mikrotikAPI.mtClient.RunArgs(strings.Split(body, " "))
	if err != nil {
		return "", err
	}

	if expect != nil {
		_, err = waitForExpected(strings.NewReader(reply.String()), expect)
	}

	return fmt.Sprintf("%s\n%s", body, reply.String()), err
}
