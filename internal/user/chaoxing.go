package user

import (
	"net/http"
	"regexp"
	"sync"
)

var uidRegrex *regexp.Regexp

func init() {
	uidRegrex, _ = regexp.Compile(`_uid=(\d+);`)
}

type ChaoXingUser struct {
	cookie   string
	username string
	uid      string
	mux      sync.RWMutex
}

var _ User = &ChaoXingUser{}

func NewChaoXingUser(cookie, name string) User {
	uid := uidRegrex.FindString(cookie)
	go autoRenewCookie()
	return &ChaoXingUser{cookie: cookie, username: name, uid: uid}
}

func (u *ChaoXingUser) SetCredentials(req *http.Request) {
	u.mux.RLock()
	cookie := u.cookie
	u.mux.RUnlock()
	req.Header.Set("Cookie", cookie)
}

func (u *ChaoXingUser) GetUID() string {
	return u.uid
}

func autoRenewCookie() {}
