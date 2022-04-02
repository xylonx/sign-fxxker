package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/xylonx/sign-fxxker/internal/config"
	"github.com/xylonx/sign-fxxker/internal/core"
	"github.com/xylonx/sign-fxxker/internal/router"
	"github.com/xylonx/zapx"
	"go.uber.org/zap"
)

var httpServer *http.Server

func StartService() (err error) {
	httpAddr := fmt.Sprintf("%v:%v", config.Config.Http.Host, config.Config.Http.Port)
	httpServer = router.NewHttpServer(&router.HttpOption{
		Addr:         httpAddr,
		ReadTimeout:  time.Second * time.Duration(config.Config.Http.ReadTimeoutSeconds),
		WriteTimeout: time.Second * time.Duration(config.Config.Http.WriteTimeoutSeconds),
		AllowOrigins: config.Config.Http.AllowOrigins,
	})

	go func() {
		zapx.Info("start http server", zap.String("host", httpAddr))
		if err := httpServer.ListenAndServe(); err != nil {
			zapx.Error("http run error", zap.Error(err))
		}
	}()

	// auto sign chaoxing course
	for _, c := range core.ChaoxingCourse {
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

func StopService(ctx context.Context) {
	httpServer.Shutdown(ctx)
}
