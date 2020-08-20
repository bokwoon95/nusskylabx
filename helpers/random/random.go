package random

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"math/rand"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit"
)

func init() {
	gofakeit.Seed(time.Now().UnixNano())
	rand.Seed(time.Now().UnixNano())
}

func Name() string {
	return gofakeit.Name()
}

func Email() string {
	return gofakeit.Email()
}

func NameAndEmail() (name, email string) {
	name = gofakeit.Name()
	email = gofakeit.Email()
	joinedName := strings.ReplaceAll(name, " ", "")
	emailDomain := email
	i := strings.Index(email, "@")
	if i > 0 {
		emailDomain = email[i:]
	}
	return name, joinedName + emailDomain
}

func TeamName() string {
	return strings.Title(gofakeit.BuzzWord() + " " + gofakeit.HackerIngverb() + " " + gofakeit.City())
}

func URL() string {
	path := gofakeit.URL()
	path = strings.TrimPrefix(path, "http:/")
	path = strings.TrimPrefix(path, "https:/")
	return path
}

func Sentence(wordCount int) string {
	return gofakeit.Sentence(wordCount)
}

func Word() string {
	return gofakeit.Word()
}

func Bool() bool {
	return gofakeit.Bool()
}

func Int() int {
	return rand.Int()
}

func Float64() float64 {
	return rand.Float64()
}

// Cryptographically random secret key
func SecretKey() string {
	arr := make([]byte, 32)
	_, err := cryptorand.Read(arr) // crypto/rand is more secure than math/rand
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(arr)
}
