package controller

import (
	"github.com/gin-gonic/gin"
)

// NewCasesController 创建接口用例控制器
func NewCasesController(router gin.IRouter) *CasesController {
	res := &CasesController{}
	r := router.Group("/cases")
	// 接口用例创建
	r.POST("/create", res.create)
	// 查询接口用例具体信息
	r.GET("/info", res.info)
	// 编辑接口用例
	r.POST("/edit", res.edit)
	// 删除接口用例
	r.DELETE("/delete", res.delete)
	// 发送POST测试请求
	r.POST("/post", res.post)
	// 发送GET测试请求
	r.POST("/get", res.get)
	return res
}

// CasesController 接口用例控制器
type CasesController struct {
}

func (c *CasesController) create(ctx *gin.Context) {

}

func (c *CasesController) info(ctx *gin.Context) {

}

func (c *CasesController) edit(ctx *gin.Context) {

}

func (c *CasesController) delete(ctx *gin.Context) {

}

func (c *CasesController) post(ctx *gin.Context) {

}

func (c *CasesController) get(ctx *gin.Context) {

}
