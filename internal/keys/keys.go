package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"strings"

	"go.uber.org/zap"
)

// GenerateKeys returns a new key pair on the elliptic curve
// source: https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go
func GenerateKeys() (*string, *string, error) {
	zap.S().Info("Generating keys...")
	sk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		zap.S().Errorf("Error Generating elliptic curve key: %v", err)
		return nil, nil, err
	}
	pk := &(sk.PublicKey)

	// Encode to strings
	// TODO is there a better to do this than simply replacing chars we don't want to use?
	replacer := strings.NewReplacer(" ", "", "-", "", "BEGIN", "", "END", "", "\n", "")

	skx509Encoded, _ := x509.MarshalECPrivateKey(sk)
	skEncoded := string(pem.EncodeToMemory(&pem.Block{Bytes: skx509Encoded}))
	skEncoded = replacer.Replace(skEncoded)

	pkx509Encoded, _ := x509.MarshalPKIXPublicKey(pk)
	pkEncoded := string(pem.EncodeToMemory(&pem.Block{Bytes: pkx509Encoded}))
	pkEncoded = replacer.Replace(pkEncoded)

	zap.S().Info("Successfully generated key!")
	return &skEncoded, &pkEncoded, nil
}
