package core

import (
	. "github.com/jzaikovs/t"
	"sync"
	"time"
)

var sessions = make(map[string]*Session) // todo: consider map[int] as it is rumored to be faster
var sessionslock = sync.RWMutex{}

func sessions_get(sid string) (session *Session, ok bool) {
	sessionslock.RLock()
	session, ok = sessions[sid]
	sessionslock.RUnlock()
	return
}

func (this *Session) save() {
	sessionslock.Lock()
	sessions[this.sid] = this
	sessionslock.Unlock()
}

type Session struct {
	sid        string
	authorized bool
	input      Input
	output     Output
	Data       Map
}

func session(input Input, output Output) *Session {
	sid, ok := input.CookieValue("sid") // session identifier is store in cookie
	if !ok {
		sid = input.HeaderValue("api_sid") // or it can be passed as header value from API client
		ok = len(sid) > 0
	}

	if ok {
		if session, ok := sessions_get(sid); ok {
			session.input = input
			session.output = output
			session.output.AddHeader("api_sid", sid)
			return session
		}
	}

	session := new(Session)
	session.input = input
	session.output = output
	session.Data = make(Map)
	session.CreateCookie(time.Now().String())
	return session
}

func (this *Session) ID() string {
	return this.sid
}

func (this *Session) IsAuth() bool {
	return this.authorized
}

func (this *Session) Destroy() {
	sessionslock.Lock()
	delete(sessions, this.sid)
	sessionslock.Unlock()
}

func (this *Session) Authorize(salt string) {
	this.Destroy()
	this.CreateCookie(salt)
	this.authorized = true
}

func (this *Session) CreateCookie(salt string) {
	this.sid = salt
	this.sid += this.input.UserAgent()
	this.sid += this.input.RemoteAddr()
	this.sid = Base64Encode(SHA1(this.sid))

	this.output.SetCookieValue("sid", this.sid)
	this.output.AddHeader("api_sid", this.sid)

	this.save()
}

// Strip removes unused references to memory,
// fo example 10k users session is stored in
// memory but we don't need to store input and output modules
func (this *Session) strip() {
	//this is needed for Go GC to do it's job
	this.input = nil
	this.output = nil
}
