package service

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/migotom/mt-bulk/internal/schema"
)

// GenerateCA generates CA key and certificate needed for sign device/client certificates.
func GenerateCA(config *schema.GeneralConfig) error {
	log.Println("[CONFIG] Generating CA")

	if _, err := os.Stat(config.Certs.OpenSSL); err != nil {
		return fmt.Errorf("Missing OpenSSL binary location in config file, %s", err.Error())
	}

	if _, err := os.Stat(config.Certs.Directory); err != nil {
		return fmt.Errorf("os stat error: %v", err)
	}
	ca := filepath.Join(config.Certs.Directory, "ca.crt")
	key := filepath.Join(config.Certs.Directory, "ca.key")

	cmd := exec.Command(config.Certs.OpenSSL, "req", "-nodes", "-new", "-x509", "-days", "3650", "-subj", "/C=US/ST=US/L=FakeTown/O=IT/CN=ca.cell-crm.com", "-keyout", key, "-out", ca)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ssl req: %v, %v", err, out)
	}

	return nil
}

// GenerateCerts generates and sign using CA certificate generic key and certificate.
func GenerateCerts(config *schema.GeneralConfig, subject string) error {
	log.Println("[CONFIG] Generating " + subject + " certificate")

	if _, err := os.Stat(config.Certs.OpenSSL); err != nil {
		return fmt.Errorf("Missing OpenSSL binary location in config file, %s", err.Error())
	}

	if _, err := os.Stat(config.Certs.Directory); err != nil {
		return fmt.Errorf("os stats error %v", err)
	}

	keyCA := filepath.Join(config.Certs.Directory, "ca.key")
	crtCA := filepath.Join(config.Certs.Directory, "ca.crt")

	key := filepath.Join(config.Certs.Directory, subject+".key")
	csr := filepath.Join(config.Certs.Directory, subject+".csr")
	crt := filepath.Join(config.Certs.Directory, subject+".crt")

	cmd := exec.Command(config.Certs.OpenSSL, "genrsa", "-out", key, "2084")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ssl genrsa: %v, %v", err, string(out))
	}

	cmd = exec.Command(config.Certs.OpenSSL, "req", "-new", "-subj", "/C=US/ST=US/L=FakeTown/O=IT/CN=mtbulk"+subject+".cell-crm.com", "-key", key, "-out", csr)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ssl req: %v, %v", err, string(out))
	}

	cmd = exec.Command(config.Certs.OpenSSL, "x509", "-req", "-days", "3650", "-in", csr, "-CA", crtCA, "-CAkey", keyCA, "-set_serial", "01", "-out", crt)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ssl x509: %v, %v", err, string(out))
	}

	return nil
}
