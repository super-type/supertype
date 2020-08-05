// TODO do we want this as a util within authenticating? Or elsewhere? I think in here is good to keep contexts boudned
// 	this is purely businsess logic regardless of database or anything...

package authenticating

import (
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
)

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
