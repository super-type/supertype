package authenticating

import "errors"

// ErrVendorNotFound is used when a vendor is not found in the database
var ErrVendorNotFound = errors.New("Vendor not found")

// ErrVendorAlreadyExists is used when attempting to create an already-used username
var ErrVendorAlreadyExists = errors.New("Vendor already exists")

// ErrRequestingAPI is used when there is an issue requested the Supertype auth lambda
var ErrRequestingAPI = errors.New("API Error")

// ErrResponseBody is used when we are unable to read response body
var ErrResponseBody = errors.New("Error reading response body")

// ErrNotAuthorized is used when a request does not contain a valid token
var ErrNotAuthorized = errors.New("Not Authorized")
