package formutil

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bokwoon95/nusskylabx/helpers/erro"
)

const multipartMaxSize = 32 << 20

const (
	ErrValueMissing erro.BaseError = "No form value found for key '%s'"
	ErrValueNotInt  erro.BaseError = "The form value '%s' has value '%s' which is not an int"
)

func ParseForm(r *http.Request) error {
	var err error
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		err = r.ParseMultipartForm(multipartMaxSize)
	} else {
		err = r.ParseForm()
	}
	return err
}

func String(r *http.Request, key string) (value string, err error) {
	if r.Form == nil {
		err = ParseForm(r)
		if err != nil {
			return value, err
		}
	}
	if values := r.Form[key]; len(values) > 0 {
		return values[0], nil
	}
	return "", erro.Errorf(ErrValueMissing, key)
}

func Strings(r *http.Request, key string) (values []string, err error) {
	if r.Form == nil {
		err = ParseForm(r)
		if err != nil {
			return values, err
		}
	}
	if values, ok := r.Form[key]; ok {
		return values, nil
	}
	return values, erro.Errorf(ErrValueMissing, key)
}

func Int(r *http.Request, key string) (value int, err error) {
	if r.Form == nil {
		err = ParseForm(r)
		if err != nil {
			return value, err
		}
	}
	if strs := r.Form[key]; len(strs) > 0 {
		value, err = strconv.Atoi(strs[0])
		if err != nil {
			return value, erro.Errorf(ErrValueNotInt, strs[0])
		}
		return value, nil
	}
	return value, erro.Errorf(ErrValueMissing, key)
}

func Ints(r *http.Request, key string) (values []int, err error) {
	if r.Form == nil {
		err = ParseForm(r)
		if err != nil {
			return values, err
		}
	}
	if strs, ok := r.Form[key]; ok {
		values = make([]int, len(strs))
		for i := range strs {
			values[i], err = strconv.Atoi(strs[i])
			if err != nil {
				return values, erro.Errorf(ErrValueNotInt, strs[i])
			}
		}
		return values, nil
	}
	return values, erro.Errorf(ErrValueMissing, key)
}
