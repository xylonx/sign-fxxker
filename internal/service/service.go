package service

import (
	"fmt"

	"github.com/xylonx/sign-fxxker/internal/config"
	"github.com/xylonx/sign-fxxker/internal/course"
	"github.com/xylonx/sign-fxxker/internal/user"
)

func StartAutoSignIn() (err error) {
	alias2user := make(map[string]user.User)
	for _, u := range config.Config.Users.Chaoxing {
		alias2user[u.Alias] = user.NewChaoXingUser(u.Cookie, u.Alias)
	}
	alias2baiduLocation := make(map[string]course.BaiduLocation)
	for _, l := range config.Config.Location {
		alias2baiduLocation[l.Alias] = course.BaiduLocation{
			Addr:      l.BaiduMapAddrName,
			Longitude: l.BaiduMapLongitude,
			Latitude:  l.BaiduMapLatitude,
		}
	}

	chaoxingCourse := []*course.ChaoXingCourse{}
	for _, c := range config.Config.Course.Chaoxing {
		if _, ok := alias2baiduLocation[c.Location]; !ok {
			return fmt.Errorf("location %s is not in location list. please add user first\n", c.Location)
		}
		for _, u := range c.Users {
			// check user validation
			if _, ok := alias2user[u]; !ok {
				return fmt.Errorf("user %s is not in user list. please add user first\n", u)
			}
			chaoxingCourse = append(chaoxingCourse, course.NewChaoXingCourse(&course.ChaoXingCourseOptions{
				CourseName:      c.Alias,
				User:            alias2user[u],
				Location:        alias2baiduLocation[c.Location],
				IntervalSeconds: config.Config.Course.IntervalSeconds,
				DelaySeconds:    config.Config.Course.DelaySeconds,
				CourseID:        c.CourseId,
				ClassID:         c.ClassId,
			}))
		}
	}

	// auto sign chaoxing course
	for _, c := range chaoxingCourse {
		status := c.StartAutoSign()
		go func(<-chan string) {
			for s := range status {
				fmt.Println(s)
				// TODO: reporter by reporter sub-module
			}
		}(status)
	}

	return nil
}
