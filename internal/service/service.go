package service

import (
	"fmt"

	"github.com/xylonx/sign-fxxker/internal/config"
	"github.com/xylonx/sign-fxxker/internal/course"
	"github.com/xylonx/sign-fxxker/internal/user"
)

func StartAutoSignIn() (err error) {
	alias2user := make(map[string]*user.User)
	for _, u := range config.Config.Users {
		alias2user[u.Alias] = user.NewUser(u.Cookie, u.Alias)
	}

	alias2baiduLocation := make(map[string]course.BaiduLocation)
	for _, l := range config.Config.Location {
		alias2baiduLocation[l.Alias] = course.BaiduLocation{
			Addr:      l.BaiduMapAddrName,
			Longitude: l.BaiduMapLongitude,
			Latitude:  l.BaiduMapLatitude,
		}
	}

	chaoxingCourse := make(map[string]*course.ChaoXingCourse, len(config.Config.Course.Chaoxing))
	for _, c := range config.Config.Course.Chaoxing {
		courseUsers := make([]*user.User, len(c.Users))
		for i := range courseUsers {
			courseUsers[i] = alias2user[c.Users[i]]
		}
		chaoxingCourse[c.Alias], err = course.NewChaoXingCourse(&course.ChaoXingCourseOptions{
			Users:           courseUsers,
			Location:        alias2baiduLocation[c.Location],
			IntervalSeconds: config.Config.Course.IntervalSeconds,
			DelaySeconds:    config.Config.Course.DelaySeconds,
			CourseID:        c.CourseId,
			ClassID:         c.ClassId,
		})
		if err != nil {
			return err
		}
	}

	for _, c := range chaoxingCourse {
		status := c.StartAutoSign()
		go func(<-chan course.SignStatus) {
			for s := range status {
				fmt.Printf("Dear %s:\n课程 %s 签到状态: %v\n附加信息: %s", s.User.GetUserName(), s.CourseName, s.Success, s.Message)
			}
		}(status)
	}

	return nil
}
