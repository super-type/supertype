package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
)

// CURVE is a generic curve which implements P-256 (see FIPS 186-3, section D.2.3)
var CURVE = elliptic.P256()

// P is the order of the underlying field
var P = CURVE.Params().P

// N is the order of the base point
var N = CURVE.Params().N

// CurvePoint is an ECDSA public key
type CurvePoint = ecdsa.PublicKey

// PointScalarMul multiplies a point on the elliptic curve
func PointScalarMul(a *CurvePoint, k *big.Int) *CurvePoint {
	x, y := a.ScalarMult(a.X, a.Y, k.Bytes())
	return &CurvePoint{CURVE, x, y}
}

// PointToBytes converts a point on the elliptic curve into a byte array
func PointToBytes(point *CurvePoint) (res []byte) {
	res = elliptic.Marshal(CURVE, point.X, point.Y)
	return
}

// GenerateKeys returns a new key pair on the elliptic curve
func GenerateKeys() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	sk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return sk, &sk.PublicKey, nil
}
