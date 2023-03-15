package dir

import (
	"log"
	"os"
	"path/filepath"
)

var (
	base                 string // 程序运行目录
	LogDir               string // 日志存储目录
	WebDir               string // web文件存储目录
	InterfaceDir         string // 接口目录文件存储目录
	UiDir                string // 前端文件存储目录
	DockingDocDir        string //对接文档文件存储目录
	TechnicalProposalDir string // 技术方案存储目录
	RootCertDir          string // 根证书管理
)

func Init() {
	base, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	LogDir = filepath.Join(base, "logs")
	WebDir = filepath.Join(base, "web")
	InterfaceDir = filepath.Join(base, "interface")
	UiDir = filepath.Join(base, "ui")
	DockingDocDir = filepath.Join(base, "dockingDoc")
	TechnicalProposalDir = filepath.Join(base, "technicalProposal")
	RootCertDir = filepath.Join(base, "rootCerts")
	_ = os.MkdirAll(LogDir, os.ModePerm)
	_ = os.MkdirAll(WebDir, os.ModePerm)
	_ = os.MkdirAll(InterfaceDir, os.ModePerm)
	_ = os.MkdirAll(UiDir, os.ModePerm)
	_ = os.MkdirAll(DockingDocDir, os.ModePerm)
	_ = os.MkdirAll(TechnicalProposalDir, os.ModePerm)
	_ = os.MkdirAll(RootCertDir, os.ModePerm)

	log.Println("程序运行目录:", base)
	log.Println("日志存储目录:", LogDir)
	log.Println("web文件存储目录:", WebDir)
	log.Println("接口目录文件存储目录:", InterfaceDir)
	log.Println("前端文件存储目录:", UiDir)
	log.Println("对接文档文件存储目录:", DockingDocDir)
	log.Println("技术方案存储目录:", TechnicalProposalDir)
	log.Println("根证书目录:", RootCertDir)
}
