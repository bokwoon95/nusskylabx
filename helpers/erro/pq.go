package erro

import (
	"errors"

	"github.com/lib/pq"
)

const (
	// Class 22 - Data Exception
	PqDataException             pq.ErrorClass = "22"
	PqInvalidTextRepresentation pq.ErrorCode  = "22P02"
)

const (
	// Class 23 — Integrity Constraint Violation
	PqIntegrityConstraintViolation pq.ErrorClass = "23"
	PqNotNullViolation             pq.ErrorCode  = "23502"
	PqForeignKeyViolation          pq.ErrorCode  = "23503"
	PqUniqueViolation              pq.ErrorCode  = "23505"
	PqCheckViolation               pq.ErrorCode  = "23514"
)

func AsPqError(err error) (*pq.Error, bool) {
	var pqerr *pq.Error
	ok := errors.As(err, &pqerr)
	return pqerr, ok
}
