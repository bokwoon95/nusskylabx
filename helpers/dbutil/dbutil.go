// Package dbutil provides database related utilities
package dbutil

import (
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

const (
	// Class 22 - Data Exception
	PqDataException             pq.ErrorClass = "22000"
	PqInvalidTextRepresentation pq.ErrorCode  = "22P02"

	// Class 23 — Integrity Constraint Violation
	ErrIntegrityConstraintViolation pq.ErrorClass = "23000"
	ErrNotNullViolation             pq.ErrorCode  = "23502"
	ErrForeignKeyViolation          pq.ErrorCode  = "23503"
	ErrUniqueViolation              pq.ErrorCode  = "23505"
	ErrCheckViolation               pq.ErrorCode  = "23514"
)

type DBError struct {
	Code   pq.ErrorCode
	Query  string
	Err    error
	Detail string
	Hint   string
}

func (e *DBError) Error() string {
	b := &strings.Builder{}
	b.WriteString(e.Err.Error())
	if e.Code != "" {
		b.WriteString(" [" + string(e.Code) + " " + e.Code.Name() + "] ")
	}
	if e.Detail != "" {
		b.WriteString(" " + e.Detail)
	}
	if e.Hint != "" {
		b.WriteString(", " + e.Hint)
	}
	return b.String()
}

func (e *DBError) Unwrap() error {
	return e.Err
}

func AsDBError(err error) (*DBError, bool) {
	dberr := &DBError{}
	ok := errors.As(err, &dberr)
	if !ok {
		var pqerr *pq.Error
		if errors.As(err, &pqerr) {
			dberr.Code = pqerr.Code
			dberr.Detail = pqerr.Detail
			dberr.Hint = pqerr.Hint
			dberr.Err = pqerr
			ok = true
		}
	}
	return dberr, ok
}

// NewDBError decorates an error returned from the database with the relevant
// query that was used (in (*DBError).Query so that handler functions can print
// out which SQL query was responsible for causing the error. If the underlying
// error is a (*pq.Error), it also copies the Code, Detail and Hint fields from
func NewDBError(err error, query string, args ...interface{}) *DBError {
	var pqerr *pq.Error
	dberr := &DBError{}
	if errors.As(err, &pqerr) {
		dberr.Code = pqerr.Code
		dberr.Detail = pqerr.Detail
		dberr.Hint = pqerr.Hint
	}
	dberr.Query = InterpolateSql(query, args...)
	dberr.Err = err
	return dberr
}

func InterpolateSql(query string, args ...interface{}) string {
	query = regexp.MustCompile(`(?m)--.*$`).ReplaceAllString(query, " ") // Remove comments
	query = regexp.MustCompile(`\\n|\\t`).ReplaceAllString(query, " ")   // Remove newlines/tabs
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")       // Replace multiple spaces with one space
	query = strings.TrimSpace(query)
	for i, arg := range args {
		var val string
		switch v := arg.(type) {
		case string:
			val = fmt.Sprintf("'%s'", v)
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			val = fmt.Sprint(arg)
		case []byte:
			// https://stackoverflow.com/a/34437656
			val = fmt.Sprintf("E'\\x%s'", strings.ToUpper(hex.EncodeToString(v)))
		case time.Time:
			val = fmt.Sprintf("'%s'", v.Format(time.RFC3339Nano))
		case nil:
			val = "NULL"
		case driver.Valuer:
			Interface, err := v.Value()
			if err != nil {
				val = "(" + err.Error() + ")"
			} else {
				switch Concrete := Interface.(type) {
				case string:
					val = fmt.Sprintf("'%s'", Concrete)
				case nil:
					val = "NULL"
				default:
					val = "'" + fmt.Sprint(arg) + "'"
				}
			}
		default:
			b, err := json.Marshal(arg)
			if err != nil {
				val = "(" + err.Error() + ")"
			} else {
				val = "'" + string(b) + "'"
			}
		}
		query = strings.ReplaceAll(query, "$"+strconv.Itoa(i+1), val)
	}
	if !strings.HasSuffix(query, ";") {
		query = query + ";"
	}
	return query
}

// Replaces all MySQL-style '?,?,...' placeholders with Postgres-style
// '$1,$2,...' placeholders.
func RebindPlaceholders(query string) string {
	buf := &bytes.Buffer{}
	i := 0
	for {
		p := strings.Index(query, "?")
		if p < 0 {
			break
		}
		if len(query[p:]) > 1 && query[p:p+2] == "??" { // Unescape ?? -> ?
			buf.WriteString(query[:p])
			buf.WriteString("?")
			query = query[p+2:]
		} else { // Replace ? -> $<number>
			i++
			buf.WriteString(query[:p])
			buf.WriteString("$" + strconv.Itoa(i))
			query = query[p+1:]
		}
	}
	buf.WriteString(query)
	return buf.String()
}
