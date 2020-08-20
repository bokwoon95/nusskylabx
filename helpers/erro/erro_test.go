package erro

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestWrapAndDump(t *testing.T) {
	is := is.New(t)
	err := Wrap(Wrap(Wrap(fmt.Errorf("gg=G"))))
	buf := &bytes.Buffer{}
	Dump(buf, err)
	is.True(buf.Len() != 0)
	is.True(Sdump(err) != "")
	is.True(!strings.Contains(S1dump(err), "\n"))
}

func TestIs(t *testing.T) {
	is := is.New(t)

	E1 := errors.New("E1")
	E2 := errors.New("E2")
	E3 := errors.New("E3")

	e3 := E3
	e4 := errors.New("E4")

	is.True(Is(e3, E1, E2, E3))
	is.True(!Is(e4, E1, E2, E3))
}
