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

type ActiveListResp struct {
	Result int `json:"result"`
	Data   struct {
		ActiveList []ActiveList `json:"activeList"`
	} `json:"data"`
}

type ActiveList struct {
	SignName  string `json:"nameOne"`
	AttendNum int    `json:"attendNum"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
	ActiveID  int64  `json:"id"`
}

type BaiduLocation struct {
	Addr      string
	Longitude float64
	Latitude  float64
}

type ChaoXingCourse struct {
	courseName string
	user       user.User
	location   BaiduLocation
	interval   time.Duration
	delay      time.Duration

	courseID int64
	classID  int64

	mux            sync.RWMutex
	qrCodeEnc      string
	qrCodeUpdateAt time.Time
}

type ChaoXingCourseOptions struct {
	CourseName      string
	User            user.User
	Location        BaiduLocation
	IntervalSeconds int64
	DelaySeconds    int64
	CourseID        int64
	ClassID         int64
}

func NewChaoXingCourse(opt *ChaoXingCourseOptions) *ChaoXingCourse {
	return &ChaoXingCourse{
		courseName: opt.CourseName,
		user:       opt.User,
		location:   opt.Location,
		interval:   time.Duration(opt.IntervalSeconds) * time.Second,
		delay:      time.Duration(opt.DelaySeconds) * time.Second,
		courseID:   opt.CourseID,
		classID:    opt.ClassID,
		mux:        sync.RWMutex{},
	}
}

func (c *ChaoXingCourse) StartAutoSign() <-chan string {
	ch := make(chan string, 2)
	go func() {
		for {
			time.Sleep(c.interval)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
			alist, err := c.getActiveList(ctx, c.courseID, c.classID)
			if err != nil {
				cancel()
				continue
			}
			cancel()

			if len(alist) != 0 {
				for _, active := range alist {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
					if err := c.autoSign(ctx, active); err != nil {
						cancel()
						ch <- fmt.Sprintf("%s 课程[%s]签到[%s]失败\n", time.Now(), c.courseName, active.SignName)
						continue
					}
					cancel()
					ch <- fmt.Sprintf("%s 课程[%s]签到[%s]成功\n", time.Now(), c.courseName, active.SignName)
				}
			}

		}
	}()
	return ch
}

func (c *ChaoXingCourse) autoSign(ctx context.Context, sign ActiveList) error {
	if time.UnixMilli(sign.EndTime).Sub(time.Now()) > c.delay*3 {
		time.Sleep(c.delay)
	}
	switch sign.SignName {
	case GeneralSign:
		return c.generalSign(ctx, sign.ActiveID)
	case LocationSign:
		return c.locationSign(ctx, c.location.Addr, sign.ActiveID, c.location.Longitude, c.location.Latitude)
	case QrCodeSign:
		enc, updatedAt := c.getQrCode()
		if updatedAt.UnixMilli() < sign.StartTime {
			return nil
		}
		return c.qrCodeSign(ctx, enc, sign.ActiveID)
	default:
		return fmt.Errorf("sign type[%v] is not supported now", sign.SignName)
	}
}

// getActiveList - return the un signed activity list
func (c *ChaoXingCourse) getActiveList(ctx context.Context, courseID, classID int64) ([]ActiveList, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, fmt.Sprintf(`https://mobilelearn.chaoxing.com/v2/apis/active/student/activelist?fid=0&courseId=%d&classId=%d&_=%d`, courseID, classID, time.Now().UnixMilli()),
		nil,
	)
	if err != nil {
		zapx.Error("generate get activeList request failed", zap.Error(err))
		return nil, err
	}
	c.user.SetCredentials(req)
	req.Header.Add("User-Agent", constant.UA)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zapx.Error("get activeList failed", zap.Error(err))
		return nil, err
	}

	alistResp := new(ActiveListResp)
	if err := json.NewDecoder(resp.Body).Decode(alistResp); err != nil {
		zapx.Error("decode activeList failed", zap.Error(err))
		return nil, err
	}

	alist := make([]ActiveList, 0, len(alistResp.Data.ActiveList))
	for _, a := range alistResp.Data.ActiveList {
		if time.UnixMilli(a.EndTime).After(time.Now()) {
			if succ, _ := c.checkUnsigned(ctx, a.ActiveID); succ {
				alist = append(alist, a)
			}
		}
	}

	return alist, nil
}

func (c *ChaoXingCourse) checkUnsigned(ctx context.Context, activeID int64) (bool, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, fmt.Sprintf(`https://mobilelearn.chaoxing.com/v2/apis/sign/getAttendInfo?activeId=%v`, activeID), nil)
	if err != nil {
		return false, err
	}
	c.user.SetCredentials(req)
	req.Header.Add("User-Agent", constant.UA)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zapx.Error("send check signed failed", zap.Error(err))
		return false, err
	}

	result := new(struct {
		Result  int    `json:"result"`
		Message string `json:"msg"`
		Data    struct {
			Status int `json:"status"`
		} `json:"data"`
	})
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		zapx.Error("unmarshal check signed response failed", zap.Error(err))
		return false, err
	}

	return result.Data.Status == 0, nil
}

func (c *ChaoXingCourse) generalSign(ctx context.Context, activeID int64) error {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, fmt.Sprintf(`https://mobilelearn.chaoxing.com/v2/apis/sign/signIn?activeId=%d`, activeID), nil)
	if err != nil {
		zapx.Error("generate generalSign request failed", zap.Error(err))
		return err
	}
	c.user.SetCredentials(req)
	req.Header.Add("User-Agent", constant.UA)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zapx.Error("general sign failed", zap.Error(err))
		return nil
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
		return err
	}

	if result.Result != 1 {
		return errors.New("general sign not success: " + result.Message)
	}

	return nil
}

func (c *ChaoXingCourse) locationSign(ctx context.Context, addr string, activeID int64, longitude, latitude float64) error {
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodGet, fmt.Sprintf(`https://mobilelearn.chaoxing.com/pptSign/stuSignajax?name=&address=%s&activeId=%d&uid=%s&clientip=&latitude=%v&longitude=%v&fid=1731&appType=15&ifTiJiao=1`, url.QueryEscape(addr), activeID, c.user.GetUID(), latitude, longitude), nil)
	c.user.SetCredentials(req)
	req.Header.Add("User-Agent", constant.UA)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zapx.Error("send location sign in request failed", zap.Error(err))
		return err
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		zapx.Error("read location sign response body failed", zap.Error(err))
		return err
	}

	if string(bs) != "success" {
		zapx.Error("location sign failed", zap.Error(err))
		return err
	}

	return nil
}

/*******************
*    QrCode sign   *
********************/

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

func (c *ChaoXingCourse) qrCodeSign(ctx context.Context, enc string, activeID int64) error {
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodGet, fmt.Sprintf(`https://mobilelearn.chaoxing.com/pptSign/stuSignajax?enc=%s&name=&activeId=%d&uid=122949003&clientip=&useragent=&latitude=-1&longitude=-1&fid=1731&appType=15`, enc, activeID), nil)
	c.user.SetCredentials(req)
	req.Header.Add("User-Agent", constant.UA)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zapx.Error("send qrCode sign in failed", zap.Error(err))
		return err
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		zapx.Error("read qrCode sign response failed", zap.Error(err))
		return err
	}

	if string(bs) != "success" {
		zapx.Error("qrCode sign failed", zap.Error(err))
		return err
	}

	return nil
}
