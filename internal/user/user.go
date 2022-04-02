package user

import (
	"regexp"
	"sync"
)

var uidRegrex *regexp.Regexp

func init() {
	uidRegrex, _ = regexp.Compile(`_uid=(\d+);`)
}

type User struct {
	cookie   string
	userName string
	uid      string
	mux      *sync.RWMutex
}

func NewUser(cookie, username string) *User {
	uid := uidRegrex.FindString(cookie)
	return &User{cookie: cookie, userName: username, uid: uid, mux: &sync.RWMutex{}}
}

func (u *User) GetCookie() string {
	u.mux.RLock()
	c := u.cookie
	u.mux.RUnlock()
	return c
}

func (u *User) GetUID() string {
	return u.uid
}

func (u *User) GetUserName() string {
	return u.userName
}

// TODO: refresh cookie
func (u *User) refreshCookie() error {
	return nil
}
