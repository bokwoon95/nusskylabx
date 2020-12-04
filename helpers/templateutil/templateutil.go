// Package templateutil provides various utility functions for golang templates
package templateutil

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Funcs contains miscellaneous Template Functions
func Funcs(funcs map[string]interface{}) map[string]interface{} {
	if funcs == nil {
		funcs = map[string]interface{}{}
	}
	funcs["_JSONify"] = JSONify
	funcs["_Sha1Hash"] = Sha1Hash
	funcs["_InputDate"] = InputDate
	funcs["_InputTime"] = InputTime
	return funcs
}

// JSONify is a Template Function that converts a generic object into a JSON
// string.  If the object cannot be marshalled into a json string, it returns
// "null" instead. To ensure that your object can be converted into json, make
// sure it has json struct tags on its fields
func JSONify(input interface{}) string {
	output, err := json.Marshal(input)
	if err != nil {
		return "null"
	}
	return string(output)
}

// Sha1Hash is a Template Function that hashes a bunch of strings together
// using the SHA1 hash (the same that Git uses) and returns the first 8
// characters from the hash
func Sha1Hash(inputs ...interface{}) (output string) {
	var strs []string
	for _, input := range inputs {
		strs = append(strs, fmt.Sprintf("%#v", input))
	}
	if len(strs) == 0 {
		return ""
	}
	joined := strings.Join(strs, ",")
	sha1Bytes := sha1.Sum([]byte(joined))
	output = hex.EncodeToString(sha1Bytes[:])
	return output[0:8]
}

func InputDate(t sql.NullTime) (datestring string, err error) {
	singapore, err := time.LoadLocation("Asia/Singapore")
	if err != nil {
		return datestring, err
	}
	if t.Valid {
		datestring = t.Time.In(singapore).Format("2006-01-02")
	}
	return datestring, nil
}

func InputTime(t sql.NullTime) (timestring string, err error) {
	singapore, err := time.LoadLocation("Asia/Singapore")
	if err != nil {
		return timestring, err
	}
	if t.Valid {
		timestring = t.Time.In(singapore).Format("15:04")
	}
	return timestring, nil
}
