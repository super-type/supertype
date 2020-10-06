package dynamo

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
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

// rk = sk_A * d^{-1}
func reKeyGen(aPriKey *ecdsa.PrivateKey, bPubKey *ecdsa.PublicKey) (*big.Int, *ecdsa.PublicKey, error) {
	// generate x,X key-pair
	priX, pubX, err := keys.GenerateKeys()
	if err != nil {
		return nil, nil, err
	}
	// get d = H3(X_A || pk_B || pk_B^{x_A})
	point := keys.PointScalarMul(bPubKey, priX.D)
	d := utils.HashToCurve(utils.ConcatBytes(utils.ConcatBytes(keys.PointToBytes(pubX), keys.PointToBytes(bPubKey)), keys.PointToBytes(point)))
	// rk = sk_A * d^{-1}
	rk := utils.BigIntMul(aPriKey.D, utils.GetInvert(d))
	rk.Mod(rk, keys.N)
	return rk, pubX, nil
}

// createReencryptionKeys creates re-encryption keys for public keys returned from DynamoDB
// todo vulnerability where if we have even one entry without a public key, this loop will throw a NPE
func createReencryptionKeys(pkList *dynamodb.ScanOutput, skVendor *ecdsa.PrivateKey) (map[string][2]string, error) {
	connections := make(map[string][2]string)

	for i := 0; i < len(pkList.Items); i++ {
		pkTempStr := *(pkList.Items[i]["pk"].S)
		pkTemp, err := utils.StringToPublicKey(&pkTempStr)
		if err != nil {
			fmt.Printf("Error converting public key string to ECDSA Public Key: %v\n", err)
		}

		// Create re-encryption keys between each pairing
		rekey, pkX, err := reKeyGen(skVendor, &pkTemp)
		if err != nil {
			fmt.Printf("Error generating re-encryption key: %v\n", err)
		}

		rekeyStr := rekey.String()
		pkXStr := utils.PublicKeyToString(pkX)

		connections[pkTempStr] = [2]string{rekeyStr, pkXStr}
	}

	return connections, nil
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
	h.Write([]byte(skVendor.D.String()))
	skHash := hex.EncodeToString(h.Sum(nil))

	// Generate Supertype ID
	supertypeID, err := generateSupertypeID(v.Password)
	if err != nil {
		return nil, err
	}

	// Generate re-encryption keys
	pk := utils.PublicKeyToString(pkVendor)
	pkList, err := establishInitialConnections(svc, &pk)
	if err != nil {
		color.Red("Failed to get list of public keys")
		return nil, storage.ErrGetListPublicKeys
	}
	rekeys, err := createReencryptionKeys(pkList, skVendor)
	if err != nil {
		color.Red("Failed to generate re-encryption keys")
		return nil, keys.ErrFailedToGenerateReencryptionKeys
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
		PublicKey:      utils.PublicKeyToString(pkVendor),
		SkHash:         skHash,
		SupertypeID:    *supertypeID,
		Connections:    rekeys,
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

	// TODO why are we returning the skVendor.D.String() here...?
	keyPair := [2]string{utils.PublicKeyToString(pkVendor), skVendor.D.String()}

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

	return &user, nil
}
