// Package auth provides various authentication related utilities
package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"net/http"
	"strings"

	"github.com/bokwoon95/nusskylabx/helpers/headers"

	"github.com/bokwoon95/nusskylabx/helpers/erro"

	"github.com/bokwoon95/nusskylabx/helpers/auth/oauth"
	"github.com/bokwoon95/nusskylabx/helpers/auth/openid"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/bcrypt"
)

const (
	ErrDeserializeOutputInvalid erro.BaseError = "deserialize output is invalid because provided signature [%s] does not match computed signature [%s]"
	ErrDeserializeInputInvalid  erro.BaseError = "deserialize input [%s] is invalid because missing '.' in string"
)

func GenerateRandomString() (string, error) {
	arr := make([]byte, 32)
	_, err := rand.Read(arr)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(arr), nil
}

// Hash will hash input byte slice and output a string. Takes in a key to salt
// the hashing algorithm
//
// THIS IS NOT SAFE FOR PASSWORD HASHING. Use HashPassword and its companion
// function CompareHashAndPassword instead
func Hash(key string, input []byte) (output string) {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Reset()
	_, _ = hash.Write(input)
	b := hash.Sum(nil)
	output = base64.URLEncoding.EncodeToString(b)
	return output
}

// Serialize will serialize any object into a string so that it can be sent
// down the wire. It can be converted back later using Deserialize. It requires
// a key to digitally sign the output
func Serialize(key string, input interface{}) (output string, err error) {
	buf := &bytes.Buffer{}
	err = gob.NewEncoder(buf).Encode(input)
	if err != nil {
		return output, fmt.Errorf("gob encoder error: %w", err)
	}
	payload := buf.Bytes()
	encodedPayload := base64.URLEncoding.EncodeToString(payload)
	signature := Hash(key, payload)
	output = encodedPayload + "." + signature // URLEncoding will never have a '.', making it safe to use as a delimiter
	return output, nil
}

// Deserialize will deserialize any input string back into an object that was
// serialized with Serialize. It requires the same key that was used to
// serialize the variable in order to verify the digital signature of the
// payload. If the digital signature doesn't match, it will return
// ErrDeserializeOutputInvalid
func Deserialize(key string, input string, outputAddr interface{}) (err error) {
	strs := strings.SplitN(input, ".", 2)
	if len(strs) < 2 {
		return erro.Errorf(ErrDeserializeInputInvalid, input)
	}
	encodedPayload := strs[0]
	providedSignature := strs[1]
	payload, err := base64.URLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return fmt.Errorf("error with base64 URL decoding string: %w", err)
	}
	computedSignature := Hash(key, payload)
	if providedSignature != computedSignature {
		return erro.Errorf(ErrDeserializeOutputInvalid, providedSignature, computedSignature)
	}
	buf := bytes.NewBuffer(payload)
	err = gob.NewDecoder(buf).Decode(outputAddr)
	if err != nil {
		return fmt.Errorf("gob decoder error: %w", err)
	}
	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CompareHashAndPassword(passwordhash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordhash), []byte(password))
}

func Redirect(
	w http.ResponseWriter, r *http.Request,
	provider, returnTo string,
	errorHandler func(http.ResponseWriter, *http.Request, error),
) {
	headers.DoNotCache(w)
	switch {
	case oauth.IsValidProvider(provider):
		oauth.Redirect(provider, returnTo, errorHandler)(w, r)
	case openid.IsValidProvider(provider):
		openid.Redirect(provider, returnTo, errorHandler)(w, r)
	case provider == "":
		errorHandler(w, r, fmt.Errorf("provider cannot be blank"))
	default:
		errorHandler(w, r, fmt.Errorf("Invalid provider: %s", provider))
	}
}

func Authenticate(returnTo string, errorHandler func(http.ResponseWriter, *http.Request, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.FormValue("state") != "": // it's from oauth
				oauth.Authenticate(returnTo, errorHandler)(next).ServeHTTP(w, r)
			case r.FormValue("openid.sreg.email") != "": // it's from openid
				provider := openid.ProviderNUS // hardcoded to NUS
				openid.Authenticate(provider, errorHandler)(next).ServeHTTP(w, r)
			default:
				errorHandler(w, r, fmt.Errorf("%s can't tell who this callback is from: %+v", returnTo, r))
			}
		})
	}
}

func IsValidProvider(provider string) bool {
	return openid.IsValidProvider(provider) || oauth.IsValidProvider(provider)
}

func AddConstProvider(funcs map[string]interface{}) map[string]interface{} {
	if funcs == nil {
		funcs = map[string]interface{}{}
	}
	funcs = openid.AddConstProvider(funcs)
	funcs = oauth.AddConstProvider(funcs)
	return funcs
}
