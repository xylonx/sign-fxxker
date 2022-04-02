package router

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/xylonx/sign-fxxker/internal/config"
	"github.com/xylonx/sign-fxxker/internal/core"
)

type HttpOption struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	AllowOrigins []string
}

func NewHttpServer(o *HttpOption) *http.Server {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// config cors
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = o.AllowOrigins
	corsConfig.AllowCredentials = true
	r.Use(cors.New(corsConfig))

	// register routers
	router := r.Group("/api/v1")

	router.GET("/course", GetAllClassHandler)
	router.PUT("/course", UpdateCourseQrEncHandler)

	return &http.Server{
		Addr:         o.Addr,
		Handler:      r,
		ReadTimeout:  o.ReadTimeout,
		WriteTimeout: o.WriteTimeout,
	}
}

////////////////
//	handler	  //
////////////////

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

func GetAllClassHandler(c *gin.Context) {
	classes := make([]string, 0, len(config.Config.Course.Chaoxing))
	for _, c := range config.Config.Course.Chaoxing {
		classes = append(classes, c.Alias)
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: classes})
}

func UpdateCourseQrEncHandler(c *gin.Context) {
	info := new(struct {
		CourseName string `json:"course_name"`
		Enc        string `json:"enc"`
	})
	if err := c.ShouldBindJSON(info); err != nil {
		c.JSON(http.StatusBadRequest, Response{Success: false, Data: err.Error()})
		return
	}

	for i := range core.ChaoxingCourse {
		if core.ChaoxingCourse[i].CourseName == info.CourseName {
			core.ChaoxingCourse[i].UpdateQrCodeEnc(time.Now(), info.Enc)
		}
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: "success"})
}
