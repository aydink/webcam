package main

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"sync"
	"time"
)

type SessionStore struct {
	sync.Mutex
	sessions map[string]*Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]*Session)}
}

func (ss *SessionStore) Set(sid string, session *Session) {
	ss.Lock()
	defer ss.Unlock()
	ss.sessions[sid] = session
}

func (ss *SessionStore) Get(sid string) *Session {
	ss.Lock()
	defer ss.Unlock()

	if v, ok := ss.sessions[sid]; ok {
		v.timeAccessed = time.Now()
		return v
	}

	return &Session{isNew: true}
}

func (ss *SessionStore) Delete(sid string) {
	delete(ss.sessions, sid)
}

func (ss *SessionStore) Reset() {
	ss.Lock()
	defer ss.Unlock()
	ss.sessions = make(map[string]*Session)
}

func (ss *SessionStore) GC() {
	ss.Lock()
	defer ss.Unlock()

	for sid, session := range ss.sessions {
		if (session.timeAccessed.Unix() + session.maxLifetime) < time.Now().Unix() {
			delete(ss.sessions, sid)
		}

	}
}

type Session struct {
	sync.Mutex
	sid          string                      // unique session id
	timeCreated  time.Time                   // last access time
	timeAccessed time.Time                   // last access time
	value        map[interface{}]interface{} // session value stored inside
	isNew        bool
	maxLifetime  int64
}

func (s *Session) Update() {
	s.Lock()
	defer s.Unlock()

	s.timeAccessed = time.Now()
}

func NewSession(value map[interface{}]interface{}) *Session {
	s := &Session{}
	s.sid = CreateSessionId()
	s.timeCreated = time.Now()
	s.value = value
	s.isNew = true
	return s
}

func CreateSessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
