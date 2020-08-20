package dbutil

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/bokwoon95/nusskylabx/helpers/random"
	"github.com/lib/pq"

	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestError(t *testing.T) {
	is := is.New(t)
	pqerr := &pq.Error{
		Code:   PqInvalidTextRepresentation,
		Detail: random.Sentence(10),
		Hint:   random.Sentence(10),
	}
	err := NewDBError(pqerr, "SELECT * FROM table WHERE arg1 = $1 AND arg2 = $2", 5, 10)
	dberr, ok := AsDBError(err)
	is.True(ok)
	is.Equal(dberr.Code, pqerr.Code)
	is.Equal(dberr.Detail, pqerr.Detail)
	is.Equal(dberr.Hint, pqerr.Hint)
	is.Equal(dberr.Query, "SELECT * FROM table WHERE arg1 = 5 AND arg2 = 10;")
	is.True(dberr.Error() != "")

	dberr2, ok := AsDBError(pqerr)
	is.True(ok)
	is.Equal(dberr2.Code, pqerr.Code)
	is.Equal(dberr2.Detail, pqerr.Detail)
	is.Equal(dberr2.Hint, pqerr.Hint)
	is.Equal(dberr2.Query, "")
	is.True(dberr2.Error() != "")
}

func TestInterpolateSql(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	now := time.Now()
	str := uuid.New().String()
	i := rand.Int()
	i32 := rand.Int31()
	i64 := rand.Int63()
	type Person struct {
		Name  string
		Age   int
		Hobby string
	}
	person := Person{"bob", 32, "cryptography"}
	b, err := json.Marshal(person)
	if err != nil {
		t.FailNow()
	}
	personJSON := string(b)

	tests := []struct {
		name  string
		query string
		args  []interface{}
		want  string
	}{
		{
			"strings and numbers",
			"SELECT * FROM table WHERE string = $1, int64 = $2, int32 = $3, int = $4;",
			[]interface{}{str, i64, i32, i},
			fmt.Sprintf("SELECT * FROM table WHERE string = '%s', int64 = %d, int32 = %d, int = %d;", str, i64, i32, i),
		},
		{
			"time and NULL",
			"SELECT $1, $2",
			[]interface{}{now, nil},
			fmt.Sprintf("SELECT '%s', NULL;", now.Format(time.RFC3339Nano)),
		},
		{
			"driver.Valuer Null and Not Null",
			"SELECT $1, $2",
			[]interface{}{
				sql.NullString{Valid: true, String: str},
				sql.NullInt64{Valid: false, Int64: i64},
			},
			fmt.Sprintf("SELECT '%s', NULL;", str),
		},
		{
			"arbitrary struct",
			"INSERT INTO table (data) VALUES ($1)",
			[]interface{}{person},
			fmt.Sprintf("INSERT INTO table (data) VALUES ('%s');", personJSON),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)
			got := InterpolateSql(tt.query, tt.args...)
			is.Equal(tt.want, got)
		})
	}
}

func TestRebindPlaceholders(t *testing.T) {
	type args struct {
		query string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"basic",
			args{"SELECT 1 FROM table WHERE arg = ? AND funarg = ?"},
			"SELECT 1 FROM table WHERE arg = $1 AND funarg = $2",
		},
		{
			"escaped ??",
			args{"SELECT 1 FROM table WHERE 'arg??' = ? AND '????funarg' = ?"},
			"SELECT 1 FROM table WHERE 'arg?' = $1 AND '??funarg' = $2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RebindPlaceholders(tt.args.query); got != tt.want {
				t.Errorf("RebindPlaceholders() = %v, want %v", got, tt.want)
			}
		})
	}
}
