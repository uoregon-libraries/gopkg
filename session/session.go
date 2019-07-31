package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// Session wraps the gorilla session to provide a better API and to ensure an
// empty session doesn't panic.  The session API exposed here only uses
// strings, not any kind of complex structures.  As such, so long as nobody
// mucks about with the session data outside these APIs, there's no risk of
// pulling a non-string data type and panicking when converting to string.
type Session struct {
	s   *sessions.Session
	w   http.ResponseWriter
	req *http.Request
}

// GetString returns the string value for the given key.  If the value doesn't
// exist in our session data, an empty string is returned.
func (s *Session) GetString(key string) string {
	var val = s.s.Values[key]
	if val == nil {
		return ""
	}
	var sval = val.(string)
	return sval
}

// SetString stores the key/value string pair to the session
func (s *Session) SetString(key, val string) error {
	s.s.Values[key] = val
	return s.s.Save(s.req, s.w)
}

// getFlash grabs the given key as a "flash" value from the session store.  If
// the key doesn't exist, or its value isn't a string, an empty string is
// returned.
func (s *Session) getFlash(key string) string {
	var data = s.s.Flashes(key)
	if len(data) > 0 {
		if str, ok := data[0].(string); ok {
			s.s.Save(s.req, s.w)
			return str
		}
	}

	return ""
}

// setFlash stores the key/value pair in the session store as a "flash" value
func (s *Session) setFlash(key, val string) {
	s.s.AddFlash(val, key)
	s.s.Save(s.req, s.w)
}

// GetInfoFlash returns the one-time "flash" value for any "info" data set up
func (s *Session) GetInfoFlash() string {
	return s.getFlash("info")
}

// SetInfoFlash stores val as the one-time "flash" value for "info" data
func (s *Session) SetInfoFlash(val string) {
	s.setFlash("info", val)
}

// GetAlertFlash returns the one-time "flash" value for any "alert" data set up
func (s *Session) GetAlertFlash() string {
	return s.getFlash("alert")
}

// SetAlertFlash stores val as the one-time "flash" value for "alert" data
func (s *Session) SetAlertFlash(val string) {
	s.setFlash("alert", val)
}
