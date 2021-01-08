package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/storage"
	"go.uber.org/zap"
)

// SetupAWSSession starts an AWS session
func SetupAWSSession() *dynamodb.DynamoDB {
	zap.S().Info("Starting AWS session...")
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		// Specify profile to load for the session's config
		// TODO do we need this?
		// Profile: "profile_name",

		// Provide SDK Config options, such as Region.
		Config: aws.Config{
			Region: aws.String("us-east-1"),
		},

		// Force enable Shared Config support
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)
	zap.S().Info("Successfully created DynamoDB client!")
	return svc
}

// Contains is just a basic slice contains function, as Golang doesn't have this
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// GenerateJWT generates a JWT on user authentication
func GenerateJWT(username string) (*string, error) {
	zap.S().Info("Generating JWT...")
	err := godotenv.Load()
	if err != nil {
		zap.S().Errorf("Error loading .env file")
		return nil, err
	}
	signingKey := os.Getenv("JWT_SIGNING_KEY")

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["user"] = username
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenStr, err := token.SignedString([]byte(signingKey))
	if err != nil {
		zap.S().Errorf("Error generating token string: %v", err)
		return nil, err
	}

	zap.S().Info("Successfully generated JWT!")
	return &tokenStr, nil
}

// GenerateSupertypeID generates a new Supertype ID for a given password
func GenerateSupertypeID(password string) (*string, error) {
	zap.S().Info("Generating Supertype ID...")
	requestBody, err := json.Marshal(map[string]string{
		"password": password,
	})
	if err != nil {
		zap.S().Errorf("Error encoding request %s : %v", fmt.Sprint(requestBody), err)
		return nil, storage.ErrMarshaling
	}

	resp, err := http.Post("https://z1lwetrbfe.execute-api.us-east-1.amazonaws.com/default/generate-nuid-credentials", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		zap.S().Errorf("Error requesting Supertype API: %v", err)
		return nil, authenticating.ErrRequestingAPI
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		zap.S().Errorf("Cannot read response body: %v", err)
		return nil, authenticating.ErrResponseBody
	}

	var supertypeID string
	json.Unmarshal(body, &supertypeID)

	zap.S().Infof("Successfully generated SupertypeID %s", supertypeID)
	return &supertypeID, nil
}

// ValidateEmail checks to see whether a valid email was entered
func ValidateEmail(email string) bool {
	var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if len(email) < 3 && len(email) > 254 {
		zap.S().Errorf("Invalid email length: %s", email)
		return false
	}
	return emailRegex.MatchString(email)
}

// IsAuthorized checks the given JWT to ensure vendor is authenticated
func IsAuthorized(endpoint func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key, Token")
		fmt.Printf("checking authed for %v\n", r.Header["Token"])
		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				err := godotenv.Load()
				if err != nil {
					zap.S().Errorf("Error loading .env file: %v", err)
					return nil, authenticating.ErrNotAuthorized
				}

				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					zap.S().Error("Error signing token")
					return nil, authenticating.ErrNotAuthorized
				}

				return []byte(os.Getenv("JWT_SIGNING_KEY")), nil
			})
			if err != nil {
				zap.S().Errorf("Error generating token: %v", err)
				return
			}

			if token.Valid {
				endpoint(w, r)
			}
		} else {
			zap.S().Error("Token was nil")
		}
	})
}

// GetAPIKeyHash returns the hashed value of the secret key
func GetAPIKeyHash(skVendor string) string {
	h := sha256.New()
	h.Write([]byte(skVendor))
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateNewSubscriberURL determines whether or not a requested Webhook URL already exists
// TODO move this to a better file location on reorg
func ValidateNewSubscriberURL(jsonString string, endpoint string) error {
	if strings.Contains(jsonString, endpoint) {
		zap.S().Errorw("Webhook URL %s already subscribed!", endpoint)
		return errors.New("Webhook URL already subscribed")
	}

	return nil
}

// AppendToSubscribers traverses levels of an attribute
// TODO move this to a better file location on reorg
func AppendToSubscribers(jsonString string, urls []interface{}, endpoint string) (*string, error) {
	urls = append(urls, endpoint)

	var result string

	for _, url := range urls {
		urlIndex := strings.Index(jsonString, url.(string))
		if urlIndex != -1 {
			quotation := urlIndex + strings.Index(jsonString[urlIndex:], `"`)
			result = jsonString[0:quotation] + `","` + endpoint + `"` + jsonString[quotation+1:]
		}
	}

	return &result, nil
}
