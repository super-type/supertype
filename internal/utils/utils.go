package utils

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/super-type/supertype/internal/keys"
)

// HashToCurve maps hash value to curve
func HashToCurve(hash []byte) *big.Int {
	hashInt := new(big.Int).SetBytes(hash)
	return hashInt.Mod(hashInt, keys.N)
}

// ConcatBytes concatenates a and b
func ConcatBytes(a, b []byte) []byte {
	var buf bytes.Buffer
	buf.Write(a)
	buf.Write(b)
	return buf.Bytes()
}

// BigIntMul multiplies two BigIntegers
func BigIntMul(a, b *big.Int) (res *big.Int) {
	res = new(big.Int).Mul(a, b)
	res.Mod(res, keys.N)
	return
}

// GetInvert gets the inverse of a BigInteger
func GetInvert(a *big.Int) (res *big.Int) {
	res = new(big.Int).ModInverse(a, keys.N)
	return
}

// PrivateKeyToString converts private key to string
func PrivateKeyToString(privateKey *ecdsa.PrivateKey) string {
	return hex.EncodeToString(privateKey.D.Bytes())
}

// PublicKeyToString converts public key to string
func PublicKeyToString(publicKey *ecdsa.PublicKey) string {
	pubKeyBytes := keys.PointToBytes(publicKey)
	return hex.EncodeToString(pubKeyBytes)
}

// StringToPublicKey converts a string back into an ECDSA Public Key
func StringToPublicKey(pkString *string) (ecdsa.PublicKey, error) {
	pkTempBytes, err := hex.DecodeString(*pkString)
	if err != nil {
		fmt.Printf("Error decoding bytes from string: %v\n", err)
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), pkTempBytes)
	publicKeyFinal := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	return publicKeyFinal, nil
}
