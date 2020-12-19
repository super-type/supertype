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
	"github.com/fatih/color"
	"github.com/super-type/supertype/internal/keys"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/storage"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// CreateVendor creates a new vendor and adds it to DynamoDB
func (d *Storage) CreateVendor(v authenticating.Vendor) (*[2]string, error) {
	// Initialize AWS session
	svc := utils.SetupAWSSession()

	// TODO we need a nice util function to get multiple attributes from the DB (i.e. username and email)

	// Get username from DynamoDB
	result, err := GetItemDynamoDB(svc, "vendor", "username", v.Username)
	if err != nil {
		return nil, err
	}
	user := Vendor{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, storage.ErrUnmarshaling
	}

	// Get email from DynamoDB
	email, err := ScanDynamoDB("vendor", "email", v.Email)
	if err != nil {
		return nil, err
	}

	// Check username, email doesn't exist
	if user.Username != "" || len(email.Items) != 0 {
		color.Red("Username or email already exists")
		return nil, authenticating.ErrVendorAlreadyExists
	}

	// Check email is a valid email address
	// TODO in a later refactor, this should be taken out of this function and put inside another... this is business logic, not database logic i.e. not depending on DynamoDB
	if len(v.Email) < 3 && len(v.Email) > 254 {
		return nil, authenticating.ErrInvalidEmailLength
	}

	if !emailRegex.MatchString(v.Email) {
		return nil, authenticating.ErrInvalidEmailMatching
	}
	parts := strings.Split(v.Email, "@")
	mx, err := net.LookupMX(parts[1])
	if err != nil || len(mx) == 0 {
		return nil, authenticating.ErrInvalidEmail
	}

	// Generate key pair for new vendor
	skVendor, pkVendor, err := keys.GenerateKeys()
	if err != nil {
		color.Red("Failed to generate keys")
		return nil, keys.ErrFailedToGenerateKeys
	}

	// Generate hash of secret key to be used as a signing measure for producing/consuming data
	h := sha256.New()
	h.Write([]byte(*skVendor))
	apiKeyHash := hex.EncodeToString(h.Sum(nil))

	// Generate Supertype ID
	supertypeID, err := utils.GenerateSupertypeID(v.Password)
	if err != nil {
		return nil, err
	}

	// Cursory check for valid email address
	if !utils.ValidateEmail(v.Email) {
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
		return nil, err
	}

	keyPair := [2]string{*pkVendor, *skVendor}

	return &keyPair, nil
}

// CreateUser creates a new user and adds it to DynamoDB
func (d *Storage) CreateUser(u authenticating.UserPassword) (*string, error) {
	// Initialize AWS session
	svc := utils.SetupAWSSession()

	// Get username from DynamoDB
	result, err := GetItemDynamoDB(svc, "user", "username", u.Username)
	if err != nil {
		return nil, err
	}
	user := User{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, storage.ErrUnmarshaling
	}

	// Check username doesn't exist
	if user.Username != "" {
		color.Red("User already exists")
		return nil, authenticating.ErrUserAlreadyExists
	}

	// Generate Supertype ID
	supertypeID, err := utils.GenerateSupertypeID(u.Password)
	if err != nil {
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
		return nil, err
	}

	success := "success"

	return &success, nil
}

// LoginVendor logs in the given vendor to the repository
func (d *Storage) LoginVendor(v authenticating.Vendor) (*authenticating.AuthenticatedVendor, error) {
	// Initialize AWS Session
	svc := utils.SetupAWSSession()

	// Get username from DynamoDB
	result, err := GetItemDynamoDB(svc, "vendor", "username", v.Username)
	if err != nil {
		return nil, err
	}
	vendor := authenticating.AuthenticatedVendor{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &vendor)
	if err != nil {
		color.Red("Error unmarshaling data")
		return nil, storage.ErrUnmarshaling
	}

	// Check vendor exists and get object
	if vendor.Username == "" {
		color.Red("Vendor not found")
		return nil, authenticating.ErrVendorNotFound
	}

	// Authenticate with NuID
	requestBody, err := json.Marshal(map[string]string{
		"password":    v.Password,
		"supertypeID": vendor.SupertypeID,
	})
	if err != nil {
		color.Red("Error encoding data")
		return nil, storage.ErrEncoding
	}

	resp, err := http.Post("https://z1lwetrbfe.execute-api.us-east-1.amazonaws.com/default/login-vendor", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		color.Red("Error requesting Supertype API")
		return nil, authenticating.ErrRequestingAPI
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		color.Red("API request gave bad response status")
		return nil, authenticating.ErrRequestingAPI
	}

	jwt, err := utils.GenerateJWT(vendor.Username)
	if err != nil {
		color.Red("Could not generate JWT")
		return nil, authenticating.ErrRequestingAPI
	}
	vendor.JWT = *jwt

	return &vendor, nil
}

// LoginUser logs in the given user to the repository
func (d *Storage) LoginUser(u authenticating.UserPassword) (*authenticating.User, error) {
	// Initialize AWS Session
	svc := utils.SetupAWSSession()

	// Get username from DynamoDB
	result, err := GetItemDynamoDB(svc, "user", "username", u.Username)
	if err != nil {
		return nil, err
	}
	user := authenticating.User{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		color.Red("Error unmarshaling data")
		return nil, storage.ErrUnmarshaling
	}

	// Check user exists and get object
	if user.Username == "" {
		color.Red("User not found")
		return nil, authenticating.ErrUserNotFound
	}

	// Authenticate with NuID
	requestBody, err := json.Marshal(map[string]string{
		"password":    u.Password,
		"supertypeID": user.SupertypeID,
	})
	if err != nil {
		color.Red("Error encoding data")
		return nil, storage.ErrEncoding
	}

	resp, err := http.Post("https://z1lwetrbfe.execute-api.us-east-1.amazonaws.com/default/login-vendor", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		color.Red("Error requesting Supertype API")
		return nil, authenticating.ErrRequestingAPI
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		color.Red("API request gave bad response status")
		return nil, authenticating.ErrRequestingAPI
	}

	// Key returned on login is the SupertypeID encrypted with the correct password as the AES encryption key
	generationKey := []byte(u.Password)
	generationPlaintext := []byte(user.SupertypeID)

	// Encrypt
	block, err := aes.NewCipher(generationKey[0:32])
	if err != nil {
		return nil, authenticating.ErrGeneratingCipherBlock
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, authenticating.ErrGeneratingIV
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(generationPlaintext))
	cfb.XORKeyStream(ciphertext, generationPlaintext)

	// Set userKey value to return on login
	user.UserKey = base64.StdEncoding.EncodeToString(ciphertext)

	return &user, nil
}

// AuthorizedLoginUser logs in the given user to the repository
func (d *Storage) AuthorizedLoginUser(u authenticating.UserPassword, apiKey string) (*authenticating.User, error) {
	// Initialize AWS Session
	svc := utils.SetupAWSSession()

	// Get username from DynamoDB
	result, err := GetItemDynamoDB(svc, "user", "username", u.Username)
	if err != nil {
		return nil, err
	}
	user := authenticating.User{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		color.Red("Error unmarshaling data")
		return nil, storage.ErrUnmarshaling
	}

	// Check user exists and get object
	if user.Username == "" {
		color.Red("User not found")
		return nil, authenticating.ErrUserNotFound
	}

	// Authenticate with NuID
	requestBody, err := json.Marshal(map[string]string{
		"password":    u.Password,
		"supertypeID": user.SupertypeID,
	})
	if err != nil {
		color.Red("Error encoding data")
		return nil, storage.ErrEncoding
	}

	resp, err := http.Post("https://z1lwetrbfe.execute-api.us-east-1.amazonaws.com/default/login-vendor", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		color.Red("Error requesting Supertype API")
		return nil, authenticating.ErrRequestingAPI
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		color.Red("API request gave bad response status")
		return nil, authenticating.ErrRequestingAPI
	}

	// Key returned on login is the SupertypeID encrypted with the correct password as the AES encryption key
	generationKey := []byte(u.Password)
	generationPlaintext := []byte(user.SupertypeID)

	// Encrypt
	block, err := aes.NewCipher(generationKey[0:32])
	if err != nil {
		return nil, authenticating.ErrGeneratingCipherBlock
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, authenticating.ErrGeneratingIV
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(generationPlaintext))
	cfb.XORKeyStream(ciphertext, generationPlaintext)

	// Set userKey value to return on login
	user.UserKey = base64.StdEncoding.EncodeToString(ciphertext)

	apiKeyHash := utils.GetAPIKeyHash(apiKey)

	// Get venor's public key given the vendor's API Key
	pk, err := ScanDynamoDBWithKeyCondition("vendor", "pk", "apiKeyHash", apiKeyHash)
	if err != nil || pk == nil {
		fmt.Println(err)
		return nil, err
	}

	pkAlreadyExists, err := ScanDynamoDBWithKeyCondition("user", "pk", "pk", *pk)
	if err != nil {
		fmt.Println(err)
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
			return nil, err
		}
	}

	return &user, nil
}
