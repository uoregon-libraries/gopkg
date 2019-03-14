package middleware

import "net/http"

// GetIP returns a string that contains one or more IP addresses which
// represent the client and values from any proxies which set the IPHeader.
//
// Note that the only reliable value here is going to be the last entry, as
// users can spoof any header that isn't explicitly changed by a proxy.  e.g.,
// if you set IPHeader to "X-Forwarded-For", but run this behind a proxy that
// doesn't use this header, you can get falsified data.
func (m *Middleware) GetIP(req *http.Request) string {
	var addr = req.RemoteAddr
	var xff = req.Header.Get(m.IPHeader)
	if xff != "" {
		addr += "," + req.RemoteAddr
	}
	return addr
}

// GetRemoteUser returns the value of the configured UserHeader.  This is only
// reliable if you have the correct header set *and* you're using an
// authentication method that actually sets said header.
//
// If the UserHeader isn't set, this returns an empty string.  If it is set,
// but there is no value for the user header, "N/A" is returned.
func (m *Middleware) GetRemoteUser(req *http.Request) string {
	if m.UserHeader == "" {
		return ""
	}

	var u = req.Header.Get(m.UserHeader)
	if u == "" {
		return "N/A"
	}
	return u
}

// ClientIdentity returns a string that can be used to determine the identity
// of a user making a request.  We use the remote user and ip address to
// generate a string in the form "{{user}} - {{ip}}" if UserHeader has a value
// (e.g., using apache basic auth), or simply ip data from GetIP.
func (m *Middleware) ClientIdentity(req *http.Request) string {
	var ip = m.GetIP(req)
	if m.UserHeader == "" {
		return ip
	}

	return m.GetRemoteUser(req) + "-" + ip
}
