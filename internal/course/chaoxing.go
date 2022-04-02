package course

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/xylonx/sign-fxxker/internal/constant"
	"github.com/xylonx/sign-fxxker/internal/user"
	"github.com/xylonx/zapx"
	"go.uber.org/zap"
)

const (
	GeneralSign  = "签到"
	LocationSign = "位置签到"
	QrCodeSign   = "二维码签到"
)

type ActiveList struct {
	Result int `json:"result"`
	Data   struct {
		ActiveList []struct {
			SignName  string `json:"nameOne"`
			AttendNum int    `json:"attendNum"`
			StartTime int64  `json:"startTime"`
			EndTime   int64  `json:"endTime"`
			ActiveID  int64  `json:"id"`
		}
	} `json:"data"`
}

type BaiduLocation struct {
	Addr      string
	Longitude float64
	Latitude  float64
}

type ChaoXingCourse struct {
	courseName string
	users      []*user.User
	location   BaiduLocation
	interval   time.Duration
	delay      time.Duration

	courseID int64
	classID  int64

	mux            *sync.RWMutex
	qrCodeEnc      string
	qrCodeUpdateAt time.Time
}

type ChaoXingCourseOptions struct {
	CourseName      string
	Users           []*user.User
	Location        BaiduLocation
	IntervalSeconds int64
	DelaySeconds    int64
	CourseID        int64
	ClassID         int64
}

func NewChaoXingCourse(opt *ChaoXingCourseOptions) (*ChaoXingCourse, error) {
	if opt == nil {
		return nil, errors.New("no option specified")
	}
	if len(opt.Users) == 0 {
		return nil, errors.New("no users specified")
	}
	return &ChaoXingCourse{
		courseName: opt.CourseName,
		users:      opt.Users,
		location:   opt.Location,
		interval:   time.Duration(opt.IntervalSeconds) * time.Second,
		delay:      time.Duration(opt.DelaySeconds) * time.Second,
		courseID:   opt.CourseID,
		classID:    opt.ClassID,
		mux:        &sync.RWMutex{},
	}, nil
}

func (c *ChaoXingCourse) UpdateQrCodeEnc(updateAt time.Time, enc string) {
	c.mux.Lock()
	c.qrCodeEnc = enc
	c.qrCodeUpdateAt = updateAt
	c.mux.Unlock()
}

func (c *ChaoXingCourse) getQrCode() (string, time.Time) {
	c.mux.RLock()
	enc := c.qrCodeEnc
	updatedAt := c.qrCodeUpdateAt
	c.mux.RUnlock()
	return enc, updatedAt
}

func (c *ChaoXingCourse) StartAutoSign() <-chan SignStatus {
	ch := make(chan SignStatus, len(c.users))
	go func() {
		for {
			alist, err := getActiveList(context.Background(), c.users[len(c.users)-1].GetCookie(), c.courseID, c.classID)
			if err != nil && len(alist.Data.ActiveList) == 0 {
				time.Sleep(c.interval)
				continue
			}
			// TODO(sign): now just sign the latest one. It can be used for all
			sign := alist.Data.ActiveList[0]
			if time.Now().UnixMilli() < sign.EndTime && sign.AttendNum == 0 {
				// !delay to avoid sign too quickly
				time.Sleep(c.delay)
				// Sign for all users
				for _, u := range c.users {
					var err error
					switch sign.SignName {
					case GeneralSign:
						_, err = generalSign(context.Background(), u.GetCookie(), sign.ActiveID)
					case LocationSign:
						_, err = locationSign(context.Background(), u.GetCookie(), u.GetUID(), c.location.Addr, sign.ActiveID, c.location.Longitude, c.location.Latitude)
					case QrCodeSign:
						enc, updatedAt := c.getQrCode()
						if updatedAt.UnixMilli() < sign.StartTime {
							break
						}
						_, err = qrCodeSign(context.Background(), u.GetCookie(), enc, sign.ActiveID)
					default:
						err = errors.New("sign type is not supported: " + sign.SignName)
					}

					if err == nil {
						ch <- SignStatus{Success: true, CourseName: c.courseName, Message: "success", User: u}
					} else {
						ch <- SignStatus{Success: false, CourseName: c.courseName, Message: err.Error(), User: u}
					}
				}
			}

			time.Sleep(c.interval)
		}
	}()
	return ch
}

// for each single course, the activeList for every students is same. Therefore, use the random ones.
func getActiveList(ctx context.Context, cookie string, courseID, classID int64) (*ActiveList, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, fmt.Sprintf(`https://mobilelearn.chaoxing.com/v2/apis/active/student/activelist?fid=0&courseId=%d&classId=%d&_=%d`, courseID, classID, time.Now().UnixMilli()),
		nil,
	)
	if err != nil {
		zapx.Error("generate get activeList request failed", zap.Error(err))
		return nil, err
	}
	req.Header.Add("Cookie", cookie)
	req.Header.Add("User-Agent", constant.UA)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zapx.Error("get activeList failed", zap.Error(err))
		return nil, err
	}

	alist := new(ActiveList)
	if err := json.NewDecoder(resp.Body).Decode(alist); err != nil {
		zapx.Error("decode activeList failed", zap.Error(err))
		return nil, err
	}

	return alist, nil
}

func generalSign(ctx context.Context, cookie string, activeID int64) (time.Time, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, fmt.Sprintf(`https://mobilelearn.chaoxing.com/v2/apis/sign/signIn?activeId=%d`, activeID), nil)
	if err != nil {
		zapx.Error("generate generalSign request failed", zap.Error(err))
		return time.Now(), err
	}
	req.Header.Add("Cookie", cookie)
	req.Header.Add("User-Agent", constant.UA)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zapx.Error("general sign failed", zap.Error(err))
		return time.Now(), nil
	}

	result := new(struct {
		Result  int    `json:"result"`
		Message string `json:"msg"`
		Data    struct {
			ID         int64  `json:"id"`
			Name       string `json:"name"`
			SubmitTime string `json:"submittime"`
		} `json:"data"`
	})
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		zapx.Error("decode general sign in response failed", zap.Error(err))
		return time.Now(), err
	}

	if result.Result != 1 {
		return time.Now(), errors.New("general sign not success: " + result.Message)
	}

	// TODO: using data.submitTime
	return time.Now(), nil
}

func locationSign(ctx context.Context, cookie, userID, addr string, activeID int64, longitude, latitude float64) (time.Time, error) {
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodGet, fmt.Sprintf(`https://mobilelearn.chaoxing.com/pptSign/stuSignajax?name=&address=%s&activeId=%d&uid=%s&clientip=&latitude=%v&longitude=%v&fid=1731&appType=15&ifTiJiao=1`, url.QueryEscape(addr), activeID, userID, latitude, longitude), nil)
	req.Header.Add("Cookie", cookie)
	req.Header.Add("User-Agent", constant.UA)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zapx.Error("send location sign in request failed", zap.Error(err))
		return time.Time{}, err
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		zapx.Error("read location sign response body failed", zap.Error(err))
		return time.Time{}, err
	}

	if string(bs) != "success" {
		zapx.Error("location sign failed", zap.Error(err))
		return time.Time{}, err
	}

	return time.Now(), nil
}

func qrCodeSign(ctx context.Context, cookie, enc string, activeID int64) (time.Time, error) {
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodGet, fmt.Sprintf(`https://mobilelearn.chaoxing.com/pptSign/stuSignajax?enc=%s&name=&activeId=%d&uid=122949003&clientip=&useragent=&latitude=-1&longitude=-1&fid=1731&appType=15`, enc, activeID), nil)
	req.Header.Add("Cookie", cookie)
	req.Header.Add("User-Agent", constant.UA)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zapx.Error("send qrCode sign in failed", zap.Error(err))
		return time.Time{}, err
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		zapx.Error("read qrCode sign response failed", zap.Error(err))
		return time.Time{}, err
	}

	if string(bs) != "success" {
		zapx.Error("qrCode sign failed", zap.Error(err))
		return time.Time{}, err
	}

	return time.Now(), nil
}
