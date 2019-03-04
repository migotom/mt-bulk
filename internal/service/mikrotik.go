package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/migotom/routeros"
)

const portAPI = "8729"

// MTAPI defines Mikrotik secure API service by extending generic device with Routeros connection.
type MTAPI struct {
	mtClient *routeros.Client
	GenericDevice
}

// TODO refactor!!
func (d *MTAPI) GetUser() string {
	if d.Host.User != "" {
		return d.Host.User
	}
	return d.AppConfig.Service["mikrotik_api"].DefaultUser
}

func (d *MTAPI) GetPasswords() []string {
	list := func() string {
		if d.Host.Pass != "" {
			return d.Host.Pass
		}
		return d.AppConfig.Service["mikrotik_api"].DefaultPass
	}()

	return strings.Split(list, ",")
}

func (d *MTAPI) GetPort() string {
	if d.Host.Port != "" {
		return d.Host.Port
	}
	return portAPI
}

func (d *MTAPI) HandleSequence(ctx context.Context, handler HandlerFunc) error {
	// load client certificate

	clientCrt := filepath.Join(d.AppConfig.Certs.Directory, "client.crt")
	clientKey := filepath.Join(d.AppConfig.Certs.Directory, "client.key")
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

	for idx, password := range d.GetPasswords() {
		select {
		case <-ctx.Done():
			return nil
		default:
			log.Printf("[IP:%s][API] Initializing connection :%s (using password #%d)", d.GenericDevice.Host.IP, d.GetPort(), idx)
			// establish SSH connection

			d.mtClient, err = routeros.DialTLS(d.GenericDevice.Host.IP+":"+d.GetPort(), d.GetUser(), password, &tlsCfg, 10*time.Second)
			if err != nil {
				continue
			}
		}
		// store valid password for this device
		d.GenericDevice.Host.Pass = password
		break
	}
	if err != nil {
		return fmt.Errorf("API handle sequence error %v", err)
	}
	if d.mtClient == nil {
		return fmt.Errorf("API no correct passwords found")
	}
	defer d.mtClient.Close()

	d.mtClient.Async()

	d.matches = make(map[string]string)

	// call handlr with sequence of operations
	return handler(d)
}

func (d MTAPI) RunCmd(body string) (string, error) {
	r, err := d.mtClient.RunArgs(strings.Split(body, " "))
	if err != nil {
		return "", err
	}

	return r.String(), nil
}

func (d MTAPI) GetDevice() *GenericDevice {
	return &d.GenericDevice
}
