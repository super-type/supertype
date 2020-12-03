package authenticating

import "errors"

// ErrVendorNotFound is used when a vendor is not found in the database
var ErrVendorNotFound = errors.New("Vendor not found")

// ErrVendorAlreadyExists is used when attempting to create an already-used username
var ErrVendorAlreadyExists = errors.New("Vendor already exists")

// ErrUserNotFound is used when a user is not found in the database
var ErrUserNotFound = errors.New("User not found")

// ErrUserAlreadyExists is used when attempting to create an already-used username
var ErrUserAlreadyExists = errors.New("User already exists")

// ErrRequestingAPI is used when there is an issue requested the Supertype auth lambda
var ErrRequestingAPI = errors.New("API Error")

// ErrResponseBody is used when we are unable to read response body
var ErrResponseBody = errors.New("Error reading response body")

// ErrNotAuthorized is used when a request does not contain a valid token
var ErrNotAuthorized = errors.New("Not Authorized")

// ErrGeneratingCipherBlock is used when there's an issue generating the cipher block
var ErrGeneratingCipherBlock = errors.New("Error generating cipher block")

// ErrGeneratingIV is used when there's an issue generating the iv
var ErrGeneratingIV = errors.New("Error generating iv")

// ErrInvalidEmailLength is used when an invalid email address length is used to create an account
var ErrInvalidEmailLength = errors.New("Invalid email address length. Account creation failed.")

// ErrInvalidEmailMatching is used when eamil regex matching fails
var ErrInvalidEmailMatching = errors.New("Invalid email - bad matching. Account creation failed.")

// ErrInvalidEmail is used when an invalid email address is used to create an account
var ErrInvalidEmail = errors.New("Invalid email address used. Account creation failed.")
