package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gbaranski/houseflow/pkg/types"
	"golang.org/x/crypto/bcrypt"
)

// ParseSignedPayload parses payload, returns respectively message and signature
func ParseSignedPayload(p []byte) (string, []byte, error) {
	splitted := strings.SplitN(string(p), ".", 2)
	if len(splitted) < 1 {
		return "", nil, fmt.Errorf("payload is invalid, it should contain payload and signature")
	}
	signature := splitted[0]
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return "", nil, fmt.Errorf("failed parsing signature %s", err.Error())
	}
	payload := splitted[1]

	return payload, decoded, nil
}

// GenerateRandomString generates random string and returns it
func GenerateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

// HashPassword hashes password with bcrypt
func HashPassword(pass []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
}

// ComparePasswordAndHash checks hash of password, returns true if they match, otherwise false
func ComparePasswordAndHash(password string, hash []byte) bool {
	return bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil
}

// MustGetEnv retrieves the environment variable
//
// Panics if doesn't exists
func MustGetEnv(key string) string {
	env, present := os.LookupEnv(key)
	if !present {
		panic(fmt.Errorf("%s environment variable is unset", key))
	}
	return env
}

// MustParseEnvKey parses base64 encoded public key, useful when you load it from ENV
//
// Panics when ENV does not exists, or length is invalid
func MustParseEnvKey(env string, size int) []byte {
	key, err := base64.StdEncoding.DecodeString(MustGetEnv(env))
	if err != nil {
		panic(fmt.Errorf("fail decode %s key from env", env))
	}
	if len(key) < size {
		panic(fmt.Errorf("invalid length of %s", env))
	}
	return key
}

// ReturnError takes http.ResponseWriter and writes err to it
func ReturnError(w http.ResponseWriter, err types.ResponseError) {
	str, _ := json.Marshal(err)
	w.WriteHeader(err.StatusCode)
	w.Write(str)
}

// NewRequestID creates new random RequestID and returns it
func NewRequestID(len int) ([]byte, error) {
	id := make([]byte, len)
	n, err := rand.Read(id)
	if err != nil {
		return nil, err
	}
	if n != len {
		return nil, fmt.Errorf("invalid written bytes")
	}

	return id, nil
}
