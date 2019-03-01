package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/migotom/mt-bulk/internal/schema"
)

// GenerateCA generates CA key and certificate needed for sign device/client certificates.
func GenerateCA(config *schema.GeneralConfig) error {
	log.Println("[CONFIG] Generating CA")

	if _, err := os.Stat(config.Certs.Directory); err != nil {
		return fmt.Errorf("Can't locate certificates directory: %s", err)
	}

	privateCA, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("Can't generate CA private key: %s", err)
	}

	privateCAFile, err := os.OpenFile(filepath.Join(config.Certs.Directory, "ca.key"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Can't create CA private key: %s", err)
	}
	pem.Encode(privateCAFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateCA)})
	privateCAFile.Close()

	publicCA := &privateCA.PublicKey

	csrCA := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Organization: []string{"IT"},
			Country:      []string{"PL"},
			Province:     []string{"PL"},
			Locality:     []string{"City"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	certificateCA, err := x509.CreateCertificate(rand.Reader, csrCA, csrCA, publicCA, privateCA)
	if err != nil {
		return fmt.Errorf("Can't generate CA certificate: %s", err)
	}

	certificateFile, err := os.Create(filepath.Join(config.Certs.Directory, "ca.crt"))
	if err != nil {
		return fmt.Errorf("Can't create CA certificate: %s", err)
	}

	pem.Encode(certificateFile, &pem.Block{Type: "CERTIFICATE", Bytes: certificateCA})
	certificateFile.Close()

	return nil
}

// GenerateCerts generates and sign using CA certificate generic key and certificate.
func GenerateCerts(config *schema.GeneralConfig, subject string) error {
	log.Println("[CONFIG] Generating " + subject + " certificate")

	if _, err := os.Stat(config.Certs.Directory); err != nil {
		return fmt.Errorf("Can't locate certificates directory: %s", err)
	}

	tlsCA, err := tls.LoadX509KeyPair(filepath.Join(config.Certs.Directory, "ca.crt"), filepath.Join(config.Certs.Directory, "ca.key"))
	if _, err := os.Stat(config.Certs.Directory); err != nil {
		return fmt.Errorf("Can't load CA certificate: %s", err)
	}

	certificateCA, err := x509.ParseCertificate(tlsCA.Certificate[0])
	if err != nil {
		return fmt.Errorf("Invalid CA certificate: %s", err)
	}

	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("Can't generate %s private key: %s", subject, err)
	}

	privateFile, err := os.OpenFile(filepath.Join(config.Certs.Directory, subject+".key"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Can't create %s private key: %s", subject, err)
	}
	pem.Encode(privateFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(private)})
	privateFile.Close()

	public := &private.PublicKey

	csr := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"IT"},
			Country:      []string{"PL"},
			Province:     []string{"PL"},
			Locality:     []string{"FakeTown"},
			CommonName:   "mtbulk" + subject + ".cell-crm.com",
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certificate, err := x509.CreateCertificate(rand.Reader, csr, certificateCA, public, tlsCA.PrivateKey)
	if err != nil {
		return fmt.Errorf("Can't generate %s certificate: %s", subject, err)
	}

	certificateFile, err := os.Create(filepath.Join(config.Certs.Directory, subject+".crt"))
	if err != nil {
		return fmt.Errorf("Can't create %s certificate: %s", subject, err)
	}

	pem.Encode(certificateFile, &pem.Block{Type: "CERTIFICATE", Bytes: certificate})
	certificateFile.Close()

	return nil
}
