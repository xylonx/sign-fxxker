package course

import "github.com/xylonx/sign-fxxker/internal/user"

type SignStatus struct {
	Success    bool
	CourseName string
	User       *user.User
	Message    string
}
