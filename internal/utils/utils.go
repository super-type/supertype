package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/dgrijalva/jwt-go"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/storage"
)

// SetupAWSSession starts an AWS session
func SetupAWSSession() *dynamodb.DynamoDB {
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

// GenerateSupertypeID generates a new Supertype ID for a given password
func GenerateSupertypeID(password string) (*string, error) {
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

// ValidateEmail checks to see whether a valid email was entered
func ValidateEmail(email string) bool {
	var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if len(email) < 3 && len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

// IsAuthorized checks the given JWT to ensure vendor is authenticated
func IsAuthorized(endpoint func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				err := godotenv.Load()
				if err != nil {
					return nil, authenticating.ErrNotAuthorized
				}

				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, authenticating.ErrNotAuthorized
				}

				return []byte(os.Getenv("JWT_SIGNING_KEY")), nil
			})
			if err != nil {
				return
			}

			if token.Valid {
				endpoint(w, r)
			}
		}
	})
}

// GetAPIKeyHash returns the hashed value of the secret key
func GetAPIKeyHash(skVendor string) string {
	h := sha256.New()
	h.Write([]byte(skVendor))
	return hex.EncodeToString(h.Sum(nil))
}

// TraverseLevels traverses levels of an attribute
func TraverseLevels(jsonString string, destination []string, levels interface{}, endpoint string) (*string, error) {
	// Check to make sure URL isn't already contained in subscribers list
	if strings.Contains(jsonString, endpoint) {
		// TODO some kind of telling message to vendor
		fmt.Println("Webhook URL already subscribed")
		return nil, errors.New("Webhook URL already subscribed")
	}

	// Register the URL granularly
	for i := 0; i < len(destination); i++ {
		if levels.(map[string]interface{})[destination[i]] == nil {
			continue
		}
		levels = levels.(map[string]interface{})[destination[i]]
	}

	levels = levels.(map[string]interface{})["subscribers"]
	urls := levels.([]interface{})
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
