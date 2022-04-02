package core

import (
	"fmt"

	"github.com/xylonx/sign-fxxker/internal/config"
	"github.com/xylonx/sign-fxxker/internal/course"
	"github.com/xylonx/sign-fxxker/internal/user"
)

var (
	Alias2user          map[string]user.User
	Alias2baiduLocation map[string]course.BaiduLocation

	ChaoxingCourse []*course.ChaoXingCourse
)

func Setup() error {
	Alias2user = make(map[string]user.User)
	for _, u := range config.Config.Users.Chaoxing {
		Alias2user[u.Alias] = user.NewChaoXingUser(u.Cookie, u.Alias)
	}
	Alias2baiduLocation = make(map[string]course.BaiduLocation)
	for _, l := range config.Config.Location {
		Alias2baiduLocation[l.Alias] = course.BaiduLocation{
			Addr:      l.BaiduMapAddrName,
			Longitude: l.BaiduMapLongitude,
			Latitude:  l.BaiduMapLatitude,
		}
	}

	ChaoxingCourse = []*course.ChaoXingCourse{}
	for _, c := range config.Config.Course.Chaoxing {
		if _, ok := Alias2baiduLocation[c.Location]; !ok {
			return fmt.Errorf("location %s is not in location list. please add user first\n", c.Location)
		}
		for _, u := range c.Users {
			// check user validation
			if _, ok := Alias2user[u]; !ok {
				return fmt.Errorf("user %s is not in user list. please add user first\n", u)
			}
			ChaoxingCourse = append(ChaoxingCourse, course.NewChaoXingCourse(&course.ChaoXingCourseOptions{
				CourseName:      c.Alias,
				User:            Alias2user[u],
				Location:        Alias2baiduLocation[c.Location],
				IntervalSeconds: config.Config.Course.IntervalSeconds,
				DelaySeconds:    config.Config.Course.DelaySeconds,
				CourseID:        c.CourseId,
				ClassID:         c.ClassId,
			}))
		}
	}
	return nil
}
