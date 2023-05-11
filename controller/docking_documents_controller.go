package controller

import (
	"encoding/json"
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
	"strconv"
	"strings"
	"time"
)

// NewDockingDocumentsController 创建版本控制器
func NewDockingDocumentsController(router gin.IRouter) *DockingDocumentsController {
	res := &DockingDocumentsController{}
	r := router.Group("/docking")
	// 创建对接文档
	r.POST("/create", ExceptProjectInterConnector, res.create)
	// 编辑对接文档
	r.POST("/edit", ExceptProjectInterConnector, res.edit)
	// 上传附件
	r.POST("/upload", ExceptProjectInterConnector, res.upload)
	// 删除附件
	r.DELETE("/remove", ExceptProjectInterConnector, res.remove)
	// 删除对接文档
	r.DELETE("/delete", ExceptProjectInterConnector, res.delete)
	// 查询所有对接文档
	r.GET("/all", ProjectMember, res.all)
	// 查询对接文档信息
	r.GET("", ProjectMember, res.info)
	// 上传文件资源
	r.POST("/assert", ExceptProjectInterConnector, res.assertPost)
	// 下载文档资源
	r.GET("/assert", ExceptProjectInterConnector, res.assertGet)
	// 导出文档
	r.GET("/export", ExceptProjectInterConnector, res.export)
	return res

}

// DockingDocumentsController 版本控制器
type DockingDocumentsController struct {
}

/**
@api {POST} /api/docking/create 创建
@apiDescription 创建对接文档

@apiName DockingCreate
@apiGroup Docking

@apiPermission 项目负责人、开发

@apiParam {String} name 对接文档名称。
@apiParam {Integer} projectId  项目ID

@apiSuccess {Integer} id 对接文档ID。

@apiParamExample {json} 请求示例
{
    "name": "test",
	"projectId": 5
}


@apiSuccessExample 成功响应
HTTP/1.1 200 OK

15
@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// create 创建对接文档
func (c *DockingDocumentsController) create(ctx *gin.Context) {
	var info entity.DockingDocuments
	err := ctx.BindJSON(&info)
	// 记录参数
	applog.L(ctx, "创建对接文档", map[string]interface{}{
		"projectId": info.ProjectId,
		"name":      info.Name,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 发布者
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	if info.ProjectId != claims.PID {
		ErrForbidden(ctx, "权限错误")
		return
	}
	info.UserId = claims.Sub
	info.Assets = "{}"
	if err = repo.DB.Create(&info).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
	// 创建文件目录
	dockingDocPath := filepath.Join(dir.DockingDocDir, strconv.Itoa(info.ID))
	if err = os.MkdirAll(dockingDocPath, os.ModePerm); err != nil {
		ErrSys(ctx, err)
		return
	}

	assertPath := filepath.Join(dockingDocPath, "asset")
	if err = os.MkdirAll(assertPath, os.ModePerm); err != nil {
		ErrSys(ctx, err)
		return
	}
	docPath := filepath.Join(dockingDocPath, "doc")
	if err = os.MkdirAll(docPath, os.ModePerm); err != nil {
		ErrSys(ctx, err)
		return
	}
	ctx.JSON(200, info.ID)
}

/**
@api {POST} /api/docking/edit 编辑
@apiDescription 编辑已经发布的对接文档内容

@apiName DockingCreate
@apiGroup Docking

@apiPermission 发布者（项目负责人、开发）

@apiParam {Integer} id  对接文档ID
@apiParam {String} [content] 版本发布说明，Markdown格式。


@apiParamExample {json} 请求示例
{
	"id": 3,
    "content": "# 问题测试"
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// edit 编辑对接文档文档内容
func (c *DockingDocumentsController) edit(ctx *gin.Context) {
	var info entity.DockingDocuments

	err := ctx.BindJSON(&info)
	applog.L(ctx, "编辑对接文档", map[string]interface{}{
		"dockingDocId": info.ID,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	var dockingDocuments entity.DockingDocuments
	// 查找对接文档
	err = repo.DB.First(&dockingDocuments, "id = ?", info.ID).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "不存在该对接文档")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 判断是否为该对接文档发布者
	if dockingDocuments.UserId != claims.Sub {
		ErrForbidden(ctx, "权限错误")
		return
	}
	dockingDocuments.Content = info.Content
	if err = repo.DB.Save(&dockingDocuments).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
	// 删除文件夹中未被引用的文件
	if err = reuint.DeleteUnreferencedFiles(info.Content, filepath.Join(dir.DockingDocDir, strconv.Itoa(info.ID), "doc")); err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {POST} /api/docking/upload 上传附件
@apiDescription 上传版本附件

@apiName DockingUpload
@apiGroup Docking

@apiPermission 发布者(项目负责人、开发)

@apiHeader {String} Content-type multipart/form-data 多类型表单固定值。

@apiParam {Integer} id  对接文档ID
@apiParam {File} file 资源文件，文档相关的图片或文件附件。

@apiSuccess {Map} assets 附件列表。


@apiSuccessExample 成功响应
HTTP/1.1 200 OK

{"202211221012472143.png": "/api/docking/assert?releaseId=6&file=202211221012472143.png"}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// upload 上传附件
func (c *DockingDocumentsController) upload(ctx *gin.Context) {
	var dockingDocuments entity.DockingDocuments
	dockingDocId, _ := strconv.Atoi(ctx.PostForm("id"))
	// 接收前端传递文件
	file, err := ctx.FormFile("file")
	//log.Println(file.Filename)
	applog.L(ctx, "上传附件", map[string]interface{}{
		"dockingDocId": dockingDocId,
	})
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if dockingDocId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 判断是否为该对接文档发布者
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	err = repo.DB.First(&dockingDocuments, "id = ?", dockingDocId).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "不存在该对接文档")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if dockingDocuments.UserId != claims.Sub {
		ErrForbidden(ctx, "权限错误")
		return
	}

	// 将附件列表json 转化为 map
	assertMap := map[string]string{}
	err = json.Unmarshal([]byte(dockingDocuments.Assets), &assertMap)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 判断文件名是否已存在
	_, ok := assertMap[file.Filename]
	if ok {
		ErrIllegal(ctx, "文件已存在")
		return
	}

	// 保存文件
	filePath := filepath.Join(dir.DockingDocDir, strconv.Itoa(dockingDocId), "asset", file.Filename)
	err = ctx.SaveUploadedFile(file, filePath)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 更新附件列表
	uri := fmt.Sprintf("/api/docking/assert?id=%d&type=%s&file=%s", dockingDocId, "asset", file.Filename)
	assertMap[file.Filename] = uri
	marshal, err := json.Marshal(assertMap)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	dockingDocuments.Assets = string(marshal)
	if err = repo.DB.Save(&dockingDocuments).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
	ctx.JSON(200, assertMap)
}

/**
@api {DELETE} /api/docking/remove 删除附件
@apiDescription 删除文档附件
该接口仅在数据库操作异常时返回500系统错误的状态码，其他情况均返回200。

@apiName DockingRemove
@apiGroup Docking

@apiPermission 发布者(项目负责人、开发)

@apiParam {Integer} id  对接文档ID
@apiParam {String} filename 文件名称。

@apiParamExample {json} 请求示例
DELETE /api/docking/remove?id=11&filename=cc.png

@apiSuccess {Map} assets 附件列表。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

{"202211221012472143.png": "/api/docking/assert?id=6&file=202211221012472143.png"}

@apiErrorExample 失败响应
HTTP/1.1 500 Bad Request

系统内部错误
*/

// remove 删除附件
func (c *DockingDocumentsController) remove(ctx *gin.Context) {
	var dockingDocuments entity.DockingDocuments
	dockingDocId, _ := strconv.Atoi(ctx.Query("id"))
	filename := ctx.Query("filename")

	// 记录日志
	applog.L(ctx, "删除附件", map[string]interface{}{
		"dockingDocId": dockingDocId,
		"filename":     filename,
	})

	if dockingDocId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 查询版本信息
	err := repo.DB.First(&dockingDocuments, "id = ?", dockingDocId).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "对接文档不存在")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 获取附件目录
	var assetMap map[string]string
	err = json.Unmarshal([]byte(dockingDocuments.Assets), &assetMap)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 判断附件是否存在
	_, ok := assetMap[filename]
	if !ok {
		ErrIllegal(ctx, "文件不存在")
		return
	}

	// 删除附件
	delete(assetMap, filename)

	removeFilePath := filepath.Join(dir.DockingDocDir, strconv.Itoa(dockingDocId), "asset", filename)

	_, err = os.Stat(removeFilePath)
	if err == nil {
		err = os.Remove(removeFilePath)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
	}

	marshal, err := json.Marshal(assetMap)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	dockingDocuments.Assets = string(marshal)
	if err = repo.DB.Save(&dockingDocuments).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
	ctx.JSON(200, assetMap)
}

/**
@api {DELETE} /api/docking/delete 删除
@apiDescription 删除已经发布的对接文档，注意该行为只能由项目负责人操作，删除对接文档的同时需要清理
在上传的相关附件。

@apiName DockingDelete
@apiGroup Docking

@apiPermission 项目负责人

@apiParam {String} id 对接文档记录ID
@apiParam {Integer} projectId  项目ID

@apiParamExample {get} 请求示例
DELETE /api/docking/delete?id=11&projectId=5

@apiSuccess {Integer} body 记录ID，由于对接文档内容可能较多，只返还ID号，若需要则调用详情接口查询。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

23

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// delete 对接文档删除
func (c *DockingDocumentsController) delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Query("id"))
	projectId, _ := strconv.Atoi(ctx.Query("projectId"))
	applog.L(ctx, "对接文档删除", map[string]interface{}{
		"id":        id,
		"projectId": projectId,
	})
	if id <= 0 || projectId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	var dockingDocuments entity.DockingDocuments
	err := repo.DB.First(&dockingDocuments, "id = ? AND project_id = ?", id, projectId).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "不存在该对接文档")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 删除对接文档
	if err = repo.DB.Delete(&dockingDocuments).Error; err != nil {
		ErrSys(ctx, err)
		return
	}

	// 删除对接文档文件夹以及文件夹内所有内容
	filePath := filepath.Join(dir.DockingDocDir, strconv.Itoa(id))
	if err = os.RemoveAll(filePath); err != nil {
		ErrSys(ctx, err)
		return
	}
	ctx.JSON(200, dockingDocuments.ID)
}

/**
@api {GET} /api/docking/all 所有对接文档
@apiDescription 查询所有对接文档概要信息。

@apiName DockingAll
@apiGroup Docking

@apiPermission 项目负责人、开发

@apiParam {String} [page=1] 页码。
@apiParam {Integer} [limit=20] 单页多少数据，默认 20。

@apiParam {Integer} projectId 项目ID

@apiParamExample {get} 请求示例
GET /api/docking/all?projectId=5&page=1&limit=30

@apiSuccess {DockingInfo[]} records 查询结果列表。
@apiSuccess {Integer} total 记录总数。
@apiSuccess {Integer} size 每页显示条数，默认 20。
@apiSuccess {Integer} current 当前页。
@apiSuccess {Integer} pages 总页数。

@apiSuccess {Object} DockingInfo 对接文档信息。
@apiSuccess {String} DockingInfo.id 记录ID。
@apiSuccess {String} DockingInfo.name 对接文档名。
@apiSuccess {String} DockingInfo.createdAt 创建时间，格式"YYYY-MM-DD HH:mm:ss"。
@apiSuccess {String} updatedAt 更新时间，格式"YYYY-MM-DD HH:mm:ss"。
@apiSuccess {Map} DockingInfo.assets 附件列表。
@apiSuccess {String} DockingInfo.content 版本更新描述。
@apiSuccess {Object} DockingInfo.publisher 发布者信息。
@apiSuccess {Integer} DockingInfo.publisher.id 用户ID。
@apiSuccess {String} DockingInfo.publisher.name 姓名。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
	"records": [{
		"id": 23,
		"name": "test",
		"createdAt": "2022-10-11 14:23:56",
	    "updatedAt": "2022-10-11 14:23:56",
		"content": "test",
		"asserts": {"202211221012472143.png": "/api/docking/assert?releaseId=6&file=202211221012472143.png"},
		"publisher": { "id" : 13, "name":"张三"}
	}],
	"total": 1,
    "size": 30,
    "current": 1,
    "pages": 1
]

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// all 查询所有版本
func (c *DockingDocumentsController) all(ctx *gin.Context) {
	projectId, _ := strconv.Atoi(ctx.Query("projectId"))
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if projectId <= 0 || page <= 0 || limit <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	dockingDocuments := []entity.DockingDocuments{}
	// 查询对接文档表
	query, tx := repo.NewPageQueryFnc(repo.DB, []entity.DockingDocuments{}, page, limit, func(db *gorm.DB) *gorm.DB {
		db = db.Order("created_at desc").Where("project_id", projectId)
		return db
	})

	err := tx.Find(&dockingDocuments).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	reqInfo := []dto.DockingDocDto{}
	for _, document := range dockingDocuments {
		user := entity.User{}
		// 查询 发布者信息
		if err = repo.DB.Where("id", document.UserId).Find(&user).Error; err != nil {
			ErrSys(ctx, err)
			return
		}

		temp := dto.DockingDocDto{}
		temp.Transform(&document, &user)
		reqInfo = append(reqInfo, temp)
	}
	query.Records = reqInfo
	ctx.JSON(200, query)
}

/**
@api {GET} /api/docking/info 对接文档详情
@apiDescription 查询指定对接文档详情。

@apiName DockingInfo
@apiGroup Docking

@apiPermission 项目负责人、开发

@apiParam {Integer} id 对接文档ID。

@apiParamExample {get} 请求示例
GET /api/docking?id=30

@apiSuccess {Integer} id ID
@apiSuccess {String} name 对接文档名。
@apiSuccess {String} createdAt 创建时间，格式"YYYY-MM-DD HH:mm:ss"。
@apiSuccess {String} updatedAt 更新时间，格式"YYYY-MM-DD HH:mm:ss"。
@apiSuccess {Map} assets 附件列表。
@apiSuccess {String} content 版本更新描述。
@apiSuccess {Object} publisher 发布者信息。
@apiSuccess {Integer} publisher.id 用户ID。
@apiSuccess {String} publisher.name 姓名。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

{
	"id": 23,
	"publisher": {"id": 7, "name": "郭小菊"},
	"name": "test",
	"content": "- 修复了若干问题...",
	"asserts": {"202211221012472143.png": "/api/docking/assert?releaseId=6&file=202211221012472143.png"},
	"createdAt": "2022-10-11 14:23:56",
	"updatedAt": "2022-10-11 14:23:56"
}

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// info 文档详细信息
func (c *DockingDocumentsController) info(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Query("id"))
	if id <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	dockingDocuments := entity.DockingDocuments{}
	err := repo.DB.First(&dockingDocuments, "id = ? ", id).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "对接文档不存在")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	user := entity.User{}
	// 查询 发布者信息
	err = repo.DB.Select("name").Where("id = ? AND is_delete = 0", dockingDocuments.UserId).Find(&user).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	reqInfo := dto.DockingDocDto{}
	reqInfo.Transform(&dockingDocuments, &user)

	ctx.JSON(200, reqInfo)
}

/**
@api {POST} /api/docking/assert 上传文档资源
@apiDescription 以表单的方式上传文档相关的图片或附件，文件上传后台需要判断文档类型是图片还是附件类型。

注意文档资源仅由项目成员可以使用，非项目成员返回无权限访问错误。

文件上传后返回资源的类型以及，资源访问的路径。

资源访问路径为文档资源的下载接口，格式为: "/api/docking/assert?id=1&type=doc&file=202210131620160001.png"

资源文件名格式为：`YYYYMMDDHHmmss` + `4位随机数`

@apiName DockingAssertPOST
@apiGroup Docking

@apiPermission 发布者（项目负责人、开发）

@apiHeader {String} Content-type multipart/form-data 多类型表单固定值。

@apiParam {Integer} id 对接文档ID。
@apiParam {File} file 资源文件，文档相关的图片或文件附件。

@apiSuccess {String} type 资源类型
<ul>

	<li>image</li>
	<li>file</li>

</ul>
@apiSuccess {String} uri 资源访问地址，资源访问路径为文档资源的下载接口，格式为: "/api/docking/assert?id=1&type=doc&file=202210131620160001.png"

@apiSuccessExample {json} 成功响应
HTTP/1.1 200 OK

	{
	    "type": "image",
	    "uri": "/api/docking/assert?id=1&type=doc&file=202210131620160001.png"
	}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// assertPost 上传文件资源
func (c *DockingDocumentsController) assertPost(ctx *gin.Context) {
	var uriDto dto.UriDto
	// 获取表单的文件
	file, err := ctx.FormFile("file")
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	// 获取表单的id
	id := ctx.PostForm("id")
	// 记录日志
	applog.L(ctx, "上传对接文档资源", map[string]interface{}{
		"id": id,
	})

	// 获取文件类型
	contentType := file.Header.Get("Content-Type")

	if strings.Contains(contentType, "image") {
		uriDto.DocType = "image"
	} else {
		uriDto.DocType = "file"
	}

	//生成文件名 - now = `YYYYMMDDHHmmss`   rand = `4位随机数`
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

	filePath := filepath.Join(dir.DockingDocDir, id, "doc", filename)
	err = ctx.SaveUploadedFile(file, filePath)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	uri := fmt.Sprintf("/api/docking/assert?id=%s&type=%s&file=%s", id, "doc", filename)
	uriDto.Uri = uri
	ctx.JSON(200, uriDto)
}

/**
@api {GET} /api/docking/assert 下载文档资源
@apiDescription 下载文档相关的资源。

用户仅能下载自己有关的项目的文档资源，非项目成员返回无权限访问错误。

@apiName DockingAssertGET
@apiGroup Docking

@apiPermission 项目成员

@apiParam {Integer} id 文档ID。
@apiParam {String} file 文件名称。
@apiParam {String} type 文档类型

@apiParamExample {http} 请求示例

GET /api/docking/assert?id=1&type=doc&file=202210131620160001.png

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// assertGet 下载文档资源
func (c *DockingDocumentsController) assertGet(ctx *gin.Context) {
	id := ctx.Query("id")
	filename := ctx.Query("file")
	filetype := ctx.Query("type")

	// 文件路径
	filePath := filepath.Join(dir.DockingDocDir, id, filetype, filename)

	// 防止用户通过 ../../ 的方式下载到操作系统内的重要文件
	if !strings.HasPrefix(filePath, filepath.Join(dir.DockingDocDir, id, filetype, filename)) {
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

	if _, err = io.Copy(ctx.Writer, file); err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {GET} /api/docking/export 导出文档
@apiDescription 导出文档。

仅可导出markdown类型的文档

@apiName DockingExport
@apiGroup Docking

@apiPermission

@apiParam {String} docId 文档ID。

@apiParamExample {http} 请求示例

GET /api/doc/export?docId=46

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// export 导出文档
func (c *DockingDocumentsController) export(ctx *gin.Context) {

}
