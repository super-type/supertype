package utils

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"

	"github.com/super-type/supertype/internal/keys"
	"golang.org/x/crypto/sha3"
)

// HashToCurve maps hash value to curve
func HashToCurve(hash []byte) *big.Int {
	hashInt := new(big.Int).SetBytes(hash)
	return hashInt.Mod(hashInt, keys.N)
}

// Sha3Hash converts plaintext to hash value
func Sha3Hash(message []byte) ([]byte, error) {
	sha := sha3.New256()
	_, err := sha.Write(message)
	if err != nil {
		return nil, err
	}
	return sha.Sum(nil), nil
}

// ConcatBytes concatenates a and b
func ConcatBytes(a, b []byte) []byte {
	var buf bytes.Buffer
	buf.Write(a)
	buf.Write(b)
	return buf.Bytes()
}

// BigIntAdd adds two BigIntegers
func BigIntAdd(a, b *big.Int) (res *big.Int) {
	res = new(big.Int).Add(a, b)
	res.Mod(res, keys.N)
	return
}

// BigIntSub subtracts two BigIntegers
func BigIntSub(a, b *big.Int) (res *big.Int) {
	res = new(big.Int)
	res.Sub(a, b)
	res.Mod(res, keys.N)
	return
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

// convert private key to string
func PrivateKeyToString(privateKey *ecdsa.PrivateKey) string {
	return hex.EncodeToString(privateKey.D.Bytes())
}

// convert public key to string
func PublicKeyToString(publicKey *ecdsa.PublicKey) string {
	pubKeyBytes := keys.PointToBytes(publicKey)
	return hex.EncodeToString(pubKeyBytes)
}
