package erro

import (
	"testing"

	"github.com/bokwoon95/nusskylabx/helpers/random"
	"github.com/lib/pq"
	"github.com/matryer/is"
)

func TestAsPqError(t *testing.T) {
	is := is.New(t)
	var err error = &pq.Error{
		Code:   PqInvalidTextRepresentation,
		Detail: random.Sentence(10),
		Hint:   random.Sentence(10),
	}
	pqerr, ok := AsPqError(err)
	is.True(ok)
	is.Equal(pqerr, err)
}
