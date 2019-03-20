package session

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

// Store wraps a gorilla session store
type Store struct {
	store sessions.Store
	name  string
}

// NewCookieStore sets up the named session as an encrypted cookie session.
// This is "secure enough" for most uses, but shouldn't be considered a safe
// place to store sensitive (e.g., passwords / FERPA) data.
func NewCookieStore(name, secret string) *Store {
	return &Store{
		store: sessions.NewCookieStore([]byte(secret)),
		// We prefix with "gopkg-" as a sort of namespace mechanism so that we
		// aren't likely to collide with other gorilla session data
		name: "gopkg-" + name,
	}
}

// Session pulls the session data from the given request/response pair
func (st *Store) Session(w http.ResponseWriter, req *http.Request) (*Session, error) {
	var s, err = st.store.Get(req, st.name)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve session: %s", err)
	}
	return &Session{s, w, req}, nil
}
