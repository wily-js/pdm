package main

import (
	"go.uber.org/zap"
	"pdm/appconf"
	"pdm/appconf/dir"
	"pdm/logg"
	"pdm/logg/applog"
	"pdm/repo"
)

func main() {
	// 初始化各级目录
	dir.Init()
	// 加载配置文件配置
	appcfg := appconf.Load()
	//fmt.Println("main appcfgACL = ", appcfg.ACL)
	// 初始化日志
	logg.InitConsole(appcfg.Debug)

	// 数据库初始化
	err := repo.Init(appcfg)
	if err != nil {
		zap.L().Fatal("持久层初始化失败", zap.Error(err))
	}
	// 初始化操作日志模块
	applog.InitLogger(appcfg)

	// 启动Web服务器
	server := NewHttpServer(appcfg)
	if err = server.ListenAndServe(); err != nil {
		zap.L().Fatal("服务启动失败", zap.Error(err))
	}
}
