package controller

import "github.com/gin-gonic/gin"

// NewCategorizeController 创建接口分类控制器
func NewCategorizeController(router gin.IRouter) *CategorizeController {
	res := &CategorizeController{}
	r := router.Group("/categorize")
	// 创建分类
	r.POST("/create", res.create)
	// 查询分类
	r.GET("/search", res.search)
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
@apiName CategorizeCreate
@apiGroup Categorize

@apiPermission 项目成员

@apiParam {String} name 接口分类名称。
@apiParam {Integer} [parentId] 父分类ID。

@apiSuccess {Integer} id 接口分类ID。
@apiSuccess {Integer} parentId 父分类ID。
@apiSuccess {String} name 接口分类名称。
@apiSuccess {String} createdAt 创建时间，格式为"YYYY-MM-DD HH:mm:ss"。

@apiParamExample {json} 请求示例
{
    "name": "张三"
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
{
    "id": 1,
	"parentId":"",
    "name": "pdm",
    "createdAt": "2020-08-24 16:26:16"
}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

接口分类已经存在
*/

func (c *CategorizeController) create(ctx *gin.Context) {

}

func (c *CategorizeController) search(ctx *gin.Context) {

}

func (c *CategorizeController) edit(ctx *gin.Context) {

}

func (c *CategorizeController) delete(ctx *gin.Context) {

}
