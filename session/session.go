package session

import (
	"crypto"
	_ "crypto/sha1" // default hash package
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/jzaikovs/t"
)

// SessionCookieName stores name of cookie used to link request with session data
var SessionCookieName = "sid"

// DefaultHash is crypto hash used to makce session ID
var DefaultHash = crypto.SHA1

// Server is interface/trates that should be implemented to use this package
type Server interface {
	CookieValue(name string) (string, bool)
	SetCookieValue(name, value string)
	UserAgent() string
	RemoteAddr() string
}

var sessions = make(map[string]*Session) // todo: consider map[int] as it is rumored to be faster
var sessionslock = sync.RWMutex{}

func getSession(sid string) (session *Session, ok bool) {
	sessionslock.RLock()
	session, ok = sessions[sid]
	sessionslock.RUnlock()
	return
}

func (session *Session) save() {
	sessionslock.Lock()
	sessions[session.sid] = session
	sessionslock.Unlock()
}

// Session represents single session from one user across requests
type Session struct {
	sid        string
	authorized bool
	server     Server
	Data       t.Map
}

// Validate validates request for session storage
// returns session ID and true/false for if session found or not
func Validate(req *http.Request) (string, bool) {
	cookie, err := req.Cookie(SessionCookieName)
	if err != nil {
		return "", false
	}

	if sesssion, ok := getSession(cookie.Value); ok {
		return sesssion.ID(), sesssion.Valid("", req.UserAgent(), req.RemoteAddr)
	}
	return "", false
}

// New creates new session, using structure which implements Server interface
func New(server Server) *Session {
	sid, ok := server.CookieValue(SessionCookieName) // session identifier is store in cookie

	if ok {
		if session, ok := getSession(sid); ok {
			session.server = server
			return session
		}
	}

	session := new(Session)
	session.server = server
	session.Data = make(t.Map)
	session.CreateCookie(time.Now().String())
	return session
}

// ID returns session ID which is equeal to cookie session id
func (session *Session) ID() string {
	return session.sid
}

// IsAuth returns true if session is authorized
func (session *Session) IsAuth() bool {
	return session.authorized
}

// Destroy destroys session data from session storage
func (session *Session) Destroy() {
	sessionslock.Lock()
	delete(sessions, session.sid)
	sessionslock.Unlock()
}

// Authorize marks session as authorized, you can pass sepecific value as "salt"/key.
// salt/key value should be used in validate call
// for new sessions ID generation, session id cookie is recreated with different ID
// to prevent session fixation attacks
func (session *Session) Authorize(salt string) {
	session.Destroy()          // destroy old session
	session.CreateCookie(salt) // create new session with new ID
	session.authorized = true
}

// CreateCookie will create cookie using salt/key
func (session *Session) CreateCookie(salt string) {
	sid := salt
	sid += session.server.UserAgent()
	sid += session.server.RemoteAddr()

	h := DefaultHash.New()
	h.Write([]byte(sid))

	session.sid = base64.URLEncoding.EncodeToString(h.Sum(nil))
	session.server.SetCookieValue(SessionCookieName, session.sid)

	session.save()
}

// Valid validates session
func (session *Session) Valid(salt, userAgent, remoteAddr string) bool {
	h := DefaultHash.New()
	h.Write([]byte(salt))
	h.Write([]byte(userAgent))
	h.Write([]byte(remoteAddr))

	return session.ID() == base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// Unlink removes unused references to memory,
// fo example 10k users session is stored in
// memory but we don't need to hold Server instance, it can be huge
func (session *Session) Unlink() {
	//this is needed for Go GC to do it's job
	session.server = nil
}
