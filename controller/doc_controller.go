package controller

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"math/rand"
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
	"regexp"
	"strconv"
	"strings"
	"time"
)

// NewDocController 创建文档控制器
func NewDocController(router gin.IRouter) *DocController {
	res := &DocController{}
	r := router.Group("/doc")
	// 创建项目文档
	r.POST("/create", ExceptProjectInterConnector, res.create)
	// 获取文档信息
	r.GET("/info", Authed, res.info)
	// 获取项目文档列表
	r.GET("/projectDocList", Authed, res.projectDocList)
	// 更新文档内容
	r.POST("/content", ExceptProjectInterConnector, res.contentPost)
	// 获取文档内容
	r.GET("/content", Authed, res.contentGet)
	// 上传文档资源
	r.POST("/assert", ExceptProjectInterConnector, res.assertPost)
	// 下载文档资源
	r.GET("/assert", ExceptProjectInterConnector, res.assertGet)
	// 获取文档编辑锁
	r.POST("/lock", ExceptProjectInterConnector, res.lock)
	// 取消编辑
	r.POST("/cancel", ExceptProjectInterConnector, res.cancel)
	// 导出文档
	r.GET("/export", ExceptProjectInterConnector, res.export)
	// 生成技术方案
	r.GET("/generate", ExceptProjectInterConnector, res.generate)
	return res
}

// DocController 文档控制器
type DocController struct {
}

/**
@api {POST} /api/doc/create 创建
@apiDescription 创建项目文档
@apiName DocCreate
@apiGroup Doc

@apiPermission 项目负责人、开发

@apiHeader {String} Content-type multipart/form-data 多类型表单固定值。

@apiParam {Integer} projectId 项目ID。
@apiParam {String} title 文档名。
@apiParam {String} [docType=markdown] 文档类型
<ul>
    <li>markdown（默认）</li>
    <li>office - 包括ppt、doc、excel等</li>
    <li>pdf</li>
    <li>txt</li>
</ul>

@apiParam {Integer} priority 优先级，默认为0，数值越大优先级越高显示位置越靠前。
@apiParam {File} [file] 文档资源。


@apiSuccess {Integer} body 文档记录ID。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

32

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// create 创建项目文档
func (c *DocController) create(ctx *gin.Context) {
	var doc entity.Document
	var project entity.Project

	var err error
	doc.ProjectId, _ = strconv.Atoi(ctx.PostForm("projectId"))
	doc.Title = ctx.PostForm("title")
	// 记录日志
	applog.L(ctx, "创建项目文档", map[string]interface{}{
		"projectId": doc.ProjectId,
		"title":     doc.Title,
	})

	if doc.ProjectId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	if doc.Title == "undefined" || doc.Title == "null" || doc.Title == "" {
		ErrIllegal(ctx, "文档名称为空")
		return
	}

	doc.DocType = ctx.PostForm("docType")
	doc.Priority, err = strconv.Atoi(ctx.PostForm("priority"))
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	err = repo.DB.Where("id = ? AND is_delete = ? ", doc.ProjectId, 0).Find(&project).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	err = repo.DB.Transaction(func(tx *gorm.DB) error {
		// 判断项目文档所属项目是否存在
		if project.ID == 0 {
			return errors.New("项目不存在或项目已被删除")
		}
		// 创建文档记录
		err = tx.Create(&doc).Error
		if err != nil {
			return err
		}

		// 创建文件目录
		docPath := filepath.Join(dir.DocDir, strconv.Itoa(doc.ID))
		err = os.MkdirAll(docPath, os.ModePerm)
		if err != nil {
			return err
		}

		// 生成文件名称
		now := time.Now().Format("20060102150405")
		r := fmt.Sprintf("%04v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10000))

		// markdown 自动生成 其他按照上传文件
		if doc.DocType == "markdown" {
			filename := fmt.Sprintf("%s%s.md", now, r)
			doc.Filename = filename
			docFile := filepath.Join(docPath, doc.Filename)
			_, err = os.Create(docFile)
			if err != nil {
				return err
			}
		} else {
			// 获取表单的文件
			file, err := ctx.FormFile("file")
			if err != nil {
				return errors.New("文件格式错误")
			}
			// 获取文件格式
			fileFormat := strings.Split(file.Filename, ".")
			filename := ""
			if len(fileFormat) == 1 {
				filename = fmt.Sprintf("%s%s", now, r)
			} else {
				filename = fmt.Sprintf("%s%s.%s", now, r, fileFormat[len(fileFormat)-1])
			}

			doc.Filename = filename
			docFile := filepath.Join(docPath, doc.Filename)
			err = ctx.SaveUploadedFile(file, docFile)
			if err != nil {
				return err
			}
		}
		find := tx.Where("id", doc.ID).Find(&entity.Document{})
		err := find.Update("filename", doc.Filename).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		ErrIllegalE(ctx, err)
		return
	}

	ctx.JSON(200, doc.ID)
}

/**
@api {GET} /api/doc/info 获取文档信息
@apiDescription 获取文档信息，注意文档信息仅由项目成员可以获取，非项目成员返回无权限访问错误。
@apiName DocInfo
@apiGroup Doc

@apiPermission 项目负责人、开发、测试、维护

@apiParam {Integer} id 文档ID。
@apiParam {Integer} projectId 项目ID。

@apiParamExample {json} 请求示例
GET /api/doc/info?id=13&projectId=1

@apiSuccess {Integer} id 文档ID。
@apiSuccess {Integer} projectId 项目ID。
@apiSuccess {String} title 标题。
@apiSuccess {Integer} docType 文档类型
<ul>

	<li>markdown（默认）</li>
	<li>office - 包括ppt、doc、excel等</li>
	<li>pdf</li>
	<li>txt</li>

</ul>
@apiSuccess {Integer} priority 优先级，默认为0，数值越大优先级越高显示位置越靠前。
@apiSuccess {String} filename 文件名称。
@apiSuccess {String} updatedAt 更新时间

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

	{
	    "id": 10,
	    "projectId": 1,
	    "title": "详细设计",
	    "docType": "markdown",
	    "priority": 0,
		"filename":"202211021516555220.png"
		"updatedAt": "2020-09-26 11:29:44",
	}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// info 获取文档信息
func (c *DocController) info(ctx *gin.Context) {
	var doc entity.Document
	var docDto dto.DocDto
	id, _ := strconv.Atoi(ctx.Query("id"))
	if id <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	projectId, _ := strconv.Atoi(ctx.Query("projectId"))
	if projectId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	err := repo.DB.Model(&doc).Where("id", id).Find(&docDto).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.JSON(200, &docDto)
}

/**
@api {GET} /api/doc/projectDocList 项目文档列表
@apiDescription 获取指定项目的项目文档列表，注意文档信息仅由项目成员可以获取，非项目成员返回无权限访问错误。
后可根据不同的用户角色有限的过滤部分文档内容。

@apiName DocProjectDocList
@apiGroup Doc

@apiPermission 项目负责人、开发、测试、维护

@apiParam {Integer} projectId 项目ID

@apiParamExample {http} 请求示例

GET /api/doc/projectDocList?projectId=1

@apiSuccess {DocInfo[]} body 项目文档列表。

@apiSuccess {Object} DocInfo 文档信息
@apiSuccess {Integer} DocInfo.id 文档ID。
@apiSuccess {String} DocInfo.title 标题。
@apiSuccess {Integer} DocInfo.projectId 项目ID。
@apiSuccess {String} DocInfo.docType 文档类型
<ul>
    <li>markdown（默认）</li>
    <li>word</li>
    <li>excel</li>
    <li>txt</li>
</ul>
@apiSuccess {Integer} DocInfo.priority 优先级，默认为0，数值越大优先级越高显示位置越靠前。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
    {
        "id": 14,
        "projectId": 1,
        "title": "详细设计",
        "docType": "markdown",
        "priority": 0
    }
]

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析

*/

// projectDocList 项目文档列表
func (c *DocController) projectDocList(ctx *gin.Context) {
	var doc entity.Document
	//var docDto dto.DocDto
	var docDtoList []dto.DocDto
	projectId, _ := strconv.Atoi(ctx.Query("projectId"))
	if projectId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	err := repo.DB.Model(&doc).Where("project_id", projectId).Order("priority desc").Find(&docDtoList).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.JSON(200, &docDtoList)
}

/**
@api {POST} /api/doc/content 更新文档内容
@apiDescription 以表单的方式上传新的文档内容替换原有的文档内容。

注意文档内容仅由项目成员可以更新，非项目成员返回无权限访问错误。

@apiName DocContentPOST
@apiGroup Doc

@apiPermission 项目负责人、开发

@apiHeader {String} Content-type multipart/form-data 多类型表单固定值。

@apiParam {Integer} docId 文档ID。
@apiParam {Integer} projectId 项目ID。
@apiParam {String} title 文档名
@apiParam {File} [file] 上传文件，文档类型非markdown时该字段更新文件。
@apiParam {String} [query] 更新类型 - file 上传文档 - title 仅修改文档名称
<ul>
    <li>file</li>
    <li>title</li>
</ul>
@apiParam {String} [content] 文档内容，当文档类型为markdown时使用该字段更新文档。
@apiParam {Boolean} [autoSave] 是否自动保存

@apiSuccess {Integer} body 文档记录ID。


@apiSuccessExample 成功响应
HTTP/1.1 200 OK

13

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// contentPost 更新文档内容
func (c *DocController) contentPost(ctx *gin.Context) {
	var doc entity.Document

	formDocId := ctx.PostForm("docId")
	docId, _ := strconv.Atoi(formDocId)
	// 记录日志
	applog.L(ctx, "更新文档内容", map[string]interface{}{
		"docId": docId,
	})
	if docId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	projectId, _ := strconv.Atoi(ctx.PostForm("projectId"))
	if projectId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	title := ctx.PostForm("title")
	if title == "" {
		ErrIllegal(ctx, "文档名称不可为空")
		return
	}

	// 判断是否为该项目的用户
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	// 获取文档信息
	err := repo.DB.First(&doc, "id = ?", docId).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "该文档不存在或被删除")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	if doc.DocType == "markdown" {
		content := ctx.PostForm("content")

		autoSave, err := strconv.ParseBool(ctx.PostForm("autoSave"))
		if err != nil {
			ErrIllegal(ctx, "参数非法，无法解析")
			return
		}

		var user entity.User
		// 判断用户是否拥有读写锁
		if autoSave == false {
			// 获取拥有该锁的用户ID
			v := editLock.Query(projectId, doc.ID, "doc")
			// 若无人拥有该锁或拥有该锁的用户为申请者本身
			if v == middle.NoLock || v == claims.Sub {
				editLock.Lock(claims.Sub, projectId, doc.ID, "doc")
				defer editLock.Unlock(projectId, doc.ID, "doc")
			} else {
				repo.DB.Where("id", v).Find(&user)
				ErrIllegal(ctx, fmt.Sprintf("%s正在编辑该文档", user.Name))
				return
			}
		}

		// 打开一个存在的文件，将原来的内容覆盖掉
		filename := filepath.Join(dir.DocDir, formDocId, doc.Filename)
		// O_WRONLY: 只写, O_TRUNC: 清空文件
		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			ErrIllegal(ctx, "文件打开错误")
			return
		}

		defer file.Close()

		// 防止删除文件本身
		fileContent := fmt.Sprintf("(%s &file=%s)", content, doc.Filename)

		// 删除文件中未被引用的文档资源
		err = reuint.DeleteUnreferencedFiles(fileContent, filepath.Join(dir.DocDir, strconv.Itoa(docId)))
		if err != nil {
			ErrSys(ctx, err)
			return
		}

		// 带缓冲区的*Writer
		writer := bufio.NewWriter(file)

		_, err = writer.WriteString(content)
		if err != nil {
			ErrSys(ctx, err)
			return
		}

		// 将缓冲区中的内容写入到文件里
		err = writer.Flush()
		if err != nil {
			ErrSys(ctx, err)
			return
		}

		err = repo.DB.Where("id", doc.ID).Find(&entity.Document{}).Update("title", title).Error
		if err != nil {
			ErrSys(ctx, err)
			return
		}
	} else {
		query := ctx.PostForm("query")
		// 若上传文档资源
		if query == "file" {
			// 获取表单的文件
			file, err := ctx.FormFile("file")
			if err != nil {
				ErrSys(ctx, err)
				return
			}

			// 删除原有文件
			if doc.Filename != "" {
				removePath := filepath.Join(dir.DocDir, formDocId, doc.Filename)
				err := os.Remove(removePath)
				if err != nil {
					ErrSys(ctx, err)
					return
				}
			}

			// 生成文件名称
			now := time.Now().Format("20060102150405")
			r := fmt.Sprintf("%04v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10000))
			// 获取文件格式
			fileFormat := strings.Split(file.Filename, ".")
			filename := ""
			if len(fileFormat) == 1 {
				filename = fmt.Sprintf("%s%s", now, r)
			} else {
				filename = fmt.Sprintf("%s%s.%s", now, r, fileFormat[len(fileFormat)-1])
			}
			doc.Filename = filename
			// 生成文件
			docFile := filepath.Join(dir.DocDir, strconv.Itoa(doc.ID), doc.Filename)
			err = ctx.SaveUploadedFile(file, docFile)
			if err != nil {
				ErrSys(ctx, err)
				return
			}

			err = repo.DB.Where("id", doc.ID).Find(&entity.Document{}).Updates(entity.Document{Filename: doc.Filename, Title: title}).Error
			if err != nil {
				ErrSys(ctx, err)
				return
			}

		} else if query == "title" {
			// 若只修改标题
			err = repo.DB.Where("id", doc.ID).Find(&entity.Document{}).Update("Title", title).Error
			if err != nil {
				ErrSys(ctx, err)
				return
			}
		}

	}

}

/**
@api {GET} /api/doc/content 获取文档内容
@apiDescription 获取文档内容。

注意文档内容仅由项目成员可以获取，非项目成员返回无权限访问错误。

@apiName DocContentGet
@apiGroup Doc

@apiPermission markdown-项目负责人、开发、测试、维护 其他类型-项目负责人、开发

@apiParam {Integer} docId 文档ID。
@apiParam {Integer} projectId 项目ID。

@apiParamExample {HTTP} 请求示例

GET /api/doc/content?docId=11&projectId=1

@apiSuccess {Object} body 文本或文件内容，若文档类型为非文本格式，则该接口为下载该文档。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

# 详细设计

第一章 详细设计说明说...

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// contentGet 获取文档内容
func (c *DocController) contentGet(ctx *gin.Context) {
	var doc entity.Document
	//var docDto dto.DocDto
	id := ctx.Query("docId")
	projectId, _ := strconv.Atoi(ctx.Query("projectId"))
	if projectId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	err := repo.DB.Where("id", id).Find(&doc).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	filePath := filepath.Join(dir.DocDir, id, doc.Filename)
	if doc.DocType == "markdown" {
		file, err := os.Open(filePath)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		ctx.JSON(200, string(content))
	} else {
		// 判断是否可下载
		claimsValue, _ := ctx.Get(middle.FlagClaims)
		claims := claimsValue.(*jwt.Claims)
		info := &entity.ProjectMember{}
		err := repo.DB.First(info, "project_id = ? AND user_id = ?", projectId, claims.Sub).Error
		if err == gorm.ErrRecordNotFound {
			ErrIllegal(ctx, "权限错误")
			return
		}
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		if info.Role == 1 || info.Role == 3 {
			ErrIllegal(ctx, "权限不足")
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			ErrIllegal(ctx, "文件解析失败")
			return
		}
		defer file.Close()

		// 下载文件名称
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", doc.Filename))

		ctx.Header("Content-Type", "application/octet-stream")

		// 流复制文件
		_, err = io.Copy(ctx.Writer, file)
		if err != nil {
			ErrSys(ctx, err)
			return
		}

	}
}

/**
@api {POST} /api/doc/assert 上传文档资源
@apiDescription 以表单的方式上传文档相关的图片或附件，文件上传后台需要判断文档类型是图片还是附件类型。

注意文档资源仅由项目成员可以使用，非项目成员返回无权限访问错误。

文件上传后返回资源的类型以及，资源访问的路径。

资源访问路径为文档资源的下载接口，格式为: "/api/doc/assert?docId=xxxx&file=202210131620160001.png"

资源文件名格式为：`YYYYMMDDHHmmss` + `4位随机数`

@apiName DocAssertPOST
@apiGroup Doc

@apiPermission 项目负责人、开发

@apiHeader {String} Content-type multipart/form-data 多类型表单固定值。

@apiParam {Integer} docId 文档ID。
@apiParam {Integer} projectId 项目ID。
@apiParam {File} file 资源文件，文档相关的图片或文件附件。

@apiSuccess {String} type 资源类型
<ul>

	<li>image</li>
	<li>file</li>

</ul>
@apiSuccess {String} uri 资源访问地址，资源访问路径为文档资源的下载接口，格式为: "/api/doc/assert?docId=xxxx&file=202210131620160001.png"

@apiSuccessExample {json} 成功响应
HTTP/1.1 200 OK

	{
	    "type": "image",
	    "uri": "/api/doc/assert?docId=xxxx&file=202210131620160001.png"
	}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// assertPost 上传文档资源
func (c *DocController) assertPost(ctx *gin.Context) {
	var docUri dto.UriDto

	// 获取表单的id
	id := ctx.PostForm("docId")
	// 记录日志
	applog.L(ctx, "上传文档资源", map[string]interface{}{
		"StageId": id,
	})
	projectId, _ := strconv.Atoi(ctx.PostForm("projectId"))
	if projectId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 获取表单的文件
	file, err := ctx.FormFile("file")
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 获取文件类型
	contentType := file.Header.Get("Content-Type")
	if strings.Contains(contentType, "image") {
		docUri.DocType = "image"
	} else {
		docUri.DocType = "file"
	}

	filename := reuint.GenTimeFileName(file.Filename)

	filePath := filepath.Join(dir.DocDir, id, filename)
	err = ctx.SaveUploadedFile(file, filePath)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	uri := fmt.Sprintf("/api/doc/assert?docId=%s&file=%s", id, filename)
	docUri.Uri = uri
	ctx.JSON(200, docUri)
}

/**
@api {GET} /api/doc/assert 下载文档资源
@apiDescription 下载文档相关的资源。

用户仅能下载自己有关的项目的文档资源，非项目成员返回无权限访问错误。

@apiName DocAssertGET
@apiGroup Doc

@apiPermission 项目负责人、开发、测试、维护

@apiParam {Integer} docId 文档ID。
@apiParam {String} file 文件名称。

@apiParamExample {http} 请求示例

GET /api/doc/assert?docId=xxxx&file=202210131620160001.png

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// assert 下载文档资源
func (c *DocController) assertGet(ctx *gin.Context) {

	id := ctx.Query("docId")
	filename := ctx.Query("file")

	// 文件路径
	filePath := filepath.Join(dir.DocDir, id, filename)

	// 防止用户通过 ../../ 的方式下载到操作系统内的重要文件
	if !strings.HasPrefix(filePath, filepath.Join(dir.DocDir, id)) {
		ErrIllegal(ctx, "文件路径错误")
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		ErrIllegal(ctx, "文件解析失败")
		return
	}
	defer file.Close()

	// 下载文件名称
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	fileNameWithSuffix := path.Base(filename)
	//获取文件的后缀(文件类型)
	fileType := path.Ext(fileNameWithSuffix)

	ctx.Header("Content-Type", reuint.GetMIME(fileType))

	_, err = io.Copy(ctx.Writer, file)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {POST} /api/doc/lock 获取文档编辑锁
@apiDescription 获取文档编辑锁
@apiName DocLock
@apiGroup Doc

@apiPermission 项目负责人、开发

@apiParam {Integer} userId 用户ID
@apiParam {Integer} projectId  项目ID
@apiParam {Integer} id        文档ID
@apiParam {String}  docType  文档类型

@apiParamExample {json} 请求示例

	{
	    "userId": 2,
	    "projectId": 5,
		"id":3,
		"docType":"doc",
	}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// lock 获取文档编辑锁
func (c *DocController) lock(ctx *gin.Context) {
	var lockDto dto.LockDto
	var user entity.User

	err := ctx.BindJSON(&lockDto)
	// 记录日志
	applog.L(ctx, "获取文档编辑锁", map[string]interface{}{
		"userId": lockDto.UserId,
		"id":     lockDto.Id,
	})
	if err != nil {
		ErrIllegal(ctx, "参数解析错误")
		return
	}

	// 获取锁信息
	v := editLock.Query(lockDto.ProjectId, lockDto.Id, lockDto.DocType)
	// 若无人拥有该锁或拥有该锁的用户为申请者本身
	if v == middle.NoLock || v == lockDto.UserId {
		editLock.Lock(lockDto.UserId, lockDto.ProjectId, lockDto.Id, lockDto.DocType)
		ctx.JSON(200, "获取到锁")
	} else {
		repo.DB.Where("id", v).Find(&user)
		ErrIllegal(ctx, fmt.Sprintf("%s正在编辑该文档", user.Name))
		return
	}
}

/**
@api {POST} /api/doc/cancel 取消编辑
@apiDescription 取消编辑
@apiName DocCancel
@apiGroup Doc

@apiPermission 项目负责人、开发

@apiParam {Integer} userId 用户ID
@apiParam {Integer} projectId  项目ID
@apiParam {Integer} id        文档ID
@apiParam {String}  docType  文档类型

@apiParamExample {json} 请求示例

	{
	    "userId": 2,
	    "projectId": 5,
		"id":3,
		"docType":"doc",
	}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数解析错误
*/

// cancel 取消编辑
func (c *DocController) cancel(ctx *gin.Context) {
	var lockDto dto.LockDto
	var doc entity.Document

	err := ctx.BindJSON(&lockDto)
	if err != nil {
		ErrIllegal(ctx, "参数解析错误")
		return
	}

	// 记录日志
	applog.L(ctx, "文档取消编辑", map[string]interface{}{
		"userId": lockDto.UserId,
		"id":     lockDto.Id,
	})

	// 获取锁信息
	v := editLock.Query(lockDto.ProjectId, lockDto.Id, lockDto.DocType)
	if v == lockDto.UserId {
		editLock.Unlock(lockDto.ProjectId, lockDto.Id, lockDto.DocType)
	} else {
		ErrIllegal(ctx, "操作异常")
		return
	}

	err = repo.DB.Where("id", lockDto.Id).Find(&doc).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	docId := strconv.Itoa(lockDto.Id)
	filePath := filepath.Join(dir.DocDir, docId, doc.Filename)
	file, err := os.Open(filePath)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 防止删除文件本身
	fileContent := fmt.Sprintf("(%s &file=%s)", string(content), doc.Filename)

	// 删除文件中未被引用的文档资源
	err = reuint.DeleteUnreferencedFiles(fileContent, filepath.Join(dir.DocDir, strconv.Itoa(doc.ID)))
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.JSON(200, string(content))
}

/**
@api {GET} /api/doc/export 导出文档
@apiDescription 导出文档。

仅可导出markdown类型的文档

@apiName DocExport
@apiGroup Doc

@apiPermission 项目负责人

@apiParam {String} docId 文档ID。

@apiParamExample {http} 请求示例

GET /api/doc/export?docId=46

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// export 导出文档
func (c *DocController) export(ctx *gin.Context) {
	var doc entity.Document
	docId := ctx.Query("docId")

	applog.L(ctx, "导出文档", map[string]interface{}{
		"docId": docId,
	})

	err := repo.DB.First(&doc, "id = ?", docId).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	//仅可导出markdown类型格式文档
	if doc.DocType != "markdown" {
		ErrIllegal(ctx, "格式错误，无法导出")
		return
	}

	docFilePath := filepath.Join(dir.DocDir, docId)

	// 生成临时文件夹 文件夹名称格式 文件名_更新时间
	temporaryFolderName := fmt.Sprintf("%s_%s", doc.Title, doc.UpdatedAt.Format("20060102"))

	temp, err := os.MkdirTemp("", temporaryFolderName)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 无论导出文件是否成功，都需要删除临时文件
	defer os.RemoveAll(temp)

	err = reuint.CopyTempDir(docFilePath, temp)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 读取文件
	temporaryFilename := filepath.Join(temp, doc.Filename)

	// 防止用户通过 ../../ 的方式下载到操作系统内的重要文件
	if !strings.HasPrefix(temporaryFilename, temp) {
		ErrIllegal(ctx, "文件路径错误")
		return
	}

	file, err := os.Open(temporaryFilename)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	file.Close()

	// 修改文件中的资源访问路径
	file, err = os.OpenFile(temporaryFilename, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	reg := regexp.MustCompile("\\/api\\/doc\\/assert\\?docId=[0-9]+(\\\\)?&file=")
	results := reg.ReplaceAllString(string(content), "./")

	// 将修改后的内容写入文件

	// 带缓冲区的*Writer
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(results)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 将缓冲区中的内容写入到文件里
	err = writer.Flush()
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	file.Close()

	//打包压缩下载
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", temporaryFolderName))
	ctx.Header("Content-Type", "application/zip")

	err = reuint.Zip(ctx.Writer, temp)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {GET} /api/doc/generate 导出文档
@apiDescription 生成技术方案。


@apiName DocGenerate
@apiGroup Doc

@apiPermission 项目负责人

@apiParam {String} docId 文档ID。

@apiParamExample {http} 请求示例

GET /api/doc/generate?docId=46

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// generate 生成技术方案
func (c *DocController) generate(ctx *gin.Context) {
	var doc entity.Document
	var technicalProposal entity.TechnicalProposal
	docId := ctx.Query("docId")

	applog.L(ctx, "生成技术方案", map[string]interface{}{
		"docId": docId,
	})
	if err := repo.DB.First(&doc, "id = ?", docId).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	projectName, err := repo.ProjectRepo.GetProjectName(ctx)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	tpDir := filepath.Join(dir.TechnicalProposalDir, projectName)
	err = repo.DB.First(&technicalProposal, "project_id = ?", claims.PID).Error
	if err == gorm.ErrRecordNotFound {
		_ = os.MkdirAll(tpDir, os.ModePerm)
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		ErrSys(ctx, err)
		return
	}
	docFilePath := filepath.Join(dir.DocDir, docId)

	// 生成目标文件夹 文件夹名称格式 文件名_更新时间
	filename := fmt.Sprintf("%s_%s", doc.Title, doc.UpdatedAt.Format("20060102"))
	destPath := filepath.Join(tpDir, filename)

	_ = os.MkdirAll(destPath, os.ModePerm)
	if err = reuint.CopyDir(docFilePath, destPath); err != nil {
		ErrSys(ctx, err)
		return
	}
	temp := filepath.Join(destPath, doc.Filename)

	dest, err := os.OpenFile(temp, os.O_RDWR, os.ModePerm)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 读取文件

	// 防止用户通过 ../../ 的方式下载到操作系统内的重要文件
	if !strings.HasPrefix(dest.Name(), tpDir) {
		ErrIllegal(ctx, "文件路径错误")
		return
	}

	content, err := io.ReadAll(dest)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	reg := regexp.MustCompile("\\/api\\/doc\\/assert\\?docId=[0-9]+(\\\\)?&file=")
	results := reg.ReplaceAllString(string(content), "./")
	// 将修改后的内容写入文件
	// 带缓冲区的*Writer
	writer := bufio.NewWriter(dest)
	if _, err = writer.WriteString(results); err != nil {
		ErrSys(ctx, err)
		return
	}

	// 将缓冲区中的内容写入到文件里
	if err = writer.Flush(); err != nil {
		ErrSys(ctx, err)
		return
	}
	dest.Close()
	res := entity.TechnicalProposal{
		Name:      doc.Filename,
		ProjectId: claims.PID,
	}
	if err = repo.DB.Create(&res).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
}
