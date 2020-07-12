package keys

import "errors"

// ErrFailedToGenerateKeys is used when we fail to generate vendor key-pair
var ErrFailedToGenerateKeys = errors.New("Failed to generate vendor key-pair")
