package dynamo

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/super-type/supertype/internal/keys"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/storage"
	"go.uber.org/zap"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// CreateVendor creates a new vendor and adds it to DynamoDB
func (d *Storage) CreateVendor(v authenticating.Vendor) (*[2]string, error) {
	zap.S().Info("Creating new vendor...")
	// Initialize AWS session
	svc := utils.SetupAWSSession()

	// TODO we need a nice util function to get multiple attributes from the DB (i.e. username and email)

	// Get username from DynamoDB
	result, err := GetItemDynamoDB(svc, "vendor", "username", v.Username)
	if err != nil {
		zap.S().Errorf("Failed to get username %s from DynamoDB: %v", v.Username, err)
		return nil, err
	}
	vendor := Vendor{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &vendor)
	if err != nil {
		zap.S().Errorf("Error unmarshaling vendor: %v", err)
		return nil, storage.ErrUnmarshaling
	}

	// Get email from DynamoDB
	email, err := ScanDynamoDB("vendor", "email", v.Email)
	if err != nil {
		zap.S().Errorf("Error getting email %s from DynamoDB: %v", v.Email, err)
		return nil, err
	}

	// Check username, email doesn't exist
	if vendor.Username != "" {
		zap.S().Errorf("Username %s already exists", vendor.Username)
		return nil, authenticating.ErrAlreadyExists
	}

	if len(email.Items) != 0 {
		zap.S().Errorf("Email %s already exists", vendor.Username)
		return nil, authenticating.ErrAlreadyExists
	}

	// Check email is a valid email address
	// TODO in a later refactor, this should be taken out of this function and put inside another... this is business logic, not database logic i.e. not depending on DynamoDB
	if len(v.Email) < 3 && len(v.Email) > 254 {
		zap.S().Errorf("Email %s invalid length %v", v.Email, len(v.Email))
		return nil, authenticating.ErrInvalidEmailLength
	}

	if !emailRegex.MatchString(v.Email) {
		zap.S().Errorf("Email %s does not match regex", v.Email)
		return nil, authenticating.ErrInvalidEmailMatching
	}
	parts := strings.Split(v.Email, "@")
	mx, err := net.LookupMX(parts[1])
	if err != nil || len(mx) == 0 {
		zap.S().Errorf("Error looking up mx: %v", err)
		return nil, authenticating.ErrInvalidEmail
	}

	// Generate key pair for new vendor
	skVendor, pkVendor, err := keys.GenerateKeys()
	if err != nil {
		zap.S().Errorf("Failed to generate keys: %v", err)
		return nil, keys.ErrFailedToGenerateKeys
	}

	// Generate hash of secret key to be used as a signing measure for producing/consuming data
	h := sha256.New()
	h.Write([]byte(*skVendor))
	apiKeyHash := hex.EncodeToString(h.Sum(nil))

	// Generate Supertype ID
	supertypeID, err := utils.GenerateSupertypeID(v.Password)
	if err != nil {
		zap.S().Errorf("Error generating SupertypeID: %v", err)
		return nil, err
	}

	// Cursory check for valid email address
	if !utils.ValidateEmail(v.Email) {
		zap.S().Errorf("Email %s invalid", v.Email)
		return nil, storage.ErrInvalidEmail
	}

	// Create a final vendor with which to upload
	createVendor := authenticating.CreateVendor{
		FirstName:      v.FirstName,
		LastName:       v.LastName,
		Email:          v.Email,
		BusinessName:   v.BusinessName,
		Username:       v.Username,
		PublicKey:      *pkVendor,
		APIKeyHash:     apiKeyHash,
		SupertypeID:    *supertypeID,
		AccountBalance: 0.0,
	}

	err = PutItemInDynamoDB(createVendor, "vendor", svc)
	if err != nil {
		zap.S().Errorf("Error putting vendor %s in DynamoDB: %v", fmt.Sprint(createVendor), err)
		return nil, err
	}

	keyPair := [2]string{*pkVendor, *skVendor}

	zap.S().Info("Successfully created new vendor!")
	return &keyPair, nil
}

// CreateUser creates a new user and adds it to DynamoDB
func (d *Storage) CreateUser(u authenticating.UserPassword) (*string, error) {
	zap.S().Info("Creating new user...")
	// Initialize AWS session
	svc := utils.SetupAWSSession()

	// Get username from DynamoDB
	result, err := GetItemDynamoDB(svc, "user", "username", u.Username)
	if err != nil {
		zap.S().Errorf("Failed to get username %s from DynamoDB: %v", u.Username, err)
		return nil, err
	}
	user := User{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		zap.S().Errorf("Error unmarshaling user: %v", err)
		return nil, storage.ErrUnmarshaling
	}

	// Check username doesn't exist
	if user.Username != "" {
		zap.S().Errorf("Error getting username %s from DynamoDB", user.Username)
		return nil, authenticating.ErrUserAlreadyExists
	}

	// Generate Supertype ID
	supertypeID, err := utils.GenerateSupertypeID(u.Password)
	if err != nil {
		zap.S().Errorf("Error generating SupertypeID: %v", err)
		return nil, err
	}

	// Create a final user with which to upload
	createUser := authenticating.User{
		Username:    u.Username,
		SupertypeID: *supertypeID,
	}

	// Upload new user to DynamoDB
	err = PutItemInDynamoDB(createUser, "user", svc)
	if err != nil {
		zap.S().Errorf("Error putting vendor %s in DynamoDB: %v", fmt.Sprint(createUser), err)
		return nil, err
	}

	success := "success"

	zap.S().Info("Successfully created new user!")
	return &success, nil
}

// LoginVendor logs in the given vendor to the repository
func (d *Storage) LoginVendor(v authenticating.Vendor) (*authenticating.AuthenticatedVendor, error) {
	zap.S().Info("Logging vendor in...")
	// Initialize AWS Session
	svc := utils.SetupAWSSession()

	// Get username from DynamoDB
	result, err := GetItemDynamoDB(svc, "vendor", "username", v.Username)
	if err != nil {
		zap.S().Errorf("Failed to get username %s from DynamoDB: %v", v.Username, err)
		return nil, err
	}
	vendor := authenticating.AuthenticatedVendor{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &vendor)
	if err != nil {
		zap.S().Errorf("Error unmarshaling user: %v", err)
		return nil, storage.ErrUnmarshaling
	}

	// Check vendor exists and get object
	if vendor.Username == "" {
		zap.S().Errorf("Error getting username %s from DynamoDB", vendor.Username)
		return nil, authenticating.ErrVendorNotFound
	}

	// Authenticate with NuID
	requestBody, err := json.Marshal(map[string]string{
		"password":    v.Password,
		"supertypeID": vendor.SupertypeID,
	})
	if err != nil {
		zap.S().Errorf("Error encoding request: %v", err)
		return nil, storage.ErrEncoding
	}

	resp, err := http.Post("https://z1lwetrbfe.execute-api.us-east-1.amazonaws.com/default/login-vendor", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		zap.S().Errorf("Error requesting Supertype API: %v", err)
		return nil, authenticating.ErrRequestingAPI
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		zap.S().Errorf("API request gave bad response status %v", resp.StatusCode)
		return nil, authenticating.ErrRequestingAPI
	}

	jwt, err := utils.GenerateJWT(vendor.Username)
	if err != nil {
		zap.S().Errorf("Could not generate JWT: %v", err)
		return nil, authenticating.ErrRequestingAPI
	}
	vendor.JWT = *jwt

	return &vendor, nil
}

// LoginUser logs in the given user to the repository
func (d *Storage) LoginUser(u authenticating.UserPassword) (*authenticating.User, error) {
	zap.S().Info("Creating user in...")
	// Initialize AWS Session
	svc := utils.SetupAWSSession()

	// Get username from DynamoDB
	result, err := GetItemDynamoDB(svc, "user", "username", u.Username)
	if err != nil {
		zap.S().Errorf("Failed to get username %s from DynamoDB: %v", u.Username, err)
		return nil, err
	}
	user := authenticating.User{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		zap.S().Errorf("Error unmarshaling user: %v", err)
		return nil, storage.ErrUnmarshaling
	}

	// Check user exists and get object
	if user.Username == "" {
		zap.S().Errorf("Error getting username %s from DynamoDB", user.Username)
		return nil, authenticating.ErrUserNotFound
	}

	// Authenticate with NuID
	requestBody, err := json.Marshal(map[string]string{
		"password":    u.Password,
		"supertypeID": user.SupertypeID,
	})
	if err != nil {
		zap.S().Errorf("Error encoding request %s : %v", fmt.Sprint(requestBody), err)
		return nil, storage.ErrEncoding
	}

	resp, err := http.Post("https://z1lwetrbfe.execute-api.us-east-1.amazonaws.com/default/login-vendor", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		zap.S().Errorf("Error requesting Supertype API: %v", err)
		return nil, authenticating.ErrRequestingAPI
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		zap.S().Errorf("API request gave bad response status %v", resp.StatusCode)
		return nil, authenticating.ErrRequestingAPI
	}

	// Key returned on login is the SupertypeID encrypted with the correct password as the AES encryption key
	generationKey := []byte(u.Password)
	generationPlaintext := []byte(user.SupertypeID)

	// Encrypt
	block, err := aes.NewCipher(generationKey[0:32])
	if err != nil {
		zap.S().Errorf("Error generating new cipher block: %v", err)
		return nil, authenticating.ErrGeneratingCipherBlock
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		zap.S().Errorf("Error generating initialization vector: %v", err)
		return nil, authenticating.ErrGeneratingIV
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(generationPlaintext))
	cfb.XORKeyStream(ciphertext, generationPlaintext)

	// Set userKey value to return on login
	user.UserKey = base64.StdEncoding.EncodeToString(ciphertext)

	zap.S().Info("Successfully logged in user!")
	return &user, nil
}

// AuthorizedLoginUser logs in the given user to the repository
func (d *Storage) AuthorizedLoginUser(u authenticating.UserPassword, apiKey string) (*authenticating.User, error) {
	zap.S().Info("Creating new user...")
	// Initialize AWS Session
	svc := utils.SetupAWSSession()

	// Get username from DynamoDB
	result, err := GetItemDynamoDB(svc, "user", "username", u.Username)
	if err != nil {
		zap.S().Errorf("Failed to get username %s from DynamoDB: %v", u.Username, err)
		return nil, err
	}
	user := authenticating.User{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		zap.S().Errorf("Error unmarshaling user: %v", err)
		return nil, storage.ErrUnmarshaling
	}

	// Check user exists and get object
	if user.Username == "" {
		zap.S().Errorf("Error getting username %s from DynamoDB", user.Username)
		return nil, authenticating.ErrUserNotFound
	}

	// Authenticate with NuID
	requestBody, err := json.Marshal(map[string]string{
		"password":    u.Password,
		"supertypeID": user.SupertypeID,
	})
	if err != nil {
		zap.S().Errorf("Error encoding request: %v", err)
		return nil, storage.ErrEncoding
	}

	resp, err := http.Post("https://z1lwetrbfe.execute-api.us-east-1.amazonaws.com/default/login-vendor", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		zap.S().Errorf("Error requesting Supertype API: %v", err)
		return nil, authenticating.ErrRequestingAPI
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		zap.S().Errorf("API request gave bad response status %v", resp.StatusCode)
		return nil, authenticating.ErrRequestingAPI
	}

	// Key returned on login is the SupertypeID encrypted with the correct password as the AES encryption key
	generationKey := []byte(u.Password)
	generationPlaintext := []byte(user.SupertypeID)

	// Encrypt
	block, err := aes.NewCipher(generationKey[0:32])
	if err != nil {
		zap.S().Errorf("Error generating new cipher block: %v", err)
		return nil, authenticating.ErrGeneratingCipherBlock
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		zap.S().Errorf("Error generating initialization vector: %v", err)
		return nil, authenticating.ErrGeneratingIV
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(generationPlaintext))
	cfb.XORKeyStream(ciphertext, generationPlaintext)

	// Set userKey value to return on login
	user.UserKey = base64.StdEncoding.EncodeToString(ciphertext)

	apiKeyHash := utils.GetAPIKeyHash(apiKey)

	// Get vendor's public key given the vendor's API Key
	pk, err := ScanDynamoDBWithKeyCondition("vendor", "pk", "apiKeyHash", apiKeyHash)
	if err != nil {
		zap.S().Errorf("Error getting vendor's public key given API Key: %v", err)
		return nil, err
	}

	if pk == nil {
		zap.S().Errorf("Public Key is empty")
	}

	pkAlreadyExists, err := ScanDynamoDBWithKeyCondition("user", "pk", "pk", *pk)
	if err != nil {
		zap.S().Errorf("Public key already exists: %v", err)
		return nil, err
	}

	if pkAlreadyExists == nil {
		userWithVendors := authenticating.UserWithVendors{}
		userWithVendors.SupertypeID = user.SupertypeID
		userWithVendors.Username = user.Username
		// Associate vendor with user
		userWithVendors.Vendors = append(userWithVendors.Vendors, *pk)

		// Upload updated user to DynamoDB
		err = PutItemInDynamoDB(userWithVendors, "user", svc)
		if err != nil {
			zap.S().Errorf("Error putting vendor %s in DynamoDB: %v", fmt.Sprint(userWithVendors), err)
			return nil, err
		}
	}

	zap.S().Info("Successfully logged in authorized user!")
	return &user, nil
}
