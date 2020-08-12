package dashboard

import "errors"

// ErrDynamoInternalError is used when there is an internal service error in DynamoDB
var ErrDynamoInternalError = errors.New("Internal server error in DynamoDB")

// ErrDynamoError is used when there is an internal service error in DynamoDB
var ErrDynamoError = errors.New("Internal server error in DynamoDB")
