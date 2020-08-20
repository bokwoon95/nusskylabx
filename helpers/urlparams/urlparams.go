package urlparams

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/bokwoon95/nusskylabx/helpers/cookies"
	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/go-chi/chi"
)

const (
	ErrChiPanic     erro.BaseError = "chi.URLParam panic https://github.com/go-chi/chi/issues/76"
	ErrParamMissing erro.BaseError = "The URL parameter '%s' is missing"
	ErrParamNotInt  erro.BaseError = "The URL parameter '%s' has value '%s' which is not an int"
)

func SetString(r *http.Request, key, value string) *http.Request {
	routeCtx, _ := r.Context().Value(chi.RouteCtxKey).(*chi.Context)
	routeCtx.URLParams.Add(key, value)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
	return r
}

func SetInt(r *http.Request, key string, value int) *http.Request {
	str := strconv.Itoa(value)
	r = SetString(r, key, str)
	return r
}

func String(r *http.Request, key string) (value string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrChiPanic
		}
	}()
	str := chi.URLParam(r, key)
	if str == "" {
		return str, erro.Wrap(erro.Errorf(ErrParamMissing, key))
	}
	return str, nil
}

func Int(r *http.Request, key string) (value int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrChiPanic
		}
	}()
	str := chi.URLParam(r, key)
	if str == "" {
		return value, erro.Wrap(erro.Errorf(ErrParamMissing, key))
	}
	value, err = strconv.Atoi(str)
	if err != nil {
		return value, erro.Wrap(erro.Errorf(ErrParamNotInt, key, str))
	}
	return value, nil
}

func PersistentString(w http.ResponseWriter, r *http.Request, key, cookiename string) (value string, err error) {
	paramvalue, err := String(r, key)
	if err != nil {
		switch {
		case errors.Is(err, ErrParamMissing):
			// No problem, check the cookies next
		default:
			return value, erro.Wrap(err)
		}
	}
	cookieValue := cookies.GetCookieValue(r, cookiename)
	value = paramvalue
	if value == "" {
		value = cookieValue
	}
	cookies.SetCookie(w, cookiename, value)
	return value, nil
}
