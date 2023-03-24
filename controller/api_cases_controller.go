package controller

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"pdm/controller/dto"
	"pdm/controller/middle"
	"pdm/logg/applog"
	"pdm/repo"
	"pdm/repo/entity"
	"pdm/reuint"
	"pdm/reuint/jwt"
	"strconv"
	"strings"
)

// NewCasesController 创建接口用例控制器
func NewCasesController(router gin.IRouter) *CasesController {
	res := &CasesController{}
	r := router.Group("/case")
	// 接口用例创建
	r.POST("/create", res.create)
	// 查询接口用例具体信息
	r.GET("/info", res.info)
	// 编辑接口用例
	r.POST("/edit", res.edit)
	// 删除接口用例
	r.DELETE("/delete", res.delete)
	// 发送测试请求
	r.POST("/send", res.send)
	return res
}

// CasesController 接口用例控制器
type CasesController struct {
}

/**
@api {POST} /api/case/create 创建接口用例
@apiDescription 创建接口用例
接口用例名称在同一级中唯一，不可重复。
该接口只负责创建接口用例名称(请求方法)，默认请求方法为GET，创建数据库记录即可。
@apiName CaseCreate
@apiGroup Case

@apiPermission 项目成员

@apiParam {String} name 接口用例名称。
@apiParam {Integer} [categorizeId] 所属分类ID。
@apiParam {Integer} [method] 接口请求方法：
<ul>
   	<li>0 - GET</li>
    <li>1 - POST</li>
	<li>2 - PUT</li>
	<li>3 - DELETE</li>
</ul>

@apiSuccess {Integer} id 接口用例ID。
@apiSuccess {String} createdAt 创建时间，格式为"YYYY-MM-DD HH:mm:ss"。
@apiSuccess {String} updatedAt 更新时间，格式为"YYYY-MM-DD HH:mm:ss"。
@apiSuccess {String} name 接口用例名称。
@apiSuccess {String} type 类型:"case"

@apiParamExample {json} 请求示例
{
    "name": "create",
    "categorizeId": 1
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

{
    "id": 2,
    "name": "pdm1",
    "createdAt": "2023-03-22 14:05:29",
    "updatedAt": "2023-03-22 14:05:29",
    "type": "case"
}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

接口名称已经存在
*/

// create 创建接口用例
func (c *CasesController) create(ctx *gin.Context) {
	var info entity.ApiCase
	if err := ctx.BindJSON(&info); err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	applog.L(ctx, "创建接口用例", map[string]interface{}{
		"name":         info.Name,
		"categorizeId": info.CategorizeId,
	})
	// 用例名称不能为空
	if len(strings.Trim(info.Name, " ")) == 0 {
		ErrIllegal(ctx, "用例名称不能为空")
		return
	}
	// 用例名称唯一
	exist, err := repo.CaseRepo.ExistName(info.Name, info.CategorizeId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if exist {
		ErrIllegal(ctx, "用例名称已经存在")
		return
	}
	// 创建人ID
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	info.UserId = claims.Sub

	if err = repo.DB.Create(&info).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
	reqInfo := dto.CategorizeListDto{}
	reqInfo.Transform(&entity.ApiCategorize{}, &info, "case")
	ctx.JSON(200, reqInfo)

}

/**
@api {GET} /api/case/info 查询接口用例信息
@apiDescription 查询出具体接口用例信息。

@apiName CaseInfo
@apiGroup Case

@apiPermission 项目成员

@apiParam {Integer} id 接口用例ID。

@apiParamExample {get} 请求示例
GET /api/case/info?id=1


@apiSuccess {Integer} id 用例ID。
@apiSuccess {String} createdAt 创建时间。
@apiSuccess {String} updatedAt 更新时间。
@apiSuccess {String} name 用例名称。
@apiSuccess {Integer} userId 创建人ID。
@apiSuccess {Integer} categorizeId 所属分类ID。
@apiSuccess {String} description 用例描述。
@apiSuccess {Integer} method 请求方法：
<ul>
   	<li>0 - GET</li>
    <li>1 - POST</li>
	<li>2 - PUT</li>
	<li>3 - DELETE</li>
</ul>
@apiSuccess {String} path 请求路径。
@apiSuccess {Map} params 请求参数。
@apiSuccess {map} headers 请求头。
@apiSuccess {Integer} bodyType 请求体类型：
<ul>
   	<li>0 - none</li>
    <li>1 - json</li>
	<li>2 - form</li>
	<li>3 - binary</li>
</ul>
@apiSuccess {Map} body 请求体。


@apiSuccessExample 成功响应
HTTP/1.1 200 OK
[
    {
        "id": 3,
        "createdAt": "2023-03-22 14:10:27",
        "updatedAt": "2023-03-22 14:10:27",
        "name": "pdm1",
        "type": "categorize"
    }
]
@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/
// info 查询具体接口用例信息
func (c *CasesController) info(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Query("id"))
	if id <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	caseInfo := entity.ApiCase{}
	err := repo.DB.First(&caseInfo, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "该接口用例不存在")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	ctx.JSON(200, &caseInfo)
}

/**
@api {POST} /api/case/edit 编辑接口用例
@apiDescription 编辑接口用例。
接口用例名称在同一级中唯一，不可重复。
@apiName CaseEdit
@apiGroup Case

@apiPermission 项目成员


@apiParam {Integer} id 用例ID。
@apiParam {String} name 用例名称。
@apiParam {String} description 用例描述。
@apiParam {Integer} method 请求方法：
<ul>
   	<li>0 - GET</li>
    <li>1 - POST</li>
	<li>2 - PUT</li>
	<li>3 - DELETE</li>
</ul>
@apiParam {String} path 请求路径。
@apiParam {String} params 请求参数。
@apiParam {String} headers 请求头。
@apiParam {Integer} bodyType 请求体类型：
<ul>
   	<li>0 - none</li>
    <li>1 - json</li>
	<li>2 - form</li>
	<li>3 - binary</li>
</ul>
@apiParam {String} body 请求体。

@apiSuccess {Integer} id 接口用例ID。

@apiParamExample {json} 请求示例
{
    "id":1,
    "name":"test",
    "description":"test",
    "method":0,
    "path":"",
    "params":"",
    "headers":"",
    "bodyType":0,
    "body":""
}
@apiSuccessExample 成功响应
HTTP/1.1 200 OK

1

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

接口名称已经存在
*/

// edit 编辑接口用例
func (c *CasesController) edit(ctx *gin.Context) {
	var info entity.ApiCase
	if err := ctx.BindJSON(&info); err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	applog.L(ctx, "编辑接口用例", map[string]interface{}{
		"id":   info.ID,
		"name": info.Name,
	})

	if info.Method < entity.MethodGet || info.Method > entity.MethodDelete ||
		info.BodyType < entity.BodyTypeNone || info.BodyType > entity.BodyTypeBinary {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 用例名称不能为空
	if len(strings.Trim(info.Name, " ")) == 0 {
		ErrIllegal(ctx, "用例名称不能为空")
		return
	}
	// 获取数据库用例信息
	caseInfo := entity.ApiCase{}
	err := repo.DB.First(&caseInfo, "id = ?", info.ID).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "该接口用例不存在")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 用例名称唯一
	if info.Name != caseInfo.Name {
		exist, err := repo.CaseRepo.ExistName(info.Name, caseInfo.CategorizeId)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		if exist {
			ErrIllegal(ctx, "用例名称已经存在")
			return
		}
	}
	if info.UserId <= 0 {
		info.UserId = caseInfo.UserId
	}
	if info.CategorizeId <= 0 {
		info.CategorizeId = caseInfo.CategorizeId
	}
	if err = repo.DB.Save(&info).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
	ctx.JSON(200, info.ID)
}

/**
@api {DELETE} /api/case/delete 删除接口用例
@apiDescription 删除接口用例，支持同时删除多个接口用例。
该接口仅在数据库操作异常时返回500系统错误的状态码，其他情况均返回200。
@apiName CaseDelete
@apiGroup Case

@apiPermission 项目负责人、接口用例创建者

@apiParam {String} ids 待删除的ID序列，多个ID用","隔开，如：ids=1,99。

@apiParamExample 请求示例
DELETE /api/case/delete?ids=1,2,3

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// delete 删除接口用例
func (c *CasesController) delete(ctx *gin.Context) {
	ids := ctx.Query("ids")
	//将string转化为[]int
	idArray := reuint.StrToIntSlice(ids)

	// 记录日志
	applog.L(ctx, "删除接口用例", map[string]interface{}{
		"ids": ids,
	})
	if err := repo.DB.Delete(&entity.ApiCase{}, "id in ?", idArray).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {POST} /api/case/send 发送测试请求
@apiDescription 发送测试请求。

@apiName CaseSend
@apiGroup Case

@apiPermission 项目成员


@apiParam {Integer} id 用例ID。
@apiParam {String} name 用例名称。
@apiParam {String} description 用例描述。
@apiParam {Integer} method 请求方法：
<ul>
   	<li>0 - GET</li>
    <li>1 - POST</li>
	<li>2 - PUT</li>
	<li>3 - DELETE</li>
</ul>
@apiParam {String} path 请求路径。
@apiParam {String} params 请求参数。
@apiParam {String} headers 请求头。
@apiParam {Integer} bodyType 请求体类型：
<ul>
   	<li>0 - none</li>
    <li>1 - json</li>
	<li>2 - form</li>
	<li>3 - binary</li>
</ul>
@apiParam {String} body 请求体。

@apiSuccess {String} respHeaders 响应头。
@apiSuccess {String} respBody 响应体。

@apiParamExample {json} 请求示例
{
    "id":1,
    "name":"test",
    "description":"test",
    "method":0,
    "path":"",
    "params":"",
    "headers":"",
    "bodyType":0,
    "body":""
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

1

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// send 发送测试请求
func (c *CasesController) send(ctx *gin.Context) {
	var info entity.ApiCase
	if err := ctx.BindJSON(&info); err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	// 处理地址
	index := strings.Index(info.Path, "?")
	if index != -1 {
		info.Path = info.Path[:index]
	}
	var req *http.Request
	var resp *http.Response
	reqInfo := dto.RespDto{}
	switch info.Method {
	case entity.MethodGet:
		// 请求地址
		url, err := reuint.GenRequestUrl(info.Path, info.Params)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		// GET请求
		req, _ = http.NewRequest(http.MethodGet, url, nil)
		// 加入请求头
		request, err := reuint.GenRequestHeader(req, info.Headers)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		// 发送请求
		resp, err = http.DefaultClient.Do(request)
		if err != nil {
			ErrIllegal(ctx, "发送请求失败")
			return
		}

	case entity.MethodPost:
		// 解析生成请求体数据
		bodyJson, bodyForm, err := reuint.GenRequestBody(info.BodyType, info.Body)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		// POST请求
		if info.BodyType == entity.BodyTypeJson {
			if bodyJson == nil {
				req, _ = http.NewRequest(http.MethodPost, info.Path, nil)
			} else {
				req, _ = http.NewRequest(http.MethodPost, info.Path, bodyJson)
			}
			// 加入请求头
			request, err := reuint.GenRequestHeader(req, info.Headers)
			if err != nil {
				ErrSys(ctx, err)
				return
			}
			resp, err = http.DefaultClient.Do(request)
			if err != nil {
				ErrIllegal(ctx, "发送请求失败")
				return
			}
		} else if info.BodyType == entity.BodyTypeForm {
			resp, err = http.PostForm(info.Path, bodyForm)
			if err != nil {
				ErrIllegal(ctx, "发送请求失败")
				return
			}
		} else if info.BodyType == entity.BodyTypeBinary {
			// TODO 二进制文件
		}

	case entity.MethodPut:
		// 解析生成请求体数据
		bodyJson, bodyForm, err := reuint.GenRequestBody(info.BodyType, info.Body)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		// POST请求
		if info.BodyType == entity.BodyTypeJson {
			if bodyJson == nil {
				req, _ = http.NewRequest(http.MethodPut, info.Path, nil)
			} else {
				req, _ = http.NewRequest(http.MethodPut, info.Path, bodyJson)
			}
			// 加入请求头
			request, err := reuint.GenRequestHeader(req, info.Headers)
			if err != nil {
				ErrSys(ctx, err)
				return
			}
			resp, err = http.DefaultClient.Do(request)
			if err != nil {
				ErrIllegal(ctx, "发送请求失败")
				return
			}
		} else if info.BodyType == entity.BodyTypeForm {
			resp, err = http.PostForm(info.Path, bodyForm)
			if err != nil {
				ErrIllegal(ctx, "发送请求失败")
				return
			}
		} else if info.BodyType == entity.BodyTypeBinary {
			// TODO 二进制文件
		}

	case entity.MethodDelete:
		// 请求地址
		url, err := reuint.GenRequestUrl(info.Path, info.Params)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		// GET请求
		req, _ = http.NewRequest(http.MethodDelete, url, nil)
		// 加入请求头
		request, err := reuint.GenRequestHeader(req, info.Headers)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		// 发送请求
		resp, err = http.DefaultClient.Do(request)
		if err != nil {
			ErrIllegal(ctx, "发送请求失败")
			return
		}
	default:
		ErrIllegal(ctx, "参数非法，无法解析")
	}
	// 生成正确格式响应体
	respBody, err := reuint.ParsingResponseBody(info.Method, resp)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	//  生成正确格式响应头
	respHeader, err := reuint.ParsingResponseHeader(resp)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	reqInfo.Transform(respHeader, respBody, "")
	ctx.JSON(200, reqInfo)
}
