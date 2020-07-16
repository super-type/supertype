package keys

import "errors"

// ErrFailedToGenerateKeys is used when we fail to generate vendor key-pair
var ErrFailedToGenerateKeys = errors.New("Failed to generate vendor key-pair")

// ErrFailedToGenerateReencryptionKeys is used when we fail to generate re-encryption keys
var ErrFailedToGenerateReencryptionKeys = errors.New("Failed to generate re-encryption keys")
