package templateutil

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// Sql contains Template Functions for working with "database/sql" types
func Sql(funcs map[string]interface{}) map[string]interface{} {
	if funcs == nil {
		funcs = map[string]interface{}{}
	}
	funcs["Sql_NullValue"] = NullValue
	return funcs
}

// NullValue returns either the underlying value of the sql Null type, or
// "<NULL>" if the value is null. If the underlying value is an empty string, a
// pair of escaped double quotes \"\" will be returned to signify an empty
// string.
func NullValue(x interface{}) string {
	switch v := x.(type) {
	case string:
		if v == "" {
			return "\"\""
		}
		return v
	case int:
		return strconv.Itoa(v)
	case sql.NullString:
		if v.Valid {
			if v.String == "" {
				return "\"\""
			}
			return v.String
		}
	case sql.NullInt64:
		if v.Valid {
			return strconv.FormatInt(v.Int64, 10)
		}
	case sql.NullTime:
		if v.Valid {
			singapore, err := time.LoadLocation("Asia/Singapore")
			if err != nil {
				panic(err)
			}
			t := v.Time.In(singapore)
			return t.String()
		}
	// Rare cases
	case int64:
		return strconv.FormatInt(v, 10)
	case sql.NullFloat64:
		if v.Valid {
			return fmt.Sprintf("%f", v.Float64)
		}
	case sql.NullInt32:
		if v.Valid {
			return strconv.FormatInt(int64(v.Int32), 10)
		}
	case sql.NullBool:
		if v.Valid {
			if v.Bool {
				return "True"
			}
			return "False"
		}
	}
	return "𝗡𝗨𝗟𝗟"
}
