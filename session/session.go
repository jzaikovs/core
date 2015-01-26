package session

import (
	"crypto"
	_ "crypto/sha1"
	"encoding/base64"
	. "github.com/jzaikovs/t"
	"sync"
	"time"
)

var (
	SESSION_ID             = "sid"
	HASH       crypto.Hash = crypto.SHA1
)

type Server interface {
	CookieValue(key string) (string, bool)
	SetCookieValue(key, value string)
	UserAgent() string
	RemoteAddr() string
}

var sessions = make(map[string]*Session) // todo: consider map[int] as it is rumored to be faster
var sessionslock = sync.RWMutex{}

func sessions_get(sid string) (session *Session, ok bool) {
	sessionslock.RLock()
	session, ok = sessions[sid]
	sessionslock.RUnlock()
	return
}

func (self *Session) save() {
	sessionslock.Lock()
	sessions[self.sid] = self
	sessionslock.Unlock()
}

type Session struct {
	sid        string
	authorized bool
	server     Server
	Data       Map
}

func New(server Server) *Session {
	sid, ok := server.CookieValue(SESSION_ID) // session identifier is store in cookie

	if ok {
		if session, ok := sessions_get(sid); ok {
			session.server = server
			return session
		}
	}

	self := new(Session)
	self.server = server
	self.Data = make(Map)
	self.CreateCookie(time.Now().String())
	return self
}

func (self *Session) ID() string {
	return self.sid
}

func (self *Session) IsAuth() bool {
	return self.authorized
}

func (self *Session) Destroy() {
	sessionslock.Lock()
	delete(sessions, self.sid)
	sessionslock.Unlock()
}

func (self *Session) Authorize(salt string) {
	self.Destroy()
	self.CreateCookie(salt)
	self.authorized = true
}

func (self *Session) CreateCookie(salt string) {
	sid := salt
	sid += self.server.UserAgent()
	sid += self.server.RemoteAddr()

	h := HASH.New()
	h.Write([]byte(sid))

	self.sid = base64.URLEncoding.EncodeToString(h.Sum(nil))
	self.server.SetCookieValue(SESSION_ID, self.sid)

	self.save()
}

// Strip removes unused references to memory,
// fo example 10k users session is stored in
// memory but we don't need to hold Server instance, it can be huge
func (self *Session) Unlink() {
	//this is needed for Go GC to do it's job
	self.server = nil
}
