package reencryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/super-type/supertype/internal/keys"
	"github.com/super-type/supertype/internal/utils"
)

type Capsule struct {
	E *ecdsa.PublicKey
	V *ecdsa.PublicKey
	S *big.Int
}

func encryptKeyGen(pubKey *ecdsa.PublicKey) (capsule *Capsule, keyBytes []byte, err error) {
	s := new(big.Int)
	// generate E,V key-pairs
	priE, pubE, err := keys.GenerateKeys()
	priV, pubV, err := keys.GenerateKeys()
	if err != nil {
		return nil, nil, err
	}
	// get H2(E || V)
	h := utils.HashToCurve(
		utils.ConcatBytes(
			keys.PointToBytes(pubE),
			keys.PointToBytes(pubV)))
	// get s = v + e * H2(E || V)
	s = utils.BigIntAdd(priV.D, utils.BigIntMul(priE.D, h))
	// get (pk_A)^{e+v}
	point := keys.PointScalarMul(pubKey, utils.BigIntAdd(priE.D, priV.D))
	// generate aes key
	keyBytes, err = utils.Sha3Hash(keys.PointToBytes(point))
	if err != nil {
		return nil, nil, err
	}
	capsule = &Capsule{
		E: pubE,
		V: pubV,
		S: s,
	}
	fmt.Println("old key:", hex.EncodeToString(keyBytes))
	return capsule, keyBytes, nil
}

// Encrypt the message
// AES GCM + Proxy Re-Encryption
func Encrypt(message string, pubKey *ecdsa.PublicKey) (cipherText []byte, capsule *Capsule, err error) {
	capsule, keyBytes, err := encryptKeyGen(pubKey)
	if err != nil {
		return nil, nil, err
	}
	key := hex.EncodeToString(keyBytes)
	// use aes gcm algorithm to encrypt
	// mark keyBytes[:12] as nonce
	cipherText, err = gcmEncrypt([]byte(message), key[:32], keyBytes[:12], nil)
	if err != nil {
		return nil, nil, err
	}
	return cipherText, capsule, nil
}

func gcmEncrypt(plaintext []byte, key string, iv []byte, additionalData []byte) (cipherText []byte, err error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	cipherText = aesgcm.Seal(nil, iv, plaintext, additionalData)
	return cipherText, nil
}

func gcmDecrypt(cipherText []byte, key string, iv []byte, additionalData []byte) (plainText []byte, err error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plainText, err = aesgcm.Open(nil, iv, cipherText, additionalData)
	if err != nil {
		return nil, err
	}
	return plainText, nil
}

// EncodeCapsule encodes capsule
func EncodeCapsule(capsule Capsule) (capsuleAsBytes []byte, err error) {
	gob.Register(elliptic.P256())
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err = enc.Encode(capsule); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeCapsule decodes capsule
func DecodeCapsule(capsuleAsBytes []byte) (capsule Capsule, err error) {
	capsule = Capsule{}
	gob.Register(elliptic.P256())
	dec := gob.NewDecoder(bytes.NewBuffer(capsuleAsBytes))
	if err = dec.Decode(&capsule); err != nil {
		return capsule, err
	}
	return capsule, nil
}

// generate re-encryption key and sends it to Server
// rk = sk_A * d^{-1}
func ReKeyGen(aPriKey *ecdsa.PrivateKey, bPubKey *ecdsa.PublicKey) (*big.Int, *ecdsa.PublicKey, error) {
	// generate x,X key-pair
	priX, pubX, err := keys.GenerateKeys()
	if err != nil {
		return nil, nil, err
	}
	// get d = H3(X_A || pk_B || pk_B^{x_A})
	point := keys.PointScalarMul(bPubKey, priX.D)
	d := utils.HashToCurve(
		utils.ConcatBytes(
			utils.ConcatBytes(
				keys.PointToBytes(pubX),
				keys.PointToBytes(bPubKey)),
			keys.PointToBytes(point)))
	// rk = sk_A * d^{-1}
	rk := utils.BigIntMul(aPriKey.D, utils.GetInvert(d))
	rk.Mod(rk, keys.N)
	return rk, pubX, nil
}

// Server executes Re-Encryption method
func ReEncryption(rk *big.Int, capsule *Capsule) (*Capsule, error) {
	// check g^s == V * E^{H2(E || V)}
	x1, y1 := keys.CURVE.ScalarBaseMult(capsule.S.Bytes())
	tempX, tempY := keys.CURVE.ScalarMult(capsule.E.X, capsule.E.Y,
		utils.HashToCurve(
			utils.ConcatBytes(
				keys.PointToBytes(capsule.E),
				keys.PointToBytes(capsule.V))).Bytes())
	x2, y2 := keys.CURVE.Add(capsule.V.X, capsule.V.Y, tempX, tempY)
	// if check failed return error
	if x1.Cmp(x2) != 0 || y1.Cmp(y2) != 0 {
		return nil, fmt.Errorf("%s", "Capsule not match")
	}
	// E' = E^{rk}, V' = V^{rk}
	newCapsule := &Capsule{
		E: keys.PointScalarMul(capsule.E, rk),
		V: keys.PointScalarMul(capsule.V, rk),
		S: capsule.S,
	}
	return newCapsule, nil
}

// Recreate the aes key then decrypt the cipherText
func Decrypt(bPriKey *ecdsa.PrivateKey, capsule *Capsule, pubX *ecdsa.PublicKey, cipherText []byte) (plainText []byte, err error) {
	keyBytes, err := decryptKeyGen(bPriKey, capsule, pubX)
	if err != nil {
		return nil, err
	}
	// recreate aes key = G((E' * V')^d)
	key := hex.EncodeToString(keyBytes)
	// use aes gcm to decrypt
	// mark keyBytes[:12] as nonce
	plainText, err = gcmDecrypt(cipherText, key[:32], keyBytes[:12], nil)
	if err != nil {
		return nil, err
	}
	return plainText, nil
}

func decryptKeyGen(bPriKey *ecdsa.PrivateKey, capsule *Capsule, pubX *ecdsa.PublicKey) (keyBytes []byte, err error) {
	// S = X_A^{sk_B}
	S := keys.PointScalarMul(pubX, bPriKey.D)
	// recreate d = H3(X_A || pk_B || S)
	d := utils.HashToCurve(
		utils.ConcatBytes(
			utils.ConcatBytes(
				keys.PointToBytes(pubX),
				keys.PointToBytes(&bPriKey.PublicKey)),
			keys.PointToBytes(S)))
	point := keys.PointScalarMul(
		keys.PointScalarAdd(capsule.E, capsule.V), d)
	keyBytes, err = utils.Sha3Hash(keys.PointToBytes(point))
	if err != nil {
		return nil, err
	}
	return keyBytes, nil
}
