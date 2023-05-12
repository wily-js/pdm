package dir

import (
	"log"
	"os"
	"path/filepath"
)

var (
	base                 string // 程序运行目录
	LogDir               string // 日志存储目录
	InterfaceDir         string // 接口目录文件存储目录
	UiDir                string // 前端文件存储目录
	TechnicalProposalDir string // 技术方案存储目录
	RootCertDir          string // 根证书管理
	AvatarDir            string // 头像存储目录
	BaseDocAreaDir       string // 基础文档区存储目录
	DocDir               string // 对接文档文件存储目录
)

func Init() {
	base, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	LogDir = filepath.Join(base, "logs")
	InterfaceDir = filepath.Join(base, "interface")
	UiDir = filepath.Join(base, "ui")
	TechnicalProposalDir = filepath.Join(base, "technicalProposal")
	RootCertDir = filepath.Join(base, "rootCerts")
	AvatarDir = filepath.Join(base, "avatar")
	BaseDocAreaDir = filepath.Join(base, "baseDocArea")
	DocDir = filepath.Join(base, "doc")

	_ = os.MkdirAll(LogDir, os.ModePerm)
	_ = os.MkdirAll(InterfaceDir, os.ModePerm)
	_ = os.MkdirAll(UiDir, os.ModePerm)
	_ = os.MkdirAll(TechnicalProposalDir, os.ModePerm)
	_ = os.MkdirAll(RootCertDir, os.ModePerm)
	_ = os.MkdirAll(AvatarDir, os.ModePerm)
	_ = os.MkdirAll(BaseDocAreaDir, os.ModePerm)
	_ = os.MkdirAll(DocDir, os.ModePerm)

	log.Println("程序运行目录:", base)
	log.Println("日志存储目录:", LogDir)
	log.Println("接口目录文件存储目录:", InterfaceDir)
	log.Println("前端文件存储目录:", UiDir)
	log.Println("技术方案存储目录:", TechnicalProposalDir)
	log.Println("根证书目录:", RootCertDir)
	log.Println("头像存储目录:", AvatarDir)
	log.Println("基础文档储目录:", BaseDocAreaDir)
	log.Println("对接文档文件存储目录:", DocDir)

}
