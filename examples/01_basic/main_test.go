package main

import (
	"testing"

	"github.com/bokwoon95/nusskylabx/helpers/testutil"
)

func TestHelloWorld(t *testing.T) {
	w, r := testutil.NewGet("", nil)
	HelloWorld(w, r)
	testutil.ResponseOK(w)
}

func TestRandomNumber(t *testing.T) {
	w, r := testutil.NewGet("", nil)
	RandomNumber(w, r)
	testutil.ResponseOK(w)
}
