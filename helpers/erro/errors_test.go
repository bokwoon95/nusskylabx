package erro

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lib/pq"
	"github.com/matryer/is"
)

func TestBaseErrorCode(t *testing.T) {
	is := is.New(t)
	tests := []struct {
		name     string
		internal string
		want     pq.ErrorCode
	}{
		{"basic", "ABCDE lorem ipsum", pq.ErrorCode("ABCDE")},
		{"empty string", "", pq.ErrorCode("")},
		{"incomplete code", "12", pq.ErrorCode("12")},
	}
	for _, tt := range tests {
		e := BaseError(tt.internal)
		is.Equal(tt.want, e.PqCode())
	}
}

func TestError(t *testing.T) {
	is := is.New(t)
	const ErrUserIdNotExist BaseError = "The user id %d does not exist"
	is.True(ErrUserIdNotExist.Error() == "The user id %d does not exist")

	err := Errorf(ErrUserIdNotExist, 1337)
	is.True(err.Error() == "The user id 1337 does not exist")
	is.True(err != ErrUserIdNotExist)          // The annotated error and the base error are not directly equal
	is.True(errors.Is(err, ErrUserIdNotExist)) // The annotated error can still be identified by its base error with errors.Is

	err2 := fmt.Errorf("wrap twice: %w", fmt.Errorf("wrap once: %w", err)) // Wrap Error into -> error
	err3, ok := AsError(err2)                                              // Unwrap error back to -> *Error
	is.True(ok)
	is.True(errors.Is(err3, ErrUserIdNotExist)) // *Error is still identifiable by its base error
}
