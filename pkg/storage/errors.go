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
