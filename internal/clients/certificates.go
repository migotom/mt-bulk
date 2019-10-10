package clients

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// GenerateCA generates CA key and certificate needed for sign device/clients certificates.
func GenerateCA(store string) error {
	log.Println("[CONFIG] Generating CA")

	if store == "" {
		return errors.New("certificates directory not provided")
	}

	if _, err := os.Stat(store); err != nil {
		return fmt.Errorf("can't locate certificates directory (%s), %s", store, err)
	}

	privateCA, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("can't generate CA private key: %s", err)
	}

	privateCAFile, err := os.OpenFile(filepath.Join(store, "ca.key"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("can't create CA private key: %s", err)
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
		return fmt.Errorf("can't generate CA certificate: %s", err)
	}

	certificateFile, err := os.Create(filepath.Join(store, "ca.crt"))
	if err != nil {
		return fmt.Errorf("can't create CA certificate: %s", err)
	}

	pem.Encode(certificateFile, &pem.Block{Type: "CERTIFICATE", Bytes: certificateCA})
	certificateFile.Close()

	return nil
}

// GenerateCerts generates and sign using CA certificate generic key and certificate.
func GenerateCerts(store, subject string) error {
	log.Println("[CONFIG] Generating " + subject + " certificate")

	if store == "" {
		return errors.New("certificates directory not provided")
	}

	if _, err := os.Stat(store); err != nil {
		return fmt.Errorf("can't locate certificates directory (%s), %s", store, err)
	}

	tlsCA, err := tls.LoadX509KeyPair(filepath.Join(store, "ca.crt"), filepath.Join(store, "ca.key"))
	if err != nil {
		return fmt.Errorf("can't load CA key pair: %s", err)
	}

	certificateCA, err := x509.ParseCertificate(tlsCA.Certificate[0])
	if err != nil {
		return fmt.Errorf("invalid CA certificate: %s", err)
	}

	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("can't generate %s private key: %s", subject, err)
	}

	privateFile, err := os.OpenFile(filepath.Join(store, subject+".key"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("can't create %s private key: %s", subject, err)
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
		return fmt.Errorf("can't generate %s certificate: %s", subject, err)
	}

	certificateFile, err := os.Create(filepath.Join(store, subject+".crt"))
	if err != nil {
		return fmt.Errorf("can't create %s certificate: %s", subject, err)
	}

	pem.Encode(certificateFile, &pem.Block{Type: "CERTIFICATE", Bytes: certificate})
	certificateFile.Close()

	return nil
}

// GenerateKeys generates private and public keys.
func GenerateKeys(store string) error {
	log.Println("[CONFIG] Generating SSH keys")

	if store == "" {
		return errors.New("keys directory not provided")
	}

	if _, err := os.Stat(store); err != nil {
		return fmt.Errorf("can't locate keys directory (%s), %s", store, err)
	}

	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("can't generate private key: %s", err)
	}

	privateFile, err := os.OpenFile(filepath.Join(store, "id_rsa.key"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("can't create private key: %s", err)
	}
	pem.Encode(privateFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(private)})
	privateFile.Close()

	public, err := ssh.NewPublicKey(&private.PublicKey)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(store, "id_rsa.pub"), ssh.MarshalAuthorizedKey(public), 0600)
	if err != nil {
		return fmt.Errorf("can't create public key: %s", err)
	}

	return nil
}
