package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"pdm/appconf/dir"
	"pdm/controller/dto"
	"pdm/controller/middle"
	"pdm/logg/applog"
	"pdm/repo"
	"pdm/repo/entity"
	"pdm/reuint"
	"pdm/reuint/jwt"
	"runtime"
	"sort"
	"strings"
	"time"
)

// NewTechnicalProposalController 创建技术方案
func NewTechnicalProposalController(router gin.IRouter) *TechnicalProposalController {
	res := &TechnicalProposalController{}
	r := router.Group("/technicalProposal")
	// 列表
	r.GET("/list", Authed, res.list)
	// 下载
	r.GET("/download", Authed, res.download)
	// 上传
	r.POST("/upload", Authed, res.upload)
	// 查看属性
	r.GET("/attr", Authed, res.attr)
	// 复制
	r.POST("/copy", Authed, res.copy)
	// 删除
	r.DELETE("/remove", Authed, res.remove)
	// 移动 / 重命名
	r.POST("/move", Authed, res.move)
	// 创建目录
	r.POST("/mkdir", Authed, res.mkdir)
	// 搜索
	r.GET("/search", Authed, res.search)

	return res
}

// TechnicalProposalController 技术方案控制器
type TechnicalProposalController struct {
}

/**
@api {GET} /api/technicalProposal/attr 查看属性
@apiDescription 查看文件或文件夹的属性。
@apiName TechnicalProposalAttr
@apiGroup TechnicalProposal

@apiPermission 已认证用户

@apiParam {String} path 文件路径，URI编码（相对于基础文档区的根目录的绝对路径）

@apiParamExample {get} 请求示例
GET /api/technicalProposal/attr?path=/%E6%A0%87%E5%87%86/

@apiSuccess {String} name 文件名。
@apiSuccess {String} updatedAt 最后一次更新时间，格式"YYYY-MM-DD HH:mm:ss"。
@apiSuccess {String} path 文件路径（相对于基础文档区的根目录的绝对路径）。
@apiSuccess {String} type 文件类型。
@apiSuccess {Integer} size 文件大小（单位 B）。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
{
	"name": "电子签章.pdf",
	"type": "file",
	"updatedAt": "2022-12-05 13:26:14",
	"path": "/标准/行业标准/国密标准/电子签章.pdf",
	"size": 1024
}

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// attr 查看属性
func (c *TechnicalProposalController) attr(ctx *gin.Context) {
	fileDir, _ := url.QueryUnescape(ctx.Query("path"))

	p := filepath.Join(dir.TechnicalProposalDir, fileDir)
	if !strings.HasPrefix(p, dir.TechnicalProposalDir) {
		ErrIllegal(ctx, "文件路径错误")
		return
	}

	info, err := os.Stat(p)
	if os.IsNotExist(err) {
		ErrIllegal(ctx, "文件不存在")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	zone := time.FixedZone("CST", 8*3600)
	res := &dto.FileItemDto{}
	res.Name = info.Name()
	res.UpdatedAt = info.ModTime().In(zone).Format("2006-01-02 15:04:05")
	res.Path = fileDir
	if info.IsDir() {
		res.Type = "dir"
	} else {
		res.Type = "file"
	}
	res.Size = info.Size()

	ctx.JSON(200, res)
}

/**
@api {POST} /api/technicalProposal/copy 复制
@apiDescription 复制文件/文件夹。
@apiName TechnicalProposalCopy
@apiGroup TechnicalProposal

@apiPermission 管理员

@apiParam {String} from 原文目录。
@apiParam {String} to 目标文件目录。

@apiParamExample {json} 请求示例
{
	"from": "/标准/国家标准/电子签章.pdf",
	"to": "/标准/行业标准/国密标准/电子签章.pdf"
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// copy 复制
func (c *TechnicalProposalController) copy(ctx *gin.Context) {
	var fileCopy dto.FileTransferDto
	err := ctx.BindJSON(&fileCopy)

	// 记录日志
	applog.L(ctx, "基础文档区复制文件", map[string]interface{}{
		"from": fileCopy.From,
		"to":   fileCopy.To,
	})

	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	from := filepath.Join(dir.TechnicalProposalDir, fileCopy.From)
	if !strings.HasPrefix(from, dir.TechnicalProposalDir) || from == dir.TechnicalProposalDir {
		ErrIllegal(ctx, "文件路径错误")
		return
	}
	info, _ := os.Stat(from)
	to := filepath.Join(dir.TechnicalProposalDir, fileCopy.To)
	if !strings.HasPrefix(to, dir.TechnicalProposalDir) || to == dir.TechnicalProposalDir || (info.IsDir() && strings.HasPrefix(to, from+string(filepath.Separator))) {
		ErrIllegal(ctx, "文件路径错误")
		return
	}
	_, err = os.Stat(to)
	if err == nil {
		ErrIllegal(ctx, "文件已存在")
		return
	}
	if os.IsNotExist(err) {
		if info.IsDir() {
			err := reuint.CopyDir(from, to)
			if err != nil {
				ErrSys(ctx, err)
				return
			}
			return
		}
		fromFile, err := os.Open(from)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		defer fromFile.Close()

		toFile, err := os.Create(to)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		defer toFile.Close()

		_, err = io.Copy(toFile, fromFile)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

}

/**
@api {DELETE} /api/technicalProposal/remove 删除
@apiDescription 删除文件或文件夹。
@apiName TechnicalProposalRemove
@apiGroup TechnicalProposal

@apiPermission 管理员

@apiParam {String} path 文件路径，URI编码（相对于基础文档区的根目录的绝对路径）

@apiParamExample {delete} 请求示例
DELETE /api/technicalProposal/remove?path=/%E6%A0%87%E5%87%86/

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// remove 删除
func (c *TechnicalProposalController) remove(ctx *gin.Context) {
	fileDir, _ := url.QueryUnescape(ctx.Query("path"))

	// 记录日志
	applog.L(ctx, "基础文档区删除文件", map[string]interface{}{
		"fileDir": fileDir,
	})

	p := filepath.Join(dir.TechnicalProposalDir, fileDir)
	if !strings.HasPrefix(p, dir.TechnicalProposalDir) || p == dir.TechnicalProposalDir {
		ErrIllegal(ctx, "文件路径错误")
		return
	}

	err := os.RemoveAll(p)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {POST} /api/technicalProposal/move 移动
@apiDescription 移动或重命名文件/文件夹。
@apiName TechnicalProposalMove
@apiGroup TechnicalProposal

@apiPermission 已认证用户

@apiParam {String} from 原文目录。
@apiParam {String} to 目标文件目录。

@apiParamExample {JSON} 移动
{
	"from": "/标准/国家标准/电子签章.pdf",
	"to": "/标准/行业标准/国密标准/电子签章.pdf"
}

@apiParamExample {JSON} 重命名
{
	"from": "/标准/行业标准/国密标准/电子签章.pdf",
	"to": "/标准/行业标准/国密标准/GBT33190-2016电子文件存储与交换格式版式文档.pdf"
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// move 移动或重命名文件
func (c *TechnicalProposalController) move(ctx *gin.Context) {
	param := &dto.FileTransferDto{}
	err := ctx.ShouldBindJSON(param)
	// 记录日志
	applog.L(ctx, "基础文档区移动或重命名文件", map[string]interface{}{
		"from": param.From,
		"to":   param.To,
	})

	if err != nil {
		ErrIllegal(ctx, "参数无法解析")
		return
	}

	param.From = filepath.Join(dir.TechnicalProposalDir, param.From)
	param.To = filepath.Join(dir.TechnicalProposalDir, param.To)
	if !strings.HasPrefix(param.From, dir.TechnicalProposalDir) || param.From == dir.TechnicalProposalDir {
		ErrIllegal(ctx, "文件原路径错误")
		return
	}
	info, err := os.Stat(param.From)
	if errors.Is(err, os.ErrNotExist) {
		ErrIllegal(ctx, "原文件不存在")
		return
	}
	if !strings.HasPrefix(param.To, dir.TechnicalProposalDir) || param.To == dir.TechnicalProposalDir || (info.IsDir() && strings.HasPrefix(param.To, param.From+string(filepath.Separator))) {
		ErrIllegal(ctx, "文件目标路径错误")
		return
	}
	if _, err := os.Stat(param.To); err == nil {
		ErrIllegal(ctx, "文件已存在")
		return
	}
	err = os.Rename(param.From, param.To)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {GET} /api/technicalProposal/list 列表
@apiDescription 返回指定目录下所有的文件或目录。
@apiName TechnicalProposalList
@apiGroup TechnicalProposal

@apiPermission 已认证用户

@apiParam {String} [base] 搜索目录，URI编码（相对于基础文档区的根目录的绝对路径）,空表示根路径。
@apiParam {String="dir", "file"} [type] 文件类型 不带该参数则文件夹和文件全部展示
<ul>
	<li>dir - 目录</li>
	<li>file - 文件</li>
</ul>
@apiParam {String} [keyword] 关键字，URI编码。

@apiParamExample {get} 请求示例
GET /api/technicalProposal/list?base=/%E6%A0%87%E5%87%86/

@apiSuccess {FileItem[]} Body 查询结果列表。

@apiSuccess (FileItem) {String} name 文件名。
@apiSuccess (FileItem) {String="dir"} [type=""] 文件类型
<ul>
	<li>dir - 目录</li>
	<li>file - 文件</li>
</ul>

@apiSuccess (FileItem) {String} path 文件路径（相对于基础文档区的根目录的绝对路径）。
@apiSuccess (FileItem) {String} updatedAt 最后一次更新时间，格式"YYYY-MM-DD HH:mm:ss"。
@apiSuccess (FileItem) {Integer} size 文件大小，默认单位为（B）。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[{
	"name": "电子签章.pdf",
	"type": "file",
	"path": "/标准/行业标准/国密标准/电子签章.pdf",
	"updatedAt": "2022-12-05 13:26:14",
	"size":200
}]

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// list 查看文件列表
func (c *TechnicalProposalController) list(ctx *gin.Context) {
	baseDir, _ := url.QueryUnescape(ctx.Query("base"))
	keyword, _ := url.QueryUnescape(ctx.Query("keyword"))
	typStr := ctx.Query("type")
	onlyDir := false
	if strings.EqualFold(typStr, "dir") {
		onlyDir = true
	}

	p := filepath.Join(dir.TechnicalProposalDir, baseDir)
	if !strings.HasPrefix(p, dir.TechnicalProposalDir) {
		ErrIllegal(ctx, "搜索基础路径错误")
		return
	}
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	projectIdList := make([]int, 0)

	if err := repo.DB.Model(&entity.ProjectMember{}).Select("project_id").Find(&projectIdList, "user_id = ?", claims.Sub).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
	projectNameList := make([]string, 0)
	if err := repo.DB.Model(&entity.Project{}).Select("name").Find(&projectNameList, "id in ? ", projectIdList).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
	res := []dto.FileItemDto{}
	zone := time.FixedZone("CST", 8*3600)
	_ = filepath.Walk(p, func(path string, info fs.FileInfo, err error) error {

		if path == p {
			// 忽略根路径
			return nil
		}

		if onlyDir && !info.IsDir() {
			// 在仅查询目录的情况忽略 文件
			return nil
		}

		if keyword != "" && !strings.Contains(info.Name(), keyword) {
			// 关键字不匹配
			return nil
		}
		if len(baseDir) <= 0 {
			for _, val := range projectNameList {
				if info.Name() != val && info.IsDir() {
					return filepath.SkipDir
				}
			}
		}
		item := &dto.FileItemDto{}
		item.Name = info.Name()
		item.UpdatedAt = info.ModTime().In(zone).Format("2006-01-02 15:04:05")
		item.Path = strings.Replace(path, dir.TechnicalProposalDir, "", 1)
		if runtime.GOOS == "windows" {
			item.Path = strings.Replace(item.Path, "\\", "/", -1)
		}
		item.Size = info.Size()
		item.Type = "file"
		if info.IsDir() {
			item.Type = "dir"
		}
		res = append(res, *item)

		if info.IsDir() {
			// 阻止递归
			return filepath.SkipDir
		}
		return nil
	})
	sort.Slice(res, func(i, j int) bool {
		if res[i].Type != res[j].Type {
			return len(res[i].Type) < len(res[j].Type)
		}
		return res[i].Name < res[j].Name
	})

	ctx.JSON(200, res)
}

/**
@api {GET} /api/technicalProposal/download 下载
@apiDescription 下载文件或文件夹（文件夹以打包成压缩包的形式下载）。
@apiName TechnicalProposalDownload
@apiGroup TechnicalProposal

@apiPermission 已认证用户

@apiParam {String} path 下载资源所在目录，URI编码（相对于基础文档区的根目录的绝对路径），多个资源以","隔开。

path= area/1,2,3,4
@apiParamExample {get} 请求示例
GET /api/technicalProposal/download?path=/%E6%A0%87%E5%87%86/,/%E6%8A%A5%E5%91%8A/

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/
// download 下载文件
func (c *TechnicalProposalController) download(ctx *gin.Context) {
	pathStr, _ := url.QueryUnescape(ctx.Query("path"))
	fileList := []string{}
	paths := ctx.QueryArray("path")
	for _, item := range paths {
		if p, _ := url.QueryUnescape(item); strings.TrimSpace(p) != "" {
			fileList = append(fileList, p)
		}
	}
	if len(fileList) == 0 {
		ErrIllegal(ctx, "下载文件路径为错误")
		return
	}

	// 记录日志
	applog.L(ctx, "基础文档区下载文件", map[string]interface{}{
		"path": pathStr,
	})

	for i, filename := range fileList {
		filePath := filepath.Join(dir.TechnicalProposalDir, filename)
		if !strings.HasPrefix(filePath, dir.TechnicalProposalDir) || filePath == dir.TechnicalProposalDir {
			ErrIllegal(ctx, "路径错误")
			return
		}
		fileList[i] = filePath
	}

	info, err := os.Stat(fileList[0])
	if err != nil {
		ErrIllegal(ctx, "路径错误")
		return
	}

	// 单文件下载
	if len(fileList) == 1 && !info.IsDir() {
		file, err := os.Open(fileList[0])
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		defer file.Close()
		// 下载文件名称
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", url.QueryEscape(info.Name())))
		//获取文件的后缀(文件类型)
		ctx.Header("Content-Type", reuint.GetMIME(path.Ext(info.Name())))
		_, err = io.Copy(ctx.Writer, file)
		if err != nil {
			ErrSys(ctx, err)
		}
		// 单文件情况提前返回
		return
	}

	// 多文件或目录 打包压缩下载
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", url.QueryEscape(info.Name())))
	ctx.Header("Content-Type", "application/zip")
	err = reuint.Zip(ctx.Writer, fileList...)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {POST} /api/technicalProposal/upload 上传
@apiDescription 以表单的方式上传文件列表。

@apiName TechnicalProposalUpload
@apiGroup TechnicalProposal

@apiPermission 已认证用户

@apiHeader {String} Content-type multipart/form-data 多类型表单固定值。

@apiParam {String} base 上传文件所在路径。
@apiParam {[]File} files 上传文件列表。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// upload 上传文件
func (c *TechnicalProposalController) upload(ctx *gin.Context) {
	// 获取表单文件
	form, _ := ctx.MultipartForm()
	files := form.File["files"]
	// 获取表单的上传文件路径
	base := ctx.PostForm("base")

	// 记录日志
	applog.L(ctx, "基础文档区上传文件", map[string]interface{}{
		"path": base,
	})
	var err1 error
	for _, file := range files {
		_ = os.MkdirAll(filepath.Join(dir.TechnicalProposalDir, base), os.ModePerm)
		filePath := filepath.Join(dir.TechnicalProposalDir, base, file.Filename)
		_, err := os.Stat(filePath)
		if err == nil {
			err1 = errors.New(file.Filename + "文件已存在")
			continue
			//return
		}
		err = ctx.SaveUploadedFile(file, filePath)
		if err != nil {
			err1 = err
			continue
			//return
		}
	}
	if err1 != nil {
		ErrIllegalE(ctx, err1)
		return
	}
}

/**
@api {POST} /api/technicalProposal/mkdir 创建目录
@apiDescription 创建目录，若包含多个级则递归创建。
@apiName TechnicalProposalMkdir
@apiGroup TechnicalProposal

@apiPermission 已认证用户

@apiParam {String} Body 目录名称，采用Linux风格路径，注意检查路径越界访问。

@apiParamExample {JSON} 创建目录
/标准/行业标准/国密标准/

@apiSuccess {String} name 文件名

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
"新建文件夹 (5)"

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/
// 创建文件夹
func (c *TechnicalProposalController) mkdir(ctx *gin.Context) {
	var base string
	err := ctx.BindJSON(&base)

	// 记录日志
	applog.L(ctx, "基础文档区创建文件夹", map[string]interface{}{
		"path": base,
	})

	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	p := filepath.Join(dir.TechnicalProposalDir, base)
	if !strings.HasPrefix(p, dir.TechnicalProposalDir) || p == dir.TechnicalProposalDir {
		ErrIllegal(ctx, "路径错误")
		return
	}
	if runtime.GOOS == "windows" {
		p = strings.Replace(p, "\\", "/", -1)
	}

	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		ErrIllegal(ctx, "创建失败")
		return
	}
	lastIndex := strings.LastIndex(p, "/")
	var name string
	if lastIndex != -1 {
		name = p[lastIndex+1:]
	}
	ctx.JSON(200, name)
}

/**
@api {GET} /api/technicalProposal/search 搜索
@apiDescription 返回所有指定目录（递归子文件）下包含关键字名称的文件或目录。
@apiName TechnicalProposalSearch
@apiGroup TechnicalProposal

@apiPermission 已认证用户

@apiParam {String} [base] 搜索目录，URI编码（相对于基础文档区的根目录的绝对路径），空表示根路径。
@apiParam {String} keyword 关键字，URI编码。

@apiParamExample {get} 请求示例
GET /api/technicalProposal/search?base=/%E6%A0%87%E5%87%86/&keyword=%E6%8A%A5%E5%91%8A

@apiSuccess {FileItem[]} Body 查询结果列表。

@apiSuccess (FileItem) {String} name 文件名。
@apiSuccess (FileItem) {String="dir", "file"} type 文件类型。
<ul>
	<li>dir - 目录</li>
	<li>file - 文件</li>
</ul>

@apiSuccess (FileItem) {Integer} size 文件大小，单位：“字节(B)”，仅在文件类型为 file 时有效。
@apiSuccess (FileItem) {String} path 文件路径（相对于基础文档区的根目录的绝对路径），注意检查文件越界。
@apiSuccess (FileItem) {String} updatedAt 最后一次更新时间，格式"YYYY-MM-DD HH:mm:ss"。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
    {
        "name": "国密标准",
        "type": "dir",
        "path": "/标准/国密标准",
        "size": 0,
        "updatedAt": "2022-12-05 16:03:18"
    },
    {
        "name": "国密标准_text.txt",
        "type": "file",
        "path": "/标准/国密标准/国密标准_text.txt",
        "size": 0,
        "updatedAt": "2022-12-05 16:02:59"
    }
]

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// search 搜索文件
func (c *TechnicalProposalController) search(ctx *gin.Context) {
	baseDir, _ := url.QueryUnescape(ctx.Query("base"))
	keyword, _ := url.QueryUnescape(ctx.Query("keyword"))

	if keyword == "" {
		ErrIllegal(ctx, "搜索关键字不能为空")
		return
	}

	p := filepath.Join(dir.TechnicalProposalDir, baseDir)
	if !strings.HasPrefix(p, dir.TechnicalProposalDir) {
		ErrIllegal(ctx, "搜索路径错误")
		return
	}

	res := []dto.FileItemDto{}
	zone := time.FixedZone("CST", 8*3600)
	_ = filepath.Walk(p, func(path string, info fs.FileInfo, err error) error {
		// 忽略根路径
		if path == p {
			return nil
		}

		if !strings.Contains(info.Name(), keyword) {
			// 关键字不匹配
			return nil
		}

		item := &dto.FileItemDto{}
		item.Name = info.Name()
		item.UpdatedAt = info.ModTime().In(zone).Format("2006-01-02 15:04:05")
		item.Path = strings.Replace(path, dir.TechnicalProposalDir, "", 1)
		if runtime.GOOS == "windows" {
			item.Path = strings.Replace(item.Path, "\\", "/", -1)
		}
		item.Size = info.Size()
		item.Type = "file"
		if info.IsDir() {
			item.Type = "dir"
		}
		res = append(res, *item)
		return nil
	})
	ctx.JSON(200, res)
}
