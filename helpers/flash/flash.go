// Package flash provides flash message utilities
package flash

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"time"

	"github.com/bokwoon95/nusskylabx/helpers/cookies"
	"github.com/bokwoon95/nusskylabx/helpers/erro"
)

// FlashMsg represents a flash message
type FlashMsg struct {
	Valid bool   `json:"Valid"`
	Value string `json:"Value"`
}

const (
	flashmsgCookiename = "_flash_message"
)

type Context string

const (
	ContextFlashMsgs Context = "ContextFlashMsgs" // map[string][]string
)

const (
	Success = "success"
	Error   = "error"
	Warning = "warning"
)

func Funcs(funcs template.FuncMap, w http.ResponseWriter, r *http.Request, key string) template.FuncMap {
	if funcs == nil {
		funcs = template.FuncMap{}
	}
	funcs["FlashMsgSuccess"] = func() string { return Success }
	funcs["FlashMsgError"] = func() string { return Error }
	funcs["FlashMsgWarning"] = func() string { return Warning }
	fe := NewEncoder(key)
	flashmsgsMap, err := fe.GetFlashMsgs(w, r)
	funcs["FlashutilGetFlashMsg"] = func(name string) (FlashMsg, error) {
		err := err
		flashmsgs := flashmsgsMap[name]
		if len(flashmsgs) > 0 {
			return flashmsgs[0], err
		}
		return FlashMsg{}, err
	}
	funcs["FlashutilGetFlashMsgs"] = func(name string) ([]FlashMsg, error) { err := err; return flashmsgsMap[name], err }
	return funcs
}

type Encoder struct {
	Key string
}

func NewEncoder(key string) Encoder {
	return Encoder{Key: key}
}

// convert an incoming map[string][]string into a map[string][]FlashMsg, then
// append it to the existing flashmsgsMap
func appendFlashMsgs(flashmsgsMap map[string][]FlashMsg, msgs map[string][]string) map[string][]FlashMsg {
	if flashmsgsMap == nil {
		flashmsgsMap = make(map[string][]FlashMsg)
	}
	for k, msgs := range msgs {
		for _, msg := range msgs {
			flashmsgsMap[k] = append(flashmsgsMap[k], FlashMsg{Valid: true, Value: msg})
		}
	}
	return flashmsgsMap
}

// SetFlashMsgs will set a bunch of flash messages into a cookie that can be
// read later by GetFlashMsgs.
//
// NOTE: This returns a new request with the flash messages in the context
// updated, so make sure to re-assign the request:
//		r, _ = flash.NewEncoder("secret_key").SetFlashMsgs(w, r, msgs)
func (fe Encoder) SetFlashMsgs(w http.ResponseWriter, r *http.Request, msgs map[string][]string) (*http.Request, error) {
	flashmsgsMap, _ := r.Context().Value(ContextFlashMsgs).(map[string][]FlashMsg)
	flashmsgsMap = appendFlashMsgs(flashmsgsMap, msgs)
	err := cookies.NewEncoder(fe.Key).EncodeVariableInCookie(w, flashmsgCookiename, flashmsgsMap, cookies.Duration(time.Second*60), cookies.AllowJS(true))
	if err != nil {
		return r, erro.Wrap(err)
	}
	r = r.WithContext(context.WithValue(r.Context(), ContextFlashMsgs, flashmsgsMap))
	return r, nil
}

// GetFlashMsgs will get all flash messages from both the cookie and request
// context, after which it will delete the cookie (hence a 'flash' message).
//
// Make sure you use an Encoder with the same seceret key as the encoder that you set the flash messages with:
//	flashmsgs, _ := flash.NewEncoder("secret_key").GetFlashMsgs(w, r)
//
// Usually you won't call this function directly in your handlers, instead you
// will include the "helpers/flash/flash.html" template in your html file which
// will automatically call this function and display the flash messages, if any.
func (fe Encoder) GetFlashMsgs(w http.ResponseWriter, r *http.Request) (map[string][]FlashMsg, error) {
	var flashmsgsMap, ctxFlashMsgs map[string][]FlashMsg
	flashmsgsMap = make(map[string][]FlashMsg)
	// It is unlikely that we will be getting flash messages from both cookie
	// and context. Even if we write a flash message to both cookie and
	// context, if GetFlashMsgs is called:
	//	1) in the same request --> only flash messages in context will be
	//	detected. A cookie cannot be set and read in the same request.
	//	2) in a separate request --> only flash messages in the cookie will be
	//	detected. The context will have expired at the end of the last request.

	// Get flash messages from cookie
	err := cookies.NewEncoder(fe.Key).DecodeVariableFromCookie(r, flashmsgCookiename, &flashmsgsMap)
	if err != nil {
		switch {
		case errors.Is(err, cookies.ErrNoCookie):
			// no big deal if we find no cookie, we will check the context next
		default:
			return flashmsgsMap, erro.Wrap(err)
		}
	}
	cookies.DeleteCookie(w, flashmsgCookiename)

	// Get flash messages from context
	ctxFlashMsgs, _ = r.Context().Value(ContextFlashMsgs).(map[string][]FlashMsg)
	for k, v := range ctxFlashMsgs {
		flashmsgsMap[k] = v
	}

	return flashmsgsMap, nil
}

// UnsetFlashMsgs will unset all flash messages under the specified keys. This
// is useful if you wish to suppress particular flash messages that you know
// were set in a previous middleware handler.
//
// NOTE: This returns a new request with the flash messages in the context
// updated, so make sure to re-assign the request:
//		r, _ = flash.NewEncoder("secret_key").UnsetFlashMsgs(flash.Success)
func (fe Encoder) UnsetFlashMsgs(w http.ResponseWriter, r *http.Request, keys ...string) (*http.Request, error) {
	flashmsgsMap, _ := r.Context().Value(ContextFlashMsgs).(map[string][]FlashMsg)
	for _, key := range keys {
		delete(flashmsgsMap, key)
	}
	err := cookies.NewEncoder(fe.Key).EncodeVariableInCookie(w, flashmsgCookiename, flashmsgsMap, cookies.Duration(time.Second*60), cookies.AllowJS(true))
	if err != nil {
		return r, erro.Wrap(err)
	}
	r = r.WithContext(context.WithValue(r.Context(), ContextFlashMsgs, flashmsgsMap))
	return r, nil
}

// UnsetFlashMsgsHandler is a http.Handler wrapper around UnsetFlashMsgs.
func (fe Encoder) UnsetFlashMsgsHandler(keys ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r, _ = fe.UnsetFlashMsgs(w, r, keys...)
			next.ServeHTTP(w, r)
		})
	}
}
