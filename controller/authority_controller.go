package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"pdm/controller/middle"
	"pdm/logg/applog"
	"pdm/repo"
	"pdm/repo/entity"
	"pdm/reuint/jwt"
)

type AuthorityController struct {
}

func NewAuthorityController(router gin.IRouter) *AuthorityController {
	res := &AuthorityController{}
	r := router.Group("/auth")
	// 进入项目
	r.POST("/enterProject", User, res.enterProject)
	// 退出项目
	r.DELETE("/exitProject", ProjectMember, res.exitProject)
	return res
}

/**
@api {POST} /api/auth/enterProject 进入项目
@apiDescription 将projectId和role存入cookies
@apiName AuthCreate
@apiGroup Auth

@apiPermission 项目成员

@apiParam {Integer} projectId 项目ID。

@apiSuccess {Integer} projectId 项目ID。
@apiSuccess {Integer} role 角色。

@apiParamExample {json} 请求示例
{
    "projectId": 1
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
{
    "projectId": 1,
	"role":1
}

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// enterProject 进入项目
func (c *AuthorityController) enterProject(ctx *gin.Context) {
	var param entity.Authority
	err := ctx.BindJSON(&param)
	log.Println(param.ProjectId)
	applog.L(ctx, "进入项目", map[string]interface{}{"projectId": param.ProjectId})
	if param.ProjectId <= 0 || err != nil {
		ErrIllegal(ctx, "参数错误，无法解析")
		return
	}

	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	if claims.Type != "user" {
		ErrIllegal(ctx, "用户类型错误")
		return
	}

	err = repo.DB.Model(&entity.ProjectMember{}).Select("role").
		First(&param.Role, "project_id = ? AND user_id = ?", param.ProjectId, claims.Sub).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "不是项目成员")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 生成新的Token包含项目ID和项目角色
	claims.PID = param.ProjectId
	claims.Role = param.Role
	token := tokenManager.GenToken(claims)
	fmt.Println(token)
	ctx.SetCookie("token", token, 8*3600, "", "", false, true)
	ctx.JSON(200, &param)
}

/**
@api {DELETE} /api/auth/exitProject 退出项目
@apiDescription 清除cookies中的projectId和role，无论清除操作是否成功均返回200状态码无任何信息。
@apiName AuthDelete
@apiGroup Auth

@apiPermission 项目成员


@apiParamExample 请求示例
DELETE /api/auth/exitProject

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

*/
// exitProject 退出项目
func (c *AuthorityController) exitProject(ctx *gin.Context) {
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	// 生成新的Token包含项目ID和项目角色
	claims.PID = 0
	claims.Role = 0
	token := tokenManager.GenToken(claims)
	ctx.SetCookie("token", token, 8*3600, "", "", false, true)
}
