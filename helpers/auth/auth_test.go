package auth

import (
	"testing"

	"github.com/bokwoon95/nusskylabx/helpers/random"
	"github.com/matryer/is"
)

func TestGenerateRandomString(t *testing.T) {
	is := is.New(t)
	str, err := GenerateRandomString()
	is.NoErr(err)
	is.True(str != "")
}

func TestHash(t *testing.T) {
	is := is.New(t)
	key, err := GenerateRandomString()
	is.NoErr(err)
	payload := []byte(random.Sentence(10))
	hashed1 := Hash(key, payload)
	hashed2 := Hash(key, payload)
	is.Equal(hashed1, hashed2)
}

func TestSerializeDeserialize(t *testing.T) {
	is := is.New(t)
	type Random struct {
		String string
		Int    int
		Bool   bool
	}
	key, err := GenerateRandomString()
	is.NoErr(err)
	data1 := Random{random.Word(), random.Int(), random.Bool()}
	str, err := Serialize(key, data1)
	is.NoErr(err)
	var data2 Random
	err = Deserialize(key, str, &data2)
	is.NoErr(err)
	is.Equal(data2, data1)
}

func TestHashPassword(t *testing.T) {
	is := is.New(t)
	password := random.Word()
	passwordHash, err := HashPassword(password)
	is.NoErr(err)
	err = CompareHashAndPassword(passwordHash, password)
	is.NoErr(err)
	err = CompareHashAndPassword(passwordHash, password+"abcd")
	is.True(err != nil)
}
