package controller

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"mime"
	"net/http"
	"pdm/appconf"
	"pdm/appconf/dir"
	"pdm/controller/middle"
)

// token管理器
var (
	tokenManager *middle.TokenManager
)

// 编辑锁
var (
	editLock *middle.EditLock
)

// RouteMapping HTTP路由注册
// r: 路由注册器
func RouteMapping(r gin.IRouter, cfg *appconf.Application) {
	// 中间件 - 拦截器 按顺序依次执行
	tokenManager = middle.NewTokenFilter()
	editLock = middle.NewEditLock()
	r.Use(
		middle.Recovery(),
		middle.Anonymous,
		tokenManager.Filter,
	)

	// 根目录默认为Web静态资源目录
	// 以可执行文件的目录下的web目录作为静态文件位置提供web服务
	r.Use(static.Serve("/ui/", static.LocalFile(dir.UiDir, true)))
	_ = mime.AddExtensionType(".js", "application/javascript")

	r.GET("/", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/ui/#/")
	})

	// 所有RestFul接口都以 /api开始
	r = r.Group("/api")
	NewLoginController(r)
	NewUserController(r)
	NewProjectController(r)
	NewSystemInfoController(r)
	NewPublicController(r)
	NewAuthorityController(r)
	NewCategorizeController(r)
	NewCasesController(r)
	NewOperationLogController(r)
	NewProgramLogController(r)
	NewSsoController(r, cfg.SSOBaseUrl)
	NewBaseDocumentAreaController(r)
	NewRootCertsController(r)
	NewDocController(r)
	NewTechnicalProposalController(r)
}
