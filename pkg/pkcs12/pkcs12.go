package pkcs12

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	gopkcs12 "software.sslmate.com/src/go-pkcs12"
)

// CreatePkcs12 creates a PKCS12 keystore from a certificate and the private key byte slices
func CreatePkcs12(cert []byte, key []byte, password string) ([]byte, error) {

	domainCert, _ := parseCrt(cert)
	privateKey, _ := parseKey(key)

	pfxData, err := gopkcs12.Encode(rand.Reader, privateKey, domainCert, nil, password)

	return pfxData, err
}

func parseCrt(cert []byte) (*x509.Certificate, error) {
	p := &pem.Block{}
	p, _ = pem.Decode(cert)
	return x509.ParseCertificate(p.Bytes)
}

func parseKey(key []byte) (*rsa.PrivateKey, error) {
	p, _ := pem.Decode(key)
	return x509.ParsePKCS1PrivateKey(p.Bytes)
}
