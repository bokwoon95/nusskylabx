// Package logutil provides logging utilities
package logutil

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strings"

	"github.com/bokwoon95/nusskylabx/helpers/dbutil"
	"github.com/go-chi/chi/middleware"
)

const (
	prefixTrace = "[TRACE] "
	prefixDebug = "[DEBUG] "
	flagTrace   = 0
	flagDebug   = log.LstdFlags | log.Llongfile | log.Lmsgprefix
)

// A Logger represents an active logging object that generates lines of
// output to an io.Writer. Each logging operation makes a single call to
// the Writer's Write method. A Logger can be used simultaneously from
// multiple goroutines; it guarantees to serialize access to the Writer.
type Logger struct {
	trace *log.Logger
	debug *log.Logger
}

// NewLogger returns a new Logger.
func NewLogger(w io.Writer) *Logger {
	l := &Logger{
		trace: log.New(ioutil.Discard, prefixTrace, flagTrace),
		debug: log.New(ioutil.Discard, prefixDebug, flagDebug),
	}
	if w != nil {
		l.trace.SetOutput(w)
		l.debug.SetOutput(w)
	}
	return l
}

// GetReqID piggybacks chi/middleware's GetReqID by obtaining the request ID of
// the request (if present) and strips away the machine name part of the
// string.
//
// NOTE: The Request ID is initially set by the chi middleware.RequestID
// middleware function, this function merely picks up on it. If
// middleware.RequestID was not called prior to this, GetReqID will simply
// return an empty string
func GetReqID(ctx context.Context) string {
	reqID := middleware.GetReqID(ctx)
	// Remove everything before the first slash in reqID
	slashIndex := strings.Index(reqID, "/")
	if slashIndex >= 0 && slashIndex+1 < len(reqID) {
		reqID = reqID[slashIndex+1:]
	}
	return reqID
}

func (l *Logger) Output(calldepth int, s string) error {
	if l.debug.Writer() == ioutil.Discard {
		return nil
	}
	return l.debug.Output(2+calldepth, s)
}

// Printf is a wrapper around log.Logger.Printf. It will print nothing if the
// output is set to ioutil.Discard, allowing for levelled logging.
// https://stackoverflow.com/a/42762815
func (l *Logger) Printf(format string, v ...interface{}) {
	if l.debug.Writer() == ioutil.Discard {
		return
	}
	_ = l.debug.Output(2, fmt.Sprintf(format, v...))
}

// Println is a wrapper around log.Logger.Println. It will print nothing if the
// output is set to ioutil.Discard, allowing for levelled logging.
func (l *Logger) Println(v ...interface{}) {
	if l.debug.Writer() == ioutil.Discard {
		return
	}
	_ = l.debug.Output(2, fmt.Sprintln(v...))
}

// Writer returns the output destination of the Logger.
func (l *Logger) Writer() io.Writer {
	return l.debug.Writer()
}

// SetOutput sets the output destination for the logger.
func (l *Logger) SetOutput(w io.Writer) {
	l.trace.SetOutput(w)
	l.debug.SetOutput(w)
}

// StartRequest should be called in the middleware at the very start of the
// request, so that it cleanly delineates between inidividual requests.
func (l *Logger) StartRequest(r *http.Request) {
	if l.trace.Writer() == ioutil.Discard {
		return
	}
	reqID := GetReqID(r.Context())
	_ = l.trace.Output(2,
		fmt.Sprintf(
			"----------------------------------------%s %s RequestID:%s----------------------------------------\n",
			r.Method, r.URL, reqID,
		),
	)
}

// TraceRequest should be called at the start of every handler function that
// accepts a *http.Request. It will print the function, file and line number of
// where the TraceRequest call was called, allowing the user to trace the chain
// functions being called in a request.
//
// It will also print out the any request ID found in the request context.
func (l *Logger) TraceRequest(r *http.Request) {
	if l.trace.Writer() == ioutil.Discard {
		return
	}
	reqID := GetReqID(r.Context())
	pc, filename, linenr, _ := runtime.Caller(1)
	_ = l.trace.Output(2, fmt.Sprintf("RequestID:%s file:line[%s:%d] function[%s]", reqID, filename, linenr, runtime.FuncForPC(pc).Name()))
}

// TraceFunc is like TraceRequest, except it is for functions that don't take
// in a *http.Request. This means it will not print the request ID, only the
// function, file and line number of where TraceFunc was called.
func (l *Logger) TraceFunc() {
	if l.trace.Writer() == ioutil.Discard {
		return
	}
	pc, filename, linenr, _ := runtime.Caller(1)
	_ = l.trace.Output(2, fmt.Sprintf("file:line[%s:%d] function[%s]", filename, linenr, runtime.FuncForPC(pc).Name()))
}

// RequestPrintf is like Printf, but it will also print out the request ID of
// the current request.
func (l *Logger) RequestPrintf(r *http.Request, format string, v ...interface{}) {
	if l.debug.Writer() == ioutil.Discard {
		return
	}
	reqID := GetReqID(r.Context())
	_ = l.debug.Output(2, fmt.Sprintf("RequestID:"+reqID+" "+format, v...))
}

// SqlPrintf takes in the same arguments as sql.DB.Query/ sql.DB.QueryRow/
// sql.DB.Exec, and logs the exact sql statement that was carried out including
// the arguments. Useful for debugging SQL queries, as it may show some
// arguments having unexpected values in the logger output.
func (l *Logger) SqlPrintf(query string, v ...interface{}) {
	if l.debug.Writer() == ioutil.Discard {
		return
	}
	l.debug.SetFlags(0)
	defer l.debug.SetFlags(flagDebug)
	pc, filename, linenr, _ := runtime.Caller(1)
	_ = l.debug.Output(2, fmt.Sprintf("file:line[%s:%d] function[%s]: "+dbutil.InterpolateSql(query, v...), filename, linenr, runtime.FuncForPC(pc).Name()))
}
