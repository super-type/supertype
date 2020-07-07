package rest

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/super-type/supertype/internal/keys"
	"github.com/super-type/supertype/internal/reencryption"

	"github.com/gorilla/mux"
	"github.com/super-type/supertype/pkg/authenticating"
)

// Router is the main router for the application
func Router(a authenticating.Service) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/healthcheck", healthcheck()).Methods("GET")
	router.HandleFunc("/loginVendor", loginVendor(a)).Methods("POST")
	router.HandleFunc("/createVendor", createVendor(a)).Methods("POST")

	return router
}

type Capsule struct {
	E *ecdsa.PublicKey
	V *ecdsa.PublicKey
	S *big.Int
}

func healthcheck() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO move all of this to create vendor... or at least some of it...
		// TODO we'll have to rewrite our encryption package

		ska, pka, err := keys.GenerateKeys()
		if err != nil {
			fmt.Printf("Error generating keys: %v\n", err)
		}

		skb, pkb, err := keys.GenerateKeys()
		if err != nil {
			fmt.Printf("Error generating keys: %v\n", err)
		}

		fmt.Printf("Private Key A to string: %v\n", hex.EncodeToString(ska.D.Bytes()))
		fmt.Printf("Private Key B to string: %v\n", hex.EncodeToString(skb.D.Bytes()))

		m := "Test"
		fmt.Printf("Original message: %v\n", m)

		cipherText, capsule, err := reencryption.Encrypt(m, pka)
		if err != nil {
			fmt.Println(err)
		}

		capsuleAsBytes, err := reencryption.EncodeCapsule(*capsule)
		if err != nil {
			fmt.Printf("Encode error: %v\n:", err)
		}

		capsuleTest, err := reencryption.DecodeCapsule(capsuleAsBytes)
		if err != nil {
			fmt.Printf("Decode error: %v\n", err)
		}

		fmt.Println("capsule before encode:", capsule)
		fmt.Println("capsule after decode:", capsuleTest)
		fmt.Println("ciphereText:", cipherText)

		rekey, px, err := reencryption.ReKeyGen(ska, pkb)
		if err != nil {
			fmt.Printf("Error generating re-encryption key: %v\n", err)
		}

		fmt.Printf("rekey: %v\n", rekey)

		fmt.Printf("rekey string: %v\n", rekey.String())

		newCapsule, err := reencryption.ReEncryption(rekey, capsule)
		if err != nil {
			fmt.Printf("Error re-encrypting: %v\n", err)
		}

		plaintext, err := reencryption.Decrypt(skb, newCapsule, px, cipherText)
		if err != nil {
			fmt.Printf("Error decrypting: %v\n", err)
		}

		fmt.Printf("Plaintext: %v\n", string(plaintext))

		fmt.Println("Healthy")
	}
}

// loginVendor returns a handler for POST /loginVendor requests
func loginVendor(a authenticating.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get new vendor details from client
		decoder := json.NewDecoder(r.Body)

		// ? should should this be authenticating, or storage?
		var vendor authenticating.Vendor
		err := decoder.Decode(&vendor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		a.LoginVendor(vendor)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("Logged in vendor.")
	}
}

func createVendor(a authenticating.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get new vendor details from client
		decoder := json.NewDecoder(r.Body)

		// ? should should this be authenticating, or storage?
		var vendor authenticating.Vendor
		err := decoder.Decode(&vendor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := a.CreateVendor(vendor)
		if err != nil {
			fmt.Printf("Error creating vendor\n")
		}

		// TODO log user in after creating account
		// TODO create re-encryption keys between this user and others

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result) // TODO return the JWT from login here...
	}
}
