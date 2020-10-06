package storage

import "errors"

// ErrFailedToReadDB is used when there is an unsuccessful DB read
var ErrFailedToReadDB = errors.New("Failed to read from DB")

// ErrFailedToWriteDB is used when there is an unsuccessful DB write
var ErrFailedToWriteDB = errors.New("Failed to write to DB")

// ErrUnmarshaling is used when we fail to unmarshal data
var ErrUnmarshaling = errors.New("Failed to unmarshal data")

// ErrMarshaling is used when we fail to marshal data
var ErrMarshaling = errors.New("Error marshaling vendor")

// ErrEncoding is used when we fail to encode JSON data
var ErrEncoding = errors.New("Error encoding JSON data")

// ErrGetListPublicKeys is used when we fail to get a list of public keys for a given vendor
var ErrGetListPublicKeys = errors.New("Error getting list of public keys")

// ErrNoObservationsForEntity is used when a vendor requests an observation for an entity who has no observations
var ErrNoObservationsForEntity = errors.New("No observations found for entity")

// ErrSkHashDoesNotMatch is used when secret keys do not match, which may flag a potential malicious attempt...
var ErrSkHashDoesNotMatch = errors.New("Secret keys do not match... Potential malicious event")

// ErrInvalidEmail is used when a user attempts to use an invalid email address
var ErrInvalidEmail = errors.New("Invalid email")
