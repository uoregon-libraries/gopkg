package middleware

import (
	"net/http"
	"time"

	"github.com/uoregon-libraries/gopkg/logger"
	"github.com/uoregon-libraries/gopkg/statusrecorder"
)

// Middleware holds configuration which may be necessary for more complex
// middleware functions, while ensuring its defaults (via middleware.NewApache,
// for instance) have pretty decent starting values for various situations.
type Middleware struct {
	// UserHeader is the header used to identify basic auth users, and defaults
	// to "X-Remote-User" for Apache.  When blank, GetRemoteUser will always
	// return the empty string.
	UserHeader string

	// IPHeader is the header which gives the real IP address behind a proxy;
	// Apache uses "X-Forwarded-For".  If this is blank, only the request's
	// RemoteAddr value is returned by GetIP.
	IPHeader string

	// Logger is necessary for any of the request logging.  It defaults to a
	// simple logger.Logger instance that writes to stderr.
	Logger *logger.Logger
}

// New returns a default Middleware structure suitable for use when an
// application is not behind any proxy
func New() *Middleware {
	return &Middleware{Logger: logger.New(logger.Debug)}
}

// NewApache returns a Middleware with values set up for Go running behind an
// Apache proxy
func NewApache() *Middleware {
	var m = New()
	m.IPHeader = "X-Forwarded-For"
	return m
}

// NewApacheBasicAuth returns a Middleware with values set up for Go running
// behind an Apache proxy which also is configured to use basic auth
func NewApacheBasicAuth() *Middleware {
	var m = NewApache()
	m.UserHeader = "X-Remote-User"
	return m
}

// NoCache is a simple Middleware function to send back the no-cache header
func (m *Middleware) NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "max-age=0, must-revalidate")
		next.ServeHTTP(w, req)
	})
}

// Log centralizes the log format and what data we log for any request
// as well as capturing the status and duration of the "real" request.
// Generally you'll want this to be the final wrapper around any other
// middleware to get valid timing.
func (m *Middleware) Log(w http.ResponseWriter, req *http.Request, next http.Handler, logfn func(format string, args ...interface{}), prefix string) {
	var sr = statusrecorder.New(w)
	var start = time.Now()
	next.ServeHTTP(sr, req)
	var ms = time.Since(start).Seconds() * 1000
	logfn("%s: [%s] %s - %d (%0.3fms)", prefix, m.ClientIdentity(req), req.URL, sr.Status, ms)
}

// RequestLog uses the logger to write an info-level log for a page request
func (m *Middleware) RequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		m.Log(w, req, next, m.Logger.Infof, "Request")
	})
}

// RequestStaticAssetLog uses the logger to write a debug-level log for static
// asset requests.  If this is used in conjunction with RequestLog, logs can
// more easily be filtered to avoid spam when unimportant requests occur.
func (m *Middleware) RequestStaticAssetLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		m.Log(w, req, next, m.Logger.Debugf, "Asset Request")
	})
}
