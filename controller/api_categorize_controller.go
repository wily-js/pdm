package controller

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pdm/controller/dto"
	"pdm/controller/middle"
	"pdm/logg/applog"
	"pdm/repo"
	"pdm/repo/entity"
	"pdm/reuint/jwt"
	"strconv"
	"strings"
)

// NewCategorizeController 创建接口分类控制器
func NewCategorizeController(router gin.IRouter) *CategorizeController {
	res := &CategorizeController{}
	r := router.Group("/categorize")
	// 创建分类
	r.POST("/create", res.create)
	// 关键字查询分类或接口
	r.GET("/search", res.search)
	// 查询出分类下的子分类和接口列表
	r.GET("/list", res.list)
	// 编辑分类
	r.POST("/edit", res.edit)
	// 删除分类
	r.DELETE("/delete", res.delete)

	return res
}

// CategorizeController 接口分类控制器
type CategorizeController struct {
}

/**
@api {POST} /api/categorize/create 创建接口分类
@apiDescription 创建接口分类
接口分类名称在同一级中唯一，不可重复。
@apiName CategorizeCreate
@apiGroup Categorize

@apiPermission 项目成员

@apiParam {String} name 接口分类名称。
@apiParam {Integer} [parentId] 父分类ID。

@apiSuccess {Integer} id 接口分类ID。
@apiSuccess {String} createdAt 创建时间，格式为"YYYY-MM-DD HH:mm:ss"。
@apiSuccess {String} updatedAt 更新时间，格式为"YYYY-MM-DD HH:mm:ss"。
@apiSuccess {String} name 接口分类名称。
@apiSuccess {String} type 类型:"categorize"

@apiParamExample {json} 请求示例
{
    "name": "pdm1"
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
{
    "id": 2,
    "name": "pdm1",
    "createdAt": "2023-03-22 14:05:29",
    "updatedAt": "2023-03-22 14:05:29",
    "type": "categorize"
}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

接口分类已经存在
*/

// create 创建分类
func (c *CategorizeController) create(ctx *gin.Context) {
	var info entity.ApiCategorize
	if err := ctx.BindJSON(&info); err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	applog.L(ctx, "创建接口分类", map[string]interface{}{
		"name":     info.Name,
		"parentId": info.ParentId,
	})
	// 分类名称不能为空
	if len(strings.Trim(info.Name, " ")) == 0 {
		ErrIllegal(ctx, "分类名称不能为空")
		return
	}
	// 分类名称唯一
	exist, err := repo.CategorizeRepo.ExistName(info.Name, info.ParentId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if exist {
		ErrIllegal(ctx, "分类名称已经存在")
		return
	}

	// 获取所属项目ID和创建人ID
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	info.UserId = claims.Sub
	info.ProjectId = claims.PID

	if err = repo.DB.Create(&info).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
	reqInfo := dto.CategorizeListDto{}
	reqInfo.Transform(&info, &entity.ApiCase{}, "categorize")
	ctx.JSON(200, reqInfo)
}

/**
@api {GET} /api/categorize/search 查找接口分类
@apiDescription 根据关键字查询分类。
@apiName CategorizeSearch
@apiGroup Categorize

@apiPermission 项目成员

@apiParam {String} [keyword] 接口分类名称。

@apiParamExample {get} 请求示例
GET /api/categorize/search?keyword=pdm

@apiSuccess {Categorize[]} Categorize 查询结果列表。

@apiSuccess {Categorize} Categorize 分类数据结构。
@apiSuccess {Integer} Categorize.id 分类ID。
@apiSuccess {String} Categorize.name 分类名称。
@apiSuccess {String} User.createdAt 创建时间。
@apiSuccess {String} User.updatedAt 更新时间。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
{
	[
		{
		    "id": 5,
		    "createdAt": "2020-09-26 11:29:44",
		    "updatedAt": "2020-09-26 11:29:44",
		    "name": "测试",
		}
    ]
}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

分类不存在
*/

// search 关键字查询
func (c *CategorizeController) search(ctx *gin.Context) {
	// TODO 关键字查询
}

/**
@api {GET} /api/categorize/list 接口分类列表
@apiDescription 查询出分类下的子分类和接口列表。
未携带参数时，查找出根分类以及根接口列表。

@apiName CategorizeList
@apiGroup Categorize

@apiPermission 项目成员

@apiParam {Integer} [parentId] 父分类ID。

@apiParamExample {get} 请求示例
GET /api/categorize/list?parentId=1

@apiSuccess {List[]} Body 查询结果列表。

@apiSuccess (List) {Integer} id 分类ID或用例ID。
@apiSuccess (List) {String} name 分类名称或用例名称。
@apiSuccess (List) {String} createdAt 创建时间。
@apiSuccess (List) {String} updatedAt 更新时间。
@apiSuccess (List) {String} type 结果类型。
<ul>
	    <li>categorize</li>
	    <li>case</li>
</ul>

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
// list 分类列表
func (c *CategorizeController) list(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Query("parentId"))
	// 获取分类列表
	categorize := []entity.ApiCategorize{}
	if err := repo.DB.Find(&categorize, "parent_id = ?", id).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
	reqInfo := make([]dto.CategorizeListDto, 0)
	for _, val := range categorize {
		temp := dto.CategorizeListDto{}
		temp.Transform(&val, &entity.ApiCase{}, "categorize")
		reqInfo = append(reqInfo, temp)
	}

	// 获取用例列表
	cases := []entity.ApiCase{}
	if err := repo.DB.Find(&cases, "categorize_id = ?", id).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
	for _, val := range cases {
		temp := dto.CategorizeListDto{}
		temp.Transform(&entity.ApiCategorize{}, &val, "case")
		reqInfo = append(reqInfo, temp)
	}
	ctx.JSON(200, reqInfo)
}

/**
@api {POST} /api/categorize/edit 编辑接口分类
@apiDescription 编辑接口分类
接口分类名称在同一级中唯一，不可重复。
@apiName CategorizeEdit
@apiGroup Categorize

@apiPermission 项目负责人，创建者

@apiParam {Integer} id 分类ID。
@apiParam {String} name 接口分类名称。

@apiSuccess {Integer} id 接口分类ID。
@apiSuccess {String} createdAt 创建时间，格式为"YYYY-MM-DD HH:mm:ss"。
@apiSuccess {String} updatedAt 更新时间，格式为"YYYY-MM-DD HH:mm:ss"。
@apiSuccess {Integer} parentId 父分类ID。
@apiSuccess {String} name 接口分类名称。
@apiSuccess {Integer} projectId 所属项目ID。
@apiSuccess {Integer} userId 创建者ID。


@apiParamExample {json} 请求示例
{
    "id": 1,
    "name": "pdm"
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
{
    "id": 2,
    "parentId": 0,
    "name": "pdm1",
    "projectId": 1,
    "userId": 1,
    "createdAt": "2023-03-22 14:05:29",
    "updatedAt": "2023-03-22 14:05:29"
}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

接口分类名称已存在
*/
// edit 编辑接口分类
func (c *CategorizeController) edit(ctx *gin.Context) {
	var info entity.ApiCategorize
	if err := ctx.BindJSON(&info); err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	applog.L(ctx, "编辑接口分类", map[string]interface{}{
		"id":   info.ID,
		"name": info.Name,
	})

	// 分类名称不能为空
	if len(strings.Trim(info.Name, " ")) == 0 {
		ErrIllegal(ctx, "分类名称不能为空")
		return
	}
	reqInfo := entity.ApiCategorize{}
	err := repo.DB.First(&reqInfo, "id = ?", info.ID).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "不存在该分类")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 同级分类名称唯一
	if info.Name != reqInfo.Name {
		exist, err := repo.CategorizeRepo.ExistName(info.Name, reqInfo.ParentId)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		if exist {
			ErrIllegal(ctx, "分类名称已经存在")
			return
		}
	}
	reqInfo.Name = info.Name
	if err = repo.DB.Save(&reqInfo).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
	ctx.JSON(200, &reqInfo)
}

/**
@api {DELETE} /api/categorize/delete 删除接口分类
@apiDescription 删除接口分类，删除时同时删除该分类下的所有子分类和接口，支持同时删除多个接口分类。
该接口仅在数据库操作异常时返回500系统错误的状态码，其他情况均返回200。
@apiName CategorizeDelete
@apiGroup Categorize

@apiPermission 项目负责人、接口分类创建者

@apiParam {String} ids 待删除的ID序列，多个ID用","隔开，如：ids=1,99。

@apiParamExample 请求示例
DELETE /api/categorize/delete?ids=1,2,3

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// delete 删除接口分类
func (c *CategorizeController) delete(ctx *gin.Context) {
	// TODO 删除接口分类
}
