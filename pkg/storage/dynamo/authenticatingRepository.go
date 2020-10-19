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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/dgrijalva/jwt-go"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/super-type/supertype/internal/keys"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/storage"
)

// generateJWT generates a JWT on user authentication
func generateJWT(username string) (*string, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	signingKey := os.Getenv("JWT_SIGNING_KEY")

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["user"] = username
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenStr, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return nil, err
	}

	return &tokenStr, nil
}

// generateSupertypeID generates a new Supertype ID for a given password
func generateSupertypeID(password string) (*string, error) {
	requestBody, err := json.Marshal(map[string]string{
		"password": password,
	})
	if err != nil {
		color.Red("Error marshaling data")
		return nil, storage.ErrMarshaling
	}

	resp, err := http.Post("https://z1lwetrbfe.execute-api.us-east-1.amazonaws.com/default/generate-nuid-credentials", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		color.Red("Error requesting Supertype API")
		return nil, authenticating.ErrRequestingAPI
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		color.Red("Can't read response body")
		return nil, authenticating.ErrResponseBody
	}

	var supertypeID string
	json.Unmarshal(body, &supertypeID)

	return &supertypeID, nil
}

// establishInitialConnections creates re-encryption keys between a newly-created vendor and all existing vendors
func establishInitialConnections(svc *dynamodb.DynamoDB, pkVendor *string) (*dynamodb.ScanOutput, error) {
	// Get all vendors' pks, except the vendor's own
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames: map[string]*string{
			"#pk": aws.String("pk"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {
				S: aws.String(*pkVendor),
			},
		},
		FilterExpression:     aws.String("pk <> :pk"),
		ProjectionExpression: aws.String("#pk"),
		TableName:            aws.String("vendor"),
	}

	result, err := svc.Scan(input)
	if err != nil {
		fmt.Printf("Err scanning vendor table: %v\n", err)
		return nil, err
	}

	return result, nil
}

// CreateVendor creates a new vendor and adds it to DynamoDB
func (d *Storage) CreateVendor(v authenticating.Vendor) (*[2]string, error) {
	// Initialize AWS session
	svc := utils.SetupAWSSession()

	// Get username from DynamoDB
	result, err := GetFromDynamoDB(svc, "vendor", "username", v.Username)
	if err != nil {
		return nil, err
	}
	vendor := Vendor{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &vendor)
	if err != nil {
		return nil, storage.ErrUnmarshaling
	}

	// Check username doesn't exist
	if vendor.Username != "" {
		color.Red("Vendor already exists")
		return nil, authenticating.ErrVendorAlreadyExists
	}

	// Generate key pair for new vendor
	skVendor, pkVendor, err := keys.GenerateKeys()
	if err != nil {
		color.Red("Failed to generate keys")
		return nil, keys.ErrFailedToGenerateKeys
	}

	// Generate hash of secret key to be used as a signing measure for producing/consuming data
	h := sha256.New()
	// h.Write([]byte(utils.PrivateKeyToString(skVendor)))
	// TODO do we want to be storing this skVendor.D...? I feel like we should just hash the sk value but maybe it doesn't matter...
	h.Write([]byte(*skVendor))
	skHash := hex.EncodeToString(h.Sum(nil))

	// Generate Supertype ID
	supertypeID, err := generateSupertypeID(v.Password)
	if err != nil {
		return nil, err
	}

	// Cursory check for valid email address
	if !ValidateEmail(v.Email) {
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
		SkHash:         skHash,
		SupertypeID:    *supertypeID,
		AccountBalance: 0.0,
	}

	// Upload new vendor to DynamoDB
	av, err := dynamodbattribute.MarshalMap(createVendor)
	if err != nil {
		color.Red("Error marshaling data")
		return nil, storage.ErrMarshaling
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("vendor"),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		color.Red("Failed to write to database")
		return nil, storage.ErrFailedToWriteDB
	}

	keyPair := [2]string{*pkVendor, *skVendor}

	return &keyPair, nil
}

// CreateUser creates a new user and adds it to DynamoDB
func (d *Storage) CreateUser(u authenticating.UserPassword) (*string, error) {
	// Initialize AWS session
	svc := utils.SetupAWSSession()

	// Get username from DynamoDB
	result, err := GetFromDynamoDB(svc, "user", "username", u.Username)
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
	supertypeID, err := generateSupertypeID(u.Password)
	if err != nil {
		return nil, err
	}

	// Create a final user with which to upload
	createUser := authenticating.User{
		Username:    u.Username,
		SupertypeID: *supertypeID,
	}

	// Upload new user to DynamoDB
	av, err := dynamodbattribute.MarshalMap(createUser)
	if err != nil {
		color.Red("Error marshaling data")
		return nil, storage.ErrMarshaling
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("user"),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		color.Red("Failed to write to database")
		return nil, storage.ErrFailedToWriteDB
	}

	success := "success"

	return &success, nil
}

// LoginVendor logs in the given vendor to the repository
func (d *Storage) LoginVendor(v authenticating.Vendor) (*authenticating.AuthenticatedVendor, error) {
	// Initialize AWS Session
	svc := utils.SetupAWSSession()

	// Get username from DynamoDB
	result, err := GetFromDynamoDB(svc, "vendor", "username", v.Username)
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

	jwt, err := generateJWT(vendor.Username)
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
	result, err := GetFromDynamoDB(svc, "user", "username", u.Username)
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
