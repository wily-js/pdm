package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"pdm/appconf"
	"pdm/controller"
)

func NewHttpServer(config *appconf.Application) *http.Server {
	var r *gin.Engine
	if config.Debug {
		r = gin.Default()
	} else {
		r = gin.New()
	}
	// 注册路路由
	controller.RouteMapping(r)
	zap.L().Info("系统启动", zap.Int("port", config.Port))
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: r,
	}
}
