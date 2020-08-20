package erro

import (
	"errors"
	"fmt"

	"github.com/lib/pq"
)

type BaseError string

func (e BaseError) Error() string {
	return string(e)
}

type Error struct {
	BaseError BaseError
	Args      []interface{}
}

func (e *Error) Error() string {
	return fmt.Sprintf(e.BaseError.Error(), e.Args...)
}

func (e *Error) Unwrap() error {
	return e.BaseError
}

func Errorf(be BaseError, a ...interface{}) error {
	return &Error{BaseError: be, Args: a}
}

func AsError(err error) (*Error, bool) {
	var skylaberr *Error
	ok := errors.As(err, &skylaberr)
	return skylaberr, ok
}

// The first five characters of the error string are assumed to be a postgres error code
func (e BaseError) PqCode() pq.ErrorCode {
	errstr := string(e)
	if len(errstr) < 5 {
		return pq.ErrorCode(errstr)
	}
	return pq.ErrorCode(errstr[:5])
}
